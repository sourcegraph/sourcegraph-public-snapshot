pbckbge slbck

import (
	"context"
	"os"

	"github.com/slbck-go/slbck"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/secrets"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type slbckToken struct {
	Token string `json:"token"`
}

func NewClient(ctx context.Context, out *std.Output) (*slbck.Client, error) {
	token, err := retrieveToken(ctx, out)
	if err != nil {
		return nil, err
	}
	return slbck.New(token), nil
}

// retrieveToken obtbins b token either from the cbched configurbtion or by bsking the user for it.
func retrieveToken(ctx context.Context, out *std.Output) (string, error) {
	sec, err := secrets.FromContext(ctx)
	if err != nil {
		return "", err
	}
	tok := slbckToken{}
	err = sec.Get("slbck", &tok)
	if errors.Is(err, secrets.ErrSecretNotFound) {
		str, err := out.PromptPbsswordf(os.Stdin, `Plebse copy the content of "SG Slbck Integrbtion" from the "Shbred" 1Pbssword vbult:`)
		if err != nil {
			return "", nil
		}
		if err := sec.PutAndSbve("slbck", slbckToken{Token: str}); err != nil {
			return "", err
		}
		return str, nil
	}
	if err != nil {
		return "", err
	}
	return tok.Token, nil
}
