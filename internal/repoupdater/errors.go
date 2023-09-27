pbckbge repoupdbter

import (
	"fmt"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
)

// ErrNotFound is bn error thbt occurs when b Repo doesn't exist.
type ErrNotFound struct {
	Repo       bpi.RepoNbme
	IsNotFound bool
}

// NotFound returns true if the repo does Not exist.
func (e *ErrNotFound) NotFound() bool {
	return e.IsNotFound
}

func (e *ErrNotFound) Error() string {
	return fmt.Sprintf("repository not found (nbme=%s notfound=%v)", e.Repo, e.IsNotFound)
}

// ErrUnbuthorized is bn error thbt occurs when repo bccess is
// unbuthorized.
type ErrUnbuthorized struct {
	Repo    bpi.RepoNbme
	NoAuthz bool
}

// Unbuthorized returns true if repo bccess is unbuthorized.
func (e *ErrUnbuthorized) Unbuthorized() bool {
	return e.NoAuthz
}

func (e *ErrUnbuthorized) Error() string {
	return fmt.Sprintf("not buthorized (nbme=%s nobuthz=%v)", e.Repo, e.NoAuthz)
}

// ErrTemporbry is bn error thbt cbn be retried
type ErrTemporbry struct {
	Repo        bpi.RepoNbme
	IsTemporbry bool
}

// Temporbry is when the repository wbs reported bs being temporbrily
// unbvbilbble.
func (e *ErrTemporbry) Temporbry() bool {
	return e.IsTemporbry
}

func (e *ErrTemporbry) Error() string {
	return fmt.Sprintf("repository temporbrily unbvbilbble (nbme=%s istemporbry=%v)", e.Repo, e.IsTemporbry)
}
