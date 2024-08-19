package valast

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"io"
	"math"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/hexops/valast/internal/bypass"
	"github.com/hexops/valast/internal/customtype"
	"golang.org/x/tools/go/packages"
	gofumpt "mvdan.cc/gofumpt/format"
)

// Options describes options for the conversion process.
type Options struct {
	// Unqualify, if true, indicates that types should be unqualified. e.g.:
	//
	// 	int(8)           -> 8
	// 	Bar{}            -> Bar{}
	// 	string("foobar") -> "foobar"
	//
	// This is set to true automatically when operating within a context where type qualification
	// is definitively not needed, e.g. when producing values for a struct or map.
	Unqualify bool

	// PackagePath, if non-zero, describes that the literal is being produced within the described
	// package path, and thus type selectors `pkg.Foo` should just be written `Foo` if the package
	// path and name match.
	PackagePath string

	// PackageName, if non-zero, describes that the literal is being produced within the described
	// package name, and thus type selectors `pkg.Foo` should just be written `Foo` if the package
	// path and name match.
	PackageName string

	// ExportedOnly indicates if only exported fields and values should be included.
	ExportedOnly bool

	// PackagePathToName, if non-nil, is called to convert a Go package path to the package name
	// written in its source. The default is DefaultPackagePathToName
	PackagePathToName func(path string) (string, error)
}

func (o *Options) withUnqualify() *Options {
	tmp := *o
	tmp.Unqualify = true
	return &tmp
}

func (o *Options) packagePathToName(path string) (string, error) {
	if o.PackagePathToName != nil {
		return o.PackagePathToName(path)
	}
	return DefaultPackagePathToName(path)
}

// DefaultPackagePathToName loads the specified package from disk to determine the package name.
func DefaultPackagePathToName(path string) (string, error) {
	pkgs, err := packages.Load(&packages.Config{Mode: packages.NeedName}, path)
	if err != nil {
		return "", err
	}
	return pkgs[0].Name, nil
}

// String converts the value v into the equivalent Go literal syntax.
//
// It is an opinionated helper for the more extensive AST function.
//
// If any error occurs, it will be returned as the string value. If handling errors is desired then
// consider using the AST function directly.
func String(v interface{}) string {
	return StringWithOptions(v, nil)
}

// StringWithOptions converts the value v into the equivalent Go literal syntax, with the specified
// options.
//
// It is an opinionated helper for the more extensive AST function.
//
// If any error occurs, it will be returned as the string value. If handling errors is desired then
// consider using the AST function directly.
func StringWithOptions(v interface{}, opt *Options) string {
	if opt == nil {
		opt = &Options{}
	}
	var buf bytes.Buffer
	result, err := AST(reflect.ValueOf(v), opt)
	if err != nil {
		return err.Error()
	}
	if opt.ExportedOnly && result.RequiresUnexported {
		return fmt.Sprintf("valast: cannot convert unexported value %T", v)
	}
	if err := gofumptFormatExpr(&buf, token.NewFileSet(), result.AST, gofumpt.Options{
		ExtraRules: true,
	}); err != nil {
		return fmt.Sprintf("valast: format: %v", err)
	}
	return buf.String()
}

// gofumptFormatExpr is a slight hack to get gofumpt to format an ast.Expr node, because the
// gofumpt/format package does not expose node-level formatting currently.
func gofumptFormatExpr(w io.Writer, fset *token.FileSet, expr ast.Expr, opt gofumpt.Options) error {
	// First use go/format to convert the expression to Go syntax.
	var tmp bytes.Buffer
	if err := format.Node(&tmp, fset, expr); err != nil {
		return err
	}

	// HACK: Split composite literals onto multiple lines to avoid extra long struct values. We
	// will defer this to gofumpt once it can perform this: https://github.com/mvdan/gofumpt/pull/70
	tmpString := string(formatCompositeLiterals([]rune(tmp.String())))

	// Create a temporary file with our expression, run gofumpt on it, and extract the result.
	fileStart := `package main

func main() {
	v := `
	fileEnd := `
}
`
	tmpFile := []byte(fileStart + tmpString + fileEnd)
	formattedFile, err := gofumpt.Source(tmpFile, opt)
	if err != nil {
		return err
	}
	formattedFile = bytes.TrimPrefix(formattedFile, []byte(fileStart))
	formattedFile = bytes.TrimSuffix(formattedFile, []byte(fileEnd))

	// Remove leading indention.
	lines := bytes.Split(formattedFile, []byte{'\n'})
	for i, line := range lines {
		lines[i] = bytes.TrimPrefix(line, []byte{'\t'})
	}
	formattedExpr := bytes.Join(lines, []byte{'\n'})
	_, err = w.Write(formattedExpr)
	return err
}

