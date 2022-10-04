package client

import (
	"encoding/json"

	"github.com/sourcegraph/sourcegraph/internal/gqltestutil"
	"github.com/sourcegraph/sourcegraph/schema"
)

func AddGitHubCodeHost(client *gqltestutil.Client, displayName string, conn *schema.GitHubConnection) error {
	b, err := json.Marshal(conn)
	if err != nil {
		return err
	}
	input := gqltestutil.AddExternalServiceInput{
		Kind:        "GITHUB",
		DisplayName: displayName,
		Config:      string(b),
	}
	_, err = client.AddExternalService(input)
	return err
}

func AddCodeHostsByGitHubOrgs(client *gqltestutil.Client, token string, displayName string, orgs []string) error {
	conn := schema.GitHubConnection{
		Url:   "https://github.com",
		Token: token,
		Orgs:  orgs,
	}
	return AddGitHubCodeHost(client, displayName, &conn)
}
