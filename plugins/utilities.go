package plugins

import (
	"math"

	"github.com/abrander/agento/timeseries"
)

func SimplePoint(key string, value interface{}) *timeseries.Point {
	return timeseries.NewPoint(key, nil, map[string]interface{}{
		"value": value,
	})
}

func PointWithTag(key string, value interface{}, tagKey string, tagValue string) *timeseries.Point {
	return timeseries.NewPoint(
		key,
		map[string]string{
			tagKey: tagValue,
		},
		map[string]interface{}{
			"value": value,
		},
	)
}

func PointWithTags(key string, value interface{}, tags map[string]string) *timeseries.Point {
	return timeseries.NewPoint(
		key,
		tags,
		map[string]interface{}{
			"value": value,
		},
	)
}

func PointsWithTags(key string, values map[string]interface{}, tags map[string]string) *timeseries.Point {
	return timeseries.NewPoint(
		key,
		tags,
		values,
	)
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
