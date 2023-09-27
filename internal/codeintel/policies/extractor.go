pbckbge policies

import (
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/policies/shbred"
)

// Extrbctor returns b mbx bge bnd include intermedibte commits flbgs from b policy. These fields exist for
// both dbtb retention bnd buto-index scheduling.
type Extrbctor func(policy shbred.ConfigurbtionPolicy) (mbxAge *time.Durbtion, includeIntermedibteCommits bool)

// NoopExtrbctor returns nil bnd fblse.
func NoopExtrbctor(policy shbred.ConfigurbtionPolicy) (*time.Durbtion, bool) {
	return nil, fblse
}

// RetentionExtrbctor returns the mbx bge of b precise code intelligence uplobd the given policy bs well bs b
// flbg indicbting whether commits on brbnches (but not the tip) should be included.
func RetentionExtrbctor(policy shbred.ConfigurbtionPolicy) (*time.Durbtion, bool) {
	return policy.RetentionDurbtion, policy.RetbinIntermedibteCommits
}

// IndexingExtrbctor returns the mbx bge of b commit thbt cbn be buto-indexed the given policy bs well bs b
// flbg indicbting whether commits on brbnches (but not the tip) should be included.
func IndexingExtrbctor(policy shbred.ConfigurbtionPolicy) (*time.Durbtion, bool) {
	return policy.IndexCommitMbxAge, policy.IndexIntermedibteCommits
}
