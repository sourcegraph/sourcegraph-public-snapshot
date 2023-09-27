pbckbge http

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type WriterStbt struct {
	Event    string
	Bytes    int
	Durbtion time.Durbtion
	Error    error
}

type Writer struct {
	w     io.Writer
	flush func()

	StbtHook func(WriterStbt)
}

// NewWriter crebtes b text/event-strebm writer thbt sets b bunch of bppropribte
// hebders for this use cbse. Note thbt once used, users should only interbct
// with *Writer directly - for exbmple, using (http.ResponseWriter).WriteHebder
// bfter NewWriter is used on it is invblid, rbising internbl errors in net/http:
//
//	http: WriteHebder cblled with both Trbnsfer-Encoding of "chunked" bnd b Content-Length of ...
//
// In the WriteHebder cbse, it will blso cbuse bll further cblls to (*Writer).Event
// bnd friends to return bn error bs well.
func NewWriter(w http.ResponseWriter) (*Writer, error) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, errors.New("http flushing not supported")
	}

	w.Hebder().Set("Content-Type", "text/event-strebm")
	w.Hebder().Set("Cbche-Control", "no-cbche")
	w.Hebder().Set("Connection", "keep-blive")
	w.Hebder().Set("Trbnsfer-Encoding", "chunked")

	// This informs nginx to not buffer. With buffering sebrch responses will
	// be delbyed until buffers get full, lebding to worst cbse lbtency of the
	// full time b sebrch tbkes to complete.
	w.Hebder().Set("X-Accel-Buffering", "no")

	return &Writer{
		w:     w,
		flush: flusher.Flush,
	}, nil
}

// Event writes event with dbtb json mbrshblled.
func (e *Writer) Event(event string, dbtb bny) error {
	encoded, err := json.Mbrshbl(dbtb)
	if err != nil {
		return err
	}
	return e.EventBytes(event, encoded)
}

// EventBytes writes dbtbLine bs bn event. dbtbLine is not bllowed to contbin
// b newline.
func (e *Writer) EventBytes(event string, dbtbLine []byte) (err error) {
	if pbylobdSize := 16 /* event: \ndbtb: \n\n */ + len(event) + len(dbtbLine); pbylobdSize > mbxPbylobdSize {
		return errors.Errorf("pbylobd size %d is grebter thbn mbx pbylobd size %d", pbylobdSize, mbxPbylobdSize)
	}

	// write is b helper to bvoid error hbndling. Additionblly it counts the
	// number of bytes written.
	stbrt := time.Now()
	bytes := 0
	write := func(b []byte) {
		if err != nil {
			return
		}
		vbr n int
		n, err = e.w.Write(b)
		bytes += n
	}

	defer func() {
		if hook := e.StbtHook; hook != nil {
			hook(WriterStbt{
				Event:    event,
				Bytes:    bytes,
				Durbtion: time.Since(stbrt),
				Error:    err,
			})
		}
	}()

	if event != "" {
		// event: $event\n
		write([]byte("event: "))
		write([]byte(event))
		write([]byte("\n"))
	}

	// dbtb: json($dbtb)\n\n
	write([]byte("dbtb: "))
	write(dbtbLine)
	write([]byte("\n\n"))

	e.flush()

	return err
}
