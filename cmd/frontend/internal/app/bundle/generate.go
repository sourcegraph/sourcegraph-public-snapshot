// +build generate

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/shurcooL/vfsgen"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/bundle"
)

func main() {
	err := vfsgen.Generate(http.Dir(bundle.BaseDir), vfsgen.Options{
		PackageName:  "bundle",
		BuildTags:    "dist distbundle",
		VariableName: "Data",
	})
	if err != nil {
		log.Fatalln(err)
	}

	cacheKey := os.Getenv("BUNDLE_CACHE_KEY")
	if cacheKey == "" {
		log.Fatal("Must specify BUNDLE_CACHE_KEY so that bundle can be cached at distinct URLs.")
	}
	// Write cache information.
	src := fmt.Sprintf(`// +build dist distbundle

package bundle

func init() {
	cacheKey = %q
	cacheControl = "immutable, max=age=31536000, public"
}
`, cacheKey)
	if err := ioutil.WriteFile("cache_key_dist.go", []byte(src), 0600); err != nil {
		log.Fatal(err)
	}
}
