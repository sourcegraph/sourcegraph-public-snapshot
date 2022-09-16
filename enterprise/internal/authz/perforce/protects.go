package perforce

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/gobwas/glob"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// p4ProtectLine is a parsed line from `p4 protects`. See:
//   - https://www.perforce.com/manuals/cmdref/Content/CmdRef/p4_protect.html#Usage_Notes_..364
//   - https://www.perforce.com/manuals/cmdref/Content/CmdRef/p4_protects.html#p4_protects
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

// Perforce wildcards file match syntax, specifically Helix server wildcards. This is the format
// we expect to get back from p4 protects.
//
// See: https://www.perforce.com/manuals/p4guide/Content/P4Guide/syntax.syntax.wildcards.html
const (
	// Matches anything including slashes. Matches recursively (everything in and below the specified directory).
	perforceWildcardMatchAll = "..."
	// Matches anything except slashes. Matches only within a single directory. Case sensitivity depends on your platform.
	perforceWildcardMatchDirectory = "*"
)

func hasPerforceWildcard(match string) bool {
	return strings.Contains(match, perforceWildcardMatchAll) ||
		strings.Contains(match, perforceWildcardMatchDirectory)
}

// PostgreSQL's SIMILAR TO equivalents for Perforce file match syntaxes.
//
// See: https://www.postgresql.org/docs/12/functions-matching.html#FUNCTIONS-SIMILARTO-REGEXP
var postgresMatchSyntax = strings.NewReplacer(
	// Matches anything, including directory slashes.
	perforceWildcardMatchAll, "%",
	// Character class that matches anything except another '/' supported.
	perforceWildcardMatchDirectory, "[^/]+",
)

// convertToPostgresMatch converts supported patterns to PostgreSQL equivalents.
func convertToPostgresMatch(match string) string {
	return postgresMatchSyntax.Replace(match)
}

// Glob syntax equivalents for _glob-escaped_ Perforce file match syntaxes.
//
// See: authz.SubRepoPermissions
var globMatchSyntax = strings.NewReplacer(
	// Matches any sequence of characters
	glob.QuoteMeta(perforceWildcardMatchAll), "**",
	// Matches any sequence of non-separator characters
	glob.QuoteMeta(perforceWildcardMatchDirectory), "*",
)

type globMatch struct {
	glob.Glob
	pattern  string
	original string
}

// convertToGlobMatch converts supported patterns to Glob, and ensures the rest of the
// match does not contain unexpected Glob patterns.
func convertToGlobMatch(match string) (globMatch, error) {
	original := match

	// Escape all glob syntax first, to ensure nothing unexpected shows up
	match = glob.QuoteMeta(match)

	// Replace glob-escaped Perforce syntax with glob syntax
	match = globMatchSyntax.Replace(match)

	// Allow a trailing '/' on trailing single wildcards
	if strings.HasSuffix(match, "*") && !strings.HasSuffix(match, "**") && !strings.HasSuffix(match, `\*`) {
		match += `{/,}`
	}

	g, err := glob.Compile(match, '/')
	return globMatch{
		Glob:     g,
		pattern:  match,
		original: original,
	}, errors.Wrap(err, "invalid pattern")
}

// matchesAgainstDepot checks if the given match affects the given depot.
func matchesAgainstDepot(match globMatch, depot string) bool {
	if match.Match(depot) {
		return true
	}

	// If the subpath includes a wildcard:
	// - depot: "//depot/main/"
	// - match: "//depot/.../file" or "//*/main/..."
	// Then we want to check if it could match this match
	if !hasPerforceWildcard(match.original) {
		return false
	}
	parts := strings.Split(match.original, perforceWildcardMatchAll)
	if len(parts) > 0 {
		// Check full prefix
		prefixMatch, err := convertToGlobMatch(parts[0] + perforceWildcardMatchAll)
		if err != nil {
			return false
		}
		if prefixMatch.Match(depot) {
			return true
		}
	}
	// Check each prefix part for perforceWildcardMatchDirectory.
	// We already tried the full match, so start at len(parts)-1, and don't traverse all
	// the way down to root unless there's a wildcard there.
	parts = strings.Split(match.original, "/")
	for i := len(parts) - 1; i > 2 || strings.Contains(parts[i], perforceWildcardMatchDirectory); i-- {
		// Depots should always be suffixed with '/'
		prefixMatch, err := convertToGlobMatch(strings.Join(parts[:i], "/") + "/")
		if err != nil {
			return false
		}
		if prefixMatch.Match(depot) {
			return true
		}
	}

	return false
}

