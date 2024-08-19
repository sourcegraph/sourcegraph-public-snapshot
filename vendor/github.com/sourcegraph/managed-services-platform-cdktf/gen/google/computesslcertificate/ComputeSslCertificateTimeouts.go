package computesslcertificate


type ComputeSslCertificateTimeouts struct {
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_ssl_certificate#create ComputeSslCertificate#create}.
	Create *string `field:"optional" json:"create" yaml:"create"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_ssl_certificate#delete ComputeSslCertificate#delete}.
	Delete *string `field:"optional" json:"delete" yaml:"delete"`
}

