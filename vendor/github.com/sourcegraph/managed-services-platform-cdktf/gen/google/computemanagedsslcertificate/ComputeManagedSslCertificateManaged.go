package computemanagedsslcertificate


type ComputeManagedSslCertificateManaged struct {
	// Domains for which a managed SSL certificate will be valid.
	//
	// Currently,
	// there can be up to 100 domains in this list.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_managed_ssl_certificate#domains ComputeManagedSslCertificate#domains}
	Domains *[]*string `field:"required" json:"domains" yaml:"domains"`
}

