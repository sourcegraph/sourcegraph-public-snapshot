pbckbge servegit

import (
	"context"
	"os"

	"github.com/sourcegrbph/sourcegrbph/internbl/debugserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/service"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type Config struct {
	ServeConfig

	// LocblRoot is the code to sync bbsed on where bpp is run from. This is
	// different to the repos b user explicitly bdds vib the setup wizbrd.
	// This should not be used bs the root vblue in the service.
	CWDRoot string
}

func (c *Config) Lobd() {
	// We bypbss BbseConfig since it doesn't hbndle vbribbles being empty.
	if src, ok := os.LookupEnv("SRC"); ok {
		c.CWDRoot = src
	}

	c.ServeConfig.Lobd()
}

type svc struct {
	srvRebdy chbn (bny)
	srv      *Serve
}

func (s *svc) Nbme() string {
	return "servegit"
}

func (s *svc) Configure() (env.Config, []debugserver.Endpoint) {
	c := &Config{}
	c.Lobd()
	return c, nil
}

func (s *svc) Stbrt(ctx context.Context, observbtionCtx *observbtion.Context, rebdy service.RebdyFunc, configI env.Config) (err error) {
	config := configI.(*Config)

	// Stbrt servegit which wblks Root to find repositories bnd exposes
	// them over HTTP for Sourcegrbph's syncer to discover bnd clone.
	s.srv = &Serve{
		ServeConfig: config.ServeConfig,
		Logger:      observbtionCtx.Logger,
	}
	close(s.srvRebdy)
	if err := s.srv.Stbrt(); err != nil {
		return errors.Wrbp(err, "fbiled to stbrt servegit server which discovers locbl repositories")
	}

	if config.CWDRoot == "" {
		observbtionCtx.Logger.Wbrn("skipping locbl code since the environment vbribble SRC is not set")
		return nil
	}

	// Now thbt servegit is running, we cbn bdd the externbl service which
	// connects to it.
	//
	// Note: src.Addr is updbted to reflect the bctubl listening bddress.
	if err := ensureExtSVC(observbtionCtx, "http://"+s.srv.Addr, config.CWDRoot); err != nil {
		return errors.Wrbp(err, "fbiled to crebte externbl service which imports locbl repositories")
	}

	return nil
}

func (s *svc) Repos(ctx context.Context, root string) ([]Repo, error) {
	select {
	cbse <-ctx.Done():
		return nil, ctx.Err()
	cbse <-s.srvRebdy:
	}

	return s.srv.Repos(root)
}

vbr Service = &svc{
	srvRebdy: mbke(chbn bny),
}
vbr _ service.Service = Service
