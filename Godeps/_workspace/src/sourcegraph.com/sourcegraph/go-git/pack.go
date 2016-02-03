package git

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"os"
	"path/filepath"
	"sync"
)

type pack struct {
	repo              *Repository
	id                string
	indexFile         *os.File
	packFile          *os.File
	indexFileErr      error
	packFileErr       error
	openIndexFileOnce sync.Once
	openPackFileOnce  sync.Once
}

func (p *pack) indexFileReader() (io.ReaderAt, error) {
	p.openIndexFileOnce.Do(func() {
		f, err := os.Open(filepath.Join(p.repo.Path, "objects", "pack", p.id+".idx"))
		if err != nil {
			p.indexFileErr = err
			return
		}
		p.indexFile = f
		if !bytes.Equal(readBytesAt(f, 0, 4), []byte{255, 't', 'O', 'c'}) {
			p.indexFileErr = errors.New("wrong magic number")
			return
		}
		if binary.BigEndian.Uint32(readBytesAt(f, 4, 4)) != 2 {
			p.indexFileErr = errors.New("unsupported index file version")
			return
		}
	})
	return p.indexFile, p.indexFileErr
}

func (p *pack) packFileReader() (io.ReaderAt, error) {
	p.openPackFileOnce.Do(func() {
		f, err := os.Open(filepath.Join(p.repo.Path, "objects", "pack", p.id+".pack"))
		if err != nil {
			p.packFileErr = err
			return
		}
		p.packFile = f
		if !bytes.HasPrefix(readBytesAt(f, 0, 4), []byte{'P', 'A', 'C', 'K'}) {
			p.packFileErr = errors.New("pack file does not min with 'PACK'")
			return
		}
		if binary.BigEndian.Uint32(readBytesAt(f, 4, 4)) != 2 {
			p.packFileErr = errors.New("unsupported pack file version")
			return
		}
	})
	return p.packFile, p.packFileErr
}

func (p *pack) Close() (err error) {
	if p.indexFile != nil {
		if thisErr := p.indexFile.Close(); thisErr != nil && err == nil {
			err = thisErr
		}
	}
	if p.packFile != nil {
		if thisErr := p.packFile.Close(); thisErr != nil && err == nil {
			err = thisErr
		}
	}
	return
}

func (p *pack) object(id ObjectID, metaOnly bool) (*Object, error) {
	r, err := p.indexFileReader()
	if err != nil {
		return nil, err
	}

	fanoutTableStart := int64(8)
	nameTableStart := fanoutTableStart + 4*256

	numObjects := binary.BigEndian.Uint32(readBytesAt(r, fanoutTableStart+4*255, 4))

	checksumTableStart := nameTableStart + 20*int64(numObjects)
	offsetTableStart := checksumTableStart + 4*int64(numObjects)
	highOffsetTableStart := offsetTableStart + 4*int64(numObjects)
	_ = highOffsetTableStart

	firstByte := id[0]
	min := uint32(0)
	if firstByte > 0 {
		min = binary.BigEndian.Uint32(readBytesAt(r, fanoutTableStart+4*int64(firstByte-1), 4))
	}
	max := binary.BigEndian.Uint32(readBytesAt(r, fanoutTableStart+4*int64(firstByte), 4)) - 1

	index, err := binarySearch(r, nameTableStart, min, max, id)
	if err != nil {
		return nil, err
	}

	offset := uint64(binary.BigEndian.Uint32(readBytesAt(r, offsetTableStart+4*int64(index), 4)))
	if offset&(1<<31) != 0 {
		highOffsetIndex := int64(offset &^ (1 << 31))
		offset = binary.BigEndian.Uint64(readBytesAt(r, highOffsetTableStart+8*highOffsetIndex, 8))
	}

	o, err := p.objectAtOffset(offset, metaOnly)
	if err != nil {
		return nil, err
	}

	return o, nil
}

