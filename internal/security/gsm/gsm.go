package gsm

import (
	"context"
	"fmt"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/googleapis/gax-go/v2"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Secret struct {
	Value       []byte
	Description string
}

type SecretSet map[string]Secret

type SecretRequest struct {
	Name        string
	Description string
}

type GSMClient interface {
	AccessSecretVersion(
		ctx context.Context,
		req *secretmanagerpb.AccessSecretVersionRequest,
		opts ...gax.CallOption) (*secretmanagerpb.AccessSecretVersionResponse, error)
	Close() error
}

func GetClient(ctx context.Context) (GSMClient, error) {
	return secretmanager.NewClient(ctx)
}

// NewSecretSet returns a set of requested secrets from Google Secret Manager
// it calls getSecretFromGSM and collects the errors. It doesn't fail when
// secrets cannot be fetched. Error checking has to be done to make sure
// the secrets you want to use have been fetched.
func NewSecretSet(ctx context.Context, client GSMClient, projectID string, requestedSecrets []SecretRequest) (SecretSet, error) {
	var errs error
	var secrets = make(SecretSet)

	for _, rs := range requestedSecrets {
		value, err := getSecretFromGSM(ctx, client, rs.Name, projectID)
		secrets[rs.Name] = Secret{
			value,
			rs.Description,
		}

		if err != nil {
			errs = errors.Append(errs, err)
		}
	}

	return secrets, errs
}

// getSecretFromGSM calls Google SecretManager and attempts to fetch the latest
// version of the secret specified by name. Returns an empty string and an error
// message when the secret cannot be found.
func getSecretFromGSM(ctx context.Context, client GSMClient, name string, projectID string) ([]byte, error) {
	// build the resource id and always fetch latest secret
	secretId := fmt.Sprintf("projects/%s/secrets/%s/versions/latest", projectID, name)

	accessRequest := &secretmanagerpb.AccessSecretVersionRequest{
		Name: secretId,
	}

	result, err := client.AccessSecretVersion(ctx, accessRequest)

	if err != nil {
		return nil, err
	}

	return result.Payload.Data, nil
}
