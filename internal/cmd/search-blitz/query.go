package main

import _ "embed"

//go:embed search.graphql
var graphQLSearchQuery string

//go:embed attribution.graphql
var graphQLAttributionQuery string
