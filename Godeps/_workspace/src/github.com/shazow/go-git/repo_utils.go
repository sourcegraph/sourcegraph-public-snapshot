package git

import (
	"bytes"
	"compress/zlib"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

func readIdxFile(path string) (*idxFile, error) {
	ifile := &idxFile{}
	ifile.indexpath = path
	ifile.packpath = path[0:len(path)-3] + "pack"
	idx, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if !bytes.HasPrefix(idx, []byte{255, 't', 'O', 'c'}) {
		return nil, errors.New("Not version 2 index file")
	}
	pos := 8
	var fanout [256]uint32
	for i := 0; i < 256; i++ {
		// TODO: use range
		fanout[i] = uint32(idx[pos])<<24 + uint32(idx[pos+1])<<16 + uint32(idx[pos+2])<<8 + uint32(idx[pos+3])
		pos += 4
	}
	numObjects := int(fanout[255])
	ids := make([]sha1, numObjects)

	for i := 0; i < numObjects; i++ {
		for j := 0; j < 20; j++ {
			ids[i][j] = idx[pos+j]
		}
		pos = pos + 20
	}
	// skip crc32 and offsetValues4
	pos += 8 * numObjects

	excessLen := len(idx) - 258*4 - 28*numObjects - 40
	var offsetValues8 []uint64
	if excessLen > 0 {
		// We have an index table, so let's read it first
		offsetValues8 = make([]uint64, excessLen/8)
		for i := 0; i < excessLen/8; i++ {
			offsetValues8[i] = uint64(idx[pos])<<070 + uint64(idx[pos+1])<<060 + uint64(idx[pos+2])<<050 + uint64(idx[pos+3])<<040 + uint64(idx[pos+4])<<030 + uint64(idx[pos+5])<<020 + uint64(idx[pos+6])<<010 + uint64(idx[pos+7])
			pos = pos + 8
		}
	}
	ifile.offsetValues = make(map[sha1]uint64, numObjects)
	pos = 258*4 + 24*numObjects
	for i := 0; i < numObjects; i++ {
		offset := uint32(idx[pos])<<24 + uint32(idx[pos+1])<<16 + uint32(idx[pos+2])<<8 + uint32(idx[pos+3])
		offset32ndbit := offset & 0x80000000
		offset31bits := offset & 0x7FFFFFFF
		if offset32ndbit == 0x80000000 {
			// it's an index entry
			ifile.offsetValues[ids[i]] = offsetValues8[offset31bits]
		} else {
			ifile.offsetValues[ids[i]] = uint64(offset31bits)
		}
		pos = pos + 4
	}
	// sha1Packfile := idx[pos : pos+20]
	// sha1Index := idx[pos+21 : pos+40]
	fi, err := os.Open(ifile.packpath)
	if err != nil {
		return nil, err
	}
	defer fi.Close()

	packVersion := make([]byte, 8)
	_, err = fi.Read(packVersion)
	if err != nil {
		return nil, err
	}
	if !bytes.HasPrefix(packVersion, []byte{'P', 'A', 'C', 'K'}) {
		return nil, errors.New("Pack file does not start with 'PACK'")
	}
	ifile.packversion = uint32(packVersion[4])<<24 + uint32(packVersion[5])<<16 + uint32(packVersion[6])<<8 + uint32(packVersion[7])
	return ifile, nil
}

// If the object is stored in its own file (i.e not in a pack file),
// this function returns the full path to the object file.
// It does not test if the file exists.
func filepathFromSHA1(rootdir, sha1 string) string {
	return filepath.Join(rootdir, "objects", sha1[:2], sha1[2:])
}

// The object length in a packfile is a bit more difficult than
// just reading the bytes. The first byte has the length in its
// lowest four bits, and if bit 7 is set, it means 'more' bytes
// will follow. These are added to the »left side« of the length
func readLenInPackFile(buf []byte) (length int, advance int) {
	advance = 0
	shift := [...]byte{0, 4, 11, 18, 25, 32, 39, 46, 53, 60}
	length = int(buf[advance] & 0x0F)
	for buf[advance]&0x80 > 0 {
		advance += 1
		length += (int(buf[advance]&0x7F) << shift[advance])
	}
	advance++
	return
}

// Read from a pack file (given by path) at position offset. If this is a
// non-delta object, the (inflated) bytes are just returned, if the object
// is a deltafied-object, we have to apply the delta to base objects
// before hand.
func readObjectBytes(path string, indexfiles *map[string]*idxFile, offset uint64, sizeonly bool) (ot ObjectType, length int64, dataRc io.ReadCloser, err error) {
	offsetInt := int64(offset)
	file, err := os.Open(path)
	if err != nil {
		return
	}

	defer func() {
		if err != nil || sizeonly {
			if file != nil {
				file.Close()
			}
		}
	}()

	pos, err := file.Seek(offsetInt, os.SEEK_SET)
	if err != nil {
		return
	}

	if pos != offsetInt {
		err = errors.New("Seek went wrong")
		return
	}

	buf := make([]byte, 1024)
	n, err := file.Read(buf)
	if err != nil {
		return
	}

	if n == 0 {
		err = errors.New("Nothing read from pack file")
		return
	}

	ot = ObjectType(buf[0] & 0x70)

	l, p := readLenInPackFile(buf)
	pos = int64(p)
	length = int64(l)

	var baseObjectOffset uint64
	switch ot {
	case ObjectCommit, ObjectTree, ObjectBlob, ObjectTag:
		if sizeonly {
			// if we are only interested in the size of the object,
			// we don't need to do more expensive stuff
			return
		}

		_, err = file.Seek(offsetInt+pos, os.SEEK_SET)
		if err != nil {
			return
		}

		dataRc, err = readerDecompressed(file, length)
		if err != nil {
			return
		}
		return
		// data, err = readCompressedDataFromFile(file, offsetInt+pos, length)

	case 0x60:
		// DELTA_ENCODED object w/ offset to base
		// Read the offset first, then calculate the starting point
		// of the base object
		num := int64(buf[pos]) & 0x7f
		for buf[pos]&0x80 > 0 {
			pos = pos + 1
			num = ((num + 1) << 7) | int64(buf[pos]&0x7f)
		}
		baseObjectOffset = uint64(offsetInt - num)
		pos = pos + 1

	case 0x70:
		// DELTA_ENCODED object w/ base BINARY_OBJID
		var id sha1
		id, err = NewId(buf[pos : pos+20])
		if err != nil {
			return
		}

		pos = pos + 20

		f := (*indexfiles)[path[0:len(path)-4]+"idx"]
		var ok bool
		if baseObjectOffset, ok = f.offsetValues[id]; !ok {
			log.Fatal("not implemented yet")
			err = errors.New("base object is not exist")
			return
		}
	}

	var (
		base   []byte
		baseRc io.ReadCloser
	)
	ot, _, baseRc, err = readObjectBytes(path, indexfiles, baseObjectOffset, false)
	if err != nil {
		return
	}

	defer func() {
		baseRc.Close()
	}()

	base, err = ioutil.ReadAll(baseRc)
	if err != nil {
		return
	}

	_, err = file.Seek(offsetInt+pos, os.SEEK_SET)
	if err != nil {
		return
	}

	rc, err := readerDecompressed(file, length)
	if err != nil {
		return
	}

	zpos := 0
	// This is the length of the base object. Do we need to know it?
	_, bytesRead := readerLittleEndianBase128Number(rc)
	//log.Println(zpos, bytesRead)
	zpos += bytesRead

	resultObjectLength, bytesRead := readerLittleEndianBase128Number(rc)
	zpos += bytesRead

	if sizeonly {
		// if we are only interested in the size of the object,
		// we don't need to do more expensive stuff
		length = resultObjectLength
		return
	}

	br := &readAter{base}
	data, err := readerApplyDelta(br, rc, resultObjectLength)

	dataRc = newBufReadCloser(data)
	return
}

// Return length as integer from zero terminated string
// and the beginning of the real object
func getLengthZeroTerminated(b []byte) (int64, int64) {
	i := 0
	var pos int
	for b[i] != 0 {
		i++
	}
	pos = i
	i--
	var length int64
	var pow int64
	pow = 1
	for i >= 0 {
		length = length + (int64(b[i])-48)*pow
		pow = pow * 10
		i--
	}
	return length, int64(pos) + 1
}

// Read the contents of the object file at path.
// Return the content type, the contents of the file and error, if any
func readObjectFile(path string, sizeonly bool) (ot ObjectType, length int64, dataRc io.ReadCloser, err error) {
	f, err := os.Open(path)
	if err != nil {
		return
	}

	defer func() {
		if err != nil || sizeonly {
			if f != nil {
				f.Close()
			}
		}
	}()

	r, err := zlib.NewReader(f)
	if err != nil {
		return
	}

	firstBufferSize := int64(1024)

	buf := make([]byte, firstBufferSize)
	_, err = r.Read(buf)
	if err != nil {
		return
	}

	spacePos := int64(bytes.IndexByte(buf, ' '))

	// "tree", "commit", "blob", ...
	switch string(buf[:spacePos]) {
	case "blob":
		ot = ObjectBlob
	case "tree":
		ot = ObjectTree
	case "commit":
		ot = ObjectCommit
	case "tag":
		ot = ObjectTag
	}

	// length starts at the position after the space
	var objstart int64
	length, objstart = getLengthZeroTerminated(buf[spacePos+1:])

	if sizeonly {
		return
	}

	objstart += spacePos + 1

	_, err = f.Seek(0, os.SEEK_SET)
	if err != nil {
		return
	}

	rc, err := readerDecompressed(f, length+objstart)
	if err != nil {
		return
	}

	_, err = io.Copy(ioutil.Discard, io.LimitReader(rc, objstart))
	if err != nil {
		return
	}

	dataRc = newReadCloser(io.LimitReader(rc, length), rc)
	return
}
