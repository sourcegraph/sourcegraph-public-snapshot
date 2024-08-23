package conversation

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/sourcegraph/managed-services-platform-cdktf/gen/slack/jsii"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/slack/conversation/internal"
)

// Represents a {@link https://registry.terraform.io/providers/pablovarela/slack/1.2.2/docs/resources/conversation slack_conversation}.
type Conversation interface {
	cdktf.TerraformResource
	ActionOnDestroy() *string
	SetActionOnDestroy(val *string)
	ActionOnDestroyInput() *string
	ActionOnUpdatePermanentMembers() *string
	SetActionOnUpdatePermanentMembers(val *string)
	ActionOnUpdatePermanentMembersInput() *string
	AdoptExistingChannel() interface{}
	SetAdoptExistingChannel(val interface{})
	AdoptExistingChannelInput() interface{}
	// Experimental.
	CdktfStack() cdktf.TerraformStack
	// Experimental.
	Connection() interface{}
	// Experimental.
	SetConnection(val interface{})
	// Experimental.
	ConstructNodeMetadata() *map[string]interface{}
	// Experimental.
	Count() interface{}
	// Experimental.
	SetCount(val interface{})
	Created() *float64
	Creator() *string
	// Experimental.
	DependsOn() *[]*string
	// Experimental.
	SetDependsOn(val *[]*string)
	// Experimental.
	ForEach() cdktf.ITerraformIterator
	// Experimental.
	SetForEach(val cdktf.ITerraformIterator)
	// Experimental.
	Fqn() *string
	// Experimental.
	FriendlyUniqueId() *string
	Id() *string
	SetId(val *string)
	IdInput() *string
	IsArchived() interface{}
	SetIsArchived(val interface{})
	IsArchivedInput() interface{}
	IsExtShared() cdktf.IResolvable
	IsGeneral() cdktf.IResolvable
	IsOrgShared() cdktf.IResolvable
	IsPrivate() interface{}
	SetIsPrivate(val interface{})
	IsPrivateInput() interface{}
	IsShared() cdktf.IResolvable
	// Experimental.
	Lifecycle() *cdktf.TerraformResourceLifecycle
	// Experimental.
	SetLifecycle(val *cdktf.TerraformResourceLifecycle)
	Name() *string
	SetName(val *string)
	NameInput() *string
	// The tree node.
	Node() constructs.Node
	PermanentMembers() *[]*string
	SetPermanentMembers(val *[]*string)
	PermanentMembersInput() *[]*string
	// Experimental.
	Provider() cdktf.TerraformProvider
	// Experimental.
	SetProvider(val cdktf.TerraformProvider)
	// Experimental.
	Provisioners() *[]interface{}
	// Experimental.
	SetProvisioners(val *[]interface{})
	Purpose() *string
	SetPurpose(val *string)
	PurposeInput() *string
	// Experimental.
	RawOverrides() interface{}
	// Experimental.
	TerraformGeneratorMetadata() *cdktf.TerraformProviderGeneratorMetadata
	// Experimental.
	TerraformMetaArguments() *map[string]interface{}
	// Experimental.
	TerraformResourceType() *string
	Topic() *string
	SetTopic(val *string)
	TopicInput() *string
	// Experimental.
	AddOverride(path *string, value interface{})
	// Experimental.
	GetAnyMapAttribute(terraformAttribute *string) *map[string]interface{}
	// Experimental.
	GetBooleanAttribute(terraformAttribute *string) cdktf.IResolvable
	// Experimental.
	GetBooleanMapAttribute(terraformAttribute *string) *map[string]*bool
	// Experimental.
	GetListAttribute(terraformAttribute *string) *[]*string
	// Experimental.
	GetNumberAttribute(terraformAttribute *string) *float64
	// Experimental.
	GetNumberListAttribute(terraformAttribute *string) *[]*float64
	// Experimental.
	GetNumberMapAttribute(terraformAttribute *string) *map[string]*float64
	// Experimental.
	GetStringAttribute(terraformAttribute *string) *string
	// Experimental.
	GetStringMapAttribute(terraformAttribute *string) *map[string]*string
	// Experimental.
	InterpolationForAttribute(terraformAttribute *string) cdktf.IResolvable
	// Overrides the auto-generated logical ID with a specific ID.
	// Experimental.
	OverrideLogicalId(newLogicalId *string)
	ResetActionOnDestroy()
	ResetActionOnUpdatePermanentMembers()
	ResetAdoptExistingChannel()
	ResetId()
	ResetIsArchived()
	// Resets a previously passed logical Id to use the auto-generated logical id again.
	// Experimental.
	ResetOverrideLogicalId()
	ResetPermanentMembers()
	ResetPurpose()
	ResetTopic()
	SynthesizeAttributes() *map[string]interface{}
	// Experimental.
	ToMetadata() interface{}
	// Returns a string representation of this construct.
	ToString() *string
	// Adds this resource to the terraform JSON output.
	// Experimental.
	ToTerraform() interface{}
}

