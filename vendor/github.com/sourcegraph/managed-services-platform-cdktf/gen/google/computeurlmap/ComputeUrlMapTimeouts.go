package computeurlmap


type ComputeUrlMapTimeouts struct {
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_url_map#create ComputeUrlMap#create}.
	Create *string `field:"optional" json:"create" yaml:"create"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_url_map#delete ComputeUrlMap#delete}.
	Delete *string `field:"optional" json:"delete" yaml:"delete"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_url_map#update ComputeUrlMap#update}.
	Update *string `field:"optional" json:"update" yaml:"update"`
}

