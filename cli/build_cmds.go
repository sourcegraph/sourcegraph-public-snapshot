package cli

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"sourcegraph.com/sourcegraph/sourcegraph/cli/cli"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/statsutil"
)

func init() {
	buildsGroup, err := cli.CLI.AddCommand("build",
		"manage builds",
		"The build subcommands manage builds.",
		&buildsCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}
	buildsGroup.Aliases = []string{"builds", "b"}

	_, err = buildsGroup.AddCommand("get",
		"get build",
		"The get subcommand gets a specific build.",
		&buildsGetCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}

	_, err = buildsGroup.AddCommand("list",
		"list builds",
		"The list subcommand lists builds by the specified criteria.",
		&buildsListCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}

	_, err = buildsGroup.AddCommand("stats",
		"get builds statistics",
		"The stats subcommand displays statistics about previous and current builds.",
		&buildsStatsCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}
}

type buildsCmd struct{}

func (c *buildsCmd) Execute(args []string) error { return nil }

type buildsGetCmd struct {
	Args struct {
		Repo string `name:"repo" description:"repo URI"`
		ID   uint64 `name:"id" description:"build ID"`
	} `positional-args:"yes" required:"true"`
}

func (c *buildsGetCmd) Execute(args []string) error {
	cl := cliClient
	opt := &sourcegraph.BuildSpec{
		Repo: c.Args.Repo,
		ID:   c.Args.ID,
	}
	build, err := cl.Builds.Get(cliContext, opt)
	if err != nil {
		return err
	}
	b, err := json.MarshalIndent(build, "", "  ")
	if err != nil {
		return err
	}
	os.Stdout.Write(b)
	fmt.Println()
	return nil
}

type buildsListCmd struct {
	Repo      string `long:"repo" description:"repo URI"`
	CommitID  string `long:"commit" description:"filter builds by commit ID"`
	Active    bool   `long:"active"`
	Queued    bool   `long:"queued"`
	Succeeded bool   `long:"succeeded"`
	Ended     bool   `long:"ended"`
	Failed    bool   `long:"failed"`
	Sort      string `long:"sort" default:"updated_at"`
	Direction string `long:"dir" default:"desc"`
}

func (c *buildsListCmd) Execute(args []string) error {
	cl := cliClient

	opt := &sourcegraph.BuildListOptions{
		Repo:        c.Repo,
		CommitID:    c.CommitID,
		Active:      c.Active,
		Queued:      c.Queued,
		Succeeded:   c.Succeeded,
		Ended:       c.Ended,
		Failed:      c.Failed,
		Sort:        c.Sort,
		Direction:   c.Direction,
		ListOptions: sourcegraph.ListOptions{PerPage: 100},
	}

	for page := int32(1); ; page++ {
		opt.ListOptions.Page = page
		builds, err := cl.Builds.List(cliContext, opt)
		if err != nil {
			return err
		}

		if len(builds.Builds) == 0 {
			break
		}

		for _, b := range builds.Builds {
			if b.Success {
				fmt.Printf(green("#%s")+" succeeded % 9s ago", b.Spec().IDString(), ago(b.EndedAt.Time()))
			} else if b.Failure {
				fmt.Printf(red("#%s")+" failed % 9s ago", b.Spec().IDString(), ago(b.EndedAt.Time()))
			} else if b.StartedAt != nil {
				fmt.Printf(cyan("#%s")+" started % 9s ago", b.Spec().IDString(), ago(b.StartedAt.Time()))
			} else {
				fmt.Printf(gray("#%s")+" queued % 9s ago", b.Spec().IDString(), ago(b.CreatedAt.Time()))
			}
			fmt.Printf("\t%s\n", b.CommitID)
		}
	}

	return nil
}

func ago(t time.Time) string {
	d := time.Since(t)
	d = (d / time.Second) * time.Second
	return d.String()
}

type buildsStatsCmd struct{}

func (c *buildsStatsCmd) Execute(args []string) error {
	cl := cliClient

	numBuilds, err := statsutil.ComputeBuildStats(cl, cliContext)
	if err != nil {
		return err
	}

	if val, ok := numBuilds["succeeded"]; ok {
		fmt.Printf("%v successful\n", val)
	}
	if val, ok := numBuilds["failed"]; ok {
		fmt.Printf("%v failed\n", val)
	}
	if val, ok := numBuilds["active"]; ok {
		fmt.Printf("%v active\n", val)
	}
	if val, ok := numBuilds["queued"]; ok {
		fmt.Printf("%v queued\n", val)
	}
	if val, ok := numBuilds["purged"]; ok {
		fmt.Printf("%v purged\n", val)
	}
	if val, ok := numBuilds["total"]; ok {
		fmt.Printf("%v total\n", val)
	}

	return nil
}
