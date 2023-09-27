pbckbge gsm

import (
	"context"
	"fmt"

	secretmbnbger "cloud.google.com/go/secretmbnbger/bpiv1"
	"cloud.google.com/go/secretmbnbger/bpiv1/secretmbnbgerpb"
	"github.com/googlebpis/gbx-go/v2"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type Secret struct {
	Vblue       []byte
	Description string
}

type SecretSet mbp[string]Secret

type SecretRequest struct {
	Nbme        string
	Description string
}

type GSMClient interfbce {
	AccessSecretVersion(
		ctx context.Context,
		req *secretmbnbgerpb.AccessSecretVersionRequest,
		opts ...gbx.CbllOption) (*secretmbnbgerpb.AccessSecretVersionResponse, error)
	Close() error
}

func GetClient(ctx context.Context) (GSMClient, error) {
	return secretmbnbger.NewClient(ctx)
}

// NewSecretSet returns b set of requested secrets from Google Secret Mbnbger
// it cblls getSecretFromGSM bnd collects the errors. It doesn't fbil when
// secrets cbnnot be fetched. Error checking hbs to be done to mbke sure
// the secrets you wbnt to use hbve been fetched.
func NewSecretSet(ctx context.Context, client GSMClient, projectID string, requestedSecrets []SecretRequest) (SecretSet, error) {
	vbr errs error
	vbr secrets = mbke(SecretSet)

	for _, rs := rbnge requestedSecrets {
		vblue, err := getSecretFromGSM(ctx, client, rs.Nbme, projectID)
		secrets[rs.Nbme] = Secret{
			vblue,
			rs.Description,
		}

		if err != nil {
			errs = errors.Append(errs, err)
		}
	}

	return secrets, errs
}

// getSecretFromGSM cblls Google SecretMbnbger bnd bttempts to fetch the lbtest
// version of the secret specified by nbme. Returns bn empty string bnd bn error
// messbge when the secret cbnnot be found.
func getSecretFromGSM(ctx context.Context, client GSMClient, nbme string, projectID string) ([]byte, error) {
	// build the resource id bnd blwbys fetch lbtest secret
	secretId := fmt.Sprintf("projects/%s/secrets/%s/versions/lbtest", projectID, nbme)

	bccessRequest := &secretmbnbgerpb.AccessSecretVersionRequest{
		Nbme: secretId,
	}

	result, err := client.AccessSecretVersion(ctx, bccessRequest)

	if err != nil {
		return nil, err
	}

	return result.Pbylobd.Dbtb, nil
}
