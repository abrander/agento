package timeseries

type (
	Database interface {
		WritePoints(points []*Point) error
	}
)
