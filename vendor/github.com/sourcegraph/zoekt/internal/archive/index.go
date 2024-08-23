// package archive provides indexing of archives from remote URLs.
package archive

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/sourcegraph/zoekt"
	"github.com/sourcegraph/zoekt/build"
)

// Options specify the archive specific indexing options.
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

// Index archive specified in opts using bopts.
func Index(opts Options, bopts build.Options) error {
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

	if opts.Incremental && bopts.IncrementalSkipIndexing() {
		return nil
	}

	a, err := openArchive(opts.Archive)
	if err != nil {
		return err
	}
	defer a.Close()

	bopts.RepositoryDescription.Source = opts.Archive
	builder, err := build.NewBuilder(bopts)
	if err != nil {
		return err
	}

	add := func(f *File) error {
		defer f.Close()

		contents, err := io.ReadAll(f)
		if err != nil {
			return err
		}

		name := stripComponents(f.Name, opts.Strip)
		if name == "" {
			return nil
		}

		return builder.Add(zoekt.Document{
			Name:     name,
			Content:  contents,
			Branches: brs,
		})
	}

	for {
		f, err := a.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if err := add(f); err != nil {
			return err
		}
	}

	return builder.Finish()
}

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