// DEPRECATED: use valast.Ptr instead.
//
// Addr returns a pointer to the given value.
//
// It is the only way to create a reference to certain values within a Go expression,
// for example since &"hello" is illegal, it can instead be written in a single expression as:
//
//	valast.Addr("hello").(*string)
func Addr(v interface{}) interface{} {
	vv := reflect.ValueOf(v)

	// Create a slice with v in it so that we have an addressable value.
	sliceType := reflect.SliceOf(vv.Type())
	slice := reflect.MakeSlice(sliceType, 1, 1)
	if v != nil {
		slice.Index(0).Set(vv)
	}
	return slice.Index(0).Addr().Interface()
}

// AddrInterface returns a pointer to the given interface value, which is determined to be of type
// T. For example, since &MyInterface(MyValue{}) is illegal, it can instead be written in a single
// expression as:
//
//	valast.AddrInterface(&MyValue{}, (*MyInterface)(nil))
//
// The second parameter should be a pointer to the interface type. This is needed because
// reflect.ValueOf(&v).Type() returns *MyValue not MyInterface, due to reflect.ValueOf taking an
// interface{} parameter and losing that type information.
func AddrInterface(v, pointerToType interface{}) interface{} {
	// Create a slice with v in it so that we have an addressable value.
	sliceType := reflect.SliceOf(reflect.TypeOf(pointerToType).Elem())
	slice := reflect.MakeSlice(sliceType, 1, 1)
	if v != nil {
		slice.Index(0).Set(reflect.ValueOf(v))
	}
	return slice.Index(0).Addr().Interface()
}

func basicLit(vv reflect.Value, kind token.Token, builtinType string, v interface{}, opt *Options, typeExprCache typeExprCache) (Result, error) {
	typeExpr, err := typeExpr(vv.Type(), opt, typeExprCache)
	if err != nil {
		return Result{}, err
	}
	if opt.Unqualify && vv.Type().Name() == builtinType && vv.Type().PkgPath() == "" {
		return Result{AST: ast.NewIdent(fmt.Sprint(v))}, nil
	}
	if opt.ExportedOnly && typeExpr.RequiresUnexported {
		return Result{RequiresUnexported: true}, nil
	}
	return Result{
		AST: &ast.CallExpr{
			Fun:  typeExpr.AST,
			Args: []ast.Expr{ast.NewIdent(fmt.Sprint(v))},
		},
		RequiresUnexported: typeExpr.RequiresUnexported,
	}, nil
}

// ErrInvalidType describes that the value is of a type that cannot be converted to an AST.
type ErrInvalidType struct {
	// Value is the actual value that was being converted.
	Value interface{}
}

// Error implements the error interface.
func (e *ErrInvalidType) Error() string {
	return fmt.Sprintf("valast: cannot convert value of type %T", e.Value)
}

// Result is a result from converting a Go value into its AST.
type Result struct {
	// AST is the actual Go AST expression for the value.
	//
	// If Options.ExportedOnly == true, and the input value was unexported this field will be nil.
	AST ast.Expr

	// OmittedUnexported indicates if unexported fields were omitted or not. Only indicative if
	// Options.ExportedOnly == true.
	OmittedUnexported bool

	// RequiresUnexported indicates if the AST requires access to unexported types/values outside
	// of the package specified in the Options, and is thus invalid code.
	RequiresUnexported bool

	// Packages is the list of packages that are used in the AST.
	Packages []string
}

