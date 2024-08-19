package provider


type PostgresqlProviderClientcert struct {
	// The SSL client certificate file path. The file must contain PEM encoded data.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/sourcegraph/postgresql/1.23.0-sg.2/docs#cert PostgresqlProvider#cert}
	Cert *string `field:"required" json:"cert" yaml:"cert"`
	// The SSL client certificate private key file path. The file must contain PEM encoded data.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/sourcegraph/postgresql/1.23.0-sg.2/docs#key PostgresqlProvider#key}
	Key *string `field:"required" json:"key" yaml:"key"`
	// Must be set to true if you are inlining the cert/key instead of using a file path.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/sourcegraph/postgresql/1.23.0-sg.2/docs#sslinline PostgresqlProvider#sslinline}
	Sslinline interface{} `field:"optional" json:"sslinline" yaml:"sslinline"`
}

