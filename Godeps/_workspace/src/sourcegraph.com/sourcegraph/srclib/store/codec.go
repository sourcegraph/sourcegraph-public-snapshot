package store

import (
	"encoding/binary"
	"encoding/json"
	"io"

	"sourcegraph.com/sourcegraph/srclib/store/pbio"
	"sourcegraph.com/sourcegraph/srclib/unit"

	"github.com/gogo/protobuf/proto"
)

// Codec is the codec used by all file-backed stores. It should only be
// set at init time or when you can guarantee that no stores will be
// using the codec.
var Codec codec = ProtobufCodec{}

// A codec is an encoder and decoder pair used by the FS-backed store
// to encode and decode data stored in files.
type codec interface {
	// NewEncoder creates a new encoder from w.
	NewEncoder(w io.Writer) encoder

	// NewDecoder creates a new decoder from r.
	NewDecoder(r io.Reader) decoder
}

type encoder interface {
	// Encode encodes the next value into v. It returns the number of
	// bytes that make up v's serialization, including any length
	// headers.
	Encode(v interface{}) (uint64, error)
}

type decoder interface {
	// Decode decodes the next value into v. It returns the number of
	// bytes that make up v's serialization, including any length
	// headers.
	Decode(v interface{}) (uint64, error)
}

type JSONCodec struct{}

func (JSONCodec) NewEncoder(w io.Writer) encoder {
	return &jsonEncoder{Writer: w}
}

type jsonEncoder struct{ io.Writer }

func (e *jsonEncoder) Encode(v interface{}) (uint64, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return 0, err
	}
	size := uint64(len(b))
	n := binary.Size(size)
	if err := binary.Write(e.Writer, binary.LittleEndian, size); err != nil {
		return 0, err
	}
	if _, err := e.Writer.Write(b); err != nil {
		return 0, err
	}
	return uint64(n + len(b)), nil
}

func (JSONCodec) NewDecoder(r io.Reader) decoder {
	return &jsonDecoder{Reader: r}
}

type jsonDecoder struct{ io.Reader }

func (d *jsonDecoder) Decode(v interface{}) (uint64, error) {
	var n uint64
	if err := binary.Read(d.Reader, binary.LittleEndian, &n); err != nil {
		return 0, err
	}
	return uint64(binary.Size(n)) + n, json.NewDecoder(io.LimitReader(d.Reader, int64(n))).Decode(v)
}

type ProtobufCodec struct{}

func (ProtobufCodec) NewEncoder(w io.Writer) encoder {
	return &protobufEncoder{w: w}
}

type protobufEncoder struct {
	w io.Writer

	j   encoder
	pbw pbio.Writer
}

func (e *protobufEncoder) Encode(v interface{}) (uint64, error) {
	switch v := v.(type) {
	case *unit.SourceUnit:
		if e.j == nil {
			e.j = JSONCodec{}.NewEncoder(e.w)
		}
		return e.j.Encode(v)
	default:
		if e.pbw == nil {
			e.pbw = pbio.NewDelimitedWriter(e.w)
		}
		return e.pbw.WriteMsg(v.(proto.Message))
	}
}

func (ProtobufCodec) NewDecoder(r io.Reader) decoder {
	return &protobufDecoder{r: r}
}

type protobufDecoder struct {
	r io.Reader

	j   decoder
	pbr pbio.Reader
}

const decodeBufSize = 4096

func (d *protobufDecoder) Decode(v interface{}) (uint64, error) {
	switch v := v.(type) {
	case *unit.SourceUnit:
		if d.j == nil {
			d.j = JSONCodec{}.NewDecoder(d.r)
		}
		return d.j.Decode(v)
	default:
		if d.pbr == nil {
			d.pbr = pbio.NewDelimitedReader(d.r, decodeBufSize, 2*1024*1024)
		}
		return d.pbr.ReadMsg(v.(proto.Message))
	}
}
