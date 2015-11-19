package local

import (
	"log"

	wd "github.com/mb0/diff"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

// chunk holds a group of consecutive additions and deletions.
type chunk struct {
	// start holds the line number relative to the hunk start
	// where the first line in this chunk starts.
	start          int
	removed, added []*sourcegraph.SourceCodeLine
}

// addAddition adds a new line to the group of additions in the chunk.
func (c *chunk) addAddition(line *sourcegraph.SourceCodeLine) {
	if c.added == nil {
		c.added = make([]*sourcegraph.SourceCodeLine, 0, 1)
	}
	c.added = append(c.added, line)
}

// addRemoval adds a new line to the group of deletions in the chunk.
func (c *chunk) addRemoval(line *sourcegraph.SourceCodeLine) {
	if c.removed == nil {
		c.removed = make([]*sourcegraph.SourceCodeLine, 0, 1)
	}
	c.removed = append(c.removed, line)
}

// isBalanced verifies if the chunk is balanced by verifying equality
// of changes.
func (c *chunk) isBalanced() bool {
	return len(c.removed) == len(c.added) && len(c.removed) != 0
}

func (c *chunk) isNil() bool { return c.removed == nil && c.added == nil }

func (c *chunk) reset() {
	c.removed = nil
	c.added = nil
	c.start = 0
}

// pairs returns a slice of pairs compatible with the mb0/diff package
// to calculate word-level differences.
func (c *chunk) pairs() []*pair {
	if !c.isBalanced() {
		return nil
	}
	p := make([]*pair, len(c.removed))
	for i := range c.removed {
		p[i] = &pair{
			IndexA: c.start + i,
			A:      c.removed[i],
			IndexB: c.start + len(c.removed) + i,
			B:      c.added[i],
		}
	}
	return p
}

// pair holds two lines of code: an addition and a deletion ready for comparison.
type pair struct {
	// A holds the deletion and B holds the addition.
	A, B *sourcegraph.SourceCodeLine

	// IndexA represents the line number offset relative to the original hunk for
	// the line at A.
	IndexA int

	// IndexB represents the line number offset relative to the original hunk for
	// the line at B.
	IndexB int
}

func (p *pair) Equal(i, j int) bool {
	a, b := p.A.Tokens[i], p.B.Tokens[j]
	if a != nil && b != nil && a.Label == b.Label {
		return true
	}
	return false
}

func wordDiff(hunk *sourcegraph.Hunk) {
	if hunk.BodySource == nil {
		log.Println("could not perform wordDiff because body is nil")
		return
	}
	if len(hunk.LinePrefixes) > len(hunk.BodySource.Lines) {
		log.Println("could not perform wordDiff because body is invalid")
		return
	}
	validChunks := make([]chunk, 0, 1)
	// check verifies if a chunk is balanced and if it is, it saves it
	// and resets it.
	check := func(c *chunk) {
		if c.isBalanced() {
			validChunks = append(validChunks, *c)
			c.reset()
		}
	}
	c := new(chunk)
	for i, t := range hunk.LinePrefixes {
		switch t {
		case '+':
			if c.isNil() {
				c.start = i
			}
			c.addAddition(hunk.BodySource.Lines[i])
		case '-':
			if c.isNil() {
				c.start = i
			}
			c.addRemoval(hunk.BodySource.Lines[i])
		case ' ':
			check(c)
			c.reset()
		}
	}
	check(c)
	for _, c := range validChunks {
		// create pairs from valid chunks
		if pairs := c.pairs(); pairs != nil {
			for _, p := range pairs {
				// mark tokens
				changes := wd.Diff(len(p.A.Tokens), len(p.B.Tokens), p)
				for _, chg := range changes {
					for i := chg.A; i < chg.A+chg.Del; i++ {
						p.A.Tokens[i].ExtraClasses = "x"
					}
					for i := chg.B; i < chg.B+chg.Ins; i++ {
						p.B.Tokens[i].ExtraClasses = "x"
					}
				}
			}
		}
	}
}
