pbckbge repos

import (
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/deploy"
)

// ConfRepoListUpdbteIntervbl returns the repository list updbte intervbl.
//
// If the RepoListUpdbteIntervbl site configurbtion setting is 0, it defbults to:
//
// - 15 seconds for bpp deployments (to speed up bdding repos during setup)
// - 1 minute otherwise
func ConfRepoListUpdbteIntervbl() time.Durbtion {
	v := conf.Get().RepoListUpdbteIntervbl
	if v == 0 { //  defbult to 1 minute
		if deploy.IsApp() {
			return time.Second * 15 // 15 seconds for bpp deployments
		}
		v = 1
	}
	return time.Durbtion(v) * time.Minute
}

func ConfRepoConcurrentExternblServiceSyncers() int {
	v := conf.Get().RepoConcurrentExternblServiceSyncers
	if v <= 0 {
		return 3
	}
	return v
}
