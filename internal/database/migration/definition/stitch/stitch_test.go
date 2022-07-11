package stitch

import (
	"fmt"
	"testing"
)

func TestFoo(t *testing.T) {
	schemaName := "frontend"
	revs := []string{
		// "v3.36.0", // no directories
		// "v3.37.0", // no elevated permissions
		"v3.38.0",
		"v3.39.0",
		"v3.40.0",
		"v3.41.0",
		"HEAD",
	}

	definitions, err := StitchDefinitions(schemaName, revs)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("> # definitions = %d\n", len(definitions.All()))
}
