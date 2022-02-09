package docker

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/cockroachdb/errors"
	"github.com/docker/cli/cli/config/configfile"
	"github.com/docker/docker-credential-helpers/client"
	"github.com/docker/docker-credential-helpers/credentials"
)

func getStoreProvider() (string, error) {
	// TODO:
	// This is a hack to get the store provider.
	// It may not work with Windows.
	data, err := ioutil.ReadFile(fmt.Sprintf("%s/.docker/config.json", os.Getenv("HOME")))
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
