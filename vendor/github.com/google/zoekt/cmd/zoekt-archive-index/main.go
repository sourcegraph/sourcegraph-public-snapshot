// Command zoekt-archive-index indexes an archive.
//
// Example via github.com:
//
//   zoekt-archive-index -incremental -commit b57cb1605fd11ba2ecfa7f68992b4b9cc791934d -name github.com/gorilla/mux -strip_components 1 https://codeload.github.com/gorilla/mux/legacy.tar.gz/b57cb1605fd11ba2ecfa7f68992b4b9cc791934d
//
//   zoekt-archive-index -branch master https://github.com/gorilla/mux/commit/b57cb1605fd11ba2ecfa7f68992b4b9cc791934d
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"reflect"
	"strings"

	"github.com/google/zoekt"
	"github.com/google/zoekt/build"
)

// stripComponents removes the specified number of leading path
// elements. Pathnames with fewer elements will return the empty string.
func stripComponents(path string, count int) string {
	for i := 0; path != "" && i < count; i++ {
		i := strings.Index(path, "/")
		if i < 0 {
			return ""
		}
		path = path[i+1:]
	}
	return path
}

// isGitOID checks if the revision is a git OID SHA string.
//
// Note: This doesn't mean the SHA exists in a repository, nor does it mean it
// isn't a ref. Git allows 40-char hexadecimal strings to be references.
func isGitOID(s string) bool {
	if len(s) != 40 {
		return false
	}
	for _, r := range s {
		if !(('0' <= r && r <= '9') ||
			('a' <= r && r <= 'f') ||
			('A' <= r && r <= 'F')) {
			return false
		}
	}
	return true
}

type Options struct {
	Incremental bool

	Archive string
	Name    string
	RepoURL string
	Branch  string
	Commit  string
	Strip   int
}

func (o *Options) SetDefaults() {
	// We guess based on the archive URL.
	u, _ := url.Parse(o.Archive)
	if u == nil {
		return
	}

	setRef := func(ref string) {
		if isGitOID(ref) && o.Commit == "" {
			o.Commit = ref
		}
		if !isGitOID(ref) && o.Branch == "" {
			o.Branch = ref
		}
	}

	switch u.Host {
	case "github.com", "codeload.github.com":
		// https://github.com/octokit/octokit.rb/commit/3d21ec53a331a6f037a91c368710b99387d012c1
		// https://github.com/octokit/octokit.rb/blob/master/README.md
		// https://github.com/octokit/octokit.rb/tree/master/lib
		// https://codeload.github.com/octokit/octokit.rb/legacy.tar.gz/master
		parts := strings.Split(u.Path, "/")
		if len(parts) > 2 && o.Name == "" {
			o.Name = fmt.Sprintf("github.com/%s/%s", parts[1], parts[2])
			o.RepoURL = fmt.Sprintf("https://github.com/%s/%s", parts[1], parts[2])
		}
		if len(parts) > 4 {
			setRef(parts[4])
			if u.Host == "github.com" {
				o.Archive = fmt.Sprintf("https://codeload.github.com/%s/%s/legacy.tar.gz/%s", parts[1], parts[2], parts[4])
			}
		}
		o.Strip = 1
	case "api.github.com":
		// https://api.github.com/repos/octokit/octokit.rb/tarball/master
		parts := strings.Split(u.Path, "/")
		if len(parts) > 2 && o.Name == "" {
			o.Name = fmt.Sprintf("github.com/%s/%s", parts[1], parts[2])
			o.RepoURL = fmt.Sprintf("https://github.com/%s/%s", parts[1], parts[2])
		}
		if len(parts) > 5 {
			setRef(parts[5])
		}
		o.Strip = 1
	}
}

func do(opts Options, bopts build.Options) error {
	opts.SetDefaults()

	if opts.Name == "" && opts.RepoURL == "" {
		return errors.New("-name or -url required")
	}
	if opts.Branch == "" {
		return errors.New("-branch required")
	}

	if opts.Name != "" {
		bopts.RepositoryDescription.Name = opts.Name
	}
	// We do not use this functionality to avoid pulling in the transitive deps of gitindex
	/*
		if opts.RepoURL != "" {
			u, err := url.Parse(opts.RepoURL)
			if err != nil {
				return err
			}
			if err := gitindex.SetTemplatesFromOrigin(&bopts.RepositoryDescription, u); err != nil {
				return err
			}
		}
	*/
	bopts.SetDefaults()
	bopts.RepositoryDescription.Branches = []zoekt.RepositoryBranch{{Name: opts.Branch, Version: opts.Commit}}
	brs := []string{opts.Branch}

	if opts.Incremental {
		versions := bopts.IndexVersions()
		if reflect.DeepEqual(versions, bopts.RepositoryDescription.Branches) {
			return nil
		}
	}

	a, err := openArchive(opts.Archive)
	if err != nil {
		return err
	}
	defer a.Close()

	builder, err := build.NewBuilder(bopts)
	if err != nil {
		return err
	}

	for {
		f, err := a.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// We do not index large files
		if f.Size > int64(bopts.SizeMax) {
			continue
		}

		contents, err := ioutil.ReadAll(f)
		if err != nil {
			return err
		}

		name := stripComponents(f.Name, opts.Strip)
		if name == "" {
			continue
		}

		err = builder.Add(zoekt.Document{
			Name:     name,
			Content:  contents,
			Branches: brs,
		})
		if err != nil {
			return err
		}
	}

	return builder.Finish()
}

func main() {
	var (
		sizeMax     = flag.Int("file_limit", 128*1024, "maximum file size")
		shardLimit  = flag.Int("shard_limit", 100<<20, "maximum corpus size for a shard")
		parallelism = flag.Int("parallelism", 4, "maximum number of parallel indexing processes.")
		indexDir    = flag.String("index", build.DefaultDir, "index directory for *.zoekt files.")
		incremental = flag.Bool("incremental", true, "only index changed repositories")
		ctags       = flag.Bool("require_ctags", false, "If set, ctags calls must succeed.")

		name   = flag.String("name", "", "The repository name for the archive")
		urlRaw = flag.String("url", "", "The repository URL for the archive")
		branch = flag.String("branch", "", "The branch name for the archive")
		commit = flag.String("commit", "", "The commit sha for the archive. If incremental this will avoid updating shards already at commit")
		strip  = flag.Int("strip_components", 0, "Remove the specified number of leading path elements. Pathnames with fewer elements will be silently skipped.")
	)
	flag.Parse()

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	if len(flag.Args()) != 1 {
		log.Fatal("expected argument for archive location")
	}
	archive := flag.Args()[0]

	bopts := build.Options{
		Parallelism:      *parallelism,
		SizeMax:          *sizeMax,
		ShardMax:         *shardLimit,
		IndexDir:         *indexDir,
		CTagsMustSucceed: *ctags,
	}
	opts := Options{
		Incremental: *incremental,

		Archive: archive,
		Name:    *name,
		RepoURL: *urlRaw,
		Branch:  *branch,
		Commit:  *commit,
		Strip:   *strip,
	}

	if err := do(opts, bopts); err != nil {
		log.Fatal(err)
	}
}
