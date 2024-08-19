package pgtype

import (
	"database/sql/driver"
	"encoding/binary"
	"fmt"

	"github.com/jackc/pgio"
)

type Int8multirange struct {
	Ranges []Int8range
	Status Status
}

func (dst *Int8multirange) Set(src interface{}) error {
	//untyped nil and typed nil interfaces are different
	if src == nil {
		*dst = Int8multirange{Status: Null}
		return nil
	}

	switch value := src.(type) {
	case Int8multirange:
		*dst = value
	case *Int8multirange:
		*dst = *value
	case string:
		return dst.DecodeText(nil, []byte(value))
	case []Int8range:
		if value == nil {
			*dst = Int8multirange{Status: Null}
		} else if len(value) == 0 {
			*dst = Int8multirange{Status: Present}
		} else {
			elements := make([]Int8range, len(value))
			for i := range value {
				if err := elements[i].Set(value[i]); err != nil {
					return err
				}
			}
			*dst = Int8multirange{
				Ranges: elements,
				Status: Present,
			}
		}
	case []*Int8range:
		if value == nil {
			*dst = Int8multirange{Status: Null}
		} else if len(value) == 0 {
			*dst = Int8multirange{Status: Present}
		} else {
			elements := make([]Int8range, len(value))
			for i := range value {
				if err := elements[i].Set(value[i]); err != nil {
					return err
				}
			}
			*dst = Int8multirange{
				Ranges: elements,
				Status: Present,
			}
		}
	default:
		return fmt.Errorf("cannot convert %v to Int8multirange", src)
	}

	return nil

}

func (dst Int8multirange) Get() interface{} {
	switch dst.Status {
	case Present:
		return dst
	case Null:
		return nil
	default:
		return dst.Status
	}
}

func (src *Int8multirange) AssignTo(dst interface{}) error {
	return fmt.Errorf("cannot assign %v to %T", src, dst)
}

func (dst *Int8multirange) DecodeText(ci *ConnInfo, src []byte) error {
	if src == nil {
		*dst = Int8multirange{Status: Null}
		return nil
	}

	utmr, err := ParseUntypedTextMultirange(string(src))
	if err != nil {
		return err
	}

	var elements []Int8range

	if len(utmr.Elements) > 0 {
		elements = make([]Int8range, len(utmr.Elements))

		for i, s := range utmr.Elements {
			var elem Int8range

			elemSrc := []byte(s)

			err = elem.DecodeText(ci, elemSrc)
			if err != nil {
				return err
			}

			elements[i] = elem
		}
	}

	*dst = Int8multirange{Ranges: elements, Status: Present}

	return nil
}

func (dst *Int8multirange) DecodeBinary(ci *ConnInfo, src []byte) error {
	if src == nil {
		*dst = Int8multirange{Status: Null}
		return nil
	}

	rp := 0

	numElems := int(binary.BigEndian.Uint32(src[rp:]))
	rp += 4

	if numElems == 0 {
		*dst = Int8multirange{Status: Present}
		return nil
	}

	elements := make([]Int8range, numElems)

	for i := range elements {
		elemLen := int(int32(binary.BigEndian.Uint32(src[rp:])))
		rp += 4
		var elemSrc []byte
		if elemLen >= 0 {
			elemSrc = src[rp : rp+elemLen]
			rp += elemLen
		}
		err := elements[i].DecodeBinary(ci, elemSrc)
		if err != nil {
			return err
		}
	}

	*dst = Int8multirange{Ranges: elements, Status: Present}
	return nil
}

func (src Int8multirange) EncodeText(ci *ConnInfo, buf []byte) ([]byte, error) {
	switch src.Status {
	case Null:
		return nil, nil
	case Undefined:
		return nil, errUndefined
	}

	buf = append(buf, '{')

	inElemBuf := make([]byte, 0, 32)
	for i, elem := range src.Ranges {
		if i > 0 {
			buf = append(buf, ',')
		}

		elemBuf, err := elem.EncodeText(ci, inElemBuf)
		if err != nil {
			return nil, err
		}
		if elemBuf == nil {
			return nil, fmt.Errorf("multi-range does not allow null range")
		} else {
			buf = append(buf, string(elemBuf)...)
		}

	}

	buf = append(buf, '}')

	return buf, nil
}

func (src Int8multirange) EncodeBinary(ci *ConnInfo, buf []byte) ([]byte, error) {
	switch src.Status {
	case Null:
		return nil, nil
	case Undefined:
		return nil, errUndefined
	}

	buf = pgio.AppendInt32(buf, int32(len(src.Ranges)))

	for i := range src.Ranges {
		sp := len(buf)
		buf = pgio.AppendInt32(buf, -1)

		elemBuf, err := src.Ranges[i].EncodeBinary(ci, buf)
		if err != nil {
			return nil, err
		}
		if elemBuf != nil {
			buf = elemBuf
			pgio.SetInt32(buf[sp:], int32(len(buf[sp:])-4))
		}
	}

	return buf, nil
}

// Scan implements the database/sql Scanner interface.
func (dst *Int8multirange) Scan(src interface{}) error {
	if src == nil {
		return dst.DecodeText(nil, nil)
	}

	switch src := src.(type) {
	case string:
		return dst.DecodeText(nil, []byte(src))
	case []byte:
		srcCopy := make([]byte, len(src))
		copy(srcCopy, src)
		return dst.DecodeText(nil, srcCopy)
	}

	return fmt.Errorf("cannot scan %T", src)
}

// Value implements the database/sql/driver Valuer interface.
func (src Int8multirange) Value() (driver.Value, error) {
	return EncodeValueText(src)
}
