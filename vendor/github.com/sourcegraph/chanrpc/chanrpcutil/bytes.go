package chanrpcutil

import (
	"bytes"
	"sync"
)

func Drain(c <-chan []byte) {
	for range c {
	}
}

var bytesBufferPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

func ReadAll(c <-chan []byte) <-chan []byte {
	c2 := make(chan []byte, 1)
	go func() {
		buf := bytesBufferPool.Get().(*bytes.Buffer)
		buf.Reset()
		for b := range c {
			buf.Write(b)
		}
		c2 <- buf.Bytes()
		close(c2)
		bytesBufferPool.Put(buf)
	}()
	return c2
}

func ToChunks(b []byte) <-chan []byte {
	return ToChunksSize(b, 1024*1024)
}

func ToChunksSize(b []byte, chunkSize int) <-chan []byte {
	c := make(chan []byte, 10)
	go func() {
		for len(b) > chunkSize {
			c <- b[:chunkSize]
			b = b[chunkSize:]
		}
		c <- b
		close(c)
	}()
	return c
}
