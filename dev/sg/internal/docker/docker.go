package docker

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cockroachdb/errors"
	"github.com/docker/cli/cli/config/configfile"
	"github.com/docker/docker-credential-helpers/client"
	"github.com/docker/docker-credential-helpers/credentials"
)

func getStoreProvider() (string, error) {
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

	if config.CredentialsStore != "" {
		return config.CredentialsStore, nil
	}

	return "", errors.New("failed to find store provider")
}

func GetCredentialsFromStore(serverAddress string) (*credentials.Credentials, error) {
	provider, err := getStoreProvider()
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