// AST converts the given value into its equivalent Go AST expression.
//
// The input must be one of these kinds:
//
//	bool
//	int, int8, int16, int32, int64
//	uint, uint8, uint16, uint32, uint64
//	uintptr
//	float32, float64
//	complex64, complex128
//	array
//	interface
//	map
//	ptr
//	slice
//	string
//	struct
//	unsafe pointer
//
// The input type is reflect.Value instead of interface{}, specifically to allow converting
// interfaces derived from struct fields or other reflection which would otherwise be lost if the
// input type is interface{}.
//
// Cyclic data structures will have their cyclic pointer values emitted twice, followed by a nil
// value. e.g. for a structure `foo` with field `bar` which points to the original `foo`:
//
//	&foo{id: 123, bar: &foo{id: 123, bar: nil}}
func AST(v reflect.Value, opt *Options) (Result, error) {
	var prof *profiler
	wantProfile, _ := strconv.ParseBool(os.Getenv("VALAST_PROFILE"))
	if wantProfile {
		prof = &profiler{}
	}
	packagesFound := make(map[string]bool)
	r, err := computeASTProfiled(v, opt, &cycleDetector{}, prof, typeExprCache{}, packagesFound)
	prof.dump()

	for k := range packagesFound {
		if k != "" {
			r.Packages = append(r.Packages, k)
		}
	}
	sort.Strings(r.Packages)

	return r, err
}

func computeASTProfiled(v reflect.Value, opt *Options, cycleDetector *cycleDetector, profiler *profiler, typeExprCache typeExprCache, packagesFound map[string]bool) (Result, error) {
	profiler.push(v)
	start := time.Now()
	r, err := computeAST(v, opt, cycleDetector, profiler, typeExprCache, packagesFound)
	profiler.pop(start)
	return r, err
}

