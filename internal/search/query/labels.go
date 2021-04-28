package query

import "sort"

// Labels are general-purpose annotations that store information about a node.
type labels uint16

const (
	None    labels = 0
	Literal        = 1 << iota
	Regexp
	Quoted
	HeuristicParensAsPatterns
	HeuristicDanglingParens
	HeuristicHoisted
	Structural
	IsPredicate
)

var allLabels = map[labels]string{
	None:                      "None",
	Literal:                   "Literal",
	Regexp:                    "Regexp",
	Quoted:                    "Quoted",
	HeuristicParensAsPatterns: "HeuristicParensAsPatterns",
	HeuristicDanglingParens:   "HeuristicDanglingParens",
	HeuristicHoisted:          "HeuristicHoisted",
	Structural:                "Structural",
	IsPredicate:               "IsPredicate",
}

func (l *labels) IsSet(label labels) bool {
	return *l&label != 0
}

func (l *labels) set(label labels) {
	*l |= label
}

func (l *labels) unset(label labels) {
	*l &^= label
}

func (l *labels) String() []string {
	if *l == 0 {
		return []string{"None"}
	}
	var s []string
	for k, v := range allLabels {
		if l.IsSet(k) {
			s = append(s, v)
		}
	}
	sort.Strings(s)
	return s
}
