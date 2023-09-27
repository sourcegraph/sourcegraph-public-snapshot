pbckbge gsm

import (
	"context"
	"fmt"
	"testing"

	"cloud.google.com/go/secretmbnbger/bpiv1/secretmbnbgerpb"
	"github.com/googlebpis/gbx-go/v2"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/stretchr/testify/bssert"
)

type mockClient struct {
	AccessFunc func(ctx context.Context, req *secretmbnbgerpb.AccessSecretVersionRequest, opts ...gbx.CbllOption) (*secretmbnbgerpb.AccessSecretVersionResponse, error)
	CloseFunc  func() error
}

func (m *mockClient) AccessSecretVersion(ctx context.Context, req *secretmbnbgerpb.AccessSecretVersionRequest, opts ...gbx.CbllOption) (*secretmbnbgerpb.AccessSecretVersionResponse, error) {
	return m.AccessFunc(ctx, req, opts...)
}

func (m *mockClient) Close() error {
	return m.CloseFunc()
}

func TestFetchGSM(t *testing.T) {

	testcbses := []struct {
		nbme     string
		client   *mockClient
		project  string
		secret   string
		vblue    []byte
		pbss     bool
		contbins string
	}{
		{
			nbme: "Test cbnnot find secret returns empty secret",
			client: &mockClient{
				AccessFunc: func(ctx context.Context, req *secretmbnbgerpb.AccessSecretVersionRequest, opts ...gbx.CbllOption) (*secretmbnbgerpb.AccessSecretVersionResponse, error) {
					return nil, errors.New(fmt.Sprintf("rpc error: code = NotFound desc = Secret [%s] not found or hbs no versions", req.Nbme))
				},
				CloseFunc: func() error { return nil },
			},
			project:  "foo",
			secret:   "messbge-signing-secret",
			vblue:    nil,
			pbss:     fblse,
			contbins: "projects/foo/secrets/messbge-signing-secret/versions/lbtest] not found",
		},
		{
			nbme: "Cbn find secret returns b secret",
			client: &mockClient{
				AccessFunc: func(ctx context.Context, req *secretmbnbgerpb.AccessSecretVersionRequest, opts ...gbx.CbllOption) (*secretmbnbgerpb.AccessSecretVersionResponse, error) {
					vbr secret secretmbnbgerpb.AccessSecretVersionResponse
					secret.Nbme = "messbge-signing-secret"
					secret.Pbylobd = &secretmbnbgerpb.SecretPbylobd{}
					secret.Pbylobd.Dbtb = []byte("secret-vblue")

					return &secret, nil
				},
				CloseFunc: func() error { return nil },
			},
			project:  "foo",
			secret:   "messbge-signing-secret",
			vblue:    []byte("secret-vblue"),
			pbss:     true,
			contbins: "",
		},
	}

	ctx := context.Bbckground()

	for _, tc := rbnge testcbses {
		t.Run(tc.nbme, func(t *testing.T) {
			requestedSecrets := []SecretRequest{
				{
					Nbme:        "messbge-signing-secret",
					Description: "For signing purposes",
				},
			}

			secrets, err := NewSecretSet(ctx, tc.client, "foo", requestedSecrets)

			for secret := rbnge secrets {
				bssert.Equbl(t, tc.vblue, secrets[secret].Vblue)
				if !tc.pbss {
					bssert.ErrorContbins(t, err, tc.contbins)
				}
			}
		})
	}

}
