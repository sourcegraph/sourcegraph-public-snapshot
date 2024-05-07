package v1

import (
	"strings"
	"testing"

	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateEventFeatureAction(t *testing.T) {
	tests := []struct {
		name    string
		feature string
		action  string
		wantErr autogold.Value
	}{
		{
			name:    "feature is empty",
			feature: "",
			action:  "valid",
			wantErr: autogold.Expect("'feature', 'action' must both be provided"),
		},
		{
			name:    "action is empty",
			feature: "valid",
			action:  "",
			wantErr: autogold.Expect("'feature', 'action' must both be provided"),
		},
		{
			name:    "feature too long",
			feature: strings.Repeat("a", featureActionMaxLength+3),
			action:  "valid",
			wantErr: autogold.Expect("'feature' must be less than 64 characters"),
		},
		{
			name:    "action too long",
			feature: "valid",
			action:  strings.Repeat("a", featureActionMaxLength+3),
			wantErr: autogold.Expect("'action' must be less than 64 characters"),
		},
		{
			name:    "feature starts with uppercase",
			feature: "Invalid",
			action:  "valid",
			wantErr: autogold.Expect("'feature' must start with a lowercase letter and contain only letters, dashes, and dots"),
		},
		{
			name:    "action starts with uppercase",
			feature: "valid",
			action:  "Invalid",
			wantErr: autogold.Expect("'action' must start with a lowercase letter and contain only letters, dashes, and dots"),
		},
		{
			name:    "feature contains invalid characters",
			feature: "invalid_feature!",
			action:  "valid",
			wantErr: autogold.Expect("'feature' must start with a lowercase letter and contain only letters, dashes, and dots"),
		},
		{
			name:    "action contains invalid characters",
			feature: "valid",
			action:  "invalid_action!",
			wantErr: autogold.Expect("'action' must start with a lowercase letter and contain only letters, dashes, and dots"),
		},
		{
			name:    "valid feature and action 1",
			feature: "valid.feature",
			action:  "valid-action",
		},
		{
			name:    "valid feature and action 2",
			feature: "validFeature.foobar",
			action:  "valid.action",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEventFeatureAction(tt.feature, tt.action)
			if tt.wantErr != nil {
				require.Error(t, err)
				tt.wantErr.Equal(t, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
