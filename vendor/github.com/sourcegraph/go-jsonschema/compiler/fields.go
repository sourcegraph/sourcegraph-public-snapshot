package compiler

import (
	"bytes"
	"go/ast"
	"go/printer"
)

type field struct {
	GoName, JSONName string
	*ast.Field
}

func (f field) GoType() string {
	var buf bytes.Buffer
	err := printer.Fprint(&buf, nil, f.Field.Type)
	if err != nil {
		panic(err)
	}
	return buf.String()
}

func (f field) GoStructFieldTag() string {
	return f.Field.Tag.Value
}

func astFields(fields []field) []*ast.Field {
	fs := make([]*ast.Field, len(fields))
	for i, f := range fields {
		fs[i] = f.Field
	}
	return fs
}
