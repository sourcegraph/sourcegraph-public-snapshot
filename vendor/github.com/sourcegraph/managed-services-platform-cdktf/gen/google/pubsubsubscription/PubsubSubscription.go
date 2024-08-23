package pubsubsubscription

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/sourcegraph/managed-services-platform-cdktf/gen/google/jsii"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/pubsubsubscription/internal"
)

// Represents a {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/pubsub_subscription google_pubsub_subscription}.
type PubsubSubscription interface {
	cdktf.TerraformResource
	AckDeadlineSeconds() *float64
	SetAckDeadlineSeconds(val *float64)
	AckDeadlineSecondsInput() *float64
	BigqueryConfig() PubsubSubscriptionBigqueryConfigOutputReference
	BigqueryConfigInput() *PubsubSubscriptionBigqueryConfig
	// Experimental.
	CdktfStack() cdktf.TerraformStack
	CloudStorageConfig() PubsubSubscriptionCloudStorageConfigOutputReference
	CloudStorageConfigInput() *PubsubSubscriptionCloudStorageConfig
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
	DeadLetterPolicy() PubsubSubscriptionDeadLetterPolicyOutputReference
	DeadLetterPolicyInput() *PubsubSubscriptionDeadLetterPolicy
	// Experimental.
	DependsOn() *[]*string
	// Experimental.
	SetDependsOn(val *[]*string)
	EffectiveLabels() cdktf.StringMap
	EnableExactlyOnceDelivery() interface{}
	SetEnableExactlyOnceDelivery(val interface{})
	EnableExactlyOnceDeliveryInput() interface{}
	EnableMessageOrdering() interface{}
	SetEnableMessageOrdering(val interface{})
	EnableMessageOrderingInput() interface{}
	ExpirationPolicy() PubsubSubscriptionExpirationPolicyOutputReference
	ExpirationPolicyInput() *PubsubSubscriptionExpirationPolicy
	Filter() *string
	SetFilter(val *string)
	FilterInput() *string
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
	Labels() *map[string]*string
	SetLabels(val *map[string]*string)
	LabelsInput() *map[string]*string
	// Experimental.
	Lifecycle() *cdktf.TerraformResourceLifecycle
	// Experimental.
	SetLifecycle(val *cdktf.TerraformResourceLifecycle)
	MessageRetentionDuration() *string
	SetMessageRetentionDuration(val *string)
	MessageRetentionDurationInput() *string
	Name() *string
	SetName(val *string)
	NameInput() *string
	// The tree node.
	Node() constructs.Node
	Project() *string
	SetProject(val *string)
	ProjectInput() *string
	// Experimental.
	Provider() cdktf.TerraformProvider
	// Experimental.
	SetProvider(val cdktf.TerraformProvider)
	// Experimental.
	Provisioners() *[]interface{}
	// Experimental.
	SetProvisioners(val *[]interface{})
	PushConfig() PubsubSubscriptionPushConfigOutputReference
	PushConfigInput() *PubsubSubscriptionPushConfig
	// Experimental.
	RawOverrides() interface{}
	RetainAckedMessages() interface{}
	SetRetainAckedMessages(val interface{})
	RetainAckedMessagesInput() interface{}
	RetryPolicy() PubsubSubscriptionRetryPolicyOutputReference
	RetryPolicyInput() *PubsubSubscriptionRetryPolicy
	// Experimental.
	TerraformGeneratorMetadata() *cdktf.TerraformProviderGeneratorMetadata
	TerraformLabels() cdktf.StringMap
	// Experimental.
	TerraformMetaArguments() *map[string]interface{}
	// Experimental.
	TerraformResourceType() *string
	Timeouts() PubsubSubscriptionTimeoutsOutputReference
	TimeoutsInput() interface{}
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
	PutBigqueryConfig(value *PubsubSubscriptionBigqueryConfig)
	PutCloudStorageConfig(value *PubsubSubscriptionCloudStorageConfig)
	PutDeadLetterPolicy(value *PubsubSubscriptionDeadLetterPolicy)
	PutExpirationPolicy(value *PubsubSubscriptionExpirationPolicy)
	PutPushConfig(value *PubsubSubscriptionPushConfig)
	PutRetryPolicy(value *PubsubSubscriptionRetryPolicy)
	PutTimeouts(value *PubsubSubscriptionTimeouts)
	ResetAckDeadlineSeconds()
	ResetBigqueryConfig()
	ResetCloudStorageConfig()
	ResetDeadLetterPolicy()
	ResetEnableExactlyOnceDelivery()
	ResetEnableMessageOrdering()
	ResetExpirationPolicy()
	ResetFilter()
	ResetId()
	ResetLabels()
	ResetMessageRetentionDuration()
	// Resets a previously passed logical Id to use the auto-generated logical id again.
	// Experimental.
	ResetOverrideLogicalId()
	ResetProject()
	ResetPushConfig()
	ResetRetainAckedMessages()
	ResetRetryPolicy()
	ResetTimeouts()
	SynthesizeAttributes() *map[string]interface{}
	// Experimental.
	ToMetadata() interface{}
	// Returns a string representation of this construct.
	ToString() *string
	// Adds this resource to the terraform JSON output.
	// Experimental.
	ToTerraform() interface{}
}

