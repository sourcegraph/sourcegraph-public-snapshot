package conftypes

import (
	"reflect"
	"time"
)

// ServiceConnections represents configuration about how the deployment
// internally connects to services. These are settings that need to be
// propagated from the frontend to other services, so that the frontend
// can be the source of truth for all configuration.
type ServiceConnections struct {
	// GitServers is the addresses of gitserver instances that should be
	// talked to.
	GitServers []string `json:"gitServers"`

	// PostgresDSN is the PostgreSQL DB data source name.
	// eg: "postgres://sg@pgsql/sourcegraph?sslmode=false"
	PostgresDSN string `json:"postgresDSN"`

	// CodeIntelPostgresDSN is the PostgreSQL DB data source name for the
	// code intel database.
	// eg: "postgres://sg@pgsql/sourcegraph_codeintel?sslmode=false"
	CodeIntelPostgresDSN string `json:"codeIntelPostgresDSN"`

	// CodeInsightsDSN is the PostgreSQL DB data source name for the
	// code insights database.
	// eg: "postgres://sg@pgsql/sourcegraph_codeintel?sslmode=false"
	CodeInsightsDSN string `json:"codeInsightsPostgresDSN"`

	// Searchers is the addresses of searcher instances that should be talked to.
	Searchers []string `json:"searchers"`
	// Zoekts is the addresses of Zoekt instances to talk to.
	Zoekts []string `json:"zoekts"`
	// ZoektListTTL is the TTL of the internal cache that Zoekt clients use to
	// cache the list of indexed repository. After TTL is over, new list will
	// get requested from Zoekt shards.
	ZoektListTTL time.Duration `json:"zoektListTTL"`
}

// RawUnified is the unparsed variant of conf.Unified.
type RawUnified struct {
	ID                 int32
	Site               string
	ServiceConnections ServiceConnections
}

// Equal tells if the two configurations are equal or not.
func (r RawUnified) Equal(other RawUnified) bool {
	return r.Site == other.Site && reflect.DeepEqual(r.ServiceConnections, other.ServiceConnections)
}
