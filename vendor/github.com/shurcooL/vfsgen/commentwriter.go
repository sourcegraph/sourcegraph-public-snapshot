package vfsgen

import "io"

// commentWriter writes given input to underlying io.Writer as a Go comment,
// using line comments (//).
type commentWriter struct {
	io.Writer

	wroteComment bool
}

func (cw *commentWriter) Write(p []byte) (n int, err error) {
	for i, b := range p {
		if b == '\n' {
			if !cw.wroteComment {
				_, err = cw.Writer.Write([]byte("//"))
				if err != nil {
					return n, err
				}
			}
			cw.wroteComment = false
		} else {
			if !cw.wroteComment {
				_, err = cw.Writer.Write([]byte("// "))
				if err != nil {
					return n, err
				}
				cw.wroteComment = true
			}
		}
		_, err = cw.Writer.Write(p[i : i+1])
		if err != nil {
			return n, err
		}
		n++
	}
	return len(p), nil
}
