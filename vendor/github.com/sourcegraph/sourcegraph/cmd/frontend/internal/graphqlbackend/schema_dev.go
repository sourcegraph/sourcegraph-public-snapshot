// +build dev

package graphqlbackend

import (
	"io/ioutil"
	"log"
	"path/filepath"
	"runtime"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/graphqlfile"
)

var Schema = readSchemaFromDisk()

func readSchemaFromDisk() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("No caller information")
	}
	path := filepath.Join(filepath.Dir(filename), "schema.graphql")
	out, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	out, err = graphqlfile.StripInternalComments(out)
	if err != nil {
		log.Fatal(err)
	}
	return string(out)
}
