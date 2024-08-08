package registry

import (
	_ "embed"
	"encoding/json"
	"sync"
)

var (
	//go:embed frozen_legacy_extensions.json
	frozenRegistryJSON []byte

	frozenRegistryDataOnce sync.Once
	frozenRegistryData     []*Extension
)

func getFrozenRegistryData() []*Extension {
	frozenRegistryDataOnce.Do(func() {
		json.Unmarshal(frozenRegistryJSON, &frozenRegistryData)
	})
	return frozenRegistryData
}
