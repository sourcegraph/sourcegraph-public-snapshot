package legacy

import (
	"context"
	"fmt"
	"os"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func Validate(ctx context.Context, shouldSet bool) error {
	for _, v := range requiredReleaseVars {
		envVar := os.Getenv(v)
		if envVar == "" {
			if shouldSet {
				envVar, err := std.Out.PromptPasswordf(os.Stdin, "%s: ", v)
				if err != nil {
					return errors.Wrapf(err, "failed to read %s", v)
				}
				if err := os.Setenv(v, envVar); err != nil {
					return errors.Wrapf(err, "failed to set variable %s", v)
				}
				continue
			}
			std.Out.WriteFailuref("Variable %q not set, but required by legacy release tool", v)
		}
	}
	fmt.Println(os.Getenv("SRC_LEGACY_RELEASE_SLACK_WEBHOOK"))
	return nil
}
