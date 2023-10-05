package perforce

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/byteutils"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	p4types "github.com/sourcegraph/sourcegraph/internal/perforce"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// PerformDebugScan will scan protections rules from r and log detailed
// information about how each line was parsed.
func PerformDebugScan(logger log.Logger, r io.Reader, depot extsvc.RepoID, ignoreRulesWithHost bool) (*authz.ExternalUserPermissions, error) {
	perms := &authz.ExternalUserPermissions{
		SubRepoPermissions: make(map[extsvc.RepoID]*authz.SubRepoPermissions),
	}
	buf, err := io.ReadAll(r)
	if err != nil {
		return perms, err
	}
	pr, err := parseP4Protects(buf)
	if err != nil {
		return perms, err
	}
	scanner := fullRepoPermsScanner(logger, perms, []extsvc.RepoID{depot})
	err = scanProtects(logger, pr, scanner, ignoreRulesWithHost)
	return perms, err
}

type perforceJSONProtect struct {
	DepotFile string  `json:"depotFile"`
	Host      string  `json:"host"`
	Line      string  `json:"line"`
	Perm      string  `json:"perm"`
	IsGroup   *string `json:"isgroup,omitempty"`
	Unmap     *string `json:"unmap,omitempty"`
	User      string  `json:"user"`
}

func parseP4Protects(out []byte) ([]*p4types.Protect, error) {
	protects := make([]*p4types.Protect, 0)

	lr := byteutils.NewLineReader(out)
	for lr.Scan() {
		line := lr.Line()

		// Trim whitespace
		line = bytes.TrimSpace(line)

		var parsedLine perforceJSONProtect
		if err := json.Unmarshal(line, &parsedLine); err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal protect line")
		}

		entityType := "user"
		if parsedLine.IsGroup != nil {
			entityType = "group"
		}

		protects = append(protects, &p4types.Protect{
			Host:        parsedLine.Host,
			EntityType:  entityType,
			EntityName:  parsedLine.User,
			Match:       parsedLine.DepotFile,
			IsExclusion: parsedLine.Unmap != nil,
			Level:       parsedLine.Perm,
		})
	}

	return protects, nil
}
