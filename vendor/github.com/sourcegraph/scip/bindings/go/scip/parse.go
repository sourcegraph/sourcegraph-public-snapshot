package scip

import (
	"io"

	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// IndexVisitor is a struct of functions rather than an interface since Go
// doesn't support adding new functions to interfaces with default
// implementations, so adding new functions here for new fields in
// the SCIP schema would break clients. Individual functions may be nil.
type IndexVisitor struct {
	VisitMetadata       func(*Metadata)
	VisitDocument       func(*Document)
	VisitExternalSymbol func(*SymbolInformation)
}

// See https://protobuf.dev/programming-guides/encoding/#varints
const maxVarintBytes = 10

// ParseStreaming processes an index by incrementally reading input from the io.Reader.
//
// Parsing takes place at Document granularity for ease of use.
func (pi *IndexVisitor) ParseStreaming(r io.Reader) error {
	// The tag is encoded as a varint with value: (field_number << 3) | wire_type
	// Varints < 128 fit in 1 byte, which means 4 bits are available for field
	// numbers. The Index type has less than 15 fields, so the tag will fit in 1 byte.
	tagBuf := make([]byte, 1)
	lenBuf := make([]byte, 0, maxVarintBytes)
	dataBuf := make([]byte, 0, 1024)

	for {
		numRead, err := r.Read(tagBuf)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return errors.Wrapf(err, "failed to read from index reader: %w")
		}
		if numRead == 0 {
			return errors.New("read 0 bytes from index")
		}
		fieldNumber, fieldType, errCode := protowire.ConsumeTag(tagBuf)
		if errCode < 0 {
			return errors.Wrap(protowire.ParseError(errCode), "failed to consume tag")
		}
		switch fieldNumber {
		// As per scip.proto, all of Metadata, Document and SymbolInformation are sub-messages
		case metadataFieldNumber, documentsFieldNumber, externalSymbolsFieldNumber:
			if fieldType != protowire.BytesType {
				return errors.Newf("expected LEN type tag for %s", indexFieldName(fieldNumber))
			}
			lenBuf = lenBuf[:0]
			dataLenUint, err := readVarint(r, &lenBuf)
			dataLen := int(dataLenUint)
			if err != nil {
				return errors.Wrapf(err, "failed to read length for %s", indexFieldName(fieldNumber))
			}
			if dataLen > cap(dataBuf) {
				dataBuf = make([]byte, dataLen)
			} else {
				dataBuf = dataBuf[:0]
				for i := 0; i < dataLen; i++ {
					dataBuf = append(dataBuf, 0)
				}
			}
			// Keep going when len == 0 instead of short-circuiting to preserve empty sub-messages
			if dataLen > 0 {
				numRead, err := io.ReadAtLeast(r, dataBuf, dataLen)
				if err != nil {
					return errors.Wrapf(err, "failed to read data for %s", indexFieldName(fieldNumber))
				}
				if numRead != dataLen {
					return errors.Newf(
						"expected to read %d bytes based on LEN but read %d bytes", dataLen, numRead)
				}
			}
			if fieldNumber == metadataFieldNumber {
				if pi.VisitMetadata != nil {
					m := Metadata{}
					if err := proto.Unmarshal(dataBuf, &m); err != nil {
						return errors.Wrapf(err, "failed to read %s", indexFieldName(fieldNumber))
					}
					pi.VisitMetadata(&m)
				}
			} else if fieldNumber == documentsFieldNumber {
				if pi.VisitDocument != nil {
					d := Document{}
					if err := proto.Unmarshal(dataBuf, &d); err != nil {
						return errors.Wrapf(err, "failed to read %s", indexFieldName(fieldNumber))
					}
					pi.VisitDocument(&d)
				}
			} else if fieldNumber == externalSymbolsFieldNumber {
				if pi.VisitExternalSymbol != nil {
					s := SymbolInformation{}
					if err := proto.Unmarshal(dataBuf, &s); err != nil {
						return errors.Wrapf(err, "failed to read %s", indexFieldName(fieldNumber))
					}
					pi.VisitExternalSymbol(&s)
				}
			} else {
				return errors.Newf(
					"added new field with number: %v in scip.Index but missing unmarshaling code", fieldNumber)
			}
		default:
			return errors.Newf(
				"added new field with number: %v in scip.Index but forgot to update streaming parser", fieldNumber)
		}
	}
}

const (
	metadataFieldNumber        = 1
	documentsFieldNumber       = 2
	externalSymbolsFieldNumber = 3
)

// readVarint attempts to read a varint, using scratchBuf for temporary storage
//
// scratchBuf should be able to accommodate any varint size
// based on its capacity, and be cleared before readVarint is called
func readVarint(r io.Reader, scratchBuf *[]byte) (uint64, error) {
	nextByteBuf := make([]byte, 1, 1)
	for i := 0; i < cap(*scratchBuf); i++ {
		numRead, err := r.Read(nextByteBuf)
		if err != nil {
			return 0, errors.Wrapf(err, "failed to read %d-th byte of Varint. soFar: %v", i, scratchBuf)
		}
		if numRead == 0 {
			return 0, errors.Newf("failed to read %d-th byte of Varint. soFar: %v", scratchBuf)
		}
		nextByte := nextByteBuf[0]
		*scratchBuf = append(*scratchBuf, nextByte)
		if nextByte <= 127 { // https://protobuf.dev/programming-guides/encoding/#varints
			// Continuation bit is not set, so Varint must've ended
			break
		}
	}
	value, errCode := protowire.ConsumeVarint(*scratchBuf)
	if errCode < 0 {
		return value, protowire.ParseError(errCode)
	}
	return value, nil
}

func indexFieldName(i protowire.Number) string {
	if i == metadataFieldNumber {
		return "metadata"
	} else if i == documentsFieldNumber {
		return "documents"
	} else if i == externalSymbolsFieldNumber {
		return "external_symbols"
	}
	return "<unknown>"
}
