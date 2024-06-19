package modelconfig

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/modelconfig/embedded"
	"github.com/sourcegraph/sourcegraph/internal/modelconfig/types"
)

// Confirm that the model data currently in the repo is well-formed and valid.
func TestEmbeddedModelConfig(t *testing.T) {
	loadModelConfig := func(t *testing.T) types.ModelConfiguration {
		t.Helper()
		cfg, err := embedded.GetCodyGatewayModelConfig()
		require.NoError(t, err)
		return *cfg
	}

	t.Run("Loads", func(t *testing.T) {
		_ = loadModelConfig(t)
	})

	t.Run("IsValid", func(t *testing.T) {
		cfg := loadModelConfig(t)

		t.Run("Metadata", func(t *testing.T) {
			assert.Equal(t, types.CurrentModelSchemaVersion, cfg.SchemaVersion)
			assert.NotEmpty(t, cfg.Revision)
		})

		t.Run("Providers", func(t *testing.T) {
			for _, provider := range cfg.Providers {
				verr := validateProvider(provider)
				assert.NoError(t, verr, "Provider %q", provider.ID)
			}
		})

		t.Run("Models", func(t *testing.T) {
			for _, model := range cfg.Models {
				verr := validateModel(model)
				assert.NoError(t, verr, "Model %q", model.ModelRef)

				// Verify the model is referencing a known provider.
				var forKnownProvider bool
				for _, provider := range cfg.Providers {
					if strings.HasPrefix(string(model.ModelRef), string(provider.ID)+"::") {
						forKnownProvider = true
						break
					}
				}
				assert.True(t, forKnownProvider, "Model %q", model.ModelRef)
			}
		})

		t.Run("DefaultModels", func(t *testing.T) {
			isKnownModel := func(mref types.ModelRef) bool {
				for _, model := range cfg.Models {
					if model.ModelRef == mref {
						return true
					}
				}
				return false
			}

			assert.True(t, isKnownModel(cfg.DefaultModels.Chat))
			assert.True(t, isKnownModel(cfg.DefaultModels.CodeCompletion))
			assert.True(t, isKnownModel(cfg.DefaultModels.FastChat))
		})
	})
}
