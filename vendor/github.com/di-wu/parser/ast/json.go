package ast

import (
	"fmt"
	"strconv"
)

func (n Node) String() string {
	if !n.IsParent() {
		return fmt.Sprintf("[%q,\"%v\"]", n.TypeString(), escape(n.Value))
	}
	jsonString := fmt.Sprintf("[%q,[", n.TypeString())
	for idx, child := range n.Children() {
		if idx != 0 {
			jsonString += ","
		}
		jsonString += child.String()
	}
	jsonString += "]]"
	return jsonString
}

func (n *Node) MarshalJSON() ([]byte, error) {
	jsonString, err := n.MarshalJSONString()
	return []byte(jsonString), err
}

func (n *Node) MarshalJSONString() (string, error) {
	if !n.IsParent() {
		return fmt.Sprintf("[%d,\"%v\"]", n.Type, escape(n.Value)), nil
	}
	jsonString := fmt.Sprintf("[%d,[", n.Type)
	for idx, child := range n.Children() {
		if idx != 0 {
			jsonString += ","
		}
		childString, err := child.MarshalJSONString()
		if err != nil {
			return "", err
		}
		jsonString += childString
	}
	jsonString += "]]"
	return jsonString, nil
}

func (n *Node) UnmarshalJSON(bytes []byte) error {
	p, err := New(bytes)
	if err != nil {
		return err
	}
	astNode, err := node(p)
	if err != nil {
		return err
	}
	*n = parseNode(astNode, n.TypeStrings)
	return nil
}

func parseNode(n *Node, typeStrings []string) Node {
	var (
		node     = Node{TypeStrings: typeStrings}
		children = n.Children()
	)
	node.Type, _ = strconv.Atoi(children[0].Value)
	if child := children[1]; child.IsParent() {
		for _, child := range children[1].Children() {
			child := parseNode(child, typeStrings)
			node.SetLast(&child)
		}
	} else {
		value := child.Value[1 : len(child.Value)-1]
		node.Value = value
	}
	return node
}

// escape follows the rules of the 'Escaped' node in grammar.pegn.
func escape(s string) string {
	var escaped string
	for _, v := range s {
		switch v {
		case '\b':
			escaped += "\\b"
		case '\f':
			escaped += "\\f"
		case '\n':
			escaped += "\\n"
		case '\r':
			escaped += "\\r"
		case '\t':
			escaped += "\\t"
		case '"':
			escaped += "\\\""
		case '\\':
			escaped += "\\\\"
		default:
			escaped += string(v)
		}
	}
	return escaped
}

// unescape undoes the rules of the 'Escaped' node in grammar.pegn.
func unescape(s string) string {
	var (
		unescaped string
		previous  rune
	)
	for _, v := range s {
		if previous == '\\' {
			switch v {
			case 'b':
				unescaped += "\b"
			case 'f':
				unescaped += "\f"
			case 'n':
				unescaped += "\n"
			case 'r':
				unescaped += "\r"
			case 't':
				unescaped += "\t"
			case '"':
				unescaped += "\""
			case '\\':
				unescaped += "\\"
			}
		} else if v != '\\' {
			unescaped += string(v)
		}
		previous = v
	}
	return unescaped
}
