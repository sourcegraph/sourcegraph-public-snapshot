package sgx

import "testing"

func TestServeCmd_configureAppURL(t *testing.T) {
	tests := []struct {
		flags      *ServeCmd
		wantAppURL string
	}{
		{
			flags:      &ServeCmd{HTTPAddr: ":8080"},
			wantAppURL: "http://localhost:8080",
		},
		{
			flags:      &ServeCmd{HTTPAddr: "myhost:8080"},
			wantAppURL: "http://myhost:8080",
		},
		{
			flags:      &ServeCmd{HTTPAddr: ":8080", AppURL: "http://example.com:1234"},
			wantAppURL: "http://example.com:1234",
		},
		{
			flags:      &ServeCmd{HTTPAddr: "other.example.com:8080", AppURL: "http://example.com:1234"},
			wantAppURL: "http://example.com:1234",
		},
	}
	for _, test := range tests {
		appURL, err := test.flags.configureAppURL()
		if err != nil {
			t.Error(err)
			continue
		}
		if appURL.String() != test.wantAppURL {
			t.Errorf("got %q, want %q", appURL, test.wantAppURL)
		}
	}
}
