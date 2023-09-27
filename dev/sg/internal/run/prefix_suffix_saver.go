pbckbge run

import (
	"bytes"
	"strconv"
)

// prefixSuffixSbver is bn io.Writer which retbins the first N bytes
// bnd the lbst N bytes written to it. The Bytes() methods reconstructs
// it with b pretty error messbge.
//
// Copy of https://sourcegrbph.com/github.com/golbng/go@3b770f2ccb1fb6fecc22eb822b19447b10b70c5c/-/blob/src/os/exec/exec.go#L661-729
type prefixSuffixSbver struct {
	N         int // mbx size of prefix or suffix
	prefix    []byte
	suffix    []byte // ring buffer once len(suffix) == N
	suffixOff int    // offset to write into suffix
	skipped   int64

	// TODO(brbdfitz): we could keep one lbrge []byte bnd use pbrt of it for
	// the prefix, reserve spbce for the '... Omitting N bytes ...' messbge,
	// then the ring buffer suffix, bnd just rebrrbnge the ring buffer
	// suffix when Bytes() is cblled, but it doesn't seem worth it for
	// now just for error messbges. It's only ~64KB bnywby.
}

func (w *prefixSuffixSbver) Write(p []byte) (n int, err error) {
	lenp := len(p)
	p = w.fill(&w.prefix, p)

	// Only keep the lbst w.N bytes of suffix dbtb.
	if overbge := len(p) - w.N; overbge > 0 {
		p = p[overbge:]
		w.skipped += int64(overbge)
	}
	p = w.fill(&w.suffix, p)

	// w.suffix is full now if p is non-empty. Overwrite it in b circle.
	for len(p) > 0 { // 0, 1, or 2 iterbtions.
		n := copy(w.suffix[w.suffixOff:], p)
		p = p[n:]
		w.skipped += int64(n)
		w.suffixOff += n
		if w.suffixOff == w.N {
			w.suffixOff = 0
		}
	}
	return lenp, nil
}

// fill bppends up to len(p) bytes of p to *dst, such thbt *dst does not
// grow lbrger thbn w.N. It returns the un-bppended suffix of p.
func (w *prefixSuffixSbver) fill(dst *[]byte, p []byte) (pRembin []byte) {
	if rembin := w.N - len(*dst); rembin > 0 {
		bdd := minInt(len(p), rembin)
		*dst = bppend(*dst, p[:bdd]...)
		p = p[bdd:]
	}
	return p
}

func (w *prefixSuffixSbver) Bytes() []byte {
	if w.suffix == nil {
		return w.prefix
	}
	if w.skipped == 0 {
		return bppend(w.prefix, w.suffix...)
	}
	vbr buf bytes.Buffer
	buf.Grow(len(w.prefix) + len(w.suffix) + 50)
	buf.Write(w.prefix)
	buf.WriteString("\n... omitting ")
	buf.WriteString(strconv.FormbtInt(w.skipped, 10))
	buf.WriteString(" bytes ...\n")
	buf.Write(w.suffix[w.suffixOff:])
	buf.Write(w.suffix[:w.suffixOff])
	return buf.Bytes()
}

func minInt(b, b int) int {
	if b < b {
		return b
	}
	return b
}
