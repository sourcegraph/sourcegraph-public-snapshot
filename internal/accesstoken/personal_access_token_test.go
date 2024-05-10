package accesstoken

// pre-commit:ignore_sourcegraph_token

import (
	"testing"
)

func TestGeneratePersonalAccessToken(t *testing.T) {
	type args struct {
		licenseKey    string
		isDevInstance bool
	}
	tests := []struct {
		name            string
		args            args
		wantTokenPrefix string
		wantBytes       []byte
		wantTokenLength int
		wantErr         bool
	}{
		{
			name: "valid token generation 1",
			args: args{
				licenseKey:    "abcdef1234abcdef1234abcdef1234abcdef1234",
				isDevInstance: false,
			},
			wantTokenPrefix: "sgp_5e37db464e9301ea_",
			wantTokenLength: 61,
			wantErr:         false,
		},
		{
			name: "valid token generation 2",
			args: args{
				licenseKey:    "foobar",
				isDevInstance: false,
			},
			wantTokenPrefix: "sgp_8844b0e0e754ec66_",
			wantTokenLength: 61,
			wantErr:         false,
		},
		{
			name: "valid token generation, dev instance",
			args: args{
				licenseKey:    "abcdef1234abcdef1234abcdef1234abcdef1234",
				isDevInstance: true,
			},
			wantTokenPrefix: "sgp_local_",
			wantTokenLength: 50,
			wantErr:         false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, _, err := GeneratePersonalAccessToken(tt.args.licenseKey, tt.args.isDevInstance)
			if (err != nil) != tt.wantErr {
				t.Errorf("GeneratePersonalAccessToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(token) != tt.wantTokenLength {
				t.Errorf("GeneratePersonalAccessToken() len(token) = %d, wantTokenLength = %d, token = '%s'", len(token), tt.wantTokenLength, token)
			}

			if len(tt.wantTokenPrefix) == 0 {
				t.Error("GeneratePersonalAccessToken() len(wantToken) is 0")
			}

			// Take the first characters to compare just the prefix
			if token[:len(tt.wantTokenPrefix)] != tt.wantTokenPrefix {
				t.Errorf("GeneratePersonalAccessToken() got = %v, want %v", token[:len(tt.wantTokenPrefix)], tt.wantTokenPrefix)
			}
		})
	}
}
