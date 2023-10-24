package accesstoken

import (
	"testing"
)

// This file contains tests for token generation code, so it legitimately contains what
// appears to be valid tokens. This is fine because those are test tokens. The comment
// just below tells the pre-commit hook to simply skip this file.
// pre-commit:ignore_sourcegraph_token

func TestGenerateDotcomUserGatewayAccessToken(t *testing.T) {
	type args struct {
		apiToken string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "valid token 1",
			args:    args{apiToken: "0123456789abcdef0123456789abcdef01234567"},
			want:    "sgd_ee6ba2aa505be17522e936ebac2c31c108d58ebfc8d483ed75a6b298506cb949",
			wantErr: false,
		},
		{
			name:    "valid token 1b",
			args:    args{apiToken: "sgp_0123456789abcdef0123456789abcdef01234567"},
			want:    "sgd_ee6ba2aa505be17522e936ebac2c31c108d58ebfc8d483ed75a6b298506cb949",
			wantErr: false,
		},
		{
			name:    "valid token 2",
			args:    args{apiToken: "sgp_abcdef0123456789abcdef0123456789abcdef01"},
			want:    "sgd_3b0ed67f378c6f62c1b17faa738f46f2889beb6856eb5432e7d1aedd0953415b",
			wantErr: false,
		},
		{
			name:    "valid token 2b",
			args:    args{apiToken: "sgph_abcdef0123456789_abcdef0123456789abcdef0123456789abcdef01"},
			want:    "sgd_3b0ed67f378c6f62c1b17faa738f46f2889beb6856eb5432e7d1aedd0953415b",
			wantErr: false,
		},
		{
			name:    "invalid token format 1",
			args:    args{apiToken: "sgp_abcdef0123456789"},
			want:    "",
			wantErr: true,
		},
		{
			name:    "invalid token format 2",
			args:    args{apiToken: "sgp_zzzzzz0123456789abcdef0123456789abcdef01"},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateDotcomUserGatewayAccessToken(tt.args.apiToken)

			if got != tt.want {
				t.Errorf("GenerateDotcomUserGatewayAccessToken() = %v, want %v", got, tt.want)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateDotcomUserGatewayAccessToken() error: got = %t, want = %t, err = %s", (err != nil), tt.wantErr, err)
			}
		})
	}
}
