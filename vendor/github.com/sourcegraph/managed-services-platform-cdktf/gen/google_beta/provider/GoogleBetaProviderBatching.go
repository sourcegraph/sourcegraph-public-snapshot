package provider


type GoogleBetaProviderBatching struct {
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#enable_batching GoogleBetaProvider#enable_batching}.
	EnableBatching interface{} `field:"optional" json:"enableBatching" yaml:"enableBatching"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#send_after GoogleBetaProvider#send_after}.
	SendAfter *string `field:"optional" json:"sendAfter" yaml:"sendAfter"`
}

