package main

import (
	"io/ioutil"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/graphqlbackend"

	"github.com/neelance/graphql-go"
)

func main() {
	json, err := graphql.MustParseSchema(graphqlbackend.Schema, nil).ToJSON()
	if err != nil {
		panic(err)
	}

	if err := ioutil.WriteFile("../../../../client/schema.json", json, 0666); err != nil {
		panic(err)
	}
}
