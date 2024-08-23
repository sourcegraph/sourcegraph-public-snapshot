package monitoringuptimecheckconfig

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/sourcegraph/managed-services-platform-cdktf/gen/google/jsii"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/monitoringuptimecheckconfig/internal"
)

// Represents a {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_uptime_check_config google_monitoring_uptime_check_config}.
type MonitoringUptimeCheckConfig interface {
	cdktf.TerraformResource
	// Experimental.
	CdktfStack() cdktf.TerraformStack
	CheckerType() *string
	SetCheckerType(val *string)
	CheckerTypeInput() *string
	// Experimental.
	Connection() interface{}
	// Experimental.
	SetConnection(val interface{})
	// Experimental.
	ConstructNodeMetadata() *map[string]interface{}
	ContentMatchers() MonitoringUptimeCheckConfigContentMatchersList
	ContentMatchersInput() interface{}
	// Experimental.
	Count() interface{}
	// Experimental.
	SetCount(val interface{})
	// Experimental.
	DependsOn() *[]*string
	// Experimental.
	SetDependsOn(val *[]*string)
	DisplayName() *string
	SetDisplayName(val *string)
	DisplayNameInput() *string
	// Experimental.
	ForEach() cdktf.ITerraformIterator
	// Experimental.
	SetForEach(val cdktf.ITerraformIterator)
	// Experimental.
	Fqn() *string
	// Experimental.
	FriendlyUniqueId() *string
	HttpCheck() MonitoringUptimeCheckConfigHttpCheckOutputReference
	HttpCheckInput() *MonitoringUptimeCheckConfigHttpCheck
	Id() *string
	SetId(val *string)
	IdInput() *string
	// Experimental.
	Lifecycle() *cdktf.TerraformResourceLifecycle
	// Experimental.
	SetLifecycle(val *cdktf.TerraformResourceLifecycle)
	MonitoredResource() MonitoringUptimeCheckConfigMonitoredResourceOutputReference
	MonitoredResourceInput() *MonitoringUptimeCheckConfigMonitoredResource
	Name() *string
	// The tree node.
	Node() constructs.Node
	Period() *string
	SetPeriod(val *string)
	PeriodInput() *string
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
	ResourceGroup() MonitoringUptimeCheckConfigResourceGroupOutputReference
	ResourceGroupInput() *MonitoringUptimeCheckConfigResourceGroup
	SelectedRegions() *[]*string
	SetSelectedRegions(val *[]*string)
	SelectedRegionsInput() *[]*string
	SyntheticMonitor() MonitoringUptimeCheckConfigSyntheticMonitorOutputReference
	SyntheticMonitorInput() *MonitoringUptimeCheckConfigSyntheticMonitor
	TcpCheck() MonitoringUptimeCheckConfigTcpCheckOutputReference
	TcpCheckInput() *MonitoringUptimeCheckConfigTcpCheck
	// Experimental.
	TerraformGeneratorMetadata() *cdktf.TerraformProviderGeneratorMetadata
	// Experimental.
	TerraformMetaArguments() *map[string]interface{}
	// Experimental.
	TerraformResourceType() *string
	Timeout() *string
	SetTimeout(val *string)
	TimeoutInput() *string
	Timeouts() MonitoringUptimeCheckConfigTimeoutsOutputReference
	TimeoutsInput() interface{}
	UptimeCheckId() *string
	UserLabels() *map[string]*string
	SetUserLabels(val *map[string]*string)
	UserLabelsInput() *map[string]*string
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
	PutContentMatchers(value interface{})
	PutHttpCheck(value *MonitoringUptimeCheckConfigHttpCheck)
	PutMonitoredResource(value *MonitoringUptimeCheckConfigMonitoredResource)
	PutResourceGroup(value *MonitoringUptimeCheckConfigResourceGroup)
	PutSyntheticMonitor(value *MonitoringUptimeCheckConfigSyntheticMonitor)
	PutTcpCheck(value *MonitoringUptimeCheckConfigTcpCheck)
	PutTimeouts(value *MonitoringUptimeCheckConfigTimeouts)
	ResetCheckerType()
	ResetContentMatchers()
	ResetHttpCheck()
	ResetId()
	ResetMonitoredResource()
	// Resets a previously passed logical Id to use the auto-generated logical id again.
	// Experimental.
	ResetOverrideLogicalId()
	ResetPeriod()
	ResetProject()
	ResetResourceGroup()
	ResetSelectedRegions()
	ResetSyntheticMonitor()
	ResetTcpCheck()
	ResetTimeouts()
	ResetUserLabels()
	SynthesizeAttributes() *map[string]interface{}
	// Experimental.
	ToMetadata() interface{}
	// Returns a string representation of this construct.
	ToString() *string
	// Adds this resource to the terraform JSON output.
	// Experimental.
	ToTerraform() interface{}
}

