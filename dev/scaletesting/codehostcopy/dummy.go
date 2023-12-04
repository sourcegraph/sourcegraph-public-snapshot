package main

import (
	"context"
	"fmt"
	"net/url"

	"github.com/sourcegraph/log"
)

type DummyCodeHostDestination struct {
	def    *CodeHostDefinition
	logger log.Logger
}

var _ CodeHostDestination = (*DummyCodeHostDestination)(nil)

func NewDummyCodeHost(logger log.Logger, def *CodeHostDefinition) *DummyCodeHostDestination {
	return &DummyCodeHostDestination{
		logger: logger.Scoped("dummy"),
		def:    def,
	}
}

func (d *DummyCodeHostDestination) GitOpts() []GitOpt {
	return nil
}

func (d *DummyCodeHostDestination) AddSSHKey(ctx context.Context) (int64, error) {
	d.logger.Info("adding SSH key")
	return 0, nil
}

func (d *DummyCodeHostDestination) DropSSHKey(ctx context.Context, keyID int64) error {
	d.logger.Info("dropping SSH key", log.Int64("keyID", keyID))
	return nil
}

func (d *DummyCodeHostDestination) CreateRepo(ctx context.Context, name string) (*url.URL, error) {
	d.logger.Info("adding repo", log.String("name", name))
	return url.Parse(fmt.Sprintf("https://dummy.local/%s", name))
}
