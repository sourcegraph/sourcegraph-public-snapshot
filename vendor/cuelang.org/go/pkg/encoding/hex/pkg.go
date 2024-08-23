// Code generated by go generate. DO NOT EDIT.

//go:generate rm pkg.go
//go:generate go run ../../gen/gen.go

package hex

import (
	"cuelang.org/go/internal/core/adt"
	"cuelang.org/go/pkg/internal"
)

func init() {
	internal.Register("encoding/hex", pkg)
}

var _ = adt.TopKind // in case the adt package isn't used

var pkg = &internal.Package{
	Native: []*internal.Builtin{{
		Name: "EncodedLen",
		Params: []internal.Param{
			{Kind: adt.IntKind},
		},
		Result: adt.IntKind,
		Func: func(c *internal.CallCtxt) {
			n := c.Int(0)
			if c.Do() {
				c.Ret = EncodedLen(n)
			}
		},
	}, {
		Name: "DecodedLen",
		Params: []internal.Param{
			{Kind: adt.IntKind},
		},
		Result: adt.IntKind,
		Func: func(c *internal.CallCtxt) {
			x := c.Int(0)
			if c.Do() {
				c.Ret = DecodedLen(x)
			}
		},
	}, {
		Name: "Decode",
		Params: []internal.Param{
			{Kind: adt.StringKind},
		},
		Result: adt.BytesKind | adt.StringKind,
		Func: func(c *internal.CallCtxt) {
			s := c.String(0)
			if c.Do() {
				c.Ret, c.Err = Decode(s)
			}
		},
	}, {
		Name: "Dump",
		Params: []internal.Param{
			{Kind: adt.BytesKind | adt.StringKind},
		},
		Result: adt.StringKind,
		Func: func(c *internal.CallCtxt) {
			data := c.Bytes(0)
			if c.Do() {
				c.Ret = Dump(data)
			}
		},
	}, {
		Name: "Encode",
		Params: []internal.Param{
			{Kind: adt.BytesKind | adt.StringKind},
		},
		Result: adt.StringKind,
		Func: func(c *internal.CallCtxt) {
			src := c.Bytes(0)
			if c.Do() {
				c.Ret = Encode(src)
			}
		},
	}},
}
