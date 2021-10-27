package perforce

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/inconshreveable/log15"

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
			log15.Warn("scanProtects",
				"line", line, "error", err)
			return err
		}
	}
	return nil
}

// scanRepoIncludesExcludes converts `p4 protects` to Postgres SIMILAR TO-compatible
// entries for including and excluding repositories.
func scanRepoIncludesExcludes(rc io.ReadCloser) (includeContains []extsvc.RepoID, excludeContains []extsvc.RepoID, err error) {
	const (
		wildcardMatchAll       = "%"     // for Perforce '...'
		wildcardMatchDirectory = "[^/]+" // for Perforce '*'
	)

	err = scanProtects(rc, func(line p4ProtectLine) error {
		// NOTE: Manipulations made to `depotContains` will affect the behaviour of
		// `(*RepoStore).ListRepoNames` - make sure to test new changes there as well.
		depotContains := line.match

		// '...' matches all files under the current working directory and all subdirectories.
		// Matches anything, including slashes, and does so across subdirectories.
		// Replace with '%' for PostgreSQL's LIKE and SIMILAR TO.
		//
		// At first, we drop trailing '...' so that we can check for prefixes (see below).
		// We assume all paths are prefixes, so add 'wildcardMatchAll' to all contains
		// later on.
		depotContains = strings.TrimRight(depotContains, ".")
		depotContains = strings.ReplaceAll(depotContains, "...", wildcardMatchAll)

		// '*' matches all characters except slashes within one directory.
		// Replace with character class that matches anything except another '/' supported
		// by PostgreSQL's SIMILAR TO.
		depotContains = strings.ReplaceAll(depotContains, "*", wildcardMatchDirectory)

		if !line.isExclusion {
			includeContains = append(includeContains, extsvc.RepoID(depotContains))
			return nil
		}

		if strings.Contains(depotContains, wildcardMatchAll) ||
			strings.Contains(depotContains, wildcardMatchDirectory) {
			// Always include wildcard matches, because we don't know what they might
			// be matching on.
			excludeContains = append(excludeContains, extsvc.RepoID(depotContains))
		} else {
			// Otherwise, only include an exclude if a corresponding include exists.
			for i, prefix := range includeContains {
				if !strings.HasPrefix(depotContains, string(prefix)) {
					continue
				}

				// Perforce ACLs can have conflict rules and the later one wins. So if there is
				// an exact match for an include prefix, we take it out.
				if depotContains == string(prefix) {
					includeContains = append(includeContains[:i], includeContains[i+1:]...)
					break
				}

				excludeContains = append(excludeContains, extsvc.RepoID(depotContains))
				break
			}
		}

		return nil
	})

	// Treat all paths as prefixes.
	for i, include := range includeContains {
		includeContains[i] = extsvc.RepoID(string(include) + wildcardMatchAll)
	}
	for i, exclude := range excludeContains {
		excludeContains[i] = extsvc.RepoID(string(exclude) + wildcardMatchAll)
	}

	return
}

// scanAllUsers converts `p4 protects` to a map of users within the protection rules.
func scanAllUsers(ctx context.Context, p *Provider, rc io.ReadCloser) (map[string]struct{}, error) {
	users := make(map[string]struct{})

	err := scanProtects(rc, func(line p4ProtectLine) error {
		if line.isExclusion {
			switch line.entityType {
			case "user":
				if line.name == "*" {
					users = make(map[string]struct{})
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
	})
	if err != nil {
		return nil, fmt.Errorf("scanAllUsers: %w", err)
	}

	return users, err
}
