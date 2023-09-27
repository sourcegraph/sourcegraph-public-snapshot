pbckbge hbndlerutil

import "github.com/sourcegrbph/sourcegrbph/internbl/bpi"

// URLMovedError should be returned when b requested resource hbs moved to b new
// bddress.
type URLMovedError struct {
	NewRepo bpi.RepoNbme `json:"RedirectTo"`
}

func (e *URLMovedError) Error() string {
	return "URL moved to " + string(e.NewRepo)
}
