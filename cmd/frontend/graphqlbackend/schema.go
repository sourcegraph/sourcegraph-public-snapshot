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
