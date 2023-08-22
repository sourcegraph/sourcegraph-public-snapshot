package gsmsecret

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/project"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/secretmanagersecret"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/secretmanagersecretversion"

	"github.com/sourcegraph/sourcegraph/internal/pointer"
)

type Output struct {
	ID      string
	Version string
}

type Config struct {
	Project project.Project

	ID    string
	Value string
}

func New(scope constructs.Construct, id string, config Config) *Output {
	secret := secretmanagersecret.NewSecretManagerSecret(scope,
		pointer.Stringf("%s-secret", id),
		&secretmanagersecret.SecretManagerSecretConfig{
			Project:  config.Project.ProjectId(),
			SecretId: &config.ID,
			Replication: &secretmanagersecret.SecretManagerSecretReplication{
				Automatic: pointer.Value(true),
			},
		})

	version := secretmanagersecretversion.NewSecretManagerSecretVersion(scope,
		pointer.Stringf("%s-version", id),
		&secretmanagersecretversion.SecretManagerSecretVersionConfig{
			Secret:     secret.Id(),
			SecretData: &config.Value,
		})

	return &Output{
		ID:      *secret.SecretId(),
		Version: *version.Version(),
	}
}
