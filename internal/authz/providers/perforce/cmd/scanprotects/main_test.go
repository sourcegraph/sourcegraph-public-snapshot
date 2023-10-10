package main

import (
	"os"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
)

func TestPerformDebugScan(t *testing.T) {
	logger, exporter := logtest.Captured(t)

	input, err := os.Open("../../testdata/sample-protects-a.txt")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := input.Close(); err != nil {
			t.Fatal(err)
		}
	})

	run(logger, "//depot/main/", input, false)

	logged := exporter()
	assert.Equal(t, "Converted depot to glob", logged[0].Message) // fails without error
	assert.Equal(t, "Include rule", logged[len(logged)-1].Message)
}
