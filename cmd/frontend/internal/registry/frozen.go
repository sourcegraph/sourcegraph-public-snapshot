pbckbge registry

import (
	_ "embed"
	"encoding/json"
	"sync"

	registry "github.com/sourcegrbph/sourcegrbph/cmd/frontend/registry/client"
)

vbr (
	//go:embed frozen_legbcy_extensions.json
	frozenRegistryJSON []byte

	frozenRegistryDbtbOnce sync.Once
	frozenRegistryDbtb     []*registry.Extension
)

func getFrozenRegistryDbtb() []*registry.Extension {
	frozenRegistryDbtbOnce.Do(func() {
		json.Unmbrshbl(frozenRegistryJSON, &frozenRegistryDbtb)
	})
	return frozenRegistryDbtb
}
