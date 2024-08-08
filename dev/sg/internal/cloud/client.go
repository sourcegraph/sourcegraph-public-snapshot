package cloud

import (
	"context"
	"net/http"
	"strings"
	"time"

	"connectrpc.com/connect"
	cloudapiv1 "github.com/sourcegraph/cloud-api/go/cloudapi/v1"
	"github.com/sourcegraph/cloud-api/go/cloudapi/v1/cloudapiv1connect"
	"github.com/sourcegraph/run"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

// ErrInstanceNotFound is returned when an instance is not found.
var ErrInstanceNotFound error = errors.New("instance not found")

// HeaderUserToken is the header name for the user token when communicating with the Cloud API.
const HeaderUserToken = "X-GCP-User-Token"

// APIEndpoint is the endpoint where Cloud API is running.
const APIEndpoint = "https://cloud-ops-dev.sgdev.org/api"

// DevEnvironment is the environment where Cloud allows ephemeral instance types
const DevEnvironment = "dev"

var _ EphemeralClient = &Client{}

type EphemeralClient interface {
	CreateInstance(context.Context, *DeploymentSpec) (*Instance, error)
	GetInstance(context.Context, string) (*Instance, error)
	ListInstances(context.Context, bool) ([]*Instance, error)
	DeleteInstance(context.Context, string) error
}

type Client struct {
	client cloudapiv1connect.InstanceServiceClient
	token  string
	email  string
}

type DeploymentSpec struct {
	Name             string
	Version          string
	License          string
	InstanceFeatures map[string]string
}

func NewDeploymentSpec(name, version, license string) *DeploymentSpec {
	features := newInstanceFeatures()
	features.SetEphemeralInstance(true)
	return &DeploymentSpec{
		Name:             name,
		Version:          version,
		License:          license,
		InstanceFeatures: features.Value(),
	}
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

func (c *Client) GetInstance(ctx context.Context, name string) (*Instance, error) {
	req := newRequestWithToken(c.token, &cloudapiv1.GetInstanceRequest{
		Name:        name,
		Environment: DevEnvironment,
	},
	)

	resp, err := c.client.GetInstance(ctx, req)
	if err != nil {
		// the error received doesn't unpack properly into grpc Status or connErr, so for now we just check that the
		// string representation contains "not found" for the instance
		if strings.Contains(err.Error(), "not found") {
			return nil, ErrInstanceNotFound
		}
		return nil, errors.Wrapf(err, "failed to get instance %q", name)
	}

	return newInstance(resp.Msg.GetInstance())
}

func (c *Client) ListInstances(ctx context.Context, all bool) ([]*Instance, error) {
	listReq := cloudapiv1.ListInstancesRequest{
		InstanceFilter: &cloudapiv1.InstanceFilter{
			Environment: pointers.Ptr(DevEnvironment),
		},
	}
	if !all {
		listReq.InstanceFilter.AdminEmail = &c.email
	}

	req := newRequestWithToken(c.token, &listReq)
	resp, err := c.client.ListInstances(
		ctx,
		req,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list instances")
	}

	return toInstances(resp.Msg.GetInstances()...)
}

func (c *Client) CreateInstance(ctx context.Context, spec *DeploymentSpec) (*Instance, error) {
	licenseKey := spec.License
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

	return newInstance(resp.Msg.GetInstance())
}

func (c *Client) UpgradeInstance(ctx context.Context, spec *DeploymentSpec) (*Instance, error) {
	req := newRequestWithToken(c.token, &cloudapiv1.UpdateInstanceVersionRequest{
		Name:        spec.Name,
		Environment: DevEnvironment,
		Version:     spec.Version,
	})
	resp, err := c.client.UpdateInstanceVersion(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to upgrade instance")
	}

	return newInstance(resp.Msg.GetInstance())
}

func (c *Client) DeleteInstance(ctx context.Context, name string) error {
	return nil
}

func (c *Client) ExtendLease(ctx context.Context, name string, extendTime time.Time) (*Instance, error) {
	req := newRequestWithToken(c.token, &cloudapiv1.UpdateInstanceLeaseRequest{
		Name:        name,
		Environment: DevEnvironment,
		Lease: &timestamppb.Timestamp{
			Seconds: extendTime.Unix(),
			Nanos:   0,
		},
	})

	resp, err := c.client.UpdateInstanceLease(ctx, req)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to extend lease for instance %q", name)
	}

	return newInstance(resp.Msg.GetInstance())
}
