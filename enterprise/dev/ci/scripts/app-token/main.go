package main

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"flag"
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
	appID := flag.String("appid", os.Getenv("GITHUB_APP_ID"), "(required) github application id.")
	keyPath := flag.String("keypath", os.Getenv("KEY_PATH"), "(required) path to private key file for github app.")
	help := flag.Bool("help", false, "Show help.")

	flag.Parse()

	if *help || len(*appID) == 0 || len(*keyPath) == 0 {
		flag.PrintDefaults()
		os.Exit(0)
	}

	jwt := genJwtToken(*appID, *keyPath)

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: jwt},
	)
	tc := oauth2.NewClient(ctx, ts)
	ghc := github.NewClient(tc)

	fmt.Println(*getInstallAccessToken(ctx, ghc))

}

func genJwtToken(appID string, keyPath string) string {
	rawPem, err := ioutil.ReadFile(keyPath)
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
