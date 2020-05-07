package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/internal/e2eutil"
)

/*
	NOTE: For easier testing, run Sourcegraph instance without volume:
			docker run --publish 7080:7080 --rm  sourcegraph/server:insiders
*/

func main() {
	baseURL := flag.String("base-url", "http://localhost:7080", "The base URL of the Sourcegraph instance")
	email := flag.String("email", "e2e@sourcegraph.com", "The email of the admin user")
	username := flag.String("username", "e2e-admin", "The username of the admin user")
	password := flag.String("password", "supersecurepassword", "The password of the admin user")
	flag.Parse()

	*baseURL = strings.TrimSuffix(*baseURL, "/")

	// TODO(jchen): Find an easy and fast way to determine if the instance has done "Site admin init".
	client, err := e2eutil.SiteAdminInit(*baseURL, *email, *username, *password)
	if err != nil {
		log15.Error("Failed to create site admin", "err", err)
		os.Exit(1)
	}
	log15.Info("Site admin has been created", "username", *username)
	//client, err := e2eutil.SignIn(*baseURL, *email, *password)
	//if err != nil {
	//	log15.Error("Failed to sign in", "err", err)
	//	os.Exit(1)
	//}
	//log15.Info("Site admin authenticated", "username", *username)

	fmt.Printf("client: %+v\n", client)
}