func computeAST(v reflect.Value, opt *Options, cycleDetector *cycleDetector, profiler *profiler, typeExprCache typeExprCache, packagesFound map[string]bool) (Result, error) {
	if opt == nil {
		opt = &Options{}
	}
	if v == (reflect.Value{}) {
		// Technically this is an invalid reflect.Value, but we handle it to be gracious in the
		// case of:
		//
		//  var x interface{}
		// 	valast.AST(reflect.ValueOf(x))
		//
		return Result{
			AST: ast.NewIdent("nil"),
		}, nil
	}

	vv := unexported(v)
	packagesFound[vv.Type().PkgPath()] = true
	switch vv.Kind() {
	case reflect.Bool:
		boolType, err := typeExpr(vv.Type(), opt, typeExprCache)
		if err != nil {
			return Result{}, err
		}
		if vv.Type().Name() == "bool" && vv.Type().PkgPath() == "" {
			return Result{AST: ast.NewIdent(fmt.Sprint(v))}, nil
		}
		if opt.ExportedOnly && boolType.RequiresUnexported {
			return Result{RequiresUnexported: true}, nil
		}
		return Result{
			AST: &ast.CallExpr{
				Fun:  boolType.AST,
				Args: []ast.Expr{ast.NewIdent(fmt.Sprint(v))},
			},
			RequiresUnexported: boolType.RequiresUnexported,
		}, nil
	case reflect.Int:
		return basicLit(vv, token.INT, "int", v, opt, typeExprCache)
	case reflect.Int8:
		return basicLit(vv, token.INT, "int8", v, opt, typeExprCache)
	case reflect.Int16:
		return basicLit(vv, token.INT, "int16", v, opt, typeExprCache)
	case reflect.Int32:
		return basicLit(vv, token.INT, "int32", v, opt, typeExprCache)
	case reflect.Int64:
		return basicLit(vv, token.INT, "int64", v, opt, typeExprCache)
	case reflect.Uint:
		return basicLit(vv, token.INT, "uint", v, opt, typeExprCache)
	case reflect.Uint8:
		return basicLit(vv, token.INT, "uint8", v, opt, typeExprCache)
	case reflect.Uint16:
		return basicLit(vv, token.INT, "uint16", v, opt, typeExprCache)
	case reflect.Uint32:
		return basicLit(vv, token.INT, "uint32", v, opt, typeExprCache)
	case reflect.Uint64:
		return basicLit(vv, token.INT, "uint64", v, opt, typeExprCache)
	case reflect.Uintptr:
		return basicLit(vv, token.INT, "uintptr", v, opt, typeExprCache)
	case reflect.Float32:
		return basicLit(vv, token.FLOAT, "float32", v, opt, typeExprCache)
	case reflect.Float64:
		return basicLit(vv, token.FLOAT, "float64", v, opt, typeExprCache)
	case reflect.Complex64:
		return basicLit(vv, token.FLOAT, "complex64", v, opt, typeExprCache)
	case reflect.Complex128:
		return basicLit(vv, token.FLOAT, "complex128", v, opt, typeExprCache)
	case reflect.Array:
		var (
			elts               []ast.Expr
			requiresUnexported bool
		)
		for i := 0; i < vv.Len(); i++ {
			elem, err := computeASTProfiled(vv.Index(i), opt.withUnqualify(), cycleDetector, profiler, typeExprCache, packagesFound)
			if err != nil {
				return Result{}, err
			}
			if elem.RequiresUnexported {
				requiresUnexported = true
			}
			elts = append(elts, elem.AST)
		}
		arrayType, err := typeExpr(vv.Type(), opt, typeExprCache)
		if err != nil {
			return Result{}, err
		}
		return Result{
			AST: &ast.CompositeLit{
				Type: arrayType.AST,
				Elts: elts,
			},
			RequiresUnexported: arrayType.RequiresUnexported || requiresUnexported,
		}, nil
	case reflect.Interface:
		if opt.ExportedOnly && !ast.IsExported(vv.Type().Name()) {
			return Result{
				AST:                nil,
				RequiresUnexported: true,
			}, nil
		}
		if opt.Unqualify {
			return computeASTProfiled(unexported(vv.Elem()), opt.withUnqualify(), cycleDetector, profiler, typeExprCache, packagesFound)
		}
		v, err := computeASTProfiled(unexported(vv.Elem()), opt, cycleDetector, profiler, typeExprCache, packagesFound)
		if err != nil {
			return Result{}, err
		}
		interfaceType, err := typeExpr(vv.Type(), opt, typeExprCache)
		if err != nil {
			return Result{}, err
		}
		return Result{
			AST: &ast.CompositeLit{
				Type: interfaceType.AST,
				Elts: []ast.Expr{v.AST},
			},
			RequiresUnexported: interfaceType.RequiresUnexported || v.RequiresUnexported,
		}, nil
	case reflect.Map:
		var (
			keyValueExprs                         []ast.Expr
			requiresUnexported, omittedUnexported bool
			keys                                  = vv.MapKeys()
		)
		sort.Slice(keys, func(i, j int) bool {
			return valueLess(keys[i], keys[j])
		})
		for _, key := range keys {
			value := vv.MapIndex(key)
			k, err := computeASTProfiled(key, opt.withUnqualify(), cycleDetector, profiler, typeExprCache, packagesFound)
			if err != nil {
				return Result{}, err
			}
			if k.RequiresUnexported {
				if opt.ExportedOnly {
					omittedUnexported = true
					continue
				}
				requiresUnexported = true
			}
			if k.OmittedUnexported {
				omittedUnexported = true
			}
			v, err := computeASTProfiled(value, opt.withUnqualify(), cycleDetector, profiler, typeExprCache, packagesFound)
			if err != nil {
				return Result{}, err
			}
			if v.RequiresUnexported {
				if opt.ExportedOnly {
					omittedUnexported = true
					continue
				}
				requiresUnexported = true
			}
			if v.OmittedUnexported {
				omittedUnexported = true
			}
			keyValueExprs = append(keyValueExprs, &ast.KeyValueExpr{
				Key:   k.AST,
				Value: v.AST,
			})
		}
		mapType, err := typeExpr(vv.Type(), opt, typeExprCache)
		if err != nil {
			return Result{}, err
		}
		return Result{
			AST: &ast.CompositeLit{
				Type: mapType.AST,
				Elts: keyValueExprs,
			},
			RequiresUnexported: requiresUnexported || mapType.RequiresUnexported,
			OmittedUnexported:  omittedUnexported,
		}, nil
	case reflect.Ptr:
		ptrType, err := typeExpr(vv.Type(), opt, typeExprCache)
		if err != nil {
			return Result{}, err
		}
		isPtrToInterface := vv.Elem().Kind() == reflect.Interface
		if !isPtrToInterface && vv.IsNil() {
			if opt.Unqualify {
				return Result{AST: ast.NewIdent("nil")}, nil
			}
			return Result{
				AST: &ast.CallExpr{
					Fun:  &ast.ParenExpr{X: ptrType.AST},
					Args: []ast.Expr{ast.NewIdent("nil")},
				},
				RequiresUnexported: ptrType.RequiresUnexported,
			}, nil
		}
		if opt.ExportedOnly && ptrType.RequiresUnexported {
			return Result{RequiresUnexported: true}, nil
		}
		if cycleDetector.push(vv.Interface()) {
			// cyclic data structure detected
			return Result{AST: ast.NewIdent("nil")}, nil
		}

		if !isPtrToInterface && !isAddressableKind(vv.Elem().Kind()) {
			if opt.Unqualify && literalNeedsQualification(vv.Elem()) {
				opt.Unqualify = false // the value must have qualification
			}
			elem, err := computeASTProfiled(vv.Elem(), opt, cycleDetector, profiler, typeExprCache, packagesFound)
			if err != nil {
				return Result{}, err
			}
			cycleDetector.pop(vv.Interface())

			// Pointers to unaddressable values can be created with help from valast.Addr.
			packagesFound["github.com/hexops/valast"] = true
			return Result{
				AST: &ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X:   ast.NewIdent("valast"),
						Sel: ast.NewIdent("Ptr"),
					},
					Args: []ast.Expr{elem.AST},
				},
				RequiresUnexported: ptrType.RequiresUnexported || elem.RequiresUnexported,
				OmittedUnexported:  elem.OmittedUnexported,
			}, nil
		}

		elem, err := computeASTProfiled(vv.Elem(), opt, cycleDetector, profiler, typeExprCache, packagesFound)
		if err != nil {
			return Result{}, err
		}
		cycleDetector.pop(vv.Interface())
		if isPtrToInterface {
			// Pointers to interfaces can be created with help from valast.AddrInterface.
			return Result{
				AST: &ast.TypeAssertExpr{
					X: &ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X:   ast.NewIdent("valast"),
							Sel: ast.NewIdent("AddrInterface"),
						},
						Args: []ast.Expr{
							elem.AST,
							&ast.CallExpr{
								Fun:  &ast.ParenExpr{X: ptrType.AST},
								Args: []ast.Expr{ast.NewIdent("nil")},
							},
						},
					},
					Type: ptrType.AST,
				},
				RequiresUnexported: ptrType.RequiresUnexported || elem.RequiresUnexported,
				OmittedUnexported:  elem.OmittedUnexported,
			}, nil
		}
		if vv.Elem().Kind() == reflect.Ptr {
			// Pointers to pointers can be created with help from valast.Addr.
			return Result{
				AST: &ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X:   ast.NewIdent("valast"),
						Sel: ast.NewIdent("Ptr"),
					},
					Args: []ast.Expr{elem.AST},
				},
				RequiresUnexported: ptrType.RequiresUnexported || elem.RequiresUnexported,
				OmittedUnexported:  elem.OmittedUnexported,
			}, nil
		}
		// Wrap custom type representations in generic pointer.
		if _, ok := customtype.Is(vv.Elem().Type()); ok {
			return Result{
				AST: pointifyASTExpr(elem.AST),
			}, nil
		}
		return Result{
			AST: &ast.UnaryExpr{
				Op: token.AND,
				X:  elem.AST,
			},
			RequiresUnexported: ptrType.RequiresUnexported || elem.RequiresUnexported,
			OmittedUnexported:  elem.OmittedUnexported,
		}, nil
	case reflect.Slice:
		var (
			elts               []ast.Expr
			requiresUnexported bool
		)
		for i := 0; i < vv.Len(); i++ {
			elem, err := computeASTProfiled(vv.Index(i), opt.withUnqualify(), cycleDetector, profiler, typeExprCache, packagesFound)
			if err != nil {
				return Result{}, err
			}
			if elem.RequiresUnexported {
				requiresUnexported = true
			}
			elts = append(elts, elem.AST)
		}
		sliceType, err := typeExpr(vv.Type(), opt, typeExprCache)
		if err != nil {
			return Result{}, err
		}
		return Result{
			AST: &ast.CompositeLit{
				Type: sliceType.AST,
				Elts: elts,
			},
			RequiresUnexported: requiresUnexported || sliceType.RequiresUnexported,
		}, nil
	case reflect.String:
		s := v.String()
		wantRawStringLiteral := len(s) > 40 && strings.Contains(s, "\n")
		wantRawStringLiteral = wantRawStringLiteral || strings.Contains(s, `"`)
		if wantRawStringLiteral && !strings.Contains(s, "`") {
			return basicLit(vv, token.STRING, "string", "`"+s+"`", opt.withUnqualify(), typeExprCache)
		}
		return basicLit(vv, token.STRING, "string", strconv.Quote(v.String()), opt.withUnqualify(), typeExprCache)
	case reflect.Struct:
		if render, ok := customtype.Is(v.Type()); ok {
			return Result{
				AST: render(v.Interface()),
			}, nil
		}

		var (
			structValue                           []ast.Expr
			requiresUnexported, omittedUnexported bool
		)
		for i := 0; i < v.NumField(); i++ {
			if unexported(v.Field(i)).IsZero() {
				continue
			}
			value, err := computeASTProfiled(unexported(v.Field(i)), opt.withUnqualify(), cycleDetector, profiler, typeExprCache, packagesFound)
			if err != nil {
				return Result{}, err
			}
			if value.RequiresUnexported {
				if opt.ExportedOnly {
					omittedUnexported = true
					continue
				}
				requiresUnexported = true
			}
			if value.OmittedUnexported {
				omittedUnexported = true
			}
			structValue = append(structValue, &ast.KeyValueExpr{
				Key:   ast.NewIdent(v.Type().Field(i).Name),
				Value: value.AST,
			})
		}
		structType, err := typeExpr(vv.Type(), opt, typeExprCache)
		if err != nil {
			return Result{}, err
		}
		if opt.ExportedOnly && structType.RequiresUnexported {
			return Result{RequiresUnexported: true}, nil
		}
		return Result{
			AST: &ast.CompositeLit{
				Type: structType.AST,
				Elts: structValue,
			},
			RequiresUnexported: structType.RequiresUnexported || requiresUnexported,
			OmittedUnexported:  omittedUnexported,
		}, nil
	case reflect.UnsafePointer:
		unsafePointerType, err := typeExpr(vv.Type(), opt, typeExprCache)
		if err != nil {
			return Result{}, err
		}
		return Result{
			AST: &ast.CallExpr{
				Fun: unsafePointerType.AST,
				Args: []ast.Expr{
					&ast.CallExpr{
						Fun:  ast.NewIdent("uintptr"),
						Args: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: fmt.Sprintf("0x%x", v.Pointer())}},
					},
				},
			},
			RequiresUnexported: unsafePointerType.RequiresUnexported,
			OmittedUnexported:  unsafePointerType.OmittedUnexported,
		}, nil
	default:
		return Result{AST: nil}, &ErrInvalidType{Value: v.Interface()}
	}
}

