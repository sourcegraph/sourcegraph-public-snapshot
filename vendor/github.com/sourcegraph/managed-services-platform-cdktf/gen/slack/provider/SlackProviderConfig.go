package provider


type SlackProviderConfig struct {
	// The Slack token.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/pablovarela/slack/1.2.2/docs#token SlackProvider#token}
	Token *string `field:"required" json:"token" yaml:"token"`
	// Alias name.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/pablovarela/slack/1.2.2/docs#alias SlackProvider#alias}
	Alias *string `field:"optional" json:"alias" yaml:"alias"`
}

