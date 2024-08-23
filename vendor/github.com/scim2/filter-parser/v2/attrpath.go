package filter

import (
	"github.com/di-wu/parser"
	"github.com/di-wu/parser/ast"
	"github.com/scim2/filter-parser/v2/internal/grammar"
	"github.com/scim2/filter-parser/v2/internal/types"
	"strings"
)

// ParseAttrPath parses the given raw data as an AttributePath.
func ParseAttrPath(raw []byte) (AttributePath, error) {
	p, err := ast.New(raw)
	if err != nil {
		return AttributePath{}, err
	}
	node, err := grammar.AttrPath(p)
	if err != nil {
		return AttributePath{}, err
	}
	if _, err := p.Expect(parser.EOD); err != nil {
		return AttributePath{}, err
	}
	return parseAttrPath(node)
}

func parseAttrPath(node *ast.Node) (AttributePath, error) {
	if node.Type != typ.AttrPath {
		return AttributePath{}, invalidTypeError(typ.AttrPath, node.Type)
	}

	// Indicates whether we encountered an attribute name.
	// These are the minimum requirements of an attribute path.
	var valid bool

	var attrPath AttributePath
	for _, node := range node.Children() {
		switch t := node.Type; t {
		case typ.URI:
			uri := node.Value
			uri = strings.TrimSuffix(uri, ":")
			attrPath.URIPrefix = &uri
		case typ.AttrName:
			name := node.Value
			if attrPath.AttributeName == "" {
				attrPath.AttributeName = name

				valid = true
			} else {
				attrPath.SubAttribute = &name
			}
		default:
			return AttributePath{}, invalidChildTypeError(typ.AttrPath, t)
		}
	}

	if !valid {
		return AttributePath{}, missingValueError(typ.AttrPath, typ.AttrName)
	}
	return attrPath, nil
}
