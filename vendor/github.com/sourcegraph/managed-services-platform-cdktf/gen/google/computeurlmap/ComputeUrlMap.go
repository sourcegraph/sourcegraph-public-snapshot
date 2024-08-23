package computeurlmap

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/sourcegraph/managed-services-platform-cdktf/gen/google/jsii"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/computeurlmap/internal"
)

// Represents a {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_url_map google_compute_url_map}.
type ComputeUrlMap interface {
	cdktf.TerraformResource
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
	CreationTimestamp() *string
	DefaultRouteAction() ComputeUrlMapDefaultRouteActionOutputReference
	DefaultRouteActionInput() *ComputeUrlMapDefaultRouteAction
	DefaultService() *string
	SetDefaultService(val *string)
	DefaultServiceInput() *string
	DefaultUrlRedirect() ComputeUrlMapDefaultUrlRedirectOutputReference
	DefaultUrlRedirectInput() *ComputeUrlMapDefaultUrlRedirect
	// Experimental.
	DependsOn() *[]*string
	// Experimental.
	SetDependsOn(val *[]*string)
	Description() *string
	SetDescription(val *string)
	DescriptionInput() *string
	Fingerprint() *string
	// Experimental.
	ForEach() cdktf.ITerraformIterator
	// Experimental.
	SetForEach(val cdktf.ITerraformIterator)
	// Experimental.
	Fqn() *string
	// Experimental.
	FriendlyUniqueId() *string
	HeaderAction() ComputeUrlMapHeaderActionOutputReference
	HeaderActionInput() *ComputeUrlMapHeaderAction
	HostRule() ComputeUrlMapHostRuleList
	HostRuleInput() interface{}
	Id() *string
	SetId(val *string)
	IdInput() *string
	// Experimental.
	Lifecycle() *cdktf.TerraformResourceLifecycle
	// Experimental.
	SetLifecycle(val *cdktf.TerraformResourceLifecycle)
	MapId() *float64
	Name() *string
	SetName(val *string)
	NameInput() *string
	// The tree node.
	Node() constructs.Node
	PathMatcher() ComputeUrlMapPathMatcherList
	PathMatcherInput() interface{}
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
	SelfLink() *string
	// Experimental.
	TerraformGeneratorMetadata() *cdktf.TerraformProviderGeneratorMetadata
	// Experimental.
	TerraformMetaArguments() *map[string]interface{}
	// Experimental.
	TerraformResourceType() *string
	Test() ComputeUrlMapTestList
	TestInput() interface{}
	Timeouts() ComputeUrlMapTimeoutsOutputReference
	TimeoutsInput() interface{}
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
	PutDefaultRouteAction(value *ComputeUrlMapDefaultRouteAction)
	PutDefaultUrlRedirect(value *ComputeUrlMapDefaultUrlRedirect)
	PutHeaderAction(value *ComputeUrlMapHeaderAction)
	PutHostRule(value interface{})
	PutPathMatcher(value interface{})
	PutTest(value interface{})
	PutTimeouts(value *ComputeUrlMapTimeouts)
	ResetDefaultRouteAction()
	ResetDefaultService()
	ResetDefaultUrlRedirect()
	ResetDescription()
	ResetHeaderAction()
	ResetHostRule()
	ResetId()
	// Resets a previously passed logical Id to use the auto-generated logical id again.
	// Experimental.
	ResetOverrideLogicalId()
	ResetPathMatcher()
	ResetProject()
	ResetTest()
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

// The jsii proxy struct for ComputeUrlMap
type jsiiProxy_ComputeUrlMap struct {
	internal.Type__cdktfTerraformResource
}

func (j *jsiiProxy_ComputeUrlMap) CdktfStack() cdktf.TerraformStack {
	var returns cdktf.TerraformStack
	_jsii_.Get(
		j,
		"cdktfStack",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMap) Connection() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"connection",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMap) ConstructNodeMetadata() *map[string]interface{} {
	var returns *map[string]interface{}
	_jsii_.Get(
		j,
		"constructNodeMetadata",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMap) Count() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"count",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMap) CreationTimestamp() *string {
	var returns *string
	_jsii_.Get(
		j,
		"creationTimestamp",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMap) DefaultRouteAction() ComputeUrlMapDefaultRouteActionOutputReference {
	var returns ComputeUrlMapDefaultRouteActionOutputReference
	_jsii_.Get(
		j,
		"defaultRouteAction",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMap) DefaultRouteActionInput() *ComputeUrlMapDefaultRouteAction {
	var returns *ComputeUrlMapDefaultRouteAction
	_jsii_.Get(
		j,
		"defaultRouteActionInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMap) DefaultService() *string {
	var returns *string
	_jsii_.Get(
		j,
		"defaultService",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMap) DefaultServiceInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"defaultServiceInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMap) DefaultUrlRedirect() ComputeUrlMapDefaultUrlRedirectOutputReference {
	var returns ComputeUrlMapDefaultUrlRedirectOutputReference
	_jsii_.Get(
		j,
		"defaultUrlRedirect",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMap) DefaultUrlRedirectInput() *ComputeUrlMapDefaultUrlRedirect {
	var returns *ComputeUrlMapDefaultUrlRedirect
	_jsii_.Get(
		j,
		"defaultUrlRedirectInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMap) DependsOn() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"dependsOn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMap) Description() *string {
	var returns *string
	_jsii_.Get(
		j,
		"description",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMap) DescriptionInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"descriptionInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMap) Fingerprint() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fingerprint",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMap) ForEach() cdktf.ITerraformIterator {
	var returns cdktf.ITerraformIterator
	_jsii_.Get(
		j,
		"forEach",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMap) Fqn() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fqn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMap) FriendlyUniqueId() *string {
	var returns *string
	_jsii_.Get(
		j,
		"friendlyUniqueId",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMap) HeaderAction() ComputeUrlMapHeaderActionOutputReference {
	var returns ComputeUrlMapHeaderActionOutputReference
	_jsii_.Get(
		j,
		"headerAction",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMap) HeaderActionInput() *ComputeUrlMapHeaderAction {
	var returns *ComputeUrlMapHeaderAction
	_jsii_.Get(
		j,
		"headerActionInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMap) HostRule() ComputeUrlMapHostRuleList {
	var returns ComputeUrlMapHostRuleList
	_jsii_.Get(
		j,
		"hostRule",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMap) HostRuleInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"hostRuleInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMap) Id() *string {
	var returns *string
	_jsii_.Get(
		j,
		"id",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMap) IdInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"idInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMap) Lifecycle() *cdktf.TerraformResourceLifecycle {
	var returns *cdktf.TerraformResourceLifecycle
	_jsii_.Get(
		j,
		"lifecycle",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMap) MapId() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"mapId",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMap) Name() *string {
	var returns *string
	_jsii_.Get(
		j,
		"name",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMap) NameInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"nameInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMap) Node() constructs.Node {
	var returns constructs.Node
	_jsii_.Get(
		j,
		"node",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMap) PathMatcher() ComputeUrlMapPathMatcherList {
	var returns ComputeUrlMapPathMatcherList
	_jsii_.Get(
		j,
		"pathMatcher",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMap) PathMatcherInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"pathMatcherInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMap) Project() *string {
	var returns *string
	_jsii_.Get(
		j,
		"project",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMap) ProjectInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"projectInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMap) Provider() cdktf.TerraformProvider {
	var returns cdktf.TerraformProvider
	_jsii_.Get(
		j,
		"provider",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMap) Provisioners() *[]interface{} {
	var returns *[]interface{}
	_jsii_.Get(
		j,
		"provisioners",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMap) RawOverrides() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"rawOverrides",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMap) SelfLink() *string {
	var returns *string
	_jsii_.Get(
		j,
		"selfLink",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMap) TerraformGeneratorMetadata() *cdktf.TerraformProviderGeneratorMetadata {
	var returns *cdktf.TerraformProviderGeneratorMetadata
	_jsii_.Get(
		j,
		"terraformGeneratorMetadata",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMap) TerraformMetaArguments() *map[string]interface{} {
	var returns *map[string]interface{}
	_jsii_.Get(
		j,
		"terraformMetaArguments",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMap) TerraformResourceType() *string {
	var returns *string
	_jsii_.Get(
		j,
		"terraformResourceType",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMap) Test() ComputeUrlMapTestList {
	var returns ComputeUrlMapTestList
	_jsii_.Get(
		j,
		"test",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMap) TestInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"testInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMap) Timeouts() ComputeUrlMapTimeoutsOutputReference {
	var returns ComputeUrlMapTimeoutsOutputReference
	_jsii_.Get(
		j,
		"timeouts",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMap) TimeoutsInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"timeoutsInput",
		&returns,
	)
	return returns
}


// Create a new {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_url_map google_compute_url_map} Resource.
func NewComputeUrlMap(scope constructs.Construct, id *string, config *ComputeUrlMapConfig) ComputeUrlMap {
	_init_.Initialize()

	if err := validateNewComputeUrlMapParameters(scope, id, config); err != nil {
		panic(err)
	}
	j := jsiiProxy_ComputeUrlMap{}

	_jsii_.Create(
		"@cdktf/provider-google.computeUrlMap.ComputeUrlMap",
		[]interface{}{scope, id, config},
		&j,
	)

	return &j
}

// Create a new {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_url_map google_compute_url_map} Resource.
func NewComputeUrlMap_Override(c ComputeUrlMap, scope constructs.Construct, id *string, config *ComputeUrlMapConfig) {
	_init_.Initialize()

	_jsii_.Create(
		"@cdktf/provider-google.computeUrlMap.ComputeUrlMap",
		[]interface{}{scope, id, config},
		c,
	)
}

func (j *jsiiProxy_ComputeUrlMap)SetConnection(val interface{}) {
	if err := j.validateSetConnectionParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"connection",
		val,
	)
}

func (j *jsiiProxy_ComputeUrlMap)SetCount(val interface{}) {
	if err := j.validateSetCountParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"count",
		val,
	)
}

func (j *jsiiProxy_ComputeUrlMap)SetDefaultService(val *string) {
	if err := j.validateSetDefaultServiceParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"defaultService",
		val,
	)
}

func (j *jsiiProxy_ComputeUrlMap)SetDependsOn(val *[]*string) {
	_jsii_.Set(
		j,
		"dependsOn",
		val,
	)
}

func (j *jsiiProxy_ComputeUrlMap)SetDescription(val *string) {
	if err := j.validateSetDescriptionParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"description",
		val,
	)
}

func (j *jsiiProxy_ComputeUrlMap)SetForEach(val cdktf.ITerraformIterator) {
	_jsii_.Set(
		j,
		"forEach",
		val,
	)
}

func (j *jsiiProxy_ComputeUrlMap)SetId(val *string) {
	if err := j.validateSetIdParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"id",
		val,
	)
}

func (j *jsiiProxy_ComputeUrlMap)SetLifecycle(val *cdktf.TerraformResourceLifecycle) {
	if err := j.validateSetLifecycleParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"lifecycle",
		val,
	)
}

func (j *jsiiProxy_ComputeUrlMap)SetName(val *string) {
	if err := j.validateSetNameParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"name",
		val,
	)
}

func (j *jsiiProxy_ComputeUrlMap)SetProject(val *string) {
	if err := j.validateSetProjectParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"project",
		val,
	)
}

func (j *jsiiProxy_ComputeUrlMap)SetProvider(val cdktf.TerraformProvider) {
	_jsii_.Set(
		j,
		"provider",
		val,
	)
}

func (j *jsiiProxy_ComputeUrlMap)SetProvisioners(val *[]interface{}) {
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
func ComputeUrlMap_IsConstruct(x interface{}) *bool {
	_init_.Initialize()

	if err := validateComputeUrlMap_IsConstructParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"@cdktf/provider-google.computeUrlMap.ComputeUrlMap",
		"isConstruct",
		[]interface{}{x},
		&returns,
	)

	return returns
}

// Experimental.
func ComputeUrlMap_IsTerraformElement(x interface{}) *bool {
	_init_.Initialize()

	if err := validateComputeUrlMap_IsTerraformElementParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"@cdktf/provider-google.computeUrlMap.ComputeUrlMap",
		"isTerraformElement",
		[]interface{}{x},
		&returns,
	)

	return returns
}

// Experimental.
func ComputeUrlMap_IsTerraformResource(x interface{}) *bool {
	_init_.Initialize()

	if err := validateComputeUrlMap_IsTerraformResourceParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"@cdktf/provider-google.computeUrlMap.ComputeUrlMap",
		"isTerraformResource",
		[]interface{}{x},
		&returns,
	)

	return returns
}

func ComputeUrlMap_TfResourceType() *string {
	_init_.Initialize()
	var returns *string
	_jsii_.StaticGet(
		"@cdktf/provider-google.computeUrlMap.ComputeUrlMap",
		"tfResourceType",
		&returns,
	)
	return returns
}

func (c *jsiiProxy_ComputeUrlMap) AddOverride(path *string, value interface{}) {
	if err := c.validateAddOverrideParameters(path, value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"addOverride",
		[]interface{}{path, value},
	)
}

func (c *jsiiProxy_ComputeUrlMap) GetAnyMapAttribute(terraformAttribute *string) *map[string]interface{} {
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

func (c *jsiiProxy_ComputeUrlMap) GetBooleanAttribute(terraformAttribute *string) cdktf.IResolvable {
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

func (c *jsiiProxy_ComputeUrlMap) GetBooleanMapAttribute(terraformAttribute *string) *map[string]*bool {
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

func (c *jsiiProxy_ComputeUrlMap) GetListAttribute(terraformAttribute *string) *[]*string {
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

func (c *jsiiProxy_ComputeUrlMap) GetNumberAttribute(terraformAttribute *string) *float64 {
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

func (c *jsiiProxy_ComputeUrlMap) GetNumberListAttribute(terraformAttribute *string) *[]*float64 {
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

func (c *jsiiProxy_ComputeUrlMap) GetNumberMapAttribute(terraformAttribute *string) *map[string]*float64 {
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

func (c *jsiiProxy_ComputeUrlMap) GetStringAttribute(terraformAttribute *string) *string {
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

func (c *jsiiProxy_ComputeUrlMap) GetStringMapAttribute(terraformAttribute *string) *map[string]*string {
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

func (c *jsiiProxy_ComputeUrlMap) InterpolationForAttribute(terraformAttribute *string) cdktf.IResolvable {
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

func (c *jsiiProxy_ComputeUrlMap) OverrideLogicalId(newLogicalId *string) {
	if err := c.validateOverrideLogicalIdParameters(newLogicalId); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"overrideLogicalId",
		[]interface{}{newLogicalId},
	)
}

func (c *jsiiProxy_ComputeUrlMap) PutDefaultRouteAction(value *ComputeUrlMapDefaultRouteAction) {
	if err := c.validatePutDefaultRouteActionParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putDefaultRouteAction",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ComputeUrlMap) PutDefaultUrlRedirect(value *ComputeUrlMapDefaultUrlRedirect) {
	if err := c.validatePutDefaultUrlRedirectParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putDefaultUrlRedirect",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ComputeUrlMap) PutHeaderAction(value *ComputeUrlMapHeaderAction) {
	if err := c.validatePutHeaderActionParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putHeaderAction",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ComputeUrlMap) PutHostRule(value interface{}) {
	if err := c.validatePutHostRuleParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putHostRule",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ComputeUrlMap) PutPathMatcher(value interface{}) {
	if err := c.validatePutPathMatcherParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putPathMatcher",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ComputeUrlMap) PutTest(value interface{}) {
	if err := c.validatePutTestParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putTest",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ComputeUrlMap) PutTimeouts(value *ComputeUrlMapTimeouts) {
	if err := c.validatePutTimeoutsParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putTimeouts",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ComputeUrlMap) ResetDefaultRouteAction() {
	_jsii_.InvokeVoid(
		c,
		"resetDefaultRouteAction",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeUrlMap) ResetDefaultService() {
	_jsii_.InvokeVoid(
		c,
		"resetDefaultService",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeUrlMap) ResetDefaultUrlRedirect() {
	_jsii_.InvokeVoid(
		c,
		"resetDefaultUrlRedirect",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeUrlMap) ResetDescription() {
	_jsii_.InvokeVoid(
		c,
		"resetDescription",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeUrlMap) ResetHeaderAction() {
	_jsii_.InvokeVoid(
		c,
		"resetHeaderAction",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeUrlMap) ResetHostRule() {
	_jsii_.InvokeVoid(
		c,
		"resetHostRule",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeUrlMap) ResetId() {
	_jsii_.InvokeVoid(
		c,
		"resetId",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeUrlMap) ResetOverrideLogicalId() {
	_jsii_.InvokeVoid(
		c,
		"resetOverrideLogicalId",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeUrlMap) ResetPathMatcher() {
	_jsii_.InvokeVoid(
		c,
		"resetPathMatcher",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeUrlMap) ResetProject() {
	_jsii_.InvokeVoid(
		c,
		"resetProject",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeUrlMap) ResetTest() {
	_jsii_.InvokeVoid(
		c,
		"resetTest",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeUrlMap) ResetTimeouts() {
	_jsii_.InvokeVoid(
		c,
		"resetTimeouts",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeUrlMap) SynthesizeAttributes() *map[string]interface{} {
	var returns *map[string]interface{}

	_jsii_.Invoke(
		c,
		"synthesizeAttributes",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComputeUrlMap) ToMetadata() interface{} {
	var returns interface{}

	_jsii_.Invoke(
		c,
		"toMetadata",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComputeUrlMap) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		c,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComputeUrlMap) ToTerraform() interface{} {
	var returns interface{}

	_jsii_.Invoke(
		c,
		"toTerraform",
		nil, // no parameters
		&returns,
	)

	return returns
}

