package monitor

import (
	"math/rand"
	"sync"
	"time"

	"github.com/abrander/agento/core"
	"github.com/abrander/agento/logger"
	"github.com/abrander/agento/plugins"
	"github.com/abrander/agento/timeseries"
	"github.com/abrander/agento/userdb"
	"github.com/influxdata/influxdb/client/v2"
)

type (
	// Scheduler is a scheduler executing probes.
	Scheduler struct {
		store   core.Store
		subject userdb.Subject
	}
)

// NewScheduler will instantiate a new scheduler. The scheduler needs a Store to
// read/write checks. If the system is not a multiuser system, userdb.God can be
// used as subject.
func NewScheduler(store core.Store, subject userdb.Subject) *Scheduler {
	return &Scheduler{
		store:   store,
		subject: subject,
	}
}

// Loop will simply loop through all probes and emit changes and execute jobs.
func (s *Scheduler) Loop(wg sync.WaitGroup, serv timeseries.Database) {
	// Make sure we have the magic localhost. Maybe we should move this somewhere else.
	_, err := s.store.GetHost(s.subject, "000000000000000000000000")
	if err != nil {
		// If localhost/localtransport does not exist. Do something.
		t, err := plugins.GetTransport("localtransport")
		if err != nil {
			logger.Red("scheduler", "localtransport plugin not found")
		}

		// Construct the magic host.
		host := &core.Host{
			ID:        "000000000000000000000000",
			AccountID: "000000000000000000000000",
			Name:      "localhost",
			Transport: t,
		}

		// Save it.
		err = s.store.AddHost(nil, host)
		if err != nil {
			logger.Red("scheduler", "Error inserting: %s", err.Error())
			wg.Done()
			return
		}

		logger.Yellow("scheduler", "Added localhost transport with id %s", host.ID)
	}

	// We tick ten times a second, this should be enough for now.
	ticker := time.Tick(time.Millisecond * 100)

	// inFlight is a list of probes id's currently running
	inFlight := make(map[string]bool)
	inFlightLock := sync.RWMutex{}
	for t := range ticker {
		// We start by extracting a list of all probes. If this gets too
		// expensive at some point, we can do it less frequent.

		probes, err := s.store.GetAllProbes(s.subject, userdb.God.GetAccountId())

		if err != nil {
			logger.Red("scheduler", "Error getting probes from store: %s", err.Error())
			continue
		}

		// We iterate the list of probes, to see if anything needs to be done.
		for _, probe := range probes {
			// Calculate the age of the last check, if the age is positive, it's
			// in the past.
			age := t.Sub(probe.LastCheck)

			// Calculate how much we should wait before executing the job. If
			// the value is positive, it's in the future.
			wait := probe.NextCheck.Sub(t)

			// Check if the probe is already executing.
			inFlightLock.RLock()
			_, found := inFlight[probe.ID]
			inFlightLock.RUnlock()

			if found {
				continue
			}

			// If the check is older than two intervals, we treat it as new.
			if age > probe.Interval*2 && wait < -probe.Interval {
				checkIn := time.Duration(rand.Int63n(int64(probe.Interval)))
				probe.NextCheck = t.Add(checkIn)
				logger.Yellow("scheduler", "[%s] %T:(%+v): start delayed by %s", probe.ID, probe.Agent, probe.Agent, checkIn)

				err = s.store.UpdateProbe(s.subject, &probe)
				if err != nil {
					logger.Red("Error updating: %v", err.Error())
				}
			} else if wait < 0 {
				// If we arrive here, wait is sub-zero, which means that we
				// should execute now.
				inFlightLock.Lock()
				inFlight[probe.ID] = true
				inFlightLock.Unlock()

				// Execute the probe in its own go routine.
				go func(probe core.Probe) {
					host := probe.Host

					// Run the job.
					start := time.Now()

					err = probe.Agent.(plugins.Agent).Gather(host.Transport)
					if err != nil {
						logger.Red("scheduler", "[%s] %T(%+v) failed in %s: %s", probe.ID, probe.Agent, probe.Agent, time.Now().Sub(start), err.Error())
					} else {
						logger.Green("scheduler", "[%s] %T(%+v) ran in %s", probe.ID, probe.Agent, probe.Agent, time.Now().Sub(start))

						points := probe.Agent.(plugins.Agent).GetPoints()

						// Tag all points with hostname and arbitrary tags.
						for index, point := range points {
							tags := point.Tags()

							tags["hostname"] = host.Name

							for key, value := range probe.Tags {
								tags[key] = value
							}

							points[index], _ = client.NewPoint(point.Name(), tags, point.Fields())
						}

						// Write results to TSDB.
						err = serv.WritePoints(points)
						if err != nil {
							logger.Red("scheduler", "[%s] %T(%+v) WritePoints(): %s", probe.ID, probe.Agent, probe.Agent, err.Error())
						}

						// Save the result
						probe.LastPoints = points
					}

					// Save the check time and schedule next check.
					probe.LastCheck = t
					probe.NextCheck = t.Add(probe.Interval)

					// Save everything back to store.
					err = s.store.UpdateProbe(s.subject, &probe)
					if err != nil {
						logger.Red("scheduler", "[%s] %T(%+v) UpdateProbe(): %s", probe.ID, probe.Agent, probe.Agent, err.Error())
					}
					// Remove the probe from inFlight map.
					inFlightLock.Lock()
					delete(inFlight, probe.ID)
					inFlightLock.Unlock()
				}(probe)
			}
		}
	}

	wg.Done()
}
