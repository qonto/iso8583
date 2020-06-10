// package iso8583 provides ...
package iso8583

import (
	"errors"
	"fmt"
	"strconv"
)

// Llvar contains bytes in non-fixed length field, first 2 symbols of field contains length
type Llvar struct {
	Value []byte
}

// NewLlvar create new Llvar field
func NewLlvar(val []byte) *Llvar {
	return &Llvar{val}
}

// IsEmpty check Llvar field for empty value
func (l *Llvar) IsEmpty() bool {
	return len(l.Value) == 0
}

// Bytes encode Llvar field to bytes
func (l *Llvar) Bytes(encoder, lenEncoder, length int) ([]byte, error) {
	val, err := UTF8ToWindows1252(l.Value)
	if err != nil {
		return nil, err
	}
	if length != -1 && len(val) > length {
		return nil, fmt.Errorf(ERR_VALUE_TOO_LONG, "Llvar", length, len(val))
	}
	if encoder != ASCII {
		return nil, errors.New(ERR_INVALID_ENCODER)
	}

	lenStr := fmt.Sprintf("%02d", len(val))
	contentLen := []byte(lenStr)
	var lenVal []byte
	switch lenEncoder {
	case ASCII:
		lenVal = contentLen
		if len(lenVal) > 2 {
			return nil, errors.New(ERR_INVALID_LENGTH_HEAD)
		}
	case rBCD:
		fallthrough
	case BCD:
		lenVal = rbcd(contentLen)
		if len(lenVal) > 1 {
			return nil, errors.New(ERR_INVALID_LENGTH_HEAD)
		}
	default:
		return nil, errors.New(ERR_INVALID_LENGTH_ENCODER)
	}
	return append(lenVal, val...), nil
}

// Load decode Llvar field from bytes
func (l *Llvar) Load(raw []byte, encoder, lenEncoder, length int) (read int, err error) {
	raw, err = UTF8ToWindows1252(raw)
	if err != nil {
		return 0, err
	}
	//parse length head:
	var contentLen int
	switch lenEncoder {
	case ASCII:
		read = 2
		contentLen, err = strconv.Atoi(string(raw[:read]))
		if err != nil {
			return 0, errors.New(ERR_PARSE_LENGTH_FAILED + ": " + string(raw[:2]))
		}
	case rBCD:
		fallthrough
	case BCD:
		read = 1
		contentLen, err = strconv.Atoi(string(bcdr2Ascii(raw[:read], 2)))
		if err != nil {
			return 0, errors.New(ERR_PARSE_LENGTH_FAILED + ": " + string(raw[0]))
		}
	default:
		return 0, errors.New(ERR_INVALID_LENGTH_ENCODER)
	}
	if len(raw) < (read + contentLen) {
		return 0, errors.New(ERR_BAD_RAW)
	}
	// parse body:
	l.Value = raw[read : read+contentLen]
	read += contentLen
	if encoder != ASCII {
		return 0, errors.New(ERR_INVALID_ENCODER)
	}

	return read, nil
}