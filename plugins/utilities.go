package plugins

import (
	"math"

	"github.com/influxdata/influxdb/client/v2"
)

func SimplePoint(key string, value interface{}) *client.Point {
	point, _ := client.NewPoint(key, nil, map[string]interface{}{
		"value": value,
	})

	return point
}

func PointWithTag(key string, value interface{}, tagKey string, tagValue string) *client.Point {
	point, _ := client.NewPoint(
		key,
		map[string]string{
			tagKey: tagValue,
		},
		map[string]interface{}{
			"value": value,
		},
	)

	return point
}

func Round(value float64, places int) float64 {
	var round float64

	pow := math.Pow(10, float64(places))

	digit := pow * value
	_, div := math.Modf(digit)
	if div >= 0.5 {
		round = math.Ceil(digit)
	} else {
		round = math.Floor(digit)
	}

	return round / pow
}
