// Extensions for Protocol Buffers to create more go like structures.
//
// Copyright (c) 2013, Vastech SA (PTY) LTD. All rights reserved.
// http://github.com/gogo/protobuf/gogoproto
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:
//
//     * Redistributions of source code must retain the above copyright
// notice, this list of conditions and the following disclaimer.
//     * Redistributions in binary form must reproduce the above
// copyright notice, this list of conditions and the following disclaimer
// in the documentation and/or other materials provided with the
// distribution.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
// A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
// OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
// LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package pbio

import (
	"bufio"
	"encoding/binary"
	"errors"
	"io"

	"github.com/gogo/protobuf/proto"
)

var (
	errLargeValue = errors.New("value is larger than 64 bits")
)

func NewDelimitedWriter(w io.Writer) Writer {
	return &varintWriter{w, make([]byte, 10), nil}
}

type varintWriter struct {
	w      io.Writer
	lenBuf []byte
	buf    []byte
}

func (w *varintWriter) WriteMsg(msg proto.Message) (uint64, error) {
	var data []byte
	if m, ok := msg.(marshaler); ok {
		n := m.Size()
		if n >= len(w.buf) {
			w.buf = make([]byte, n)
		}
		if _, err := m.MarshalTo(w.buf); err != nil {
			return 0, err
		}
		data = w.buf[:n]
	} else {
		var err error
		data, err = proto.Marshal(msg)
		if err != nil {
			return 0, err
		}
	}
	length := uint64(len(data))
	n := binary.PutUvarint(w.lenBuf, length)
	if _, err := w.w.Write(w.lenBuf[:n]); err != nil {
		return 0, err
	}
	_, err := w.w.Write(data)
	return uint64(n) + length, err
}

func NewDelimitedReader(r io.Reader, bufSize, maxSize int) Reader {
	return &varintReader{bufio.NewReaderSize(r, bufSize), nil, maxSize}
}

type varintReader struct {
	r       *bufio.Reader
	buf     []byte
	maxSize int
}

func (r *varintReader) ReadMsg(msg proto.Message) (uint64, error) {
	n, length64, err := readUvarint(r.r)
	if err != nil {
		return 0, err
	}
	length := int(length64)
	if length < 0 || length > r.maxSize {
		return 0, io.ErrShortBuffer
	}
	if len(r.buf) < length {
		r.buf = make([]byte, length)
	}
	buf := r.buf[:length]
	if _, err := io.ReadFull(r.r, buf); err != nil {
		return 0, err
	}
	return uint64(n) + length64, proto.Unmarshal(buf, msg)
}

// readUvarint reads an encoded unsigned integer from r and returns it
// as a uint64. It returns the int number of bytes read.
//
// It is adapted from Go's encoding/binary package and modified to
// return the number of bytes read.
func readUvarint(r io.ByteReader) (int, uint64, error) {
	var x uint64
	var s uint
	for i := 1; ; i++ {
		b, err := r.ReadByte()
		if err != nil {
			return i, x, err
		}
		if b < 0x80 {
			if i > 9 || i == 9 && b > 1 {
				return i, x, errors.New("binary: varint overflows a 64-bit integer")
			}
			return i, x | uint64(b)<<s, nil
		}
		x |= uint64(b&0x7f) << s
		s += 7
	}
}
