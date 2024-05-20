package assets

import _ "embed"

// Icon is a data-uri encoded svg.
// We need to embed the icons in our output svg
// rather than using urls to externally hosted svgs
// as _most_ sites block requests to external content for security reasons
//
// Note: when adding a new icon ensure there is no trailing whitespace/newline
type Icon string

var (
	//go:embed internet
	Internet Icon
	//go:embed cloudflare
	Cloudflare Icon
	//go:embed externalipaddress
	CloudExternalIPAddress Icon
	//go:embed loadbalancer
	CloudLoadBalancer Icon
	//go:embed cloudarmor
	CloudArmor Icon
	//go:embed cloudrun
	CloudRun Icon
	//go:embed cloudmemorystore
	CloudMemorystore Icon
	//go:embed bigquery
	BigQuery Icon
	//go:embed cloudsql
	CloudSQL Icon
	//go:embed cloudmonitoring
	CloudMonitoring Icon
	//go:embed cloudtrace
	CloudTrace Icon
	//go:embed sentry
	Sentry Icon
	//go:embed opsgenie
	Opsgenie Icon
	//go:embed vpc
	VPC Icon
)
