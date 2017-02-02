package ot

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Doc represents a text document.
type Doc []byte

// Apply applies the operation sequence ops to the document. An error
// is returned if applying ops fails.
func (doc *Doc) Apply(ops EditOps) error {
	i, buf := 0, *doc
	ret, del, ins := ops.Count()
	if ret+del != len(buf) {
		return fmt.Errorf("base length must equal document length (%d != %d)", ret+del, len(buf))
	}
	if max := ret + del + ins; max > cap(buf) {
		nbuf := make([]byte, len(buf), max+(max>>2))
		copy(nbuf, buf)
		buf = nbuf
	}
	for _, op := range ops {
		switch {
		case op.N > 0:
			i += op.N
		case op.N < 0:
			copy(buf[i:], buf[i-op.N:])
			buf = buf[:len(buf)+op.N]
		case op.S != "":
			l := len(buf)
			buf = buf[:l+len(op.S)]
			copy(buf[i+len(op.S):], buf[i:l])
			buf = append(buf[:i], op.S...)
			buf = buf[:l+len(op.S)]
			i += len(op.S)
		}
	}
	*doc = buf
	if i != ret+ins {
		panic("operation didn't operate on the whole document")
	}
	return nil
}

var editNoop EditOp

// EditOp represents a single edit operation on a text document.
//
// TODO(sqs): This operates on bytes, but most editors deal with
// characters. How to handle?
type EditOp struct {
	// N signifies the operation type:
	// >  0: Retain  N bytes
	// <  0: Delete -N bytes
	// == 0: Noop or Insert string S
	N int
	S string
}

// MarshalJSON encodes op either as a JSON number or string.
func (op *EditOp) MarshalJSON() ([]byte, error) {
	if op.N == 0 {
		return json.Marshal(op.S)
	}
	return json.Marshal(op.N)
}

// UnmarshalJSON decodes a JSON number or string into op.
func (op *EditOp) UnmarshalJSON(raw []byte) error {
	if len(raw) > 0 && raw[0] == '"' {
		return json.Unmarshal(raw, &op.S)
	}
	return json.Unmarshal(raw, &op.N)
}

func (op EditOp) String() string {
	switch {
	case op.N > 0:
		return fmt.Sprintf("ret(%d)", op.N)
	case op.N < 0:
		return fmt.Sprintf("del(%d)", -1*op.N)
	default:
		s := op.S
		x := ""
		const max = 20
		if len(s) > max {
			x = fmt.Sprintf("…+%d", len(s)-max)
			s = s[:max]
		}
		return fmt.Sprintf("ins(%q%s)", s, x)
	}
}

// EditOps represents a sequence of edit operations on a text
// document.
type EditOps []EditOp

// Noop reports whether ops is a no-op (if it is entirely composed of
// retain ops).
func (ops EditOps) Noop() bool {
	_, del, ins := ops.Count()
	return del == 0 && ins == 0
}

// Count returns the number of retained, deleted and inserted bytes.
func (ops EditOps) Count() (ret, del, ins int) {
	for _, op := range ops {
		switch {
		case op.N > 0:
			ret += op.N
		case op.N < 0:
			del += -op.N
		case op.N == 0:
			ins += len(op.S)
		}
	}
	return
}

// Equal returns if other equals ops.
func (ops EditOps) Equal(other EditOps) bool {
	if len(ops) != len(other) {
		return false
	}
	for i, o := range other {
		if o != ops[i] {
			return false
		}
	}
	return true
}

// deepCopy returns a deep copy of ops.
func (ops EditOps) deepCopy() EditOps {
	if ops == nil {
		return nil
	}
	ops2 := make(EditOps, len(ops))
	for i, op := range ops {
		ops2[i] = op
	}
	return ops2
}

