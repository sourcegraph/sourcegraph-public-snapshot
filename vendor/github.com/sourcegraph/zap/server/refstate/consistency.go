package refstate

import (
	"errors"
	"fmt"
	"strings"

	"github.com/sourcegraph/zap"
)

func compareRefPointerInfo(p zap.RefPointer, r zap.RefState) error {
	var diffs []string
	if p.Target != r.Target {
		diffs = append(diffs, fmt.Sprintf("target: %q != %q", p.Target, r.Target))
	}
	if (p.Base == nil) != (r.Data == nil) {
		diffs = append(diffs, fmt.Sprintf("base: %v != %v", p.Base, r.Data))
	} else if p.Base != nil {
		if p.Base.GitBase != r.Data.GitBase {
			diffs = append(diffs, fmt.Sprintf("git base: %q != %q", p.Base.GitBase, r.Data.GitBase))
		}
		if p.Base.GitBranch != r.Data.GitBranch {
			diffs = append(diffs, fmt.Sprintf("git branch: %q != %q", p.Base.GitBranch, r.Data.GitBranch))
		}
	}
	if len(diffs) == 0 {
		return nil
	}
	return errors.New(strings.Join(diffs, ", "))
}

func toRefPointer(p *zap.RefPointer) *zap.RefPointer {
	if p == nil {
		return nil
	}
	return &zap.RefPointer{Base: p.Base, Target: p.Target}
}
