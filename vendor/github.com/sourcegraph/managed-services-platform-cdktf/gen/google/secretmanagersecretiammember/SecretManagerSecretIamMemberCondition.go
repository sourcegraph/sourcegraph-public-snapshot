package secretmanagersecretiammember


type SecretManagerSecretIamMemberCondition struct {
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/secret_manager_secret_iam_member#expression SecretManagerSecretIamMember#expression}.
	Expression *string `field:"required" json:"expression" yaml:"expression"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/secret_manager_secret_iam_member#title SecretManagerSecretIamMember#title}.
	Title *string `field:"required" json:"title" yaml:"title"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/secret_manager_secret_iam_member#description SecretManagerSecretIamMember#description}.
	Description *string `field:"optional" json:"description" yaml:"description"`
}

