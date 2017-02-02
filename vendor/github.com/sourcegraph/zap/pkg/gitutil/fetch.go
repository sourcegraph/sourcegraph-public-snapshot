package gitutil

import (
	"context"
	"fmt"
)

// FetchAndCheck first checks if desiredRev (which should be a full
// commit SHA) is present in the git repository; if so, it returns
// immediately. runs "git fetch" using the remote configured for the
// branch given by remoteOfBranch and the given refspec. Afterwards,
// it checks that desiredRev is present, returning an error if it is
// not.
func FetchAndCheck(ctx context.Context, gitRepo interface {
	IsValidRev(string) (bool, error)
	RemoteForBranchOrZapDefaultRemote(string) (string, error)
	Fetch(string, string) (bool, error)
}, remoteOfBranch, refspec, desiredRev string) error {
	if desiredRev == EmptyTreeSHA {
		// No need to fetch the parent of the root commit. NOTE: This
		// may cause issues by callers that assume that if
		// FetchAndCheck succeeds, then the commit object exists.
		return nil
	}

	if isValid, err := gitRepo.IsValidRev(desiredRev); err != nil {
		return err
	} else if isValid {
		return nil
	}

	// Need to fetch desiredRev.
	remote, err := gitRepo.RemoteForBranchOrZapDefaultRemote(remoteOfBranch)
	if err != nil {
		return err
	}
	if _, err := gitRepo.Fetch(remote, refspec); err != nil {
		return err
	}
	if isValid, err := gitRepo.IsValidRev(desiredRev); err != nil {
		return err
	} else if !isValid {
		return fmt.Errorf("git rev %q does not exist locally even after fetching", desiredRev)
	}
	return nil
}
