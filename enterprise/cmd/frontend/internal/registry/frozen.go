package registry

import (
	_ "embed"
	"encoding/json"
	"sync"

	registry "github.com/sourcegraph/sourcegraph/cmd/frontend/registry/client"
)

var (
	//go:embed frozen_legacy_extensions.json
	frozenRegistryJSON []byte

	frozenRegistryDataOnce sync.Once
	frozenRegistryData     []*registry.Extension
)

func getFrozenRegistryData() []*registry.Extension {
	frozenRegistryDataOnce.Do(func() {
		json.Unmarshal(frozenRegistryJSON, &frozenRegistryData)
	})
	return frozenRegistryData
}
