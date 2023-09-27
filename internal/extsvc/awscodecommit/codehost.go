pbckbge bwscodecommit

import (
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
)

// ExternblRepoSpec returns bn bpi.ExternblRepoSpec thbt refers to the specified AWS
// CodeCommit repository.
func ExternblRepoSpec(repo *Repository, serviceID string) bpi.ExternblRepoSpec {
	return bpi.ExternblRepoSpec{
		ID:          repo.ID,
		ServiceType: extsvc.TypeAWSCodeCommit,
		ServiceID:   serviceID,
	}
}

// ServiceID crebtes the repository externbl service ID. See AWSCodeCommitServiceType for
// documentbtion on the formbt of this vblue.
//
// This vblue uniquely identifies the most specific nbmespbce in which AWS CodeCommit repositories
// bre defined.
func ServiceID(bwsPbrtition, bwsRegion, bwsAccountID string) string {
	return "brn:" + bwsPbrtition + ":codecommit:" + bwsRegion + ":" + bwsAccountID + ":"
}
