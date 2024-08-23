package cloudrunv2service

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/sourcegraph/managed-services-platform-cdktf/gen/google/jsii"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/cloudrunv2service/internal"
)

// Represents a {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_run_v2_service google_cloud_run_v2_service}.
type CloudRunV2Service interface {
	cdktf.TerraformResource
	Annotations() *map[string]*string
	SetAnnotations(val *map[string]*string)
	AnnotationsInput() *map[string]*string
	BinaryAuthorization() CloudRunV2ServiceBinaryAuthorizationOutputReference
	BinaryAuthorizationInput() *CloudRunV2ServiceBinaryAuthorization
	// Experimental.
	CdktfStack() cdktf.TerraformStack
	Client() *string
	SetClient(val *string)
	ClientInput() *string
	ClientVersion() *string
	SetClientVersion(val *string)
	ClientVersionInput() *string
	Conditions() CloudRunV2ServiceConditionsList
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
	Creator() *string
	CustomAudiences() *[]*string
	SetCustomAudiences(val *[]*string)
	CustomAudiencesInput() *[]*string
	DeleteTime() *string
	// Experimental.
	DependsOn() *[]*string
	// Experimental.
	SetDependsOn(val *[]*string)
	Description() *string
	SetDescription(val *string)
	DescriptionInput() *string
	EffectiveAnnotations() cdktf.StringMap
	EffectiveLabels() cdktf.StringMap
	Etag() *string
	ExpireTime() *string
	// Experimental.
	ForEach() cdktf.ITerraformIterator
	// Experimental.
	SetForEach(val cdktf.ITerraformIterator)
	// Experimental.
	Fqn() *string
	// Experimental.
	FriendlyUniqueId() *string
	Generation() *string
	Id() *string
	SetId(val *string)
	IdInput() *string
	Ingress() *string
	SetIngress(val *string)
	IngressInput() *string
	Labels() *map[string]*string
	SetLabels(val *map[string]*string)
	LabelsInput() *map[string]*string
	LastModifier() *string
	LatestCreatedRevision() *string
	LatestReadyRevision() *string
	LaunchStage() *string
	SetLaunchStage(val *string)
	LaunchStageInput() *string
	// Experimental.
	Lifecycle() *cdktf.TerraformResourceLifecycle
	// Experimental.
	SetLifecycle(val *cdktf.TerraformResourceLifecycle)
	Location() *string
	SetLocation(val *string)
	LocationInput() *string
	Name() *string
	SetName(val *string)
	NameInput() *string
	// The tree node.
	Node() constructs.Node
	ObservedGeneration() *string
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
	Reconciling() cdktf.IResolvable
	Template() CloudRunV2ServiceTemplateOutputReference
	TemplateInput() *CloudRunV2ServiceTemplate
	TerminalCondition() CloudRunV2ServiceTerminalConditionList
	// Experimental.
	TerraformGeneratorMetadata() *cdktf.TerraformProviderGeneratorMetadata
	TerraformLabels() cdktf.StringMap
	// Experimental.
	TerraformMetaArguments() *map[string]interface{}
	// Experimental.
	TerraformResourceType() *string
	Timeouts() CloudRunV2ServiceTimeoutsOutputReference
	TimeoutsInput() interface{}
	Traffic() CloudRunV2ServiceTrafficList
	TrafficInput() interface{}
	TrafficStatuses() CloudRunV2ServiceTrafficStatusesList
	Uid() *string
	UpdateTime() *string
	Uri() *string
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
	PutBinaryAuthorization(value *CloudRunV2ServiceBinaryAuthorization)
	PutTemplate(value *CloudRunV2ServiceTemplate)
	PutTimeouts(value *CloudRunV2ServiceTimeouts)
	PutTraffic(value interface{})
	ResetAnnotations()
	ResetBinaryAuthorization()
	ResetClient()
	ResetClientVersion()
	ResetCustomAudiences()
	ResetDescription()
	ResetId()
	ResetIngress()
	ResetLabels()
	ResetLaunchStage()
	// Resets a previously passed logical Id to use the auto-generated logical id again.
	// Experimental.
	ResetOverrideLogicalId()
	ResetProject()
	ResetTimeouts()
	ResetTraffic()
	SynthesizeAttributes() *map[string]interface{}
	// Experimental.
	ToMetadata() interface{}
	// Returns a string representation of this construct.
	ToString() *string
	// Adds this resource to the terraform JSON output.
	// Experimental.
	ToTerraform() interface{}
}

