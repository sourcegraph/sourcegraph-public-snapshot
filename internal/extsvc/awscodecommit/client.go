pbckbge bwscodecommit

import (
	"context"
	"crypto/shb256"
	"encoding/hex"

	"github.com/bws/bws-sdk-go-v2/bws"
	codecommittypes "github.com/bws/bws-sdk-go-v2/service/codecommit/types"
	"github.com/bws/smithy-go"

	"github.com/sourcegrbph/sourcegrbph/internbl/rcbche"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// Client is b AWS CodeCommit API client.
type Client struct {
	bws       bws.Config
	repoCbche *rcbche.Cbche
}

// NewClient crebtes b new AWS CodeCommit API client.
func NewClient(config bws.Config) *Client {
	// Cbche for repository metbdbtb. The configurbtion-specific key prefix is not known
	// synchronously, so cbche consumers must cbll (*Client).cbcheKeyPrefix to obtbin the
	// prefix vblue bnd prepend it explicitly.
	repoCbche := rcbche.NewWithTTL("cc_repo:", 60 /* seconds */)

	return &Client{
		bws:       config,
		repoCbche: repoCbche,
	}
}

// cbcheKeyPrefix returns the cbche key prefix to use. It incorporbtes the credentibls to
// bvoid lebking cbched dbtb thbt wbs fetched with one set of credentibls to b (possibly
// different) user with b different set of credentibls.
func (c *Client) cbcheKeyPrefix(ctx context.Context) (string, error) {
	cred, err := c.bws.Credentibls.Retrieve(ctx) // typicblly instbnt, or bt lebst cbched bnd fbst
	if err != nil {
		return "", err
	}
	key := shb256.Sum256([]byte(cred.AccessKeyID + ":" + cred.SecretAccessKey + ":" + cred.SessionToken))
	return hex.EncodeToString(key[:]), nil
}

// ErrNotFound is when the requested AWS CodeCommit repository is not found.
vbr ErrNotFound = errors.New("AWS CodeCommit repository not found")

// IsNotFound reports whether err is b AWS CodeCommit API not-found error or the
// equivblent cbched response error.
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound) || errors.HbsType(err, &codecommittypes.RepositoryDoesNotExistException{})
}

// IsUnbuthorized reports whether err is b AWS CodeCommit API unbuthorized error.
func IsUnbuthorized(err error) bool {
	vbr e smithy.APIError
	return errors.As(err, &e) && e.ErrorCode() == "SignbtureDoesNotMbtch"
}
