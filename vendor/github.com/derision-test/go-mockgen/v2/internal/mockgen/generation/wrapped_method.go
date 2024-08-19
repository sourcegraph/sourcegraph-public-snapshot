package generation

import (
	"github.com/dave/jennifer/jen"
	"github.com/derision-test/go-mockgen/v2/internal/mockgen/types"
)

type wrappedMethod struct {
	*types.Method
	iface             *types.Interface
	dotlessParamTypes []jen.Code
	paramTypes        []jen.Code
	resultTypes       []jen.Code
	signature         jen.Code
}

func wrapMethod(iface *types.Interface, method *types.Method, outputImportPath string) *wrappedMethod {
	m := &wrappedMethod{
		Method:            method,
		iface:             iface,
		dotlessParamTypes: generateParamTypes(method, iface.ImportPath, outputImportPath, true),
		paramTypes:        generateParamTypes(method, iface.ImportPath, outputImportPath, false),
		resultTypes:       generateResultTypes(method, iface.ImportPath, outputImportPath),
	}

	m.signature = jen.Func().Params(m.paramTypes...).Params(m.resultTypes...)
	return m
}

func generateParamTypes(method *types.Method, importPath, outputImportPath string, omitDots bool) []jen.Code {
	params := make([]jen.Code, 0, len(method.Params))
	for i, typ := range method.Params {
		params = append(params, generateType(
			typ,
			importPath,
			outputImportPath,
			method.Variadic && i == len(method.Params)-1 && !omitDots,
		))
	}

	return params
}

func generateResultTypes(method *types.Method, importPath, outputImportPath string) []jen.Code {
	results := make([]jen.Code, 0, len(method.Results))
	for _, typ := range method.Results {
		results = append(results, generateType(
			typ,
			importPath,
			outputImportPath,
			false,
		))
	}

	return results
}
