package buildserver

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"
)

// Adapted from github.com/golang/gddo/gosrc.

// directory represents a directory on a version control service.
type directory struct {
	importPath   string // the Go import path for this package
	projectRoot  string // import path prefix for all packages in the project
	resolvedPath string // import path after resolving go-import meta tags, if any
	cloneURL     string // the VCS clone URL
	repoPrefix   string // the path to this directory inside the repo, if set
	vcs          string // one of "git", "hg", "svn", etc.
	rev          string // the VCS revision specifier, if any
}

var errNoMatch = errors.New("no match")

func resolveImportPath(client *http.Client, importPath string) (*directory, error) {
	if d, err := resolveStaticImportPath(importPath); err == nil {
		return d, nil
	} else if err != nil && err != errNoMatch {
		return d, nil
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
			rev:         runtime.Version(),
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
			importPath:   importPath,
			resolvedPath: importPath + ".git",
			projectRoot:  repo,
			cloneURL:     "https://" + repo,
			vcs:          "git",
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

func resolveDynamicImportPath(client *http.Client, importPath string) (*directory, error) {
	metaProto, im, err := fetchMeta(client, importPath)
	if err != nil {
		return nil, err
	}

	if im.prefix != importPath {
		var imRoot *importMeta
		metaProto, imRoot, err = fetchMeta(client, im.prefix)
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

	resolvedPath := repo + dirName
	dir, err := resolveStaticImportPath(resolvedPath)
	if err == errNoMatch {
		resolvedPath = repo + "." + im.vcs + dirName
	}

	if dir == nil {
		dir = &directory{}
	}
	dir.importPath = importPath
	dir.projectRoot = im.prefix
	dir.resolvedPath = resolvedPath
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

func fetchMeta(client *http.Client, importPath string) (scheme string, im *importMeta, err error) {
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
			return scheme, nil, err
		}
	}
	defer resp.Body.Close()
	im, err = parseMeta(scheme, importPath, resp.Body)
	return scheme, im, err
}

func parseMeta(scheme, importPath string, r io.Reader) (im *importMeta, err error) {
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
			if nameAttr != "go-import" {
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
		}
	}
	if im == nil {
		return nil, fmt.Errorf("%s at %s://%s", errorMessage, scheme, importPath)
	}
	return im, nil
}

func attrValue(attrs []xml.Attr, name string) string {
	for _, a := range attrs {
		if strings.EqualFold(a.Name.Local, name) {
			return a.Value
		}
	}
	return ""
}
