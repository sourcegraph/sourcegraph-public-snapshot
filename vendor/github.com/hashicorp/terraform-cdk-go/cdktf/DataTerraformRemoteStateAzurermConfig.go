// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf


// Experimental.
type DataTerraformRemoteStateAzurermConfig struct {
	// Experimental.
	Defaults *map[string]interface{} `field:"optional" json:"defaults" yaml:"defaults"`
	// Experimental.
	Workspace *string `field:"optional" json:"workspace" yaml:"workspace"`
	// (Required) The Name of the Storage Container within the Storage Account.
	// Experimental.
	ContainerName *string `field:"required" json:"containerName" yaml:"containerName"`
	// (Required) The name of the Blob used to retrieve/store Terraform's State file inside the Storage Container.
	// Experimental.
	Key *string `field:"required" json:"key" yaml:"key"`
	// (Required) The Name of the Storage Account.
	// Experimental.
	StorageAccountName *string `field:"required" json:"storageAccountName" yaml:"storageAccountName"`
	// access_key - (Optional) The Access Key used to access the Blob Storage Account.
	//
	// This can also be sourced from the ARM_ACCESS_KEY environment variable.
	// Experimental.
	AccessKey *string `field:"optional" json:"accessKey" yaml:"accessKey"`
	// (Optional) The password associated with the Client Certificate specified in client_certificate_path.
	//
	// This can also be sourced from the
	// ARM_CLIENT_CERTIFICATE_PASSWORD environment variable.
	// Experimental.
	ClientCertificatePassword *string `field:"optional" json:"clientCertificatePassword" yaml:"clientCertificatePassword"`
	// (Optional) The path to the PFX file used as the Client Certificate when authenticating as a Service Principal.
	//
	// This can also be sourced from the
	// ARM_CLIENT_CERTIFICATE_PATH environment variable.
	// Experimental.
	ClientCertificatePath *string `field:"optional" json:"clientCertificatePath" yaml:"clientCertificatePath"`
	// (Optional) The Client ID of the Service Principal.
	//
	// This can also be sourced from the ARM_CLIENT_ID environment variable.
	// Experimental.
	ClientId *string `field:"optional" json:"clientId" yaml:"clientId"`
	// (Optional) The Client Secret of the Service Principal.
	//
	// This can also be sourced from the ARM_CLIENT_SECRET environment variable.
	// Experimental.
	ClientSecret *string `field:"optional" json:"clientSecret" yaml:"clientSecret"`
	// (Optional) The Custom Endpoint for Azure Resource Manager. This can also be sourced from the ARM_ENDPOINT environment variable.
	//
	// NOTE: An endpoint should only be configured when using Azure Stack.
	// Experimental.
	Endpoint *string `field:"optional" json:"endpoint" yaml:"endpoint"`
	// (Optional) The Azure Environment which should be used.
	//
	// This can also be sourced from the ARM_ENVIRONMENT environment variable.
	//  Possible values are public, china, german, stack and usgovernment. Defaults to public.
	// Experimental.
	Environment *string `field:"optional" json:"environment" yaml:"environment"`
	// (Optional) The Hostname of the Azure Metadata Service (for example management.azure.com), used to obtain the Cloud Environment when using a Custom Azure Environment. This can also be sourced from the ARM_METADATA_HOSTNAME Environment Variable.).
	// Experimental.
	MetadataHost *string `field:"optional" json:"metadataHost" yaml:"metadataHost"`
	// (Optional) The path to a custom Managed Service Identity endpoint which is automatically determined if not specified.
	//
	// This can also be sourced from the ARM_MSI_ENDPOINT environment variable.
	// Experimental.
	MsiEndpoint *string `field:"optional" json:"msiEndpoint" yaml:"msiEndpoint"`
	// (Optional) The bearer token for the request to the OIDC provider.
	//
	// This can
	// also be sourced from the ARM_OIDC_REQUEST_TOKEN or
	// ACTIONS_ID_TOKEN_REQUEST_TOKEN environment variables.
	// Experimental.
	OidcRequestToken *string `field:"optional" json:"oidcRequestToken" yaml:"oidcRequestToken"`
	// (Optional) The URL for the OIDC provider from which to request an ID token.
	//
	// This can also be sourced from the ARM_OIDC_REQUEST_URL or
	// ACTIONS_ID_TOKEN_REQUEST_URL environment variables.
	// Experimental.
	OidcRequestUrl *string `field:"optional" json:"oidcRequestUrl" yaml:"oidcRequestUrl"`
	// (Optional) The ID token when authenticating using OpenID Connect (OIDC).
	//
	// This can also be sourced from the ARM_OIDC_TOKEN environment variable.
	// Experimental.
	OidcToken *string `field:"optional" json:"oidcToken" yaml:"oidcToken"`
	// (Optional) The path to a file containing an ID token when authenticating using OpenID Connect (OIDC).
	//
	// This can also be sourced from the ARM_OIDC_TOKEN_FILE_PATH environment variable.
	// Experimental.
	OidcTokenFilePath *string `field:"optional" json:"oidcTokenFilePath" yaml:"oidcTokenFilePath"`
	// (Required) The Name of the Resource Group in which the Storage Account exists.
	// Experimental.
	ResourceGroupName *string `field:"optional" json:"resourceGroupName" yaml:"resourceGroupName"`
	// (Optional) The SAS Token used to access the Blob Storage Account.
	//
	// This can also be sourced from the ARM_SAS_TOKEN environment variable.
	// Experimental.
	SasToken *string `field:"optional" json:"sasToken" yaml:"sasToken"`
	// (Optional) Should the Blob used to store the Terraform Statefile be snapshotted before use?
	//
	// Defaults to false. This value can also be sourced
	// from the ARM_SNAPSHOT environment variable.
	// Experimental.
	Snapshot *bool `field:"optional" json:"snapshot" yaml:"snapshot"`
	// (Optional) The Subscription ID in which the Storage Account exists.
	//
	// This can also be sourced from the ARM_SUBSCRIPTION_ID environment variable.
	// Experimental.
	SubscriptionId *string `field:"optional" json:"subscriptionId" yaml:"subscriptionId"`
	// (Optional) The Tenant ID in which the Subscription exists.
	//
	// This can also be sourced from the ARM_TENANT_ID environment variable.
	// Experimental.
	TenantId *string `field:"optional" json:"tenantId" yaml:"tenantId"`
	// (Optional) Should AzureAD Authentication be used to access the Blob Storage Account.
	//
	// This can also be sourced from the ARM_USE_AZUREAD environment
	// variable.
	//
	// Note: When using AzureAD for Authentication to Storage you also need to
	// ensure the Storage Blob Data Owner role is assigned.
	// Experimental.
	UseAzureadAuth *bool `field:"optional" json:"useAzureadAuth" yaml:"useAzureadAuth"`
	// (Optional) Should MSAL be used for authentication instead of ADAL, and should Microsoft Graph be used instead of Azure Active Directory Graph?
	//
	// Defaults to true.
	//
	// Note: In Terraform 1.2 the Azure Backend uses MSAL (and Microsoft Graph)
	// rather than ADAL (and Azure Active Directory Graph) for authentication by
	// default - you can disable this by setting use_microsoft_graph to false.
	// This setting will be removed in Terraform 1.3, due to Microsoft's
	// deprecation of ADAL.
	// Experimental.
	UseMicrosoftGraph *bool `field:"optional" json:"useMicrosoftGraph" yaml:"useMicrosoftGraph"`
	// (Optional) Should Managed Service Identity authentication be used?
	//
	// This can also be sourced from the ARM_USE_MSI environment variable.
	// Experimental.
	UseMsi *bool `field:"optional" json:"useMsi" yaml:"useMsi"`
	// (Optional) Should OIDC authentication be used? This can also be sourced from the ARM_USE_OIDC environment variable.
	//
	// Note: When using OIDC for authentication, use_microsoft_graph
	// must be set to true (which is the default).
	// Experimental.
	UseOidc *bool `field:"optional" json:"useOidc" yaml:"useOidc"`
}

