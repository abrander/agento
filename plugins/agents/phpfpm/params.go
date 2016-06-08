package phpfpm

import (
	"encoding/binary"
	"io"
)

type (
	// params is a simple type used for format parameters for FastCGI.
	params map[string]string
)

// size will predict the needed space for encoding.
func (p params) size() uint16 {
	accum := 0

	for key, value := range p {
		accum += 2
		accum += len(key)
		accum += len(value)
	}

	return uint16(accum)
}

// write the encoded parameters to w.
func (p params) write(w io.Writer) error {
	for key, value := range p {
		keyLen := len(key)
		valueLen := len(value)

		if keyLen > 255 || valueLen > 255 {
			continue
		}

		err := binary.Write(w, binary.BigEndian, byte(keyLen))
		if err != nil {
			return err
		}

		err = binary.Write(w, binary.BigEndian, byte(valueLen))
		if err != nil {
			return err
		}

		_, err = w.Write([]byte(key))
		if err != nil {
			return err
		}

		_, err = w.Write([]byte(value))
		if err != nil {
			return err
		}
	}

	return nil
}
