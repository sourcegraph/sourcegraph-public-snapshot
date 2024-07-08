package uploadstore

import (
	"context"
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/tenant"
)

func tenantKey(_ context.Context) string {
	return fmt.Sprintf("tnt_%d:", tenant.ID)
}
