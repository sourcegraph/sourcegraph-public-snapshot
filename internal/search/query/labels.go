package query

import "sort"

// Labels are general-purpose annotations that store information about a node.
type labels uint8

const (
	None    labels = 0
	Literal        = 1 << iota
	Quoted
	HeuristicParensAsPatterns
	HeuristicDanglingParens
	HeuristicHoisted
)

var allLabels = map[labels]string{
	None:                      "None",
	Literal:                   "Literal",
	Quoted:                    "Quoted",
	HeuristicParensAsPatterns: "HeuristicParensAsPatterns",
	HeuristicDanglingParens:   "HeuristicDanglingParens",
	HeuristicHoisted:          "HeuristicHoisted",
}

func Strings(labels labels) []string {
	if labels == 0 {
		return []string{"None"}
	}
	var s []string
	for k, v := range allLabels {
		if k&labels != 0 {
			s = append(s, v)
		}
	}
	sort.Strings(s)
	return s
}