// literalNeedsQualification tells if a literal value needs qualification or not when initializing
// a value of type `interface{}`, e.g. being passed into the valast.Addr() helper function.
func literalNeedsQualification(v reflect.Value) bool {
	k := v.Kind()

	// Simple cases: Types whose literal values are always implicitly qualified
	if k == reflect.Bool ||
		k == reflect.String ||
		k == reflect.Int ||
		k == reflect.Array ||
		k == reflect.Chan ||
		k == reflect.Func ||
		k == reflect.Interface ||
		k == reflect.Map ||
		k == reflect.Ptr ||
		k == reflect.Slice ||
		k == reflect.Struct ||
		k == reflect.UnsafePointer {
		return false
	}

	// Floats. If passed to a function accepting an `interface{}` value:
	//
	// * A whole number `1234` would be considered an integer.
	// * A non-whole number `3.14` would be considered `float64`
	//
	if k == reflect.Float64 && v.Float() != math.Trunc(v.Float()) {
		return false // A float64 and not a whole number, so no qualification needed.
	}
	return true // needs qualification
}

func unexported(v reflect.Value) reflect.Value {
	if v == (reflect.Value{}) {
		return v
	}
	return bypass.UnsafeReflectValue(v)
}

// pointifyASTExpr wraps an expression in a call to the `Ptr` helper function.
//
//	valast.Ptr(//...)
func pointifyASTExpr(e ast.Expr) ast.Expr {
	return &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   ast.NewIdent("valast"),
			Sel: ast.NewIdent("Ptr"),
		},
		Args: []ast.Expr{e},
	}
}
