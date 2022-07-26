package codeownership

import (
	"fmt"
	"strings"
)

type RulesTrie struct {
	Rules []Rule
	Root  *Node
}

type Rule struct {
	Pattern string
}

func (t *RulesTrie) String() string {
	return t.Root.String()
}

type Node struct {
	RuleIndices []int
	Children    map[string]*Node
}

func (t *Node) String() string {
	return t.StringWithIndentation(0)
}

func (t *Node) StringWithIndentation(indentation int) string {
	str := ""
	indentationString := strings.Repeat(" ", indentation)
	for rule := range t.RuleIndices {
		str += fmt.Sprintf("%s%d\n", indentationString, rule)
	}
	for key, child := range t.Children {
		str += fmt.Sprintf("%s%s:\n", indentationString, key)
		str += child.StringWithIndentation(indentation + 1)
	}
	return str
}
