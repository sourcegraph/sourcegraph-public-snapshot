package ast

import "fmt"

// Capture is a structure to indicate that the value should be converted to a
// node. If one of the children returns a node, then that node gets returned
type Capture struct {
	// Type of the node.
	Type int
	// TypeStrings contains all the string representations of the available types.
	TypeStrings []string
	// Value is the expression to capture the value of the node.
	Value interface{}
}

func (c Capture) String() string {
	if 0 <= c.Type && c.Type < len(c.TypeStrings) {
		return fmt.Sprintf("%s", c.TypeStrings[c.Type])
	}
	return fmt.Sprintf("{%03d}", c.Type)
}
