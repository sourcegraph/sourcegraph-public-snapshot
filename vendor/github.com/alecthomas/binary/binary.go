package binary

import (
	"bytes"
	"encoding"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"reflect"
)

var (
	LittleEndian  = binary.LittleEndian
	BigEndian     = binary.BigEndian
	DefaultEndian = LittleEndian
)

func Marshal(v interface{}) ([]byte, error) {
	b := &bytes.Buffer{}
	if err := NewEncoder(b).Encode(v); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func Unmarshal(b []byte, v interface{}) error {
	return NewDecoder(bytes.NewReader(b)).Decode(v)
}

type Encoder struct {
	Order binary.ByteOrder
	w     io.Writer
	buf   []byte
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		Order: DefaultEndian,
		w:     w,
		buf:   make([]byte, 8),
	}
}

func (e *Encoder) writeVarint(v int) error {
	l := binary.PutUvarint(e.buf, uint64(v))
	_, err := e.w.Write(e.buf[:l])
	return err
}

func (b *Encoder) Encode(v interface{}) (err error) {
	switch cv := v.(type) {
	case encoding.BinaryMarshaler:
		buf, err := cv.MarshalBinary()
		if err != nil {
			return err
		}
		if err = b.writeVarint(len(buf)); err != nil {
			return err
		}
		_, err = b.w.Write(buf)

	case []byte: // fast-path byte arrays
		if err = b.writeVarint(len(cv)); err != nil {
			return
		}
		_, err = b.w.Write(cv)

	default:
		rv := reflect.Indirect(reflect.ValueOf(v))
		t := rv.Type()
		switch t.Kind() {
		case reflect.Array:
			l := t.Len()
			for i := 0; i < l; i++ {
				if err = b.Encode(rv.Index(i).Addr().Interface()); err != nil {
					return
				}
			}

		case reflect.Slice:
			l := rv.Len()
			if err = b.writeVarint(l); err != nil {
				return
			}
			for i := 0; i < l; i++ {
				if err = b.Encode(rv.Index(i).Addr().Interface()); err != nil {
					return
				}
			}

		case reflect.Struct:
			l := rv.NumField()
			for i := 0; i < l; i++ {
				if v := rv.Field(i); v.CanSet() && t.Field(i).Name != "_" {
					// take the address of the field, so structs containing structs
					// are correctly encoded.
					if err = b.Encode(v.Addr().Interface()); err != nil {
						return
					}
				}
			}

		case reflect.Map:
			l := rv.Len()
			if err = b.writeVarint(l); err != nil {
				return
			}
			for _, key := range rv.MapKeys() {
				value := rv.MapIndex(key)
				if err = b.Encode(key.Interface()); err != nil {
					return err
				}
				if err = b.Encode(value.Interface()); err != nil {
					return err
				}
			}

		case reflect.String:
			if err = b.writeVarint(rv.Len()); err != nil {
				return
			}
			_, err = b.w.Write([]byte(rv.String()))

		case reflect.Bool:
			var out byte
			if rv.Bool() {
				out = 1
			}
			err = binary.Write(b.w, b.Order, out)

		case reflect.Int:
			err = binary.Write(b.w, b.Order, int64(rv.Int()))

		case reflect.Uint:
			err = binary.Write(b.w, b.Order, int64(rv.Uint()))

		case reflect.Int8, reflect.Uint8, reflect.Int16, reflect.Uint16,
			reflect.Int32, reflect.Uint32, reflect.Int64, reflect.Uint64,
			reflect.Float32, reflect.Float64,
			reflect.Complex64, reflect.Complex128:
			err = binary.Write(b.w, b.Order, v)

		default:
			return errors.New("binary: unsupported type " + t.String())
		}
	}
	return
}

type byteReader struct {
	io.Reader
}

func (b *byteReader) ReadByte() (byte, error) {
	var buf [1]byte
	if _, err := io.ReadFull(b, buf[:]); err != nil {
		return 0, err
	}
	return buf[0], nil
}

type Decoder struct {
	Order binary.ByteOrder
	r     *byteReader
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		Order: DefaultEndian,
		r:     &byteReader{r},
	}
}

func (d *Decoder) Decode(v interface{}) (err error) {
	// Check if the type implements the encoding.BinaryUnmarshaler interface, and use it if so.
	if i, ok := v.(encoding.BinaryUnmarshaler); ok {
		var l uint64
		if l, err = binary.ReadUvarint(d.r); err != nil {
			return
		}
		buf := make([]byte, l)
		_, err = d.r.Read(buf)
		return i.UnmarshalBinary(buf)
	}

	// Otherwise, use reflection.
	rv := reflect.Indirect(reflect.ValueOf(v))
	if !rv.CanAddr() {
		return errors.New("binary: can only Decode to pointer type")
	}
	t := rv.Type()

	switch t.Kind() {
	case reflect.Array:
		len := t.Len()
		for i := 0; i < int(len); i++ {
			if err = d.Decode(rv.Index(i).Addr().Interface()); err != nil {
				return
			}
		}

	case reflect.Slice:
		var l uint64
		if l, err = binary.ReadUvarint(d.r); err != nil {
			return
		}
		if t.Kind() == reflect.Slice {
			rv.Set(reflect.MakeSlice(t, int(l), int(l)))
		} else if int(l) != t.Len() {
			return fmt.Errorf("binary: encoded size %d != real size %d", l, t.Len())
		}
		for i := 0; i < int(l); i++ {
			if err = d.Decode(rv.Index(i).Addr().Interface()); err != nil {
				return
			}
		}

	case reflect.Struct:
		l := rv.NumField()
		for i := 0; i < l; i++ {
			if v := rv.Field(i); v.CanSet() && t.Field(i).Name != "_" {
				if err = d.Decode(v.Addr().Interface()); err != nil {
					return
				}
			}
		}

	case reflect.Map:
		var l uint64
		if l, err = binary.ReadUvarint(d.r); err != nil {
			return
		}
		kt := t.Key()
		vt := t.Elem()
		rv.Set(reflect.MakeMap(t))
		for i := 0; i < int(l); i++ {
			kv := reflect.Indirect(reflect.New(kt))
			if err = d.Decode(kv.Addr().Interface()); err != nil {
				return
			}
			vv := reflect.Indirect(reflect.New(vt))
			if err = d.Decode(vv.Addr().Interface()); err != nil {
				return
			}
			rv.SetMapIndex(kv, vv)
		}

	case reflect.String:
		var l uint64
		if l, err = binary.ReadUvarint(d.r); err != nil {
			return
		}
		buf := make([]byte, l)
		_, err = d.r.Read(buf)
		rv.SetString(string(buf))

	case reflect.Bool:
		var out byte
		err = binary.Read(d.r, d.Order, &out)
		rv.SetBool(out != 0)

	case reflect.Int:
		var out int64
		err = binary.Read(d.r, d.Order, &out)
		rv.SetInt(out)

	case reflect.Uint:
		var out uint64
		err = binary.Read(d.r, d.Order, &out)
		rv.SetUint(out)

	case reflect.Int8, reflect.Uint8, reflect.Int16, reflect.Uint16,
		reflect.Int32, reflect.Uint32, reflect.Int64, reflect.Uint64,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128:
		err = binary.Read(d.r, d.Order, v)

	default:
		return errors.New("binary: unsupported type " + t.String())
	}
	return
}
