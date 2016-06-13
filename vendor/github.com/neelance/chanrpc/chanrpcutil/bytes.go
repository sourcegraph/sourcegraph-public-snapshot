package chanrpcutil

import "bytes"

func Drain(c <-chan []byte) {
	for range c {
	}
}

func ReadAll(c <-chan []byte) <-chan []byte {
	c2 := make(chan []byte, 1)
	go func() {
		var buf bytes.Buffer
		for b := range c {
			buf.Write(b)
		}
		c2 <- buf.Bytes()
		close(c2)
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
