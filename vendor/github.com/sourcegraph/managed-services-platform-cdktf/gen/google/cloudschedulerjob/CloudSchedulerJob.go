package cloudschedulerjob

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/sourcegraph/managed-services-platform-cdktf/gen/google/jsii"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/cloudschedulerjob/internal"
)

// Represents a {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_scheduler_job google_cloud_scheduler_job}.
type CloudSchedulerJob interface {
	cdktf.TerraformResource
	AppEngineHttpTarget() CloudSchedulerJobAppEngineHttpTargetOutputReference
	AppEngineHttpTargetInput() *CloudSchedulerJobAppEngineHttpTarget
	AttemptDeadline() *string
	SetAttemptDeadline(val *string)
	AttemptDeadlineInput() *string
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
	// Experimental.
	DependsOn() *[]*string
	// Experimental.
	SetDependsOn(val *[]*string)
	Description() *string
	SetDescription(val *string)
	DescriptionInput() *string
	// Experimental.
	ForEach() cdktf.ITerraformIterator
	// Experimental.
	SetForEach(val cdktf.ITerraformIterator)
	// Experimental.
	Fqn() *string
	// Experimental.
	FriendlyUniqueId() *string
	HttpTarget() CloudSchedulerJobHttpTargetOutputReference
	HttpTargetInput() *CloudSchedulerJobHttpTarget
	Id() *string
	SetId(val *string)
	IdInput() *string
	// Experimental.
	Lifecycle() *cdktf.TerraformResourceLifecycle
	// Experimental.
	SetLifecycle(val *cdktf.TerraformResourceLifecycle)
	Name() *string
	SetName(val *string)
	NameInput() *string
	// The tree node.
	Node() constructs.Node
	Paused() interface{}
	SetPaused(val interface{})
	PausedInput() interface{}
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
	PubsubTarget() CloudSchedulerJobPubsubTargetOutputReference
	PubsubTargetInput() *CloudSchedulerJobPubsubTarget
	// Experimental.
	RawOverrides() interface{}
	Region() *string
	SetRegion(val *string)
	RegionInput() *string
	RetryConfig() CloudSchedulerJobRetryConfigOutputReference
	RetryConfigInput() *CloudSchedulerJobRetryConfig
	Schedule() *string
	SetSchedule(val *string)
	ScheduleInput() *string
	State() *string
	// Experimental.
	TerraformGeneratorMetadata() *cdktf.TerraformProviderGeneratorMetadata
	// Experimental.
	TerraformMetaArguments() *map[string]interface{}
	// Experimental.
	TerraformResourceType() *string
	Timeouts() CloudSchedulerJobTimeoutsOutputReference
	TimeoutsInput() interface{}
	TimeZone() *string
	SetTimeZone(val *string)
	TimeZoneInput() *string
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
	PutAppEngineHttpTarget(value *CloudSchedulerJobAppEngineHttpTarget)
	PutHttpTarget(value *CloudSchedulerJobHttpTarget)
	PutPubsubTarget(value *CloudSchedulerJobPubsubTarget)
	PutRetryConfig(value *CloudSchedulerJobRetryConfig)
	PutTimeouts(value *CloudSchedulerJobTimeouts)
	ResetAppEngineHttpTarget()
	ResetAttemptDeadline()
	ResetDescription()
	ResetHttpTarget()
	ResetId()
	// Resets a previously passed logical Id to use the auto-generated logical id again.
	// Experimental.
	ResetOverrideLogicalId()
	ResetPaused()
	ResetProject()
	ResetPubsubTarget()
	ResetRegion()
	ResetRetryConfig()
	ResetSchedule()
	ResetTimeouts()
	ResetTimeZone()
	SynthesizeAttributes() *map[string]interface{}
	// Experimental.
	ToMetadata() interface{}
	// Returns a string representation of this construct.
	ToString() *string
	// Adds this resource to the terraform JSON output.
	// Experimental.
	ToTerraform() interface{}
}

// The jsii proxy struct for CloudSchedulerJob
type jsiiProxy_CloudSchedulerJob struct {
	internal.Type__cdktfTerraformResource
}

