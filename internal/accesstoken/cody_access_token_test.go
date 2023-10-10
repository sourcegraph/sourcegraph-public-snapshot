package accesstoken

import "testing"

func TestGenerateDotcomUserGatewayAccessToken(t *testing.T) {
	type args struct {
		apiToken string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "valid token 1",
			args: args{apiToken: "0123456789abcdef0123456789abcdef01234567"},
			want: "sgd_ee6ba2aa505be17522e936ebac2c31c108d58ebfc8d483ed75a6b298506cb949",
		},
		{
			name: "valid token 1b",
			args: args{apiToken: "sgp_0123456789abcdef0123456789abcdef01234567"},
			want: "sgd_ee6ba2aa505be17522e936ebac2c31c108d58ebfc8d483ed75a6b298506cb949",
		},
		{
			name: "valid token 2",
			args: args{apiToken: "sgp_abcdef0123456789abcdef0123456789abcdef01"},
			want: "sgd_3b0ed67f378c6f62c1b17faa738f46f2889beb6856eb5432e7d1aedd0953415b",
		},
		{
			name: "valid token 2b",
			args: args{apiToken: "sgph_abcdef0123456789_abcdef0123456789abcdef0123456789abcdef01"},
			want: "sgd_3b0ed67f378c6f62c1b17faa738f46f2889beb6856eb5432e7d1aedd0953415b",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := GenerateDotcomUserGatewayAccessToken(tt.args.apiToken); got != tt.want {
				if err != nil {
					t.Errorf("GenerateDotcomUserGatewayAccessToken() returned error: %s", err)
				}
				t.Errorf("GenerateDotcomUserGatewayAccessToken() = %v, want %v", got, tt.want)
			}
		})
	}
}
