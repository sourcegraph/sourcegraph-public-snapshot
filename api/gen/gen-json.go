package main

import (
	"io/ioutil"

	"github.com/neelance/graphql-go"
	"sourcegraph.com/sourcegraph/sourcegraph/api"
)

func main() {
	json, err := graphql.SchemaToJSON(api.Schema)
	if err != nil {
		panic(err)
	}

	if err := ioutil.WriteFile("schema.json", json, 0666); err != nil {
		panic(err)
	}
}
