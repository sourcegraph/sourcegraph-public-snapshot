// Package cli exposes a CLI flag that the program uses to select the
// local store to use (e.g., PostgreSQL or filesystem).
//
// It contains a registry of available stores, and the CLI flag's
// value must refer to one of these.
package cli

import (
	"fmt"
	"log"

	"strings"

	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/server/serverctx"
	"src.sourcegraph.com/sourcegraph/sgx/cli"
	"src.sourcegraph.com/sourcegraph/store"
)

// Flags exposes CLI flags to the sgx program that configure which store implementation to use for each interface.
type Flags struct {
	Store string `short:"s" long:"local-store" description:"backing store for local (non-federated) data" default:"fs" value-name:"TYPE"`
}

func (f *Flags) stores() store.Stores {
	s, ok := stores[f.Store]
	if !ok {
		log.Fatalf("invalid --local-store value: no store named %q", f.Store)
	}
	return *s
}

var ActiveFlags Flags

func init() {
	cli.PostInit = append(cli.PostInit, func() {
		g, err := cli.Serve.AddGroup("Local store", "Local store configuration", &ActiveFlags)
		if err != nil {
			log.Fatal(err)
		}

		for _, opt := range g.Options() {
			switch opt.LongName {
			case "local-store":
				opt.Description += fmt.Sprintf(" (choices: %s)", strings.Join(storeNames, " "))
			}
		}
	})

	serverctx.Funcs = append(serverctx.Funcs, func(ctx context.Context) (context.Context, error) {
		return store.WithStores(ctx, ActiveFlags.stores()), nil
	})
}

var (
	stores     map[string]*store.Stores
	storeNames []string
)

func RegisterStores(name string, s *store.Stores) {
	if stores == nil {
		stores = map[string]*store.Stores{}
	}
	stores[name] = s
	storeNames = append(storeNames, name)
}