// The jsii proxy struct for Conversation
type jsiiProxy_Conversation struct {
	internal.Type__cdktfTerraformResource
}

func (j *jsiiProxy_Conversation) ActionOnDestroy() *string {
	var returns *string
	_jsii_.Get(
		j,
		"actionOnDestroy",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Conversation) ActionOnDestroyInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"actionOnDestroyInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Conversation) ActionOnUpdatePermanentMembers() *string {
	var returns *string
	_jsii_.Get(
		j,
		"actionOnUpdatePermanentMembers",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Conversation) ActionOnUpdatePermanentMembersInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"actionOnUpdatePermanentMembersInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Conversation) AdoptExistingChannel() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"adoptExistingChannel",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Conversation) AdoptExistingChannelInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"adoptExistingChannelInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Conversation) CdktfStack() cdktf.TerraformStack {
	var returns cdktf.TerraformStack
	_jsii_.Get(
		j,
		"cdktfStack",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Conversation) Connection() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"connection",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Conversation) ConstructNodeMetadata() *map[string]interface{} {
	var returns *map[string]interface{}
	_jsii_.Get(
		j,
		"constructNodeMetadata",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Conversation) Count() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"count",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Conversation) Created() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"created",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Conversation) Creator() *string {
	var returns *string
	_jsii_.Get(
		j,
		"creator",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Conversation) DependsOn() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"dependsOn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Conversation) ForEach() cdktf.ITerraformIterator {
	var returns cdktf.ITerraformIterator
	_jsii_.Get(
		j,
		"forEach",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Conversation) Fqn() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fqn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Conversation) FriendlyUniqueId() *string {
	var returns *string
	_jsii_.Get(
		j,
		"friendlyUniqueId",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Conversation) Id() *string {
	var returns *string
	_jsii_.Get(
		j,
		"id",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Conversation) IdInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"idInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Conversation) IsArchived() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"isArchived",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Conversation) IsArchivedInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"isArchivedInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Conversation) IsExtShared() cdktf.IResolvable {
	var returns cdktf.IResolvable
	_jsii_.Get(
		j,
		"isExtShared",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Conversation) IsGeneral() cdktf.IResolvable {
	var returns cdktf.IResolvable
	_jsii_.Get(
		j,
		"isGeneral",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Conversation) IsOrgShared() cdktf.IResolvable {
	var returns cdktf.IResolvable
	_jsii_.Get(
		j,
		"isOrgShared",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Conversation) IsPrivate() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"isPrivate",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Conversation) IsPrivateInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"isPrivateInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Conversation) IsShared() cdktf.IResolvable {
	var returns cdktf.IResolvable
	_jsii_.Get(
		j,
		"isShared",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Conversation) Lifecycle() *cdktf.TerraformResourceLifecycle {
	var returns *cdktf.TerraformResourceLifecycle
	_jsii_.Get(
		j,
		"lifecycle",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Conversation) Name() *string {
	var returns *string
	_jsii_.Get(
		j,
		"name",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Conversation) NameInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"nameInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Conversation) Node() constructs.Node {
	var returns constructs.Node
	_jsii_.Get(
		j,
		"node",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Conversation) PermanentMembers() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"permanentMembers",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Conversation) PermanentMembersInput() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"permanentMembersInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Conversation) Provider() cdktf.TerraformProvider {
	var returns cdktf.TerraformProvider
	_jsii_.Get(
		j,
		"provider",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Conversation) Provisioners() *[]interface{} {
	var returns *[]interface{}
	_jsii_.Get(
		j,
		"provisioners",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Conversation) Purpose() *string {
	var returns *string
	_jsii_.Get(
		j,
		"purpose",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Conversation) PurposeInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"purposeInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Conversation) RawOverrides() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"rawOverrides",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Conversation) TerraformGeneratorMetadata() *cdktf.TerraformProviderGeneratorMetadata {
	var returns *cdktf.TerraformProviderGeneratorMetadata
	_jsii_.Get(
		j,
		"terraformGeneratorMetadata",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Conversation) TerraformMetaArguments() *map[string]interface{} {
	var returns *map[string]interface{}
	_jsii_.Get(
		j,
		"terraformMetaArguments",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Conversation) TerraformResourceType() *string {
	var returns *string
	_jsii_.Get(
		j,
		"terraformResourceType",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Conversation) Topic() *string {
	var returns *string
	_jsii_.Get(
		j,
		"topic",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Conversation) TopicInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"topicInput",
		&returns,
	)
	return returns
}


// Create a new {@link https://registry.terraform.io/providers/pablovarela/slack/1.2.2/docs/resources/conversation slack_conversation} Resource.
func NewConversation(scope constructs.Construct, id *string, config *ConversationConfig) Conversation {
	_init_.Initialize()

	if err := validateNewConversationParameters(scope, id, config); err != nil {
		panic(err)
	}
	j := jsiiProxy_Conversation{}

	_jsii_.Create(
		"@cdktf/provider-slack.conversation.Conversation",
		[]interface{}{scope, id, config},
		&j,
	)

	return &j
}

// Create a new {@link https://registry.terraform.io/providers/pablovarela/slack/1.2.2/docs/resources/conversation slack_conversation} Resource.
func NewConversation_Override(c Conversation, scope constructs.Construct, id *string, config *ConversationConfig) {
	_init_.Initialize()

	_jsii_.Create(
		"@cdktf/provider-slack.conversation.Conversation",
		[]interface{}{scope, id, config},
		c,
	)
}

func (j *jsiiProxy_Conversation)SetActionOnDestroy(val *string) {
	if err := j.validateSetActionOnDestroyParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"actionOnDestroy",
		val,
	)
}

func (j *jsiiProxy_Conversation)SetActionOnUpdatePermanentMembers(val *string) {
	if err := j.validateSetActionOnUpdatePermanentMembersParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"actionOnUpdatePermanentMembers",
		val,
	)
}

func (j *jsiiProxy_Conversation)SetAdoptExistingChannel(val interface{}) {
	if err := j.validateSetAdoptExistingChannelParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"adoptExistingChannel",
		val,
	)
}

func (j *jsiiProxy_Conversation)SetConnection(val interface{}) {
	if err := j.validateSetConnectionParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"connection",
		val,
	)
}

func (j *jsiiProxy_Conversation)SetCount(val interface{}) {
	if err := j.validateSetCountParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"count",
		val,
	)
}

func (j *jsiiProxy_Conversation)SetDependsOn(val *[]*string) {
	_jsii_.Set(
		j,
		"dependsOn",
		val,
	)
}

func (j *jsiiProxy_Conversation)SetForEach(val cdktf.ITerraformIterator) {
	_jsii_.Set(
		j,
		"forEach",
		val,
	)
}

func (j *jsiiProxy_Conversation)SetId(val *string) {
	if err := j.validateSetIdParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"id",
		val,
	)
}

func (j *jsiiProxy_Conversation)SetIsArchived(val interface{}) {
	if err := j.validateSetIsArchivedParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"isArchived",
		val,
	)
}

