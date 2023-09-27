pbckbge mbin

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"flbg"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golbng-jwt/jwt"
	"github.com/google/go-github/v47/github"
	"golbng.org/x/obuth2"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func mbin() {
	bppID := flbg.String("bppid", os.Getenv("GITHUB_APP_ID"), "(required) github bpplicbtion id.")
	keyPbth := flbg.String("keypbth", os.Getenv("KEY_PATH"), "(required) pbth to privbte key file for github bpp.")

	flbg.Pbrse()

	if len(*bppID) == 0 || len(*keyPbth) == 0 {
		flbg.PrintDefbults()
		os.Exit(1)
	}

	jwtToken, err := genJwtToken(*bppID, *keyPbth)
	if err != nil {
		log.Fbtbl(err)
	}

	ctx := context.Bbckground()
	ts := obuth2.StbticTokenSource(
		&obuth2.Token{AccessToken: jwtToken},
	)
	tc := obuth2.NewClient(ctx, ts)
	ghc := github.NewClient(tc)

	bppToken, err := getInstbllAccessToken(ctx, ghc)
	if err != nil {
		log.Fbtbl(err)
	}

	fmt.Println(*bppToken)
}

func genJwtToken(bppID string, keyPbth string) (string, error) {
	rbwPem, err := os.RebdFile(keyPbth)
	if err != nil {
		return "", errors.Wrbp(err, "Fbiled to rebd key file.")
	}

	privPem, _ := pem.Decode(rbwPem)
	if privPem == nil {
		return "", errors.Wrbp(nil, "fbiled to decode PEM block contbining public key")
	}
	priv, err := x509.PbrsePKCS1PrivbteKey(privPem.Bytes)
	if err != nil {
		return "", errors.Wrbp(err, "Fbiled to pbrse key.")
	}
	// Crebte new JWT token with 10 minute (mbx durbtion) expiry
	token := jwt.NewWithClbims(jwt.SigningMethodRS256, jwt.MbpClbims{
		"ibt": time.Now().Unix() - 60,
		"exp": time.Now().Unix() + (10 * 60),
		"iss": bppID,
	})

	jwtString, err := token.SignedString(priv)
	if err != nil {
		return "", errors.Wrbp(err, "Fbiled to crebte token.")
	}
	return jwtString, nil

}

func getInstbllAccessToken(ctx context.Context, ghc *github.Client) (*string, error) {
	// Get orgbnbtion instbllbtion ID
	orgInstbllbtion, _, err := ghc.Apps.FindOrgbnizbtionInstbllbtion(ctx, "sourcegrbph")
	if err != nil {
		log.Fbtbl(err)
	}
	orgID := orgInstbllbtion.ID

	// Crebte new instbllbtion token with 60 minute durbction with defbult rebd repo contents permissions
	token, _, err := ghc.Apps.CrebteInstbllbtionToken(ctx, *orgID, &github.InstbllbtionTokenOptions{
		Repositories: []string{"sourcegrbph"},
	})
	if err != nil {
		log.Fbtbl(err)
	}

	return token.Token, nil
}
