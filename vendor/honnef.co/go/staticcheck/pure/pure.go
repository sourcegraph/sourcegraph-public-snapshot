// Package pure analyzes functions for purity.
//
// It currently does a (very) conservative analysis, forbidding most
// language constructs. It's in turn only able to detect simple pure
// functions.
//
// The analysis may be improved in the future.
package pure

import (
	"go/token"
	"go/types"

	"honnef.co/go/ssa"
)

type State struct {
	purity map[*ssa.Function]bool
}

func (s *State) IsPure(fn *ssa.Function) (ret bool) {
	if s.purity == nil {
		s.purity = map[*ssa.Function]bool{}
	}
	if p, ok := s.purity[fn]; ok {
		return p
	}
	// stop infinite recursion by considering recursive calls unpure.
	s.purity[fn] = false
	defer func() {
		s.purity[fn] = ret
	}()

	for _, param := range fn.Params {
		if _, ok := param.Type().Underlying().(*types.Basic); !ok {
			return false
		}
	}

	if fn.Blocks == nil {
		return false
	}
	checkCall := func(c *ssa.CallCommon) bool {
		if c.IsInvoke() {
			return false
		}
		builtin, ok := c.Value.(*ssa.Builtin)
		if !ok {
			if c.StaticCallee() != fn {
				if c.StaticCallee() == nil || !s.IsPure(c.StaticCallee()) {
					return false
				}
			}
		} else {
			switch builtin.Name() {
			case "len", "cap", "make", "new":
			default:
				return false
			}
		}
		return true
	}
	for _, b := range fn.Blocks {
		for _, ins := range b.Instrs {
			switch ins := ins.(type) {
			case *ssa.Call:
				if !checkCall(ins.Common()) {
					return false
				}
			case *ssa.Defer:
				if !checkCall(&ins.Call) {
					return false
				}
			case *ssa.Select:
				return false
			case *ssa.Send:
				return false
			case *ssa.Go:
				return false
			case *ssa.Panic:
				return false
			case *ssa.Store:
				return false
			case *ssa.FieldAddr:
				return false
			case *ssa.UnOp:
				if ins.Op == token.MUL || ins.Op == token.AND {
					return false
				}
			}
		}
	}
	return true
}
