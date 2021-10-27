package perforce

import (
	"bufio"
	"context"
	"io"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/gobwas/glob"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
)

// p4ProtectLine is a parsed line from `p4 protects`. See:
//  - https://www.perforce.com/manuals/cmdref/Content/CmdRef/p4_protect.html#Usage_Notes_..364
//  - https://www.perforce.com/manuals/cmdref/Content/CmdRef/p4_protects.html#p4_protects
type p4ProtectLine struct {
	level      string // e.g. read
	entityType string // e.g. user
	name       string // e.g. alice
	match      string // raw match, e.g. //Sourcegraph/, trimmed of leading '-' for exclusion

	// isExclusion is whether the match is an exclusion or inclusion (had a leading '-' or not)
	// which indicates access should be revoked
	isExclusion bool
}

// revokesReadAccess returns true if the line's access level is able to revoke
// read account for a depot prefix.
func (p *p4ProtectLine) revokesReadAccess() bool {
	_, canRevokeReadAccess := map[string]struct{}{
		"list":   {},
		"read":   {},
		"=read":  {},
		"open":   {},
		"write":  {},
		"review": {},
		"owner":  {},
		"admin":  {},
		"super":  {},
	}[p.level]
	return canRevokeReadAccess
}

// grantsReadAccess returns true if the line's access level is able to grant
// read account for a depot prefix.
func (p *p4ProtectLine) grantsReadAccess() bool {
	_, canGrantReadAccess := map[string]struct{}{
		"read":   {},
		"=read":  {},
		"open":   {},
		"=open":  {},
		"write":  {},
		"=write": {},
		"review": {},
		"owner":  {},
		"admin":  {},
		"super":  {},
	}[p.level]
	return canGrantReadAccess
}

// affectsReadAccess returns true if this line changes read access.
func (p *p4ProtectLine) affectsReadAccess() bool {
	return (p.isExclusion && p.revokesReadAccess()) ||
		(!p.isExclusion && p.grantsReadAccess())
}

// Perforce wildcards file match syntax.
//
// See: https://www.perforce.com/manuals/cmdref/Content/CmdRef/filespecs.html
const (
	// Matches all characters except slashes within one directory.
	perforceWildcardMatchAll = "..."
	// Matches all files under the current working directory and all subdirectories.
	// Matches anything, including slashes, and does so across subdirectories.
	perforceWildcardMatchDirectory = "*"
)

// PostgreSQL's SIMILAR TO equivalents for Perforce file match syntaxes.
//
// See: https://www.postgresql.org/docs/12/functions-matching.html#FUNCTIONS-SIMILARTO-REGEXP
var postgresMatchSyntax = map[string]string{
	// Matches anything, including directory slashes.
	perforceWildcardMatchAll: "%",
	// Character class that matches anything except another '/' supported.
	perforceWildcardMatchDirectory: "[^/]+",
}

// convertToPostgresMatch converts supported patterns to PostgreSQL equivalents.
func convertToPostgresMatch(match string) string {
	for perforce, postgres := range postgresMatchSyntax {
		match = strings.ReplaceAll(match, perforce, postgres)
	}
	return match
}

// Glob equivalents for Perforce file match syntaxes.
var globMatchSyntax = map[string]string{
	// Matches any sequence of characters
	perforceWildcardMatchAll: "**",
	// Matches any sequence of non-separator characters
	perforceWildcardMatchDirectory: "*",
}

type globMatch struct {
	glob.Glob
	raw string
}

// convertToGlobMatch converts supported patterns to Glob, and ensures the rest of the
// match does not contain unexpected Glob patterns.
func convertToGlobMatch(match string) (globMatch, error) {
	// Escape all glob syntax first, to ensure nothing unexpected shows up
	match = glob.QuoteMeta(match)

	// Replace glob-escaped Perforce syntax with glob syntax
	for perforce, globSyntax := range globMatchSyntax {
		match = strings.ReplaceAll(match, glob.QuoteMeta(perforce), globSyntax)
	}

	// Allow a tailing '/' on trailing single wildcards
	if strings.HasSuffix(match, "*") && !strings.HasSuffix(match, "**") && !strings.HasSuffix(match, `\*`) {
		match += `{/,}`
	}

	g, err := glob.Compile(match, '/')
	return globMatch{
		Glob: g,
		raw:  match,
	}, errors.Wrap(err, "invalid pattern")
}

// scanProtects is a utility function for processing values from `p4 protects`.
// It handles skipping comments, cleaning whitespace, parsing relevant fields, and
// skipping entries that do not affect read access.
func scanProtects(rc io.Reader, processLine func(p4ProtectLine) error) error {
	scanner := bufio.NewScanner(rc)
	for scanner.Scan() {
		line := scanner.Text()

		// Skip comments
		if strings.HasPrefix(line, "##") {
			continue
		}

		// Trim trailing comments
		i := strings.Index(line, "##")
		if i > -1 {
			line = line[:i]
		}

		// Trim whitespace
		line = strings.TrimSpace(line)

		// Split into fields
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}

		// Parse line
		parsedLine := p4ProtectLine{
			level:      fields[0],
			entityType: fields[1],
			name:       fields[2],
			match:      fields[4],
		}
		if strings.HasPrefix(parsedLine.match, "-") {
			parsedLine.isExclusion = true           // is an exclusion
			parsedLine.match = parsedLine.match[1:] // trim leading -
		}

		// We only care about read access. If the permission doesn't change read access,
		// then we exit early.
		if !parsedLine.affectsReadAccess() {
			continue
		}

		// Do stuff to line
		if err := processLine(parsedLine); err != nil {
			return err
		}
	}
	return nil
}

