package refs_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/lang/golang/internal/refs"
)

func TestParseFile(t *testing.T) {
	pos := func(s string) token.Position {
		f := strings.Fields(s)
		fp := strings.Split(f[0], ":")
		line, err := strconv.Atoi(fp[1])
		if err != nil {
			panic(err)
		}
		column, err := strconv.Atoi(fp[2])
		if err != nil {
			panic(err)
		}
		offs, err := strconv.Atoi(strings.TrimSuffix(f[2], ")"))
		if err != nil {
			panic(err)
		}
		return token.Position{
			Filename: fp[0],
			Line:     line,
			Column:   column,
			Offset:   offs,
		}
	}
	cases := []struct {
		Filename string
		Want     []*refs.Ref
	}{
		{
			Filename: "testdata/empty.go",
			Want:     nil,
		},
		{
			Filename: "testdata/imports.go",
			Want: []*refs.Ref{
				&refs.Ref{Def: refs.Def{ImportPath: "net/http", Path: ""}, Position: pos("testdata/imports.go:3:8 (offset 21)")},
			},
		},
		{
			Filename: "testdata/http-request-headers.go",
			Want: []*refs.Ref{
				&refs.Ref{Def: refs.Def{ImportPath: "net/http", Path: ""}, Position: pos("testdata/http-request-headers.go:4:2 (offset 24)")},
				&refs.Ref{Def: refs.Def{ImportPath: "net/http", Path: "Request"}, Position: pos("testdata/http-request-headers.go:8:13 (offset 78)")},
				&refs.Ref{Def: refs.Def{ImportPath: "net/http", Path: "Request Header"}, Position: pos("testdata/http-request-headers.go:9:4 (offset 113)")},
				&refs.Ref{Def: refs.Def{ImportPath: "net/http", Path: "Request Header"}, Position: pos("testdata/http-request-headers.go:11:4 (offset 172)")},
			},
		},
		{
			Filename: "testdata/convoluted.go",
			Want: []*refs.Ref{
				&refs.Ref{Def: refs.Def{ImportPath: "net/http", Path: ""}, Position: pos("testdata/convoluted.go:3:8 (offset 21)")},
				&refs.Ref{Def: refs.Def{ImportPath: "net/http", Path: "Client"}, Position: pos("testdata/convoluted.go:6:8 (offset 75)")},
				&refs.Ref{Def: refs.Def{ImportPath: "net/http", Path: "RoundTripper RoundTrip"}, Position: pos("testdata/convoluted.go:15:14 (offset 179)")},
				&refs.Ref{Def: refs.Def{ImportPath: "net/http", Path: "Client Transport"}, Position: pos("testdata/convoluted.go:15:4 (offset 169)")},
				&refs.Ref{Def: refs.Def{ImportPath: "net/http", Path: "RoundTripper RoundTrip"}, Position: pos("testdata/convoluted.go:19:25 (offset 301)")},
				&refs.Ref{Def: refs.Def{ImportPath: "net/http", Path: "Client Transport"}, Position: pos("testdata/convoluted.go:19:15 (offset 291)")},
			},
		},
		{
			Filename: "testdata/defs.go",
			Want: []*refs.Ref{
				&refs.Ref{Def: refs.Def{ImportPath: "net/http", Path: ""}, Position: pos("testdata/defs.go:3:8 (offset 21)")},
				&refs.Ref{Def: refs.Def{ImportPath: "net/http", Path: "Client"}, Position: pos("testdata/defs.go:6:10 (offset 78)")},
				&refs.Ref{Def: refs.Def{ImportPath: "net/http", Path: "RoundTripper"}, Position: pos("testdata/defs.go:7:9 (offset 119)")},
				&refs.Ref{Def: refs.Def{ImportPath: "net/http", Path: "Client"}, Position: pos("testdata/defs.go:10:21 (offset 182)")},
				&refs.Ref{Def: refs.Def{ImportPath: "net/http", Path: "Client"}, Position: pos("testdata/defs.go:12:12 (offset 246)")},
				&refs.Ref{Def: refs.Def{ImportPath: "net/http", Path: "RoundTripper"}, Position: pos("testdata/defs.go:13:11 (offset 295)")},
				&refs.Ref{Def: refs.Def{ImportPath: "net/http", Path: "Client"}, Position: pos("testdata/defs.go:20:13 (offset 392)")},
			},
		},
		{
			Filename: "testdata/vars.go",
			Want: []*refs.Ref{
				&refs.Ref{Def: refs.Def{ImportPath: "net/http", Path: ""}, Position: pos("testdata/vars.go:3:8 (offset 21)")},
				&refs.Ref{Def: refs.Def{ImportPath: "net/http", Path: "Client"}, Position: pos("testdata/vars.go:6:14 (offset 74)")},
				&refs.Ref{Def: refs.Def{ImportPath: "net/http", Path: "RoundTripper"}, Position: pos("testdata/vars.go:8:12 (offset 124)")},
				&refs.Ref{Def: refs.Def{ImportPath: "net/http", Path: "Client Transport"}, Position: pos("testdata/vars.go:12:3 (offset 225)")},
				&refs.Ref{Def: refs.Def{ImportPath: "net/http", Path: "Client"}, Position: pos("testdata/vars.go:11:12 (offset 194)")},
			},
		},
	}
	for _, c := range cases {
		t.Run(c.Filename, func(t *testing.T) {
			cont, err := ioutil.ReadFile(c.Filename)
			if err != nil {
				t.Fatal(err)
			}
			cfg := refs.Default()
			astFile, err := parser.ParseFile(cfg.FileSet, c.Filename, cont, 0)
			if err != nil {
				t.Fatal(err)
			}
			var allRefs []*refs.Ref
			err = cfg.Refs("refstest", []*ast.File{astFile}, func(r *refs.Ref) {
				allRefs = append(allRefs, r)
			})
			if err != nil {
				t.Fatal(err)
			}
			if len(allRefs) != len(c.Want) {
				t.Log("got", len(allRefs), "refs:")
				for i, r := range allRefs {
					t.Logf("    %d. %+v\n", i, r)
				}
				t.Log("want", len(c.Want), "refs:")
				for i, r := range c.Want {
					t.Logf("    %d. %+v\n", i, r)
				}
				t.FailNow()
			}
			for i, ref := range allRefs {
				if !reflect.DeepEqual(ref, c.Want[i]) {
					t.Log("got", len(allRefs), "refs:")
					for i, r := range allRefs {
						t.Logf("    %d. %+v\n", i, r)
					}
					t.Log("want", len(c.Want), "refs:")
					for i, r := range c.Want {
						t.Logf("    %d. %+v\n", i, r)
					}
					t.FailNow()
				}
			}
		})
	}
}
