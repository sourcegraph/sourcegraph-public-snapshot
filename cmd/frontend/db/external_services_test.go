package db

import (
	"context"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"testing"
)

func TestExternalServicesStore_ValidateConfig(t *testing.T) {
	tests := map[string]struct {
		kind, config string
		setup        func(t *testing.T)
		teardown     func()
		wantErr      string
	}{
		"0 errors": {
			kind:    "GITHUB",
			config:  `{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc"}`,
			wantErr: "",
		},
		"1 error": {
			kind:    "GITHUB",
			config:  `{"url": "https://github.com", "repositoryQuery": ["none"], "token": ""}`,
			wantErr: "1 error occurred:\n\t* token: String length must be greater than or equal to 1\n\n",
		},
		"2 errors": {
			kind:    "GITHUB",
			config:  `{"url": "https://github.com", "repositoryQuery": ["none"], "token": "", "x": 123}`,
			wantErr: "2 errors occurred:\n\t* Additional property x is not allowed\n\t* token: String length must be greater than or equal to 1\n\n",
		},
		"no conflicting rate limit": {
			kind:   "GITHUB",
			config: `{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc", "rateLimit": {"enabled": true, "requestsPerHour": 5000}}`,
			setup: func(t *testing.T) {
				Mocks.ExternalServices.List = func(opt ExternalServicesListOptions) ([]*types.ExternalService, error) {
					return nil, nil
				}
			},
			wantErr: "",
		},
		"conflicting rate limit": {
			kind:   "GITHUB",
			config: `{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc", "rateLimit": {"enabled": true, "requestsPerHour": 5000}}`,
			setup: func(t *testing.T) {
				t.Cleanup(func() {
					Mocks.ExternalServices.List = nil
				})
				Mocks.ExternalServices.List = func(opt ExternalServicesListOptions) ([]*types.ExternalService, error) {
					return []*types.ExternalService{
						{
							ID:          1,
							Kind:        "GITHUB",
							DisplayName: "GITHUB 1",
							Config:      `{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc", "rateLimit": {"enabled": true, "requestsPerHour": 5000}}`,
						},
					}, nil
				}
			},
			wantErr: "1 error occurred:\n\t* existing external service, \"GITHUB 1\", already has a rate limit set\n\n",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if test.setup != nil {
				test.setup(t)
			}

			err := (&ExternalServicesStore{}).ValidateConfig(context.Background(), 0, test.kind, test.config, nil)
			var errStr string
			if err != nil {
				errStr = err.Error()
			}
			if errStr != test.wantErr {
				t.Errorf("got error %q, want %q", errStr, test.wantErr)
			}
		})
	}
}
