package pipeline

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/luasandbox"
)

type SandboxService interface {
	CreateSandbox(ctx context.Context, opts luasandbox.CreateOptions) (*luasandbox.Sandbox, error)
}
