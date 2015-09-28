package util

import (
	"testing"

	"github.com/golang/protobuf/proto"
	descriptor "github.com/golang/protobuf/protoc-gen-go/descriptor"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
)

func TestParseParams(t *testing.T) {
	params := ParseParams(&plugin.CodeGeneratorRequest{
		Parameter: proto.String("key =value,abc = d ef , z = g "),
	})
	if len(params) != 3 {
		t.Fatal("expected 3 arguments got", len(params))
	}
	if params["key"] != "value" {
		t.Fatal(`"key" != "value"`)
	}
	if params["abc"] != "d ef" {
		t.Fatal(`"abc" != "d ef"`)
	}
	if params["z"] != "g" {
		t.Fatal(`"z" != "g"`)
	}
}

func TestIsFullyQualified(t *testing.T) {
	tests := map[string]bool{
		".google.protobuf.UninterpretedOption": true,
		".google.protobuf.FieldOptions.CType":  true,
		"protobuf.FieldOptions.CType":          false,
		"UninterpretedOption":                  false,
	}
	for symbolPath, want := range tests {
		got := IsFullyQualified(symbolPath)
		if got != want {
			t.Fatalf("got %v want %v", got, want)
		}
	}
}

func TestTrimElem(t *testing.T) {
	tests := []struct {
		symbolPath, want string
		n                int
	}{
		// Standard cases:
		{"a.b.c", "b.c", 1},
		{".a.b.c", "b.c", 1},
		{".a.b.c", ".a.b", -1},

		// Extreme cases:
		{"a.b.c", "", 1000},
		{".a.b.c", "", 1000},
		{"a.b.c", "", -1000},
		{".a.b.c", "", -1000},

		{".a.b.c.d", "b.c.d", 1},
		{"a.b.c.d", "b.c.d", 1},
		{".a.b.c.d", "c.d", 2},
		{"a.b.c.d.e", "", 1000},
	}
	for _, tst := range tests {
		got := TrimElem(tst.symbolPath, tst.n)
		if got != tst.want {
			t.Logf("symbolPath=%q\n", tst.symbolPath)
			t.Logf("n=%v\n", tst.n)
			t.Fatalf("got %q want %q\n", got, tst.want)
		}
	}
}

func TestCountElem(t *testing.T) {
	tests := []struct {
		symbolPath string
		want       int
	}{
		{"a.b.c", 3},
		{".a.b.c", 3},
		{"a.b.c.d", 4},
		{"a", 1},
		{".", 0},
		{"", 0},
	}
	for _, tst := range tests {
		got := CountElem(tst.symbolPath)
		if got != tst.want {
			t.Logf("symbolPath=%q\n", tst.symbolPath)
			t.Fatalf("got %v want %v\n", got, tst.want)
		}
	}
}

func TestPackageName(t *testing.T) {
	got := PackageName(&descriptor.FileDescriptorProto{
		Package: proto.String("foo"),
	})
	if got != "foo" {
		t.Fatalf("expected explicit package name \"foo\", got %q\n", got)
	}

	got = PackageName(&descriptor.FileDescriptorProto{
		Name: proto.String("some/arbitrary/file.proto"),
	})
	if got != "file" {
		t.Fatalf("expected derived package name \"file\", got %q\n", got)
	}
}
