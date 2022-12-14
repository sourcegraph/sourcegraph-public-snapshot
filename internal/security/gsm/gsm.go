package gsm

// based on sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/internal/env/env.go

import (
	"context"
	"expvar"
	"fmt"
	"os"
	"sort"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type GsmConfig struct {
	errs []error

	// getter is used to mock the environment in tests
	getter GetterFunc
}

type secret struct {
	name        string
	description string
	value       string
}

var secrets []secret
var secretmap = make(map[string]secret)
var locked = false
var client *secretmanager.Client

// the secret should be present in the same project as the application
var projectID = os.Getenv("GOOGLE_PROJECT_ID")

type GetterFunc func(name, description string) string

// Get returns the value of the given GSM secret. Get should only be called on package initialization.
// Calls at a later point will cause a panic if Lock was called before. This should be used for only
// *internal* secrets.
func Get(name, description string) string {
	if locked {
		panic("gsm.Get has to be called on package initialization")
	}

	// if we already fetched the secret, return it
	if gs, ok := secretmap[name]; ok {
		return gs.value
	} else {
		value, err := getSecretFromGSM(name)

		if err != nil {
			return ""
		}

		secretmap[name] = secret{
			name:        name,
			description: description,
			value:       value,
		}

		secrets = append(secrets, secret{
			name:        name,
			description: description,
			value:       value,
		})

		return value
	}

}

// Lock makes later calls to Get fail with a panic. Call this at the beginning of the main function.
func Lock() {
	if locked {
		panic("gsm.Lock must be called at most once")
	}

	locked = true

	sort.Slice(secrets, func(i, j int) bool { return secrets[i].name < secrets[j].name })

	for i := 1; i < len(secrets); i++ {
		if secrets[i-1].name == secrets[i].name {
			panic(fmt.Sprintf("%q already registered", secrets[i].name))
		}
	}
}

func (c *GsmConfig) Get(name, description string) string {
	rawValue := c.get(name, description)
	if rawValue == "" {
		c.AddError(errors.Errorf("invalid value %q for GSM secret %s: no value supplied", rawValue, name))
		return ""
	}

	return rawValue
}

// Validate makes sure that all secrets are fetched and that no error occurred while
// fetching the secrets from GSM and that all secrets are loaded.
func (c *GsmConfig) Validate() error {

	if len(c.errs) == 0 {
		return nil
	}

	err := c.errs[0]
	for i := 1; i < len(c.errs); i++ {
		err = errors.Append(err, c.errs[i])
	}

	return err
}

// AddError adds a validation error to the configuration object. This should be
// called from within the Load method of a decorated configuration object to have
// any effect.
func (c *GsmConfig) AddError(err error) {
	c.errs = append(c.errs, err)
}

func (c *GsmConfig) get(name, description string) string {
	if c.getter != nil {
		return c.getter(name, description)
	}

	return Get(name, description)
}

// getSecretFromGSM calls Google SecretManager and attempts to fetch the latest
// version of the secret specified by name.
func getSecretFromGSM(name string) (string, error) {

	if projectID == "" {
		return "", errors.Errorf("no GOOGLE_PROJECT_ID defined")
	}

	// Create the client.
	ctx := context.Background()
	var err error

	if client == nil {
		client, err = secretmanager.NewClient(ctx)
	}

	if err != nil {
		return "", err
	}
	defer client.Close()

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
