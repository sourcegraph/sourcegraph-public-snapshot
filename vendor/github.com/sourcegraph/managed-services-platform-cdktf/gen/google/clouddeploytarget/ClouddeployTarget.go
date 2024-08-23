package clouddeploytarget

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/sourcegraph/managed-services-platform-cdktf/gen/google/jsii"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/clouddeploytarget/internal"
)

// Represents a {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/clouddeploy_target google_clouddeploy_target}.
type ClouddeployTarget interface {
	cdktf.TerraformResource
	Annotations() *map[string]*string
	SetAnnotations(val *map[string]*string)
	AnnotationsInput() *map[string]*string
	AnthosCluster() ClouddeployTargetAnthosClusterOutputReference
	AnthosClusterInput() *ClouddeployTargetAnthosCluster
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
	CreateTime() *string
	CustomTarget() ClouddeployTargetCustomTargetOutputReference
	CustomTargetInput() *ClouddeployTargetCustomTarget
	// Experimental.
	DependsOn() *[]*string
	// Experimental.
	SetDependsOn(val *[]*string)
	DeployParameters() *map[string]*string
	SetDeployParameters(val *map[string]*string)
	DeployParametersInput() *map[string]*string
	Description() *string
	SetDescription(val *string)
	DescriptionInput() *string
	EffectiveAnnotations() cdktf.StringMap
	EffectiveLabels() cdktf.StringMap
	Etag() *string
	ExecutionConfigs() ClouddeployTargetExecutionConfigsList
	ExecutionConfigsInput() interface{}
	// Experimental.
	ForEach() cdktf.ITerraformIterator
	// Experimental.
	SetForEach(val cdktf.ITerraformIterator)
	// Experimental.
	Fqn() *string
	// Experimental.
	FriendlyUniqueId() *string
	Gke() ClouddeployTargetGkeOutputReference
	GkeInput() *ClouddeployTargetGke
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
	Location() *string
	SetLocation(val *string)
	LocationInput() *string
	MultiTarget() ClouddeployTargetMultiTargetOutputReference
	MultiTargetInput() *ClouddeployTargetMultiTarget
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
	// Experimental.
	RawOverrides() interface{}
	RequireApproval() interface{}
	SetRequireApproval(val interface{})
	RequireApprovalInput() interface{}
	Run() ClouddeployTargetRunOutputReference
	RunInput() *ClouddeployTargetRun
	TargetId() *string
	// Experimental.
	TerraformGeneratorMetadata() *cdktf.TerraformProviderGeneratorMetadata
	TerraformLabels() cdktf.StringMap
	// Experimental.
	TerraformMetaArguments() *map[string]interface{}
	// Experimental.
	TerraformResourceType() *string
	Timeouts() ClouddeployTargetTimeoutsOutputReference
	TimeoutsInput() interface{}
	Uid() *string
	UpdateTime() *string
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
	PutAnthosCluster(value *ClouddeployTargetAnthosCluster)
	PutCustomTarget(value *ClouddeployTargetCustomTarget)
	PutExecutionConfigs(value interface{})
	PutGke(value *ClouddeployTargetGke)
	PutMultiTarget(value *ClouddeployTargetMultiTarget)
	PutRun(value *ClouddeployTargetRun)
	PutTimeouts(value *ClouddeployTargetTimeouts)
	ResetAnnotations()
	ResetAnthosCluster()
	ResetCustomTarget()
	ResetDeployParameters()
	ResetDescription()
	ResetExecutionConfigs()
	ResetGke()
	ResetId()
	ResetLabels()
	ResetMultiTarget()
	// Resets a previously passed logical Id to use the auto-generated logical id again.
	// Experimental.
	ResetOverrideLogicalId()
	ResetProject()
	ResetRequireApproval()
	ResetRun()
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

// The jsii proxy struct for ClouddeployTarget
type jsiiProxy_ClouddeployTarget struct {
	internal.Type__cdktfTerraformResource
}

func (j *jsiiProxy_ClouddeployTarget) Annotations() *map[string]*string {
	var returns *map[string]*string
	_jsii_.Get(
		j,
		"annotations",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployTarget) AnnotationsInput() *map[string]*string {
	var returns *map[string]*string
	_jsii_.Get(
		j,
		"annotationsInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployTarget) AnthosCluster() ClouddeployTargetAnthosClusterOutputReference {
	var returns ClouddeployTargetAnthosClusterOutputReference
	_jsii_.Get(
		j,
		"anthosCluster",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployTarget) AnthosClusterInput() *ClouddeployTargetAnthosCluster {
	var returns *ClouddeployTargetAnthosCluster
	_jsii_.Get(
		j,
		"anthosClusterInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployTarget) CdktfStack() cdktf.TerraformStack {
	var returns cdktf.TerraformStack
	_jsii_.Get(
		j,
		"cdktfStack",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployTarget) Connection() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"connection",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployTarget) ConstructNodeMetadata() *map[string]interface{} {
	var returns *map[string]interface{}
	_jsii_.Get(
		j,
		"constructNodeMetadata",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployTarget) Count() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"count",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployTarget) CreateTime() *string {
	var returns *string
	_jsii_.Get(
		j,
		"createTime",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployTarget) CustomTarget() ClouddeployTargetCustomTargetOutputReference {
	var returns ClouddeployTargetCustomTargetOutputReference
	_jsii_.Get(
		j,
		"customTarget",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployTarget) CustomTargetInput() *ClouddeployTargetCustomTarget {
	var returns *ClouddeployTargetCustomTarget
	_jsii_.Get(
		j,
		"customTargetInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployTarget) DependsOn() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"dependsOn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployTarget) DeployParameters() *map[string]*string {
	var returns *map[string]*string
	_jsii_.Get(
		j,
		"deployParameters",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployTarget) DeployParametersInput() *map[string]*string {
	var returns *map[string]*string
	_jsii_.Get(
		j,
		"deployParametersInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployTarget) Description() *string {
	var returns *string
	_jsii_.Get(
		j,
		"description",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployTarget) DescriptionInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"descriptionInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployTarget) EffectiveAnnotations() cdktf.StringMap {
	var returns cdktf.StringMap
	_jsii_.Get(
		j,
		"effectiveAnnotations",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployTarget) EffectiveLabels() cdktf.StringMap {
	var returns cdktf.StringMap
	_jsii_.Get(
		j,
		"effectiveLabels",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployTarget) Etag() *string {
	var returns *string
	_jsii_.Get(
		j,
		"etag",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployTarget) ExecutionConfigs() ClouddeployTargetExecutionConfigsList {
	var returns ClouddeployTargetExecutionConfigsList
	_jsii_.Get(
		j,
		"executionConfigs",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployTarget) ExecutionConfigsInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"executionConfigsInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployTarget) ForEach() cdktf.ITerraformIterator {
	var returns cdktf.ITerraformIterator
	_jsii_.Get(
		j,
		"forEach",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployTarget) Fqn() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fqn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployTarget) FriendlyUniqueId() *string {
	var returns *string
	_jsii_.Get(
		j,
		"friendlyUniqueId",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployTarget) Gke() ClouddeployTargetGkeOutputReference {
	var returns ClouddeployTargetGkeOutputReference
	_jsii_.Get(
		j,
		"gke",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployTarget) GkeInput() *ClouddeployTargetGke {
	var returns *ClouddeployTargetGke
	_jsii_.Get(
		j,
		"gkeInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployTarget) Id() *string {
	var returns *string
	_jsii_.Get(
		j,
		"id",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployTarget) IdInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"idInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployTarget) Labels() *map[string]*string {
	var returns *map[string]*string
	_jsii_.Get(
		j,
		"labels",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployTarget) LabelsInput() *map[string]*string {
	var returns *map[string]*string
	_jsii_.Get(
		j,
		"labelsInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployTarget) Lifecycle() *cdktf.TerraformResourceLifecycle {
	var returns *cdktf.TerraformResourceLifecycle
	_jsii_.Get(
		j,
		"lifecycle",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployTarget) Location() *string {
	var returns *string
	_jsii_.Get(
		j,
		"location",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployTarget) LocationInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"locationInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployTarget) MultiTarget() ClouddeployTargetMultiTargetOutputReference {
	var returns ClouddeployTargetMultiTargetOutputReference
	_jsii_.Get(
		j,
		"multiTarget",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployTarget) MultiTargetInput() *ClouddeployTargetMultiTarget {
	var returns *ClouddeployTargetMultiTarget
	_jsii_.Get(
		j,
		"multiTargetInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployTarget) Name() *string {
	var returns *string
	_jsii_.Get(
		j,
		"name",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployTarget) NameInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"nameInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployTarget) Node() constructs.Node {
	var returns constructs.Node
	_jsii_.Get(
		j,
		"node",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployTarget) Project() *string {
	var returns *string
	_jsii_.Get(
		j,
		"project",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployTarget) ProjectInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"projectInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployTarget) Provider() cdktf.TerraformProvider {
	var returns cdktf.TerraformProvider
	_jsii_.Get(
		j,
		"provider",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployTarget) Provisioners() *[]interface{} {
	var returns *[]interface{}
	_jsii_.Get(
		j,
		"provisioners",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployTarget) RawOverrides() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"rawOverrides",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployTarget) RequireApproval() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"requireApproval",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployTarget) RequireApprovalInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"requireApprovalInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployTarget) Run() ClouddeployTargetRunOutputReference {
	var returns ClouddeployTargetRunOutputReference
	_jsii_.Get(
		j,
		"run",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployTarget) RunInput() *ClouddeployTargetRun {
	var returns *ClouddeployTargetRun
	_jsii_.Get(
		j,
		"runInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployTarget) TargetId() *string {
	var returns *string
	_jsii_.Get(
		j,
		"targetId",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployTarget) TerraformGeneratorMetadata() *cdktf.TerraformProviderGeneratorMetadata {
	var returns *cdktf.TerraformProviderGeneratorMetadata
	_jsii_.Get(
		j,
		"terraformGeneratorMetadata",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployTarget) TerraformLabels() cdktf.StringMap {
	var returns cdktf.StringMap
	_jsii_.Get(
		j,
		"terraformLabels",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployTarget) TerraformMetaArguments() *map[string]interface{} {
	var returns *map[string]interface{}
	_jsii_.Get(
		j,
		"terraformMetaArguments",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployTarget) TerraformResourceType() *string {
	var returns *string
	_jsii_.Get(
		j,
		"terraformResourceType",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployTarget) Timeouts() ClouddeployTargetTimeoutsOutputReference {
	var returns ClouddeployTargetTimeoutsOutputReference
	_jsii_.Get(
		j,
		"timeouts",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployTarget) TimeoutsInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"timeoutsInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployTarget) Uid() *string {
	var returns *string
	_jsii_.Get(
		j,
		"uid",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployTarget) UpdateTime() *string {
	var returns *string
	_jsii_.Get(
		j,
		"updateTime",
		&returns,
	)
	return returns
}


// Create a new {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/clouddeploy_target google_clouddeploy_target} Resource.
func NewClouddeployTarget(scope constructs.Construct, id *string, config *ClouddeployTargetConfig) ClouddeployTarget {
	_init_.Initialize()

	if err := validateNewClouddeployTargetParameters(scope, id, config); err != nil {
		panic(err)
	}
	j := jsiiProxy_ClouddeployTarget{}

	_jsii_.Create(
		"@cdktf/provider-google.clouddeployTarget.ClouddeployTarget",
		[]interface{}{scope, id, config},
		&j,
	)

	return &j
}

// Create a new {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/clouddeploy_target google_clouddeploy_target} Resource.
func NewClouddeployTarget_Override(c ClouddeployTarget, scope constructs.Construct, id *string, config *ClouddeployTargetConfig) {
	_init_.Initialize()

	_jsii_.Create(
		"@cdktf/provider-google.clouddeployTarget.ClouddeployTarget",
		[]interface{}{scope, id, config},
		c,
	)
}

func (j *jsiiProxy_ClouddeployTarget)SetAnnotations(val *map[string]*string) {
	if err := j.validateSetAnnotationsParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"annotations",
		val,
	)
}

func (j *jsiiProxy_ClouddeployTarget)SetConnection(val interface{}) {
	if err := j.validateSetConnectionParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"connection",
		val,
	)
}

func (j *jsiiProxy_ClouddeployTarget)SetCount(val interface{}) {
	if err := j.validateSetCountParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"count",
		val,
	)
}

func (j *jsiiProxy_ClouddeployTarget)SetDependsOn(val *[]*string) {
	_jsii_.Set(
		j,
		"dependsOn",
		val,
	)
}

func (j *jsiiProxy_ClouddeployTarget)SetDeployParameters(val *map[string]*string) {
	if err := j.validateSetDeployParametersParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"deployParameters",
		val,
	)
}

func (j *jsiiProxy_ClouddeployTarget)SetDescription(val *string) {
	if err := j.validateSetDescriptionParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"description",
		val,
	)
}

func (j *jsiiProxy_ClouddeployTarget)SetForEach(val cdktf.ITerraformIterator) {
	_jsii_.Set(
		j,
		"forEach",
		val,
	)
}

func (j *jsiiProxy_ClouddeployTarget)SetId(val *string) {
	if err := j.validateSetIdParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"id",
		val,
	)
}

func (j *jsiiProxy_ClouddeployTarget)SetLabels(val *map[string]*string) {
	if err := j.validateSetLabelsParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"labels",
		val,
	)
}

func (j *jsiiProxy_ClouddeployTarget)SetLifecycle(val *cdktf.TerraformResourceLifecycle) {
	if err := j.validateSetLifecycleParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"lifecycle",
		val,
	)
}

func (j *jsiiProxy_ClouddeployTarget)SetLocation(val *string) {
	if err := j.validateSetLocationParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"location",
		val,
	)
}

func (j *jsiiProxy_ClouddeployTarget)SetName(val *string) {
	if err := j.validateSetNameParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"name",
		val,
	)
}

func (j *jsiiProxy_ClouddeployTarget)SetProject(val *string) {
	if err := j.validateSetProjectParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"project",
		val,
	)
}

func (j *jsiiProxy_ClouddeployTarget)SetProvider(val cdktf.TerraformProvider) {
	_jsii_.Set(
		j,
		"provider",
		val,
	)
}

func (j *jsiiProxy_ClouddeployTarget)SetProvisioners(val *[]interface{}) {
	if err := j.validateSetProvisionersParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"provisioners",
		val,
	)
}

func (j *jsiiProxy_ClouddeployTarget)SetRequireApproval(val interface{}) {
	if err := j.validateSetRequireApprovalParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"requireApproval",
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
func ClouddeployTarget_IsConstruct(x interface{}) *bool {
	_init_.Initialize()

	if err := validateClouddeployTarget_IsConstructParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"@cdktf/provider-google.clouddeployTarget.ClouddeployTarget",
		"isConstruct",
		[]interface{}{x},
		&returns,
	)

	return returns
}

// Experimental.
func ClouddeployTarget_IsTerraformElement(x interface{}) *bool {
	_init_.Initialize()

	if err := validateClouddeployTarget_IsTerraformElementParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"@cdktf/provider-google.clouddeployTarget.ClouddeployTarget",
		"isTerraformElement",
		[]interface{}{x},
		&returns,
	)

	return returns
}

// Experimental.
func ClouddeployTarget_IsTerraformResource(x interface{}) *bool {
	_init_.Initialize()

	if err := validateClouddeployTarget_IsTerraformResourceParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"@cdktf/provider-google.clouddeployTarget.ClouddeployTarget",
		"isTerraformResource",
		[]interface{}{x},
		&returns,
	)

	return returns
}

func ClouddeployTarget_TfResourceType() *string {
	_init_.Initialize()
	var returns *string
	_jsii_.StaticGet(
		"@cdktf/provider-google.clouddeployTarget.ClouddeployTarget",
		"tfResourceType",
		&returns,
	)
	return returns
}

func (c *jsiiProxy_ClouddeployTarget) AddOverride(path *string, value interface{}) {
	if err := c.validateAddOverrideParameters(path, value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"addOverride",
		[]interface{}{path, value},
	)
}

func (c *jsiiProxy_ClouddeployTarget) GetAnyMapAttribute(terraformAttribute *string) *map[string]interface{} {
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

func (c *jsiiProxy_ClouddeployTarget) GetBooleanAttribute(terraformAttribute *string) cdktf.IResolvable {
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

func (c *jsiiProxy_ClouddeployTarget) GetBooleanMapAttribute(terraformAttribute *string) *map[string]*bool {
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

func (c *jsiiProxy_ClouddeployTarget) GetListAttribute(terraformAttribute *string) *[]*string {
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

func (c *jsiiProxy_ClouddeployTarget) GetNumberAttribute(terraformAttribute *string) *float64 {
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

func (c *jsiiProxy_ClouddeployTarget) GetNumberListAttribute(terraformAttribute *string) *[]*float64 {
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

func (c *jsiiProxy_ClouddeployTarget) GetNumberMapAttribute(terraformAttribute *string) *map[string]*float64 {
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

func (c *jsiiProxy_ClouddeployTarget) GetStringAttribute(terraformAttribute *string) *string {
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

func (c *jsiiProxy_ClouddeployTarget) GetStringMapAttribute(terraformAttribute *string) *map[string]*string {
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

func (c *jsiiProxy_ClouddeployTarget) InterpolationForAttribute(terraformAttribute *string) cdktf.IResolvable {
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

func (c *jsiiProxy_ClouddeployTarget) OverrideLogicalId(newLogicalId *string) {
	if err := c.validateOverrideLogicalIdParameters(newLogicalId); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"overrideLogicalId",
		[]interface{}{newLogicalId},
	)
}

func (c *jsiiProxy_ClouddeployTarget) PutAnthosCluster(value *ClouddeployTargetAnthosCluster) {
	if err := c.validatePutAnthosClusterParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putAnthosCluster",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ClouddeployTarget) PutCustomTarget(value *ClouddeployTargetCustomTarget) {
	if err := c.validatePutCustomTargetParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putCustomTarget",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ClouddeployTarget) PutExecutionConfigs(value interface{}) {
	if err := c.validatePutExecutionConfigsParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putExecutionConfigs",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ClouddeployTarget) PutGke(value *ClouddeployTargetGke) {
	if err := c.validatePutGkeParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putGke",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ClouddeployTarget) PutMultiTarget(value *ClouddeployTargetMultiTarget) {
	if err := c.validatePutMultiTargetParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putMultiTarget",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ClouddeployTarget) PutRun(value *ClouddeployTargetRun) {
	if err := c.validatePutRunParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putRun",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ClouddeployTarget) PutTimeouts(value *ClouddeployTargetTimeouts) {
	if err := c.validatePutTimeoutsParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putTimeouts",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ClouddeployTarget) ResetAnnotations() {
	_jsii_.InvokeVoid(
		c,
		"resetAnnotations",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ClouddeployTarget) ResetAnthosCluster() {
	_jsii_.InvokeVoid(
		c,
		"resetAnthosCluster",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ClouddeployTarget) ResetCustomTarget() {
	_jsii_.InvokeVoid(
		c,
		"resetCustomTarget",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ClouddeployTarget) ResetDeployParameters() {
	_jsii_.InvokeVoid(
		c,
		"resetDeployParameters",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ClouddeployTarget) ResetDescription() {
	_jsii_.InvokeVoid(
		c,
		"resetDescription",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ClouddeployTarget) ResetExecutionConfigs() {
	_jsii_.InvokeVoid(
		c,
		"resetExecutionConfigs",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ClouddeployTarget) ResetGke() {
	_jsii_.InvokeVoid(
		c,
		"resetGke",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ClouddeployTarget) ResetId() {
	_jsii_.InvokeVoid(
		c,
		"resetId",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ClouddeployTarget) ResetLabels() {
	_jsii_.InvokeVoid(
		c,
		"resetLabels",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ClouddeployTarget) ResetMultiTarget() {
	_jsii_.InvokeVoid(
		c,
		"resetMultiTarget",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ClouddeployTarget) ResetOverrideLogicalId() {
	_jsii_.InvokeVoid(
		c,
		"resetOverrideLogicalId",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ClouddeployTarget) ResetProject() {
	_jsii_.InvokeVoid(
		c,
		"resetProject",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ClouddeployTarget) ResetRequireApproval() {
	_jsii_.InvokeVoid(
		c,
		"resetRequireApproval",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ClouddeployTarget) ResetRun() {
	_jsii_.InvokeVoid(
		c,
		"resetRun",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ClouddeployTarget) ResetTimeouts() {
	_jsii_.InvokeVoid(
		c,
		"resetTimeouts",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ClouddeployTarget) SynthesizeAttributes() *map[string]interface{} {
	var returns *map[string]interface{}

	_jsii_.Invoke(
		c,
		"synthesizeAttributes",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ClouddeployTarget) ToMetadata() interface{} {
	var returns interface{}

	_jsii_.Invoke(
		c,
		"toMetadata",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ClouddeployTarget) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		c,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ClouddeployTarget) ToTerraform() interface{} {
	var returns interface{}

	_jsii_.Invoke(
		c,
		"toTerraform",
		nil, // no parameters
		&returns,
	)

	return returns
}

