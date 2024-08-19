package valast

import (
	"fmt"
	"go/ast"
	"go/token"
	"reflect"

	"github.com/hexops/valast/internal/customtype"
)

// RegisterType registers a type that for representation in a custom manner with
// valast. If valast encounters a value or pointer to a value of this type, it
// will use the given render func to generate the appropriate AST representation.
//
// This is useful if a type's fields are private, and can only be represented
// through a constructor - see stdtypes.go for examples.
//
// This mechanism currently only works with struct types.
func RegisterType[T any](render func(value T) ast.Expr) {
	customtype.Register(render)
}

type cacheKeyOptions struct {
	Unqualify    bool
	PackagePath  string
	PackageName  string
	ExportedOnly bool
}

type cacheKey struct {
	v   reflect.Type
	opt cacheKeyOptions
}

func newCacheKey(v reflect.Type, opt *Options) cacheKey {
	return cacheKey{v: v, opt: cacheKeyOptions{
		Unqualify:    opt.Unqualify,
		PackagePath:  opt.PackagePath,
		PackageName:  opt.PackageName,
		ExportedOnly: opt.ExportedOnly,
	}}
}

type typeExprCache map[cacheKey]Result

// typeExpr returns an AST type expression for the value v.
//
// It is cached to avoid building type expressions again for types we've already seen, which can
// get quite complex (see BenchmarkComplexType.)
func typeExpr(v reflect.Type, opt *Options, cache typeExprCache) (Result, error) {
	key := newCacheKey(v, opt)
	if cached, ok := cache[key]; ok {
		return cached, nil
	}

	result, err := uncachedTypeExpr(v, opt, cache)
	if err != nil {
		return Result{}, err
	}
	cache[key] = result
	return result, nil
}

