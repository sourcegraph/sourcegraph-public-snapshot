package store

import (
	"github.com/sourcegraph/sourcegraph/internal/codeintel/policies/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

func scanPolicy(s dbutil.Scanner) (policy shared.Policy, err error) {
	return policy, s.Scan(
		&policy.ID,
	)
}

var scanPolicies = basestore.NewSliceScanner(scanPolicy)
