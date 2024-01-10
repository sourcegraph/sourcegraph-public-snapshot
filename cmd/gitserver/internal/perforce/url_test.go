package perforce

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/vcs"
)

func TestDecomposePerforceRemoteURL(t *testing.T) {
	t.Run("not a perforce scheme", func(t *testing.T) {
		remoteURL, _ := vcs.ParseURL("https://www.google.com")
		_, _, _, _, err := DecomposePerforceRemoteURL(remoteURL)
		assert.Error(t, err)
	})

	// Tests are driven from "Examples" from the page:
	// https://www.perforce.com/manuals/cmdref/Content/CmdRef/P4PORT.html
	tests := []struct {
		cloneURL     string
		wantHost     string
		wantUsername string
		wantPassword string
		wantDepot    string
	}{
		{
			cloneURL:     "perforce://admin:password@ssl:111.222.333.444:1666//Sourcegraph/",
			wantHost:     "ssl:111.222.333.444:1666",
			wantUsername: "admin",
			wantPassword: "password",
			wantDepot:    "//Sourcegraph/",
		},
		{
			cloneURL:     "perforce://admin@ssl:111.222.333.444:1666//Sourcegraph/",
			wantHost:     "ssl:111.222.333.444:1666",
			wantUsername: "admin",
			wantDepot:    "//Sourcegraph/",
		},
		{
			cloneURL:  "perforce://ssl:111.222.333.444:1666//Sourcegraph/",
			wantHost:  "ssl:111.222.333.444:1666",
			wantDepot: "//Sourcegraph/",
		},
		{
			cloneURL: "perforce://ssl:111.222.333.444:1666",
			wantHost: "ssl:111.222.333.444:1666",
		},

		{
			cloneURL:     "perforce://admin:password@ssl6:[::]:1818ssl64:[::]:1818//Sourcegraph/",
			wantHost:     "ssl6:[::]:1818ssl64:[::]:1818",
			wantUsername: "admin",
			wantPassword: "password",
			wantDepot:    "//Sourcegraph/",
		},
		{
			cloneURL:     "perforce://admin:password@tcp6:[2001:db8::123]:1818//Sourcegraph/Cloud/",
			wantHost:     "tcp6:[2001:db8::123]:1818",
			wantUsername: "admin",
			wantPassword: "password",
			wantDepot:    "//Sourcegraph/Cloud/",
		},
	}
	for _, test := range tests {
		t.Run(test.cloneURL, func(t *testing.T) {
			remoteURL, _ := vcs.ParseURL(test.cloneURL)
			username, password, host, depot, err := DecomposePerforceRemoteURL(remoteURL)
			if err != nil {
				t.Fatal(err)
			}

			if host != test.wantHost {
				t.Fatalf("Host: want %q but got %q", test.wantHost, host)
			}
			if username != test.wantUsername {
				t.Fatalf("Username: want %q but got %q", test.wantUsername, username)
			}
			if password != test.wantPassword {
				t.Fatalf("Password: want %q but got %q", test.wantPassword, password)
			}
			if depot != test.wantDepot {
				t.Fatalf("Depot: want %q but got %q", test.wantDepot, depot)
			}
		})
	}
}
