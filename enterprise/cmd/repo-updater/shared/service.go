package shared

import (
	shared "github.com/sourcegraph/sourcegraph/cmd/repo-updater/shared"
	"github.com/sourcegraph/sourcegraph/internal/service"
)

var Service service.Service = shared.NewServiceWithEnterpriseInit(EnterpriseInit)
