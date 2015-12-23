package sgx

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/sgx/cli"
	"src.sourcegraph.com/sourcegraph/util/statsutil"
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

	_, err = buildsGroup.AddCommand("list",
		"list builds",
		"The list subcommand lists builds by the specified criteria.",
		&buildsListCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}

	_, err = buildsGroup.AddCommand("repo",
		"get repo build info",
		"The get-repo subcommand gets the latest repo build.",
		&buildsGetRepoBuildInfoCmd{},
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

type buildsListCmd struct {
	N         int    `short:"n" description:"number of builds to show" default:"5"`
	Repo      string `long:"repo" description:"repo URI"`
	CommitID  string `long:"commit" description:"filter builds by commit ID"`
	Queued    bool   `long:"queued"`
	Succeeded bool   `long:"succeeded"`
	Ended     bool   `long:"ended"`
	Failed    bool   `long:"failed"`
	Sort      string `long:"sort" default:"updated_at"`
	Direction string `long:"dir" default:"desc"`
}

func (c *buildsListCmd) Execute(args []string) error {
	cl := cli.Client()

	opt := &sourcegraph.BuildListOptions{
		Repo:        c.Repo,
		CommitID:    c.CommitID,
		Queued:      c.Queued,
		Succeeded:   c.Succeeded,
		Ended:       c.Ended,
		Failed:      c.Failed,
		Sort:        c.Sort,
		Direction:   c.Direction,
		ListOptions: sourcegraph.ListOptions{PerPage: int32(c.N)},
	}
	builds, err := cl.Builds.List(cli.Ctx, opt)
	if err != nil {
		return err
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

	return nil
}

type buildsGetRepoBuildInfoCmd struct {
	Args struct {
		Repo []string `name:"repositories to fetch build info for"`
	} `positional-args:"yes"`
}

func (c *buildsGetRepoBuildInfoCmd) Execute(args []string) error {
	cl := cli.Client()

	for _, repo := range c.Args.Repo {
		repo, rev := sourcegraph.ParseRepoAndCommitID(repo)
		build, err := cl.Builds.GetRepoBuild(cli.Ctx,
			&sourcegraph.RepoRevSpec{
				RepoSpec: sourcegraph.RepoSpec{URI: repo},
				Rev:      rev,
			},
		)
		if err != nil {
			return err
		}
		fmt.Println(repo)
		b, err := json.MarshalIndent(build, "\t", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(b))
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
	cl := cli.Client()

	numBuilds, err := statsutil.ComputeBuildStats(cl, cli.Ctx)
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
