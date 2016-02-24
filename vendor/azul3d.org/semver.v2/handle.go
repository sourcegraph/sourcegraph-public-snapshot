// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package semver

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"sort"
	"strings"
)

// goGetTmpl is the HTML template that is served to the "go get" command line
// tool.
var goGetTmpl = template.Must(template.New("").Parse(`<html>
	<head>
		<meta name="go-import" content="{{.Prefix}} {{.VCS}} {{.RepoRoot}}">
		{{if .GoSource}}
			<meta name="go-source" content="{{.GoSource}}">
		{{end}}
	</head>
	<body>
go get {{.PkgPath}}
	</body>
</html>`))

// Handler implements a semantic versioning HTTP request handler.
type Handler struct {
	// The host of this application, e.g. "example.org".
	Host string

	// If set to true then HTTPS is not used by default when a request's URL
	// is missing a schema.
	NoSecure bool

	// The matcher used to resolve package URL's to their associated
	// repositories.
	Matcher

	// HTTP client to utilize for outgoing requests to Git servers, if nil then
	// http.DefaultClient is used.
	Client *http.Client
}

// Handle asks this handler to handle the given HTTP request by writing the
// appropriate response to the HTTP response writer.
func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) (s Status, err error) {
	// See if we can relate the requested URL to a repository URL.
	repo, err := h.Match(h.sanitize(r.Method, r.URL))
	if err != nil {
		if err == ErrNotPackageURL {
			// For an invalid package path, the request was unhandled and there
			// was no error.
			return Unhandled, nil
		}

		// For HTTP errors, we write a HTTP response and tell the caller the
		// request was handled OK.
		httpErr, ok := err.(*HTTPError)
		if ok {
			// Send the HTTP error.
			w.WriteHeader(httpErr.Status)
			fmt.Fprintf(w, "%s\n", httpErr)
			return Handled, nil
		}

		// For any other error, the request is unhandled and there was an
		// error.
		return Unhandled, err
	}

	// Default to HTTPS scheme.
	if repo.Scheme == "" {
		if h.NoSecure {
			repo.Scheme = "http"
		} else {
			repo.Scheme = "https"
		}
	}

	// Parse the query.
	query, _ := url.ParseQuery(r.URL.RawQuery)

	// POST git-upload-pack is responded to by simply redirecting their actual
	// request to the repository itself.
	if r.Method == "POST" && strings.HasSuffix(r.URL.Path, "/git-upload-pack") {
		target := &url.URL{
			Scheme: repo.Scheme,
			Host:   repo.Host,
			Path:   path.Join(repo.Path, "/git-upload-pack"),
		}
		w.Header().Set("Location", target.String())
		w.WriteHeader(http.StatusMovedPermanently)
		return Handled, nil
	}

	// GET info/refs?service=git-receive-pack is responded to by simply
	// redirecting their actual request to the repository itself.
	if r.Method == "GET" && strings.HasSuffix(r.URL.Path, "/info/refs") && query.Get("service") == "git-receive-pack" {
		target := &url.URL{
			Scheme:   repo.Scheme,
			Host:     repo.Host,
			Path:     path.Join(repo.Path, "/info/refs"),
			RawQuery: "service=git-receive-pack",
		}
		w.Header().Set("Location", target.String())
		w.WriteHeader(http.StatusMovedPermanently)
		return Handled, nil
	}

	// Create a URL to the target repo's /info/refs path.
	target := &url.URL{
		Scheme:   repo.Scheme,
		Host:     repo.URL.Host,
		Path:     path.Join(repo.URL.Path, "/info/refs"),
		RawQuery: "service=git-upload-pack",
	}

	// Modify the binary /info/refs blob. We do this now so that go get will
	// not find packages that do not exist.
	refs, err, status := h.modifyRefs(target, repo.Version)
	if err != nil {
		w.WriteHeader(status)
		fmt.Fprintf(w, "%s\n", err)
		return Handled, nil
	}

	// If the client is the `go get` tool, then we serve them a small template
	// that mostly just contains the go-import meta tag.
	if r.Method == "GET" && len(query.Get("go-get")) > 0 {
		pkgRoot := path.Join(h.Host, strings.TrimSuffix(r.URL.Path, repo.SubPath))
		repoRoot := repo.Scheme + "://" + pkgRoot
		err = goGetTmpl.Execute(w, map[string]interface{}{
			"VCS":      "git",
			"RepoRoot": repoRoot,
			"Prefix":   pkgRoot,
			"PkgPath":  path.Join(h.Host, r.URL.Path),
			"GoSource": repo.GoSource,
		})
		return Handled, err
	}

	// GET info/refs?service=git-upload-pack is responded to by fetching the
	// literal page at the remote repository, and serving a modified version of
	// it.
	if r.Method == "GET" && strings.HasSuffix(r.URL.Path, "/info/refs") && query.Get("service") == "git-upload-pack" {
		// Set correct content type header, return any IO error.
		w.Header().Set("Content-Type", "application/x-git-upload-pack-advertisement")
		_, err = io.Copy(w, bytes.NewReader(refs))
		return Handled, err
	}

	// It's no request that we recognize, but it is still a valid package URL
	// according to the Relate function. This means e.g. someone went to the
	// package page in their browser.
	return PkgPage, nil
}

