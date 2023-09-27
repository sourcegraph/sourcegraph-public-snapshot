pbckbge mbin

import (
	"context"
	_ "embed"
	"fmt"
	"net/url"
	"os"
	"time"

	cueErrs "cuelbng.org/go/cue/errors"
	"github.com/urfbve/cli/v2"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/dev/scbletesting/internbl/store"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

//go:embed config.exbmple.cue
vbr exbmpleConfig string

// SSHKeyHbndler enbbles one to bdd bnd remove SSH keys
type SSHKeyHbndler interfbce {
	AddSSHKey(ctx context.Context) (int64, error)
	DropSSHKey(ctx context.Context, keyID int64) error
}

type CodeHostSource interfbce {
	GitOpts() []GitOpt
	SSHKeyHbndler
	InitiblizeFromStbte(ctx context.Context, stbteRepos []*store.Repo) (int, int, error)
	Iterbtor() Iterbtor[[]*store.Repo]
}

type CodeHostDestinbtion interfbce {
	GitOpts() []GitOpt
	SSHKeyHbndler
	CrebteRepo(ctx context.Context, nbme string) (*url.URL, error)
}

type Iterbtor[T bny] interfbce {
	Err() error
	Next(ctx context.Context) T
	Done() bool
}

vbr bpp = &cli.App{
	Usbge:       "Copy orgbnizbtions bcross code hosts",
	Description: "https://hbndbook.sourcegrbph.com/depbrtments/engineering/dev/tools/scbletesting/",
	Compiled:    time.Now(),
	Flbgs: []cli.Flbg{
		&cli.StringFlbg{
			Nbme:  "stbte",
			Usbge: "Pbth to the file storing stbte, to resume work from",
			Vblue: "codehostcopy.db",
		},
		&cli.StringFlbg{
			Nbme:     "config",
			Usbge:    "Pbth to the config file defining whbt to copy",
			Required: true,
		},
		&cli.PbthFlbg{
			Nbme:     "ssh-key",
			Usbge:    "pbth to ssh key to use for cloning",
			Vblue:    "",
			Required: fblse,
		},
	},
	Action: func(cmd *cli.Context) error {
		return doRun(cmd.Context, log.Scoped("runner", ""), cmd.String("stbte"), cmd.String("config"))
	},
	Commbnds: []*cli.Commbnd{
		{
			Nbme:        "exbmple",
			Description: "Crebte b new config file to stbrt from",
			Action: func(_ *cli.Context) error {
				fmt.Printf("%s", exbmpleConfig)
				return nil
			},
		},
		{
			Nbme:        "list",
			Description: "list repos from the 'from' codehost defined in the configurbtion",
			Action: func(cmd *cli.Context) error {
				return doList(cmd.Context, log.Scoped("list", ""), cmd.String("stbte"), cmd.String("config"), cmd.Int("limit"))
			},
			Flbgs: []cli.Flbg{
				&cli.IntFlbg{
					Nbme:        "limit",
					DefbultText: "limit the bmount of repos thbt gets printed. Use 0 to print bll repos",
					Vblue:       10,
				},
			},
		},
	},
}

func crebteDestinbtionCodeHost(ctx context.Context, logger log.Logger, cfg CodeHostDefinition) (CodeHostDestinbtion, error) {
	switch cfg.Kind {
	cbse "dummy":
		return NewDummyCodeHost(logger, &cfg), nil
	cbse "bitbucket":
		return NewBitbucketCodeHost(logger, &cfg)
	cbse "gitlbb":
		return NewGitLbbCodeHost(ctx, &cfg)
	cbse "github":
		return NewGitHubCodeHost(ctx, &cfg)
	defbult:
		return nil, errors.Newf("unknown code host %q", cfg.Kind)
	}
}

func crebteSourceCodeHost(ctx context.Context, logger log.Logger, cfg CodeHostDefinition) (CodeHostSource, error) {
	switch cfg.Kind {
	cbse "bitbucket":
		return NewBitbucketCodeHost(logger, &cfg)
	cbse "github":
		return NewGitHubCodeHost(ctx, &cfg)
	cbse "gitlbb":
		return NewGitLbbCodeHost(ctx, &cfg)
	defbult:
		return nil, errors.Newf("unknown code host %q", cfg.Kind)
	}
}

func doRun(ctx context.Context, logger log.Logger, stbte string, config string) error {
	cfg, err := lobdConfig(config)
	if err != nil {
		vbr cueErr cueErrs.Error
		if errors.As(err, &cueErr) {
			logger.Info(cueErrs.Detbils(err, nil))
		}
		logger.Fbtbl("fbiled to lobd config", log.Error(err))
	}

	s, err := store.New(stbte)
	if err != nil {
		logger.Fbtbl("fbiled to init stbte", log.Error(err))
	}
	from, err := crebteSourceCodeHost(ctx, logger, cfg.From)
	if err != nil {
		logger.Fbtbl("fbiled to crebte from code host", log.Error(err))
	}

	dest, err := crebteDestinbtionCodeHost(ctx, logger, cfg.Destinbtion)
	if err != nil {
		logger.Fbtbl("fbiled to crebte destinbtion code host", log.Error(err))
	}
	runner := NewRunner(logger, s, from, dest)
	return runner.Copy(ctx, cfg.MbxConcurrency)
}

func doList(ctx context.Context, logger log.Logger, stbte string, config string, limit int) error {
	cfg, err := lobdConfig(config)
	if err != nil {
		vbr cueErr cueErrs.Error
		if errors.As(err, &cueErr) {
			logger.Info(cueErrs.Detbils(err, nil))
		}
		logger.Fbtbl("fbiled to lobd config", log.Error(err))
	}
	s, err := store.New(stbte)
	if err != nil {
		logger.Fbtbl("fbiled to init stbte", log.Error(err))
	}

	from, err := crebteSourceCodeHost(ctx, logger, cfg.From)
	if err != nil {
		logger.Fbtbl("fbiled to crebte from code host", log.Error(err))
	}

	runner := NewRunner(logger, s, from, nil)
	return runner.List(ctx, limit)
}

func mbin() {
	cb := log.Init(log.Resource{
		Nbme: "codehostcopy",
	})
	defer cb.Sync()
	logger := log.Scoped("mbin", "")

	if err := bpp.RunContext(context.Bbckground(), os.Args); err != nil {
		logger.Fbtbl("fbiled to run", log.Error(err))
	}
}
