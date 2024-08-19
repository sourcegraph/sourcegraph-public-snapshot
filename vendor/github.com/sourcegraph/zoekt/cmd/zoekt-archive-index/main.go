// Command zoekt-archive-index indexes an archive.
//
// Example via github.com:
//
//	zoekt-archive-index -incremental -commit b57cb1605fd11ba2ecfa7f68992b4b9cc791934d -name github.com/gorilla/mux -strip_components 1 https://codeload.github.com/gorilla/mux/legacy.tar.gz/b57cb1605fd11ba2ecfa7f68992b4b9cc791934d
//
//	zoekt-archive-index -branch master https://github.com/gorilla/mux/commit/b57cb1605fd11ba2ecfa7f68992b4b9cc791934d
package main

import (
	"flag"
	"log"

	"go.uber.org/automaxprocs/maxprocs"

	"github.com/sourcegraph/zoekt/cmd"
	"github.com/sourcegraph/zoekt/internal/archive"
)

func main() {
	var (
		incremental = flag.Bool("incremental", true, "only index changed repositories")

		name   = flag.String("name", "", "The repository name for the archive")
		urlRaw = flag.String("url", "", "The repository URL for the archive")
		branch = flag.String("branch", "", "The branch name for the archive")
		commit = flag.String("commit", "", "The commit sha for the archive. If incremental this will avoid updating shards already at commit")
		strip  = flag.Int("strip_components", 0, "Remove the specified number of leading path elements. Pathnames with fewer elements will be silently skipped.")

		downloadLimitMbps = flag.Int64("download-limit-mbps", 0, "If non-zero, limit archive downloads to specified amount in megabits per second")
	)
	flag.Parse()

	// Tune GOMAXPROCS to match Linux container CPU quota.
	_, _ = maxprocs.Set()

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	if len(flag.Args()) != 1 {
		log.Fatal("expected argument for archive location")
	}
	archiveURL := flag.Args()[0]
	bopts := cmd.OptionsFromFlags()
	opts := archive.Options{
		Incremental: *incremental,

		Archive: archiveURL,
		Name:    *name,
		RepoURL: *urlRaw,
		Branch:  *branch,
		Commit:  *commit,
		Strip:   *strip,
	}

	// Sourcegraph specific: Limit HTTP traffic
	limitHTTPDefaultClient(*downloadLimitMbps)

	if err := archive.Index(opts, *bopts); err != nil {
		log.Fatal(err)
	}
}
