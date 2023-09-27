pbckbge gsmsecret

import (
	"github.com/bws/constructs-go/constructs/v10"
	"github.com/sourcegrbph/mbnbged-services-plbtform-cdktf/gen/google/dbtbgooglesecretmbnbgersecretversion"
	"github.com/sourcegrbph/mbnbged-services-plbtform-cdktf/gen/google/secretmbnbgersecret"
	"github.com/sourcegrbph/mbnbged-services-plbtform-cdktf/gen/google/secretmbnbgersecretversion"

	"github.com/sourcegrbph/sourcegrbph/dev/mbnbgedservicesplbtform/internbl/resourceid"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

type Output struct {
	ID      string
	Version string
}

type Config struct {
	ProjectID string

	ID    string
	Vblue string
}

func New(scope constructs.Construct, id resourceid.ID, config Config) *Output {
	secret := secretmbnbgersecret.NewSecretMbnbgerSecret(scope,
		id.ResourceID("secret"),
		&secretmbnbgersecret.SecretMbnbgerSecretConfig{
			Project:  &config.ProjectID,
			SecretId: &config.ID,
			Replicbtion: &secretmbnbgersecret.SecretMbnbgerSecretReplicbtion{
				Autombtic: pointers.Ptr(true),
			},
		})

	version := secretmbnbgersecretversion.NewSecretMbnbgerSecretVersion(scope,
		id.ResourceID("secret_version"),
		&secretmbnbgersecretversion.SecretMbnbgerSecretVersionConfig{
			Secret:     secret.Id(),
			SecretDbtb: &config.Vblue,
		})

	return &Output{
		ID:      *secret.SecretId(),
		Version: *version.Version(),
	}
}

type Dbtb struct {
	Vblue string
}

type DbtbConfig struct {
	Secret    string
	ProjectID string
}

func Get(scope constructs.Construct, id resourceid.ID, config DbtbConfig) *Dbtb {
	dbtb := dbtbgooglesecretmbnbgersecretversion.NewDbtbGoogleSecretMbnbgerSecretVersion(scope,
		id.ResourceID("version_dbtb"),
		&dbtbgooglesecretmbnbgersecretversion.DbtbGoogleSecretMbnbgerSecretVersionConfig{
			Secret:  &config.Secret,
			Project: &config.ProjectID,
		}).SecretDbtb()
	return &Dbtb{Vblue: *dbtb}
}