func (j *jsiiProxy_CloudSchedulerJob) AppEngineHttpTarget() CloudSchedulerJobAppEngineHttpTargetOutputReference {
	var returns CloudSchedulerJobAppEngineHttpTargetOutputReference
	_jsii_.Get(
		j,
		"appEngineHttpTarget",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJob) AppEngineHttpTargetInput() *CloudSchedulerJobAppEngineHttpTarget {
	var returns *CloudSchedulerJobAppEngineHttpTarget
	_jsii_.Get(
		j,
		"appEngineHttpTargetInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJob) AttemptDeadline() *string {
	var returns *string
	_jsii_.Get(
		j,
		"attemptDeadline",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJob) AttemptDeadlineInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"attemptDeadlineInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJob) CdktfStack() cdktf.TerraformStack {
	var returns cdktf.TerraformStack
	_jsii_.Get(
		j,
		"cdktfStack",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJob) Connection() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"connection",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJob) ConstructNodeMetadata() *map[string]interface{} {
	var returns *map[string]interface{}
	_jsii_.Get(
		j,
		"constructNodeMetadata",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJob) Count() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"count",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJob) DependsOn() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"dependsOn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJob) Description() *string {
	var returns *string
	_jsii_.Get(
		j,
		"description",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJob) DescriptionInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"descriptionInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJob) ForEach() cdktf.ITerraformIterator {
	var returns cdktf.ITerraformIterator
	_jsii_.Get(
		j,
		"forEach",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJob) Fqn() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fqn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJob) FriendlyUniqueId() *string {
	var returns *string
	_jsii_.Get(
		j,
		"friendlyUniqueId",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJob) HttpTarget() CloudSchedulerJobHttpTargetOutputReference {
	var returns CloudSchedulerJobHttpTargetOutputReference
	_jsii_.Get(
		j,
		"httpTarget",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJob) HttpTargetInput() *CloudSchedulerJobHttpTarget {
	var returns *CloudSchedulerJobHttpTarget
	_jsii_.Get(
		j,
		"httpTargetInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJob) Id() *string {
	var returns *string
	_jsii_.Get(
		j,
		"id",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJob) IdInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"idInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJob) Lifecycle() *cdktf.TerraformResourceLifecycle {
	var returns *cdktf.TerraformResourceLifecycle
	_jsii_.Get(
		j,
		"lifecycle",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJob) Name() *string {
	var returns *string
	_jsii_.Get(
		j,
		"name",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJob) NameInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"nameInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJob) Node() constructs.Node {
	var returns constructs.Node
	_jsii_.Get(
		j,
		"node",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJob) Paused() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"paused",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJob) PausedInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"pausedInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJob) Project() *string {
	var returns *string
	_jsii_.Get(
		j,
		"project",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJob) ProjectInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"projectInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJob) Provider() cdktf.TerraformProvider {
	var returns cdktf.TerraformProvider
	_jsii_.Get(
		j,
		"provider",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJob) Provisioners() *[]interface{} {
	var returns *[]interface{}
	_jsii_.Get(
		j,
		"provisioners",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJob) PubsubTarget() CloudSchedulerJobPubsubTargetOutputReference {
	var returns CloudSchedulerJobPubsubTargetOutputReference
	_jsii_.Get(
		j,
		"pubsubTarget",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJob) PubsubTargetInput() *CloudSchedulerJobPubsubTarget {
	var returns *CloudSchedulerJobPubsubTarget
	_jsii_.Get(
		j,
		"pubsubTargetInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJob) RawOverrides() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"rawOverrides",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJob) Region() *string {
	var returns *string
	_jsii_.Get(
		j,
		"region",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJob) RegionInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"regionInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJob) RetryConfig() CloudSchedulerJobRetryConfigOutputReference {
	var returns CloudSchedulerJobRetryConfigOutputReference
	_jsii_.Get(
		j,
		"retryConfig",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJob) RetryConfigInput() *CloudSchedulerJobRetryConfig {
	var returns *CloudSchedulerJobRetryConfig
	_jsii_.Get(
		j,
		"retryConfigInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJob) Schedule() *string {
	var returns *string
	_jsii_.Get(
		j,
		"schedule",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJob) ScheduleInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"scheduleInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJob) State() *string {
	var returns *string
	_jsii_.Get(
		j,
		"state",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJob) TerraformGeneratorMetadata() *cdktf.TerraformProviderGeneratorMetadata {
	var returns *cdktf.TerraformProviderGeneratorMetadata
	_jsii_.Get(
		j,
		"terraformGeneratorMetadata",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJob) TerraformMetaArguments() *map[string]interface{} {
	var returns *map[string]interface{}
	_jsii_.Get(
		j,
		"terraformMetaArguments",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJob) TerraformResourceType() *string {
	var returns *string
	_jsii_.Get(
		j,
		"terraformResourceType",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJob) Timeouts() CloudSchedulerJobTimeoutsOutputReference {
	var returns CloudSchedulerJobTimeoutsOutputReference
	_jsii_.Get(
		j,
		"timeouts",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJob) TimeoutsInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"timeoutsInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJob) TimeZone() *string {
	var returns *string
	_jsii_.Get(
		j,
		"timeZone",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJob) TimeZoneInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"timeZoneInput",
		&returns,
	)
	return returns
}


// Create a new {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_scheduler_job google_cloud_scheduler_job} Resource.
func NewCloudSchedulerJob(scope constructs.Construct, id *string, config *CloudSchedulerJobConfig) CloudSchedulerJob {
	_init_.Initialize()

	if err := validateNewCloudSchedulerJobParameters(scope, id, config); err != nil {
		panic(err)
	}
	j := jsiiProxy_CloudSchedulerJob{}

	_jsii_.Create(
		"@cdktf/provider-google.cloudSchedulerJob.CloudSchedulerJob",
		[]interface{}{scope, id, config},
		&j,
	)

	return &j
}

// Create a new {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_scheduler_job google_cloud_scheduler_job} Resource.
func NewCloudSchedulerJob_Override(c CloudSchedulerJob, scope constructs.Construct, id *string, config *CloudSchedulerJobConfig) {
	_init_.Initialize()

	_jsii_.Create(
		"@cdktf/provider-google.cloudSchedulerJob.CloudSchedulerJob",
		[]interface{}{scope, id, config},
		c,
	)
}

func (j *jsiiProxy_CloudSchedulerJob)SetAttemptDeadline(val *string) {
	if err := j.validateSetAttemptDeadlineParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"attemptDeadline",
		val,
	)
}

func (j *jsiiProxy_CloudSchedulerJob)SetConnection(val interface{}) {
	if err := j.validateSetConnectionParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"connection",
		val,
	)
}

func (j *jsiiProxy_CloudSchedulerJob)SetCount(val interface{}) {
	if err := j.validateSetCountParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"count",
		val,
	)
}

func (j *jsiiProxy_CloudSchedulerJob)SetDependsOn(val *[]*string) {
	_jsii_.Set(
		j,
		"dependsOn",
		val,
	)
}

func (j *jsiiProxy_CloudSchedulerJob)SetDescription(val *string) {
	if err := j.validateSetDescriptionParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"description",
		val,
	)
}

func (j *jsiiProxy_CloudSchedulerJob)SetForEach(val cdktf.ITerraformIterator) {
	_jsii_.Set(
		j,
		"forEach",
		val,
	)
}

func (j *jsiiProxy_CloudSchedulerJob)SetId(val *string) {
	if err := j.validateSetIdParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"id",
		val,
	)
}

func (j *jsiiProxy_CloudSchedulerJob)SetLifecycle(val *cdktf.TerraformResourceLifecycle) {
	if err := j.validateSetLifecycleParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"lifecycle",
		val,
	)
}

func (j *jsiiProxy_CloudSchedulerJob)SetName(val *string) {
	if err := j.validateSetNameParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"name",
		val,
	)
}

func (j *jsiiProxy_CloudSchedulerJob)SetPaused(val interface{}) {
	if err := j.validateSetPausedParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"paused",
		val,
	)
}

func (j *jsiiProxy_CloudSchedulerJob)SetProject(val *string) {
	if err := j.validateSetProjectParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"project",
		val,
	)
}

func (j *jsiiProxy_CloudSchedulerJob)SetProvider(val cdktf.TerraformProvider) {
	_jsii_.Set(
		j,
		"provider",
		val,
	)
}

func (j *jsiiProxy_CloudSchedulerJob)SetProvisioners(val *[]interface{}) {
	if err := j.validateSetProvisionersParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"provisioners",
		val,
	)
}

func (j *jsiiProxy_CloudSchedulerJob)SetRegion(val *string) {
	if err := j.validateSetRegionParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"region",
		val,
	)
}

func (j *jsiiProxy_CloudSchedulerJob)SetSchedule(val *string) {
	if err := j.validateSetScheduleParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"schedule",
		val,
	)
}

func (j *jsiiProxy_CloudSchedulerJob)SetTimeZone(val *string) {
	if err := j.validateSetTimeZoneParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"timeZone",
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
func CloudSchedulerJob_IsConstruct(x interface{}) *bool {
	_init_.Initialize()

	if err := validateCloudSchedulerJob_IsConstructParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"@cdktf/provider-google.cloudSchedulerJob.CloudSchedulerJob",
		"isConstruct",
		[]interface{}{x},
		&returns,
	)

	return returns
}

// Experimental.
func CloudSchedulerJob_IsTerraformElement(x interface{}) *bool {
	_init_.Initialize()

	if err := validateCloudSchedulerJob_IsTerraformElementParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"@cdktf/provider-google.cloudSchedulerJob.CloudSchedulerJob",
		"isTerraformElement",
		[]interface{}{x},
		&returns,
	)

	return returns
}

