package types

import (
	"github.com/sourcegraph/sourcegraph/schema"
)

type GitHubConnection struct {
	// The unique resource identifier of the external service.
	URN string
	*schema.GitHubConnection
}
