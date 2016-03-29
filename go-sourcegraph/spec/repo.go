package spec

import (
	"bytes"
	"regexp"
)

const (
	// RepoPattern is the regexp pattern that matches RepoSpec strings
	// ("repo" or "domain.com/repo" or "domain.com/path/to/repo").
	RepoPattern = `(?P<repo>(?:[^/.@][^/@]*/)*(?:[^/.@][^/@]*))`

	// RepoRevPattern is the regexp pattern that matches RepoRevSpec
	// strings (which encode a repository path, optional revision, and
	// optional commit).
	RepoRevPattern = RepoPattern + `(?:@` + RevPattern + `)?`

	// RevPattern is the regexp pattern that matches a VCS revision
	// and, optionally, a resolved commit ID. The format is "rev" or
	// "rev===commit".
	RevPattern = unresolvedRevPattern + `(?:` + resolvedRevSep + CommitPattern + `)?`

	// unresolvedRevPattern is the regexp pattern that matches a VCS
	// revision specifier (e.g., "master" or "my/branch~1") without
	// the "===" indicating the resolved commit ID. The revision may
	// not contain "=" or "/." to avoid ambiguity.
	unresolvedRevPattern = `(?P<rev>[^/=]+(?:/[^/.=][^/=]*)*)`

	// CommitPattern is the regexp pattern that matches absolute
	// (40-character) hexidecimal commit IDs.
	CommitPattern = `(?P<commit>[[:xdigit:]]{40})`

	resolvedRevSep = `===`

	// PathNoLeadingDotComponentPattern is a pattern that matches any
	// string that doesn't contain "/.".
	PathNoLeadingDotComponentPattern = `(?:[^/]*(?:/` + noDotDotOrSlash + `)*)`

	// noDotDotOrSlash matches a single path component and does not
	// permit "..".
	noDotDotOrSlash = `(?:[^/.]+[^/]*)+`
)

var (
	repoPattern    = regexp.MustCompile("^" + RepoPattern + "$")
	repoRevPattern = regexp.MustCompile("^" + RepoRevPattern + "$")
	revPattern     = regexp.MustCompile("^" + RevPattern + "$")
)

// ParseRepo parses a RepoSpec string. If spec is invalid, an
// InvalidError is returned.
func ParseRepo(spec string) (repo string, err error) {
	if m := repoPattern.FindStringSubmatch(spec); m != nil {
		repo = m[1]
		return
	}
	return "", InvalidError{"RepoSpec", spec, nil}
}

// RepoString returns a RepoSpec string. It is the inverse of
// ParseRepo. It does not check the validity of the inputs.
func RepoString(repo string) string { return repo }

// ParseRepoRev parses a RepoRevSpec string. If spec is invalid, an
// InvalidError is returned.
func ParseRepoRev(spec string) (repo, rev, commitID string, err error) {
	if m := repoRevPattern.FindStringSubmatch(spec); m != nil {
		repo = m[1]
		if len(m) >= 3 {
			rev = m[2]
		}
		if len(m) >= 4 {
			commitID = m[3]
		}
		return
	}
	return "", "", "", InvalidError{"RepoRevSpec", spec, nil}
}

// RepoRevString returns a RepoRevSpec string. It is the inverse of
// ParseRepoRev. It does not check the validity of the inputs.
func RepoRevString(repo, rev, commitID string) string {
	s := repo
	if rrev := ResolvedRevString(rev, commitID); rrev != "" {
		return s + "@" + rrev
	}
	return repo
}

// ParseResolvedRev parses a ResolvedRevSpec string ("rev" or
// "rev===commit").
func ParseResolvedRev(spec string) (rev, commitID string) {
	if m := revPattern.FindStringSubmatch(spec); m != nil {
		rev = m[1]
		if len(m) >= 3 {
			commitID = m[2]
		}
		return
	}
	return spec, ""
}

// ResolvedRevString returns a ResolvedRevSpec string. It is the
// inverse of ParseResolvedRev. It does not check the validity of the
// inputs.
func ResolvedRevString(rev, commitID string) string {
	n := len(rev)
	if commitID != "" {
		n += len(resolvedRevSep) + len(commitID)
	}
	buf := bytes.NewBuffer(make([]byte, 0, n))
	buf.WriteString(rev)
	if commitID != "" {
		buf.Write([]byte(resolvedRevSep))
		buf.WriteString(commitID)
	}
	return buf.String()
}
