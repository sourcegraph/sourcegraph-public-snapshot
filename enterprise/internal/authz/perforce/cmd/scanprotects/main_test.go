package main

import (
	"os"
	"testing"

	"github.com/sourcegraph/log/logtest"
)

func TestPerformDebugScan(t *testing.T) {
	logger := logtest.Scoped(t)

	input, err := os.Open("../../testdata/sample-protects-a.txt")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := input.Close(); err != nil {
			t.Fatal(err)
		}
	})

	run(logger, "//depot/main/", input)
}
