// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf


// Deprecated: CDK for Terraform no longer supports the swift backend. Terraform deprecated swift in v1.2.3 and removed it in v1.3.
type DataTerraformRemoteStateSwiftConfig struct {
	// Deprecated: CDK for Terraform no longer supports the swift backend. Terraform deprecated swift in v1.2.3 and removed it in v1.3.
	Defaults *map[string]interface{} `field:"optional" json:"defaults" yaml:"defaults"`
	// Deprecated: CDK for Terraform no longer supports the swift backend. Terraform deprecated swift in v1.2.3 and removed it in v1.3.
	Workspace *string `field:"optional" json:"workspace" yaml:"workspace"`
	// Deprecated: CDK for Terraform no longer supports the swift backend. Terraform deprecated swift in v1.2.3 and removed it in v1.3.
	Container *string `field:"required" json:"container" yaml:"container"`
	// Deprecated: CDK for Terraform no longer supports the swift backend. Terraform deprecated swift in v1.2.3 and removed it in v1.3.
	ApplicationCredentialId *string `field:"optional" json:"applicationCredentialId" yaml:"applicationCredentialId"`
	// Deprecated: CDK for Terraform no longer supports the swift backend. Terraform deprecated swift in v1.2.3 and removed it in v1.3.
	ApplicationCredentialName *string `field:"optional" json:"applicationCredentialName" yaml:"applicationCredentialName"`
	// Deprecated: CDK for Terraform no longer supports the swift backend. Terraform deprecated swift in v1.2.3 and removed it in v1.3.
	ApplicationCredentialSecret *string `field:"optional" json:"applicationCredentialSecret" yaml:"applicationCredentialSecret"`
	// Deprecated: CDK for Terraform no longer supports the swift backend. Terraform deprecated swift in v1.2.3 and removed it in v1.3.
	ArchiveContainer *string `field:"optional" json:"archiveContainer" yaml:"archiveContainer"`
	// Deprecated: CDK for Terraform no longer supports the swift backend. Terraform deprecated swift in v1.2.3 and removed it in v1.3.
	AuthUrl *string `field:"optional" json:"authUrl" yaml:"authUrl"`
	// Deprecated: CDK for Terraform no longer supports the swift backend. Terraform deprecated swift in v1.2.3 and removed it in v1.3.
	CacertFile *string `field:"optional" json:"cacertFile" yaml:"cacertFile"`
	// Deprecated: CDK for Terraform no longer supports the swift backend. Terraform deprecated swift in v1.2.3 and removed it in v1.3.
	Cert *string `field:"optional" json:"cert" yaml:"cert"`
	// Deprecated: CDK for Terraform no longer supports the swift backend. Terraform deprecated swift in v1.2.3 and removed it in v1.3.
	Cloud *string `field:"optional" json:"cloud" yaml:"cloud"`
	// Deprecated: CDK for Terraform no longer supports the swift backend. Terraform deprecated swift in v1.2.3 and removed it in v1.3.
	DefaultDomain *string `field:"optional" json:"defaultDomain" yaml:"defaultDomain"`
	// Deprecated: CDK for Terraform no longer supports the swift backend. Terraform deprecated swift in v1.2.3 and removed it in v1.3.
	DomainId *string `field:"optional" json:"domainId" yaml:"domainId"`
	// Deprecated: CDK for Terraform no longer supports the swift backend. Terraform deprecated swift in v1.2.3 and removed it in v1.3.
	DomainName *string `field:"optional" json:"domainName" yaml:"domainName"`
	// Deprecated: CDK for Terraform no longer supports the swift backend. Terraform deprecated swift in v1.2.3 and removed it in v1.3.
	ExpireAfter *string `field:"optional" json:"expireAfter" yaml:"expireAfter"`
	// Deprecated: CDK for Terraform no longer supports the swift backend. Terraform deprecated swift in v1.2.3 and removed it in v1.3.
	Insecure *bool `field:"optional" json:"insecure" yaml:"insecure"`
	// Deprecated: CDK for Terraform no longer supports the swift backend. Terraform deprecated swift in v1.2.3 and removed it in v1.3.
	Key *string `field:"optional" json:"key" yaml:"key"`
	// Deprecated: CDK for Terraform no longer supports the swift backend. Terraform deprecated swift in v1.2.3 and removed it in v1.3.
	Password *string `field:"optional" json:"password" yaml:"password"`
	// Deprecated: CDK for Terraform no longer supports the swift backend. Terraform deprecated swift in v1.2.3 and removed it in v1.3.
	ProjectDomainId *string `field:"optional" json:"projectDomainId" yaml:"projectDomainId"`
	// Deprecated: CDK for Terraform no longer supports the swift backend. Terraform deprecated swift in v1.2.3 and removed it in v1.3.
	ProjectDomainName *string `field:"optional" json:"projectDomainName" yaml:"projectDomainName"`
	// Deprecated: CDK for Terraform no longer supports the swift backend. Terraform deprecated swift in v1.2.3 and removed it in v1.3.
	RegionName *string `field:"optional" json:"regionName" yaml:"regionName"`
	// Deprecated: CDK for Terraform no longer supports the swift backend. Terraform deprecated swift in v1.2.3 and removed it in v1.3.
	StateName *string `field:"optional" json:"stateName" yaml:"stateName"`
	// Deprecated: CDK for Terraform no longer supports the swift backend. Terraform deprecated swift in v1.2.3 and removed it in v1.3.
	TenantId *string `field:"optional" json:"tenantId" yaml:"tenantId"`
	// Deprecated: CDK for Terraform no longer supports the swift backend. Terraform deprecated swift in v1.2.3 and removed it in v1.3.
	TenantName *string `field:"optional" json:"tenantName" yaml:"tenantName"`
	// Deprecated: CDK for Terraform no longer supports the swift backend. Terraform deprecated swift in v1.2.3 and removed it in v1.3.
	Token *string `field:"optional" json:"token" yaml:"token"`
	// Deprecated: CDK for Terraform no longer supports the swift backend. Terraform deprecated swift in v1.2.3 and removed it in v1.3.
	UserDomainId *string `field:"optional" json:"userDomainId" yaml:"userDomainId"`
	// Deprecated: CDK for Terraform no longer supports the swift backend. Terraform deprecated swift in v1.2.3 and removed it in v1.3.
	UserDomainName *string `field:"optional" json:"userDomainName" yaml:"userDomainName"`
	// Deprecated: CDK for Terraform no longer supports the swift backend. Terraform deprecated swift in v1.2.3 and removed it in v1.3.
	UserId *string `field:"optional" json:"userId" yaml:"userId"`
	// Deprecated: CDK for Terraform no longer supports the swift backend. Terraform deprecated swift in v1.2.3 and removed it in v1.3.
	UserName *string `field:"optional" json:"userName" yaml:"userName"`
}

