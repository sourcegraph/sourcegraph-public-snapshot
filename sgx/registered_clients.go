package sgx

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"sourcegraph.com/sqs/pbtypes"
	"src.sourcegraph.com/sourcegraph/auth/idkey"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/sgx/cli"
	"src.sourcegraph.com/sourcegraph/util/timeutil"
)

func init() {
	g, err := cli.CLI.AddCommand("registered-clients",
		"manage registered API clients",
		"The registered-clients subcommands manage registered API clients.",
		&regClientsCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}
	g.Aliases = []string{"clients", "rc"}

	_, err = g.AddCommand("create",
		"create (register) an API client",
		"The create subcommand creates (registers) an API clients.",
		&regClientsCreateCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}

	c, err := g.AddCommand("list",
		"list registered API clients",
		"The list subcommand lists registered API clients.",
		&regClientsListCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}
	c.Aliases = []string{"ls"}

	_, err = g.AddCommand("current",
		"gets info about the currently authenticated registered API client",
		"The current subcommand gets info about the currently authenticated registered API client.",
		&regClientsCurrentCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}

	c, err = g.AddCommand("update",
		"updates a registered API client",
		"The update subcommand updates a registered API client.",
		&regClientsUpdateCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}

	c, err = g.AddCommand("delete",
		"deletes a registered API client",
		"The rm subcommand deletes a registered API client.",
		&regClientsDeleteCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}
	c.Aliases = []string{"rm"}
}

type regClientsCmd struct{}

func (*regClientsCmd) Execute(args []string) error { return nil }

type regClientsListCmd struct {
	Detail bool `short:"d" long:"detail" description:"show full details"`
}

func (c *regClientsListCmd) Execute(args []string) error {
	cl := cli.Client()

	opt := sourcegraph.RegisteredClientListOptions{
		ListOptions: sourcegraph.ListOptions{Page: 1},
	}
	for {
		clients, err := cl.RegisteredClients.List(cli.Ctx, &opt)
		if err != nil {
			return err
		}
		for _, client := range clients.Clients {
			if c.Detail {
				printRegisteredClient(client)
			} else {
				fmt.Printf("%- 48s   %- 20s\n", client.ID, timeutil.TimeAgo(client.CreatedAt))
			}
		}
		if !clients.HasMore {
			break
		}
		opt.Page++
	}
	return nil
}

type regClientsCreateCmd struct {
	ClientName  string `long:"client-name"`
	ClientURI   string `long:"client-uri"`
	RedirectURI string `long:"redirect-uri"`
	Description string `long:"description"`
	Type        string `long:"type" default:"SourcegraphServer"`
	IDKeyFile   string `short:"i" long:"id-key-file" description:"path to file containing ID key (only public key is transmitted)" default:"$SGPATH/id.pem"`
}

func (c *regClientsCreateCmd) Execute(args []string) error {
	cl := cli.Client()

	typ, ok := sourcegraph.RegisteredClientType_value[c.Type]
	if !ok {
		return fmt.Errorf("invalid --type %q; choices are %+v", c.Type, sourcegraph.RegisteredClientType_value)
	}

	c.IDKeyFile = os.ExpandEnv(c.IDKeyFile)
	data, err := ioutil.ReadFile(c.IDKeyFile)
	if err != nil {
		return err
	}
	idKey, err := idkey.New(data)
	if err != nil {
		return err
	}
	log.Printf("# Using public key from file %s", c.IDKeyFile)
	jwks, err := idKey.MarshalJWKSPublicKey()
	if err != nil {
		return err
	}

	regClient := &sourcegraph.RegisteredClient{
		ID:          idKey.ID,
		ClientName:  c.ClientName,
		ClientURI:   c.ClientURI,
		JWKS:        string(jwks),
		Description: c.Description,
		Type:        sourcegraph.RegisteredClientType(typ),
	}
	if c.RedirectURI != "" {
		regClient.RedirectURIs = []string{c.RedirectURI}
	}

	regClient, err = cl.RegisteredClients.Create(cli.Ctx, regClient)
	if err != nil {
		return err
	}

	log.Println("# Registered API client:")
	printRegisteredClient(regClient)
	return nil
}

type regClientsCurrentCmd struct{}

func (c *regClientsCurrentCmd) Execute(args []string) error {
	cl := cli.Client()

	client, err := cl.RegisteredClients.GetCurrent(cli.Ctx, &pbtypes.Void{})
	if err != nil {
		return err
	}
	printRegisteredClient(client)
	return nil
}

func printRegisteredClient(c *sourcegraph.RegisteredClient) {
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(b))
}

type regClientsUpdateCmd struct {
	ClientName    string `long:"client-name"`
	ClientURI     string `long:"client-uri"`
	RedirectURI   string `long:"redirect-uri"`
	Description   string `long:"description"`
	AllowLogins   string `long:"allow-logins" description:"set to 'all' to allow any user to login to this client" default:"restricted"`
	DefaultAccess string `long:"default-access" description:"set to 'write' to grant write access to new users (eg. for LDAP auth)" default:"read"`

	Args struct {
		ClientID string `name:"CLIENT-ID"`
	} `positional-args:"yes" count:"1"`
}

func (c *regClientsUpdateCmd) Execute(args []string) error {
	cl := cli.Client()

	rc, err := cl.RegisteredClients.Get(cli.Ctx, &sourcegraph.RegisteredClientSpec{ID: c.Args.ClientID})
	if err != nil {
		return err
	}

	fmt.Print(rc.ID, ": ")
	if c.ClientName != "" {
		rc.ClientName = c.ClientName
	}
	if c.ClientURI != "" {
		rc.ClientURI = c.ClientURI
	}
	if c.RedirectURI != "" {
		rc.RedirectURIs = []string{c.RedirectURI}
	}
	if c.Description != "" {
		rc.Description = c.Description
	}
	if c.AllowLogins != "" {
		if rc.Meta == nil {
			rc.Meta = map[string]string{}
		}
		rc.Meta["allow-logins"] = c.AllowLogins
	}
	if c.DefaultAccess != "" {
		if rc.Meta == nil {
			rc.Meta = map[string]string{}
		}
		rc.Meta["default-access"] = c.DefaultAccess
	}
	if _, err := cl.RegisteredClients.Update(cli.Ctx, rc); err != nil {
		return err
	}
	fmt.Println("updated")
	return nil
}

type regClientsDeleteCmd struct {
	Args struct {
		ClientIDs []string `name:"CLIENT-IDs"`
	} `positional-args:"yes"`
}

func (c *regClientsDeleteCmd) Execute(args []string) error {
	cl := cli.Client()

	for _, clientID := range c.Args.ClientIDs {
		fmt.Print(clientID, ": ")
		if _, err := cl.RegisteredClients.Delete(cli.Ctx, &sourcegraph.RegisteredClientSpec{ID: clientID}); err != nil {
			return err
		}
		fmt.Println("deleted")
	}
	return nil
}
