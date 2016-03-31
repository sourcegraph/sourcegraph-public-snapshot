package spec

import (
	"bytes"
	"regexp"
)

const (
	// RepoPattern is the regexp pattern that matches RepoSpec strings
	// ("repo" or "domain.com/repo" or "domain.com/path/to/repo").
	RepoPattern = `(?P<repo>(?:` + pathComponentNotDelim + `/)*` + pathComponentNotDelim + `)`

	RepoPathDelim         = "-"
	pathComponentNotDelim = `(?:[^@=/` + RepoPathDelim + `]|(?:[^=/@]{2,}))`

	// RevPattern is the regexp pattern that matches a VCS revision
	// and, optionally, a resolved commit ID. The format is "rev" or
	// "rev===commit".
	RevPattern = unresolvedRevPattern + `(?:` + resolvedRevSep + CommitPattern + `)?`

	// unresolvedRevPattern is the regexp pattern that matches a VCS
	// revision specifier (e.g., "master" or "my/branch~1") without
	// the "===" indicating the resolved commit ID. The revision may
	// not contain "=" or "@" to avoid ambiguity.
	unresolvedRevPattern = `(?P<rev>(?:` + pathComponentNotDelim + `/)*` + pathComponentNotDelim + `)`

	// CommitPattern is the regexp pattern that matches absolute
	// (40-character) hexidecimal commit IDs.
	CommitPattern = `(?P<commit>[[:xdigit:]]{40})`

	resolvedRevSep = `===`

	resolvedCommitSep = `===`
)

var (
	repoPattern = regexp.MustCompile("^" + RepoPattern + "$")
	revPattern  = regexp.MustCompile("^" + RevPattern + "$")
)

// ParseRepo parses a RepoSpec string. If spec is invalid, an
// InvalidError is returned.
func ParseRepo(spec string) (repo string, err error) {
	if m := repoPattern.FindStringSubmatch(spec); len(m) > 0 {
		repo = m[0]
		return
	}
	return "", InvalidError{"RepoSpec", spec, nil}
}

// RepoString returns a RepoSpec string. It is the inverse of
// ParseRepo. It does not check the validity of the inputs.
func RepoString(repo string) string { return repo }

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
