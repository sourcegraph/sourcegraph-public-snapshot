package ctags

import (
	"os/exec"
	"reflect"
	"testing"
)

func TestParser(t *testing.T) {
	// TODO(sqs): find a way to make it easy to run these tests in local dev (w/o needing to install universal-ctags) and CI
	const command = "universal-ctags"
	if _, err := exec.LookPath(command); err != nil {
		t.Skipf("command not in PATH: %s", command)
	}

	p, err := NewParser(command)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()

	java := `
package com.sourcegraph;
import a.b.c;
class A implements B extends C {
  public static int D = 1;
  public int E;
  public A() {
    E = 2;
  }
  public int F() {
    E++;
  }
}
`
	name := "com/sourcegraph/A.java"
	got, err := p.Parse(name, []byte(java))
	if err != nil {
		t.Error(err)
	}

	want := []Entry{
		{
			Kind:     "package",
			Language: "Java",
			Line:     2,
			Name:     "com.sourcegraph",
			Path:     "com/sourcegraph/A.java",
		},
		{
			Kind:     "class",
			Language: "Java",
			Line:     4,
			Name:     "A",
			Path:     "com/sourcegraph/A.java",
		},

		{
			Kind:       "field",
			Language:   "Java",
			Line:       5,
			Name:       "D",
			Parent:     "A",
			ParentKind: "class",
			Path:       "com/sourcegraph/A.java",
		},
		{
			Kind:       "field",
			Language:   "Java",
			Line:       6,
			Name:       "E",
			Parent:     "A",
			ParentKind: "class",
			Path:       "com/sourcegraph/A.java",
		},
		{
			Kind:       "method",
			Language:   "Java",
			Line:       7,
			Name:       "A",
			Parent:     "A",
			ParentKind: "class",
			Path:       "com/sourcegraph/A.java",
			Signature:  "()",
		},
		{
			Kind:       "method",
			Language:   "Java",
			Line:       10,
			Name:       "F",
			Parent:     "A",
			ParentKind: "class",
			Path:       "com/sourcegraph/A.java",
			Signature:  "()",
		},
	}

	for i := range want {
		got[i].Pattern = ""
		if !reflect.DeepEqual(got[i], want[i]) {
			t.Fatalf("got %#v, want %#v", got[i], want[i])
		}
	}
}
