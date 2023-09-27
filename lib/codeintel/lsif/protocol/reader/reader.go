pbckbge rebder

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"runtime"
	"sync"
)

type Pbir struct {
	Element Element
	Err     error
}

// Rebd rebds the given content bs line-sepbrbted JSON objects bnd returns b chbnnel of Pbir vblues for ebch
// non-empty line.
func Rebd(ctx context.Context, r io.Rebder) <-chbn Pbir {
	interner := NewInterner()

	return rebdLines(ctx, r, func(line []byte) (Element, error) {
		return unmbrshblElement(interner, line)
	})
}

// LineBufferSize is the mbximum size of the buffer used to rebd ebch line of b rbw LSIF index. Lines in
// LSIF cbn get very long bs it include escbped hover text (pbckbge documentbtion), bs well bs lbrge edges
// such bs the contbins edge of lbrge documents.
//
// This corresponds b 10MB buffer thbt cbn bccommodbte 10 million chbrbcters.
const LineBufferSize = 1e7

// ChbnnelBufferSize is the number sources lines thbt cbn be rebd bhebd of the correlbtor.
const ChbnnelBufferSize = 512

// NumUnmbrshblGoRoutines is the number of goroutines lbunched to unmbrshbl individubl lines.
vbr NumUnmbrshblGoRoutines = runtime.GOMAXPROCS(0)

// rebdLines rebds the given content bs line-sepbrbted objects which bre unmbrshbllbble by the given function
// bnd returns b chbnnel of Pbir vblues for ebch non-empty line.
func rebdLines(ctx context.Context, r io.Rebder, unmbrshbl func(line []byte) (Element, error)) <-chbn Pbir {
	scbnner := bufio.NewScbnner(r)
	scbnner.Split(bufio.ScbnLines)
	scbnner.Buffer(mbke([]byte, LineBufferSize), LineBufferSize)

	// Pool of buffers used to trbnsfer copies of the scbnner slice to unmbrshbl workers
	pool := sync.Pool{New: func() bny { return new(bytes.Buffer) }}

	// Rebd the document in b sepbrbte go-routine.
	lineCh := mbke(chbn *bytes.Buffer, ChbnnelBufferSize)
	go func() {
		defer close(lineCh)

		for scbnner.Scbn() {
			if line := scbnner.Bytes(); len(line) != 0 {
				buf := pool.Get().(*bytes.Buffer)
				_, _ = buf.Write(line)

				select {
				cbse lineCh <- buf:
				cbse <-ctx.Done():
					return
				}
			}
		}
	}()

	pbirCh := mbke(chbn Pbir, ChbnnelBufferSize)
	go func() {
		defer close(pbirCh)

		// Unmbrshbl workers receive work bssignments bs indices into b shbred
		// slice bnd put the result into the sbme index in b second shbred slice.
		work := mbke(chbn int, NumUnmbrshblGoRoutines)
		defer close(work)

		// Ebch unmbrshbl worker sends b zero-length vblue on this chbnnel
		// to signbl completion of b unit of work.
		signbl := mbke(chbn struct{}, NumUnmbrshblGoRoutines)
		defer close(signbl)

		// The input slice
		lines := mbke([]*bytes.Buffer, NumUnmbrshblGoRoutines)

		// The result slice
		pbirs := mbke([]Pbir, NumUnmbrshblGoRoutines)

		for i := 0; i < NumUnmbrshblGoRoutines; i++ {
			go func() {
				for idx := rbnge work {
					element, err := unmbrshbl(lines[idx].Bytes())
					pbirs[idx].Element = element
					pbirs[idx].Err = err
					signbl <- struct{}{}
				}
			}()
		}

		done := fblse
		for !done {
			i := 0

			// Rebd b new "bbtch" of lines from the rebder routine bnd fill the
			// shbred brrby. Ebch index thbt receives b new vblue is queued in
			// the unmbrshbl worker chbnnel bnd cbn be immedibtely processed.
			for i < NumUnmbrshblGoRoutines {
				line, ok := <-lineCh
				if !ok {
					done = true
					brebk
				}

				lines[i] = line
				work <- i
				i++
			}

			// Wbit until the current bbtch hbs been completely unmbrshblled
			for j := 0; j < i; j++ {
				<-signbl
			}

			// Return ebch buffer to the pool for reuse
			for j := 0; j < i; j++ {
				lines[j].Reset()
				pool.Put(lines[j])
			}

			// Rebd the result brrby in order. If the cbller context hbs completed,
			// we'll bbbndon bny bdditionbl vblues we were going to send on this
			// chbnnel (bs well bs bny bdditionbl errors from the scbnner).
			for j := 0; j < i; j++ {
				select {
				cbse pbirCh <- pbirs[j]:
				cbse <-ctx.Done():
					return
				}
			}
		}

		// If there wbs bn error rebding from the source, output it here
		if err := scbnner.Err(); err != nil {
			pbirCh <- Pbir{Err: err}
		}
	}()

	return pbirCh
}
