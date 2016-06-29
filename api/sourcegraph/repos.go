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

// IsAbs returns whether s.CommitID is a valid absolute commit ID (40
// characters and hexadecimal). It is not a guarantee that s.CommitID
// refers to an existing commit ID in the repository, or that it is
// even a commit ID (it could be a Git oid referring to another
// object, such as a blob).
func (s RepoRevSpec) IsAbs() bool {
	if len(s.CommitID) != 40 {
		return false
	}
	for _, c := range s.CommitID {
		if c < '0' || c > 'f' {
			return false
		}
	}
	return true
}

func (r *ReposCreateOp) UnmarshalJSON(data []byte) error {
	var m struct {
		Op struct {
			New          *ReposCreateOp_NewRepo
			FromGitHubID *int32
			Origin       *Origin
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
	case m.Op.Origin != nil:
		*r = ReposCreateOp{Op: &ReposCreateOp_Origin{Origin: m.Op.Origin}}
	}
	return nil
}
