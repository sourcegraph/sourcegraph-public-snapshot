package annotate

import (
	"bytes"
	"errors"
	"io"
)

type Annotation struct {
	// Start and End byte offsets (not rune offsets).
	Start, End int

	Left, Right []byte
	WantInner   int
}

type Annotations []*Annotation

func (a Annotations) Len() int { return len(a) }
func (a Annotations) Less(i, j int) bool {
	// Sort by start position, breaking ties by preferring longer
	// matches.
	ai, aj := a[i], a[j]
	if ai.Start == aj.Start {
		if ai.End == aj.End {
			return ai.WantInner < aj.WantInner
		}
		return ai.End > aj.End
	} else {
		return ai.Start < aj.Start
	}
}
func (a Annotations) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

// Annotates src with annotations in anns.
//
// Annotating an empty byte array always returns an empty byte array.
//
// Assumes anns is sorted (using sort.Sort(anns)).
func Annotate(src []byte, anns Annotations, writeContent func(io.Writer, []byte)) ([]byte, error) {
	out := bytes.NewBuffer(make([]byte, 0, len(src)+20*len(anns)))
	var err error

	// Keep a stack of open annotations (i.e., that have been opened and not yet
	// closed).
	var open []*Annotation

	for b := range src {
		// Open annotations that begin here.
		for i, a := range anns {
			if a.Start < 0 || a.Start == b {
				if a.Start < 0 {
					err = ErrStartOutOfBounds
				}

				out.Write(a.Left)

				if a.Start == a.End {
					out.Write(a.Right)
				} else {
					// Put this annotation on the stack of annotations that will need
					// to be closed. We remove it from anns at the end of the loop
					// (to avoid modifying anns while we're iterating over it).
					open = append(open, a)
				}
			} else if a.Start > b {
				// Remove all annotations that we opened (we already put them on the
				// stack of annotations that will need to be closed).
				anns = anns[i:]
				break
			}
		}

		if writeContent == nil {
			out.Write(src[b : b+1])
		} else {
			writeContent(out, src[b:b+1])
		}

		// Close annotations that end after this byte, handling overlapping
		// elements as described below. Elements of open are ordered by their
		// annotation start position.

		// We need to close all annotatations ending after this byte, as well as
		// annotations that overlap this annotation's end and should reopen
		// after it closes.
		var toClose []*Annotation

		// Find annotations ending after this byte.
		minStart := 0 // start of the leftmost annotation closing here
		for i := len(open) - 1; i >= 0; i-- {
			a := open[i]
			if a.End == b+1 {
				toClose = append(toClose, a)
				if minStart == 0 || a.Start < minStart {
					minStart = a.Start
				}
				open = append(open[:i], open[i+1:]...)
			}
		}

		// Find annotations that overlap annotations closing after this and
		// that should reopen after it closes.
		if toClose != nil {
			for i := len(open) - 1; i >= 0; i-- {
				if a := open[i]; a.Start > minStart {
					out.Write(a.Right)
				}
			}
		}

		for _, a := range toClose {
			out.Write(a.Right)
		}

		if toClose != nil {
			for _, a := range open {
				if a.Start > minStart {
					out.Write(a.Left)
				}
			}
		}
	}

	if len(open) > 0 {
		if err == ErrStartOutOfBounds {
			err = ErrStartAndEndOutOfBounds
		} else {
			err = ErrEndOutOfBounds
		}

		// Clean up by closing unclosed annotations, in the order they would
		// have been closed in.
		for i := len(open) - 1; i >= 0; i-- {
			a := open[i]
			out.Write(a.Right)
		}
	}

	return out.Bytes(), err
}

var (
	ErrStartOutOfBounds       = errors.New("annotation start out of bounds")
	ErrEndOutOfBounds         = errors.New("annotation end out of bounds")
	ErrStartAndEndOutOfBounds = errors.New("annotations start and end out of bounds")
)

func IsOutOfBounds(err error) bool {
	return err == ErrStartOutOfBounds || err == ErrEndOutOfBounds || err == ErrStartAndEndOutOfBounds
}

func annLefts(as []*Annotation) []string {
	ls := make([]string, len(as))
	for i, a := range as {
		ls[i] = string(a.Left)
	}
	return ls
}
