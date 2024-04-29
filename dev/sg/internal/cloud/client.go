package cloud

import (
	"context"
	"net/http"
	"os"
	"strings"

	"connectrpc.com/connect"
	cloudapiv1 "github.com/sourcegraph/cloud-api/go/cloudapi/v1"
	"github.com/sourcegraph/cloud-api/go/cloudapi/v1/cloudapiv1connect"
	"github.com/sourcegraph/run"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

// HeaderUserToken is the header name for the user token when communicating with the Cloud API.
const HeaderUserToken = "X-GCP-User-Token"

// APIEndpoint is the endpoint where Cloud API is running.
const APIEndpoint = "https://cloud-ops-dev.sgdev.org/api"

// DevEnvironment is the environment where Cloud allows ephemeral instance types
const DevEnvironment = "dev"

// EphemeralInstanceType is the instace type we should use when creating an instance with the cloud API.
// It is set to internal because in cloud, internal instance types does not have metrics or security enabled.
const EphemeralInstanceType = "internal"

type Client struct {
	client cloudapiv1connect.InstanceServiceClient
	token  string
	email  string
}

type DeploymentSpec struct {
	Name             string
	Version          string
	InstanceFeatures map[string]string
}

func NewDeploymentSpec(name, version string) *DeploymentSpec {
	return &DeploymentSpec{
		Name:    name,
		Version: version,
		InstanceFeatures: map[string]string{
			"ephemeral": "true", // need to have this to make the instance ephemeral
		},
	}
}

func GetGCloudAccount(ctx context.Context) (string, error) {
	return run.Cmd(ctx, "gcloud config get account").Run().String()
}

func NewClient(ctx context.Context, email, endpoint string) (*Client, error) {
	// have to use IDENTITY token not ACCESS token!
	token, err := run.Cmd(ctx, "gcloud auth print-identity-token").Run().String()
	if err != nil {
		return nil, errors.Newf("failed to get gcloud auth token: %v", err)
	}

	client := cloudapiv1connect.NewInstanceServiceClient(
		http.DefaultClient,
		endpoint,
	)

	return &Client{client: client, email: email, token: token}, nil
}

func validateEmail(email string) error {
	if len(email) == 0 {
		return errors.New("gcloud account email is empty")
	}

	if !strings.Contains(email, "@sourcegraph.com") {
		return errors.Newf("gcloud account email %q is not a valid Sourcegraph email", email)
	}

	return nil
}

func newRequestWithToken[T any](token string, message *T) *connect.Request[T] {
	req := connect.NewRequest(message)
	req.Header().Add(HeaderUserToken, token)
	req.Header().Add("Content-Type", "application/json")
	return req
}

func (c *Client) ListInstances(ctx context.Context) ([]*Instance, error) {
	req := newRequestWithToken(c.token, &cloudapiv1.ListInstancesRequest{
		InstanceFilter: &cloudapiv1.InstanceFilter{
			AdminEmail: &c.email,
		},
	})
	resp, err := c.client.ListInstances(
		ctx,
		req,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list instances")
	}

	return toInstances(resp.Msg.GetInstances()...), nil
}

func (c *Client) DeployVersion(ctx context.Context, spec *DeploymentSpec) (*Instance, error) {
	// TODO(burmudar): Better method to get LicenseKeys
	licenseKey := os.Getenv("EPHEMERAL_LICENSE_KEY")
	if licenseKey == "" {
		return nil, errors.New("no license key - the env var 'EPHEMERAL_LICENSE_KEY' is empty")
	}
	req := newRequestWithToken(c.token, &cloudapiv1.CreateInstanceRequest{
		Name:    spec.Name,
		Version: &spec.Version,
		// We use internal since internal means the instance will be launched with no security, no metrics
		// this is also why the environment is set to Dev. Don't want a cloud ephemeral instance with not security in prod :)
		InstanceType:     cloudapiv1.InstanceType_INSTANCE_TYPE_INTERNAL,
		InstanceFeatures: spec.InstanceFeatures,
		Environment:      pointers.Ptr(DevEnvironment),
		AdminEmail:       &c.email,
		LicenseKey:       &licenseKey,
		GcpRegion:        pointers.Ptr("us-central1"),
	})
	resp, err := c.client.CreateInstance(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to deploy instance")
	}

	return newInstance(resp.Msg.GetInstance()), nil
}

func (c *Client) DeleteInstance(ctx context.Context, name string) error {
	return nil
}