// The jsii proxy struct for PubsubSubscription
type jsiiProxy_PubsubSubscription struct {
	internal.Type__cdktfTerraformResource
}

func (j *jsiiProxy_PubsubSubscription) AckDeadlineSeconds() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"ackDeadlineSeconds",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscription) AckDeadlineSecondsInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"ackDeadlineSecondsInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscription) BigqueryConfig() PubsubSubscriptionBigqueryConfigOutputReference {
	var returns PubsubSubscriptionBigqueryConfigOutputReference
	_jsii_.Get(
		j,
		"bigqueryConfig",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscription) BigqueryConfigInput() *PubsubSubscriptionBigqueryConfig {
	var returns *PubsubSubscriptionBigqueryConfig
	_jsii_.Get(
		j,
		"bigqueryConfigInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscription) CdktfStack() cdktf.TerraformStack {
	var returns cdktf.TerraformStack
	_jsii_.Get(
		j,
		"cdktfStack",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscription) CloudStorageConfig() PubsubSubscriptionCloudStorageConfigOutputReference {
	var returns PubsubSubscriptionCloudStorageConfigOutputReference
	_jsii_.Get(
		j,
		"cloudStorageConfig",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscription) CloudStorageConfigInput() *PubsubSubscriptionCloudStorageConfig {
	var returns *PubsubSubscriptionCloudStorageConfig
	_jsii_.Get(
		j,
		"cloudStorageConfigInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscription) Connection() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"connection",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscription) ConstructNodeMetadata() *map[string]interface{} {
	var returns *map[string]interface{}
	_jsii_.Get(
		j,
		"constructNodeMetadata",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscription) Count() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"count",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscription) DeadLetterPolicy() PubsubSubscriptionDeadLetterPolicyOutputReference {
	var returns PubsubSubscriptionDeadLetterPolicyOutputReference
	_jsii_.Get(
		j,
		"deadLetterPolicy",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscription) DeadLetterPolicyInput() *PubsubSubscriptionDeadLetterPolicy {
	var returns *PubsubSubscriptionDeadLetterPolicy
	_jsii_.Get(
		j,
		"deadLetterPolicyInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscription) DependsOn() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"dependsOn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscription) EffectiveLabels() cdktf.StringMap {
	var returns cdktf.StringMap
	_jsii_.Get(
		j,
		"effectiveLabels",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscription) EnableExactlyOnceDelivery() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"enableExactlyOnceDelivery",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscription) EnableExactlyOnceDeliveryInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"enableExactlyOnceDeliveryInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscription) EnableMessageOrdering() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"enableMessageOrdering",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscription) EnableMessageOrderingInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"enableMessageOrderingInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscription) ExpirationPolicy() PubsubSubscriptionExpirationPolicyOutputReference {
	var returns PubsubSubscriptionExpirationPolicyOutputReference
	_jsii_.Get(
		j,
		"expirationPolicy",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscription) ExpirationPolicyInput() *PubsubSubscriptionExpirationPolicy {
	var returns *PubsubSubscriptionExpirationPolicy
	_jsii_.Get(
		j,
		"expirationPolicyInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscription) Filter() *string {
	var returns *string
	_jsii_.Get(
		j,
		"filter",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscription) FilterInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"filterInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscription) ForEach() cdktf.ITerraformIterator {
	var returns cdktf.ITerraformIterator
	_jsii_.Get(
		j,
		"forEach",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscription) Fqn() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fqn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscription) FriendlyUniqueId() *string {
	var returns *string
	_jsii_.Get(
		j,
		"friendlyUniqueId",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscription) Id() *string {
	var returns *string
	_jsii_.Get(
		j,
		"id",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscription) IdInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"idInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscription) Labels() *map[string]*string {
	var returns *map[string]*string
	_jsii_.Get(
		j,
		"labels",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscription) LabelsInput() *map[string]*string {
	var returns *map[string]*string
	_jsii_.Get(
		j,
		"labelsInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscription) Lifecycle() *cdktf.TerraformResourceLifecycle {
	var returns *cdktf.TerraformResourceLifecycle
	_jsii_.Get(
		j,
		"lifecycle",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscription) MessageRetentionDuration() *string {
	var returns *string
	_jsii_.Get(
		j,
		"messageRetentionDuration",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscription) MessageRetentionDurationInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"messageRetentionDurationInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscription) Name() *string {
	var returns *string
	_jsii_.Get(
		j,
		"name",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscription) NameInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"nameInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscription) Node() constructs.Node {
	var returns constructs.Node
	_jsii_.Get(
		j,
		"node",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscription) Project() *string {
	var returns *string
	_jsii_.Get(
		j,
		"project",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscription) ProjectInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"projectInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscription) Provider() cdktf.TerraformProvider {
	var returns cdktf.TerraformProvider
	_jsii_.Get(
		j,
		"provider",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscription) Provisioners() *[]interface{} {
	var returns *[]interface{}
	_jsii_.Get(
		j,
		"provisioners",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscription) PushConfig() PubsubSubscriptionPushConfigOutputReference {
	var returns PubsubSubscriptionPushConfigOutputReference
	_jsii_.Get(
		j,
		"pushConfig",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscription) PushConfigInput() *PubsubSubscriptionPushConfig {
	var returns *PubsubSubscriptionPushConfig
	_jsii_.Get(
		j,
		"pushConfigInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscription) RawOverrides() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"rawOverrides",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscription) RetainAckedMessages() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"retainAckedMessages",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscription) RetainAckedMessagesInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"retainAckedMessagesInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscription) RetryPolicy() PubsubSubscriptionRetryPolicyOutputReference {
	var returns PubsubSubscriptionRetryPolicyOutputReference
	_jsii_.Get(
		j,
		"retryPolicy",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscription) RetryPolicyInput() *PubsubSubscriptionRetryPolicy {
	var returns *PubsubSubscriptionRetryPolicy
	_jsii_.Get(
		j,
		"retryPolicyInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscription) TerraformGeneratorMetadata() *cdktf.TerraformProviderGeneratorMetadata {
	var returns *cdktf.TerraformProviderGeneratorMetadata
	_jsii_.Get(
		j,
		"terraformGeneratorMetadata",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscription) TerraformLabels() cdktf.StringMap {
	var returns cdktf.StringMap
	_jsii_.Get(
		j,
		"terraformLabels",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscription) TerraformMetaArguments() *map[string]interface{} {
	var returns *map[string]interface{}
	_jsii_.Get(
		j,
		"terraformMetaArguments",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscription) TerraformResourceType() *string {
	var returns *string
	_jsii_.Get(
		j,
		"terraformResourceType",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscription) Timeouts() PubsubSubscriptionTimeoutsOutputReference {
	var returns PubsubSubscriptionTimeoutsOutputReference
	_jsii_.Get(
		j,
		"timeouts",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscription) TimeoutsInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"timeoutsInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscription) Topic() *string {
	var returns *string
	_jsii_.Get(
		j,
		"topic",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscription) TopicInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"topicInput",
		&returns,
	)
	return returns
}


// Create a new {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/pubsub_subscription google_pubsub_subscription} Resource.
func NewPubsubSubscription(scope constructs.Construct, id *string, config *PubsubSubscriptionConfig) PubsubSubscription {
	_init_.Initialize()

	if err := validateNewPubsubSubscriptionParameters(scope, id, config); err != nil {
		panic(err)
	}
	j := jsiiProxy_PubsubSubscription{}

	_jsii_.Create(
		"@cdktf/provider-google.pubsubSubscription.PubsubSubscription",
		[]interface{}{scope, id, config},
		&j,
	)

	return &j
}

// Create a new {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/pubsub_subscription google_pubsub_subscription} Resource.
func NewPubsubSubscription_Override(p PubsubSubscription, scope constructs.Construct, id *string, config *PubsubSubscriptionConfig) {
	_init_.Initialize()

	_jsii_.Create(
		"@cdktf/provider-google.pubsubSubscription.PubsubSubscription",
		[]interface{}{scope, id, config},
		p,
	)
}

func (j *jsiiProxy_PubsubSubscription)SetAckDeadlineSeconds(val *float64) {
	if err := j.validateSetAckDeadlineSecondsParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"ackDeadlineSeconds",
		val,
	)
}

func (j *jsiiProxy_PubsubSubscription)SetConnection(val interface{}) {
	if err := j.validateSetConnectionParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"connection",
		val,
	)
}

func (j *jsiiProxy_PubsubSubscription)SetCount(val interface{}) {
	if err := j.validateSetCountParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"count",
		val,
	)
}

func (j *jsiiProxy_PubsubSubscription)SetDependsOn(val *[]*string) {
	_jsii_.Set(
		j,
		"dependsOn",
		val,
	)
}

func (j *jsiiProxy_PubsubSubscription)SetEnableExactlyOnceDelivery(val interface{}) {
	if err := j.validateSetEnableExactlyOnceDeliveryParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"enableExactlyOnceDelivery",
		val,
	)
}

func (j *jsiiProxy_PubsubSubscription)SetEnableMessageOrdering(val interface{}) {
	if err := j.validateSetEnableMessageOrderingParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"enableMessageOrdering",
		val,
	)
}

func (j *jsiiProxy_PubsubSubscription)SetFilter(val *string) {
	if err := j.validateSetFilterParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"filter",
		val,
	)
}

func (j *jsiiProxy_PubsubSubscription)SetForEach(val cdktf.ITerraformIterator) {
	_jsii_.Set(
		j,
		"forEach",
		val,
	)
}

func (j *jsiiProxy_PubsubSubscription)SetId(val *string) {
	if err := j.validateSetIdParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"id",
		val,
	)
}

func (j *jsiiProxy_PubsubSubscription)SetLabels(val *map[string]*string) {
	if err := j.validateSetLabelsParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"labels",
		val,
	)
}

func (j *jsiiProxy_PubsubSubscription)SetLifecycle(val *cdktf.TerraformResourceLifecycle) {
	if err := j.validateSetLifecycleParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"lifecycle",
		val,
	)
}

func (j *jsiiProxy_PubsubSubscription)SetMessageRetentionDuration(val *string) {
	if err := j.validateSetMessageRetentionDurationParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"messageRetentionDuration",
		val,
	)
}

func (j *jsiiProxy_PubsubSubscription)SetName(val *string) {
	if err := j.validateSetNameParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"name",
		val,
	)
}

func (j *jsiiProxy_PubsubSubscription)SetProject(val *string) {
	if err := j.validateSetProjectParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"project",
		val,
	)
}

func (j *jsiiProxy_PubsubSubscription)SetProvider(val cdktf.TerraformProvider) {
	_jsii_.Set(
		j,
		"provider",
		val,
	)
}

func (j *jsiiProxy_PubsubSubscription)SetProvisioners(val *[]interface{}) {
	if err := j.validateSetProvisionersParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"provisioners",
		val,
	)
}

func (j *jsiiProxy_PubsubSubscription)SetRetainAckedMessages(val interface{}) {
	if err := j.validateSetRetainAckedMessagesParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"retainAckedMessages",
		val,
	)
}

func (j *jsiiProxy_PubsubSubscription)SetTopic(val *string) {
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
func PubsubSubscription_IsConstruct(x interface{}) *bool {
	_init_.Initialize()

	if err := validatePubsubSubscription_IsConstructParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"@cdktf/provider-google.pubsubSubscription.PubsubSubscription",
		"isConstruct",
		[]interface{}{x},
		&returns,
	)

	return returns
}

// Experimental.
func PubsubSubscription_IsTerraformElement(x interface{}) *bool {
	_init_.Initialize()

	if err := validatePubsubSubscription_IsTerraformElementParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"@cdktf/provider-google.pubsubSubscription.PubsubSubscription",
		"isTerraformElement",
		[]interface{}{x},
		&returns,
	)

	return returns
}

// Experimental.
func PubsubSubscription_IsTerraformResource(x interface{}) *bool {
	_init_.Initialize()

	if err := validatePubsubSubscription_IsTerraformResourceParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"@cdktf/provider-google.pubsubSubscription.PubsubSubscription",
		"isTerraformResource",
		[]interface{}{x},
		&returns,
	)

	return returns
}

func PubsubSubscription_TfResourceType() *string {
	_init_.Initialize()
	var returns *string
	_jsii_.StaticGet(
		"@cdktf/provider-google.pubsubSubscription.PubsubSubscription",
		"tfResourceType",
		&returns,
	)
	return returns
}

func (p *jsiiProxy_PubsubSubscription) AddOverride(path *string, value interface{}) {
	if err := p.validateAddOverrideParameters(path, value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		p,
		"addOverride",
		[]interface{}{path, value},
	)
}

func (p *jsiiProxy_PubsubSubscription) GetAnyMapAttribute(terraformAttribute *string) *map[string]interface{} {
	if err := p.validateGetAnyMapAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *map[string]interface{}

	_jsii_.Invoke(
		p,
		"getAnyMapAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (p *jsiiProxy_PubsubSubscription) GetBooleanAttribute(terraformAttribute *string) cdktf.IResolvable {
	if err := p.validateGetBooleanAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		p,
		"getBooleanAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (p *jsiiProxy_PubsubSubscription) GetBooleanMapAttribute(terraformAttribute *string) *map[string]*bool {
	if err := p.validateGetBooleanMapAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *map[string]*bool

	_jsii_.Invoke(
		p,
		"getBooleanMapAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (p *jsiiProxy_PubsubSubscription) GetListAttribute(terraformAttribute *string) *[]*string {
	if err := p.validateGetListAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *[]*string

	_jsii_.Invoke(
		p,
		"getListAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (p *jsiiProxy_PubsubSubscription) GetNumberAttribute(terraformAttribute *string) *float64 {
	if err := p.validateGetNumberAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *float64

	_jsii_.Invoke(
		p,
		"getNumberAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (p *jsiiProxy_PubsubSubscription) GetNumberListAttribute(terraformAttribute *string) *[]*float64 {
	if err := p.validateGetNumberListAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *[]*float64

	_jsii_.Invoke(
		p,
		"getNumberListAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (p *jsiiProxy_PubsubSubscription) GetNumberMapAttribute(terraformAttribute *string) *map[string]*float64 {
	if err := p.validateGetNumberMapAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *map[string]*float64

	_jsii_.Invoke(
		p,
		"getNumberMapAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (p *jsiiProxy_PubsubSubscription) GetStringAttribute(terraformAttribute *string) *string {
	if err := p.validateGetStringAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.Invoke(
		p,
		"getStringAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (p *jsiiProxy_PubsubSubscription) GetStringMapAttribute(terraformAttribute *string) *map[string]*string {
	if err := p.validateGetStringMapAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *map[string]*string

	_jsii_.Invoke(
		p,
		"getStringMapAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (p *jsiiProxy_PubsubSubscription) InterpolationForAttribute(terraformAttribute *string) cdktf.IResolvable {
	if err := p.validateInterpolationForAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		p,
		"interpolationForAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (p *jsiiProxy_PubsubSubscription) OverrideLogicalId(newLogicalId *string) {
	if err := p.validateOverrideLogicalIdParameters(newLogicalId); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		p,
		"overrideLogicalId",
		[]interface{}{newLogicalId},
	)
}

func (p *jsiiProxy_PubsubSubscription) PutBigqueryConfig(value *PubsubSubscriptionBigqueryConfig) {
	if err := p.validatePutBigqueryConfigParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		p,
		"putBigqueryConfig",
		[]interface{}{value},
	)
}

func (p *jsiiProxy_PubsubSubscription) PutCloudStorageConfig(value *PubsubSubscriptionCloudStorageConfig) {
	if err := p.validatePutCloudStorageConfigParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		p,
		"putCloudStorageConfig",
		[]interface{}{value},
	)
}

func (p *jsiiProxy_PubsubSubscription) PutDeadLetterPolicy(value *PubsubSubscriptionDeadLetterPolicy) {
	if err := p.validatePutDeadLetterPolicyParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		p,
		"putDeadLetterPolicy",
		[]interface{}{value},
	)
}

func (p *jsiiProxy_PubsubSubscription) PutExpirationPolicy(value *PubsubSubscriptionExpirationPolicy) {
	if err := p.validatePutExpirationPolicyParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		p,
		"putExpirationPolicy",
		[]interface{}{value},
	)
}

func (p *jsiiProxy_PubsubSubscription) PutPushConfig(value *PubsubSubscriptionPushConfig) {
	if err := p.validatePutPushConfigParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		p,
		"putPushConfig",
		[]interface{}{value},
	)
}

func (p *jsiiProxy_PubsubSubscription) PutRetryPolicy(value *PubsubSubscriptionRetryPolicy) {
	if err := p.validatePutRetryPolicyParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		p,
		"putRetryPolicy",
		[]interface{}{value},
	)
}

func (p *jsiiProxy_PubsubSubscription) PutTimeouts(value *PubsubSubscriptionTimeouts) {
	if err := p.validatePutTimeoutsParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		p,
		"putTimeouts",
		[]interface{}{value},
	)
}

func (p *jsiiProxy_PubsubSubscription) ResetAckDeadlineSeconds() {
	_jsii_.InvokeVoid(
		p,
		"resetAckDeadlineSeconds",
		nil, // no parameters
	)
}

func (p *jsiiProxy_PubsubSubscription) ResetBigqueryConfig() {
	_jsii_.InvokeVoid(
		p,
		"resetBigqueryConfig",
		nil, // no parameters
	)
}

func (p *jsiiProxy_PubsubSubscription) ResetCloudStorageConfig() {
	_jsii_.InvokeVoid(
		p,
		"resetCloudStorageConfig",
		nil, // no parameters
	)
}

func (p *jsiiProxy_PubsubSubscription) ResetDeadLetterPolicy() {
	_jsii_.InvokeVoid(
		p,
		"resetDeadLetterPolicy",
		nil, // no parameters
	)
}

func (p *jsiiProxy_PubsubSubscription) ResetEnableExactlyOnceDelivery() {
	_jsii_.InvokeVoid(
		p,
		"resetEnableExactlyOnceDelivery",
		nil, // no parameters
	)
}

func (p *jsiiProxy_PubsubSubscription) ResetEnableMessageOrdering() {
	_jsii_.InvokeVoid(
		p,
		"resetEnableMessageOrdering",
		nil, // no parameters
	)
}

func (p *jsiiProxy_PubsubSubscription) ResetExpirationPolicy() {
	_jsii_.InvokeVoid(
		p,
		"resetExpirationPolicy",
		nil, // no parameters
	)
}

func (p *jsiiProxy_PubsubSubscription) ResetFilter() {
	_jsii_.InvokeVoid(
		p,
		"resetFilter",
		nil, // no parameters
	)
}

func (p *jsiiProxy_PubsubSubscription) ResetId() {
	_jsii_.InvokeVoid(
		p,
		"resetId",
		nil, // no parameters
	)
}

func (p *jsiiProxy_PubsubSubscription) ResetLabels() {
	_jsii_.InvokeVoid(
		p,
		"resetLabels",
		nil, // no parameters
	)
}

func (p *jsiiProxy_PubsubSubscription) ResetMessageRetentionDuration() {
	_jsii_.InvokeVoid(
		p,
		"resetMessageRetentionDuration",
		nil, // no parameters
	)
}

func (p *jsiiProxy_PubsubSubscription) ResetOverrideLogicalId() {
	_jsii_.InvokeVoid(
		p,
		"resetOverrideLogicalId",
		nil, // no parameters
	)
}

func (p *jsiiProxy_PubsubSubscription) ResetProject() {
	_jsii_.InvokeVoid(
		p,
		"resetProject",
		nil, // no parameters
	)
}

func (p *jsiiProxy_PubsubSubscription) ResetPushConfig() {
	_jsii_.InvokeVoid(
		p,
		"resetPushConfig",
		nil, // no parameters
	)
}

func (p *jsiiProxy_PubsubSubscription) ResetRetainAckedMessages() {
	_jsii_.InvokeVoid(
		p,
		"resetRetainAckedMessages",
		nil, // no parameters
	)
}

func (p *jsiiProxy_PubsubSubscription) ResetRetryPolicy() {
	_jsii_.InvokeVoid(
		p,
		"resetRetryPolicy",
		nil, // no parameters
	)
}

func (p *jsiiProxy_PubsubSubscription) ResetTimeouts() {
	_jsii_.InvokeVoid(
		p,
		"resetTimeouts",
		nil, // no parameters
	)
}

func (p *jsiiProxy_PubsubSubscription) SynthesizeAttributes() *map[string]interface{} {
	var returns *map[string]interface{}

	_jsii_.Invoke(
		p,
		"synthesizeAttributes",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (p *jsiiProxy_PubsubSubscription) ToMetadata() interface{} {
	var returns interface{}

	_jsii_.Invoke(
		p,
		"toMetadata",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (p *jsiiProxy_PubsubSubscription) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		p,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (p *jsiiProxy_PubsubSubscription) ToTerraform() interface{} {
	var returns interface{}

	_jsii_.Invoke(
		p,
		"toTerraform",
		nil, // no parameters
		&returns,
	)

	return returns
}

