package agento

import (
	"io/ioutil"
	"math"
	"strconv"
	"strings"

	"github.com/influxdb/influxdb/client"
)

func SimplePoint(key string, value interface{}) client.Point {
	return client.Point{
		Measurement: key,
		Fields: map[string]interface{}{
			"value": value,
		},
	}
}

func PointWithTag(key string, value interface{}, tagKey string, tagValue string) client.Point {
	return client.Point{
		Tags: map[string]string{
			tagKey: tagValue,
		},
		Measurement: key,
		Fields: map[string]interface{}{
			"value": value,
		},
	}
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
