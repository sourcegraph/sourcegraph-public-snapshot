package luasandbox

import (
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func NewService() *Service {
	return newService(observation.NewContext(log.Scoped("luasandbox")))
}
