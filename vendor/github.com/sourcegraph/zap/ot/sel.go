package ot

import "fmt"

// Sel represents a 0-indexed character selection range. The first
// element is the anchor index of the selection (where the selection
// began), and the second element is the position index of the
// selection (where the cursor is).
type Sel [2]uint

func (s *Sel) String() string { return fmt.Sprintf("%d-%d", s[0], s[1]) }

func (s *Sel) start() uint { return s[0] }
func (s *Sel) end() uint   { return s[1] }

func (s *Sel) clone() *Sel {
	if s == nil {
		return nil
	}
	tmp := *s
	return &tmp
}

func cloneSels(usels map[string]*Sel) map[string]*Sel {
	c := make(map[string]*Sel, len(usels))
	for u, sel := range usels {
		c[u] = sel.clone()
	}
	return c
}

func mergeSels(a, b map[string]*Sel) map[string]*Sel {
	c := make(map[string]*Sel, (len(a)+len(b))/2)
	for u, s := range a {
		c[u] = s.clone()
	}
	for u, s := range b {
		c[u] = s.clone()
	}
	return c
}

// AdjustSel returns the selections after adjusting for the edit.
//
// TODO(sqs): support multiple sels, and add tests for multiple sels
// to TestAdjustSel as well.
func AdjustSel(sel *Sel, edit EditOps) *Sel {
	if sel == nil {
		return nil
	}

	if len(edit) == 0 {
		return sel
	}

	sign := func(n int) int {
		if n >= 0 {
			return 1
		}
		return -1
	}
	min := func(a, b uint) uint {
		if a < b {
			return a
		}
		return b
	}

	var c uint
	for _, op := range edit {
		var opStart, opEnd uint
		var n int
		switch {
		case op.N > 0:
			c += uint(op.N)
			continue
		case op.N < 0:
			opStart = c
			c += uint(-1 * op.N)
			opEnd = c
			n = op.N
		case op.S != "":
			opStart = c
			c += uint(len(op.S))
			opEnd = c
			n = len(op.S)
		}

		opEnd = min(opEnd, sel.end())

		// TODO(sqs) TODO!: check for uint overflow of 'sel[0] +=' and
		// 'sel[1] +=' in either direction
		//
		// TODO(sqs) TODO!: handle reverse selections
		if opStart < sel.start() {
			startOverlap := int(min(opEnd, sel.start())) - int(opStart)
			if startOverlap < 0 {
				panic(fmt.Sprintf("startOverlap (%d) < 0", startOverlap))
			}
			sel[0] += uint(startOverlap * sign(n))
		}
		if opEnd <= sel.end() {
			endOverlap := int(min(opEnd, sel.end())) - int(opStart)
			if endOverlap < 0 {
				endOverlap = 0
			}
			sel[1] += uint(endOverlap * sign(n))
		}
	}
	return sel
}
