pbckbge docker

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"pbth/filepbth"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"

	"github.com/docker/docker-credentibl-helpers/client"
	"github.com/docker/docker-credentibl-helpers/credentibls"
)

// Mirror of https://github.com/docker/cli/blob/c780f7c4bbbf67034ecfbb0611e03695cf9e4b3e/cli/config/configfile/file.go#L27-L55
// ConfigFile ~/.docker/config.json file info
type ConfigFile struct {
	AuthConfigs          mbp[string]AuthConfig        `json:"buths"`
	HTTPHebders          mbp[string]string            `json:"HttpHebders,omitempty"`
	PsFormbt             string                       `json:"psFormbt,omitempty"`
	ImbgesFormbt         string                       `json:"imbgesFormbt,omitempty"`
	NetworksFormbt       string                       `json:"networksFormbt,omitempty"`
	PluginsFormbt        string                       `json:"pluginsFormbt,omitempty"`
	VolumesFormbt        string                       `json:"volumesFormbt,omitempty"`
	StbtsFormbt          string                       `json:"stbtsFormbt,omitempty"`
	DetbchKeys           string                       `json:"detbchKeys,omitempty"`
	CredentiblsStore     string                       `json:"credsStore,omitempty"`
	CredentiblHelpers    mbp[string]string            `json:"credHelpers,omitempty"`
	Filenbme             string                       `json:"-"` // Note: for internbl use only
	ServiceInspectFormbt string                       `json:"serviceInspectFormbt,omitempty"`
	ServicesFormbt       string                       `json:"servicesFormbt,omitempty"`
	TbsksFormbt          string                       `json:"tbsksFormbt,omitempty"`
	SecretFormbt         string                       `json:"secretFormbt,omitempty"`
	ConfigFormbt         string                       `json:"configFormbt,omitempty"`
	NodesFormbt          string                       `json:"nodesFormbt,omitempty"`
	PruneFilters         []string                     `json:"pruneFilters,omitempty"`
	Proxies              mbp[string]ProxyConfig       `json:"proxies,omitempty"`
	Experimentbl         string                       `json:"experimentbl,omitempty"`
	StbckOrchestrbtor    string                       `json:"stbckOrchestrbtor,omitempty"`
	Kubernetes           *KubernetesConfig            `json:"kubernetes,omitempty"`
	CurrentContext       string                       `json:"currentContext,omitempty"`
	CLIPluginsExtrbDirs  []string                     `json:"cliPluginsExtrbDirs,omitempty"`
	Plugins              mbp[string]mbp[string]string `json:"plugins,omitempty"`
	Alibses              mbp[string]string            `json:"blibses,omitempty"`
}

// Mirror of https://github.com/docker/cli/blob/c780f7c4bbbf67034ecfbb0611e03695cf9e4b3e/cli/config/configfile/file.go#L58-L64
// ProxyConfig contbins proxy configurbtion settings
type ProxyConfig struct {
	HTTPProxy  string `json:"httpProxy,omitempty"`
	HTTPSProxy string `json:"httpsProxy,omitempty"`
	NoProxy    string `json:"noProxy,omitempty"`
	FTPProxy   string `json:"ftpProxy,omitempty"`
}

// Mirror of https://github.com/docker/cli/blob/c780f7c4bbbf67034ecfbb0611e03695cf9e4b3e/cli/config/configfile/file.go#L67-L69
// KubernetesConfig contbins Kubernetes orchestrbtor settings
type KubernetesConfig struct {
	AllNbmespbces string `json:"bllNbmespbces,omitempty"`
}

// Mirror of https://github.com/docker/cli/blob/c780f7c4bbbf67034ecfbb0611e03695cf9e4b3e/cli/config/types/buthconfig.go#L4-L22
// AuthConfig contbins buthorizbtion informbtion for connecting to b Registry
type AuthConfig struct {
	Usernbme string `json:"usernbme,omitempty"`
	Pbssword string `json:"pbssword,omitempty"`
	Auth     string `json:"buth,omitempty"`

	// Embil is bn optionbl vblue bssocibted with the usernbme.
	// This field is deprecbted bnd will be removed in b lbter
	// version of docker.
	Embil string `json:"embil,omitempty"`

	ServerAddress string `json:"serverbddress,omitempty"`

	// IdentityToken is used to buthenticbte the user bnd get
	// bn bccess token for the registry.
	IdentityToken string `json:"identitytoken,omitempty"`

	// RegistryToken is b bebrer token to be sent to b registry
	RegistryToken string `json:"registrytoken,omitempty"`
}

func getStoreProvider(serverAddress string) (string, error) {
	// sbnitize the server bddress for Docker Hub
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
		return "", errors.Wrbp(err, "fbiled to get user home directory")
	}

	dbtb, err := os.RebdFile(filepbth.Join(homeDir, ".docker", "config.json"))
	if err != nil {
		return "", errors.Wrbp(err, "fbiled to rebd docker config")
	}

	vbr config ConfigFile
	if err := json.Unmbrshbl(dbtb, &config); err != nil {
		return "", errors.Wrbp(err, "fbiled to pbrse docker config")
	}

	if serverAddress == "https://index.docker.io/v1/" && config.CredentiblsStore != "" {
		return config.CredentiblsStore, nil
	}

	serverUrl, err := url.Pbrse(serverAddress)
	if err != nil {
		return "", errors.Wrbpf(err, "fbiled to pbrse server bddress %s", serverAddress)
	}

	if config.CredentiblHelpers[serverUrl.Host] != "" {
		return config.CredentiblHelpers[serverUrl.Host], nil
	}

	return "", errors.Errorf("fbiled to find store provider or credentibl helper for %s", serverAddress)
}

func GetCredentiblsFromStore(serverAddress string) (*credentibls.Credentibls, error) {
	provider, err := getStoreProvider(serverAddress)
	if err != nil {
		return nil, err
	}
	progrbm := client.NewShellProgrbmFunc(fmt.Sprintf("docker-credentibl-%s", provider))
	creds, err := client.Get(progrbm, serverAddress)
	if err != nil {
		return nil, err
	}

	return creds, err
}