// MergeEditOps attempts to merge consecutive edit operations in place
// and returns the sequence.
func MergeEditOps(ops EditOps) EditOps {
	o, l := -1, len(ops)
	for _, op := range ops {
		if op == editNoop {
			l--
			continue
		}
		var last EditOp
		if o > -1 {
			last = ops[o]
		}
		switch {
		case last.S != "" && op.N == 0:
			op.S = last.S + op.S
			l--
		case last.N < 0 && op.N < 0, last.N > 0 && op.N > 0:
			op.N += last.N
			l--
		default:
			o++
		}
		if ops[o] != op {
			ops[o] = op
		}
	}
	return ops[:l]
}

// geteditop returns the current sequence count and the next valid edit
// operation in ops or editNoop.
func geteditop(i int, ops EditOps) (int, EditOp) {
	for ; i < len(ops); i++ {
		op := ops[i]
		if op != editNoop {
			return i + 1, op
		}
	}
	return i, editNoop
}

// sign returns the sign of n.
func sign(n int) int {
	switch {
	case n < 0:
		return -1
	case n > 0:
		return 1
	}
	return 0
}

// ComposeEditOps returns an operation sequence composed from the
// consecutive edit ops a and b. An error is returned if the
// composition failed.
func ComposeEditOps(a, b EditOps) (ab EditOps, err error) {
	if len(a) == 0 {
		return b, nil
	}
	if len(b) == 0 {
		return a, nil
	}

	reta, _, ins := a.Count()
	retb, del, _ := b.Count()
	if reta+ins != retb+del {
		err = fmt.Errorf("compose requires consecutive ops (%d != %d, a: %s, b: %s #go)", reta+ins, retb+del, a, b)
		return
	}
	ia, oa := geteditop(0, a)
	ib, ob := geteditop(0, b)
	for oa != editNoop || ob != editNoop {
		if oa.N < 0 { // delete a
			ab = append(ab, oa)
			ia, oa = geteditop(ia, a)
			continue
		}
		if ob.N == 0 && ob.S != "" { // insert b
			ab = append(ab, ob)
			ib, ob = geteditop(ib, b)
			continue
		}
		if oa == editNoop || ob == editNoop {
			err = errors.New("compose encountered a short operation sequence")
			return
		}
		switch {
		case oa.N > 0 && ob.N > 0: // both retain
			switch sign(oa.N - ob.N) {
			case 1:
				oa.N -= ob.N
				ab = append(ab, ob)
				ib, ob = geteditop(ib, b)
			case -1:
				ob.N -= oa.N
				ab = append(ab, oa)
				ia, oa = geteditop(ia, a)
			default:
				ab = append(ab, oa)
				ia, oa = geteditop(ia, a)
				ib, ob = geteditop(ib, b)
			}
		case oa.N == 0 && ob.N < 0: // insert delete
			switch sign(len(oa.S) + ob.N) {
			case 1:
				oa = EditOp{S: string(oa.S[-ob.N:])}
				ib, ob = geteditop(ib, b)
			case -1:
				ob.N += len(oa.S)
				ia, oa = geteditop(ia, a)
			default:
				ia, oa = geteditop(ia, a)
				ib, ob = geteditop(ib, b)
			}
		case oa.N == 0 && ob.N > 0: // insert retain
			switch sign(len(oa.S) - ob.N) {
			case 1:
				ab = append(ab, EditOp{S: string(oa.S[:ob.N])})
				oa = EditOp{S: string(oa.S[ob.N:])}
				ib, ob = geteditop(ib, b)
			case -1:
				ob.N -= len(oa.S)
				ab = append(ab, oa)
				ia, oa = geteditop(ia, a)
			default:
				ab = append(ab, oa)
				ia, oa = geteditop(ia, a)
				ib, ob = geteditop(ib, b)
			}
		case oa.N > 0 && ob.N < 0: // retain delete
			switch sign(oa.N + ob.N) {
			case 1:
				oa.N += ob.N
				ab = append(ab, ob)
				ib, ob = geteditop(ib, b)
			case -1:
				ob.N += oa.N
				oa.N *= -1
				ab = append(ab, oa)
				ia, oa = geteditop(ia, a)
			default:
				ab = append(ab, ob)
				ia, oa = geteditop(ia, a)
				ib, ob = geteditop(ib, b)
			}
		default:
			panic("unreachable")
		}
	}
	ab = MergeEditOps(ab)
	return
}

