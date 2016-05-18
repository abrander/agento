package server

import (
	"testing"
)

func TestComputeKey(t *testing.T) {
	cases := map[string]Sample{
		"1:id:a=bc=de=f": Sample{
			Type:       1,
			Identifier: "id",
			Tags: map[string]string{
				"a": "b",
				"c": "d",
				"e": "f",
			},
		},
		"1:id:a=bc=de=f2": Sample{
			Type:       1,
			Identifier: "id",
			Tags: map[string]string{
				"c": "d",
				"a": "b",
				"e": "f2",
			},
		},
		"1::a=bc=de=f": Sample{
			Type:       1,
			Identifier: "",
			Tags: map[string]string{
				"c": "d",
				"a": "b",
				"e": "f",
			},
		},
	}

	for expectedKey, sample := range cases {
		computedKey := sample.computeKey()
		if computedKey != expectedKey {
			t.Errorf("Key is '%s', should be '%s'", computedKey, expectedKey)
		}
	}
}
