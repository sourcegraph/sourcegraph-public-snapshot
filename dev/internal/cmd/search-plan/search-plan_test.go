package main

import (
	"os"
	"testing"
)

func TestRun(t *testing.T) {
	err := run(os.Stdout, []string{"search-plan", "-dotcom", "-pattern_type=literal", `content:"hello\nworld"`})
	if err != nil {
		t.Error(err)
	}
}
