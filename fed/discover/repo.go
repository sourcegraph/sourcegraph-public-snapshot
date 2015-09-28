package discover

import (
	"fmt"

	"golang.org/x/net/context"
)

// Repo performs the discovery process for the specified repo (by
// calling QuickRepoFuncs then RepoFuncs).
//
// The returned Info's NewContext method should be used to add the
// configuration necessary for performing the federated operation to
// the current context.
//
// It returns an Info instead of just adding the discovery results to
// ctx for two reasons:
//
//  1. Using an Info interface lets us guarantee that the results will
//     be Stringers (which helps with debugging--it's hard to get a
//     grasp for what some arbitrary Context represents).
//
//  2. It lets the caller use a different Context for discovery
//     vs. the actual API operations they'll perform. For example,
//     they use an HTTP client with a shorter timeout for discovery
//     but use the default HTTP client for subsequent operations.
func Repo(ctx context.Context, repo string) (Info, error) {
	do := func(funcs []RepoFunc) (Info, error) {
		var errs []error
		for _, f := range funcs {
			info, err := f(ctx, repo)
			if IsNotFound(err) {
				errs = append(errs, err.(*NotFoundError).Err)
				continue
			} else if err != nil {
				return nil, err
			}
			return info, nil
		}
		return nil, &NotFoundError{Type: "repo", Input: repo, Err: fmt.Errorf("%v", errs)}
	}

	if info, err := do(QuickRepoFuncs); info != nil || (err != nil && !IsNotFound(err)) {
		return info, err
	}
	return do(RepoFuncs)
}

var (
	// RepoFuncs is a list of functions that perform repo discovery.
	//
	// It should be modified at init time only by packages that need
	// to register prefix discovery functions (e.g., handling GitHub
	// repositories).
	//
	// The ordering of RepoFuncs is undefined and should not be relied
	// on.
	RepoFuncs []RepoFunc

	// QuickRepoFuncs is like RepoFuncs, but with the additional
	// constraint that the funcs should run immediately without
	// requiring network access or performing other expensive
	// operations. QuickRepoFuncs are called before RepoFuncs during
	// discovery.
	//
	// The ordering of QuickRepoFuncs is undefined and should not be
	// relied on. Therefore, funcs in the list should not overlap in
	// the repo paths they handle (e.g., by only handling paths with a
	// certain prefix).
	QuickRepoFuncs []RepoFunc
)

// A RepoFunc performs discovery on the specified repo path. If
// successful (i.e., this discovery func successfully discovered a
// Sourcegraph instance that hosts a specified repo), the returned
// Info encapsulates the config (endpoint URLs, etc.) that should be
// used to perform the operation. Otherwise, a non-nil error is
// returned.
//
// If a NotFoundError is returned, discovery proceeds with the next
// discovery func. Otherwise the process fails immediately with the
// first non-nil error it encounters.
type RepoFunc func(ctx context.Context, repo string) (Info, error)
