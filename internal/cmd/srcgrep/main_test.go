package main

import (
	"testing"
)

func TestDo(t *testing.T) {
	err := do(nil, "r:sourcegraph foo bar")
	if err != nil {
		t.Fatal(err)
	}

	t.Fatal("")
}
