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

type originMapEntry struct {
	Prefix string
	Origin string
}

var originMap []*originMapEntry

func init() {
	ogmp, err := parseFromEnv()
	if err != nil {
		log.Fatal(err)
	}
	originMap = append(ogmp, &originMapEntry{Prefix: "github.com/", Origin: "https://github.com/%.git"})
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

func parseFromEnv() ([]*originMapEntry, error) {
	return parse(env.Get("ORIGIN_MAP", "", `space separated list of mappings from repo name prefix to origin url, for example "github.com/!https://github.com/%.git"`))
}

func parse(raw string) (originMap []*originMapEntry, err error) {
	for _, e := range strings.Fields(raw) {
		p := strings.Split(e, "!")
		if len(p) != 2 {
			return nil, fmt.Errorf("invalid ORIGIN_MAP entry: %s", e)
		}
		originMap = append(originMap, &originMapEntry{Prefix: p[0], Origin: p[1]})
	}
	return
}
