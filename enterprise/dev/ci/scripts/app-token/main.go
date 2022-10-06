package main

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/go-github/v47/github"
	"golang.org/x/oauth2"
)

func main() {
	appID := os.Getenv("GITHUB_APP_ID")

	jwt := genJwtToken(appID)

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: jwt},
	)

	tc := oauth2.NewClient(ctx, ts)
	ghc := github.NewClient(tc)

	fmt.Println(*getInstallAccessToken(ctx, ghc))

}

func genJwtToken(appID string) string {
	rawPem, err := ioutil.ReadFile("private.pem")
	if err != nil {
		log.Fatal(err)
	}

	privPem, _ := pem.Decode(rawPem)
	priv, err := x509.ParsePKCS1PrivateKey(privPem.Bytes)
	if err != nil {
		log.Fatal(err)
	}
	// Create new JWT token with 10 minute (max duration) expiry
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iat": time.Now().Unix() - 60,
		"exp": time.Now().Unix() + (10 * 60),
		"iss": appID,
	})

	jwtString, err := token.SignedString(priv)
	if err != nil {
		log.Fatal(err)
	}

	return jwtString

}

func getInstallAccessToken(ctx context.Context, ghc *github.Client) *string {
	// Get organation installation ID
	orgInstallation, _, err := ghc.Apps.FindOrganizationInstallation(ctx, "sourcegraph")
	if err != nil {
		log.Fatal(err)
	}
	orgID := orgInstallation.ID

	// Create new installation token with 60 minute duraction with default read contents permissions
	token, _, err := ghc.Apps.CreateInstallationToken(ctx, *orgID, &github.InstallationTokenOptions{
		Repositories: []string{"sourcegraph"},
	})
	if err != nil {
		log.Fatal(err)
	}

	return token.Token

}
