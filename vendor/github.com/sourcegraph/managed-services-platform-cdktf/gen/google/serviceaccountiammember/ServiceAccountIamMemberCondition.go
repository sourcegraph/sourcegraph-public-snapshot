package serviceaccountiammember


type ServiceAccountIamMemberCondition struct {
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/service_account_iam_member#expression ServiceAccountIamMember#expression}.
	Expression *string `field:"required" json:"expression" yaml:"expression"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/service_account_iam_member#title ServiceAccountIamMember#title}.
	Title *string `field:"required" json:"title" yaml:"title"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/service_account_iam_member#description ServiceAccountIamMember#description}.
	Description *string `field:"optional" json:"description" yaml:"description"`
}

