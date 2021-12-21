package api

import "fmt"

type Name struct {
	Value string
	Kind  int
}

func NewSimpleName(value string) *Name {
	return &Name{Value: value}
}

func (n *Name) String() string {
	if n.Kind == 0 {
		return n.Value
	}
	return fmt.Sprintf("%d_%s", n.Kind, n.Value)
}

func (n *Name) IsEqual(other *Name) bool {
	return n.Kind == other.Kind &&
		n.Value == other.Value
}
