package conftypes

import "github.com/sourcegraph/sourcegraph/schema"

type UnifiedWatchable interface {
	Watchable
	UnifiedQuerier
}

type UnifiedQuerier interface {
	ServiceConnectionQuerier
	SiteConfigQuerier
}

type WatchableSiteConfig interface {
	SiteConfigQuerier
	Watchable
}

type ServiceConnectionQuerier interface {
	ServiceConnections() ServiceConnections
}

type SiteConfigQuerier interface {
	SiteConfig() schema.SiteConfiguration
}

type Watchable interface {
	Watch(func())
}
