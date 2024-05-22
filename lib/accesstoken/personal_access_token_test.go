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
		{
			name:    "sgph_ prefix and local instance-identifier",
			args:    args{token: "sgph_local_abcdef1234abcdef1234abcdef1234abcdef1234"},
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
