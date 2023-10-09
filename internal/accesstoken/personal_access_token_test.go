package accesstoken

import (
	"testing"
)

func TestParsePersonalAccessToken(t *testing.T) {
	type args struct {
		token string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// Valid test cases
		{
			name:    "no prefix",
			args:    args{token: "abcdef1234abcdef1234abcdef1234abcdef1234"},
			want:    "abcdef1234abcdef1234abcdef1234abcdef1234",
			wantErr: false,
		},
		{
			name:    "sgp_ prefix",
			args:    args{token: "sgp_abcdef1234abcdef1234abcdef1234abcdef1234"},
			want:    "abcdef1234abcdef1234abcdef1234abcdef1234",
			wantErr: false,
		},
		{
			name:    "sgph_ prefix",
			args:    args{token: "sgph_abcdef1234abcdef1234abcdef1234abcdef1234"},
			want:    "abcdef1234abcdef1234abcdef1234abcdef1234",
			wantErr: false,
		},
		{
			name:    "sgph_ prefix and instance-identifier",
			args:    args{token: "sgph_0123456789abcdef_abcdef1234abcdef1234abcdef1234abcdef1234"},
			want:    "abcdef1234abcdef1234abcdef1234abcdef1234",
			wantErr: false,
		},
		// Error cases
		{
			name:    "no prefix, invalid length",
			args:    args{token: "abc123"},
			want:    "",
			wantErr: true,
		},
		{
			name:    "invalid prefix, invalid length",
			args:    args{token: "sgptest_abcdef1234abcdef1234abcdef1234abcdef1234"},
			want:    "",
			wantErr: true,
		},
		{
			name:    "prefix, invalid length",
			args:    args{token: "sgp_abcdef1234abcdef1234abcdef1234abcdef12345"},
			want:    "",
			wantErr: true,
		},
		{
			name:    "too-short instance identifer",
			args:    args{token: "sgph_01234_abcdef1234abcdef1234abcdef1234abcdef1234"},
			want:    "",
			wantErr: true,
		},
		{
			name:    "too-long instance identifer",
			args:    args{token: "sgph_0123456789abcdef0_abcdef1234abcdef1234abcdef1234abcdef1234"},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePersonalAccessToken(tt.args.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParsePersonalAccessToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParsePersonalAccessToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGeneratePersonalAccessToken(t *testing.T) {
	type args struct {
		licenseKey    string
		isDevInstance bool
	}
	tests := []struct {
		name      string
		args      args
		wantToken string
		wantBytes []byte
		wantErr   bool
	}{
		{
			name:      "valid token generation 1",
			args:      args{licenseKey: "abcdef1234abcdef1234abcdef1234abcdef1234", isDevInstance: false},
			wantToken: "sgph_5e37db464e9301ea_",
			wantErr:   false,
		},
		{
			name:      "valid token generation 2",
			args:      args{licenseKey: "foobar", isDevInstance: false},
			wantToken: "sgph_8844b0e0e754ec66_",
			wantErr:   false,
		},
		{
			name:      "valid token generation, dev instance",
			args:      args{licenseKey: "abcdef1234abcdef1234abcdef1234abcdef1234", isDevInstance: true},
			wantToken: "sgph_local_",
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, _, err := GeneratePersonalAccessToken(tt.args.licenseKey, tt.args.isDevInstance)
			if (err != nil) != tt.wantErr {
				t.Errorf("GeneratePersonalAccessToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(tt.wantToken) == 0 {
				t.Error("GeneratePersonalAccessToken() len(wantToken) is 0")
			}

			// Take the first characters to compare just the prefix
			if token[:len(tt.wantToken)] != tt.wantToken {
				t.Errorf("GeneratePersonalAccessToken() got = %v, want %v", token[:len(tt.wantToken)], tt.wantToken)
			}
		})
	}
}
