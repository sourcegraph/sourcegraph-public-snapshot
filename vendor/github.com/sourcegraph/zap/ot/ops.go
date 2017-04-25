package ot

import (
	"encoding/json"
	"fmt"
)

// Ops is an ordered list of operations performed on a workspace.
type Ops []Op

// Equal reports if ops == other.
func (ops Ops) Equal(other Ops) bool {
	if len(ops) == 0 && len(other) == 0 {
		return true
	}
	if len(ops) != len(other) {
		return false
	}
	for i, o := range ops {
		if o == nil && other[i] != nil {
			return false
		}
		if o != nil && !o.Equal(other[i]) {
			return false
		}
	}
	return true
}

func (ops Ops) String() string {
	if ops == nil {
		return "<nil>"
	}
	return fmt.Sprint(([]Op)(ops))
}

func (ops *Ops) UnmarshalJSON(b []byte) error {
	type opType struct {
		Type string `json:"type"`
	}
	var rawOps []*json.RawMessage
	if err := json.Unmarshal(b, &rawOps); err != nil {
		return err
	}
	for _, r := range rawOps {
		var t *opType
		if err := json.Unmarshal(*r, &t); err != nil {
			return err
		}
		switch t.Type {
		case "create":
			var op *FileCreate
			if err := json.Unmarshal(*r, &op); err != nil {
				return err
			}
			*ops = append(*ops, *op)
		case "delete":
			var op *FileDelete
			if err := json.Unmarshal(*r, &op); err != nil {
				return err
			}
			*ops = append(*ops, *op)
		case "copy":
			var op *FileCopy
			if err := json.Unmarshal(*r, &op); err != nil {
				return err
			}
			*ops = append(*ops, *op)
		case "rename":
			var op *FileRename
			if err := json.Unmarshal(*r, &op); err != nil {
				return err
			}
			*ops = append(*ops, *op)
		case "truncate":
			var op *FileTruncate
			if err := json.Unmarshal(*r, &op); err != nil {
				return err
			}
			*ops = append(*ops, *op)
		case "edit":
			var op *FileEdit
			if err := json.Unmarshal(*r, &op); err != nil {
				return err
			}
			*ops = append(*ops, *op)
		case "gitHead":
			var op *GitHead
			if err := json.Unmarshal(*r, &op); err != nil {
				return err
			}
			*ops = append(*ops, *op)
		}
	}
	return nil
}

func (ops Ops) shallowCopy() Ops {
	tmp := make(Ops, len(ops))
	copy(tmp, ops)
	return tmp
}

func (ops Ops) DeepCopy() Ops {
	ops = ops.shallowCopy()
	for i, op := range ops {
		if op != nil {
			ops[i] = op.Copy()
		}
	}
	return ops
}

func (ops Ops) Noop() bool {
	// OT_TODO: is this correct?
	return len(ops) == 0
}

type sortableOps Ops // unexported alias to avoid polluting public API

func (ops sortableOps) Len() int      { return len(ops) }
func (ops sortableOps) Swap(i, j int) { ops[i], ops[j] = ops[j], ops[i] }
func (ops sortableOps) Less(i, j int) bool {
	if ops[i] == nil {
		return ops[j] != nil
	}
	return ops[i].Less(ops[j])
}

// join concats two Ops slices together.
func join(a, b Ops) Ops {
	ab := make(Ops, len(a)+len(b))
	copy(ab[:len(a)], a)
	copy(ab[len(a):len(a)+len(b)], b)
	return ab
}
