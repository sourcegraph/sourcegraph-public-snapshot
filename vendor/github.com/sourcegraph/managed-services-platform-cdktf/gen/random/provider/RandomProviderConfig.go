package provider


type RandomProviderConfig struct {
	// Alias name.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/random/3.5.1/docs#alias RandomProvider#alias}
	Alias *string `field:"optional" json:"alias" yaml:"alias"`
}

