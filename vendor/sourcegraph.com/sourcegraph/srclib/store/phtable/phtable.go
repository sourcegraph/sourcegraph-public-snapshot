package phtable

import (
	"bytes"
	"encoding/binary"
	"io"
	"io/ioutil"
)

// CHD hash table lookup.
type CHD struct {
	// Random hash function table.
	r []uint64
	// Array of indices into hash function table r. We assume there aren't
	// more than 2^16 hash functions O_o
	indices []uint16
	// Final table of values.
	keys   [][]byte
	values [][]byte

	el uint32

	StoreKeys        bool
	ValuesAreVarints bool
	valueVarints     []uint64
}

func hasher(data []byte) uint64 {
	var hash uint64 = 14695981039346656037
	for _, c := range data {
		hash ^= uint64(c)
		hash *= 1099511628211
	}
	return hash
}

// Read a serialized CHD.
func Read(r io.Reader) (*CHD, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return Mmap(b, false)
}

func ReadVarints(r io.Reader) (*CHD, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return Mmap(b, true)
}

// Mmap creates a new CHD aliasing the CHD structure over an existing byte region (typically mmapped).
func Mmap(b []byte, isVarints bool) (*CHD, error) {
	c := &CHD{ValuesAreVarints: isVarints}

	bi := &sliceReader{b: b}

	// Read vector of hash functions.
	rl := bi.ReadInt()
	c.r = bi.ReadUint64Array(rl)

	// Read hash function indices.
	il := bi.ReadInt()
	c.indices = bi.ReadUint16Array(il)

	c.el = bi.ReadInt32()
	c.StoreKeys = bi.ReadInt() != 0

	if c.StoreKeys {
		c.keys = make([][]byte, c.el)
	}
	if c.ValuesAreVarints {
		c.valueVarints = make([]uint64, c.el)
	} else {
		c.values = make([][]byte, c.el)
	}

	for i := uint32(0); i < c.el; i++ {
		if c.StoreKeys {
			kl := bi.ReadUvarint()
			c.keys[i] = bi.Read(kl)
		}
		vl := bi.ReadUvarint()
		if c.ValuesAreVarints {
			c.valueVarints[i] = vl
		} else {
			c.values[i] = bi.Read(vl)
		}
	}

	return c, nil
}

func (c *CHD) getIndex(key []byte) (uint64, bool) {
	r0 := c.r[0]
	h := hasher(key) ^ r0
	i := h % uint64(len(c.indices))
	ri := c.indices[i]
	// This can occur if there were unassigned slots in the hash table.
	if ri >= uint16(len(c.r)) {
		return 0, false
	}
	r := c.r[ri]
	ti := (h ^ r) % uint64(c.el)
	// fmt.Printf("r[0]=%d, h=%d, i=%d, ri=%d, r=%d, ti=%d\n", c.r[0], h, i, ri, r, ti)
	return ti, true
}

// Get an entry from the hash table.
func (c *CHD) Get(key []byte) []byte {
	if c.ValuesAreVarints {
		panic("ValuesAreVarints")
	}
	ti, found := c.getIndex(key)
	if !found {
		return nil
	}
	if c.StoreKeys {
		k := c.keys[ti]
		if bytes.Compare(k, key) != 0 {
			return nil
		}
	}
	return c.values[ti]
}

func (c *CHD) GetUint64(key []byte) (uint64, bool) {
	if !c.ValuesAreVarints {
		panic("!ValuesAreVarints")
	}
	ti, found := c.getIndex(key)
	if !found {
		return 0, false
	}
	if c.StoreKeys {
		k := c.keys[ti]
		if bytes.Compare(k, key) != 0 {
			return 0, false
		}
	}
	return c.valueVarints[ti], true
}

func (c *CHD) Len() int {
	return len(c.keys)
}

// Iterate over entries in the hash table.
func (c *CHD) Iterate() *Iterator {
	if len(c.keys) == 0 {
		return nil
	}
	return &Iterator{c: c}
}

// Serialize the CHD. The serialized form is conducive to mmapped access. See
// the Mmap function for details.
func (c *CHD) Write(w io.Writer) error {
	write := func(nd ...interface{}) error {
		for _, d := range nd {
			if err := binary.Write(w, binary.LittleEndian, d); err != nil {
				return err
			}
		}
		return nil
	}

	var storeKeys uint32
	if c.StoreKeys {
		storeKeys = 1
	}

	data := []interface{}{
		uint32(len(c.r)), c.r,
		uint32(len(c.indices)), c.indices,
		uint32(c.el),
		storeKeys,
	}

	if err := write(data...); err != nil {
		return err
	}

	vb := make([]byte, binary.MaxVarintLen64)
	for i := 0; i < int(c.el); i++ {
		if c.StoreKeys {
			k := c.keys[i]
			n := binary.PutUvarint(vb, uint64(len(k)))
			if _, err := w.Write(vb[:n]); err != nil {
				return err
			}
			if _, err := w.Write(k); err != nil {
				return err
			}
		}
		if c.ValuesAreVarints {
			v := c.valueVarints[i]
			n := binary.PutUvarint(vb, v)
			if _, err := w.Write(vb[:n]); err != nil {
				return err
			}
		} else {
			v := c.values[i]
			n := binary.PutUvarint(vb, uint64(len(v)))
			if _, err := w.Write(vb[:n]); err != nil {
				return err
			}
			if _, err := w.Write(v); err != nil {
				return err
			}
		}
	}
	return nil
}

type Iterator struct {
	i int
	c *CHD
}

func (c *Iterator) Get() (key []byte, value []byte) {
	return c.c.keys[c.i], c.c.values[c.i]
}

func (c *Iterator) Next() *Iterator {
	c.i++
	if c.i >= len(c.c.keys) {
		return nil
	}
	return c
}
