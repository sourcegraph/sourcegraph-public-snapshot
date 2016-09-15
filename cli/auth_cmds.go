package cli

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/dgrijalva/jwt-go"

	"sourcegraph.com/sourcegraph/sourcegraph/cli/cli"

	"sourcegraph.com/sqs/pbtypes"
)

func init() {
	authGroup, err := cli.CLI.AddCommand("auth",
		"generate auth tokens and tickets",
		"The auth subcommands generate authentication tokens and tickets.",
		&authCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}

	identifyC, err := authGroup.AddCommand("whoami",
		"identify the current user",
		"The whoami ('who am I?') subcommand prints out information about the currently authenticated user.",
		&authIdentifyCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}
	identifyC.Aliases = []string{"id"}

	_, err = authGroup.AddCommand("jwt",
		"decode a JWT",
		"Decodes a JWT.",
		&authJWTCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}
}

type authCmd struct{}

func (c *authCmd) Execute(args []string) error { return nil }

type authIdentifyCmd struct{}

func (c *authIdentifyCmd) Execute(args []string) error {
	cl := cliClient
	authInfo, err := cl.Auth.Identify(cliContext, &pbtypes.Void{})
	if err != nil {
		return err
	}

	if authInfo.UID != 0 {
		fmt.Printf("%s (%d)\n", authInfo.Login, authInfo.UID)
	}

	return nil
}

type authJWTCmd struct {
}

func (c *authJWTCmd) Execute(args []string) error {
	log.Println("(reading on stdin...)")

	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return err
	}
	log.Println()

	tok, err := jwt.Parse(string(data), func(*jwt.Token) (interface{}, error) {
		return nil, nil
	})
	if err != nil {
		return err
	}
	fmt.Println("(NO VERIFICATION PERFORMED; COULD BE FORGED)")
	fmt.Println()
	fmt.Println("## Header")
	fmt.Println(tok.Header)
	fmt.Println()
	fmt.Println("## Claims")
	for k, v := range tok.Claims {
		fmt.Println(k+":", v)
	}

	return nil
}
