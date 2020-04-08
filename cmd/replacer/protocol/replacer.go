// Package protocol contains structures used by the replacer API.
package protocol

import (
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
)

// Request represents a request to replacer
type Request struct {
	// Repo is the name of the repository to search. eg "github.com/gorilla/mux"
	Repo api.RepoName

	// URL specifies the repository's Git remote URL (for gitserver). It is optional. See
	// (gitserver.ExecRequest).URL for documentation on what it is used for.
	URL string

	// Commit is which commit to search. It is required to be resolved,
	// not a ref like HEAD or master. eg
	// "599cba5e7b6137d46ddf58fb1765f5d928e69604"
	Commit api.CommitID

	// The amount of time to wait for a repo archive to fetch.
	// It is parsed with time.ParseDuration.
	//
	// This timeout should be low when searching across many repos
	// so that unfetched repos don't delay the search, and because we are likely
	// to get results from the repos that have already been fetched.
	//
	// This timeout should be high when searching across a single repo
	// because returning results slowly is better than returning no results at all.
	//
	// This only times out how long we wait for the fetch request;
	// the fetch will still happen in the background so future requests don't have to wait.
	FetchTimeout string

	RewriteSpecification
}

type RewriteSpecification struct {
	// A template pattern that expresses what to match.
	MatchTemplate string

	// A template pattern that expresses how matches should be rewritten.
	RewriteTemplate string

	// A file extension suffix filtering which files to process (e.g., ".go")
	FileExtension string

	// A directory prefix to exclude (e.g., vendor)
	DirectoryExclude string
}

// GitserverRepo returns the repository information necessary to perform gitserver requests.
func (r Request) GitserverRepo() gitserver.Repo { return gitserver.Repo{Name: r.Repo, URL: r.URL} }
