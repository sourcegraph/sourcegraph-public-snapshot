package conversation

import (
	"github.com/hashicorp/terraform-cdk-go/cdktf"
)

type ConversationConfig struct {
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
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/pablovarela/slack/1.2.2/docs/resources/conversation#is_private Conversation#is_private}.
	IsPrivate interface{} `field:"required" json:"isPrivate" yaml:"isPrivate"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/pablovarela/slack/1.2.2/docs/resources/conversation#name Conversation#name}.
	Name *string `field:"required" json:"name" yaml:"name"`
	// Either of none or archive.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/pablovarela/slack/1.2.2/docs/resources/conversation#action_on_destroy Conversation#action_on_destroy}
	ActionOnDestroy *string `field:"optional" json:"actionOnDestroy" yaml:"actionOnDestroy"`
	// Either of none or kick.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/pablovarela/slack/1.2.2/docs/resources/conversation#action_on_update_permanent_members Conversation#action_on_update_permanent_members}
	ActionOnUpdatePermanentMembers *string `field:"optional" json:"actionOnUpdatePermanentMembers" yaml:"actionOnUpdatePermanentMembers"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/pablovarela/slack/1.2.2/docs/resources/conversation#adopt_existing_channel Conversation#adopt_existing_channel}.
	AdoptExistingChannel interface{} `field:"optional" json:"adoptExistingChannel" yaml:"adoptExistingChannel"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/pablovarela/slack/1.2.2/docs/resources/conversation#id Conversation#id}.
	//
	// Please be aware that the id field is automatically added to all resources in Terraform providers using a Terraform provider SDK version below 2.
	// If you experience problems setting this value it might not be settable. Please take a look at the provider documentation to ensure it should be settable.
	Id *string `field:"optional" json:"id" yaml:"id"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/pablovarela/slack/1.2.2/docs/resources/conversation#is_archived Conversation#is_archived}.
	IsArchived interface{} `field:"optional" json:"isArchived" yaml:"isArchived"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/pablovarela/slack/1.2.2/docs/resources/conversation#permanent_members Conversation#permanent_members}.
	PermanentMembers *[]*string `field:"optional" json:"permanentMembers" yaml:"permanentMembers"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/pablovarela/slack/1.2.2/docs/resources/conversation#purpose Conversation#purpose}.
	Purpose *string `field:"optional" json:"purpose" yaml:"purpose"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/pablovarela/slack/1.2.2/docs/resources/conversation#topic Conversation#topic}.
	Topic *string `field:"optional" json:"topic" yaml:"topic"`
}

