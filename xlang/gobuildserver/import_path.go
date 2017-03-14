package gobuildserver

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/httputil"
)

// Adapted from github.com/golang/gddo/gosrc.

// noGoGetDomains is a list of domains we do not attempt standard go vanity
// import resolution. Instead we take an educated guess based on the URL how
// to create the directory struct.
var noGoGetDomains = strings.Split(env.Get("NO_GO_GET_DOMAINS", "", "list of domains to NOT perform go get on. Separated by ','"), ",")

type directory struct {
	importPath  string // the Go import path for this package
	projectRoot string // import path prefix for all packages in the project
	cloneURL    string // the VCS clone URL
	repoPrefix  string // the path to this directory inside the repo, if set
	vcs         string // one of "git", "hg", "svn", etc.
	rev         string // the VCS revision specifier, if any
}

var errNoMatch = errors.New("no match")

// ResolveImportPathCloneURL returns the clone URL for a Go package import path.
func ResolveImportPathCloneURL(importPath string) (string, error) {
	d, err := resolveImportPath(httputil.CachingClient, importPath)
	if err != nil {
		return "", err
	}
	return d.cloneURL, nil
}

func resolveImportPath(client *http.Client, importPath string) (*directory, error) {
	if d, err := resolveStaticImportPath(importPath); err == nil {
		return d, nil
	} else if err != nil && err != errNoMatch {
		return nil, err
	}
	return resolveDynamicImportPath(client, importPath)
}

func resolveStaticImportPath(importPath string) (*directory, error) {
	if _, isStdlib := stdlibPackagePaths[importPath]; isStdlib {
		return &directory{
			importPath:  importPath,
			projectRoot: "",
			cloneURL:    "https://github.com/golang/go",
			repoPrefix:  "src",
			vcs:         "git",
			rev:         RuntimeVersion,
		}, nil
	}

	// This allows a user to set a list of domains that are considered to be
	// non-go-gettable, i.e. standard git repositories. Some on-prem customers
	// use setups like this, where they directly import non-go-gettable git
	// repository URLs like "mygitolite.aws.me.org/mux.git/subpkg"
	for _, domain := range noGoGetDomains {
		domain = strings.TrimSpace(domain)
		if domain == "" {
			continue
		}
		if !strings.HasPrefix(importPath, domain) {
			continue
		}

		if !strings.Contains(importPath, ".git") {
			// Assume GitHub-like where two path elements is the project
			// root.
			parts := strings.SplitN(importPath, "/", 4)
			if len(parts) < 3 {
				return nil, fmt.Errorf("invalid GitHub-like import path: %q", importPath)
			}
			repo := parts[0] + "/" + parts[1] + "/" + parts[2]
			return &directory{
				importPath:  importPath,
				projectRoot: repo,
				cloneURL:    "http://" + repo,
				vcs:         "git",
			}, nil
		}

		// TODO(slimsag): We assume that .git only shows up
		// once in the import path. Not always true, but generally
		// should be in 99% of cases.
		split := strings.Split(importPath, ".git")
		if len(split) != 2 {
			return nil, fmt.Errorf("expected one .git in %q", importPath)
		}

		return &directory{
			importPath:  importPath,
			projectRoot: split[0] + ".git",
			cloneURL:    "http://" + split[0] + ".git",
			vcs:         "git",
		}, nil
	}

	switch {
	case strings.HasPrefix(importPath, "github.com/"):
		parts := strings.SplitN(importPath, "/", 4)
		if len(parts) < 3 {
			return nil, fmt.Errorf("invalid github.com/golang.org import path: %q", importPath)
		}
		repo := parts[0] + "/" + parts[1] + "/" + parts[2]
		return &directory{
			importPath:  importPath,
			projectRoot: repo,
			cloneURL:    "https://" + repo,
			vcs:         "git",
		}, nil

	case strings.HasPrefix(importPath, "golang.org/x/"):
		d, err := resolveStaticImportPath(strings.Replace(importPath, "golang.org/x/", "github.com/golang/", 1))
		if err != nil {
			return nil, err
		}
		d.projectRoot = strings.Replace(d.projectRoot, "github.com/golang/", "golang.org/x/", 1)
		return d, nil
	}
	return nil, errNoMatch
}

// gopkgSrcTemplate matches the go-source dir templates specified by the
// popular gopkg.in
var gopkgSrcTemplate = regexp.MustCompile(`https://(github.com/[^/]*/[^/]*)/tree/([^/]*)\{/dir\}`)

