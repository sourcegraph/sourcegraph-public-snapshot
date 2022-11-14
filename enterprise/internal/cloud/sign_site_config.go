//go:build ignore

// Command sign_site_config signs the site config for Sourcegraph Cloud.
//
// # REQUIREMENTS
//
// You must provide a private key and a site config file to be signed.
//
// To sign site configs that are valid for Sourcegraph Cloud instances, you must use the private key at
// https://team-sourcegraph.1password.com/vaults/dnrhbauihkhjs5ag6vszsme45a/allitems/m4rqoaoujjwesf6twwqyr3lpde.
//
// To create a test private key that will NOT generate valid licenses, use:
//
//	ssh-keygen -t ed25519
//
// EXAMPLE
//
//	go run sign_site_config.go -private-key key.pem -site-config site-config.json
package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"golang.org/x/crypto/ssh"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/cloud"
)

var (
	privateKeyFile = flag.String("private-key", "", "file containing private key to sign site config")
	siteConfigFile = flag.String("site-config", "", "file containing site config to be signed")
)

func main() {
	flag.Parse()
	log.SetFlags(0)

	if *privateKeyFile == "" {
		log.Fatal("A private key file must be explicitly indicated, but was not.")
	}
	if *siteConfigFile == "" {
		log.Fatal("A site config file must be explicitly indicated, but was not.")
	}

	privateKeyData, err := os.ReadFile(*privateKeyFile)
	if err != nil {
		log.Fatalf("Failed to read private key: %v", err)
	}
	privateKey, err := ssh.ParsePrivateKey(privateKeyData)
	if err != nil {
		log.Fatalf("Failed to parse private key: %v", err)
	}

	siteConfigData, err := os.ReadFile(*siteConfigFile)
	if err != nil {
		log.Fatalf("Failed to read site config: %v", err)
	}

	signature, err := privateKey.Sign(rand.Reader, siteConfigData)
	if err != nil {
		log.Fatalf("Failed to sign site config: %v", err)
	}

	signedData, err := json.Marshal(
		cloud.SignedSiteConfig{
			Signature:  signature,
			SiteConfig: siteConfigData,
		},
	)
	if err != nil {
		log.Fatalf("Failed to marshal signed site data: %v", err)
	}

	fmt.Println(base64.RawURLEncoding.EncodeToString(signedData))
}
