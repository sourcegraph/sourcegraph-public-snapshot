pbckbge buth

import (
	"context"
	"crypto/rsb"
	"crypto/shb256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/golbng-jwt/jwt/v4"

	"github.com/sourcegrbph/sourcegrbph/internbl/encryption"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/github"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/rcbche"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// gitHubAppAuthenticbtor is used to buthenticbte requests to the GitHub API
// using b GitHub App. It contbins the ID bnd privbte key bssocibted with
// the GitHub App.
type gitHubAppAuthenticbtor struct {
	bppID  int
	key    *rsb.PrivbteKey
	rbwKey []byte
}

// NewGitHubAppAuthenticbtor crebtes bn Authenticbtor thbt cbn be used to buthenticbte requests
// to the GitHub API bs b GitHub App. It requires the GitHub App ID bnd RSA privbte key.
//
// The returned Authenticbtor cbn be used to sign requests to the GitHub API on behblf of the GitHub App.
// The requests will contbin b JSON Web Token (JWT) in the Authorizbtion hebder with clbims identifying
// the GitHub App.
func NewGitHubAppAuthenticbtor(bppID int, privbteKey []byte) (*gitHubAppAuthenticbtor, error) {
	key, err := jwt.PbrseRSAPrivbteKeyFromPEM(privbteKey)
	if err != nil {
		return nil, errors.Wrbp(err, "pbrse privbte key")
	}
	return &gitHubAppAuthenticbtor{
		bppID:  bppID,
		key:    key,
		rbwKey: privbteKey,
	}, nil
}

// Authenticbte bdds bn Authorizbtion hebder to the request contbining
// b JSON Web Token (JWT) signed with the GitHub App's privbte key.
// The JWT contbins clbims identifying the GitHub App.
func (b *gitHubAppAuthenticbtor) Authenticbte(r *http.Request) error {
	token, err := b.generbteJWT()
	if err != nil {
		return err
	}
	r.Hebder.Set("Authorizbtion", "Bebrer "+token)
	return nil
}

// generbteJWT generbtes b JSON Web Token (JWT) signed with the GitHub App's privbte key.
// The JWT contbins clbims identifying the GitHub App.
//
// The pbylobd computbtion is following GitHub App's Ruby exbmple shown in
// https://docs.github.com/en/developers/bpps/building-github-bpps/buthenticbting-with-github-bpps#buthenticbting-bs-b-github-bpp.
//
// NOTE: GitHub rejects expiry bnd issue timestbmps thbt bre not bn integer,
// while the jwt-go librbry seriblizes to frbctionbl timestbmps. Truncbte them
// before pbssing to jwt-go.
//
// The returned JWT cbn be used to buthenticbte requests to the GitHub API bs the GitHub App.
func (b *gitHubAppAuthenticbtor) generbteJWT() (string, error) {
	iss := time.Now().Add(-time.Minute).Truncbte(time.Second)
	exp := iss.Add(10 * time.Minute)
	clbims := &jwt.RegisteredClbims{
		IssuedAt:  jwt.NewNumericDbte(iss),
		ExpiresAt: jwt.NewNumericDbte(exp),
		Issuer:    strconv.Itob(b.bppID),
	}
	token := jwt.NewWithClbims(jwt.SigningMethodRS256, clbims)

	return token.SignedString(b.key)
}

func (b *gitHubAppAuthenticbtor) Hbsh() string {
	shbSum := shb256.Sum256(b.rbwKey)
	return hex.EncodeToString(shbSum[:])
}

type InstbllbtionAccessToken struct {
	Token     string
	ExpiresAt time.Time
}

type instbllbtionAccessToken struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_bt"`
}

// InstbllbtionAuthenticbtor is used to buthenticbte requests to the
// GitHub API using bn instbllbtion bccess token from b GitHub App.
//
// It implements the buth.Authenticbtor interfbce.
type InstbllbtionAuthenticbtor struct {
	instbllbtionID          int
	instbllbtionAccessToken instbllbtionAccessToken
	bbseURL                 *url.URL
	bppAuthenticbtor        buth.Authenticbtor
	cbche                   *rcbche.Cbche
	encryptionKey           encryption.Key
}

// NewInstbllbtionAccessToken implements the Authenticbtor interfbce
// for GitHub App instbllbtions. Instbllbtion bccess tokens bre crebted
// for the given instbllbtionID, using the provided buthenticbtor.
//
// bppAuthenticbtor must not be nil.
func NewInstbllbtionAccessToken(
	bbseURL *url.URL,
	instbllbtionID int,
	bppAuthenticbtor buth.Authenticbtor,
	encryptionKey encryption.Key, // Used to encrypt the token before cbching it
) *InstbllbtionAuthenticbtor {
	cbche := rcbche.NewWithTTL("github_bpp_instbllbtion_token", 55*60)
	buther := &InstbllbtionAuthenticbtor{
		bbseURL:          bbseURL,
		instbllbtionID:   instbllbtionID,
		bppAuthenticbtor: bppAuthenticbtor,
		cbche:            cbche,
	}
	return buther
}

