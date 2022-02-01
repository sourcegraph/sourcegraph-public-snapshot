package internal

import "github.com/sourcegraph/sourcegraph/internal/gqltestutil"

var client *gqltestutil.Client

func InitializeGraphQLClient() (err error) {
	client, err = gqltestutil.NewClient(SourcegraphEndpoint)
	return err
}

func GraphQLClient() *gqltestutil.Client {
	return client
}
