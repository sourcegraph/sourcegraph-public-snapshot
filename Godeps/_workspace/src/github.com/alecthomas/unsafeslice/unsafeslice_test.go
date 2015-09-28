package unsafeslice

import (
	"bytes"
	"encoding/binary"
	"github.com/stretchrcom/testify/assert"
	"testing"
)

func TestUnsafeSlice64(t *testing.T) {
	w := &bytes.Buffer{}
	d := []uint64{0xdead, 0xbeef, 0xb334}
	binary.Write(w, binary.LittleEndian, d)
	v := Uint64SliceFromByteSlice(w.Bytes())
	assert.Equal(t, d, v)
}

func TestUnsafeSlice32(t *testing.T) {
	w := &bytes.Buffer{}
	d := []uint32{0xdead, 0xbeef, 0xb334}
	binary.Write(w, binary.LittleEndian, d)
	v := Uint32SliceFromByteSlice(w.Bytes())
	assert.Equal(t, d, v)
}

func TestUnsafeSlice16(t *testing.T) {
	w := &bytes.Buffer{}
	d := []uint16{0xdead, 0xbeef, 0xb334}
	binary.Write(w, binary.LittleEndian, d)
	v := Uint16SliceFromByteSlice(w.Bytes())
	assert.Equal(t, d, v)
}

type Struct struct {
	A uint8
	B uint32
}

func makeTestStructBuffer() []byte {
	w := &bytes.Buffer{}
	a := &Struct{0xab, 0xdead}
	b := &Struct{0xce, 0xbeef}
	// Write struct values with padding
	binary.Write(w, binary.LittleEndian, a.A)
	w.Write([]byte{0, 0, 0})
	binary.Write(w, binary.LittleEndian, a.B)
	binary.Write(w, binary.LittleEndian, b.A)
	w.Write([]byte{0, 0, 0})
	binary.Write(w, binary.LittleEndian, b.B)
	return w.Bytes()
}

func TestUnsafeSliceStruct(t *testing.T) {
	var v []Struct
	b := makeTestStructBuffer()
	assert.Nil(t, v)
	StructSliceFromByteSlice(b, &v)
	assert.NotNil(t, v)
	assert.Equal(t, len(v), 2)
	assert.Equal(t, v[0].A, uint8(0xab))
	assert.Equal(t, v[0].B, uint32(0xdead))
	assert.Equal(t, v[1].A, uint8(0xce))
	assert.Equal(t, v[1].B, uint32(0xbeef))
}

func TestByteSliceFromStructSlice(t *testing.T) {
	a := []Struct{
		Struct{0xab, 0xdead},
		Struct{0xce, 0xbeef},
	}
	b := ByteSliceFromStructSlice(a)
	assert.Equal(t, 16, len(b))
	assert.Equal(t, makeTestStructBuffer(), b)
	assert.True(t, bytes.Compare(makeTestStructBuffer(), b) == 0)

	b = ByteSliceFromStructSlice([]Struct{})
	assert.Equal(t, len(b), 0)
}
