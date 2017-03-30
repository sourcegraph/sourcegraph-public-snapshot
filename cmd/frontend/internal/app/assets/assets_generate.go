// +build generate

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/shurcooL/vfsgen"
)

func getMainBundleFilename(dir string) (string, error) {
	fis, err := ioutil.ReadDir(dir)
	if err != nil {
		return "", err
	}

	var candidates []string
	for _, fi := range fis {
		if fi.Mode().IsRegular() && strings.HasPrefix(fi.Name(), "main.") && strings.HasSuffix(fi.Name(), ".js") {
			candidates = append(candidates, fi.Name())
		}
	}

	if len(candidates) == 0 {
		return "", fmt.Errorf("No output directories in %s. You must first run Webpack via `yarn run build` in the ui directory to produce the output directory.", dir)
	} else if len(candidates) != 1 {
		return "", fmt.Errorf("Multiple output directories in %s. There must be exactly one. Did `yarn run build` not properly clean up %s before producing output?", dir, dir)
	}
	return candidates[0], nil
}

func main() {
	// Find the hashed assets dir.
	dir := "../../ui/assets/"
	mainBundleFilename, err := getMainBundleFilename(dir)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Webpack main bundle file is", mainBundleFilename)

	err = vfsgen.Generate(http.Dir(dir), vfsgen.Options{
		PackageName:  "assets",
		BuildTags:    "dist",
		VariableName: "Assets",
	})
	if err != nil {
		log.Fatalln(err)
	}

	src := fmt.Sprintf(`// +build dist

package assets

const mainJavaScriptBundlePath = %q
`, "/"+mainBundleFilename)
	if err := ioutil.WriteFile("main_bundle_dist.go", []byte(src), 0600); err != nil {
		log.Fatal(err)
	}
}
