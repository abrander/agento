package plugins

import (
	"encoding/json"

	"github.com/influxdb/influxdb/client"
)

type Results map[string]interface{}

func (r Results) GetPoints() []client.Point {
	points := make([]client.Point, 0, 300)

	for _, p := range r {
		pp := p.(Plugin)
		points = append(points, pp.GetPoints()...)
	}

	return points
}

func (r *Results) UnmarshalJSON(b []byte) error {
	var tmp = map[string]json.RawMessage{}

	err := json.Unmarshal(b, &tmp)
	if err != nil {
		return err
	}
	for t, v := range tmp {
		constructor, ok := pluginConstructors[t]
		if ok {
			res := constructor()

			err = json.Unmarshal(v, &res)
			if err != nil {
				return err
			}

			(*r)[t] = res
		}
		// Fail silently if we don't know the type to allow forward compatibility
	}

	return nil
}
