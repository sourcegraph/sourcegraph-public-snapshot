package executor

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel"
)

func Init(ctx context.Context, enterpriseServices *enterprise.Services) error {
	handler, err := codeintel.NewCodeIntelUploadHandler(ctx, true)
	if err != nil {
		return err
	}

	proxyHandler, err := newInternalProxyHandler(handler)
	if err != nil {
		return err
	}

	enterpriseServices.NewExecutorProxyHandler = proxyHandler
	return nil
}