// The jsii proxy struct for MonitoringUptimeCheckConfig
type jsiiProxy_MonitoringUptimeCheckConfig struct {
	internal.Type__cdktfTerraformResource
}

func (j *jsiiProxy_MonitoringUptimeCheckConfig) CdktfStack() cdktf.TerraformStack {
	var returns cdktf.TerraformStack
	_jsii_.Get(
		j,
		"cdktfStack",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfig) CheckerType() *string {
	var returns *string
	_jsii_.Get(
		j,
		"checkerType",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfig) CheckerTypeInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"checkerTypeInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfig) Connection() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"connection",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfig) ConstructNodeMetadata() *map[string]interface{} {
	var returns *map[string]interface{}
	_jsii_.Get(
		j,
		"constructNodeMetadata",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfig) ContentMatchers() MonitoringUptimeCheckConfigContentMatchersList {
	var returns MonitoringUptimeCheckConfigContentMatchersList
	_jsii_.Get(
		j,
		"contentMatchers",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfig) ContentMatchersInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"contentMatchersInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfig) Count() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"count",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfig) DependsOn() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"dependsOn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfig) DisplayName() *string {
	var returns *string
	_jsii_.Get(
		j,
		"displayName",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfig) DisplayNameInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"displayNameInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfig) ForEach() cdktf.ITerraformIterator {
	var returns cdktf.ITerraformIterator
	_jsii_.Get(
		j,
		"forEach",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfig) Fqn() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fqn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfig) FriendlyUniqueId() *string {
	var returns *string
	_jsii_.Get(
		j,
		"friendlyUniqueId",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfig) HttpCheck() MonitoringUptimeCheckConfigHttpCheckOutputReference {
	var returns MonitoringUptimeCheckConfigHttpCheckOutputReference
	_jsii_.Get(
		j,
		"httpCheck",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfig) HttpCheckInput() *MonitoringUptimeCheckConfigHttpCheck {
	var returns *MonitoringUptimeCheckConfigHttpCheck
	_jsii_.Get(
		j,
		"httpCheckInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfig) Id() *string {
	var returns *string
	_jsii_.Get(
		j,
		"id",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfig) IdInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"idInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfig) Lifecycle() *cdktf.TerraformResourceLifecycle {
	var returns *cdktf.TerraformResourceLifecycle
	_jsii_.Get(
		j,
		"lifecycle",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfig) MonitoredResource() MonitoringUptimeCheckConfigMonitoredResourceOutputReference {
	var returns MonitoringUptimeCheckConfigMonitoredResourceOutputReference
	_jsii_.Get(
		j,
		"monitoredResource",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfig) MonitoredResourceInput() *MonitoringUptimeCheckConfigMonitoredResource {
	var returns *MonitoringUptimeCheckConfigMonitoredResource
	_jsii_.Get(
		j,
		"monitoredResourceInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfig) Name() *string {
	var returns *string
	_jsii_.Get(
		j,
		"name",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfig) Node() constructs.Node {
	var returns constructs.Node
	_jsii_.Get(
		j,
		"node",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfig) Period() *string {
	var returns *string
	_jsii_.Get(
		j,
		"period",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfig) PeriodInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"periodInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfig) Project() *string {
	var returns *string
	_jsii_.Get(
		j,
		"project",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfig) ProjectInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"projectInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfig) Provider() cdktf.TerraformProvider {
	var returns cdktf.TerraformProvider
	_jsii_.Get(
		j,
		"provider",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfig) Provisioners() *[]interface{} {
	var returns *[]interface{}
	_jsii_.Get(
		j,
		"provisioners",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfig) RawOverrides() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"rawOverrides",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfig) ResourceGroup() MonitoringUptimeCheckConfigResourceGroupOutputReference {
	var returns MonitoringUptimeCheckConfigResourceGroupOutputReference
	_jsii_.Get(
		j,
		"resourceGroup",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfig) ResourceGroupInput() *MonitoringUptimeCheckConfigResourceGroup {
	var returns *MonitoringUptimeCheckConfigResourceGroup
	_jsii_.Get(
		j,
		"resourceGroupInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfig) SelectedRegions() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"selectedRegions",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfig) SelectedRegionsInput() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"selectedRegionsInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfig) SyntheticMonitor() MonitoringUptimeCheckConfigSyntheticMonitorOutputReference {
	var returns MonitoringUptimeCheckConfigSyntheticMonitorOutputReference
	_jsii_.Get(
		j,
		"syntheticMonitor",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfig) SyntheticMonitorInput() *MonitoringUptimeCheckConfigSyntheticMonitor {
	var returns *MonitoringUptimeCheckConfigSyntheticMonitor
	_jsii_.Get(
		j,
		"syntheticMonitorInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfig) TcpCheck() MonitoringUptimeCheckConfigTcpCheckOutputReference {
	var returns MonitoringUptimeCheckConfigTcpCheckOutputReference
	_jsii_.Get(
		j,
		"tcpCheck",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfig) TcpCheckInput() *MonitoringUptimeCheckConfigTcpCheck {
	var returns *MonitoringUptimeCheckConfigTcpCheck
	_jsii_.Get(
		j,
		"tcpCheckInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfig) TerraformGeneratorMetadata() *cdktf.TerraformProviderGeneratorMetadata {
	var returns *cdktf.TerraformProviderGeneratorMetadata
	_jsii_.Get(
		j,
		"terraformGeneratorMetadata",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfig) TerraformMetaArguments() *map[string]interface{} {
	var returns *map[string]interface{}
	_jsii_.Get(
		j,
		"terraformMetaArguments",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfig) TerraformResourceType() *string {
	var returns *string
	_jsii_.Get(
		j,
		"terraformResourceType",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfig) Timeout() *string {
	var returns *string
	_jsii_.Get(
		j,
		"timeout",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfig) TimeoutInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"timeoutInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfig) Timeouts() MonitoringUptimeCheckConfigTimeoutsOutputReference {
	var returns MonitoringUptimeCheckConfigTimeoutsOutputReference
	_jsii_.Get(
		j,
		"timeouts",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfig) TimeoutsInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"timeoutsInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfig) UptimeCheckId() *string {
	var returns *string
	_jsii_.Get(
		j,
		"uptimeCheckId",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfig) UserLabels() *map[string]*string {
	var returns *map[string]*string
	_jsii_.Get(
		j,
		"userLabels",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfig) UserLabelsInput() *map[string]*string {
	var returns *map[string]*string
	_jsii_.Get(
		j,
		"userLabelsInput",
		&returns,
	)
	return returns
}


// Create a new {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_uptime_check_config google_monitoring_uptime_check_config} Resource.
func NewMonitoringUptimeCheckConfig(scope constructs.Construct, id *string, config *MonitoringUptimeCheckConfigConfig) MonitoringUptimeCheckConfig {
	_init_.Initialize()

	if err := validateNewMonitoringUptimeCheckConfigParameters(scope, id, config); err != nil {
		panic(err)
	}
	j := jsiiProxy_MonitoringUptimeCheckConfig{}

	_jsii_.Create(
		"@cdktf/provider-google.monitoringUptimeCheckConfig.MonitoringUptimeCheckConfig",
		[]interface{}{scope, id, config},
		&j,
	)

	return &j
}

// Create a new {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_uptime_check_config google_monitoring_uptime_check_config} Resource.
func NewMonitoringUptimeCheckConfig_Override(m MonitoringUptimeCheckConfig, scope constructs.Construct, id *string, config *MonitoringUptimeCheckConfigConfig) {
	_init_.Initialize()

	_jsii_.Create(
		"@cdktf/provider-google.monitoringUptimeCheckConfig.MonitoringUptimeCheckConfig",
		[]interface{}{scope, id, config},
		m,
	)
}

func (j *jsiiProxy_MonitoringUptimeCheckConfig)SetCheckerType(val *string) {
	if err := j.validateSetCheckerTypeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"checkerType",
		val,
	)
}

func (j *jsiiProxy_MonitoringUptimeCheckConfig)SetConnection(val interface{}) {
	if err := j.validateSetConnectionParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"connection",
		val,
	)
}

func (j *jsiiProxy_MonitoringUptimeCheckConfig)SetCount(val interface{}) {
	if err := j.validateSetCountParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"count",
		val,
	)
}

func (j *jsiiProxy_MonitoringUptimeCheckConfig)SetDependsOn(val *[]*string) {
	_jsii_.Set(
		j,
		"dependsOn",
		val,
	)
}

func (j *jsiiProxy_MonitoringUptimeCheckConfig)SetDisplayName(val *string) {
	if err := j.validateSetDisplayNameParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"displayName",
		val,
	)
}

func (j *jsiiProxy_MonitoringUptimeCheckConfig)SetForEach(val cdktf.ITerraformIterator) {
	_jsii_.Set(
		j,
		"forEach",
		val,
	)
}

func (j *jsiiProxy_MonitoringUptimeCheckConfig)SetId(val *string) {
	if err := j.validateSetIdParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"id",
		val,
	)
}

func (j *jsiiProxy_MonitoringUptimeCheckConfig)SetLifecycle(val *cdktf.TerraformResourceLifecycle) {
	if err := j.validateSetLifecycleParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"lifecycle",
		val,
	)
}

func (j *jsiiProxy_MonitoringUptimeCheckConfig)SetPeriod(val *string) {
	if err := j.validateSetPeriodParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"period",
		val,
	)
}

func (j *jsiiProxy_MonitoringUptimeCheckConfig)SetProject(val *string) {
	if err := j.validateSetProjectParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"project",
		val,
	)
}

func (j *jsiiProxy_MonitoringUptimeCheckConfig)SetProvider(val cdktf.TerraformProvider) {
	_jsii_.Set(
		j,
		"provider",
		val,
	)
}

func (j *jsiiProxy_MonitoringUptimeCheckConfig)SetProvisioners(val *[]interface{}) {
	if err := j.validateSetProvisionersParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"provisioners",
		val,
	)
}

func (j *jsiiProxy_MonitoringUptimeCheckConfig)SetSelectedRegions(val *[]*string) {
	if err := j.validateSetSelectedRegionsParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"selectedRegions",
		val,
	)
}

func (j *jsiiProxy_MonitoringUptimeCheckConfig)SetTimeout(val *string) {
	if err := j.validateSetTimeoutParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"timeout",
		val,
	)
}

func (j *jsiiProxy_MonitoringUptimeCheckConfig)SetUserLabels(val *map[string]*string) {
	if err := j.validateSetUserLabelsParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"userLabels",
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
func MonitoringUptimeCheckConfig_IsConstruct(x interface{}) *bool {
	_init_.Initialize()

	if err := validateMonitoringUptimeCheckConfig_IsConstructParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"@cdktf/provider-google.monitoringUptimeCheckConfig.MonitoringUptimeCheckConfig",
		"isConstruct",
		[]interface{}{x},
		&returns,
	)

	return returns
}

// Experimental.
func MonitoringUptimeCheckConfig_IsTerraformElement(x interface{}) *bool {
	_init_.Initialize()

	if err := validateMonitoringUptimeCheckConfig_IsTerraformElementParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"@cdktf/provider-google.monitoringUptimeCheckConfig.MonitoringUptimeCheckConfig",
		"isTerraformElement",
		[]interface{}{x},
		&returns,
	)

	return returns
}

// Experimental.
func MonitoringUptimeCheckConfig_IsTerraformResource(x interface{}) *bool {
	_init_.Initialize()

	if err := validateMonitoringUptimeCheckConfig_IsTerraformResourceParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"@cdktf/provider-google.monitoringUptimeCheckConfig.MonitoringUptimeCheckConfig",
		"isTerraformResource",
		[]interface{}{x},
		&returns,
	)

	return returns
}

func MonitoringUptimeCheckConfig_TfResourceType() *string {
	_init_.Initialize()
	var returns *string
	_jsii_.StaticGet(
		"@cdktf/provider-google.monitoringUptimeCheckConfig.MonitoringUptimeCheckConfig",
		"tfResourceType",
		&returns,
	)
	return returns
}

func (m *jsiiProxy_MonitoringUptimeCheckConfig) AddOverride(path *string, value interface{}) {
	if err := m.validateAddOverrideParameters(path, value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		m,
		"addOverride",
		[]interface{}{path, value},
	)
}

func (m *jsiiProxy_MonitoringUptimeCheckConfig) GetAnyMapAttribute(terraformAttribute *string) *map[string]interface{} {
	if err := m.validateGetAnyMapAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *map[string]interface{}

	_jsii_.Invoke(
		m,
		"getAnyMapAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (m *jsiiProxy_MonitoringUptimeCheckConfig) GetBooleanAttribute(terraformAttribute *string) cdktf.IResolvable {
	if err := m.validateGetBooleanAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		m,
		"getBooleanAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (m *jsiiProxy_MonitoringUptimeCheckConfig) GetBooleanMapAttribute(terraformAttribute *string) *map[string]*bool {
	if err := m.validateGetBooleanMapAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *map[string]*bool

	_jsii_.Invoke(
		m,
		"getBooleanMapAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (m *jsiiProxy_MonitoringUptimeCheckConfig) GetListAttribute(terraformAttribute *string) *[]*string {
	if err := m.validateGetListAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *[]*string

	_jsii_.Invoke(
		m,
		"getListAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (m *jsiiProxy_MonitoringUptimeCheckConfig) GetNumberAttribute(terraformAttribute *string) *float64 {
	if err := m.validateGetNumberAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *float64

	_jsii_.Invoke(
		m,
		"getNumberAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (m *jsiiProxy_MonitoringUptimeCheckConfig) GetNumberListAttribute(terraformAttribute *string) *[]*float64 {
	if err := m.validateGetNumberListAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *[]*float64

	_jsii_.Invoke(
		m,
		"getNumberListAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (m *jsiiProxy_MonitoringUptimeCheckConfig) GetNumberMapAttribute(terraformAttribute *string) *map[string]*float64 {
	if err := m.validateGetNumberMapAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *map[string]*float64

	_jsii_.Invoke(
		m,
		"getNumberMapAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (m *jsiiProxy_MonitoringUptimeCheckConfig) GetStringAttribute(terraformAttribute *string) *string {
	if err := m.validateGetStringAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.Invoke(
		m,
		"getStringAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (m *jsiiProxy_MonitoringUptimeCheckConfig) GetStringMapAttribute(terraformAttribute *string) *map[string]*string {
	if err := m.validateGetStringMapAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *map[string]*string

	_jsii_.Invoke(
		m,
		"getStringMapAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (m *jsiiProxy_MonitoringUptimeCheckConfig) InterpolationForAttribute(terraformAttribute *string) cdktf.IResolvable {
	if err := m.validateInterpolationForAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		m,
		"interpolationForAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (m *jsiiProxy_MonitoringUptimeCheckConfig) OverrideLogicalId(newLogicalId *string) {
	if err := m.validateOverrideLogicalIdParameters(newLogicalId); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		m,
		"overrideLogicalId",
		[]interface{}{newLogicalId},
	)
}

func (m *jsiiProxy_MonitoringUptimeCheckConfig) PutContentMatchers(value interface{}) {
	if err := m.validatePutContentMatchersParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		m,
		"putContentMatchers",
		[]interface{}{value},
	)
}

func (m *jsiiProxy_MonitoringUptimeCheckConfig) PutHttpCheck(value *MonitoringUptimeCheckConfigHttpCheck) {
	if err := m.validatePutHttpCheckParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		m,
		"putHttpCheck",
		[]interface{}{value},
	)
}

func (m *jsiiProxy_MonitoringUptimeCheckConfig) PutMonitoredResource(value *MonitoringUptimeCheckConfigMonitoredResource) {
	if err := m.validatePutMonitoredResourceParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		m,
		"putMonitoredResource",
		[]interface{}{value},
	)
}

func (m *jsiiProxy_MonitoringUptimeCheckConfig) PutResourceGroup(value *MonitoringUptimeCheckConfigResourceGroup) {
	if err := m.validatePutResourceGroupParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		m,
		"putResourceGroup",
		[]interface{}{value},
	)
}

func (m *jsiiProxy_MonitoringUptimeCheckConfig) PutSyntheticMonitor(value *MonitoringUptimeCheckConfigSyntheticMonitor) {
	if err := m.validatePutSyntheticMonitorParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		m,
		"putSyntheticMonitor",
		[]interface{}{value},
	)
}

func (m *jsiiProxy_MonitoringUptimeCheckConfig) PutTcpCheck(value *MonitoringUptimeCheckConfigTcpCheck) {
	if err := m.validatePutTcpCheckParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		m,
		"putTcpCheck",
		[]interface{}{value},
	)
}

func (m *jsiiProxy_MonitoringUptimeCheckConfig) PutTimeouts(value *MonitoringUptimeCheckConfigTimeouts) {
	if err := m.validatePutTimeoutsParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		m,
		"putTimeouts",
		[]interface{}{value},
	)
}

func (m *jsiiProxy_MonitoringUptimeCheckConfig) ResetCheckerType() {
	_jsii_.InvokeVoid(
		m,
		"resetCheckerType",
		nil, // no parameters
	)
}

func (m *jsiiProxy_MonitoringUptimeCheckConfig) ResetContentMatchers() {
	_jsii_.InvokeVoid(
		m,
		"resetContentMatchers",
		nil, // no parameters
	)
}

func (m *jsiiProxy_MonitoringUptimeCheckConfig) ResetHttpCheck() {
	_jsii_.InvokeVoid(
		m,
		"resetHttpCheck",
		nil, // no parameters
	)
}

func (m *jsiiProxy_MonitoringUptimeCheckConfig) ResetId() {
	_jsii_.InvokeVoid(
		m,
		"resetId",
		nil, // no parameters
	)
}

func (m *jsiiProxy_MonitoringUptimeCheckConfig) ResetMonitoredResource() {
	_jsii_.InvokeVoid(
		m,
		"resetMonitoredResource",
		nil, // no parameters
	)
}

func (m *jsiiProxy_MonitoringUptimeCheckConfig) ResetOverrideLogicalId() {
	_jsii_.InvokeVoid(
		m,
		"resetOverrideLogicalId",
		nil, // no parameters
	)
}

func (m *jsiiProxy_MonitoringUptimeCheckConfig) ResetPeriod() {
	_jsii_.InvokeVoid(
		m,
		"resetPeriod",
		nil, // no parameters
	)
}

func (m *jsiiProxy_MonitoringUptimeCheckConfig) ResetProject() {
	_jsii_.InvokeVoid(
		m,
		"resetProject",
		nil, // no parameters
	)
}

func (m *jsiiProxy_MonitoringUptimeCheckConfig) ResetResourceGroup() {
	_jsii_.InvokeVoid(
		m,
		"resetResourceGroup",
		nil, // no parameters
	)
}

func (m *jsiiProxy_MonitoringUptimeCheckConfig) ResetSelectedRegions() {
	_jsii_.InvokeVoid(
		m,
		"resetSelectedRegions",
		nil, // no parameters
	)
}

func (m *jsiiProxy_MonitoringUptimeCheckConfig) ResetSyntheticMonitor() {
	_jsii_.InvokeVoid(
		m,
		"resetSyntheticMonitor",
		nil, // no parameters
	)
}

func (m *jsiiProxy_MonitoringUptimeCheckConfig) ResetTcpCheck() {
	_jsii_.InvokeVoid(
		m,
		"resetTcpCheck",
		nil, // no parameters
	)
}

func (m *jsiiProxy_MonitoringUptimeCheckConfig) ResetTimeouts() {
	_jsii_.InvokeVoid(
		m,
		"resetTimeouts",
		nil, // no parameters
	)
}

func (m *jsiiProxy_MonitoringUptimeCheckConfig) ResetUserLabels() {
	_jsii_.InvokeVoid(
		m,
		"resetUserLabels",
		nil, // no parameters
	)
}

func (m *jsiiProxy_MonitoringUptimeCheckConfig) SynthesizeAttributes() *map[string]interface{} {
	var returns *map[string]interface{}

	_jsii_.Invoke(
		m,
		"synthesizeAttributes",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (m *jsiiProxy_MonitoringUptimeCheckConfig) ToMetadata() interface{} {
	var returns interface{}

	_jsii_.Invoke(
		m,
		"toMetadata",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (m *jsiiProxy_MonitoringUptimeCheckConfig) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		m,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (m *jsiiProxy_MonitoringUptimeCheckConfig) ToTerraform() interface{} {
	var returns interface{}

	_jsii_.Invoke(
		m,
		"toTerraform",
		nil, // no parameters
		&returns,
	)

	return returns
}

