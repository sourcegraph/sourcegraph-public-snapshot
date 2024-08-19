package generation

import (
	"fmt"

	"github.com/dave/jennifer/jen"
)

func generateMockInterfaceMethod(iface *wrappedInterface, method *wrappedMethod, outputImportPath string) jen.Code {
	mockFuncFieldName := fmt.Sprintf("%sFunc", method.Name)
	mockFuncCallStructName := fmt.Sprintf("%s%s%sFuncCall", iface.prefix, iface.titleName, method.Name)
	commentText := fmt.Sprintf(
		`%s delegates to the next hook function in the queue and stores the parameter and result values of this invocation.`,
		method.Name,
	)

	paramNames := make([]jen.Code, 0, len(method.Params))
	argumentExpressions := make([]jen.Code, 0, len(method.Params))
	for i := 0; i < len(method.Params); i++ {
		name := fmt.Sprintf("v%d", i)

		nameExpression := jen.Id(name)
		if method.Variadic && i == len(method.Params)-1 {
			nameExpression = compose(nameExpression, jen.Op("..."))
		}

		paramNames = append(paramNames, jen.Id(name))
		argumentExpressions = append(argumentExpressions, nameExpression)
	}

	resultNames := make([]jen.Code, 0, len(method.Results))
	for i := 0; i < len(method.Results); i++ {
		resultNames = append(resultNames, jen.Id(fmt.Sprintf("r%d", i)))
	}

	functionExpression := jen.Id("m").Dot(mockFuncFieldName).Dot("nextHook").Call()
	callStatement := functionExpression.Call(argumentExpressions...)
	callInstanceExpression := compose(addTypes(jen.Id(mockFuncCallStructName), iface.TypeParams, outputImportPath, false), jen.Values(append(paramNames, resultNames...)...))
	appendFuncCall := jen.Id("m").Dot(mockFuncFieldName).Dot("appendCall").Call(callInstanceExpression)
	returnStatement := jen.Return()

	if len(method.Results) != 0 {
		assignmentTarget := jen.Id("r0")
		returnStatement = returnStatement.Id("r0")

		for i := 1; i < len(method.Results); i++ {
			assignmentTarget = assignmentTarget.Op(",").Id(fmt.Sprintf("r%d", i))
			returnStatement = returnStatement.Op(",").Id(fmt.Sprintf("r%d", i))
		}

		callStatement = compose(assignmentTarget.Op(":="), callStatement)
	}

	return generateMockMethod(iface, method, commentText, outputImportPath,
		callStatement,   // r<n>, ... := m.<MethodName>Func.nextHook()(Param<n>, ...)
		appendFuncCall,  // m.<MethodName>Func.appendCall(<InterfaceName><MethodName>FuncCall{Param<n>, ..., r<n>, ...})
		returnStatement, // return r<n>, ...
	)
}

func generateMockMethod(
	iface *wrappedInterface,
	method *wrappedMethod,
	commentText string,
	outputImportPath string,
	body ...jen.Code,
) jen.Code {
	params := make([]jen.Code, 0, len(method.paramTypes))
	for i, param := range method.paramTypes {
		params = append(params, compose(jen.Id(fmt.Sprintf("v%d", i)), param))
	}

	receiver := compose(jen.Id("m").Op("*"), addTypes(jen.Id(iface.mockStructName), iface.TypeParams, outputImportPath, false))
	methodDeclaration := jen.Func().Params(receiver).Id(method.Name).Params(params...).Params(method.resultTypes...).Block(body...)
	return addComment(methodDeclaration, 1, commentText)
}