func (j *jsiiProxy_Conversation)SetIsPrivate(val interface{}) {
	if err := j.validateSetIsPrivateParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"isPrivate",
		val,
	)
}

func (j *jsiiProxy_Conversation)SetLifecycle(val *cdktf.TerraformResourceLifecycle) {
	if err := j.validateSetLifecycleParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"lifecycle",
		val,
	)
}

func (j *jsiiProxy_Conversation)SetName(val *string) {
	if err := j.validateSetNameParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"name",
		val,
	)
}

func (j *jsiiProxy_Conversation)SetPermanentMembers(val *[]*string) {
	if err := j.validateSetPermanentMembersParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"permanentMembers",
		val,
	)
}

func (j *jsiiProxy_Conversation)SetProvider(val cdktf.TerraformProvider) {
	_jsii_.Set(
		j,
		"provider",
		val,
	)
}

func (j *jsiiProxy_Conversation)SetProvisioners(val *[]interface{}) {
	if err := j.validateSetProvisionersParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"provisioners",
		val,
	)
}

func (j *jsiiProxy_Conversation)SetPurpose(val *string) {
	if err := j.validateSetPurposeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"purpose",
		val,
	)
}

func (j *jsiiProxy_Conversation)SetTopic(val *string) {
	if err := j.validateSetTopicParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"topic",
		val,
	)
}

