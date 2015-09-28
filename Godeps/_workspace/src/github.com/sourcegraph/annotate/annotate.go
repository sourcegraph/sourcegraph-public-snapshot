package annotate

import (
	"bytes"
	"errors"
	"io"
	"sort"
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

	// Keep a stack of annotations we should close at all future rune offsets.
	closeAnnsAtByte := make(map[int]Annotations, len(src)/10)

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
					closeAnnsAtByte[a.End] = append(closeAnnsAtByte[a.End], a)
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

		// Close annotations that after this rune.
		if closeAnns, present := closeAnnsAtByte[b+1]; present {
			for i := len(closeAnns) - 1; i >= 0; i-- {
				out.Write(closeAnns[i].Right)
			}
			delete(closeAnnsAtByte, b+1)
		}
	}

	if unclosed := len(closeAnnsAtByte); unclosed > 0 {
		if err == ErrStartOutOfBounds {
			err = ErrStartAndEndOutOfBounds
		} else {
			err = ErrEndOutOfBounds
		}

		// Clean up by closing unclosed annotations, in the order they would have been
		// closed in.
		unclosedAnns := make(Annotations, 0, len(closeAnnsAtByte))
		for _, anns := range closeAnnsAtByte {
			unclosedAnns = append(unclosedAnns, anns...)
		}
		sort.Sort(sort.Reverse(unclosedAnns))
		for _, a := range unclosedAnns {
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
