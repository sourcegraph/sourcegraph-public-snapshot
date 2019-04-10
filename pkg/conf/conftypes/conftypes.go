package conftypes

import "reflect"

// ServiceConnections represents configuration about how the deployment
// internally connects to services. These are settings that need to be
// propagated from the frontend to other services, so that the frontend
// can be the source of truth for all configuration.
type ServiceConnections struct {
	// GitServers is the addresses of gitserver instances that should be talked
	// to.
	GitServers []string `json:"gitServers"`

	// PostgreSQL environment variables configured on the frontend.
	PGSSLMODE, PGUSER, PGPASSWORD, PGHOST, PGPORT, PGDATABASE string
}

// RawUnified is the unparsed variant of conf.Unified.
type RawUnified struct {
	Site, Critical     string
	ServiceConnections ServiceConnections
}

// Equal tells if the two configurations are equal or not.
func (r RawUnified) Equal(other RawUnified) bool {
	return r.Site == other.Site && r.Critical == other.Critical && reflect.DeepEqual(r.ServiceConnections, other.ServiceConnections)
}
