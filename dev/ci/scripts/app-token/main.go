package main

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/go-github/v55/github"
	"golang.org/x/oauth2"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func main() {
	appID := flag.String("appid", os.Getenv("GITHUB_APP_ID"), "(required) github application id.")
	keyPath := flag.String("keypath", os.Getenv("KEY_PATH"), "(required) path to private key file for github app.")

	flag.Parse()

	if len(*appID) == 0 || len(*keyPath) == 0 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	jwtToken, err := genJwtToken(*appID, *keyPath)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: jwtToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	ghc := github.NewClient(tc)

	appToken, err := getInstallAccessToken(ctx, ghc)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(*appToken)
}

func genJwtToken(appID string, keyPath string) (string, error) {
	rawPem, err := os.ReadFile(keyPath)
	if err != nil {
		return "", errors.Wrap(err, "Failed to read key file.")
	}

	privPem, _ := pem.Decode(rawPem)
	if privPem == nil {
		return "", errors.Wrap(nil, "failed to decode PEM block containing public key")
	}
	priv, err := x509.ParsePKCS1PrivateKey(privPem.Bytes)
	if err != nil {
		return "", errors.Wrap(err, "Failed to parse key.")
	}
	// Create new JWT token with 10 minute (max duration) expiry
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iat": time.Now().Unix() - 60,
		"exp": time.Now().Unix() + (10 * 60),
		"iss": appID,
	})

	jwtString, err := token.SignedString(priv)
	if err != nil {
		return "", errors.Wrap(err, "Failed to create token.")
	}
	return jwtString, nil
}

func getInstallAccessToken(ctx context.Context, ghc *github.Client) (*string, error) {
	// Get organation installation ID
	orgInstallation, _, err := ghc.Apps.FindOrganizationInstallation(ctx, "sourcegraph")
	if err != nil {
		log.Fatal(err)
	}
	orgID := orgInstallation.ID

	// Create new installation token with 60 minute duraction with default read repo contents permissions
	token, _, err := ghc.Apps.CreateInstallationToken(ctx, *orgID, &github.InstallationTokenOptions{
		Repositories: []string{"sourcegraph"},
	})
	if err != nil {
		log.Fatal(err)
	}

	return token.Token, nil
}