// PerformDebugScan will scan protections rules from r and log detailed
// information about how each line was parsed.
func PerformDebugScan(logger log.Logger, r io.Reader, depot extsvc.RepoID) (*authz.ExternalUserPermissions, error) {
	perms := &authz.ExternalUserPermissions{
		SubRepoPermissions: make(map[extsvc.RepoID]*authz.SubRepoPermissions),
	}
	scanner := fullRepoPermsScanner(logger, perms, []extsvc.RepoID{depot})
	err := scanProtects(logger, r, scanner)
	return perms, err
}

// protectsScanner provides callbacks for scanning the output of `p4 protects`.
type protectsScanner struct {
	// Called on the parsed contents of each `p4 protects` line.
	processLine func(p4ProtectLine) error
	// Called upon completion of processing of all lines.
	finalize func() error
}

// scanProtects is a utility function for processing values from `p4 protects`.
// It handles skipping comments, cleaning whitespace, parsing relevant fields, and
// skipping entries that do not affect read access.
func scanProtects(logger log.Logger, rc io.Reader, s *protectsScanner) error {
	logger = logger.Scoped("scanProtects", "")
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

		logger.Debug("Scanning protects line", log.String("line", line))

		// Split into fields
		fields := strings.Fields(line)
		if len(fields) < 5 {
			logger.Debug("Line has less than 5 fields, discarding")
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
			logger.Debug("Line does not affect read access, discarding")
			continue
		}

		// Do stuff to line
		if err := s.processLine(parsedLine); err != nil {
			return err
		}
	}
	var finalizeErr error
	if s.finalize != nil {
		finalizeErr = s.finalize()
	}
	scanErr := scanner.Err()
	return errors.CombineErrors(scanErr, finalizeErr)
}