// Experimental.
func CloudSchedulerJob_IsTerraformResource(x interface{}) *bool {
	_init_.Initialize()

	if err := validateCloudSchedulerJob_IsTerraformResourceParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"@cdktf/provider-google.cloudSchedulerJob.CloudSchedulerJob",
		"isTerraformResource",
		[]interface{}{x},
		&returns,
	)

	return returns
}

func CloudSchedulerJob_TfResourceType() *string {
	_init_.Initialize()
	var returns *string
	_jsii_.StaticGet(
		"@cdktf/provider-google.cloudSchedulerJob.CloudSchedulerJob",
		"tfResourceType",
		&returns,
	)
	return returns
}

func (c *jsiiProxy_CloudSchedulerJob) AddOverride(path *string, value interface{}) {
	if err := c.validateAddOverrideParameters(path, value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"addOverride",
		[]interface{}{path, value},
	)
}

func (c *jsiiProxy_CloudSchedulerJob) GetAnyMapAttribute(terraformAttribute *string) *map[string]interface{} {
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

func (c *jsiiProxy_CloudSchedulerJob) GetBooleanAttribute(terraformAttribute *string) cdktf.IResolvable {
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

func (c *jsiiProxy_CloudSchedulerJob) GetBooleanMapAttribute(terraformAttribute *string) *map[string]*bool {
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

func (c *jsiiProxy_CloudSchedulerJob) GetListAttribute(terraformAttribute *string) *[]*string {
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

func (c *jsiiProxy_CloudSchedulerJob) GetNumberAttribute(terraformAttribute *string) *float64 {
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

func (c *jsiiProxy_CloudSchedulerJob) GetNumberListAttribute(terraformAttribute *string) *[]*float64 {
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

func (c *jsiiProxy_CloudSchedulerJob) GetNumberMapAttribute(terraformAttribute *string) *map[string]*float64 {
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

func (c *jsiiProxy_CloudSchedulerJob) GetStringAttribute(terraformAttribute *string) *string {
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

func (c *jsiiProxy_CloudSchedulerJob) GetStringMapAttribute(terraformAttribute *string) *map[string]*string {
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

func (c *jsiiProxy_CloudSchedulerJob) InterpolationForAttribute(terraformAttribute *string) cdktf.IResolvable {
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

func (c *jsiiProxy_CloudSchedulerJob) OverrideLogicalId(newLogicalId *string) {
	if err := c.validateOverrideLogicalIdParameters(newLogicalId); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"overrideLogicalId",
		[]interface{}{newLogicalId},
	)
}

func (c *jsiiProxy_CloudSchedulerJob) PutAppEngineHttpTarget(value *CloudSchedulerJobAppEngineHttpTarget) {
	if err := c.validatePutAppEngineHttpTargetParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putAppEngineHttpTarget",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_CloudSchedulerJob) PutHttpTarget(value *CloudSchedulerJobHttpTarget) {
	if err := c.validatePutHttpTargetParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putHttpTarget",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_CloudSchedulerJob) PutPubsubTarget(value *CloudSchedulerJobPubsubTarget) {
	if err := c.validatePutPubsubTargetParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putPubsubTarget",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_CloudSchedulerJob) PutRetryConfig(value *CloudSchedulerJobRetryConfig) {
	if err := c.validatePutRetryConfigParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putRetryConfig",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_CloudSchedulerJob) PutTimeouts(value *CloudSchedulerJobTimeouts) {
	if err := c.validatePutTimeoutsParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putTimeouts",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_CloudSchedulerJob) ResetAppEngineHttpTarget() {
	_jsii_.InvokeVoid(
		c,
		"resetAppEngineHttpTarget",
		nil, // no parameters
	)
}

func (c *jsiiProxy_CloudSchedulerJob) ResetAttemptDeadline() {
	_jsii_.InvokeVoid(
		c,
		"resetAttemptDeadline",
		nil, // no parameters
	)
}

func (c *jsiiProxy_CloudSchedulerJob) ResetDescription() {
	_jsii_.InvokeVoid(
		c,
		"resetDescription",
		nil, // no parameters
	)
}

func (c *jsiiProxy_CloudSchedulerJob) ResetHttpTarget() {
	_jsii_.InvokeVoid(
		c,
		"resetHttpTarget",
		nil, // no parameters
	)
}

func (c *jsiiProxy_CloudSchedulerJob) ResetId() {
	_jsii_.InvokeVoid(
		c,
		"resetId",
		nil, // no parameters
	)
}

func (c *jsiiProxy_CloudSchedulerJob) ResetOverrideLogicalId() {
	_jsii_.InvokeVoid(
		c,
		"resetOverrideLogicalId",
		nil, // no parameters
	)
}

func (c *jsiiProxy_CloudSchedulerJob) ResetPaused() {
	_jsii_.InvokeVoid(
		c,
		"resetPaused",
		nil, // no parameters
	)
}

func (c *jsiiProxy_CloudSchedulerJob) ResetProject() {
	_jsii_.InvokeVoid(
		c,
		"resetProject",
		nil, // no parameters
	)
}

func (c *jsiiProxy_CloudSchedulerJob) ResetPubsubTarget() {
	_jsii_.InvokeVoid(
		c,
		"resetPubsubTarget",
		nil, // no parameters
	)
}

func (c *jsiiProxy_CloudSchedulerJob) ResetRegion() {
	_jsii_.InvokeVoid(
		c,
		"resetRegion",
		nil, // no parameters
	)
}

func (c *jsiiProxy_CloudSchedulerJob) ResetRetryConfig() {
	_jsii_.InvokeVoid(
		c,
		"resetRetryConfig",
		nil, // no parameters
	)
}

func (c *jsiiProxy_CloudSchedulerJob) ResetSchedule() {
	_jsii_.InvokeVoid(
		c,
		"resetSchedule",
		nil, // no parameters
	)
}

func (c *jsiiProxy_CloudSchedulerJob) ResetTimeouts() {
	_jsii_.InvokeVoid(
		c,
		"resetTimeouts",
		nil, // no parameters
	)
}

func (c *jsiiProxy_CloudSchedulerJob) ResetTimeZone() {
	_jsii_.InvokeVoid(
		c,
		"resetTimeZone",
		nil, // no parameters
	)
}

func (c *jsiiProxy_CloudSchedulerJob) SynthesizeAttributes() *map[string]interface{} {
	var returns *map[string]interface{}

	_jsii_.Invoke(
		c,
		"synthesizeAttributes",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (c *jsiiProxy_CloudSchedulerJob) ToMetadata() interface{} {
	var returns interface{}

	_jsii_.Invoke(
		c,
		"toMetadata",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (c *jsiiProxy_CloudSchedulerJob) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		c,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (c *jsiiProxy_CloudSchedulerJob) ToTerraform() interface{} {
	var returns interface{}

	_jsii_.Invoke(
		c,
		"toTerraform",
		nil, // no parameters
		&returns,
	)

	return returns
}

