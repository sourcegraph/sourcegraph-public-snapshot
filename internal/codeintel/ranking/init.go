package ranking

import (
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads"
	"github.com/sourcegraph/sourcegraph/internal/memo"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// GetService creates or returns an already-initialized ranking service.
// If the service is not yet initialized, it will use the provided dependencies.
func GetService(
	uploadSvc *uploads.Service,
) *Service {
	svc, _ := initServiceMemo.Init(serviceDependencies{
		uploadSvc,
	})

	return svc
}

type serviceDependencies struct {
	uploadsService *uploads.Service
}

var initServiceMemo = memo.NewMemoizedConstructorWithArg(func(deps serviceDependencies) (*Service, error) {
	return newService(
		deps.uploadsService,
		scopedContext("service"),
	), nil
})

func scopedContext(component string) *observation.Context {
	return observation.ScopedContext("codeintel", "ranking", component)
}
