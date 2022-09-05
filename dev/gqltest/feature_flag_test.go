package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFeatureFlags(t *testing.T) {
	const featureFlagOverrideFragment = `fragment FeatureFlagOverrideData on FeatureFlagOverride {
		id
		namespace {
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
		ID        string
		Namespace struct {
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
		  rolloutBasisPoints
		  overrides {
			...FeatureFlagOverrideData
		  }
		}
	}`

	type featureFlagResult struct {
		Name               string
		Value              *bool
		RolloutBasisPoints *int
		Overrides          []featureFlagOverrideResult
	}

	createFeatureFlag := func(name string, value *bool, rolloutBasisPoints *int) (featureFlagResult, error) {
		m := featureFlagFragment + featureFlagOverrideFragment + `
		mutation CreateFeatureFlag($name: String!, $value: Boolean, $rollout: Int) {
			createFeatureFlag(
				name: $name,
				value: $value,
				rolloutBasisPoints: $rollout,
			) {
				...FeatureFlagData
			}
		}`

		var res struct {
			Data struct {
				CreateFeatureFlag featureFlagResult
			}
		}
		params := map[string]any{"name": name, "value": value, "rollout": rolloutBasisPoints}
		err := client.GraphQL("", m, params, &res)
		return res.Data.CreateFeatureFlag, err
	}

	updateFeatureFlag := func(name string, value *bool, rolloutBasisPoints *int) (featureFlagResult, error) {
		m := featureFlagFragment + featureFlagOverrideFragment + `
		mutation UpdateFeatureFlag($name: String!, $value: Boolean, $rollout: Int) {
			updateFeatureFlag(
				name: $name,
				value: $value,
				rolloutBasisPoints: $rollout
			) {
				...FeatureFlagData
			}
		}`

		var res struct {
			Data struct {
				UpdateFeatureFlag featureFlagResult
			}
		}
		params := map[string]any{"name": name, "value": value, "rollout": rolloutBasisPoints}
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
		params := map[string]any{"name": name}
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

	// NOTE: these tests are intended to run in order, and not in parallel.
	// The orders matter for create, update, delete, list.

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
				Name:               "test_rollout",
				RolloutBasisPoints: &int343,
				Overrides:          []featureFlagOverrideResult{},
			}
			require.Equal(t, expected, res)
		})

		t.Run("BadArgsError", func(t *testing.T) {
			int343 := 343
			boolTrue := true
			_, err := createFeatureFlag("test_rollout", &boolTrue, &int343)
			require.Error(t, err)
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
				Name:               "test_rollout",
				RolloutBasisPoints: &int344,
				Overrides:          []featureFlagOverrideResult{},
			}
			require.Equal(t, expected, res)
		})

		t.Run("NonextantError", func(t *testing.T) {
			int344 := 344
			_, err := updateFeatureFlag("test_nonextant", nil, &int344)
			require.Error(t, err)
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
			t.Cleanup(func() {
				err := deleteFeatureFlag("test_concrete")
				require.NoError(t, err)
			})

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

	createOverride := func(namespace string, flagName string, value bool) (featureFlagOverrideResult, error) {
		m := featureFlagOverrideFragment + `
		mutation CreateFeatureFlagOverride($namespace: ID!, $flagName: String!, $value: Boolean!) {
			createFeatureFlagOverride(
				namespace: $namespace,
				flagName: $flagName,
				value: $value,
			) {
				...FeatureFlagOverrideData
			}
		}`

		var res struct {
			Data struct {
				CreateFeatureFlagOverride featureFlagOverrideResult
			}
		}
		params := map[string]any{"namespace": namespace, "flagName": flagName, "value": value}
		err := client.GraphQL("", m, params, &res)
		return res.Data.CreateFeatureFlagOverride, err
	}

	updateOverride := func(id string, value bool) (featureFlagOverrideResult, error) {
		m := featureFlagOverrideFragment + `
		mutation UpdateFeatureFlagOverride($id: ID!, $value: Boolean!) {
			updateFeatureFlagOverride(
				id: $id,
				value: $value,
			) {
				...FeatureFlagOverrideData
			}
		}`

		var res struct {
			Data struct {
				UpdateFeatureFlagOverride featureFlagOverrideResult
			}
		}
		params := map[string]any{"id": id, "value": value}
		err := client.GraphQL("", m, params, &res)
		return res.Data.UpdateFeatureFlagOverride, err
	}

	deleteOverride := func(id string) error {
		m := `
		mutation DeleteFeatureFlagOverride($id: ID!) {
			deleteFeatureFlagOverride(
				id: $id,
			) {
				alwaysNil
			}
		}`

		params := map[string]any{"id": id}
		return client.GraphQL("", m, params, nil)
	}

	t.Run("Overrides", func(t *testing.T) {
		orgID, err := client.CreateOrganization("testoverrides", "test")
		require.NoError(t, err)
		t.Cleanup(func() {
			client.DeleteOrganization(orgID)
		})

		userID, err := client.CreateUser("testuseroverrides", "test@override.com")
		require.NoError(t, err)
		removeTestUserAfterTest(t, userID)

		boolTrue := true
		flag, err := createFeatureFlag("test_override", &boolTrue, nil)
		require.NoError(t, err)
		t.Cleanup(func() {
			deleteFeatureFlag("test_override")
		})

		overrideT := t
		t.Run("Create", func(t *testing.T) {
			t.Run("OrgOverride", func(t *testing.T) {
				res, err := createOverride(orgID, flag.Name, false)
				require.NoError(t, err)
				overrideT.Cleanup(func() {
					deleteOverride(res.ID)
				})

				require.Equal(t, res.Namespace.ID, orgID)
				require.Equal(t, res.TargetFlag.Name, flag.Name)
				require.Equal(t, res.Value, false)

				t.Run("Update", func(t *testing.T) {
					updated, err := updateOverride(res.ID, true)
					require.NoError(t, err)
					require.Equal(t, updated.Value, true)
				})

			})

			t.Run("UserOverride", func(t *testing.T) {
				res, err := createOverride(userID, flag.Name, false)
				require.NoError(t, err)
				overrideT.Cleanup(func() {
					deleteOverride(res.ID)
				})

				require.Equal(t, res.Namespace.ID, userID)
				require.Equal(t, res.TargetFlag.Name, flag.Name)
				require.Equal(t, res.Value, false)
			})

			t.Run("NonextantFlag", func(t *testing.T) {
				_, err = createOverride(orgID, "test_nonextant", true)
				require.Error(t, err)
			})

			t.Run("NonextantUser", func(t *testing.T) {
				userString := "nonextant"
				_, err := createOverride(userString, "test_nonextant", true)
				require.Error(t, err)
			})

			t.Run("NonextantOrg", func(t *testing.T) {
				orgID := "nonextant"
				_, err := createOverride(orgID, "test_nonextant", true)
				require.Error(t, err)
			})
		})

		t.Run("ListFlagsIncludesOverride", func(t *testing.T) {
			res, err := listFeatureFlags()
			require.NoError(t, err)

			require.Len(t, res, 1)
			require.Len(t, res[0].Overrides, 2)

			o1 := res[0].Overrides[0]
			o2 := res[0].Overrides[1]
			require.Equal(t, o1.Namespace.ID, orgID)
			require.Equal(t, o2.Namespace.ID, userID)
			require.Equal(t, o1.TargetFlag.Name, "test_override")
			require.Equal(t, o2.TargetFlag.Name, "test_override")
			require.Equal(t, o1.Value, true)
			require.Equal(t, o2.Value, false)
		})
	})
}
