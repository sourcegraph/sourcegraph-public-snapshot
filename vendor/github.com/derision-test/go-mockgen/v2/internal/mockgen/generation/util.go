package generation

import (
	"github.com/dave/jennifer/jen"
	"github.com/derision-test/go-mockgen/v2/internal/mockgen/types"
)

func compose(stmt *jen.Statement, tail ...jen.Code) *jen.Statement {
	head := *stmt
	for _, value := range tail {
		head = append(head, value)
	}

	return &head
}

func addComment(code *jen.Statement, level int, commentText string) *jen.Statement {
	if commentText == "" {
		return code
	}

	comment := generateComment(level, commentText)
	return compose(comment, code)
}

func selfAppend(sliceRef *jen.Statement, value jen.Code) jen.Code {
	return compose(sliceRef, jen.Op("=").Id("append").Call(sliceRef, value))
}

func addTypes(code *jen.Statement, typeParams []types.TypeParam, outputImportPath string, includeTypes bool) *jen.Statement {
	if len(typeParams) == 0 {
		return code
	}

	types := make([]jen.Code, 0, len(typeParams))
	for _, typeParam := range typeParams {
		if includeTypes {
			types = append(types, compose(jen.Id(typeParam.Name), generateType(typeParam.Type, "", outputImportPath, false)))
		} else {
			types = append(types, jen.Id(typeParam.Name))
		}
	}

	return compose(code, jen.Types(types...))
}
