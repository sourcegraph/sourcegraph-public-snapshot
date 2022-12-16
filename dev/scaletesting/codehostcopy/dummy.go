package main

import (
	"context"
	"fmt"
	"net/url"

	"github.com/sourcegraph/log"
)

type DummyCodeHost struct {
	def    *CodeHostDefinition
	logger log.Logger
}

var _ CodeHostDestination = (*DummyCodeHost)(nil)

func NewDummyCodeHost(logger log.Logger, def *CodeHostDefinition) *DummyCodeHost {
	return &DummyCodeHost{
		logger: logger.Scoped("dummy", "DummyCodeHost, pretending to perform actions"),
		def:    def,
	}
}

func (d *DummyCodeHost) GitOpts() []GitOpt {
	return nil
}

func (d *DummyCodeHost) AddSSHKey(ctx context.Context) (int64, error) {
	d.logger.Info("adding SSH key")
	return 0, nil
}

func (d *DummyCodeHost) DropSSHKey(ctx context.Context, keyID int64) error {
	d.logger.Info("dropping SSH key", log.Int64("keyID", keyID))
	return nil
}

func (d *DummyCodeHost) CreateRepo(ctx context.Context, name string) (*url.URL, error) {
	d.logger.Info("adding repo", log.String("name", name))
	return url.Parse(fmt.Sprintf("https://dummmy.local/%s", name))
}