// scanRepoIncludesExcludes converts `p4 protects` to Postgres SIMILAR TO-compatible
// entries for including and excluding depots as "repositories".
func repoIncludesExcludesScanner(perms *authz.ExternalUserPermissions) *protectsScanner {
	return &protectsScanner{
		processLine: func(line p4ProtectLine) error {
			// We drop trailing '...' so that we can check for prefixes (see below).
			line.match = strings.TrimRight(line.match, ".")

			// NOTE: Manipulations made to `depotContains` will affect the behaviour of
			// `(*RepoStore).ListMinimalRepos` - make sure to test new changes there as well.
			depotContains := convertToPostgresMatch(line.match)

			if !line.isExclusion {
				perms.IncludeContains = append(perms.IncludeContains, extsvc.RepoID(depotContains))
				return nil
			}

			if hasPerforceWildcard(line.match) {
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
		},
		finalize: func() error {
			// Treat all Contains paths as prefixes.
			for i, include := range perms.IncludeContains {
				perms.IncludeContains[i] = extsvc.RepoID(convertToPostgresMatch(string(include) + perforceWildcardMatchAll))
			}
			for i, exclude := range perms.ExcludeContains {
				perms.ExcludeContains[i] = extsvc.RepoID(convertToPostgresMatch(string(exclude) + perforceWildcardMatchAll))
			}
			return nil
		},
	}
}

// fullRepoPermsScanner converts `p4 protects` to a 1:1 implementation of Sourcegraph
// authorization, including sub-repo perms and exact depot-as-repo matches.
func fullRepoPermsScanner(logger log.Logger, perms *authz.ExternalUserPermissions, configuredDepots []extsvc.RepoID) *protectsScanner {
	logger = logger.Scoped("fullRepoPermsScanner", "")
	// Get glob equivalents of all depots
	var configuredDepotMatches []globMatch
	for _, depot := range configuredDepots {
		// treat depots as wildcards
		m, err := convertToGlobMatch(string(depot) + "**")
		if err != nil {
			logger.Error("unexpected failure to convert depot to pattern - using a no-op pattern",
				log.String("depot", string(depot)),
				log.Error(err))
			continue
		}
		logger.Debug("Converted depot to glob", log.String("depot", string(depot)), log.String("glob", m.pattern))
		// preserve original name by overriding the wildcard version of the original text
		m.original = string(depot)
		configuredDepotMatches = append(configuredDepotMatches, m)
	}

	// relevantDepots determines the set of configured depots relevant to the given globMatch
	relevantDepots := func(m globMatch) (depots []extsvc.RepoID) {
		for i, depot := range configuredDepotMatches {
			if depot.Match(m.original) || matchesAgainstDepot(m, depot.original) {
				depots = append(depots, configuredDepots[i])
			}
		}
		return
	}

	// Helper function for retrieving an existing SubRepoPermissions or instantiating one
	getSubRepoPerms := func(repo extsvc.RepoID) *authz.SubRepoPermissions {
		if _, ok := perms.SubRepoPermissions[repo]; !ok {
			perms.SubRepoPermissions[repo] = &authz.SubRepoPermissions{}
		}
		return perms.SubRepoPermissions[repo]
	}

	// Store seen patterns for reference and matching against conflict rules
	patternsToGlob := make(map[string]globMatch)

	return &protectsScanner{
		processLine: func(line p4ProtectLine) error {
			match, err := convertToGlobMatch(line.match)
			if err != nil {
				return err
			}
			patternsToGlob[match.pattern] = match

			// Depots that this match pertains to
			depots := relevantDepots(match)

			depotStrings := make([]string, len(depots))
			for i := range depots {
				depotStrings[i] = string(depots[i])
			}
			logger.Debug("Relevant depots", log.Strings("depots", depotStrings))

			// Handle inclusions
			if !line.isExclusion {
				// Grant access to specified paths
				for _, depot := range depots {
					srp := getSubRepoPerms(depot)
					newIncludes := convertRulesForWildcardDepotMatch(match, depot, patternsToGlob)
					srp.PathIncludes = append(srp.PathIncludes, newIncludes...)
					logger.Debug("Adding include rules", log.Strings("rules", newIncludes))

					var i int
					for _, exclude := range srp.PathExcludes {
						// Perforce ACLs can have conflicting rules and the later one wins, so
						// if we get a match here we drop the existing rule.
						originalExclude, exists := patternsToGlob[exclude]
						if !exists {
							i++
							continue
						}
						checkWithDepotAdded := !strings.HasPrefix(originalExclude.pattern, "//") && match.Match(string(depot)+originalExclude.pattern)
						if originalExclude.Match(match.original) || checkWithDepotAdded {
							logger.Debug("Removing conflicting exclude rule", log.String("rule", originalExclude.pattern))
							srp.PathExcludes = append(srp.PathExcludes[:i], srp.PathExcludes[i+1:]...)
						} else {
							i++
						}
					}
				}

				return nil
			}

			for _, depot := range depots {
				srp := getSubRepoPerms(depot)

				// Special case: exclude entire depot
				if strings.TrimPrefix(match.original, string(depot)) == perforceWildcardMatchAll {
					logger.Debug("Exclude entire depot, removing all include rules")
					srp.PathIncludes = nil
				}

				newExcludes := convertRulesForWildcardDepotMatch(match, depot, patternsToGlob)
				srp.PathExcludes = append(srp.PathExcludes, newExcludes...)
				logger.Debug("Adding exclude rules", log.Strings("rules", newExcludes))

				var i int
				for _, include := range srp.PathIncludes {
					// Perforce ACLs can have conflicting rules and the later one wins, so
					// if we get a match here we drop the existing rule.
					originalInclude, exists := patternsToGlob[include]
					if !exists {
						i++
						continue
					}
					checkWithDepotAdded := !strings.HasPrefix(originalInclude.pattern, "//") && match.Match(string(depot)+originalInclude.pattern)
					if match.Match(originalInclude.original) || checkWithDepotAdded {
						logger.Debug("Removing conflicting include rule", log.String("rule", originalInclude.pattern))
						srp.PathIncludes = append(srp.PathIncludes[:i], srp.PathIncludes[i+1:]...)
					} else {
						i++
					}
				}
			}

			return nil
		},
		finalize: func() error {
			// iterate over configuredDepots to be deterministic
			for _, depot := range configuredDepots {
				srp, exists := perms.SubRepoPermissions[depot]
				if !exists {
					continue
				} else if len(srp.PathIncludes) == 0 {
					// Depots with no inclusions can just be dropped
					delete(perms.SubRepoPermissions, depot)
					continue
				}

				// Rules should not include the depot name. We want them to be relative so that
				// we can match even if repo name transformations have occurred, for example a
				// repositoryPathPattern has been used. We also need to remove any `//` prefixes
				// which are included in all Helix server rules.
				depotString := string(depot)
				for i := range srp.PathIncludes {
					srp.PathIncludes[i] = strings.TrimPrefix(srp.PathIncludes[i], depotString)
					srp.PathIncludes[i] = strings.TrimPrefix(srp.PathIncludes[i], "//")
				}
				for i := range srp.PathExcludes {
					srp.PathExcludes[i] = strings.TrimPrefix(srp.PathExcludes[i], depotString)
					srp.PathExcludes[i] = strings.TrimPrefix(srp.PathExcludes[i], "//")
				}

				// Add to repos users can access
				perms.Exacts = append(perms.Exacts, depot)
			}
			return nil
		},
	}
}

func convertRulesForWildcardDepotMatch(match globMatch, depot extsvc.RepoID, patternsToGlob map[string]globMatch) []string {
	logger := log.Scoped("convertRulesForWildcardDepotMatch", "")
	if !strings.Contains(match.pattern, "**") && !strings.Contains(match.pattern, "*") {
		return []string{match.pattern}
	}
	trimmedRule := strings.TrimPrefix(match.pattern, "//")
	trimmedDepot := strings.TrimSuffix(strings.TrimPrefix(string(depot), "//"), "/")
	parts := strings.Split(trimmedRule, "/")
	newRules := make([]string, 0, len(parts))
	depotOnlyMatchesDoubleWildcard := true
	for i := range parts {
		maybeDepotMatch := strings.Join(parts[:i+1], "/")
		maybePathRule := strings.Join(parts[i+1:], "/")
		depotMatchGlob, err := glob.Compile(maybeDepotMatch, '/')
		if err != nil {
			logger.Warn(fmt.Sprintf("error compiling %s to glob: %v", maybeDepotMatch, err))
			continue
		}
		if depotMatchGlob.Match(trimmedDepot) {
			// special case: depot match ends with **
			if strings.HasSuffix(maybeDepotMatch, "**") {
				if maybePathRule == "" {
					maybePathRule = "**"
				} else {
					maybePathRule = fmt.Sprintf("**/%s", maybePathRule)
				}
			}
			if maybeDepotMatch != "**" {
				depotOnlyMatchesDoubleWildcard = false
				newGlobMatch, err := convertToGlobMatch(maybePathRule)
				if err != nil {
					logger.Warn(fmt.Sprintf("error converting to glob match: %s\n", err))
				}
				patternsToGlob[newGlobMatch.pattern] = newGlobMatch
			}
			newRules = append(newRules, maybePathRule)
		}
	}
	if depotOnlyMatchesDoubleWildcard {
		// in this case, the original rule will work fine, so no need to convert.
		return []string{match.pattern}
	}
	return newRules
}

// allUsersScanner converts `p4 protects` to a map of users within the protection rules.
func allUsersScanner(ctx context.Context, p *Provider, users map[string]struct{}) *protectsScanner {
	logger := log.Scoped("allUsersScanner", "")
	return &protectsScanner{
		processLine: func(line p4ProtectLine) error {
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
					if err := p.excludeGroupMembers(ctx, line.name, users); err != nil {
						return err
					}
				default:
					logger.Warn("authz.perforce.Provider.FetchRepoPerms.unrecognizedType", log.String("type", line.entityType))
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
				if err := p.includeGroupMembers(ctx, line.name, users); err != nil {
					return err
				}
			default:
				logger.Warn("authz.perforce.Provider.FetchRepoPerms.unrecognizedType", log.String("type", line.entityType))
			}

			return nil
		},
	}
}