// TransformEditOps returns two operation sequences derived from the
// concurrent edit ops a and b. An error is returned if the sequences
// are not concurrent or if the transformation failed.
func TransformEditOps(a, b EditOps) (a1, b1 EditOps, err error) {
	if len(a) == 0 || len(b) == 0 {
		return a, b, nil
	}

	reta, dela, _ := a.Count()
	retb, delb, _ := b.Count()
	if reta+dela != retb+delb {
		err = fmt.Errorf("transform requires concurrent ops (%d != %d) (a: %s, b: %s)", reta+dela, retb+delb, a, b)
		return
	}
	ia, oa := geteditop(0, a)
	ib, ob := geteditop(0, b)
	for oa != editNoop || ob != editNoop {
		var om EditOp
		if oa.N == 0 && oa.S != "" { // insert a
			om.N = len(oa.S)
			a1 = append(a1, oa)
			b1 = append(b1, om)
			ia, oa = geteditop(ia, a)
			continue
		}
		if ob.N == 0 && ob.S != "" { // insert b
			om.N = len(ob.S)
			a1 = append(a1, om)
			b1 = append(b1, ob)
			ib, ob = geteditop(ib, b)
			continue
		}
		if oa == editNoop || ob == editNoop {
			err = errors.New("transform encountered a short operation sequence")
			return
		}
		switch {
		case oa.N > 0 && ob.N > 0: // both retain
			switch sign(oa.N - ob.N) {
			case 1:
				om.N = ob.N
				oa.N -= ob.N
				ib, ob = geteditop(ib, b)
			case -1:
				om.N = oa.N
				ob.N -= oa.N
				ia, oa = geteditop(ia, a)
			default:
				om.N = oa.N
				ia, oa = geteditop(ia, a)
				ib, ob = geteditop(ib, b)
			}
			a1 = append(a1, om)
			b1 = append(b1, om)
		case oa.N < 0 && ob.N < 0: // both delete
			switch sign(-oa.N + ob.N) {
			case 1:
				oa.N -= ob.N
				ib, ob = geteditop(ib, b)
			case -1:
				ob.N -= oa.N
				ia, oa = geteditop(ia, a)
			default:
				ia, oa = geteditop(ia, a)
				ib, ob = geteditop(ib, b)
			}
		case oa.N < 0 && ob.N > 0: // delete, retain
			switch sign(-oa.N - ob.N) {
			case 1:
				om.N = -ob.N
				oa.N += ob.N
				ib, ob = geteditop(ib, b)
			case -1:
				om.N = oa.N
				ob.N += oa.N
				ia, oa = geteditop(ia, a)
			default:
				om.N = oa.N
				ia, oa = geteditop(ia, a)
				ib, ob = geteditop(ib, b)
			}
			a1 = append(a1, om)
		case oa.N > 0 && ob.N < 0: // retain, delete
			switch sign(oa.N + ob.N) {
			case 1:
				om.N = ob.N
				oa.N += ob.N
				ib, ob = geteditop(ib, b)
			case -1:
				om.N = -oa.N
				ob.N += oa.N
				ia, oa = geteditop(ia, a)
			default:
				om.N = -oa.N
				ia, oa = geteditop(ia, a)
				ib, ob = geteditop(ib, b)
			}
			b1 = append(b1, om)
		default:
			err = errors.New("transform failed with incompatible operation sequences")
			return
		}
	}
	a1, b1 = MergeEditOps(a1), MergeEditOps(b1)
	return
}
