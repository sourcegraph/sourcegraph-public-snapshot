package sourcegraph

import (
	"encoding/json"
	"net/url"
)

// IsSystemOfRecord returns true iff this repository is the source of truth (not a mirror, etc)
func (r *Repo) IsSystemOfRecord() bool {
	return !r.Mirror
}

// Returns the repository's canonical clone URL
func (r *Repo) CloneURL() *url.URL {
	var cloneURL string
	if r.HTTPCloneURL != "" {
		cloneURL = r.HTTPCloneURL
	} else if r.SSHCloneURL != "" {
		cloneURL = string(r.SSHCloneURL)
	} else {
		cloneURL = r.URI
	}
	u, _ := url.Parse(cloneURL)
	return u
}

// RepoSpec returns the RepoSpec that specifies r.
func (r *Repo) RepoSpec() RepoSpec {
	return RepoSpec{URI: r.URI}
}

// IsZero reports whether s.URI is the zero value.
func (s RepoSpec) IsZero() bool { return s.URI == "" }

// Resolved reports whether the revspec has been fully resolved to an
// absolute (40-char) commit ID.
func (s RepoRevSpec) Resolved() bool {
	return s.Rev != "" && len(s.CommitID) == 40
}

func (r *RepoResolution) UnmarshalJSON(data []byte) error {
	var m struct {
		Result struct {
			Repo       *RepoSpec
			RemoteRepo *RemoteRepo
		}
	}
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	switch {
	case m.Result.Repo != nil:
		*r = RepoResolution{Result: &RepoResolution_Repo{Repo: m.Result.Repo}}
	case m.Result.RemoteRepo != nil:
		*r = RepoResolution{Result: &RepoResolution_RemoteRepo{RemoteRepo: m.Result.RemoteRepo}}
	}
	return nil
}

func (r *ReposCreateOp) UnmarshalJSON(data []byte) error {
	var m struct {
		Op struct {
			New          *ReposCreateOp_NewRepo
			FromGitHubID *int32
		}
	}
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	switch {
	case m.Op.New != nil:
		*r = ReposCreateOp{Op: &ReposCreateOp_New{New: m.Op.New}}
	case m.Op.FromGitHubID != nil:
		*r = ReposCreateOp{Op: &ReposCreateOp_FromGitHubID{FromGitHubID: *m.Op.FromGitHubID}}
	}
	return nil
}
