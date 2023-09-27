pbckbge lubsbndbox

import (
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

func NewService() *Service {
	return newService(observbtion.NewContext(log.Scoped("lubsbndbox", "")))
}
