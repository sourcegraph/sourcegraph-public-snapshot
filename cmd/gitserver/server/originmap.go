// Package originmap maps Sourcegraph repository URIs to repository
// origins (i.e., clone URLs). It accepts external customization via
// the ORIGIN_MAP environment variable.
//
// It always includes the mapping
// "github.com/!https://github.com/%.git" (github.com ->
// https://github.com/%.git)
package server

import (
	"fmt"
	"log"
	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
)

type prefixAndOrgin struct {
	Prefix string
	Origin string
}

var originMapEnv = env.Get("ORIGIN_MAP", "", `space separated list of mappings from repo name prefix to origin url, for example "github.com/!https://github.com/%.git"`)
var gitoliteHostsEnv = env.Get("GITOLITE_HOSTS", "", `space separated list of mappings from repo name prefix to gitolite hosts"`)

var originMap []prefixAndOrgin
var gitoliteHostMap []prefixAndOrgin

func init() {
	var err error
	originMap, err = parse(originMapEnv)
	if err != nil {
		log.Fatal(err)
	}

	gitoliteHostMap, err = parse(gitoliteHostsEnv)
	if err != nil {
		log.Fatal(err)
	}
	for _, entry := range gitoliteHostMap {
		originMap = append(originMap, prefixAndOrgin{Prefix: entry.Prefix, Origin: entry.Origin + "/%"})
	}

	originMap = append(originMap, prefixAndOrgin{Prefix: "github.com/", Origin: "https://github.com/%.git"})
}

// Map maps the repo URI to the repository origin (clone URL). Returns empty string if no mapping was found.
func OriginMap(repoURI string) string {
	for _, entry := range originMap {
		if strings.HasPrefix(repoURI, entry.Prefix) {
			return strings.Replace(entry.Origin, "%", strings.TrimPrefix(repoURI, entry.Prefix), 1)
		}
	}
	return ""
}

func parse(raw string) (originMap []prefixAndOrgin, err error) {
	for _, e := range strings.Fields(raw) {
		p := strings.Split(e, "!")
		if len(p) != 2 {
			return nil, fmt.Errorf("invalid entry: %s", e)
		}
		originMap = append(originMap, prefixAndOrgin{Prefix: p[0], Origin: p[1]})
	}
	return
}
