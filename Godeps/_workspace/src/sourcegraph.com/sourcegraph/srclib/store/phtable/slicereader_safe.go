// +build !386,!amd64,!arm

package phtable

import (
	"encoding/binary"
	"fmt"
)

// Read values and typed vectors from a byte slice without copying where possible.
type sliceReader struct {
	b          []byte
	start, end uint64
}

func (b *sliceReader) Read(size uint64) []byte {
	b.start, b.end = b.end, b.end+size
	if b.start == b.end {
		return nil
	}
	return b.b[b.start:b.end]
}

func (b *sliceReader) ReadUint64Array(n uint64) []uint64 {
	b.start, b.end = b.end, b.end+n*8
	out := make([]uint64, n)
	buf := b.b[b.start:b.end]
	for i := 0; i < len(buf); i += 8 {
		out[i>>3] = binary.LittleEndian.Uint64(buf[i : i+8])
	}
	return out
}

func (b *sliceReader) ReadUint16Array(n uint64) []uint16 {
	b.start, b.end = b.end, b.end+n*2
	out := make([]uint16, n)
	buf := b.b[b.start:b.end]
	for i := 0; i < len(buf); i += 2 {
		out[i>>1] = binary.LittleEndian.Uint16(buf[i : i+2])
	}
	return out
}

func (b *sliceReader) ReadInt() uint64 {
	return uint64(binary.LittleEndian.Uint32(b.Read(4)))
}

func (b *sliceReader) ReadInt32() uint32 {
	return binary.LittleEndian.Uint32(b.Read(4))
}

func (b *sliceReader) ReadUvarint() uint64 {
	v, n := binary.Uvarint(b.b[b.end:])
	if n <= 0 {
		panic(fmt.Sprintf("Uvarint error: n == %d (buf size %d)", n, b.end-b.start))
	}
	b.start, b.end = b.end, b.end+uint64(n)
	return v
}
