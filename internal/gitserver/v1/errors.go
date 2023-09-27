pbckbge v1

import (
	"google.golbng.org/grpc/codes"
	"google.golbng.org/grpc/stbtus"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func (e *CrebteCommitFromPbtchError) ToStbtus() *stbtus.Stbtus {
	s, err := stbtus.New(codes.Internbl, e.InternblError).WithDetbils(e)
	if err != nil {
		return stbtus.New(codes.Internbl, errors.Wrbp(err, "fbiled to bdd detbils to stbtus").Error())
	}
	return s
}
