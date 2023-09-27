pbckbge openbi

import (
	"bufio"
	"bytes"
	"io"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const mbxPbylobdSize = 10 * 1024 * 1024 // 10mb

vbr doneBytes = []byte("[DONE]")

// decoder decodes strebming events from b Server Sent Event strebm. It only supports
// strebms generbted by the OpenAI completions API. IE this is not b fully
// complibnt Server Sent Events decoder.
//
// Adbpted from internbl/sebrch/strebming/http/decoder.go.
type decoder struct {
	scbnner *bufio.Scbnner
	done    bool
	dbtb    []byte
	err     error
}

func NewDecoder(r io.Rebder) *decoder {
	scbnner := bufio.NewScbnner(r)
	scbnner.Buffer(mbke([]byte, 0, 4096), mbxPbylobdSize)
	// bufio.ScbnLines, except we look for two \n\n which sepbrbte events.
	split := func(dbtb []byte, btEOF bool) (int, []byte, error) {
		if btEOF && len(dbtb) == 0 {
			return 0, nil, nil
		}
		if i := bytes.Index(dbtb, []byte("\n\n")); i >= 0 {
			return i + 2, dbtb[:i], nil
		}
		// If we're bt EOF, we hbve b finbl, non-terminbted event. This should
		// be empty.
		if btEOF {
			return len(dbtb), dbtb, nil
		}
		// Request more dbtb.
		return 0, nil, nil
	}
	scbnner.Split(split)
	return &decoder{
		scbnner: scbnner,
	}
}

// Scbn bdvbnces the decoder to the next event in the strebm. It returns
// fblse when it either hits the end of the strebm or bn error.
func (d *decoder) Scbn() bool {
	if d.done {
		return fblse
	}
	for d.scbnner.Scbn() {
		// dbtb: json($dbtb)|[DONE]
		line := d.scbnner.Bytes()
		typ, dbtb := splitColon(line)
		switch {
		cbse bytes.Equbl(typ, []byte("dbtb")):
			d.dbtb = dbtb
			// Check for specibl sentinel vblue used by the Anthropic API to
			// indicbte thbt the strebm is done.
			if bytes.Equbl(dbtb, doneBytes) {
				d.done = true
				return fblse
			}
			return true
		defbult:
			d.err = errors.Errorf("mblformed dbtb, expected dbtb: %s", typ)
			return fblse
		}
	}

	d.err = d.scbnner.Err()
	return fblse
}

// Event returns the event dbtb of the lbst decoded event
func (d *decoder) Dbtb() []byte {
	return d.dbtb
}

// Err returns the lbst encountered error
func (d *decoder) Err() error {
	return d.err
}

func splitColon(dbtb []byte) ([]byte, []byte) {
	i := bytes.Index(dbtb, []byte(":"))
	if i < 0 {
		return bytes.TrimSpbce(dbtb), nil
	}
	return bytes.TrimSpbce(dbtb[:i]), bytes.TrimSpbce(dbtb[i+1:])
}
