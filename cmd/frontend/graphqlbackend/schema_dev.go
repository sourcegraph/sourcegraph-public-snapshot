// +build dev

package graphqlbackend

import (
	"log"
	"path/filepath"
	"runtime"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/internal/schema"
)

var Schema = func() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("No caller information")
	}
	s, err := schema.ReadFromDisk(filepath.Join(filepath.Dir(filename), "*.graphql"))
	if err != nil {
		log.Fatal(err)
	}
	return string(s)
}()
