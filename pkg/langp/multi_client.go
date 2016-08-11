package langp

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf/feature"
)

const (
	Go   Language = "Go"
	Java Language = "Java"
)

// Language represents a programming language.
type Language string

// extToLanguage is a map of file extensions to their respective language.
var extToLanguage = map[string]Language{
	".go":   Go,
	".java": Java,
}

// DefaultClient is the default language processor client.
var DefaultClient *MultiClient

func init() {
	if !feature.Features.Universe {
		return
	}

	newClient := func(v string) *Client {
		client, err := NewClient(os.Getenv(v))
		if err != nil {
			log.Fatalf("$%s %v", v, err)
		}
		return client
	}
	DefaultClient = &MultiClient{
		Clients: map[Language]*Client{
			Go:   newClient("SG_GO_LANGUAGE_PROCESSOR"),
			Java: newClient("SG_JAVA_LANGUAGE_PROCESSOR"),
		},
	}
}

// MultiClient is a client which wraps multiple underlying clients and is
// responsible for invoking the proper client (or combining results) depending
// on the request / which langauge the source file is.
type MultiClient struct {
	// Clients is a map of languages to their respective clients.
	Clients map[Language]*Client
}

// Prepare invokes Prepare on each underlying client returning the first error
// that occurs, if any.
func (mc *MultiClient) Prepare(r *RepoRev) error {
	for _, cl := range mc.Clients {
		if err := cl.Prepare(r); err != nil {
			return err
		}
	}
	return nil
}

// find finds the client related to the file extension for filename.
func (mc *MultiClient) find(filename string) (*Client, error) {
	language, ok := extToLanguage[filepath.Ext(filename)]
	if !ok {
		return nil, fmt.Errorf("MultiClient: no language registered for extension %q", filepath.Ext(filename))
	}
	client, ok := mc.Clients[language]
	if !ok {
		return nil, fmt.Errorf("MultiClient: no client registered for language %s", language)
	}
	return client, nil
}

// Definition invokes Definition on the client whose language matches p.File.
func (mc *MultiClient) Definition(p *Position) (*Range, error) {
	c, err := mc.find(p.File)
	if err != nil {
		return nil, err
	}
	return c.Definition(p)
}

// Hover invokes Hover on the client whose language matches p.File.
func (mc *MultiClient) Hover(p *Position) (*Hover, error) {
	c, err := mc.find(p.File)
	if err != nil {
		return nil, err
	}
	return c.Hover(p)
}

// LocalRefs invokes LocalRefs on the client whose language matches p.File.
func (mc *MultiClient) LocalRefs(p *Position) (*LocalRefs, error) {
	c, err := mc.find(p.File)
	if err != nil {
		return nil, err
	}
	return c.LocalRefs(p)
}

// ExternalRefs invokes ExternalRefs for each client and combines the results,
// returning the first error that occurs, if any.
func (mc *MultiClient) ExternalRefs(r *RepoRev) (*ExternalRefs, error) {
	result := &ExternalRefs{}
	for _, c := range mc.Clients {
		v, err := c.ExternalRefs(r)
		if err != nil {
			return nil, err
		}
		result.Defs = append(result.Defs, v.Defs...)
	}
	return result, nil
}

// ExportedSymbols invokes ExportedSymbols for each client and combines the
// results, returning the first error that occurs, if any.
func (mc *MultiClient) ExportedSymbols(r *RepoRev) (*ExportedSymbols, error) {
	result := &ExportedSymbols{}
	for _, c := range mc.Clients {
		v, err := c.ExportedSymbols(r)
		if err != nil {
			return nil, err
		}
		result.Defs = append(result.Defs, v.Defs...)
	}
	return result, nil
}
