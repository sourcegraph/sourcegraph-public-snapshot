package provider


type Nobl9ProviderConfig struct {
	// the [Client ID](https://docs.nobl9.com/sloctl-user-guide/#configuration) of your Nobl9 account required to connect to Nobl9.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/nobl9/nobl9/0.22.0/docs#client_id Nobl9Provider#client_id}
	ClientId *string `field:"required" json:"clientId" yaml:"clientId"`
	// the [Client Secret](https://docs.nobl9.com/sloctl-user-guide/#configuration) of your Nobl9 account required to connect to Nobl9.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/nobl9/nobl9/0.22.0/docs#client_secret Nobl9Provider#client_secret}
	ClientSecret *string `field:"required" json:"clientSecret" yaml:"clientSecret"`
	// Alias name.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/nobl9/nobl9/0.22.0/docs#alias Nobl9Provider#alias}
	Alias *string `field:"optional" json:"alias" yaml:"alias"`
	// Nobl9 API URL.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/nobl9/nobl9/0.22.0/docs#ingest_url Nobl9Provider#ingest_url}
	IngestUrl *string `field:"optional" json:"ingestUrl" yaml:"ingestUrl"`
	// Authorization service configuration.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/nobl9/nobl9/0.22.0/docs#okta_auth_server Nobl9Provider#okta_auth_server}
	OktaAuthServer *string `field:"optional" json:"oktaAuthServer" yaml:"oktaAuthServer"`
	// Authorization service URL.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/nobl9/nobl9/0.22.0/docs#okta_org_url Nobl9Provider#okta_org_url}
	OktaOrgUrl *string `field:"optional" json:"oktaOrgUrl" yaml:"oktaOrgUrl"`
	// Nobl9 [Organization ID](https://docs.nobl9.com/API_Documentation/api-endpoints-for-slo-annotations/#common-headers) that contains resources managed by the Nobl9 Terraform provider.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/nobl9/nobl9/0.22.0/docs#organization Nobl9Provider#organization}
	Organization *string `field:"optional" json:"organization" yaml:"organization"`
	// Nobl9 project used when importing resources.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/nobl9/nobl9/0.22.0/docs#project Nobl9Provider#project}
	Project *string `field:"optional" json:"project" yaml:"project"`
}

