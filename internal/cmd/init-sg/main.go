package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/sourcegraph/sourcegraph/internal/gqltestutil"
)

var client *gqltestutil.Client

var (
	initSG   = flag.NewFlagSet("initserver", flag.ExitOnError)
	addRepos = flag.NewFlagSet("addrepos", flag.ExitOnError)

	baseURL  = initSG.String("baseurl", os.Getenv("SOURCEGRAPH_BASE_URL"), "The base URL of the Sourcegraph instance. (Required)")
	email    = initSG.String("email", os.Getenv("TEST_USER_EMAIL"), "The email of the admin user. (Required)")
	username = initSG.String("username", os.Getenv("SOURCEGRAPH_SUDO_USER"), "The username of the admin user. (Required)")
	password = initSG.String("password", os.Getenv("TEST_USER_PASSWORD"), "The password of the admin user. (Required)")

	githubToken    = addRepos.String("githubtoken", os.Getenv("GITHUB_TOKEN"), "The github access token that will be used to authenticate an external service. (Required)")
	addReposConfig = addRepos.String("config", "", "Path to the external service config. (Required)")

	home    = os.Getenv("HOME")
	profile = home + "/.profile"
)

func main() {
	flag.Parse()

	if len(os.Args) < 2 {
		fmt.Println("initSG or addRepos subcommand is required")
		flag.PrintDefaults()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "initSG":
		initSG.Parse(os.Args[2:])
		initSourceGraph()
	case "addRepos":
		addRepos.Parse(os.Args[2:])
		addReposCommand()
	case "default":
		flag.PrintDefaults()
		os.Exit(1)
	}

}