// The jsii proxy struct for CloudRunV2Service
type jsiiProxy_CloudRunV2Service struct {
	internal.Type__cdktfTerraformResource
}

func (j *jsiiProxy_CloudRunV2Service) Annotations() *map[string]*string {
	var returns *map[string]*string
	_jsii_.Get(
		j,
		"annotations",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2Service) AnnotationsInput() *map[string]*string {
	var returns *map[string]*string
	_jsii_.Get(
		j,
		"annotationsInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2Service) BinaryAuthorization() CloudRunV2ServiceBinaryAuthorizationOutputReference {
	var returns CloudRunV2ServiceBinaryAuthorizationOutputReference
	_jsii_.Get(
		j,
		"binaryAuthorization",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2Service) BinaryAuthorizationInput() *CloudRunV2ServiceBinaryAuthorization {
	var returns *CloudRunV2ServiceBinaryAuthorization
	_jsii_.Get(
		j,
		"binaryAuthorizationInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2Service) CdktfStack() cdktf.TerraformStack {
	var returns cdktf.TerraformStack
	_jsii_.Get(
		j,
		"cdktfStack",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2Service) Client() *string {
	var returns *string
	_jsii_.Get(
		j,
		"client",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2Service) ClientInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"clientInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2Service) ClientVersion() *string {
	var returns *string
	_jsii_.Get(
		j,
		"clientVersion",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2Service) ClientVersionInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"clientVersionInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2Service) Conditions() CloudRunV2ServiceConditionsList {
	var returns CloudRunV2ServiceConditionsList
	_jsii_.Get(
		j,
		"conditions",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2Service) Connection() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"connection",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2Service) ConstructNodeMetadata() *map[string]interface{} {
	var returns *map[string]interface{}
	_jsii_.Get(
		j,
		"constructNodeMetadata",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2Service) Count() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"count",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2Service) CreateTime() *string {
	var returns *string
	_jsii_.Get(
		j,
		"createTime",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2Service) Creator() *string {
	var returns *string
	_jsii_.Get(
		j,
		"creator",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2Service) CustomAudiences() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"customAudiences",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2Service) CustomAudiencesInput() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"customAudiencesInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2Service) DeleteTime() *string {
	var returns *string
	_jsii_.Get(
		j,
		"deleteTime",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2Service) DependsOn() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"dependsOn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2Service) Description() *string {
	var returns *string
	_jsii_.Get(
		j,
		"description",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2Service) DescriptionInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"descriptionInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2Service) EffectiveAnnotations() cdktf.StringMap {
	var returns cdktf.StringMap
	_jsii_.Get(
		j,
		"effectiveAnnotations",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2Service) EffectiveLabels() cdktf.StringMap {
	var returns cdktf.StringMap
	_jsii_.Get(
		j,
		"effectiveLabels",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2Service) Etag() *string {
	var returns *string
	_jsii_.Get(
		j,
		"etag",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2Service) ExpireTime() *string {
	var returns *string
	_jsii_.Get(
		j,
		"expireTime",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2Service) ForEach() cdktf.ITerraformIterator {
	var returns cdktf.ITerraformIterator
	_jsii_.Get(
		j,
		"forEach",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2Service) Fqn() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fqn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2Service) FriendlyUniqueId() *string {
	var returns *string
	_jsii_.Get(
		j,
		"friendlyUniqueId",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2Service) Generation() *string {
	var returns *string
	_jsii_.Get(
		j,
		"generation",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2Service) Id() *string {
	var returns *string
	_jsii_.Get(
		j,
		"id",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2Service) IdInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"idInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2Service) Ingress() *string {
	var returns *string
	_jsii_.Get(
		j,
		"ingress",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2Service) IngressInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"ingressInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2Service) Labels() *map[string]*string {
	var returns *map[string]*string
	_jsii_.Get(
		j,
		"labels",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2Service) LabelsInput() *map[string]*string {
	var returns *map[string]*string
	_jsii_.Get(
		j,
		"labelsInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2Service) LastModifier() *string {
	var returns *string
	_jsii_.Get(
		j,
		"lastModifier",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2Service) LatestCreatedRevision() *string {
	var returns *string
	_jsii_.Get(
		j,
		"latestCreatedRevision",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2Service) LatestReadyRevision() *string {
	var returns *string
	_jsii_.Get(
		j,
		"latestReadyRevision",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2Service) LaunchStage() *string {
	var returns *string
	_jsii_.Get(
		j,
		"launchStage",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2Service) LaunchStageInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"launchStageInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2Service) Lifecycle() *cdktf.TerraformResourceLifecycle {
	var returns *cdktf.TerraformResourceLifecycle
	_jsii_.Get(
		j,
		"lifecycle",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2Service) Location() *string {
	var returns *string
	_jsii_.Get(
		j,
		"location",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2Service) LocationInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"locationInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2Service) Name() *string {
	var returns *string
	_jsii_.Get(
		j,
		"name",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2Service) NameInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"nameInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2Service) Node() constructs.Node {
	var returns constructs.Node
	_jsii_.Get(
		j,
		"node",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2Service) ObservedGeneration() *string {
	var returns *string
	_jsii_.Get(
		j,
		"observedGeneration",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2Service) Project() *string {
	var returns *string
	_jsii_.Get(
		j,
		"project",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2Service) ProjectInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"projectInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2Service) Provider() cdktf.TerraformProvider {
	var returns cdktf.TerraformProvider
	_jsii_.Get(
		j,
		"provider",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2Service) Provisioners() *[]interface{} {
	var returns *[]interface{}
	_jsii_.Get(
		j,
		"provisioners",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2Service) RawOverrides() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"rawOverrides",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2Service) Reconciling() cdktf.IResolvable {
	var returns cdktf.IResolvable
	_jsii_.Get(
		j,
		"reconciling",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2Service) Template() CloudRunV2ServiceTemplateOutputReference {
	var returns CloudRunV2ServiceTemplateOutputReference
	_jsii_.Get(
		j,
		"template",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2Service) TemplateInput() *CloudRunV2ServiceTemplate {
	var returns *CloudRunV2ServiceTemplate
	_jsii_.Get(
		j,
		"templateInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2Service) TerminalCondition() CloudRunV2ServiceTerminalConditionList {
	var returns CloudRunV2ServiceTerminalConditionList
	_jsii_.Get(
		j,
		"terminalCondition",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2Service) TerraformGeneratorMetadata() *cdktf.TerraformProviderGeneratorMetadata {
	var returns *cdktf.TerraformProviderGeneratorMetadata
	_jsii_.Get(
		j,
		"terraformGeneratorMetadata",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2Service) TerraformLabels() cdktf.StringMap {
	var returns cdktf.StringMap
	_jsii_.Get(
		j,
		"terraformLabels",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2Service) TerraformMetaArguments() *map[string]interface{} {
	var returns *map[string]interface{}
	_jsii_.Get(
		j,
		"terraformMetaArguments",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2Service) TerraformResourceType() *string {
	var returns *string
	_jsii_.Get(
		j,
		"terraformResourceType",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2Service) Timeouts() CloudRunV2ServiceTimeoutsOutputReference {
	var returns CloudRunV2ServiceTimeoutsOutputReference
	_jsii_.Get(
		j,
		"timeouts",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2Service) TimeoutsInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"timeoutsInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2Service) Traffic() CloudRunV2ServiceTrafficList {
	var returns CloudRunV2ServiceTrafficList
	_jsii_.Get(
		j,
		"traffic",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2Service) TrafficInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"trafficInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2Service) TrafficStatuses() CloudRunV2ServiceTrafficStatusesList {
	var returns CloudRunV2ServiceTrafficStatusesList
	_jsii_.Get(
		j,
		"trafficStatuses",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2Service) Uid() *string {
	var returns *string
	_jsii_.Get(
		j,
		"uid",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2Service) UpdateTime() *string {
	var returns *string
	_jsii_.Get(
		j,
		"updateTime",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2Service) Uri() *string {
	var returns *string
	_jsii_.Get(
		j,
		"uri",
		&returns,
	)
	return returns
}


// Create a new {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_run_v2_service google_cloud_run_v2_service} Resource.
func NewCloudRunV2Service(scope constructs.Construct, id *string, config *CloudRunV2ServiceConfig) CloudRunV2Service {
	_init_.Initialize()

	if err := validateNewCloudRunV2ServiceParameters(scope, id, config); err != nil {
		panic(err)
	}
	j := jsiiProxy_CloudRunV2Service{}

	_jsii_.Create(
		"@cdktf/provider-google.cloudRunV2Service.CloudRunV2Service",
		[]interface{}{scope, id, config},
		&j,
	)

	return &j
}

// Create a new {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_run_v2_service google_cloud_run_v2_service} Resource.
func NewCloudRunV2Service_Override(c CloudRunV2Service, scope constructs.Construct, id *string, config *CloudRunV2ServiceConfig) {
	_init_.Initialize()

	_jsii_.Create(
		"@cdktf/provider-google.cloudRunV2Service.CloudRunV2Service",
		[]interface{}{scope, id, config},
		c,
	)
}

func (j *jsiiProxy_CloudRunV2Service)SetAnnotations(val *map[string]*string) {
	if err := j.validateSetAnnotationsParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"annotations",
		val,
	)
}

func (j *jsiiProxy_CloudRunV2Service)SetClient(val *string) {
	if err := j.validateSetClientParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"client",
		val,
	)
}

func (j *jsiiProxy_CloudRunV2Service)SetClientVersion(val *string) {
	if err := j.validateSetClientVersionParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"clientVersion",
		val,
	)
}

func (j *jsiiProxy_CloudRunV2Service)SetConnection(val interface{}) {
	if err := j.validateSetConnectionParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"connection",
		val,
	)
}

func (j *jsiiProxy_CloudRunV2Service)SetCount(val interface{}) {
	if err := j.validateSetCountParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"count",
		val,
	)
}

func (j *jsiiProxy_CloudRunV2Service)SetCustomAudiences(val *[]*string) {
	if err := j.validateSetCustomAudiencesParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"customAudiences",
		val,
	)
}

func (j *jsiiProxy_CloudRunV2Service)SetDependsOn(val *[]*string) {
	_jsii_.Set(
		j,
		"dependsOn",
		val,
	)
}

func (j *jsiiProxy_CloudRunV2Service)SetDescription(val *string) {
	if err := j.validateSetDescriptionParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"description",
		val,
	)
}

func (j *jsiiProxy_CloudRunV2Service)SetForEach(val cdktf.ITerraformIterator) {
	_jsii_.Set(
		j,
		"forEach",
		val,
	)
}

func (j *jsiiProxy_CloudRunV2Service)SetId(val *string) {
	if err := j.validateSetIdParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"id",
		val,
	)
}

func (j *jsiiProxy_CloudRunV2Service)SetIngress(val *string) {
	if err := j.validateSetIngressParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"ingress",
		val,
	)
}

func (j *jsiiProxy_CloudRunV2Service)SetLabels(val *map[string]*string) {
	if err := j.validateSetLabelsParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"labels",
		val,
	)
}

func (j *jsiiProxy_CloudRunV2Service)SetLaunchStage(val *string) {
	if err := j.validateSetLaunchStageParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"launchStage",
		val,
	)
}

func (j *jsiiProxy_CloudRunV2Service)SetLifecycle(val *cdktf.TerraformResourceLifecycle) {
	if err := j.validateSetLifecycleParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"lifecycle",
		val,
	)
}

func (j *jsiiProxy_CloudRunV2Service)SetLocation(val *string) {
	if err := j.validateSetLocationParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"location",
		val,
	)
}

func (j *jsiiProxy_CloudRunV2Service)SetName(val *string) {
	if err := j.validateSetNameParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"name",
		val,
	)
}

func (j *jsiiProxy_CloudRunV2Service)SetProject(val *string) {
	if err := j.validateSetProjectParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"project",
		val,
	)
}

func (j *jsiiProxy_CloudRunV2Service)SetProvider(val cdktf.TerraformProvider) {
	_jsii_.Set(
		j,
		"provider",
		val,
	)
}

func (j *jsiiProxy_CloudRunV2Service)SetProvisioners(val *[]interface{}) {
	if err := j.validateSetProvisionersParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"provisioners",
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
func CloudRunV2Service_IsConstruct(x interface{}) *bool {
	_init_.Initialize()

	if err := validateCloudRunV2Service_IsConstructParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"@cdktf/provider-google.cloudRunV2Service.CloudRunV2Service",
		"isConstruct",
		[]interface{}{x},
		&returns,
	)

	return returns
}

// Experimental.
func CloudRunV2Service_IsTerraformElement(x interface{}) *bool {
	_init_.Initialize()

	if err := validateCloudRunV2Service_IsTerraformElementParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"@cdktf/provider-google.cloudRunV2Service.CloudRunV2Service",
		"isTerraformElement",
		[]interface{}{x},
		&returns,
	)

	return returns
}

// Experimental.
func CloudRunV2Service_IsTerraformResource(x interface{}) *bool {
	_init_.Initialize()

	if err := validateCloudRunV2Service_IsTerraformResourceParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"@cdktf/provider-google.cloudRunV2Service.CloudRunV2Service",
		"isTerraformResource",
		[]interface{}{x},
		&returns,
	)

	return returns
}

func CloudRunV2Service_TfResourceType() *string {
	_init_.Initialize()
	var returns *string
	_jsii_.StaticGet(
		"@cdktf/provider-google.cloudRunV2Service.CloudRunV2Service",
		"tfResourceType",
		&returns,
	)
	return returns
}

func (c *jsiiProxy_CloudRunV2Service) AddOverride(path *string, value interface{}) {
	if err := c.validateAddOverrideParameters(path, value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"addOverride",
		[]interface{}{path, value},
	)
}

func (c *jsiiProxy_CloudRunV2Service) GetAnyMapAttribute(terraformAttribute *string) *map[string]interface{} {
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

func (c *jsiiProxy_CloudRunV2Service) GetBooleanAttribute(terraformAttribute *string) cdktf.IResolvable {
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

func (c *jsiiProxy_CloudRunV2Service) GetBooleanMapAttribute(terraformAttribute *string) *map[string]*bool {
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

func (c *jsiiProxy_CloudRunV2Service) GetListAttribute(terraformAttribute *string) *[]*string {
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

func (c *jsiiProxy_CloudRunV2Service) GetNumberAttribute(terraformAttribute *string) *float64 {
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

func (c *jsiiProxy_CloudRunV2Service) GetNumberListAttribute(terraformAttribute *string) *[]*float64 {
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

func (c *jsiiProxy_CloudRunV2Service) GetNumberMapAttribute(terraformAttribute *string) *map[string]*float64 {
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

func (c *jsiiProxy_CloudRunV2Service) GetStringAttribute(terraformAttribute *string) *string {
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

func (c *jsiiProxy_CloudRunV2Service) GetStringMapAttribute(terraformAttribute *string) *map[string]*string {
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

func (c *jsiiProxy_CloudRunV2Service) InterpolationForAttribute(terraformAttribute *string) cdktf.IResolvable {
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

func (c *jsiiProxy_CloudRunV2Service) OverrideLogicalId(newLogicalId *string) {
	if err := c.validateOverrideLogicalIdParameters(newLogicalId); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"overrideLogicalId",
		[]interface{}{newLogicalId},
	)
}

func (c *jsiiProxy_CloudRunV2Service) PutBinaryAuthorization(value *CloudRunV2ServiceBinaryAuthorization) {
	if err := c.validatePutBinaryAuthorizationParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putBinaryAuthorization",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_CloudRunV2Service) PutTemplate(value *CloudRunV2ServiceTemplate) {
	if err := c.validatePutTemplateParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putTemplate",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_CloudRunV2Service) PutTimeouts(value *CloudRunV2ServiceTimeouts) {
	if err := c.validatePutTimeoutsParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putTimeouts",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_CloudRunV2Service) PutTraffic(value interface{}) {
	if err := c.validatePutTrafficParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putTraffic",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_CloudRunV2Service) ResetAnnotations() {
	_jsii_.InvokeVoid(
		c,
		"resetAnnotations",
		nil, // no parameters
	)
}

func (c *jsiiProxy_CloudRunV2Service) ResetBinaryAuthorization() {
	_jsii_.InvokeVoid(
		c,
		"resetBinaryAuthorization",
		nil, // no parameters
	)
}

func (c *jsiiProxy_CloudRunV2Service) ResetClient() {
	_jsii_.InvokeVoid(
		c,
		"resetClient",
		nil, // no parameters
	)
}

func (c *jsiiProxy_CloudRunV2Service) ResetClientVersion() {
	_jsii_.InvokeVoid(
		c,
		"resetClientVersion",
		nil, // no parameters
	)
}

func (c *jsiiProxy_CloudRunV2Service) ResetCustomAudiences() {
	_jsii_.InvokeVoid(
		c,
		"resetCustomAudiences",
		nil, // no parameters
	)
}

func (c *jsiiProxy_CloudRunV2Service) ResetDescription() {
	_jsii_.InvokeVoid(
		c,
		"resetDescription",
		nil, // no parameters
	)
}

func (c *jsiiProxy_CloudRunV2Service) ResetId() {
	_jsii_.InvokeVoid(
		c,
		"resetId",
		nil, // no parameters
	)
}

func (c *jsiiProxy_CloudRunV2Service) ResetIngress() {
	_jsii_.InvokeVoid(
		c,
		"resetIngress",
		nil, // no parameters
	)
}

func (c *jsiiProxy_CloudRunV2Service) ResetLabels() {
	_jsii_.InvokeVoid(
		c,
		"resetLabels",
		nil, // no parameters
	)
}

func (c *jsiiProxy_CloudRunV2Service) ResetLaunchStage() {
	_jsii_.InvokeVoid(
		c,
		"resetLaunchStage",
		nil, // no parameters
	)
}

func (c *jsiiProxy_CloudRunV2Service) ResetOverrideLogicalId() {
	_jsii_.InvokeVoid(
		c,
		"resetOverrideLogicalId",
		nil, // no parameters
	)
}

func (c *jsiiProxy_CloudRunV2Service) ResetProject() {
	_jsii_.InvokeVoid(
		c,
		"resetProject",
		nil, // no parameters
	)
}

func (c *jsiiProxy_CloudRunV2Service) ResetTimeouts() {
	_jsii_.InvokeVoid(
		c,
		"resetTimeouts",
		nil, // no parameters
	)
}

func (c *jsiiProxy_CloudRunV2Service) ResetTraffic() {
	_jsii_.InvokeVoid(
		c,
		"resetTraffic",
		nil, // no parameters
	)
}

func (c *jsiiProxy_CloudRunV2Service) SynthesizeAttributes() *map[string]interface{} {
	var returns *map[string]interface{}

	_jsii_.Invoke(
		c,
		"synthesizeAttributes",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (c *jsiiProxy_CloudRunV2Service) ToMetadata() interface{} {
	var returns interface{}

	_jsii_.Invoke(
		c,
		"toMetadata",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (c *jsiiProxy_CloudRunV2Service) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		c,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (c *jsiiProxy_CloudRunV2Service) ToTerraform() interface{} {
	var returns interface{}

	_jsii_.Invoke(
		c,
		"toTerraform",
		nil, // no parameters
		&returns,
	)

	return returns
}

