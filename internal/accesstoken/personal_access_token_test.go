package accesstoken

import "testing"

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
		{
			name:    "token with no prefix",
			args:    args{token: "abc123"},
			want:    "abc123",
			wantErr: false,
		},
		{
			name:    "token with sgp_ prefix",
			args:    args{token: "sgp_abc123"},
			want:    "abc123",
			wantErr: false,
		},
		{
			name:    "token with sgph_ prefix",
			args:    args{token: "sgph_abc123"},
			want:    "abc123",
			wantErr: false,
		},
		{
			name:    "token with sgph_ prefix and instance-identifier",
			args:    args{token: "sgph_instanceidentifier_abc123"},
			want:    "abc123",
			wantErr: false,
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
