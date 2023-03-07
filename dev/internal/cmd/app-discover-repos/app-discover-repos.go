package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/service/servegit"
	"github.com/sourcegraph/sourcegraph/internal/singleprogram/filepicker"
)

const usage = `

app-discover-repos runs the same discovery logic used by app to discover local
repositories. It will print some additional debug information.`

func main() {
	liblog := log.Init(log.Resource{
		Name:       "app-discover-repos",
		Version:    "dev",
		InstanceID: os.Getenv("HOSTNAME"),
	})
	defer liblog.Sync()

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n\n%s\n\n", os.Args[0], strings.TrimSpace(usage))
		flag.PrintDefaults()
	}

	var c servegit.Config
	c.Load()

	root := flag.String("root", c.Root, "the directory we search from.")
	block := flag.Bool("block", false, "by default we stream out the repos we find. This is not exactly what sourcegraph uses, so enable this flag for the same behaviour.")
	picker := flag.Bool("picker", false, "try run the file picker.")
	verbose := flag.Bool("v", false, "verbose output")

	flag.Parse()

	if *picker {
		p, ok := filepicker.Lookup(log.Scoped("picker", ""))
		if !ok {
			fmt.Fprintf(os.Stderr, "filepicker not found\n")
		} else {
			path, err := p(context.Background())
			if err != nil {
				fmt.Fprintf(os.Stderr, "filepicker error: %v\n", err)
				os.Exit(1)
			}
			fmt.Fprintf(os.Stderr, "filepicker picked %q\n", path)
			*root = path
		}
	}

	c.Root = *root

	srv := &servegit.Serve{
		Config: c,
		Logger: log.Scoped("serve", ""),
	}

	printRepo := func(r servegit.Repo) {
		if *verbose {
			fmt.Printf("%s\t%s\t%s\n", r.Name, r.URI, r.ClonePath)
		} else {
			fmt.Println(r.Name)
		}
	}

	if *block {
		repos, err := srv.Repos()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Repos returned error: %v\n", err)
			os.Exit(1)
		}
		for _, r := range repos {
			printRepo(r)
		}
	} else {
		repoC := make(chan servegit.Repo, 4)
		go func() {
			defer close(repoC)
			err := srv.Walk(repoC)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Walk returned error: %v\n", err)
				os.Exit(1)
			}
		}()
		for r := range repoC {
			printRepo(r)
		}
	}
}
