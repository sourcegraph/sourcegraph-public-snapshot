package workspace

import (
	"encoding/json"
	"os"
)

type cniConfig struct {
	CNIVersion string `json:"cniVersion"`
	Name       string `json:"name"`
	Plugins    []any  `json:"plugins"`
}

type cniBridgePluginConfig struct {
	Type             string                    `json:"type"`
	Bridge           string                    `json:"bridge"`
	IsGateway        bool                      `json:"isGateway"`
	IsDefaultGateway bool                      `json:"isDefaultGateway"`
	PromiscMode      bool                      `json:"promiscMode"`
	IPMasq           bool                      `json:"ipMasq"`
	IPAM             cniBridgePluginConfigIPAM `json:"ipam"`
}

type cniBridgePluginConfigIPAM struct {
	Type   string `json:"type"`
	Subnet string `json:"subnet"`
}

type cniPortmapPluginConfig struct {
	Type         string                             `json:"type"`
	Capabilities cniPortmapPluginConfigCapabilities `json:"capabilities"`
}
type cniPortmapPluginConfigCapabilities struct {
	PortMappings bool `json:"portMappings"`
}

type cniFirewallPluginConfig struct {
	Type string `json:"type"`
}

type cniIsolationPluginConfig struct {
	Type string `json:"type"`
}

type cniBandwidthPluginConfig struct {
	Type         string `json:"type"`
	Name         string `json:"name"`
	IngressRate  int32  `json:"ingressRate"`
	IngressBurst int32  `json:"ingressBurst"`
	EgressRate   int32  `json:"egressRate"`
	EgressBurst  int32  `json:"egressBurst"`
}

// SetupCNI generates a CNI config to be used for VM creation.
// It writes a config to the temporary directory, which is then
// passed to ignite later.
func SetupCNI(tmpDir string) error {
	c := cniConfig{
		CNIVersion: "0.4.0",
		Name:       "ignite-cni-bridge",
		Plugins: []any{
			cniBridgePluginConfig{
				Type:             "bridge",
				Bridge:           "ignite0",
				IsGateway:        true,
				IsDefaultGateway: true,
				PromiscMode:      false,
				IPMasq:           true,
				IPAM: cniBridgePluginConfigIPAM{
					Type:   "host-local",
					Subnet: "10.61.0.0/16",
				},
			},
			cniPortmapPluginConfig{
				Type: "portmap",
				Capabilities: cniPortmapPluginConfigCapabilities{
					PortMappings: true,
				},
			},
			cniFirewallPluginConfig{
				Type: "firewall",
			},
			cniIsolationPluginConfig{
				Type: "isolation",
			},
			cniBandwidthPluginConfig{
				Type:         "bandwidth",
				Name:         "slowdown",
				IngressRate:  524288000,
				IngressBurst: 1048576000,
				EgressRate:   524288000,
				EgressBurst:  1048576000,
			},
		},
	}
	f, err := os.CreateTemp(tmpDir, "cni-config.json")
	if err != nil {
		return err
	}
	mc, err := json.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(f.Name(), mc, os.ModePerm)
}
