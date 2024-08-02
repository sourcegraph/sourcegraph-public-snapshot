package perforce

import (
	"bufio"
	"io"
	"strings"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	p4types "github.com/sourcegraph/sourcegraph/internal/perforce"
)

// PerformDebugScan will scan protections rules from r and log detailed
// information about how each line was parsed.
func PerformDebugScan(logger log.Logger, r io.Reader, depot extsvc.RepoID, ignoreRulesWithHost bool) (*authz.ExternalUserPermissions, error) {
	perms := &authz.ExternalUserPermissions{
		SubRepoPermissions: make(map[extsvc.RepoID]*authz.SubRepoPermissionsWithIPs),
	}

	pr, err := parseP4ProtectsRaw(r)
	if err != nil {
		return perms, err
	}

	scanner := fullRepoPermsScanner(logger, perms, []extsvc.RepoID{depot})
	err = scanProtects(logger, pr, scanner, ignoreRulesWithHost)
	return perms, err
}

func parseP4ProtectsRaw(rc io.Reader) ([]*p4types.Protect, error) {
	protects := make([]*p4types.Protect, 0)

	scanner := bufio.NewScanner(rc)
	for scanner.Scan() {
		line := scanner.Text()

		// Trim whitespace
		line = strings.TrimSpace(line)

		// Skip comments and blank lines
		if strings.HasPrefix(line, "##") || line == "" {
			continue
		}

		// Trim trailing comments
		if i := strings.Index(line, "##"); i > -1 {
			line = line[:i]
		}

		// Split into fields
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}

		parsedLine := p4ProtectLine{
			level:      fields[0],
			entityType: fields[1],
			name:       fields[2],
			match:      fields[4],
		}
		if strings.HasPrefix(parsedLine.match, "-") {
			parsedLine.isExclusion = true                                // is an exclusion
			parsedLine.match = strings.TrimPrefix(parsedLine.match, "-") // trim leading -
		}

		protects = append(protects, &p4types.Protect{
			Level:       parsedLine.level,
			EntityType:  parsedLine.entityType,
			EntityName:  parsedLine.name,
			Host:        fields[3],
			Match:       parsedLine.match,
			IsExclusion: parsedLine.isExclusion,
		})
	}

	return protects, scanner.Err()
}
