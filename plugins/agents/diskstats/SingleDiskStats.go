package diskstats

import (
	"encoding/json"
	"strconv"

	"github.com/abrander/agento/plugins"
)

type SingleDiskStats struct {
	ReadsCompleted  float64
	ReadsMerged     float64
	ReadSectors     float64
	ReadTime        float64
	WritesCompleted float64
	WritesMerged    float64
	WriteSectors    float64
	WriteTime       float64
	IoInProgress    int64
	IoTime          float64
	IoWeightedTime  float64
}

func (s *SingleDiskStats) ReadArray(data []string) {
	l := len(data)

	if l > 3 {
		s.ReadsCompleted, _ = strconv.ParseFloat(data[3], 64)
	}

	if l > 4 {
		s.ReadsMerged, _ = strconv.ParseFloat(data[4], 64)
	}

	if l > 5 {
		s.ReadSectors, _ = strconv.ParseFloat(data[5], 64)
	}

	if l > 6 {
		s.ReadTime, _ = strconv.ParseFloat(data[6], 64)
	}

	if l > 7 {
		s.WritesCompleted, _ = strconv.ParseFloat(data[7], 64)
	}

	if l > 8 {
		s.WritesMerged, _ = strconv.ParseFloat(data[8], 64)
	}

	if l > 9 {
		s.WriteSectors, _ = strconv.ParseFloat(data[9], 64)
	}

	if l > 10 {
		s.WriteTime, _ = strconv.ParseFloat(data[10], 64)
	}

	if l > 11 {
		s.IoInProgress, _ = strconv.ParseInt(data[11], 10, 32)
	}

	if l > 12 {
		s.IoTime, _ = strconv.ParseFloat(data[12], 64)
	}

	if l > 13 {
		s.IoWeightedTime, _ = strconv.ParseFloat(data[13], 64)
	}
}

func (s SingleDiskStats) MarshalJSON() ([]byte, error) {
	var a [11]float64

	a[0] = plugins.Round(s.ReadsCompleted, 1)
	a[1] = plugins.Round(s.ReadsMerged, 1)
	a[2] = plugins.Round(s.ReadSectors, 1)
	a[3] = plugins.Round(s.ReadTime, 1)
	a[4] = plugins.Round(s.WritesCompleted, 1)
	a[5] = plugins.Round(s.WritesMerged, 1)
	a[6] = plugins.Round(s.WriteSectors, 1)
	a[7] = plugins.Round(s.WriteTime, 1)
	a[8] = plugins.Round(float64(s.IoInProgress), 0)
	a[9] = plugins.Round(s.IoTime, 1)
	a[10] = plugins.Round(s.IoWeightedTime, 1)

	return json.Marshal(a)
}

func (s *SingleDiskStats) UnmarshalJSON(b []byte) error {
	var a [11]float64

	err := json.Unmarshal(b, &a)
	if err != nil {
		return err
	}

	s.ReadsCompleted = a[0]
	s.ReadsMerged = a[1]
	s.ReadSectors = a[2]
	s.ReadTime = a[3]
	s.WritesCompleted = a[4]
	s.WritesMerged = a[5]
	s.WriteSectors = a[6]
	s.WriteTime = a[7]
	s.IoInProgress = int64(a[8])
	s.IoTime = a[9]
	s.IoWeightedTime = a[10]

	return err
}
