package provider


type GoogleProviderBatching struct {
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs#enable_batching GoogleProvider#enable_batching}.
	EnableBatching interface{} `field:"optional" json:"enableBatching" yaml:"enableBatching"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs#send_after GoogleProvider#send_after}.
	SendAfter *string `field:"optional" json:"sendAfter" yaml:"sendAfter"`
}

