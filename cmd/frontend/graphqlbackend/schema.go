package graphqlbackend

import (
	_ "embed"
)

// MainSchema is the main raw graqhql schema.
//go:embed schema.graphql
var MainSchema string

// BatchesSchema is the Batch Changes raw graqhql schema.
//go:embed batches.graphql
var BatchesSchema string

// CodeIntelSchema is the Code Intel raw graqhql schema.
//go:embed codeintel.graphql
var CodeIntelSchema string

// DotcomSchema is the Dotcom schema extension raw graqhql schema.
//go:embed dotcom.graphql
var DotcomSchema string

// LicenseSchema is the Licensing raw graqhql schema.
//go:embed license.graphql
var LicenseSchema string

// CodeMonitorsSchema is the Code Monitoring raw graqhql schema.
//go:embed code_monitors.graphql
var CodeMonitorsSchema string

// InsightsSchema is the Code Insights raw graqhql schema.
//go:embed insights.graphql
var InsightsSchema string