// sanitize performs sanitiization of the given URL such that it can be passed
// directly to the relation function. It strips exactly the following prefixes
// from the URL:
//
//  /info/refs
//  /git-upload-pack
//  .git (some tools, e.g. godoc, incorrectly append this)
//
// Additionally, all query terms and any URL fragment are removed. The returned
// URL pointer is a copy (thus the original is unmodified).
func (h *Handler) sanitize(method string, u *url.URL) *url.URL {
	// Make a copy of the URL so we are not modifying the original one.
	cp := *u
	u = &cp

	// Remove all query terms and any URL fragment.
	u.RawQuery = ""
	u.Fragment = ""

	// Trim the suffixes.
	u.Path = strings.TrimSuffix(u.Path, "/info/refs")
	u.Path = strings.TrimSuffix(u.Path, "/git-upload-pack")
	u.Path = strings.TrimSuffix(u.Path, ".git")

	// Ensure the URL has a proper host.
	u.Host = h.Host
	return u
}

// modifyRefs downloads the given /info/refs URL and modifies it to download
// the given version branch/tag of the git repository.
//
// The returned integer is the HTTP status code to be sent in the event of an
// error.
func (h *Handler) modifyRefs(target *url.URL, v Version) ([]byte, error, int) {
	// Choose the appropriate HTTP client.
	client := h.Client
	if client == nil {
		client = http.DefaultClient
	}

	// Download the /info/refs?service=get-info-pack Git smart reply.
	resp, err := client.Get(target.String())
	if err != nil {
		return nil, err, http.StatusBadGateway
	}
	defer resp.Body.Close()

	// Read the entire body.
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err, http.StatusBadGateway
	}

	// Parse the info/refs data.
	refs, err := gitParseRefs(data)
	if err != nil {
		return nil, err, http.StatusInternalServerError
	}

	// Swap refs/heads/master record hash with our desired tag/branch hash.
	for _, ref := range refs.records {
		if ref.Name == "refs/heads/master" {
			hash, ok := h.chooseRef(refs.records, v)
			if !ok {
				// We don't actually have the requested version.
				return nil, fmt.Errorf("Requested version does not exist."), http.StatusNotFound
			}
			ref.Hash = hash
			break
		}
	}

	// Return the encoded and modified info/refs data.
	return refs.Bytes(), nil, http.StatusOK
}

type refVersion struct {
	Version
	*gitRef
}
type refsByVersion []refVersion

func (s refsByVersion) Len() int           { return len(s) }
func (s refsByVersion) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s refsByVersion) Less(i, j int) bool { return s[i].Version.Less(s[j].Version) }

// chooseRef chooses the best ref in the list for the given version. It returns
// ok=false if no ref could be chosen for the given version (i.e. the given
// version does not exist).
func (h *Handler) chooseRef(refs []*gitRef, v Version) (chosenHash string, ok bool) {
	var verList refsByVersion
	var master *gitRef
	for _, ref := range refs {
		// Trim the head and tags prefix. If the strings have different lengths
		// then we are certain it is a head or tag string.
		head := strings.TrimPrefix(ref.Name, "refs/heads/")
		isHead := len(head) != len(ref.Name)

		tag := strings.TrimPrefix(ref.Name, "refs/tags/")
		isTag := len(tag) != len(ref.Name)

		if !isTag && !isHead {
			// We're not interested (e.g. a pull request or something else).
			continue
		}

		// Store master reference.
		if head == "master" {
			master = ref
		}

		// Parse the version string.
		var refV Version
		if isHead {
			// A head ref.
			refV = ParseVersion(head)
		} else {
			// A tag ref.
			refV = ParseVersion(tag)
		}

		// Ensure that the major versions (and unstable statuses) match the one
		// we desire. If they don't then we skip this version.
		if refV.Major != v.Major || refV.Unstable != v.Unstable {
			continue
		}

		// Add it to the version list for sorting.
		verList = append(verList, refVersion{
			Version: refV,
			gitRef:  ref,
		})
	}

	if len(verList) == 0 {
		// No branch/tag with that version. If we wanted v0 then we can just
		// use the master branch.
		if v.Major == 0 {
			return master.BestHash(), true
		}
		return "", false
	}

	// Sort the version list.
	sort.Sort(sort.Reverse(verList))

	// What if the version list contains both a tag and branch with the same
	// version? We always choose the branch. Do this now.
	if len(verList) >= 2 {
		if verList[0].Version == verList[1].Version {
			// The versions are the same. Which is the branch?
			if strings.HasPrefix(verList[0].Name, "refs/heads") {
				return verList[0].BestHash(), true
			}

			// It's the other one then.
			return verList[1].BestHash(), true
		}
	}

	return verList[0].BestHash(), true
}
