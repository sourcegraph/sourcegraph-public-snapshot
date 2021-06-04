package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFeatureFlags(t *testing.T) {
	const featureFlagOverrideFragment = `fragment FeatureFlagOverrideData on FeatureFlagOverride {
		id
		user {
			username 
		}
		org {
			id
		}
		targetFlag {
			...on FeatureFlagBoolean {
				name
			}
			...on FeatureFlagRollout{
				name
			}
		}
		value
	}`

	type featureFlagOverrideResult struct {
		ID   string
		User struct {
			Username string
		}
		Org struct {
			ID string
		}
		TargetFlag struct {
			Name string
		}
		Value bool
	}

	const featureFlagFragment = `fragment FeatureFlagData on FeatureFlag {
		...on FeatureFlagBoolean{
		  name
		  value
		  overrides {
			...FeatureFlagOverrideData
		  }
		}
		...on FeatureFlagRollout {
		  name
		  rollout
		  overrides {
			...FeatureFlagOverrideData
		  }
		}
	}`

	type featureFlagResult struct {
		Name      string
		Value     *bool
		Rollout   *int
		Overrides []featureFlagOverrideResult
	}

	createFeatureFlag := func(name string, value *bool, rollout *int) (featureFlagResult, error) {
		m := featureFlagFragment + featureFlagOverrideFragment + `
		mutation CreateFeatureFlag($name: String!, $value: Boolean, $rollout: Int) {
			createFeatureFlag(
				name: $name,
				value: $value,
				rollout: $rollout,
			) {
				...FeatureFlagData
			}
		}`

		var res struct {
			Data struct {
				CreateFeatureFlag featureFlagResult
			}
		}
		params := map[string]interface{}{"name": name, "value": value, "rollout": rollout}
		err := client.GraphQL("", m, params, &res)
		return res.Data.CreateFeatureFlag, err
	}

	updateFeatureFlag := func(name string, value *bool, rollout *int) (featureFlagResult, error) {
		m := featureFlagFragment + featureFlagOverrideFragment + `
		mutation UpdateFeatureFlag($name: String!, $value: Boolean, $rollout: Int) {
			updateFeatureFlag(
				name: $name,
				value: $value,
				rollout: $rollout
			) {
				...FeatureFlagData
			}
		}`

		var res struct {
			Data struct {
				UpdateFeatureFlag featureFlagResult
			}
		}
		params := map[string]interface{}{"name": name, "value": value, "rollout": rollout}
		err := client.GraphQL("", m, params, &res)
		return res.Data.UpdateFeatureFlag, err
	}

	deleteFeatureFlag := func(name string) error {
		m := `mutation DeleteFeatureFlag($name: String!){
			deleteFeatureFlag(
				name: $name,
			) {
				alwaysNil		
			}
		}`
		params := map[string]interface{}{"name": name}
		return client.GraphQL("", m, params, nil)
	}

	listFeatureFlags := func() ([]featureFlagResult, error) {
		m := featureFlagFragment + featureFlagOverrideFragment + `
		query ListFeatureFlags{
			featureFlags{
				...FeatureFlagData
			}
		}`

		var res struct {
			Data struct {
				FeatureFlags []featureFlagResult
			}
		}
		err := client.GraphQL("", m, nil, &res)
		return res.Data.FeatureFlags, err
	}

	t.Run("Create", func(t *testing.T) {
		t.Run("Concrete", func(t *testing.T) {
			boolTrue := true
			res, err := createFeatureFlag("test_concrete", &boolTrue, nil)
			require.NoError(t, err)

			expected := featureFlagResult{
				Name:      "test_concrete",
				Value:     &boolTrue,
				Overrides: []featureFlagOverrideResult{},
			}
			require.Equal(t, expected, res)
		})

		t.Run("Rollout", func(t *testing.T) {
			int343 := 343
			res, err := createFeatureFlag("test_rollout", nil, &int343)
			require.NoError(t, err)

			expected := featureFlagResult{
				Name:      "test_rollout",
				Rollout:   &int343,
				Overrides: []featureFlagOverrideResult{},
			}
			require.Equal(t, expected, res)
		})
	})

	t.Run("Update", func(t *testing.T) {
		t.Run("Concrete", func(t *testing.T) {
			boolFalse := false
			res, err := updateFeatureFlag("test_concrete", &boolFalse, nil)
			require.NoError(t, err)

			expected := featureFlagResult{
				Name:      "test_concrete",
				Value:     &boolFalse,
				Overrides: []featureFlagOverrideResult{},
			}
			require.Equal(t, expected, res)
		})

		t.Run("Rollout", func(t *testing.T) {
			int344 := 344
			res, err := updateFeatureFlag("test_rollout", nil, &int344)
			require.NoError(t, err)

			expected := featureFlagResult{
				Name:      "test_rollout",
				Rollout:   &int344,
				Overrides: []featureFlagOverrideResult{},
			}
			require.Equal(t, expected, res)
		})
	})

	t.Run("Delete", func(t *testing.T) {
		t.Run("Concrete", func(t *testing.T) {
			err := deleteFeatureFlag("test_concrete")
			require.NoError(t, err)
		})

		t.Run("Rollout", func(t *testing.T) {
			err := deleteFeatureFlag("test_rollout")
			require.NoError(t, err)
		})
	})

	t.Run("List", func(t *testing.T) {
		t.Run("None", func(t *testing.T) {
			res, err := listFeatureFlags()
			require.NoError(t, err)
			require.Len(t, res, 0)
		})

		t.Run("Some", func(t *testing.T) {
			// Create a feature flag first
			boolTrue := true
			_, err := createFeatureFlag("test_concrete", &boolTrue, nil)
			require.NoError(t, err)

			// Then see if it shows up when we list it
			res, err := listFeatureFlags()
			require.NoError(t, err)

			expected := []featureFlagResult{{
				Name:      "test_concrete",
				Value:     &boolTrue,
				Overrides: []featureFlagOverrideResult{},
			}}
			require.Equal(t, res, expected)
		})
	})

	t.Run("Overrides", func(t *testing.T) {

	})
}
