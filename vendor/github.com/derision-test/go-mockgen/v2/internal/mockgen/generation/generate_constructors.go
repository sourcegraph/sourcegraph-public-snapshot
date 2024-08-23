package generation

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/dave/jennifer/jen"
	"github.com/derision-test/go-mockgen/v2/internal/mockgen/types"
)

func generateMockStructConstructor(iface *wrappedInterface, constructorPrefix, outputImportPath string) jen.Code {
	makeField := func(method *wrappedMethod) jen.Code {
		return makeDefaultHookField(iface, method, outputImportPath, generateNoopFunction(iface, method, outputImportPath))
	}

	name := fmt.Sprintf("New%s%s", constructorPrefix, iface.mockStructName)
	commentText := []string{
		fmt.Sprintf(`%s creates a new mock of the %s interface.`, name, iface.Name),
		`All methods return zero values for all results, unless overwritten.`,
	}
	return generateConstructor(iface, strings.Join(commentText, " "), name, nil, outputImportPath, makeField)
}

func generateMockStructStrictConstructor(iface *wrappedInterface, constructorPrefix, outputImportPath string) jen.Code {
	makeField := func(method *wrappedMethod) jen.Code {
		return makeDefaultHookField(iface, method, outputImportPath, generatePanickingFunction(iface, method, outputImportPath))
	}

	name := fmt.Sprintf("NewStrict%s%s", constructorPrefix, iface.mockStructName)
	commentText := []string{
		fmt.Sprintf(`%s creates a new mock of the %s interface.`, name, iface.Name),
		`All methods panic on invocation, unless overwritten.`,
	}
	return generateConstructor(iface, strings.Join(commentText, " "), name, nil, outputImportPath, makeField)
}

func generateMockStructFromConstructor(iface *wrappedInterface, constructorPrefix, outputImportPath string) jen.Code {
	if !unicode.IsUpper([]rune(iface.Name)[0]) {
		surrogateStructName := fmt.Sprintf("surrogateMock%s", iface.titleName)
		surrogateDefinition := generateSurrogateInterface(iface, surrogateStructName, outputImportPath)
		name := jen.Id(surrogateStructName)
		constructor := generateMockStructFromConstructorCommon(iface, name, constructorPrefix, outputImportPath)
		return compose(surrogateDefinition, constructor)
	}

	importPath := sanitizeImportPath(iface.ImportPath, outputImportPath)
	name := jen.Qual(importPath, iface.Name)
	return generateMockStructFromConstructorCommon(iface, name, constructorPrefix, outputImportPath)
}

func generateMockStructFromConstructorCommon(iface *wrappedInterface, ifaceName *jen.Statement, constructorPrefix, outputImportPath string) jen.Code {
	makeField := func(method *wrappedMethod) jen.Code {
		// i.<MethodName>
		return makeDefaultHookField(iface, method, outputImportPath, jen.Id("i").Dot(method.Name))
	}

	name := fmt.Sprintf("New%s%sFrom", constructorPrefix, iface.mockStructName)
	commentText := []string{
		fmt.Sprintf(`%s creates a new mock of the %s interface.`, name, iface.mockStructName),
		`All methods delegate to the given implementation, unless overwritten.`,
	}

	// (i <InterfaceName>)
	params := []jen.Code{compose(jen.Id("i"), addTypes(ifaceName, iface.TypeParams, outputImportPath, false))}
	return generateConstructor(iface, strings.Join(commentText, " "), name, params, outputImportPath, makeField)
}

func generateConstructor(
	iface *wrappedInterface,
	commentText string,
	methodName string,
	params []jen.Code,
	outputImportPath string,
	makeField func(method *wrappedMethod) jen.Code,
) jen.Code {
	constructorFields := make([]jen.Code, 0, len(iface.Methods))
	for _, method := range iface.wrappedMethods {
		constructorFields = append(constructorFields, makeField(method))
	}

	// return &Mock<Name>{ <constructorField>, ... }
	returnStatement := compose(jen.Return(), generateStructInitializer(iface.mockStructName, outputImportPath, iface.TypeParams, constructorFields...))
	results := []jen.Code{addTypes(jen.Op("*").Id(iface.mockStructName), iface.TypeParams, outputImportPath, false)}
	functionDeclaration := compose(addTypes(jen.Func().Id(methodName), iface.TypeParams, outputImportPath, true), jen.Params(params...).Params(results...).Block(returnStatement))
	return addComment(functionDeclaration, 1, commentText)
}

func generateNoopFunction(iface *wrappedInterface, method *wrappedMethod, outputImportPath string) jen.Code {
	rt := make([]jen.Code, 0, len(method.resultTypes))
	for i, resultType := range method.resultTypes {
		// (r0 <typ1>, r1 <type2>, ...)
		rt = append(rt, compose(jen.Id(fmt.Sprintf("r%d", i)), resultType))
	}

	// Note: an empty return here returns the zero valued variables r0, r1, ...
	return jen.Func().Params(method.paramTypes...).Params(rt...).Block(jen.Return())
}

func generatePanickingFunction(iface *wrappedInterface, method *wrappedMethod, outputImportPath string) jen.Code {
	// panic("unexpected invocation of <Struct>.<Method>")
	panicStatement := jen.Panic(jen.Lit(fmt.Sprintf("unexpected invocation of %s.%s", iface.mockStructName, method.Method.Name)))
	return jen.Func().Params(method.paramTypes...).Params(method.resultTypes...).Block(panicStatement)
}

func generateSurrogateInterface(iface *wrappedInterface, surrogateName, outputImportPath string) *jen.Statement {
	surrogateCommentText := strings.Join([]string{
		fmt.Sprintf(`%s is a copy of the %s interface (from the package %s).`, surrogateName, iface.Name, iface.ImportPath),
		`It is redefined here as it is unexported in the source package.`,
	}, " ")

	signatures := make([]jen.Code, 0, len(iface.wrappedMethods))
	for _, method := range iface.wrappedMethods {
		signatures = append(signatures, jen.Id(method.Name).Params(method.paramTypes...).Params(method.resultTypes...))
	}

	// type <SurrogateName> interface { <MethodName>(<Param #n>, ...) (<Result #n>, ...), ... }
	typeDeclaration := addTypes(jen.Type().Id(surrogateName), iface.Interface.TypeParams, outputImportPath, true).Interface(signatures...).Line()
	return addComment(typeDeclaration, 1, surrogateCommentText)
}

func makeDefaultHookField(iface *wrappedInterface, method *wrappedMethod, outputImportPath string, function jen.Code) jen.Code {
	fieldName := fmt.Sprintf("%sFunc", method.Name)
	structName := fmt.Sprintf("%s%s%sFunc", iface.prefix, iface.titleName, method.Name)

	initializer := generateStructInitializer(structName, outputImportPath, iface.TypeParams, compose(
		jen.Id("defaultHook").Op(":"),
		function,
	))

	// <fieldName>: &StructName{ defaultHook: <Function> }
	return compose(jen.Id(fieldName), jen.Op(":"), initializer)
}

func generateStructInitializer(structName string, outputImportPath string, typeParams []types.TypeParam, fields ...jen.Code) jen.Code {
	// &<StructName>{ fields, ... }
	return compose(addTypes(jen.Op("&").Id(structName), typeParams, outputImportPath, false), jen.Values(padFields(fields)...))
}

func padFields(fields []jen.Code) []jen.Code {
	paddedFields := make([]jen.Code, 0, len(fields)+1)
	for _, field := range fields {
		paddedFields = append(paddedFields, compose(jen.Line(), field))
	}

	return append(paddedFields, jen.Line())
}
