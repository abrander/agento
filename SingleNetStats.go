package agento

import (
	"encoding/json"
	"strconv"
)

type SingleNetStats struct {
	RxBytes      float64
	RxPackets    float64
	RxErrors     float64
	RxDropped    float64
	RxFifo       float64
	RxFrame      float64
	RxCompressed float64
	RxMulticast  float64
	TxBytes      float64
	TxPackets    float64
	TxErrors     float64
	TxDropped    float64
	TxFifo       float64
	TxCollisions float64
	TxCarrier    float64
	TxCompressed float64
}

func (s *SingleNetStats) ReadArray(data []string) {
	l := len(data)

	if l > 1 {
		s.RxBytes, _ = strconv.ParseFloat(data[1], 64)
	}

	if l > 2 {
		s.RxPackets, _ = strconv.ParseFloat(data[2], 64)
	}

	if l > 3 {
		s.RxErrors, _ = strconv.ParseFloat(data[3], 64)
	}

	if l > 4 {
		s.RxDropped, _ = strconv.ParseFloat(data[4], 64)
	}

	if l > 5 {
		s.RxFifo, _ = strconv.ParseFloat(data[5], 64)
	}

	if l > 6 {
		s.RxFrame, _ = strconv.ParseFloat(data[6], 64)
	}

	if l > 7 {
		s.RxCompressed, _ = strconv.ParseFloat(data[7], 64)
	}

	if l > 8 {
		s.RxMulticast, _ = strconv.ParseFloat(data[8], 64)
	}

	if l > 9 {
		s.TxBytes, _ = strconv.ParseFloat(data[9], 64)
	}

	if l > 10 {
		s.TxPackets, _ = strconv.ParseFloat(data[10], 64)
	}

	if l > 11 {
		s.TxErrors, _ = strconv.ParseFloat(data[11], 64)
	}

	if l > 12 {
		s.TxDropped, _ = strconv.ParseFloat(data[12], 64)
	}

	if l > 13 {
		s.TxFifo, _ = strconv.ParseFloat(data[13], 64)
	}

	if l > 14 {
		s.TxCollisions, _ = strconv.ParseFloat(data[14], 64)
	}

	if l > 15 {
		s.TxCarrier, _ = strconv.ParseFloat(data[15], 64)
	}

	if l > 16 {
		s.TxCompressed, _ = strconv.ParseFloat(data[16], 64)
	}
}

func (s SingleNetStats) MarshalJSON() ([]byte, error) {
	var a [16]float64

	a[0] = Round(s.RxBytes, 0)
	a[1] = Round(s.RxPackets, 0)
	a[2] = Round(s.RxErrors, 0)
	a[3] = Round(s.RxDropped, 0)
	a[4] = Round(s.RxFifo, 0)
	a[5] = Round(s.RxFrame, 0)
	a[6] = Round(s.RxCompressed, 0)
	a[7] = Round(s.RxMulticast, 0)
	a[8] = Round(s.TxBytes, 0)
	a[9] = Round(s.TxPackets, 0)
	a[10] = Round(s.TxErrors, 0)
	a[11] = Round(s.TxDropped, 0)
	a[12] = Round(s.TxFifo, 0)
	a[13] = Round(s.TxCollisions, 0)
	a[14] = Round(s.TxCarrier, 0)
	a[15] = Round(s.TxCompressed, 0)

	return json.Marshal(a)
}

func (s *SingleNetStats) UnmarshalJSON(b []byte) error {
	var a [16]float64

	err := json.Unmarshal(b, &a)
	if err != nil {
		return err
	}

	s.RxBytes = a[0]
	s.RxPackets = a[1]
	s.RxErrors = a[2]
	s.RxDropped = a[3]
	s.RxFifo = a[4]
	s.RxFrame = a[5]
	s.RxCompressed = a[6]
	s.RxMulticast = a[7]
	s.TxBytes = a[8]
	s.TxPackets = a[9]
	s.TxErrors = a[10]
	s.TxDropped = a[11]
	s.TxFifo = a[12]
	s.TxCollisions = a[13]
	s.TxCarrier = a[14]
	s.TxCompressed = a[15]

	return err
}

func (s *SingleNetStats) Sub(previous *SingleNetStats, factor float64) *SingleNetStats {
	diff := SingleNetStats{}

	diff.RxBytes = (s.RxBytes - previous.RxBytes) / factor
	diff.RxPackets = (s.RxPackets - previous.RxPackets) / factor
	diff.RxErrors = (s.RxErrors - previous.RxErrors) / factor
	diff.RxDropped = (s.RxDropped - previous.RxDropped) / factor
	diff.RxFifo = (s.RxFifo - previous.RxFifo) / factor
	diff.RxFrame = (s.RxFrame - previous.RxFrame) / factor
	diff.RxCompressed = (s.RxCompressed - previous.RxCompressed) / factor
	diff.RxMulticast = (s.RxMulticast - previous.RxMulticast) / factor
	diff.TxBytes = (s.TxBytes - previous.TxBytes) / factor
	diff.TxPackets = (s.TxPackets - previous.TxPackets) / factor
	diff.TxErrors = (s.TxErrors - previous.TxErrors) / factor
	diff.TxDropped = (s.TxDropped - previous.TxDropped) / factor
	diff.TxFifo = (s.TxFifo - previous.TxFifo) / factor
	diff.TxCollisions = (s.TxCollisions - previous.TxCollisions) / factor
	diff.TxCarrier = (s.TxCarrier - previous.TxCarrier) / factor
	diff.TxCompressed = (s.TxCompressed - previous.TxCompressed) / factor

	return &diff
}
