package graphqlbackend

import "github.com/sourcegraph/sourcegraph/pkg/vcs/git"

type highlightedRange struct {
	line      int32
	character int32
	length    int32
}

func (h *highlightedRange) Line() int32      { return h.line }
func (h *highlightedRange) Character() int32 { return h.character }
func (h *highlightedRange) Length() int32    { return h.length }

type highlightedString struct {
	value      string
	highlights []*highlightedRange
}

func (s *highlightedString) Value() string                   { return s.value }
func (s *highlightedString) Highlights() []*highlightedRange { return s.highlights }

func fromVCSHighlights(vcsHighlights []git.Highlight) []*highlightedRange {
	highlights := make([]*highlightedRange, len(vcsHighlights))
	for i, vh := range vcsHighlights {
		highlights[i] = &highlightedRange{
			line:      int32(vh.Line),
			character: int32(vh.Character),
			length:    int32(vh.Length),
		}
	}
	return highlights
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_156(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		
