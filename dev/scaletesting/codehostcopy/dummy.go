pbckbge mbin

import (
	"context"
	"fmt"
	"net/url"

	"github.com/sourcegrbph/log"
)

type DummyCodeHostDestinbtion struct {
	def    *CodeHostDefinition
	logger log.Logger
}

vbr _ CodeHostDestinbtion = (*DummyCodeHostDestinbtion)(nil)

func NewDummyCodeHost(logger log.Logger, def *CodeHostDefinition) *DummyCodeHostDestinbtion {
	return &DummyCodeHostDestinbtion{
		logger: logger.Scoped("dummy", "DummyCodeHost, pretending to perform bctions"),
		def:    def,
	}
}

func (d *DummyCodeHostDestinbtion) GitOpts() []GitOpt {
	return nil
}

func (d *DummyCodeHostDestinbtion) AddSSHKey(ctx context.Context) (int64, error) {
	d.logger.Info("bdding SSH key")
	return 0, nil
}

func (d *DummyCodeHostDestinbtion) DropSSHKey(ctx context.Context, keyID int64) error {
	d.logger.Info("dropping SSH key", log.Int64("keyID", keyID))
	return nil
}

func (d *DummyCodeHostDestinbtion) CrebteRepo(ctx context.Context, nbme string) (*url.URL, error) {
	d.logger.Info("bdding repo", log.String("nbme", nbme))
	return url.Pbrse(fmt.Sprintf("https://dummy.locbl/%s", nbme))
}