func (p *pack) objectAtOffset(offset uint64, metaOnly bool) (*Object, error) {
	r, err := p.packFileReader()
	if err != nil {
		return nil, err
	}

	br := bufio.NewReader(io.NewSectionReader(r, int64(offset), math.MaxInt64-int64(offset))) // avoid overflow

	x, err := binary.ReadUvarint(br)
	if err != nil {
		return nil, err
	}
	typ := ObjectType(x & 0x70)
	size := x&^0x7f>>3 + x&0xf

	switch typ {
	case ObjectCommit, ObjectTree, ObjectBlob, ObjectTag:
		if metaOnly {
			return &Object{typ, size, nil}, nil
		}

		data, err := readAndDecompress(br, size)
		if err != nil {
			return nil, err
		}
		return &Object{typ, size, data}, nil

	case objectOfsDelta, objectRefDelta:
		var base *Object
		switch typ {
		case objectOfsDelta:
			relOffset, err := readOffset(br)
			if err != nil {
				return nil, err
			}
			base, err = p.objectAtOffset(offset-relOffset, false)
			if err != nil {
				return nil, err
			}

		case objectRefDelta:
			id := make([]byte, 20)
			if _, err := io.ReadFull(br, id); err != nil {
				return nil, err
			}
			base, err = p.repo.object(ObjectID(id), false)
			if err != nil {
				return nil, err
			}
		}

		d, err := readAndDecompress(br, size)
		if err != nil {
			return nil, err
		}

		_, n := binary.Uvarint(d) // length of base object
		d = d[n:]
		resultObjectLength, n := binary.Uvarint(d)
		d = d[n:]
		if metaOnly {
			return &Object{base.Type, resultObjectLength, nil}, nil
		}

		data, err := applyDelta(base.Data, d, resultObjectLength)
		if err != nil {
			return nil, err
		}
		return &Object{base.Type, resultObjectLength, data}, nil

	default:
		return nil, errors.New("unexpected type")
	}
}

func readOffset(r io.ByteReader) (uint64, error) {
	var offset uint64
	for {
		b, err := r.ReadByte()
		if err != nil {
			return 0, err
		}
		offset |= uint64(b & 0x7f)
		if b&0x80 == 0 {
			return offset, nil
		}
		offset = (offset + 1) << 7
	}
}

func readBytesAt(r io.ReaderAt, offset int64, len int) []byte {
	buf := make([]byte, len)
	r.ReadAt(buf, offset) // error ignored, unexpected content needs to be handled by caller, just like a corrupted file
	return buf
}

func binarySearch(r io.ReaderAt, tableStart int64, min, max uint32, id ObjectID) (uint32, error) {
	for min <= max {
		mid := min + ((max - min) / 2) // avoid overflow
		midID := ObjectID(readBytesAt(r, tableStart+20*int64(mid), 20))
		if midID == id {
			return mid, nil
		}
		if midID < id {
			if mid == math.MaxUint32 {
				break
			}
			min = mid + 1
			continue
		}
		if mid == 0 {
			break
		}
		max = mid - 1
	}
	return 0, ObjectNotFound(id)
}

func readAndDecompress(r io.Reader, inflatedSize uint64) ([]byte, error) {
	zr, err := zlib.NewReader(r)
	if err != nil {
		return nil, err
	}
	defer zr.Close()

	buf := make([]byte, inflatedSize)
	if _, err := io.ReadFull(zr, buf); err != nil {
		return nil, err
	}
	return buf, nil
}

func applyDelta(base, delta []byte, resultLen uint64) ([]byte, error) {
	res := make([]byte, resultLen)
	insertPoint := res
	for {
		if len(delta) == 0 {
			return res, nil
		}
		opcode := delta[0]
		delta = delta[1:]

		if opcode&0x80 == 0 {
			// copy from delta
			copy(insertPoint, delta[:opcode])
			insertPoint = insertPoint[opcode:]
			delta = delta[opcode:]
			continue
		}

		// copy from base
		readNum := func(len uint) uint64 {
			var x uint64
			for i := uint(0); i < len; i++ {
				if opcode&1 != 0 {
					x |= uint64(delta[0]) << (i * 8)
					delta = delta[1:]
				}
				opcode >>= 1
			}
			return x
		}
		_ = readNum
		copyOffset := readNum(4)
		copyLength := readNum(3)
		if copyLength == 0 {
			copyLength = 1 << 16
		}
		copy(insertPoint, base[copyOffset:copyOffset+copyLength])
		insertPoint = insertPoint[copyLength:]
	}
}
