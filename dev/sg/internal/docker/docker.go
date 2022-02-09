package docker

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"

	"github.com/cockroachdb/errors"
	"github.com/docker/cli/cli/config/configfile"
	"github.com/docker/docker-credential-helpers/client"
	"github.com/docker/docker-credential-helpers/credentials"
)

func getStoreProvider(serverAddress string) (string, error) {
	// sanitize the server address for Docker Hub
	if serverAddress == "" ||
		serverAddress == "https://index.docker.io" ||
		serverAddress == "https://registry-1.docker.io" ||
		serverAddress == "https://registry.docker.io" ||
		serverAddress == "https://docker.io" ||
		serverAddress == "https://registry.hub.docker.com" ||
		serverAddress == "https://index.docker.io/v2/" {
		serverAddress = "https://index.docker.io/v1/"
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	data, err := ioutil.ReadFile(filepath.Join(homeDir, ".docker", "config.json"))
	if err != nil {
		return "", err
	}

	var config configfile.ConfigFile
	if err := json.Unmarshal(data, &config); err != nil {
		return "", err
	}

	if serverAddress == "https://index.docker.io/v1/" && config.CredentialsStore != "" {
		return config.CredentialsStore, nil
	}

	url, err := url.Parse(serverAddress)
	if err != nil {
		return "", errors.Newf("failed to parse server address %s", serverAddress)
	}

	if config.CredentialHelpers[url.Host] != "" {
		return config.CredentialHelpers[url.Host], nil
	}

	return "", errors.Newf("failed to find store provider or credential helper for %s", serverAddress)
}

func GetCredentialsFromStore(serverAddress string) (*credentials.Credentials, error) {
	provider, err := getStoreProvider(serverAddress)
	if err != nil {
		return nil, err
	}
	program := client.NewShellProgramFunc(fmt.Sprintf("docker-credential-%s", provider))
	credentials, err := client.Get(program, serverAddress)
	if err != nil {
		return nil, err
	}

	return credentials, err
}
