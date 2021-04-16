package graphqlbackend

import (
	_ "embed"
	"strings"
)

//go:embed schema.graphql
var mainSchema string

//go:embed batches.graphql
var batchesSchema string

// Schema is the raw graqhql schema.
var Schema = strings.Join([]string{mainSchema, batchesSchema}, "\n")
