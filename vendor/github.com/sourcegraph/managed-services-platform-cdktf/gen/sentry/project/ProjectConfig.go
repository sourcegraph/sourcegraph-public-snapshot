package project

import (
	"github.com/hashicorp/terraform-cdk-go/cdktf"
)

type ProjectConfig struct {
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
	// The name for the project.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/jianyuan/sentry/0.12.3/docs/resources/project#name Project#name}
	Name *string `field:"required" json:"name" yaml:"name"`
	// The slug of the organization the project belongs to.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/jianyuan/sentry/0.12.3/docs/resources/project#organization Project#organization}
	Organization *string `field:"required" json:"organization" yaml:"organization"`
	// Whether to create a default key.
	//
	// By default, Sentry will create a key for you. If you wish to manage keys manually, set this to false and create keys using the `sentry_key` resource.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/jianyuan/sentry/0.12.3/docs/resources/project#default_key Project#default_key}
	DefaultKey interface{} `field:"optional" json:"defaultKey" yaml:"defaultKey"`
	// Whether to create a default issue alert.
	//
	// Defaults to true where the behavior is to alert the user on every new issue.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/jianyuan/sentry/0.12.3/docs/resources/project#default_rules Project#default_rules}
	DefaultRules interface{} `field:"optional" json:"defaultRules" yaml:"defaultRules"`
	// The maximum amount of time (in seconds) to wait between scheduling digests for delivery.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/jianyuan/sentry/0.12.3/docs/resources/project#digests_max_delay Project#digests_max_delay}
	DigestsMaxDelay *float64 `field:"optional" json:"digestsMaxDelay" yaml:"digestsMaxDelay"`
	// The minimum amount of time (in seconds) to wait between scheduling digests for delivery after the initial scheduling.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/jianyuan/sentry/0.12.3/docs/resources/project#digests_min_delay Project#digests_min_delay}
	DigestsMinDelay *float64 `field:"optional" json:"digestsMinDelay" yaml:"digestsMinDelay"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/jianyuan/sentry/0.12.3/docs/resources/project#id Project#id}.
	//
	// Please be aware that the id field is automatically added to all resources in Terraform providers using a Terraform provider SDK version below 2.
	// If you experience problems setting this value it might not be settable. Please take a look at the provider documentation to ensure it should be settable.
	Id *string `field:"optional" json:"id" yaml:"id"`
	// The optional platform for this project.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/jianyuan/sentry/0.12.3/docs/resources/project#platform Project#platform}
	Platform *string `field:"optional" json:"platform" yaml:"platform"`
	// Hours in which an issue is automatically resolve if not seen after this amount of time.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/jianyuan/sentry/0.12.3/docs/resources/project#resolve_age Project#resolve_age}
	ResolveAge *float64 `field:"optional" json:"resolveAge" yaml:"resolveAge"`
	// The optional slug for this project.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/jianyuan/sentry/0.12.3/docs/resources/project#slug Project#slug}
	Slug *string `field:"optional" json:"slug" yaml:"slug"`
	// The slug of the team to create the project for. **Deprecated** Use `teams` instead.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/jianyuan/sentry/0.12.3/docs/resources/project#team Project#team}
	Team *string `field:"optional" json:"team" yaml:"team"`
	// The slugs of the teams to create the project for.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/jianyuan/sentry/0.12.3/docs/resources/project#teams Project#teams}
	Teams *[]*string `field:"optional" json:"teams" yaml:"teams"`
}

