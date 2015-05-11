package agento

import (
	"encoding/json"
	"strconv"
)

type SingleCpuStat struct {
	User      float64
	Nice      float64
	System    float64
	Idle      float64
	IoWait    float64 // Since 2.5.41
	Irq       float64 // Since 2.6.0-test4
	SoftIrq   float64 // Since 2.6.0-test4
	Steal     float64 // Since 2.6.11
	Guest     float64 // Since 2.6.24
	GuestNice float64 // Since 2.6.33
}

func (s *SingleCpuStat) ReadArray(data []string) {
	l := len(data)

	if l > 4 {
		s.User, _ = strconv.ParseFloat(data[1], 64)
		s.Nice, _ = strconv.ParseFloat(data[2], 64)
		s.System, _ = strconv.ParseFloat(data[3], 64)
		s.Idle, _ = strconv.ParseFloat(data[4], 64)
	}

	if l > 5 {
		s.IoWait, _ = strconv.ParseFloat(data[5], 64)
	}

	if l > 6 {
		s.Irq, _ = strconv.ParseFloat(data[6], 64)
	}

	if l > 7 {
		s.SoftIrq, _ = strconv.ParseFloat(data[7], 64)
	}

	if l > 8 {
		s.Steal, _ = strconv.ParseFloat(data[8], 64)
	}

	if l > 9 {
		s.Guest, _ = strconv.ParseFloat(data[9], 64)
	}

	if l > 10 {
		s.GuestNice, _ = strconv.ParseFloat(data[10], 64)
	}
}

func (s SingleCpuStat) MarshalJSON() ([]byte, error) {
	var a [10]float64

	a[0] = Round(s.User, 1)
	a[1] = Round(s.Nice, 1)
	a[2] = Round(s.System, 1)
	a[3] = Round(s.Idle, 1)
	a[4] = Round(s.IoWait, 1)
	a[5] = Round(s.Irq, 1)
	a[6] = Round(s.SoftIrq, 1)
	a[7] = Round(s.Steal, 1)
	a[8] = Round(s.Guest, 1)
	a[9] = Round(s.GuestNice, 1)

	return json.Marshal(a)
}

func (s *SingleCpuStat) UnmarshalJSON(b []byte) error {
	var a [10]float64

	err := json.Unmarshal(b, &a)
	if err != nil {
		return err
	}

	s.User = a[0]
	s.Nice = a[1]
	s.System = a[2]
	s.Idle = a[3]
	s.IoWait = a[4]
	s.Irq = a[5]
	s.SoftIrq = a[6]
	s.Steal = a[7]
	s.Guest = a[8]
	s.GuestNice = a[9]

	return err
}

func (s *SingleCpuStat) Sub(previous *SingleCpuStat, factor float64) *SingleCpuStat {
	diff := SingleCpuStat{}

	diff.User = (s.User - previous.User) / factor
	diff.Nice = (s.Nice - previous.Nice) / factor
	diff.System = (s.System - previous.System) / factor
	diff.Idle = (s.Idle - previous.Idle) / factor
	diff.IoWait = (s.IoWait - previous.IoWait) / factor
	diff.Irq = (s.Irq - previous.Irq) / factor
	diff.SoftIrq = (s.SoftIrq - previous.SoftIrq) / factor
	diff.Steal = (s.Steal - previous.Steal) / factor
	diff.Guest = (s.Guest - previous.Guest) / factor
	diff.GuestNice = (s.GuestNice - previous.GuestNice) / factor

	return &diff
}
