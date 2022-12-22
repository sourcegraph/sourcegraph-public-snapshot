package registry

import (
	"context"
	_ "embed"
	"encoding/json"
	"sync"

	frontendregistry "github.com/sourcegraph/sourcegraph/cmd/frontend/registry/api"
	registry "github.com/sourcegraph/sourcegraph/cmd/frontend/registry/client"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/registry/stores"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

// useFrozenRegistryData is whether we should serve the frozen registry data from sourcegraph.com as
// of 2022-12-22 instead of from the DB. This is a step along the way toward removing the extensions
// database code (as part of the removal of the legacy Sourcegraph extension API).
func useFrozenRegistryData(ctx context.Context, db database.DB) bool {
	ffs, _ := db.FeatureFlags().GetGlobalFeatureFlags(ctx)
	return ffs["frozen-registry-data"]
}

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

func frozenRegistryList(ctx context.Context, db database.DB, opt stores.ExtensionsListOptions) ([]*registry.Extension, error) {
	xs := getFrozenRegistryData()
	return frontendregistry.FilterRegistryExtensions(xs, opt.Query), nil
}

func frozenRegistryGetByUUID(ctx context.Context, db database.DB, uuid string) (*registry.Extension, error) {
	xs := getFrozenRegistryData()
	return frontendregistry.FindRegistryExtension(xs, "uuid", uuid), nil
}

func frozenRegistryGetByExtensionID(ctx context.Context, db database.DB, extensionID string) (*registry.Extension, error) {
	xs := getFrozenRegistryData()
	return frontendregistry.FindRegistryExtension(xs, "extensionID", extensionID), nil
}
