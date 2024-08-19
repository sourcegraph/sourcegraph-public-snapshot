package secretmanagersecret

import (
	"github.com/hashicorp/terraform-cdk-go/cdktf"
)

type SecretManagerSecretConfig struct {
	// Experimental.
	Connection interface{} `field:"optional" json:"connection" yaml:"connection"`
	// Experimental.
	Count interface{} `field:"optional" json:"count" yaml:"count"`
	// Experimental.
	DependsOn *[]cdktf.ITerraformDependable `field:"optional" json:"dependsOn" yaml:"dependsOn"`
	// Experimental.
	ForEach cdktf.ITerraformIterator `field:"optional" json:"forEach" yaml:"forEach"`
	// Experimental.
	Lifecycle *cdktf.TerraformResourceLifecycle `field:"optional" json:"lifecycle" yaml:"lifecycle"`
	// Experimental.
	Provider cdktf.TerraformProvider `field:"optional" json:"provider" yaml:"provider"`
	// Experimental.
	Provisioners *[]interface{} `field:"optional" json:"provisioners" yaml:"provisioners"`
	// replication block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/secret_manager_secret#replication SecretManagerSecret#replication}
	Replication *SecretManagerSecretReplication `field:"required" json:"replication" yaml:"replication"`
	// This must be unique within the project.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/secret_manager_secret#secret_id SecretManagerSecret#secret_id}
	SecretId *string `field:"required" json:"secretId" yaml:"secretId"`
	// Custom metadata about the secret.
	//
	// Annotations are distinct from various forms of labels. Annotations exist to allow
	// client tools to store their own state information without requiring a database.
	//
	// Annotation keys must be between 1 and 63 characters long, have a UTF-8 encoding of
	// maximum 128 bytes, begin and end with an alphanumeric character ([a-z0-9A-Z]), and
	// may have dashes (-), underscores (_), dots (.), and alphanumerics in between these
	// symbols.
	//
	// The total size of annotation keys and values must be less than 16KiB.
	//
	// An object containing a list of "key": value pairs. Example:
	// { "name": "wrench", "mass": "1.3kg", "count": "3" }.
	//
	//
	// *Note**: This field is non-authoritative, and will only manage the annotations present in your configuration.
	// Please refer to the field 'effective_annotations' for all of the annotations present on the resource.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/secret_manager_secret#annotations SecretManagerSecret#annotations}
	Annotations *map[string]*string `field:"optional" json:"annotations" yaml:"annotations"`
	// Timestamp in UTC when the Secret is scheduled to expire.
	//
	// This is always provided on output, regardless of what was sent on input.
	// A timestamp in RFC3339 UTC "Zulu" format, with nanosecond resolution and up to nine fractional digits. Examples: "2014-10-02T15:01:23Z" and "2014-10-02T15:01:23.045123456Z".
	// Only one of 'expire_time' or 'ttl' can be provided.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/secret_manager_secret#expire_time SecretManagerSecret#expire_time}
	ExpireTime *string `field:"optional" json:"expireTime" yaml:"expireTime"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/secret_manager_secret#id SecretManagerSecret#id}.
	//
	// Please be aware that the id field is automatically added to all resources in Terraform providers using a Terraform provider SDK version below 2.
	// If you experience problems setting this value it might not be settable. Please take a look at the provider documentation to ensure it should be settable.
	Id *string `field:"optional" json:"id" yaml:"id"`
	// The labels assigned to this Secret.
	//
	// Label keys must be between 1 and 63 characters long, have a UTF-8 encoding of maximum 128 bytes,
	// and must conform to the following PCRE regular expression: [\p{Ll}\p{Lo}][\p{Ll}\p{Lo}\p{N}_-]{0,62}
	//
	// Label values must be between 0 and 63 characters long, have a UTF-8 encoding of maximum 128 bytes,
	// and must conform to the following PCRE regular expression: [\p{Ll}\p{Lo}\p{N}_-]{0,63}
	//
	// No more than 64 labels can be assigned to a given resource.
	//
	// An object containing a list of "key": value pairs. Example:
	// { "name": "wrench", "mass": "1.3kg", "count": "3" }.
	//
	//
	// *Note**: This field is non-authoritative, and will only manage the labels present in your configuration.
	// Please refer to the field 'effective_labels' for all of the labels present on the resource.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/secret_manager_secret#labels SecretManagerSecret#labels}
	Labels *map[string]*string `field:"optional" json:"labels" yaml:"labels"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/secret_manager_secret#project SecretManagerSecret#project}.
	Project *string `field:"optional" json:"project" yaml:"project"`
	// rotation block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/secret_manager_secret#rotation SecretManagerSecret#rotation}
	Rotation *SecretManagerSecretRotation `field:"optional" json:"rotation" yaml:"rotation"`
	// timeouts block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/secret_manager_secret#timeouts SecretManagerSecret#timeouts}
	Timeouts *SecretManagerSecretTimeouts `field:"optional" json:"timeouts" yaml:"timeouts"`
	// topics block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/secret_manager_secret#topics SecretManagerSecret#topics}
	Topics interface{} `field:"optional" json:"topics" yaml:"topics"`
	// The TTL for the Secret.
	//
	// A duration in seconds with up to nine fractional digits, terminated by 's'. Example: "3.5s".
	// Only one of 'ttl' or 'expire_time' can be provided.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/secret_manager_secret#ttl SecretManagerSecret#ttl}
	Ttl *string `field:"optional" json:"ttl" yaml:"ttl"`
	// Mapping from version alias to version name.
	//
	// A version alias is a string with a maximum length of 63 characters and can contain
	// uppercase and lowercase letters, numerals, and the hyphen (-) and underscore ('_')
	// characters. An alias string must start with a letter and cannot be the string
	// 'latest' or 'NEW'. No more than 50 aliases can be assigned to a given secret.
	//
	// An object containing a list of "key": value pairs. Example:
	// { "name": "wrench", "mass": "1.3kg", "count": "3" }.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/secret_manager_secret#version_aliases SecretManagerSecret#version_aliases}
	VersionAliases *map[string]*string `field:"optional" json:"versionAliases" yaml:"versionAliases"`
	// Secret Version TTL after destruction request.
	//
	// This is a part of the delayed delete feature on Secret Version.
	// For secret with versionDestroyTtl>0, version destruction doesn't happen immediately
	// on calling destroy instead the version goes to a disabled state and
	// the actual destruction happens after this TTL expires.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/secret_manager_secret#version_destroy_ttl SecretManagerSecret#version_destroy_ttl}
	VersionDestroyTtl *string `field:"optional" json:"versionDestroyTtl" yaml:"versionDestroyTtl"`
}

