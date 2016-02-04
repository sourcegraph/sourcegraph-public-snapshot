// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package semver

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"
)

var rePkgVersion = regexp.MustCompile(`^([a-zA-Z0-9-]+).(v[0-9]+[\.]?[0-9]*[\.]?[0-9]*(?:\-unstable)?)`)

// github is a Matcher that represents a single GitHub user or organization.
type github string

// githubGoSource returns a go-source meta-tag for the given repository and go
// get URL.
func githubGoSource(r *Repo, u *url.URL) string {
	// The Go package path corresponding to the repository root, for example:
	//
	//  right: azul3d.org/gfx.v2
	//  wrong: azul3d.org/gfx.v2/window
	//
	prefix := path.Join(u.Host, strings.TrimSuffix(u.Path, r.SubPath))

	// Default to godoc.org's home:
	home := "_"

	// A basic GitHub repository URL.
	ghURL := *r.URL
	if len(ghURL.Scheme) == 0 {
		ghURL.Scheme = "https"
	}

	// Build a directory-view URL like so:
	//
	//  https://github.com/go-yaml/yaml/tree/v2{/dir}
	//
	dirURL := ghURL
	dirURL.Path = path.Join(dirURL.Path, "tree", r.Version.String())
	dir := dirURL.String() + "{/dir}"

	// Build a file-view URL like so:
	//
	//  https://github.com/go-yaml/yaml/blob/v2{/dir}/{file}#L{line}
	//
	fileURL := ghURL
	fileURL.Path = path.Join(fileURL.Path, "blob", r.Version.String())
	file := fileURL.String() + "{/dir}/{file}#L{line}"

	return strings.Join([]string{prefix, home, dir, file}, " ")
}

// Match implements the Matcher interface.
func (user github) Match(u *url.URL) (repo *Repo, err error) {
	// Split the path elements. If any element is an empty string then it
	// is because there are two consecutive slashes ("/a//b/c") or the path
	// ends with a trailing slash ("example.com/pkg.v1/").
	//
	// If more than one element contains a version match then it's invalid
	// as well ("example.com/foo.v1/bar.v1/something.v2").
	var (
		rel         = strings.TrimPrefix(u.Path, "/")
		s           = strings.Split(rel, "/")
		versionElem = -1   // Index of version element in s.
		version     string // e.g. "v3".
		pkgName     string // e.g. "pkg" from "foo/bar/pkg.v3/sub".
	)
	for index, elem := range s {
		if len(elem) == 0 {
			// Path has two consecutive slashes ("/a//b/c") or ends with
			// trailing slash.
			return nil, ErrNotPackageURL
		}
		m := rePkgVersion.FindStringSubmatch(elem)
		if m != nil {
			if versionElem != -1 {
				// Multiple versions in path.
				return nil, ErrNotPackageURL
			}
			pkgName = m[1]
			version = m[2]
			versionElem = index
		}
	}
	if versionElem == -1 {
		// No version in path.
		return nil, ErrNotPackageURL
	}

	// Parse the version string.
	v := ParseVersion(version)
	if v.Minor > 0 || v.Patch > 0 {
		return nil, &HTTPError{
			error:  fmt.Errorf("Import path may only contain major version."),
			Status: http.StatusNotFound,
		}
	}

	// Everything in the path up to the path element index [found] is part
	// of the repository name. We replace all slashes with dashes (the same
	// thing GitHub does if you try to create a repository with slashes in
	// the name).
	repoName := strings.Join(append(s[:versionElem], pkgName), "-")
	repoSubPath := strings.Join(s[versionElem+1:], "/")
	repo = &Repo{
		Version: v,
		SubPath: repoSubPath,
		URL: &url.URL{
			Scheme: u.Scheme,
			Host:   "github.com",
			Path:   path.Join(string(user), repoName),
		},
	}

	// Attach the go-source meta-tag.
	repo.GoSource = githubGoSource(repo, u)

	// TODO(slimsag): godoc.org requires that repos end in .git: very strange.
	repo.URL.Path += ".git"
	return
}

// GitHub returns a URL Matcher that operates on a single GitHub user or
// organization. For instance if the service was running at example.com and the
// user string was "bob", it would match URLS in the pattern of:
//
//  example.com/pkg.v3 → github.com/bob/pkg (branch/tag v3, v3.N, or v3.N.M)
//  example.com/folder/pkg.v3 → github.com/bob/folder-pkg (branch/tag v3, v3.N, or v3.N.M)
//  example.com/multi/folder/pkg.v3 → github.com/bob/multi-folder-pkg (branch/tag v3, v3.N, or v3.N.M)
//  example.com/folder/pkg.v3/subpkg → github.com/bob/folder-pkg (branch/tag v3, v3.N, or v3.N.M)
//  example.com/pkg.v3/folder/subpkg → github.com/bob/pkg (branch/tag v3, v3.N, or v3.N.M)
//  example.com/pkg.v3-unstable → github.com/bob/pkg (branch/tag v3-unstable, v3.N-unstable, or v3.N.M-unstable)
//
func GitHub(user string) Matcher {
	return github(user)
}
