package zoekt

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"unsafe"
)

// Wire-format of map[uint32]MinimalRepoListEntry is pretty straightforward:
//
// byte(2) version
// uvarint(len(minimal))
// uvarint(sum(len(entry.Branches) for entry in minimal))
// for repoID, entry in minimal:
//   uvarint(repoID)
//   byte(entry.HasSymbols)
//   uvarint(entry.IndexTimeUnix)
//   uvarint(len(entry.Branches))
//   for b in entry.Branches:
//     str(b.Name)
//     str(b.Version)
//
// Version 1 was the same, except it didn't have the IndexTimeUnix field.

// reposMapEncode implements an efficient encoder for ReposMap.
func reposMapEncode(minimal ReposMap) ([]byte, error) {
	if minimal == nil {
		return nil, nil
	}

	var b bytes.Buffer
	var enc [binary.MaxVarintLen64]byte
	varint := func(n int) {
		m := binary.PutUvarint(enc[:], uint64(n))
		b.Write(enc[:m])
	}
	str := func(s string) {
		varint(len(s))
		b.WriteString(s)
	}
	strSize := func(s string) int {
		return binary.PutUvarint(enc[:], uint64(len(s))) + len(s)
	}

	// We calculate this up front so when decoding we only need to allocate the
	// underlying array once.
	allBranchesLen := 0
	for _, entry := range minimal {
		allBranchesLen += len(entry.Branches)
	}

	// Calculate size
	size := 1 // version
	size += binary.PutUvarint(enc[:], uint64(len(minimal)))
	size += binary.PutUvarint(enc[:], uint64(allBranchesLen))
	for repoID, entry := range minimal {
		size += binary.PutUvarint(enc[:], uint64(repoID))
		size += 1 // HasSymbols
		size += binary.PutUvarint(enc[:], uint64(entry.IndexTimeUnix))
		size += binary.PutUvarint(enc[:], uint64(len(entry.Branches)))
		for _, b := range entry.Branches {
			size += strSize(b.Name)
			size += strSize(b.Version)
		}
	}
	b.Grow(size)

	// Version
	b.WriteByte(2)

	// Length
	varint(len(minimal))

	varint(allBranchesLen)

	for repoID, entry := range minimal {
		varint(int(repoID))

		hasSymbols := byte(1)
		if !entry.HasSymbols {
			hasSymbols = 0
		}
		b.WriteByte(hasSymbols)

		varint(int(entry.IndexTimeUnix))

		varint(len(entry.Branches))
		for _, b := range entry.Branches {
			str(b.Name)
			str(b.Version)
		}
	}

	return b.Bytes(), nil
}

// reposMapDecode implements an efficient decoder for map[string]struct{}.
func reposMapDecode(b []byte) (ReposMap, error) {
	// nil input
	if len(b) == 0 {
		return nil, nil
	}

	// binaryReader returns strings pointing into b to avoid allocations. We
	// don't own b, so we create a copy of it.
	r := binaryReader{
		typ: "ReposMap",
		b:   append([]byte{}, b...),
	}

	// Version
	var readIndexTime bool
	v := r.byt()
	switch v {
	case 1:
		readIndexTime = false
	case 2:
		readIndexTime = true
	default:
		return nil, fmt.Errorf("unsupported stringSet encoding version %d", v)
	}

	// Length
	l := r.uvarint()
	m := make(map[uint32]MinimalRepoListEntry, l)

	// Pre-allocate slice for all branches
	allBranchesLen := r.uvarint()
	allBranches := make([]RepositoryBranch, 0, allBranchesLen)

	for i := 0; i < l; i++ {
		repoID := r.uvarint()
		hasSymbols := r.byt() == 1
		var indexTimeUnix int64
		if readIndexTime {
			indexTimeUnix = int64(r.uvarint())
		}
		lb := r.uvarint()
		for i := 0; i < lb; i++ {
			allBranches = append(allBranches, RepositoryBranch{
				Name:    r.str(),
				Version: r.str(),
			})
		}
		branches := allBranches[len(allBranches)-lb:]
		m[uint32(repoID)] = MinimalRepoListEntry{
			HasSymbols:    hasSymbols,
			Branches:      branches,
			IndexTimeUnix: indexTimeUnix,
		}
	}

	return m, r.err
}

type binaryReader struct {
	typ string
	b   []byte
	err error
}

func (b *binaryReader) uvarint() int {
	x, n := binary.Uvarint(b.b)
	if n < 0 {
		b.b = nil
		b.err = fmt.Errorf("malformed %s", b.typ)
		return 0
	}
	b.b = b.b[n:]
	return int(x)
}

func (b *binaryReader) str() string {
	l := b.uvarint()
	if l > len(b.b) {
		b.b = nil
		b.err = fmt.Errorf("malformed %s", b.typ)
		return ""
	}
	s := b2s(b.b[:l])
	b.b = b.b[l:]
	return s
}

func (b *binaryReader) byt() byte {
	if len(b.b) < 1 {
		b.b = nil
		b.err = fmt.Errorf("malformed %s", b.typ)
		return 0
	}
	x := b.b[0]
	b.b = b.b[1:]
	return x
}

func b2s(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
