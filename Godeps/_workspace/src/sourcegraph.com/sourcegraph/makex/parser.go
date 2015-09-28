package makex

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
)

// Parse parses a Makefile into a *Makefile struct.
//
// TODO(sqs): super hacky.
func Parse(data []byte) (*Makefile, error) {
	var mf Makefile

	lines := bytes.Split(data, []byte{'\n'})
	var rule *BasicRule
	for lineno, lineBytes := range lines {
		line := string(lineBytes)
		if strings.HasPrefix(line, "\t") {
			if rule == nil {
				return nil, fmt.Errorf("line %d: indented recipe not inside a rule", lineno)
			}
			recipe := strings.TrimPrefix(line, "\t")
			recipe = ExpandAutoVars(rule, recipe)
			rule.RecipeCmds = append(rule.RecipeCmds, recipe)
		} else if strings.Contains(line, ":") {
			sep := strings.Index(line, ":")
			targets := strings.Fields(line[:sep])
			if len(targets) > 1 {
				return nil, errMultipleTargetsUnsupported(lineno)
			}
			target := targets[0]
			prereqs := strings.Fields(line[sep+1:])
			prereqs = uniqAndSort(prereqs)
			rule = &BasicRule{TargetFile: target, PrereqFiles: prereqs}
			mf.Rules = append(mf.Rules, rule)
		} else {
			rule = nil
		}
	}

	return &mf, nil
}

func errMultipleTargetsUnsupported(lineno int) error {
	return fmt.Errorf("line %d: rule with multiple targets is yet implemented", lineno)
}

func uniqAndSort(strs []string) []string {
	sort.Strings(strs)
	uniq := make([]string, 0, len(strs))
	for i, s := range strs {
		if i == 0 || strs[i-1] != s {
			uniq = append(uniq, s)
		}
	}
	return uniq
}