// Checks if `x` is a construct.
//
// Use this method instead of `instanceof` to properly detect `Construct`
// instances, even when the construct library is symlinked.
//
// Explanation: in JavaScript, multiple copies of the `constructs` library on
// disk are seen as independent, completely different libraries. As a
// consequence, the class `Construct` in each copy of the `constructs` library
// is seen as a different class, and an instance of one class will not test as
// `instanceof` the other class. `npm install` will not create installations
// like this, but users may manually symlink construct libraries together or
// use a monorepo tool: in those cases, multiple copies of the `constructs`
// library can be accidentally installed, and `instanceof` will behave
// unpredictably. It is safest to avoid using `instanceof`, and using
// this type-testing method instead.
//
// Returns: true if `x` is an object created from a class which extends `Construct`.
func Conversation_IsConstruct(x interface{}) *bool {
	_init_.Initialize()

	if err := validateConversation_IsConstructParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"@cdktf/provider-slack.conversation.Conversation",
		"isConstruct",
		[]interface{}{x},
		&returns,
	)

	return returns
}

// Experimental.
func Conversation_IsTerraformElement(x interface{}) *bool {
	_init_.Initialize()

	if err := validateConversation_IsTerraformElementParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"@cdktf/provider-slack.conversation.Conversation",
		"isTerraformElement",
		[]interface{}{x},
		&returns,
	)

	return returns
}

// Experimental.
func Conversation_IsTerraformResource(x interface{}) *bool {
	_init_.Initialize()

	if err := validateConversation_IsTerraformResourceParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"@cdktf/provider-slack.conversation.Conversation",
		"isTerraformResource",
		[]interface{}{x},
		&returns,
	)

	return returns
}

func Conversation_TfResourceType() *string {
	_init_.Initialize()
	var returns *string
	_jsii_.StaticGet(
		"@cdktf/provider-slack.conversation.Conversation",
		"tfResourceType",
		&returns,
	)
	return returns
}

func (c *jsiiProxy_Conversation) AddOverride(path *string, value interface{}) {
	if err := c.validateAddOverrideParameters(path, value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"addOverride",
		[]interface{}{path, value},
	)
}

