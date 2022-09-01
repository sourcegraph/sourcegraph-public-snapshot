package resolvers

import (
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
)

type RetentionPolicyMatchCandidate struct {
	*dbstore.ConfigurationPolicy
	Matched           bool
	ProtectingCommits []string
}
