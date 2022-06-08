package dependencies

import (
	"context"
	"os"
	"path/filepath"

	"github.com/sourcegraph/run"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/check"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func categoryCloneRepositories() category {
	return category{
		Name:      "Clone repositories",
		DependsOn: []string{depsBaseUtilities},
		Checks: []*dependency{
			{
				Name: "SSH authentication with GitHub.com",
				Description: `Make sure that you can clone git repositories from GitHub via SSH.
See here on how to set that up:

https://docs.github.com/en/authentication/connecting-to-github-with-ssh`,
				Check: func(ctx context.Context, cio check.IO, args CheckArgs) error {
					if args.Teammate {
						return check.CommandOutputContains(
							"ssh -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no -T git@github.com",
							"successfully authenticated")(ctx)
					}
					// otherwise, we don't need auth set up at all, since everything is OSS
					return nil
				},
				// TODO we might be able to automate this fix
			},
			{
				Name:        "github.com/sourcegraph/sourcegraph",
				Description: `The 'sourcegraph' repository contains the Sourcegraph codebase and everything to run Sourcegraph locally.`,
				Check: func(ctx context.Context, cio check.IO, args CheckArgs) error {
					if _, err := root.RepositoryRoot(); err == nil {
						return nil
					}

					ok, err := pathExists("sourcegraph")
					if !ok || err != nil {
						return errors.New("'sg setup' is not run in sourcegraph and repository is also not found in current directory")
					}
					return nil
				},
				Fix: func(ctx context.Context, cio check.IO, args CheckArgs) error {
					var cmd *run.Command
					if args.Teammate {
						cmd = run.Cmd(ctx, `git clone git@github.com:sourcegraph/sourcegraph.git`)
					} else {
						cmd = run.Cmd(ctx, `git clone https://github.com/sourcegraph/sourcegraph.git`)
					}
					return cmd.Run().StreamLines(cio.Write)
				},
			},
			{
				Name: "github.com/sourcegraph/dev-private",
				Description: `In order to run the local development environment as a Sourcegraph teammate,
	you'll need to clone another repository: github.com/sourcegraph/dev-private.

	It contains convenient preconfigured settings and code host connections.

	It needs to be cloned into the same folder as sourcegraph/sourcegraph,
	so they sit alongside each other, like this:

	/dir
	|-- dev-private
	+-- sourcegraph

	NOTE: You can ignore this if you're not a Sourcegraph teammate.`,
				Enabled: teammatesOnly(),
				Check: func(ctx context.Context, cio check.IO, args CheckArgs) error {
					ok, err := pathExists("dev-private")
					if ok && err == nil {
						return nil
					}
					wd, err := os.Getwd()
					if err != nil {
						return errors.Wrap(err, "failed to check for dev-private repository")
					}

					p := filepath.Join(wd, "..", "dev-private")
					ok, err = pathExists(p)
					if ok && err == nil {
						return nil
					}
					return errors.New("could not find dev-private repository either in current directory or one above")
				},
				Fix: cmdAction(`git clone git@github.com:sourcegraph/dev-private.git`),
			},
		},
	}
}
