package freecache

import (
	"bytes"
	"errors"
	"fmt"
	"io"
)

var ErrOutOfRange = errors.New("out of range")

// Ring buffer has a fixed size, when data exceeds the
// size, old data will be overwritten by new data.
// It only contains the data in the stream from begin to end
type RingBuf struct {
	begin int64 // beginning offset of the data stream.
	end   int64 // ending offset of the data stream.
	data  []byte
	index int //range from '0' to 'len(rb.data)-1'
}

func NewRingBuf(size int, begin int64) (rb RingBuf) {
	rb.data = make([]byte, size)
	rb.begin = begin
	rb.end = begin
	rb.index = 0
	return
}

// Create a copy of the buffer.
func (rb *RingBuf) Dump() []byte {
	dump := make([]byte, len(rb.data))
	copy(dump, rb.data)
	return dump
}

func (rb *RingBuf) String() string {
	return fmt.Sprintf("[size:%v, start:%v, end:%v, index:%v]", len(rb.data), rb.begin, rb.end, rb.index)
}

func (rb *RingBuf) Size() int64 {
	return int64(len(rb.data))
}

func (rb *RingBuf) Begin() int64 {
	return rb.begin
}

func (rb *RingBuf) End() int64 {
	return rb.end
}

// read up to len(p), at off of the data stream.
func (rb *RingBuf) ReadAt(p []byte, off int64) (n int, err error) {
	if off > rb.end || off < rb.begin {
		err = ErrOutOfRange
		return
	}
	var readOff int
	if rb.end-rb.begin < int64(len(rb.data)) {
		readOff = int(off - rb.begin)
	} else {
		readOff = rb.index + int(off-rb.begin)
	}
	if readOff >= len(rb.data) {
		readOff -= len(rb.data)
	}
	readEnd := readOff + int(rb.end-off)
	if readEnd <= len(rb.data) {
		n = copy(p, rb.data[readOff:readEnd])
	} else {
		n = copy(p, rb.data[readOff:])
		if n < len(p) {
			n += copy(p[n:], rb.data[:readEnd-len(rb.data)])
		}
	}
	if n < len(p) {
		err = io.EOF
	}
	return
}

func (rb *RingBuf) Write(p []byte) (n int, err error) {
	if len(p) > len(rb.data) {
		err = ErrOutOfRange
		return
	}
	for n < len(p) {
		written := copy(rb.data[rb.index:], p[n:])
		rb.end += int64(written)
		n += written
		rb.index += written
		if rb.index >= len(rb.data) {
			rb.index -= len(rb.data)
		}
	}
	if int(rb.end-rb.begin) > len(rb.data) {
		rb.begin = rb.end - int64(len(rb.data))
	}
	return
}

func (rb *RingBuf) WriteAt(p []byte, off int64) (n int, err error) {
	if off+int64(len(p)) > rb.end || off < rb.begin {
		err = ErrOutOfRange
		return
	}
	var writeOff int
	if rb.end-rb.begin < int64(len(rb.data)) {
		writeOff = int(off - rb.begin)
	} else {
		writeOff = rb.index + int(off-rb.begin)
	}
	if writeOff > len(rb.data) {
		writeOff -= len(rb.data)
	}
	writeEnd := writeOff + int(rb.end-off)
	if writeEnd <= len(rb.data) {
		n = copy(rb.data[writeOff:writeEnd], p)
	} else {
		n = copy(rb.data[writeOff:], p)
		if n < len(p) {
			n += copy(rb.data[:writeEnd-len(rb.data)], p[n:])
		}
	}
	return
}

func (rb *RingBuf) EqualAt(p []byte, off int64) bool {
	if off+int64(len(p)) > rb.end || off < rb.begin {
		return false
	}
	var readOff int
	if rb.end-rb.begin < int64(len(rb.data)) {
		readOff = int(off - rb.begin)
	} else {
		readOff = rb.index + int(off-rb.begin)
	}
	if readOff >= len(rb.data) {
		readOff -= len(rb.data)
	}
	readEnd := readOff + len(p)
	if readEnd <= len(rb.data) {
		return bytes.Equal(p, rb.data[readOff:readEnd])
	} else {
		firstLen := len(rb.data) - readOff
		equal := bytes.Equal(p[:firstLen], rb.data[readOff:])
		if equal {
			secondLen := len(p) - firstLen
			equal = bytes.Equal(p[firstLen:], rb.data[:secondLen])
		}
		return equal
	}
}

// Evacuate read the data at off, then write it to the the data stream,
// Keep it from being overwritten by new data.
func (rb *RingBuf) Evacuate(off int64, length int) (newOff int64) {
	if off+int64(length) > rb.end || off < rb.begin {
		return -1
	}
	var readOff int
	if rb.end-rb.begin < int64(len(rb.data)) {
		readOff = int(off - rb.begin)
	} else {
		readOff = rb.index + int(off-rb.begin)
	}
	if readOff >= len(rb.data) {
		readOff -= len(rb.data)
	}

	if readOff == rb.index {
		// no copy evacuate
		rb.index += length
		if rb.index >= len(rb.data) {
			rb.index -= len(rb.data)
		}
	} else if readOff < rb.index {
		var n = copy(rb.data[rb.index:], rb.data[readOff:readOff+length])
		rb.index += n
		if rb.index == len(rb.data) {
			rb.index = copy(rb.data, rb.data[readOff+n:readOff+length])
		}
	} else {
		var readEnd = readOff + length
		var n int
		if readEnd <= len(rb.data) {
			n = copy(rb.data[rb.index:], rb.data[readOff:readEnd])
			rb.index += n
			if rb.index == len(rb.data) {
				rb.index = copy(rb.data, rb.data[readOff+n:readEnd])
			}
		} else {
			n = copy(rb.data[rb.index:], rb.data[readOff:])
			rb.index += n
			var tail = length - n
			n = copy(rb.data[rb.index:], rb.data[:tail])
			rb.index += n
			if rb.index == len(rb.data) {
				rb.index = copy(rb.data, rb.data[n:tail])
			}
		}
	}
	newOff = rb.end
	rb.end += int64(length)
	if rb.begin < rb.end-int64(len(rb.data)) {
		rb.begin = rb.end - int64(len(rb.data))
	}
	return
}

func (rb *RingBuf) Resize(newSize int) {
	if len(rb.data) == newSize {
		return
	}
	newData := make([]byte, newSize)
	var offset int
	if rb.end-rb.begin == int64(len(rb.data)) {
		offset = rb.index
	}
	if int(rb.end-rb.begin) > newSize {
		discard := int(rb.end-rb.begin) - newSize
		offset = (offset + discard) % len(rb.data)
		rb.begin = rb.end - int64(newSize)
	}
	n := copy(newData, rb.data[offset:])
	if n < newSize {
		copy(newData[n:], rb.data[:offset])
	}
	rb.data = newData
	rb.index = 0
}

func (rb *RingBuf) Skip(length int64) {
	rb.end += length
	rb.index += int(length)
	for rb.index >= len(rb.data) {
		rb.index -= len(rb.data)
	}
	if int(rb.end-rb.begin) > len(rb.data) {
		rb.begin = rb.end - int64(len(rb.data))
	}
}
