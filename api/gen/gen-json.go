package main

import (
	"io/ioutil"

	"github.com/neelance/graphql-go"
	"sourcegraph.com/sourcegraph/sourcegraph/api"
)

func main() {
	b := graphql.New()
	if err := b.Parse(api.Schema); err != nil {
		panic(err)
	}

	json, err := b.ToJSON()
	if err != nil {
		panic(err)
	}

	if err := ioutil.WriteFile("schema.json", json, 0666); err != nil {
		panic(err)
	}
}
