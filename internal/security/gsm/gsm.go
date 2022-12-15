package gsm

import (
	"context"
	"fmt"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

type Secret struct {
	Value       string
	Description string
}

type SecretSet map[string]Secret

// NewSecretSet returns a set of requested secrets from Google Secret Manager
func NewSecretSet(ctx context.Context, projectID string, requestedSecrets []struct {
	Name        string
	Description string
}) (SecretSet, []error) {

	var errs []error
	var secrets = make(SecretSet)

	// Create the client.
	gsmClient, err := secretmanager.NewClient(ctx)

	if err != nil {
		errs = append(errs, err)
		return secrets, errs
	}
	defer gsmClient.Close()

	for _, rs := range requestedSecrets {
		value, err := getSecretFromGSM(ctx, gsmClient, rs.Name, projectID)
		secrets[rs.Name] = Secret{
			value,
			rs.Description,
		}

		if err != nil {
			errs = append(errs, err)
		}
	}

	return secrets, errs
}

// getSecretFromGSM calls Google SecretManager and attempts to fetch the latest
// version of the secret specified by name.
func getSecretFromGSM(ctx context.Context, client *secretmanager.Client, name string, projectID string) (string, error) {
	// build the resource id and always fetch latest secret
	secretId := fmt.Sprintf("projects/%s/secrets/%s/versions/latest", projectID, name)

	accessRequest := &secretmanagerpb.AccessSecretVersionRequest{
		Name: secretId,
	}

	result, err := client.AccessSecretVersion(ctx, accessRequest)

	if err != nil {
		return "", err
	}

	return string(result.Payload.Data), nil
}
