// Code generated by go generate. DO NOT EDIT.

//go:generate rm pkg.go
//go:generate go run ../../gen/gen.go

package json

import (
	"cuelang.org/go/internal/core/adt"
	"cuelang.org/go/pkg/internal"
)

func init() {
	internal.Register("encoding/json", pkg)
}

var _ = adt.TopKind // in case the adt package isn't used

var pkg = &internal.Package{
	Native: []*internal.Builtin{{
		Name: "Valid",
		Params: []internal.Param{
			{Kind: adt.BytesKind | adt.StringKind},
		},
		Result: adt.BoolKind,
		Func: func(c *internal.CallCtxt) {
			data := c.Bytes(0)
			if c.Do() {
				c.Ret = Valid(data)
			}
		},
	}, {
		Name: "Compact",
		Params: []internal.Param{
			{Kind: adt.BytesKind | adt.StringKind},
		},
		Result: adt.StringKind,
		Func: func(c *internal.CallCtxt) {
			src := c.Bytes(0)
			if c.Do() {
				c.Ret, c.Err = Compact(src)
			}
		},
	}, {
		Name: "Indent",
		Params: []internal.Param{
			{Kind: adt.BytesKind | adt.StringKind},
			{Kind: adt.StringKind},
			{Kind: adt.StringKind},
		},
		Result: adt.StringKind,
		Func: func(c *internal.CallCtxt) {
			src, prefix, indent := c.Bytes(0), c.String(1), c.String(2)
			if c.Do() {
				c.Ret, c.Err = Indent(src, prefix, indent)
			}
		},
	}, {
		Name: "HTMLEscape",
		Params: []internal.Param{
			{Kind: adt.BytesKind | adt.StringKind},
		},
		Result: adt.StringKind,
		Func: func(c *internal.CallCtxt) {
			src := c.Bytes(0)
			if c.Do() {
				c.Ret = HTMLEscape(src)
			}
		},
	}, {
		Name: "Marshal",
		Params: []internal.Param{
			{Kind: adt.TopKind},
		},
		Result: adt.StringKind,
		Func: func(c *internal.CallCtxt) {
			v := c.Value(0)
			if c.Do() {
				c.Ret, c.Err = Marshal(v)
			}
		},
	}, {
		Name: "MarshalStream",
		Params: []internal.Param{
			{Kind: adt.TopKind},
		},
		Result: adt.StringKind,
		Func: func(c *internal.CallCtxt) {
			v := c.Value(0)
			if c.Do() {
				c.Ret, c.Err = MarshalStream(v)
			}
		},
	}, {
		Name: "UnmarshalStream",
		Params: []internal.Param{
			{Kind: adt.BytesKind | adt.StringKind},
		},
		Result: adt.TopKind,
		Func: func(c *internal.CallCtxt) {
			data := c.Bytes(0)
			if c.Do() {
				c.Ret, c.Err = UnmarshalStream(data)
			}
		},
	}, {
		Name: "Unmarshal",
		Params: []internal.Param{
			{Kind: adt.BytesKind | adt.StringKind},
		},
		Result: adt.TopKind,
		Func: func(c *internal.CallCtxt) {
			b := c.Bytes(0)
			if c.Do() {
				c.Ret, c.Err = Unmarshal(b)
			}
		},
	}, {
		Name: "Validate",
		Params: []internal.Param{
			{Kind: adt.BytesKind | adt.StringKind},
			{Kind: adt.TopKind},
		},
		Result: adt.BoolKind,
		Func: func(c *internal.CallCtxt) {
			b, v := c.Bytes(0), c.Value(1)
			if c.Do() {
				c.Ret, c.Err = Validate(b, v)
			}
		},
	}},
}
