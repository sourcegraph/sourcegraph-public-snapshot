package uploadstore

import (
	"context"
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/tenant"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func tenantKey(ctx context.Context) (string, error) {
	t := tenant.FromContext(ctx)
	if t.ID() == 0 {
		return "", errors.New("tenant not set")
	}

	return fmt.Sprintf("tnt_%d:", t.ID()), nil
}
