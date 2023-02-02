package docker

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/docker/docker-credential-helpers/client"
	"github.com/docker/docker-credential-helpers/credentials"
)

// Mirror of https://github.com/docker/cli/blob/c780f7c4abaf67034ecfaa0611e03695cf9e4a3e/cli/config/configfile/file.go#L27-L55
// ConfigFile ~/.docker/config.json file info
type ConfigFile struct {
	AuthConfigs          map[string]AuthConfig        `json:"auths"`
	HTTPHeaders          map[string]string            `json:"HttpHeaders,omitempty"`
	PsFormat             string                       `json:"psFormat,omitempty"`
	ImagesFormat         string                       `json:"imagesFormat,omitempty"`
	NetworksFormat       string                       `json:"networksFormat,omitempty"`
	PluginsFormat        string                       `json:"pluginsFormat,omitempty"`
	VolumesFormat        string                       `json:"volumesFormat,omitempty"`
	StatsFormat          string                       `json:"statsFormat,omitempty"`
	DetachKeys           string                       `json:"detachKeys,omitempty"`
	CredentialsStore     string                       `json:"credsStore,omitempty"`
	CredentialHelpers    map[string]string            `json:"credHelpers,omitempty"`
	Filename             string                       `json:"-"` // Note: for internal use only
	ServiceInspectFormat string                       `json:"serviceInspectFormat,omitempty"`
	ServicesFormat       string                       `json:"servicesFormat,omitempty"`
	TasksFormat          string                       `json:"tasksFormat,omitempty"`
	SecretFormat         string                       `json:"secretFormat,omitempty"`
	ConfigFormat         string                       `json:"configFormat,omitempty"`
	NodesFormat          string                       `json:"nodesFormat,omitempty"`
	PruneFilters         []string                     `json:"pruneFilters,omitempty"`
	Proxies              map[string]ProxyConfig       `json:"proxies,omitempty"`
	Experimental         string                       `json:"experimental,omitempty"`
	StackOrchestrator    string                       `json:"stackOrchestrator,omitempty"`
	Kubernetes           *KubernetesConfig            `json:"kubernetes,omitempty"`
	CurrentContext       string                       `json:"currentContext,omitempty"`
	CLIPluginsExtraDirs  []string                     `json:"cliPluginsExtraDirs,omitempty"`
	Plugins              map[string]map[string]string `json:"plugins,omitempty"`
	Aliases              map[string]string            `json:"aliases,omitempty"`
}

// Mirror of https://github.com/docker/cli/blob/c780f7c4abaf67034ecfaa0611e03695cf9e4a3e/cli/config/configfile/file.go#L58-L64
// ProxyConfig contains proxy configuration settings
type ProxyConfig struct {
	HTTPProxy  string `json:"httpProxy,omitempty"`
	HTTPSProxy string `json:"httpsProxy,omitempty"`
	NoProxy    string `json:"noProxy,omitempty"`
	FTPProxy   string `json:"ftpProxy,omitempty"`
}

// Mirror of https://github.com/docker/cli/blob/c780f7c4abaf67034ecfaa0611e03695cf9e4a3e/cli/config/configfile/file.go#L67-L69
// KubernetesConfig contains Kubernetes orchestrator settings
type KubernetesConfig struct {
	AllNamespaces string `json:"allNamespaces,omitempty"`
}

// Mirror of https://github.com/docker/cli/blob/c780f7c4abaf67034ecfaa0611e03695cf9e4a3e/cli/config/types/authconfig.go#L4-L22
// AuthConfig contains authorization information for connecting to a Registry
type AuthConfig struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Auth     string `json:"auth,omitempty"`

	// Email is an optional value associated with the username.
	// This field is deprecated and will be removed in a later
	// version of docker.
	Email string `json:"email,omitempty"`

	ServerAddress string `json:"serveraddress,omitempty"`

	// IdentityToken is used to authenticate the user and get
	// an access token for the registry.
	IdentityToken string `json:"identitytoken,omitempty"`

	// RegistryToken is a bearer token to be sent to a registry
	RegistryToken string `json:"registrytoken,omitempty"`
}

func getStoreProvider(serverAddress string) (string, error) {
	// sanitize the server address for Docker Hub
	if serverAddress == "" ||
		serverAddress == "https://index.docker.io" ||
		serverAddress == "https://registry-1.docker.io" ||
		serverAddress == "https://registry.docker.io" ||
		serverAddress == "https://docker.io" ||
		serverAddress == "https://registry.hub.docker.com" ||
		serverAddress == "https://index.docker.io/v2/" {
		serverAddress = "https://registry.hub.docker.com/v2"
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", errors.Wrap(err, "failed to get user home directory")
	}

	data, err := os.ReadFile(filepath.Join(homeDir, ".docker", "config.json"))
	if err != nil {
		return "", errors.Wrap(err, "failed to read docker config")
	}

	var config ConfigFile
	if err := json.Unmarshal(data, &config); err != nil {
		return "", errors.Wrap(err, "failed to parse docker config")
	}

	if serverAddress == "https://index.docker.io/v1/" && config.CredentialsStore != "" {
		return config.CredentialsStore, nil
	}

	serverUrl, err := url.Parse(serverAddress)
	if err != nil {
		return "", errors.Wrapf(err, "failed to parse server address %s", serverAddress)
	}

	if config.CredentialHelpers[serverUrl.Host] != "" {
		return config.CredentialHelpers[serverUrl.Host], nil
	}

	return "", errors.Errorf("failed to find store provider or credential helper for %s", serverAddress)
}

func GetCredentialsFromStore(serverAddress string) (*credentials.Credentials, error) {
	provider, err := getStoreProvider(serverAddress)
	if err != nil {
		return nil, err
	}
	program := client.NewShellProgramFunc(fmt.Sprintf("docker-credential-%s", provider))
	creds, err := client.Get(program, serverAddress)
	if err != nil {
		return nil, err
	}

	return creds, err
}
