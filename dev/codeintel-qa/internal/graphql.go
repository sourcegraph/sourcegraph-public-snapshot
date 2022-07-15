package internal

import (
	"os"

	"github.com/sourcegraph/sourcegraph/internal/gqltestutil"
)

var client *gqltestutil.Client

func InitializeGraphQLClient() (err error) {
	client, err = gqltestutil.NewClient(SourcegraphEndpoint, os.Stderr, os.Stderr)
	return err
}

func GraphQLClient() *gqltestutil.Client {
	return client
}
