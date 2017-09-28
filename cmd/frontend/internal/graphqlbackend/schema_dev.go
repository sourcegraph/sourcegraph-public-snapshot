// +build dev

package graphqlbackend

import (
	"io/ioutil"
	"log"
	"path/filepath"
	"runtime"
)

var Schema string

func init() {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("No caller information")
	}
	path := filepath.Join(filepath.Dir(filename), "schema.graphql")
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	Schema = string(raw)
}