func resolveDynamicImportPath(client *http.Client, importPath string) (*directory, error) {
	metaProto, im, sm, err := fetchMeta(client, importPath)
	if err != nil {
		return nil, err
	}

	if im.prefix != importPath {
		var imRoot *importMeta
		metaProto, imRoot, _, err = fetchMeta(client, im.prefix)
		if err != nil {
			return nil, err
		}
		if *imRoot != *im {
			return nil, fmt.Errorf("project root mismatch: %q != %q", *imRoot, *im)
		}
	}

	// clonePath is the repo URL from import meta tag, with the "scheme://" prefix removed.
	// It should be used for cloning repositories.
	// repo is the repo URL from import meta tag, with the "scheme://" prefix removed, and
	// a possible ".vcs" suffix trimmed.
	i := strings.Index(im.repo, "://")
	if i < 0 {
		return nil, fmt.Errorf("bad repo URL: %s", im.repo)
	}
	clonePath := im.repo[i+len("://"):]
	repo := strings.TrimSuffix(clonePath, "."+im.vcs)
	dirName := importPath[len(im.prefix):]

	var dir *directory
	if sm != nil {
		m := gopkgSrcTemplate.FindStringSubmatch(sm.dirTemplate)
		if len(m) > 0 {
			// We are doing best effort, so we ignore err
			dir, _ = resolveStaticImportPath(m[1] + dirName)
			if dir != nil {
				dir.rev = m[2]
			}
		}
	}

	if dir == nil {
		// We are doing best effort, so we ignore err
		dir, _ = resolveStaticImportPath(repo + dirName)
	}

	if dir == nil {
		dir = &directory{}
	}
	dir.importPath = importPath
	dir.projectRoot = im.prefix
	if dir.cloneURL == "" {
		dir.cloneURL = metaProto + "://" + repo + "." + im.vcs
	}
	dir.vcs = im.vcs
	return dir, nil
}

// importMeta represents the values in a go-import meta tag.
//
// See https://golang.org/cmd/go/#hdr-Remote_import_paths.
type importMeta struct {
	prefix string // the import path corresponding to the repository root
	vcs    string // one of "git", "hg", "svn", etc.
	repo   string // root of the VCS repo containing a scheme and not containing a .vcs qualifier
}

// sourceMeta represents the values in a go-source meta tag.
type sourceMeta struct {
	prefix       string
	projectURL   string
	dirTemplate  string
	fileTemplate string
}

func fetchMeta(client *http.Client, importPath string) (scheme string, im *importMeta, sm *sourceMeta, err error) {
	uri := importPath
	if !strings.Contains(uri, "/") {
		// Add slash for root of domain.
		uri = uri + "/"
	}
	uri = uri + "?go-get=1"

	scheme = "https"
	resp, err := client.Get(scheme + "://" + uri)
	if err != nil || resp.StatusCode != 200 {
		if err == nil {
			resp.Body.Close()
		}
		scheme = "http"
		resp, err = client.Get(scheme + "://" + uri)
		if err != nil {
			return scheme, nil, nil, err
		}
	}
	defer resp.Body.Close()
	im, sm, err = parseMeta(scheme, importPath, resp.Body)
	return scheme, im, sm, err
}

func parseMeta(scheme, importPath string, r io.Reader) (im *importMeta, sm *sourceMeta, err error) {
	errorMessage := "go-import meta tag not found"

	d := xml.NewDecoder(r)
	d.Strict = false
metaScan:
	for {
		t, tokenErr := d.Token()
		if tokenErr != nil {
			break metaScan
		}
		switch t := t.(type) {
		case xml.EndElement:
			if strings.EqualFold(t.Name.Local, "head") {
				break metaScan
			}
		case xml.StartElement:
			if strings.EqualFold(t.Name.Local, "body") {
				break metaScan
			}
			if !strings.EqualFold(t.Name.Local, "meta") {
				continue metaScan
			}
			nameAttr := attrValue(t.Attr, "name")
			if nameAttr != "go-import" && nameAttr != "go-source" {
				continue metaScan
			}
			fields := strings.Fields(attrValue(t.Attr, "content"))
			if len(fields) < 1 {
				continue metaScan
			}
			prefix := fields[0]
			if !strings.HasPrefix(importPath, prefix) ||
				!(len(importPath) == len(prefix) || importPath[len(prefix)] == '/') {
				// Ignore if root is not a prefix of the  path. This allows a
				// site to use a single error page for multiple repositories.
				continue metaScan
			}
			switch nameAttr {
			case "go-import":
				if len(fields) != 3 {
					errorMessage = "go-import meta tag content attribute does not have three fields"
					continue metaScan
				}
				if im != nil {
					im = nil
					errorMessage = "more than one go-import meta tag found"
					break metaScan
				}
				im = &importMeta{
					prefix: prefix,
					vcs:    fields[1],
					repo:   fields[2],
				}
			case "go-source":
				if sm != nil {
					// Ignore extra go-source meta tags.
					continue metaScan
				}
				if len(fields) != 4 {
					continue metaScan
				}
				sm = &sourceMeta{
					prefix:       prefix,
					projectURL:   fields[1],
					dirTemplate:  fields[2],
					fileTemplate: fields[3],
				}
			}
		}
	}
	if im == nil {
		return nil, nil, fmt.Errorf("%s at %s://%s", errorMessage, scheme, importPath)
	}
	if sm != nil && sm.prefix != im.prefix {
		sm = nil
	}
	return im, sm, nil
}

func attrValue(attrs []xml.Attr, name string) string {
	for _, a := range attrs {
		if strings.EqualFold(a.Name.Local, name) {
			return a.Value
		}
	}
	return ""
}
