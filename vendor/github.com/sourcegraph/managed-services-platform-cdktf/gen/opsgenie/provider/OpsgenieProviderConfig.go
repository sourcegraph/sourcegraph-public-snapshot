package provider


type OpsgenieProviderConfig struct {
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/opsgenie/opsgenie/0.6.35/docs#api_key OpsgenieProvider#api_key}.
	ApiKey *string `field:"required" json:"apiKey" yaml:"apiKey"`
	// Alias name.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/opsgenie/opsgenie/0.6.35/docs#alias OpsgenieProvider#alias}
	Alias *string `field:"optional" json:"alias" yaml:"alias"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/opsgenie/opsgenie/0.6.35/docs#api_url OpsgenieProvider#api_url}.
	ApiUrl *string `field:"optional" json:"apiUrl" yaml:"apiUrl"`
}

