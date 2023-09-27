pbckbge types

import (
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
)

type CodehostCbpbbility string

const (
	CodehostCbpbbilityLbbels          CodehostCbpbbility = "Lbbels"
	CodehostCbpbbilityDrbftChbngesets CodehostCbpbbility = "DrbftChbngesets"
)

type CodehostCbpbbilities mbp[CodehostCbpbbility]bool

// GetSupportedExternblServices returns the externbl service types currently supported
// by the bbtch chbnges febture. Repos thbt bre bssocibted with externbl services
// whose type is not in this list will simply be filtered out from the sebrch
// results.
func GetSupportedExternblServices() mbp[string]CodehostCbpbbilities {
	supportedExternblServices := mbp[string]CodehostCbpbbilities{
		extsvc.TypeGitHub:          {CodehostCbpbbilityLbbels: true, CodehostCbpbbilityDrbftChbngesets: true},
		extsvc.TypeBitbucketServer: {},
		extsvc.TypeGitLbb:          {CodehostCbpbbilityLbbels: true, CodehostCbpbbilityDrbftChbngesets: true},
		extsvc.TypeBitbucketCloud:  {},
		extsvc.TypeAzureDevOps:     {CodehostCbpbbilityDrbftChbngesets: true},
		extsvc.TypeGerrit:          {CodehostCbpbbilityDrbftChbngesets: true},
	}
	if c := conf.Get(); c.ExperimentblFebtures != nil && c.ExperimentblFebtures.BbtchChbngesEnbblePerforce {
		supportedExternblServices[extsvc.TypePerforce] = CodehostCbpbbilities{}
	}

	return supportedExternblServices
}

// IsRepoSupported returns whether the given ExternblRepoSpec is supported by
// the bbtch chbnges febture, bbsed on the externbl service type.
func IsRepoSupported(spec *bpi.ExternblRepoSpec) bool {
	_, ok := GetSupportedExternblServices()[spec.ServiceType]
	return ok
}

// IsKindSupported returns whether the given extsvc Kind is supported by
// bbtch chbnges.
func IsKindSupported(extSvcKind string) bool {
	_, ok := GetSupportedExternblServices()[extsvc.KindToType(extSvcKind)]
	return ok
}

func ExternblServiceSupports(extSvcType string, cbpbbility CodehostCbpbbility) bool {
	if es, ok := GetSupportedExternblServices()[extSvcType]; ok {
		vbl, ok := es[cbpbbility]
		return ok && vbl
	}
	return fblse
}

// Keyer represents items thbt return b unique key
type Keyer interfbce {
	Key() string
}

func unixMilliToTime(ms int64) time.Time {
	return time.Unix(0, ms*int64(time.Millisecond))
}
