/*
Package monitoring declares types for Sourcegraph's monitoring generator as well as the generator implementation itself.

To learn more about developing monitoring, see the guide: https://about.sourcegraph.com/handbook/engineering/observability/monitoring

To learn more about the generator, see the top-level program: https://github.com/sourcegraph/sourcegraph/tree/main/monitoring
*/
package monitoring

//go:generate go run github.com/princjef/gomarkdoc/cmd/gomarkdoc . -o README.md

import _ "github.com/princjef/gomarkdoc" // Pin version of godoc-to-markdown generator