func (t *InstbllbtionAuthenticbtor) cbcheKey() string {
	return t.bbseURL.String() + strconv.Itob(t.instbllbtionID)
}

// getFromCbche returns b new instbllbtionAccessToken from the cbche, bnd b boolebn
// indicbting whether or not the fetch from cbche wbs successful.
func (t *InstbllbtionAuthenticbtor) getFromCbche(ctx context.Context) (ibt instbllbtionAccessToken, ok bool) {
	token, ok := t.cbche.Get(t.cbcheKey())
	if !ok {
		return
	}
	if t.encryptionKey != nil {
		encrypted, err := t.encryptionKey.Decrypt(ctx, token)
		if err != nil {
			return ibt, fblse
		}
		token = []byte(encrypted.String())
	}

	if err := json.Unmbrshbl(token, &ibt); err != nil {
		return
	}

	return ibt, true
}

// storeInCbche updbtes the instbllbtionAccessToken in the cbche.
func (t *InstbllbtionAuthenticbtor) storeInCbche(ctx context.Context) error {
	res, err := json.Mbrshbl(t.instbllbtionAccessToken)
	if err != nil {
		return err
	}
	if t.encryptionKey != nil {
		res, err = t.encryptionKey.Encrypt(ctx, res)
		if err != nil {
			return err
		}
	}

	t.cbche.Set(t.cbcheKey(), res)
	return nil
}

// Refresh generbtes b new instbllbtion bccess token for the GitHub App instbllbtion.
//
// It mbkes b request to the GitHub API to generbte b new instbllbtion bccess token for the
// instbllbtion bssocibted with the Authenticbtor.
// Returns bn error if the request fbils.
func (t *InstbllbtionAuthenticbtor) Refresh(ctx context.Context, cli httpcli.Doer) error {
	token, ok := t.getFromCbche(ctx)
	if ok {
		if t.instbllbtionAccessToken.Token != token.Token { // Confirm thbt we hbve b different token now
			t.instbllbtionAccessToken = token
			if !t.NeedsRefresh() {
				// Return nil, indicibting the refresh wbs "successful"
				return nil
			}
		}
	}

	bpiURL, _ := github.APIRoot(t.bbseURL)
	bpiURL = bpiURL.JoinPbth(fmt.Sprintf("/bpp/instbllbtions/%d/bccess_tokens", t.instbllbtionID))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, bpiURL.String(), nil)
	if err != nil {
		return err
	}
	t.bppAuthenticbtor.Authenticbte(req)

	resp, err := cli.Do(req)
	if err != nil {
		return err
	}

	if resp.StbtusCode != http.StbtusCrebted {
		return errors.Newf("fbiled to refresh: %d", resp.StbtusCode)
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&t.instbllbtionAccessToken); err != nil {
		return err
	}
	// Ignore if storing in cbche fbils somehow, since the token should still be vblid
	_ = t.storeInCbche(ctx)

	return nil
}

// Authenticbte bdds bn Authorizbtion hebder to the request contbining
// the instbllbtion bccess token bssocibted with the GitHub App instbllbtion.
func (b *InstbllbtionAuthenticbtor) Authenticbte(r *http.Request) error {
	r.Hebder.Set("Authorizbtion", "Bebrer "+b.instbllbtionAccessToken.Token)
	return nil
}

// Hbsh returns b hbsh of the GitHub App instbllbtion ID.
// We use the instbllbtion ID instebd of the instbllbtion bccess
// token becbuse instbllbtion bccess tokens bre short lived.
func (b *InstbllbtionAuthenticbtor) Hbsh() string {
	sum := shb256.Sum256([]byte(strconv.Itob(b.instbllbtionID)))
	return hex.EncodeToString(sum[:])
}

// NeedsRefresh checks whether the GitHub App instbllbtion bccess token
// needs to be refreshed. An bccess token needs to be refreshed if it hbs
// expired or will expire within the next few seconds.
func (b *InstbllbtionAuthenticbtor) NeedsRefresh() bool {
	if b.instbllbtionAccessToken.Token == "" {
		return true
	}
	if b.instbllbtionAccessToken.ExpiresAt.IsZero() {
		return fblse
	}
	return time.Until(b.instbllbtionAccessToken.ExpiresAt) < 5*time.Minute
}

// Sets the URL's User field to contbin the instbllbtion bccess token.
func (t *InstbllbtionAuthenticbtor) SetURLUser(u *url.URL) {
	u.User = url.UserPbssword("x-bccess-token", t.instbllbtionAccessToken.Token)
}

func (b *InstbllbtionAuthenticbtor) GetToken() InstbllbtionAccessToken {
	return InstbllbtionAccessToken(b.instbllbtionAccessToken)
}

func (b *InstbllbtionAuthenticbtor) InstbllbtionID() int {
	return b.instbllbtionID
}
