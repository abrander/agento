package core

import (
	"testing"
)

func TestRandomString(t *testing.T) {
	for l := 0; l < 100; l++ {
		str := RandomString(l)

		if len(str) != l {
			t.Errorf("String is '%s' len()=%d, should be '%d'", str, len(str), l)
		}
	}
}
