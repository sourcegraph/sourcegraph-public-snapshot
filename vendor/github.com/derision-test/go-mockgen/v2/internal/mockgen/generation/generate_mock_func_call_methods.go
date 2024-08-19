package generation

import (
	"fmt"
	"strings"

	"github.com/dave/jennifer/jen"
)

func generateMockFuncCallArgsMethod(iface *wrappedInterface, method *wrappedMethod, outputImportPath string) jen.Code {
	if method.Variadic {
		return generateMockFuncCallArgsMethodVariadic(iface, method, outputImportPath)
	}

	return generateMockFuncCallArgsMethodNonVariadic(iface, method, outputImportPath)
}

func generateMockFuncCallArgsMethodNonVariadic(iface *wrappedInterface, method *wrappedMethod, outputImportPath string) jen.Code {
	commentText := `Args returns an interface slice containing the arguments of this invocation.`

	valueExpressions := make([]jen.Code, 0, len(method.Params))
	for i := range method.Params {
		valueExpressions = append(valueExpressions, jen.Id("c").Dot(fmt.Sprintf("Arg%d", i)))
	}
	returnStatement := jen.Return().Index().Interface().Values(valueExpressions...)

	results := []jen.Code{jen.Index().Interface()}
	return generateMockFuncCallMethod(iface, outputImportPath, method, "Args", commentText, nil, results,
		returnStatement, // return []interface{ c.Arg<n>, ... }
	)
}

func generateMockFuncCallArgsMethodVariadic(iface *wrappedInterface, method *wrappedMethod, outputImportPath string) jen.Code {
	commentText := strings.Join([]string{
		`Args returns an interface slice containing the arguments of this invocation.`,
		`The variadic slice argument is flattened in this array such that one positional argument and three variadic arguments would result in a slice of four, not two.`,
	}, " ")

	valueExpressions := make([]jen.Code, 0, len(method.Params))
	for i := range method.Params {
		valueExpressions = append(valueExpressions, jen.Id("c").Dot(fmt.Sprintf("Arg%d", i)))
	}

	lastIndex := len(valueExpressions) - 1
	trailingDeclaration := jen.Id("trailing").Op(":=").Index().Interface().Values()
	loopCondition := compose(jen.Id("_").Op(",").Id("val").Op(":=").Range(), valueExpressions[lastIndex])
	loopBody := selfAppend(jen.Id("trailing"), jen.Id("val"))
	loopStatement := jen.For(loopCondition).Block(loopBody)
	simpleValuesExpression := jen.Index().Interface().Values(valueExpressions[:lastIndex]...)
	returnStatement := jen.Return().Append(simpleValuesExpression, jen.Id("trailing").Op("..."))

	results := []jen.Code{jen.Index().Interface()}
	return generateMockFuncCallMethod(iface, outputImportPath, method, "Args", commentText, nil, results,
		trailingDeclaration,                   // trailingDeclaration := []interface{}
		loopStatement, jen.Line(), jen.Line(), // for _, val := range Arg<lastIndex> { trailing = append(trailing, val) }
		returnStatement, // return append([]interface{ <values>, ... }, trailing...)
	)
}

func generateMockFuncCallResultsMethod(iface *wrappedInterface, method *wrappedMethod, outputImportPath string) jen.Code {
	commentText := `Results returns an interface slice containing the results of this invocation.`

	values := make([]jen.Code, 0, len(method.Results))
	for i := range method.Results {
		values = append(values, jen.Id("c").Dot(fmt.Sprintf("Result%d", i)))
	}

	returnStatement := jen.Return().Index().Interface().Values(values...)

	results := []jen.Code{jen.Index().Interface()}
	return generateMockFuncCallMethod(iface, outputImportPath, method, "Results", commentText, nil, results,
		returnStatement, // return []interface{ c.Result<n>, ... }
	)
}

func generateMockFuncCallMethod(
	iface *wrappedInterface,
	outputImportPath string,
	method *wrappedMethod,
	methodName string,
	commentText string,
	params, results []jen.Code,
	body ...jen.Code,
) jen.Code {
	mockFuncCallStructName := fmt.Sprintf("%s%s%sFuncCall", iface.prefix, iface.titleName, method.Name)
	receiver := compose(jen.Id("c"), addTypes(jen.Id(mockFuncCallStructName), iface.TypeParams, outputImportPath, false))
	methodDeclaration := jen.Func().Params(receiver).Id(methodName).Params(params...).Params(results...).Block(body...)
	return addComment(methodDeclaration, 1, commentText)
}
