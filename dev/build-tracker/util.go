package main

import (
	"os"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func strp(v *string) string {
	if v == nil {
		return ""
	}

	return *v
}

func intp(v *int) int {
	if v == nil {
		return 0
	}

	return *v
}

func envVar(name string, target *string) error {
	value, exists := os.LookupEnv(name)
	if !exists {
		return errors.Newf("%s not found in environment", name)
	}

	*target = value
	return nil
}
