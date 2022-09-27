package main

import (
	"os"
	"strconv"

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

func envVar[T any](name string, target *T) error {
	value, exists := os.LookupEnv(name)
	if !exists {
		return errors.Newf("%s not found in environment", name)
	}

	switch p := any(target).(type) {
	case *bool:
		{
			v, err := strconv.ParseBool(value)
			if err != nil {
				return err
			}

			*p = v
		}
	case *string:
		{
			*p = value
		}
	default:
		panic(errors.Newf("unsuporrted target type %T", target))
	}

	return nil
}
