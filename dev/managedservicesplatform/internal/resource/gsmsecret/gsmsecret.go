package gsmsecret

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/datagooglesecretmanagersecretversion"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/secretmanagersecret"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/secretmanagersecretversion"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resourceid"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

type Output struct {
	ID      string
	Version string
}

type Config struct {
	ProjectID string

	ID    string
	Value string
}

func New(scope constructs.Construct, id resourceid.ID, config Config) *Output {
	secret := secretmanagersecret.NewSecretManagerSecret(scope,
		id.TerraformID("secret"),
		&secretmanagersecret.SecretManagerSecretConfig{
			Project:  &config.ProjectID,
			SecretId: &config.ID,
			Replication: &secretmanagersecret.SecretManagerSecretReplication{
				Automatic: pointers.Ptr(true),
			},
		})

	version := secretmanagersecretversion.NewSecretManagerSecretVersion(scope,
		id.TerraformID("secret_version"),
		&secretmanagersecretversion.SecretManagerSecretVersionConfig{
			Secret:     secret.Id(),
			SecretData: &config.Value,
		})

	return &Output{
		ID:      *secret.SecretId(),
		Version: *version.Version(),
	}
}

type Data struct {
	Value string
}

type DataConfig struct {
	Secret    string
	ProjectID string
}

func Get(scope constructs.Construct, id resourceid.ID, config DataConfig) *Data {
	data := datagooglesecretmanagersecretversion.NewDataGoogleSecretManagerSecretVersion(scope,
		id.TerraformID("version_data"),
		&datagooglesecretmanagersecretversion.DataGoogleSecretManagerSecretVersionConfig{
			Secret:  &config.Secret,
			Project: &config.ProjectID,
		}).SecretData()
	return &Data{Value: *data}
}
