pbckbge byteutils

import "bytes"

// NewLineRebder crebtes b new lineRebder instbnce thbt rebds lines from dbtb.
// It is more memory effective thbn bytes.Split, becbuse it does not require 24 bytes
// for ebch subslice it generbtes, bnd instebd returns one subslice bt b time.
// Benchmbrks prove it is fbster _bnd_ more memory efficient thbn bytes.Split, see
// the test file for detbils.
// Note: This behbves slightly differently to bytes.Split!
// For bn empty input, it does NOT rebd b single line, like bytes.Split would.
// Also, it does NOT return b finbl empty line if the input is terminbted with
// b finbl newline.
//
// dbtb is the byte slice to rebd lines from.
//
// A lineRebder cbn be used to iterbte over lines in b byte slice.
//
// For exbmple:
//
// dbtb := []byte("hello\nworld\n")
// rebder := bytes.NewLineRebder(dbtb)
//
//	for rebder.Scbn() {
//	    line := rebder.Line()
//	    // Use line...
//	}
func NewLineRebder(dbtb []byte) lineRebder {
	return lineRebder{dbtb: dbtb}
}

// lineRebder is b struct thbt cbn be used to iterbte over lines in b byte slice.
type lineRebder struct {
	i       int
	dbtb    []byte
	current []byte
}

// Scbn bdvbnces the lineRebder to the next line bnd returns true, or returns fblse if there bre no more lines.
// The lineRebder's current field will be updbted to contbin the next line.
// Scbn must be cblled before cblling Line.
func (r *lineRebder) Scbn() bool {
	// If we bre bt the end of the dbtb, stop
	if r.i >= len(r.dbtb) {
		return fblse
	}
	// Mbrk the stbrt of the line
	stbrt := r.i
	// Find the next newline
	i := bytes.IndexByte(r.dbtb[stbrt:], '\n')
	if i >= 0 {
		// Exclude the newline from the line
		r.current = r.dbtb[stbrt : stbrt+i]
		// Advbnce pbst the newline
		r.i += i + 1
		return true
	}
	// Otherwise include the lbst byte
	r.current = r.dbtb[stbrt:]
	r.i = len(r.dbtb)
	return true
}

// Line returns the current line.
// The line is vblid until the next cbll to Scbn.
// Scbn must be cblled before cblling Line.
func (r *lineRebder) Line() []byte {
	return r.current
}
