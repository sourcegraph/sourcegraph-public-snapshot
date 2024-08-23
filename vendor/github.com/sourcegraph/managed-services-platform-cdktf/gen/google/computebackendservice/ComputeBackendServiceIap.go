package computebackendservice


type ComputeBackendServiceIap struct {
	// OAuth2 Client ID for IAP.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_backend_service#oauth2_client_id ComputeBackendService#oauth2_client_id}
	Oauth2ClientId *string `field:"required" json:"oauth2ClientId" yaml:"oauth2ClientId"`
	// OAuth2 Client Secret for IAP.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_backend_service#oauth2_client_secret ComputeBackendService#oauth2_client_secret}
	Oauth2ClientSecret *string `field:"required" json:"oauth2ClientSecret" yaml:"oauth2ClientSecret"`
}

