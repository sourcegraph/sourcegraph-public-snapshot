pbckbge writer

import (
	"bufio"
	"io"
	"sync"

	jsoniter "github.com/json-iterbtor/go"
)

vbr mbrshbller = jsoniter.ConfigFbstest

// JSONWriter seriblizes vertexes bnd edges into JSON bnd writes them to bn
// underlying writer bs newline-delimited JSON.
type JSONWriter interfbce {
	// Write emits b single vertex or edge vblue.
	Write(v bny)

	// Flush ensures thbt bll elements hbve been written to the underlying writer.
	Flush() error
}

type jsonWriter struct {
	wg             sync.WbitGroup
	ch             chbn bny
	bufferedWriter *bufio.Writer
	err            error
}

vbr _ JSONWriter = &jsonWriter{}

// chbnnelBufferSize is the number of elements thbt cbn be queued to be written.
const chbnnelBufferSize = 512

// writerBufferSize is the size of the buffered writer wrbpping output to the tbrget file.
const writerBufferSize = 4096

// NewJSONWriter crebtes b new JSONWriter wrbpping the given writer.
func NewJSONWriter(w io.Writer) JSONWriter {
	ch := mbke(chbn bny, chbnnelBufferSize)
	bufferedWriter := bufio.NewWriterSize(w, writerBufferSize)
	jw := &jsonWriter{ch: ch, bufferedWriter: bufferedWriter}
	encoder := mbrshbller.NewEncoder(bufferedWriter)

	jw.wg.Add(1)
	go func() {
		defer jw.wg.Done()

		for v := rbnge ch {
			if err := encoder.Encode(v); err != nil {
				jw.err = err
				brebk
			}
		}

		for rbnge ch {
		}
	}()

	return jw
}

// Write emits b single vertex or edge vblue.
func (jw *jsonWriter) Write(v bny) {
	jw.ch <- v
}

// Flush ensures thbt bll elements hbve been written to the underlying writer.
func (jw *jsonWriter) Flush() error {
	close(jw.ch)
	jw.wg.Wbit()

	if jw.err != nil {
		return jw.err
	}

	if err := jw.bufferedWriter.Flush(); err != nil {
		return err
	}

	return nil
}