// scanRepoIncludesExcludes converts `p4 protects` to Postgres SIMILAR TO-compatible
// entries for including and excluding depots as "repositories".
func repoIncludesExcludesScanner(perms *authz.ExternalUserPermissions) func(p4ProtectLine) error {
	return func(line p4ProtectLine) error {
		// We drop trailing '...' so that we can check for prefixes (see below).
		line.match = strings.TrimRight(line.match, ".")

		// NOTE: Manipulations made to `depotContains` will affect the behaviour of
		// `(*RepoStore).ListRepoNames` - make sure to test new changes there as well.
		depotContains := convertToPostgresMatch(line.match)

		if !line.isExclusion {
			perms.IncludeContains = append(perms.IncludeContains, extsvc.RepoID(depotContains))
			return nil
		}

		if strings.Contains(line.match, perforceWildcardMatchAll) ||
			strings.Contains(line.match, perforceWildcardMatchDirectory) {
			// Always include wildcard matches, because we don't know what they might
			// be matching on.
			perms.ExcludeContains = append(perms.ExcludeContains, extsvc.RepoID(depotContains))
			return nil
		}

		// Otherwise, only include an exclude if a corresponding include exists.
		for i, prefix := range perms.IncludeContains {
			if !strings.HasPrefix(depotContains, string(prefix)) {
				continue
			}

			// Perforce ACLs can have conflict rules and the later one wins. So if there is
			// an exact match for an include prefix, we take it out.
			if depotContains == string(prefix) {
				perms.IncludeContains = append(perms.IncludeContains[:i], perms.IncludeContains[i+1:]...)
				break
			}

			perms.ExcludeContains = append(perms.ExcludeContains, extsvc.RepoID(depotContains))
			break
		}

		return nil
	}
}

// fullRepoPermsScanner converts `p4 protects` to a 1:1 implementation of Sourcegraph
// authorization, including sub-repo perms.
func fullRepoPermsScanner(perms *authz.ExternalUserPermissions, configuredDepots []extsvc.RepoID) func(p4ProtectLine) error {
	relevantDepots := func(m globMatch) (depots []extsvc.RepoID) {
		for _, depot := range configuredDepots {
			if m.Match(string(depot)) {
				depots = append(depots, depot)
			}
		}
		return
	}

	seenRepos := make(map[extsvc.RepoID]struct{})
	appendRepos := func(ids ...extsvc.RepoID) {
		for _, id := range ids {
			if _, seen := seenRepos[id]; !seen {
				perms.Exacts = append(perms.Exacts, id)
				seenRepos[id] = struct{}{}
			}
		}
	}

	return func(line p4ProtectLine) error {
		match, err := convertToGlobMatch(line.match)
		if err != nil {
			return err
		}

		if !line.isExclusion {
			// Access is granted to *some* part of the depot, so user has access to the
			// root repo.
			depots := relevantDepots(match)
			appendRepos(depots...)

			// Grant access to specified paths
			for _, depot := range depots {
				perms := perms.SubRepoPermissions[depot]
				perms.PathIncludes = append(perms.PathIncludes, match.raw)
			}

			return nil
		}

		// TODO exclusio ncases

		return nil
	}
}

// allUsersScanner converts `p4 protects` to a map of users within the protection rules.
func allUsersScanner(ctx context.Context, p *Provider, users map[string]struct{}) func(p4ProtectLine) error {
	return func(line p4ProtectLine) error {
		if line.isExclusion {
			switch line.entityType {
			case "user":
				if line.name == "*" {
					for u := range users {
						delete(users, u)
					}
				} else {
					delete(users, line.name)
				}
			case "group":
				members, err := p.getGroupMembers(ctx, line.name)
				if err != nil {
					return errors.Wrapf(err, "list members of group %q", line.name)
				}
				for _, member := range members {
					delete(users, member)
				}

			default:
				log15.Warn("authz.perforce.Provider.FetchRepoPerms.unrecognizedType", "type", line.entityType)
			}

			return nil
		}

		switch line.entityType {
		case "user":
			if line.name == "*" {
				all, err := p.getAllUsers(ctx)
				if err != nil {
					return errors.Wrap(err, "list all users")
				}
				for _, user := range all {
					users[user] = struct{}{}
				}
			} else {
				users[line.name] = struct{}{}
			}
		case "group":
			members, err := p.getGroupMembers(ctx, line.name)
			if err != nil {
				return errors.Wrapf(err, "list members of group %q", line.name)
			}
			for _, member := range members {
				users[member] = struct{}{}
			}

		default:
			log15.Warn("authz.perforce.Provider.FetchRepoPerms.unrecognizedType", "type", line.entityType)
		}

		return nil
	}
}