func (c *jsiiProxy_Conversation) GetAnyMapAttribute(terraformAttribute *string) *map[string]interface{} {
	if err := c.validateGetAnyMapAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *map[string]interface{}

	_jsii_.Invoke(
		c,
		"getAnyMapAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (c *jsiiProxy_Conversation) GetBooleanAttribute(terraformAttribute *string) cdktf.IResolvable {
	if err := c.validateGetBooleanAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		c,
		"getBooleanAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (c *jsiiProxy_Conversation) GetBooleanMapAttribute(terraformAttribute *string) *map[string]*bool {
	if err := c.validateGetBooleanMapAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *map[string]*bool

	_jsii_.Invoke(
		c,
		"getBooleanMapAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (c *jsiiProxy_Conversation) GetListAttribute(terraformAttribute *string) *[]*string {
	if err := c.validateGetListAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *[]*string

	_jsii_.Invoke(
		c,
		"getListAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (c *jsiiProxy_Conversation) GetNumberAttribute(terraformAttribute *string) *float64 {
	if err := c.validateGetNumberAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *float64

	_jsii_.Invoke(
		c,
		"getNumberAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (c *jsiiProxy_Conversation) GetNumberListAttribute(terraformAttribute *string) *[]*float64 {
	if err := c.validateGetNumberListAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *[]*float64

	_jsii_.Invoke(
		c,
		"getNumberListAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (c *jsiiProxy_Conversation) GetNumberMapAttribute(terraformAttribute *string) *map[string]*float64 {
	if err := c.validateGetNumberMapAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *map[string]*float64

	_jsii_.Invoke(
		c,
		"getNumberMapAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (c *jsiiProxy_Conversation) GetStringAttribute(terraformAttribute *string) *string {
	if err := c.validateGetStringAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.Invoke(
		c,
		"getStringAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (c *jsiiProxy_Conversation) GetStringMapAttribute(terraformAttribute *string) *map[string]*string {
	if err := c.validateGetStringMapAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *map[string]*string

	_jsii_.Invoke(
		c,
		"getStringMapAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (c *jsiiProxy_Conversation) InterpolationForAttribute(terraformAttribute *string) cdktf.IResolvable {
	if err := c.validateInterpolationForAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		c,
		"interpolationForAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (c *jsiiProxy_Conversation) OverrideLogicalId(newLogicalId *string) {
	if err := c.validateOverrideLogicalIdParameters(newLogicalId); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"overrideLogicalId",
		[]interface{}{newLogicalId},
	)
}

func (c *jsiiProxy_Conversation) ResetActionOnDestroy() {
	_jsii_.InvokeVoid(
		c,
		"resetActionOnDestroy",
		nil, // no parameters
	)
}

func (c *jsiiProxy_Conversation) ResetActionOnUpdatePermanentMembers() {
	_jsii_.InvokeVoid(
		c,
		"resetActionOnUpdatePermanentMembers",
		nil, // no parameters
	)
}

func (c *jsiiProxy_Conversation) ResetAdoptExistingChannel() {
	_jsii_.InvokeVoid(
		c,
		"resetAdoptExistingChannel",
		nil, // no parameters
	)
}

func (c *jsiiProxy_Conversation) ResetId() {
	_jsii_.InvokeVoid(
		c,
		"resetId",
		nil, // no parameters
	)
}

func (c *jsiiProxy_Conversation) ResetIsArchived() {
	_jsii_.InvokeVoid(
		c,
		"resetIsArchived",
		nil, // no parameters
	)
}

func (c *jsiiProxy_Conversation) ResetOverrideLogicalId() {
	_jsii_.InvokeVoid(
		c,
		"resetOverrideLogicalId",
		nil, // no parameters
	)
}

func (c *jsiiProxy_Conversation) ResetPermanentMembers() {
	_jsii_.InvokeVoid(
		c,
		"resetPermanentMembers",
		nil, // no parameters
	)
}

func (c *jsiiProxy_Conversation) ResetPurpose() {
	_jsii_.InvokeVoid(
		c,
		"resetPurpose",
		nil, // no parameters
	)
}

func (c *jsiiProxy_Conversation) ResetTopic() {
	_jsii_.InvokeVoid(
		c,
		"resetTopic",
		nil, // no parameters
	)
}

func (c *jsiiProxy_Conversation) SynthesizeAttributes() *map[string]interface{} {
	var returns *map[string]interface{}

	_jsii_.Invoke(
		c,
		"synthesizeAttributes",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (c *jsiiProxy_Conversation) ToMetadata() interface{} {
	var returns interface{}

	_jsii_.Invoke(
		c,
		"toMetadata",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (c *jsiiProxy_Conversation) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		c,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (c *jsiiProxy_Conversation) ToTerraform() interface{} {
	var returns interface{}

	_jsii_.Invoke(
		c,
		"toTerraform",
		nil, // no parameters
		&returns,
	)

	return returns
}