func uncachedTypeExpr(v reflect.Type, opt *Options, cache typeExprCache) (Result, error) {
	if v.Kind() != reflect.UnsafePointer && v.Name() != "" {
		pkgPath := v.PkgPath()
		if pkgPath != "" && pkgPath != opt.PackagePath {
			pkgName, err := opt.packagePathToName(v.PkgPath())
			if err != nil {
				return Result{}, err
			}
			if pkgName != opt.PackageName {
				return Result{
					AST:                &ast.SelectorExpr{X: ast.NewIdent(pkgName), Sel: ast.NewIdent(v.Name())},
					RequiresUnexported: !ast.IsExported(v.Name()),
				}, nil
			}
		}
		return Result{
			AST:                ast.NewIdent(v.Name()),
			RequiresUnexported: false,
		}, nil
	}
	switch v.Kind() {
	case reflect.Array:
		elemType, err := typeExpr(v.Elem(), opt, cache)
		if err != nil {
			return Result{}, err
		}
		return Result{
			AST: &ast.ArrayType{
				Len: &ast.BasicLit{Kind: token.INT, Value: fmt.Sprint(v.Len())},
				Elt: elemType.AST,
			},
			RequiresUnexported: elemType.RequiresUnexported,
		}, nil
	case reflect.Interface:
		var methods []*ast.Field
		var requiresUnexported bool
		for i := 0; i < v.NumMethod(); i++ {
			method := v.Method(i)
			methodType, err := typeExpr(method.Type, opt, cache)
			if err != nil {
				return Result{}, err
			}
			if methodType.RequiresUnexported {
				requiresUnexported = true
			}
			methods = append(methods, &ast.Field{
				Names: []*ast.Ident{ast.NewIdent(method.Name)},
				Type:  methodType.AST,
			})
		}
		return Result{
			AST:                &ast.InterfaceType{Methods: &ast.FieldList{List: methods}},
			RequiresUnexported: requiresUnexported,
		}, nil
	case reflect.Func:
		// Note: reflect cannot determine parameter/result names. See https://groups.google.com/g/golang-nuts/c/nM_ZhL7fuGc
		var (
			requiresUnexported bool
			params             []*ast.Field
		)
		for i := 0; i < v.NumIn(); i++ {
			param := v.In(i)
			paramType, err := typeExpr(param, opt, cache)
			if err != nil {
				return Result{}, err
			}
			if paramType.RequiresUnexported {
				requiresUnexported = true
			}
			params = append(params, &ast.Field{
				Type: paramType.AST,
			})
		}
		var results []*ast.Field
		for i := 0; i < v.NumOut(); i++ {
			result := v.Out(i)
			resultType, err := typeExpr(result, opt, cache)
			if err != nil {
				return Result{}, err
			}
			if resultType.RequiresUnexported {
				requiresUnexported = true
			}
			results = append(results, &ast.Field{
				Type: resultType.AST,
			})
		}
		return Result{
			AST: &ast.FuncType{
				Params:  &ast.FieldList{List: params},
				Results: &ast.FieldList{List: results},
			},
			RequiresUnexported: requiresUnexported,
		}, nil
	case reflect.Map:
		keyType, err := typeExpr(v.Key(), opt, cache)
		if err != nil {
			return Result{}, err
		}
		valueType, err := typeExpr(v.Elem(), opt, cache)
		if err != nil {
			return Result{}, err
		}
		return Result{
			AST: &ast.MapType{
				Key:   keyType.AST,
				Value: valueType.AST,
			},
			RequiresUnexported: keyType.RequiresUnexported || valueType.RequiresUnexported,
		}, nil
	case reflect.Ptr:
		ptrType, err := typeExpr(v.Elem(), opt, cache)
		if err != nil {
			return Result{}, err
		}
		return Result{
			AST:                &ast.StarExpr{X: ptrType.AST},
			RequiresUnexported: ptrType.RequiresUnexported,
		}, nil
	case reflect.Slice:
		elemType, err := typeExpr(v.Elem(), opt, cache)
		if err != nil {
			return Result{}, err
		}
		return Result{
			AST:                &ast.ArrayType{Elt: elemType.AST},
			RequiresUnexported: elemType.RequiresUnexported,
		}, nil
	case reflect.Struct:
		var (
			fields                                []*ast.Field
			requiresUnexported, omittedUnexported bool
		)
		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)
			fieldType, err := typeExpr(field.Type, opt, cache)
			if err != nil {
				return Result{}, err
			}
			if fieldType.RequiresUnexported {
				requiresUnexported = true
				if opt.ExportedOnly {
					return Result{RequiresUnexported: true}, nil
				}
			}
			if fieldType.OmittedUnexported {
				omittedUnexported = true
			}
			fields = append(fields, &ast.Field{
				Names: []*ast.Ident{ast.NewIdent(field.Name)},
				Type:  fieldType.AST,
			})
		}
		return Result{
			AST: &ast.StructType{
				Fields: &ast.FieldList{List: fields},
			},
			RequiresUnexported: requiresUnexported,
			OmittedUnexported:  omittedUnexported,
		}, nil
	case reflect.UnsafePointer:
		// Note: For a plain unsafe.Pointer type, v.PkgPath() does not report "unsafe" but rather
		// an empty string "".
		isPlainUnsafePointer := v.String() == "unsafe.Pointer"
		if !isPlainUnsafePointer && v.Name() != "" {
			pkgPath := v.PkgPath()
			if pkgPath != "" && pkgPath != opt.PackagePath {
				pkgName, err := opt.packagePathToName(v.PkgPath())
				if err != nil {
					return Result{}, err
				}
				if pkgName != opt.PackageName {
					return Result{
						AST:                &ast.SelectorExpr{X: ast.NewIdent(pkgName), Sel: ast.NewIdent(v.Name())},
						RequiresUnexported: !ast.IsExported(v.Name()),
					}, nil
				}
			}
			return Result{
				AST:                ast.NewIdent(v.Name()),
				RequiresUnexported: false,
			}, nil
		}
		return Result{AST: &ast.SelectorExpr{X: ast.NewIdent("unsafe"), Sel: ast.NewIdent("Pointer")}}, nil
	default:
		return Result{AST: ast.NewIdent(v.Name())}, nil
	}
}
