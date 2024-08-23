package computeurlmap


type ComputeUrlMapTest struct {
	// Host portion of the URL.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_url_map#host ComputeUrlMap#host}
	Host *string `field:"required" json:"host" yaml:"host"`
	// Path portion of the URL.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_url_map#path ComputeUrlMap#path}
	Path *string `field:"required" json:"path" yaml:"path"`
	// The backend service or backend bucket link that should be matched by this test.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_url_map#service ComputeUrlMap#service}
	Service *string `field:"required" json:"service" yaml:"service"`
	// Description of this test case.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_url_map#description ComputeUrlMap#description}
	Description *string `field:"optional" json:"description" yaml:"description"`
}

