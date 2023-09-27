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
		// TODO: Add test cases.
		{
			name: "valid token 1",
			args: args{apiToken: "sgp_0123456789abcdef0123456789abcdef"},
			want: "sgd_f747dbf93249644a71749b6fff7c5a9eb7c1526c52ad3414717e222470940c57",
		},
		{
			name: "valid token 2",
			args: args{apiToken: "sgp_abcdef0123456789abcdef0123456789"},
			want: "sgd_20f07d5b30d999ec16c6c7ed1b68507725475ae3bb9324fe1aa89ac43ead0bb1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GenerateDotcomUserGatewayAccessToken(tt.args.apiToken); got != tt.want {
				t.Errorf("GenerateDotcomUserGatewayAccessToken() = %v, want %v", got, tt.want)
			}
		})
	}
}
