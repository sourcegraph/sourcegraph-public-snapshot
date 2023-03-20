package run

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/apiclient"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestValidateSrcCLIVersion(t *testing.T) {
	execCommand = fakeExecCommand
	defer func() { execCommand = exec.CommandContext }()

	tests := []struct {
		name           string
		latestVersion  string
		currentVersion string
		expectedErr    error
		isSrcPatchErr  bool
	}{
		{
			name:           "Matches",
			latestVersion:  "1.2.3",
			currentVersion: "1.2.3",
		},
		{
			name:           "Current patch behind",
			latestVersion:  "1.2.3",
			currentVersion: "1.2.2",
			expectedErr:    errors.New("consider upgrading actual=1.2.2, latest=1.2.3: installed src-cli is not the latest version"),
			isSrcPatchErr:  true,
		},
		{
			name:           "Latest patch behind",
			latestVersion:  "1.2.2",
			currentVersion: "1.2.3",
		},
		{
			name:           "Current minor behind",
			latestVersion:  "1.2.3",
			currentVersion: "1.1.0",
			expectedErr:    errors.New("installed src-cli is not the recommended version, consider switching actual=1.1.0, recommended=1.2.3"),
			isSrcPatchErr:  false,
		},
		{
			name:           "Latest minor behind",
			latestVersion:  "1.1.0",
			currentVersion: "1.2.0",
			expectedErr:    errors.New("installed src-cli is not the recommended version, consider switching actual=1.2.0, recommended=1.1.0"),
			isSrcPatchErr:  false,
		},
		{
			name:           "Current major behind",
			latestVersion:  "2.0.0",
			currentVersion: "1.0.0",
			expectedErr:    errors.New("installed src-cli is not the recommended version, consider switching actual=1.0.0, recommended=2.0.0"),
			isSrcPatchErr:  false,
		},
		{
			name:           "Latest major behind",
			latestVersion:  "1.0.0",
			currentVersion: "2.0.0",
			expectedErr:    errors.New("installed src-cli is not the recommended version, consider switching actual=2.0.0, recommended=1.0.0"),
			isSrcPatchErr:  false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server, client := newTestServerAndClient(t, func(w http.ResponseWriter, r *http.Request) {
				err := json.NewEncoder(w).Encode(struct {
					Version string `json:"version"`
				}{test.latestVersion})
				require.NoError(t, err)
			})
			defer server.Close()

			mockedStdout = fmt.Sprintf("Current version: %s", test.currentVersion)

			err := validateSrcCLIVersion(context.Background(), client, apiclient.EndpointOptions{URL: server.URL})
			if test.expectedErr != nil {
				assert.NotNil(t, err)
				assert.Equal(t, errors.Is(err, ErrSrcPatchBehind), test.isSrcPatchErr)
				assert.Equal(t, test.expectedErr.Error(), err.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestValidateAuthorizationToken(t *testing.T) {
	tests := []struct {
		name                string
		statusCode          int
		expectedErr         error
		isUnauthorizedError bool
	}{
		{
			name:       "Valid response",
			statusCode: http.StatusOK,
		},
		{
			name:                "Unauthorized",
			statusCode:          http.StatusUnauthorized,
			expectedErr:         AuthorizationFailedErr,
			isUnauthorizedError: true,
		},
		{
			name:                "Internal server error",
			statusCode:          http.StatusInternalServerError,
			expectedErr:         errors.New("failed to validate authorization token: unexpected status code 500"),
			isUnauthorizedError: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server, client := newTestServerAndClient(t, func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(test.statusCode)
			})
			defer server.Close()

			err := validateAuthorizationToken(context.Background(), client, apiclient.EndpointOptions{URL: server.URL})
			if test.expectedErr != nil {
				assert.NotNil(t, err)
				assert.Equal(t, errors.Is(err, AuthorizationFailedErr), test.isUnauthorizedError)
				assert.Equal(t, test.expectedErr.Error(), err.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

// mockedExitStatus is set the exit status of the exec.Command.
var mockedExitStatus = 0

// mockedStdout is set the stdout of the exec.Command.
var mockedStdout string

// fakeExecCommand returns an exec.CommandContext configured to call TestExecCommandHelper with the exit status set to
// mockedExitStatus and stdout set to mockedStdout.
func fakeExecCommand(ctx context.Context, command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestExecCommandHelper", "--", command}
	cs = append(cs, args...)
	cmd := exec.CommandContext(ctx, os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1",
		fmt.Sprintf("STDOUT=%s", mockedStdout),
		fmt.Sprintf("EXIT_STATUS=%d", mockedExitStatus)}
	return cmd
}

// TestExecCommandHelper a fake test that fakeExecCommand will run instead of calling the actual exec.CommandContext.
func TestExecCommandHelper(t *testing.T) {
	// Since this function must be big T test. We don't want to actually test anything. So if GO_WANT_HELPER_PROCESS
	// is not set, just exit right away.
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	_, err := fmt.Fprint(os.Stdout, os.Getenv("STDOUT"))
	require.NoError(t, err)

	i, err := strconv.Atoi(os.Getenv("EXIT_STATUS"))
	require.NoError(t, err)

	os.Exit(i)
}

func newTestServerAndClient(t *testing.T, handlerFunc func(w http.ResponseWriter, r *http.Request)) (*httptest.Server, *apiclient.BaseClient) {
	server := httptest.NewServer(http.HandlerFunc(handlerFunc))
	client, err := apiclient.NewBaseClient(apiclient.BaseClientOptions{
		EndpointOptions: apiclient.EndpointOptions{
			URL: server.URL,
		},
	})
	require.NoError(t, err)

	return server, client
}
