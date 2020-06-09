package types

import (
	"github.com/sourcegraph/sourcegraph/schema"
)

type GitHubConnection struct {
	URN string
	*schema.GitHubConnection
}
