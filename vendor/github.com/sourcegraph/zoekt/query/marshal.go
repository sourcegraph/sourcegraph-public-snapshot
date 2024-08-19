package query

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"unsafe"

	"github.com/RoaringBitmap/roaring"
)

func branchesReposEncode(brs []BranchRepos) ([]byte, error) {
	var b bytes.Buffer
	var enc [binary.MaxVarintLen64]byte
	varint := func(n uint64) {
		m := binary.PutUvarint(enc[:], n)
		b.Write(enc[:m])
	}
	str := func(s string) {
		varint(uint64(len(s)))
		b.WriteString(s)
	}
	strSize := func(s string) uint64 {
		return uint64(binary.PutUvarint(enc[:], uint64(len(s))) + len(s))
	}

	// Calculate size
	size := uint64(1) // version
	size += uint64(binary.PutUvarint(enc[:], uint64(len(brs))))
	for _, br := range brs {
		size += strSize(br.Branch)
		idsSize := br.Repos.GetSerializedSizeInBytes()
		size += uint64(binary.PutUvarint(enc[:], idsSize))
		size += idsSize
	}

	b.Grow(int(size))

	// Version
	b.WriteByte(1)

	// Length
	varint(uint64(len(brs)))

	for _, br := range brs {
		str(br.Branch)
		l := br.Repos.GetSerializedSizeInBytes()
		varint(l)

		n, err := br.Repos.WriteTo(&b)
		if err != nil {
			return nil, err
		}

		if uint64(n) != l {
			return nil, io.ErrShortWrite
		}
	}

	return b.Bytes(), nil
}

func branchesReposDecode(b []byte) ([]BranchRepos, error) {
	// binaryReader returns strings pointing into b to avoid allocations. We
	// don't own b, so we create a copy of it.
	r := binaryReader{b: append(make([]byte, 0, len(b)), b...)}

	// Version
	if v := r.byt(); v != 1 {
		return nil, fmt.Errorf("unsupported BranchRepos encoding version %d", v)
	}

	l := r.uvarint() // Length
	brs := make([]BranchRepos, l)

	for i := 0; i < l; i++ {
		brs[i].Branch = r.str()
		brs[i].Repos = r.bitmap()
	}

	return brs, r.err
}

// We implement a custom binary marshaller for a set of file names. See commit
// 6c893ff323647b0419fac46ee462532401bf3283 for context on this code.
// Additionally this code is based on that commit.
//
// Wire-format of map[string]struct{} is pretty straightforward:
//
// byte(1) version
// uvarint(len(map))
// for k in map:
//   uvarint(len(k))
//   bytes(k)
//
// The above format gives about the same size encoding as gob does. However,
// gob doesn't have a specialization for map[string]struct{} so we get to
// avoid a lot of intermediate allocations.
//
// The above adds up to a huge improvement, worth the extra complexity:
//
//   name                  old time/op    new time/op    delta
//   FileNameSet_Encode-8    91.2µs ± 2%    36.8µs ± 1%  -59.69%  (p=0.000 n=10+9)
//   FileNameSet_Decode-8     143µs ± 1%      54µs ± 1%  -61.96%  (p=0.000 n=8+9)
//
//   name                  old bytes      new bytes      delta
//   FileNameSet_Encode-8    12.1kB ± 0%    11.1kB ± 0%   -8.63%  (p=0.000 n=10+10)
//
//   name                  old alloc/op   new alloc/op   delta
//   FileNameSet_Encode-8    16.0kB ± 0%    12.3kB ± 0%  -23.20%  (p=0.000 n=10+10)
//   FileNameSet_Decode-8    76.7kB ± 0%    72.3kB ± 0%   -5.77%  (p=0.000 n=10+10)
//
//   name                  old allocs/op  new allocs/op  delta
//   FileNameSet_Encode-8     1.00k ± 0%     0.00k ± 0%  -99.90%  (p=0.000 n=10+10)
//   FileNameSet_Decode-8     1.20k ± 0%     0.18k ± 0%  -85.27%  (p=0.000 n=10+10)

// stringSetEncode implements an efficient encoder for map[string]struct{}.
func stringSetEncode(set map[string]struct{}) ([]byte, error) {
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

	// Calculate size
	size := 1 // version
	size += binary.PutUvarint(enc[:], uint64(len(set)))
	for k := range set {
		size += strSize(k)
	}
	b.Grow(size)

	// Version
	b.WriteByte(1)

	// Length
	varint(len(set))

	for k := range set {
		str(k)
	}

	return b.Bytes(), nil
}

// stringSetDecode implements an efficient decoder for map[string]struct{}.
func stringSetDecode(b []byte) (map[string]struct{}, error) {
	// binaryReader returns strings pointing into b to avoid allocations. We
	// don't own b, so we create a copy of it.
	r := binaryReader{b: append([]byte{}, b...)}

	// Version
	if v := r.byt(); v != 1 {
		return nil, fmt.Errorf("unsupported stringSet encoding version %d", v)
	}

	// Length
	l := r.uvarint()
	set := make(map[string]struct{}, l)

	for i := 0; i < l; i++ {
		set[r.str()] = struct{}{}
	}

	return set, r.err
}

type binaryReader struct {
	b   []byte
	err error
}

func (b *binaryReader) uvarint() int {
	x, n := binary.Uvarint(b.b)
	if n < 0 {
		b.b = nil
		b.err = errors.New("malformed RepoBranches")
		return 0
	}
	b.b = b.b[n:]
	return int(x)
}

func (b *binaryReader) str() string {
	l := b.uvarint()
	if l > len(b.b) {
		b.b = nil
		b.err = errors.New("malformed RepoBranches")
		return ""
	}
	s := b2s(b.b[:l])
	b.b = b.b[l:]
	return s
}

func (b *binaryReader) bitmap() *roaring.Bitmap {
	l := b.uvarint()
	if l > len(b.b) {
		b.b = nil
		b.err = errors.New("malformed BranchRepos")
		return nil
	}
	r := roaring.New()
	_, b.err = r.FromBuffer(b.b[:l])
	b.b = b.b[l:]
	return r
}

func (b *binaryReader) byt() byte {
	if len(b.b) < 1 {
		b.b = nil
		b.err = errors.New("malformed RepoBranches")
		return 0
	}
	x := b.b[0]
	b.b = b.b[1:]
	return x
}

func b2s(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
