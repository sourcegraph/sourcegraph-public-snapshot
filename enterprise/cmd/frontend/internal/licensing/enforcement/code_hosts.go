package enforcement

import (
	"context"
	"fmt"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/db"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
)

// CodeHostsStore is implemented by any type that can act as a
// repository for code hosts (e.g. GitHub, GitLab).
type CodeHostsStore interface {
	Count(context.Context, db.CodeHostsListOptions) (int, error)
}

// NewPreCreateCodeHostHook enforces any per-tier validations prior to
// creating a new code host.
func NewPreCreateCodeHostHook(codeHosts CodeHostsStore) func(ctx context.Context) error {
	if !licensing.EnforceTiers {
		return nil
	}

	return func(ctx context.Context) error {
		// First we need to determine how many code hosts are permissible.
		var maxCodeHostCount int
		if mockGetCodeHostsLimit != nil {
			maxCodeHostCount = mockGetCodeHostsLimit(ctx)
		}

		// TODO(flying-robot, unknwon) - wrap licensing Info type with per-tier
		// limits that we can reference here. The count check below will pass
		// the validation right now since maxCodeHostCount = 0.

		// Next we'll grab the current count of code hosts.
		codeHostCount, err := codeHosts.Count(ctx, db.CodeHostsListOptions{})
		if err != nil {
			return err
		}

		// If we have none configured or we're under the limit, we can pass the
		// validation. Otherwise an error will be returned. Note that we consider
		// a maximum of 0 to be "unlimited", which is consistent with other checks.
		if codeHostCount == 0 || maxCodeHostCount == 0 || codeHostCount < maxCodeHostCount {
			return nil
		}
		return errcode.NewPresentationError(
			fmt.Sprintf(
				"Unable to create code host: the current plan cannot exceed %d code hosts (this instance now has %d hosts). Contact Sourcegraph to learn more at https://about.sourcegraph.com/contact/sales.",
				maxCodeHostCount,
				codeHostCount,
			),
		)
	}
}

// These mocks are used for test purposes.
var (
	mockGetCodeHostsLimit func(_ context.Context) int
)
