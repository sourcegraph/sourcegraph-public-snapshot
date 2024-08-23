package types

import (
	"go/ast"
	"go/types"
	"sort"
)

type Interface struct {
	Name       string
	ImportPath string
	TypeParams []TypeParam
	Methods    []*Method

	// Prefix is set on extraction based on the current PackageOptions
	Prefix string
}

type TypeParam struct {
	Name string
	Type types.Type
}

func newInterfaceFromTypeSpec(name, importPath string, typeSpec *ast.TypeSpec, underlyingType *types.Interface, ps *types.TypeParamList) *Interface {
	methodMap := make(map[string]*Method, underlyingType.NumMethods())
	for i := 0; i < underlyingType.NumMethods(); i++ {
		method := underlyingType.Method(i)
		name := method.Name()
		methodMap[name] = newMethodFromSignature(name, method.Type().(*types.Signature))
	}

	methodNames := make([]string, 0, len(methodMap))
	for k := range methodMap {
		methodNames = append(methodNames, k)
	}
	sort.Strings(methodNames)

	methods := make([]*Method, 0, len(methodNames))
	for _, name := range methodNames {
		methods = append(methods, methodMap[name])
	}

	var typeParams []TypeParam
	if typeSpec.TypeParams != nil && ps != nil {
		for i, field := range typeSpec.TypeParams.List {
			for _, name := range field.Names {
				typeParams = append(typeParams, TypeParam{Name: name.Name, Type: ps.At(i).Constraint()})
			}
		}
	}

	return &Interface{
		Name:       name,
		ImportPath: importPath,
		TypeParams: typeParams,
		Methods:    methods,
	}
}
