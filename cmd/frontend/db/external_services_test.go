package db

import "testing"

func TestExternalServicesStore_ValidateConfig(t *testing.T) {
	tests := map[string]struct {
		kind, config string
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
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			err := (&ExternalServicesStore{}).ValidateConfig(test.kind, test.config, nil)
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
