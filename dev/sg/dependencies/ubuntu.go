package dependencies

import (
	"context"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Ubuntu declares Ubuntu dependencies.
var Ubuntu = []category{
	{
		Name: "TODO",
		Enabled: func(ctx context.Context, args CheckArgs) error {
			return errors.New("not implemented")
		},
	},
}
