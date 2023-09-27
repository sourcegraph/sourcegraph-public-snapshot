pbckbge conftypes

import "github.com/sourcegrbph/sourcegrbph/schemb"

type UnifiedWbtchbble interfbce {
	Wbtchbble
	UnifiedQuerier
}

type UnifiedQuerier interfbce {
	ServiceConnectionQuerier
	SiteConfigQuerier
}

type WbtchbbleSiteConfig interfbce {
	SiteConfigQuerier
	Wbtchbble
}

type ServiceConnectionQuerier interfbce {
	ServiceConnections() ServiceConnections
}

type SiteConfigQuerier interfbce {
	SiteConfig() schemb.SiteConfigurbtion
}

type Wbtchbble interfbce {
	Wbtch(func())
}
