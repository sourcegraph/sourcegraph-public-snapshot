package gsm

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/googleapis/gax-go/v2"
	"github.com/stretchr/testify/assert"
)

type MockClient struct {
	AccessFunc func(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.AccessSecretVersionResponse, error)
	CloseFunc  func() error
}

func (m *MockClient) AccessSecretVersion(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.AccessSecretVersionResponse, error) {
	return m.AccessFunc(ctx, req, opts...)
}

func (m *MockClient) Close() error {
	return m.CloseFunc()
}

func TestFetchGSM(t *testing.T) {

	testcases := []struct {
		name     string
		client   *MockClient
		project  string
		secret   string
		value    string
		pass     bool
		contains string
	}{
		{
			name: "Test cannot find secret returns empty secret",
			client: &MockClient{
				AccessFunc: func(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.AccessSecretVersionResponse, error) {
					return nil, errors.New(fmt.Sprintf("rpc error: code = NotFound desc = Secret [%s] not found or has no versions", req.Name))
				},
				CloseFunc: func() error { return nil },
			},
			project:  "foo",
			secret:   "message-signing-secret",
			value:    "",
			pass:     false,
			contains: fmt.Sprintf("projects/foo/secrets/message-signing-secret/versions/latest] not found"),
		},
	}

	ctx := context.Background()

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			Client = tc.client
			secret, err := getSecretFromGSM(ctx, tc.secret, tc.project)

			assert.Equal(t, secret, tc.value)
			if !tc.pass {
				assert.ErrorContains(t, err, tc.contains)
			}
		})
	}

}
