package computenetwork

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/sourcegraph/managed-services-platform-cdktf/gen/google/jsii"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/computenetwork/internal"
)

// Represents a {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_network google_compute_network}.
type ComputeNetwork interface {
	cdktf.TerraformResource
	AutoCreateSubnetworks() interface{}
	SetAutoCreateSubnetworks(val interface{})
	AutoCreateSubnetworksInput() interface{}
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
	DeleteDefaultRoutesOnCreate() interface{}
	SetDeleteDefaultRoutesOnCreate(val interface{})
	DeleteDefaultRoutesOnCreateInput() interface{}
	// Experimental.
	DependsOn() *[]*string
	// Experimental.
	SetDependsOn(val *[]*string)
	Description() *string
	SetDescription(val *string)
	DescriptionInput() *string
	EnableUlaInternalIpv6() interface{}
	SetEnableUlaInternalIpv6(val interface{})
	EnableUlaInternalIpv6Input() interface{}
	// Experimental.
	ForEach() cdktf.ITerraformIterator
	// Experimental.
	SetForEach(val cdktf.ITerraformIterator)
	// Experimental.
	Fqn() *string
	// Experimental.
	FriendlyUniqueId() *string
	GatewayIpv4() *string
	Id() *string
	SetId(val *string)
	IdInput() *string
	InternalIpv6Range() *string
	SetInternalIpv6Range(val *string)
	InternalIpv6RangeInput() *string
	// Experimental.
	Lifecycle() *cdktf.TerraformResourceLifecycle
	// Experimental.
	SetLifecycle(val *cdktf.TerraformResourceLifecycle)
	Mtu() *float64
	SetMtu(val *float64)
	MtuInput() *float64
	Name() *string
	SetName(val *string)
	NameInput() *string
	NetworkFirewallPolicyEnforcementOrder() *string
	SetNetworkFirewallPolicyEnforcementOrder(val *string)
	NetworkFirewallPolicyEnforcementOrderInput() *string
	// The tree node.
	Node() constructs.Node
	NumericId() *string
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
	RoutingMode() *string
	SetRoutingMode(val *string)
	RoutingModeInput() *string
	SelfLink() *string
	// Experimental.
	TerraformGeneratorMetadata() *cdktf.TerraformProviderGeneratorMetadata
	// Experimental.
	TerraformMetaArguments() *map[string]interface{}
	// Experimental.
	TerraformResourceType() *string
	Timeouts() ComputeNetworkTimeoutsOutputReference
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
	PutTimeouts(value *ComputeNetworkTimeouts)
	ResetAutoCreateSubnetworks()
	ResetDeleteDefaultRoutesOnCreate()
	ResetDescription()
	ResetEnableUlaInternalIpv6()
	ResetId()
	ResetInternalIpv6Range()
	ResetMtu()
	ResetNetworkFirewallPolicyEnforcementOrder()
	// Resets a previously passed logical Id to use the auto-generated logical id again.
	// Experimental.
	ResetOverrideLogicalId()
	ResetProject()
	ResetRoutingMode()
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

// The jsii proxy struct for ComputeNetwork
type jsiiProxy_ComputeNetwork struct {
	internal.Type__cdktfTerraformResource
}

func (j *jsiiProxy_ComputeNetwork) AutoCreateSubnetworks() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"autoCreateSubnetworks",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeNetwork) AutoCreateSubnetworksInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"autoCreateSubnetworksInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeNetwork) CdktfStack() cdktf.TerraformStack {
	var returns cdktf.TerraformStack
	_jsii_.Get(
		j,
		"cdktfStack",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeNetwork) Connection() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"connection",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeNetwork) ConstructNodeMetadata() *map[string]interface{} {
	var returns *map[string]interface{}
	_jsii_.Get(
		j,
		"constructNodeMetadata",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeNetwork) Count() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"count",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeNetwork) DeleteDefaultRoutesOnCreate() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"deleteDefaultRoutesOnCreate",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeNetwork) DeleteDefaultRoutesOnCreateInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"deleteDefaultRoutesOnCreateInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeNetwork) DependsOn() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"dependsOn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeNetwork) Description() *string {
	var returns *string
	_jsii_.Get(
		j,
		"description",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeNetwork) DescriptionInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"descriptionInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeNetwork) EnableUlaInternalIpv6() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"enableUlaInternalIpv6",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeNetwork) EnableUlaInternalIpv6Input() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"enableUlaInternalIpv6Input",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeNetwork) ForEach() cdktf.ITerraformIterator {
	var returns cdktf.ITerraformIterator
	_jsii_.Get(
		j,
		"forEach",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeNetwork) Fqn() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fqn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeNetwork) FriendlyUniqueId() *string {
	var returns *string
	_jsii_.Get(
		j,
		"friendlyUniqueId",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeNetwork) GatewayIpv4() *string {
	var returns *string
	_jsii_.Get(
		j,
		"gatewayIpv4",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeNetwork) Id() *string {
	var returns *string
	_jsii_.Get(
		j,
		"id",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeNetwork) IdInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"idInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeNetwork) InternalIpv6Range() *string {
	var returns *string
	_jsii_.Get(
		j,
		"internalIpv6Range",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeNetwork) InternalIpv6RangeInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"internalIpv6RangeInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeNetwork) Lifecycle() *cdktf.TerraformResourceLifecycle {
	var returns *cdktf.TerraformResourceLifecycle
	_jsii_.Get(
		j,
		"lifecycle",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeNetwork) Mtu() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"mtu",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeNetwork) MtuInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"mtuInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeNetwork) Name() *string {
	var returns *string
	_jsii_.Get(
		j,
		"name",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeNetwork) NameInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"nameInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeNetwork) NetworkFirewallPolicyEnforcementOrder() *string {
	var returns *string
	_jsii_.Get(
		j,
		"networkFirewallPolicyEnforcementOrder",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeNetwork) NetworkFirewallPolicyEnforcementOrderInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"networkFirewallPolicyEnforcementOrderInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeNetwork) Node() constructs.Node {
	var returns constructs.Node
	_jsii_.Get(
		j,
		"node",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeNetwork) NumericId() *string {
	var returns *string
	_jsii_.Get(
		j,
		"numericId",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeNetwork) Project() *string {
	var returns *string
	_jsii_.Get(
		j,
		"project",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeNetwork) ProjectInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"projectInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeNetwork) Provider() cdktf.TerraformProvider {
	var returns cdktf.TerraformProvider
	_jsii_.Get(
		j,
		"provider",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeNetwork) Provisioners() *[]interface{} {
	var returns *[]interface{}
	_jsii_.Get(
		j,
		"provisioners",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeNetwork) RawOverrides() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"rawOverrides",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeNetwork) RoutingMode() *string {
	var returns *string
	_jsii_.Get(
		j,
		"routingMode",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeNetwork) RoutingModeInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"routingModeInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeNetwork) SelfLink() *string {
	var returns *string
	_jsii_.Get(
		j,
		"selfLink",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeNetwork) TerraformGeneratorMetadata() *cdktf.TerraformProviderGeneratorMetadata {
	var returns *cdktf.TerraformProviderGeneratorMetadata
	_jsii_.Get(
		j,
		"terraformGeneratorMetadata",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeNetwork) TerraformMetaArguments() *map[string]interface{} {
	var returns *map[string]interface{}
	_jsii_.Get(
		j,
		"terraformMetaArguments",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeNetwork) TerraformResourceType() *string {
	var returns *string
	_jsii_.Get(
		j,
		"terraformResourceType",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeNetwork) Timeouts() ComputeNetworkTimeoutsOutputReference {
	var returns ComputeNetworkTimeoutsOutputReference
	_jsii_.Get(
		j,
		"timeouts",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeNetwork) TimeoutsInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"timeoutsInput",
		&returns,
	)
	return returns
}


// Create a new {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_network google_compute_network} Resource.
func NewComputeNetwork(scope constructs.Construct, id *string, config *ComputeNetworkConfig) ComputeNetwork {
	_init_.Initialize()

	if err := validateNewComputeNetworkParameters(scope, id, config); err != nil {
		panic(err)
	}
	j := jsiiProxy_ComputeNetwork{}

	_jsii_.Create(
		"@cdktf/provider-google.computeNetwork.ComputeNetwork",
		[]interface{}{scope, id, config},
		&j,
	)

	return &j
}

// Create a new {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_network google_compute_network} Resource.
func NewComputeNetwork_Override(c ComputeNetwork, scope constructs.Construct, id *string, config *ComputeNetworkConfig) {
	_init_.Initialize()

	_jsii_.Create(
		"@cdktf/provider-google.computeNetwork.ComputeNetwork",
		[]interface{}{scope, id, config},
		c,
	)
}

func (j *jsiiProxy_ComputeNetwork)SetAutoCreateSubnetworks(val interface{}) {
	if err := j.validateSetAutoCreateSubnetworksParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"autoCreateSubnetworks",
		val,
	)
}

func (j *jsiiProxy_ComputeNetwork)SetConnection(val interface{}) {
	if err := j.validateSetConnectionParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"connection",
		val,
	)
}

func (j *jsiiProxy_ComputeNetwork)SetCount(val interface{}) {
	if err := j.validateSetCountParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"count",
		val,
	)
}

func (j *jsiiProxy_ComputeNetwork)SetDeleteDefaultRoutesOnCreate(val interface{}) {
	if err := j.validateSetDeleteDefaultRoutesOnCreateParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"deleteDefaultRoutesOnCreate",
		val,
	)
}

func (j *jsiiProxy_ComputeNetwork)SetDependsOn(val *[]*string) {
	_jsii_.Set(
		j,
		"dependsOn",
		val,
	)
}

func (j *jsiiProxy_ComputeNetwork)SetDescription(val *string) {
	if err := j.validateSetDescriptionParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"description",
		val,
	)
}

func (j *jsiiProxy_ComputeNetwork)SetEnableUlaInternalIpv6(val interface{}) {
	if err := j.validateSetEnableUlaInternalIpv6Parameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"enableUlaInternalIpv6",
		val,
	)
}

func (j *jsiiProxy_ComputeNetwork)SetForEach(val cdktf.ITerraformIterator) {
	_jsii_.Set(
		j,
		"forEach",
		val,
	)
}

func (j *jsiiProxy_ComputeNetwork)SetId(val *string) {
	if err := j.validateSetIdParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"id",
		val,
	)
}

func (j *jsiiProxy_ComputeNetwork)SetInternalIpv6Range(val *string) {
	if err := j.validateSetInternalIpv6RangeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"internalIpv6Range",
		val,
	)
}

func (j *jsiiProxy_ComputeNetwork)SetLifecycle(val *cdktf.TerraformResourceLifecycle) {
	if err := j.validateSetLifecycleParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"lifecycle",
		val,
	)
}

func (j *jsiiProxy_ComputeNetwork)SetMtu(val *float64) {
	if err := j.validateSetMtuParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"mtu",
		val,
	)
}

func (j *jsiiProxy_ComputeNetwork)SetName(val *string) {
	if err := j.validateSetNameParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"name",
		val,
	)
}

func (j *jsiiProxy_ComputeNetwork)SetNetworkFirewallPolicyEnforcementOrder(val *string) {
	if err := j.validateSetNetworkFirewallPolicyEnforcementOrderParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"networkFirewallPolicyEnforcementOrder",
		val,
	)
}

func (j *jsiiProxy_ComputeNetwork)SetProject(val *string) {
	if err := j.validateSetProjectParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"project",
		val,
	)
}

func (j *jsiiProxy_ComputeNetwork)SetProvider(val cdktf.TerraformProvider) {
	_jsii_.Set(
		j,
		"provider",
		val,
	)
}

func (j *jsiiProxy_ComputeNetwork)SetProvisioners(val *[]interface{}) {
	if err := j.validateSetProvisionersParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"provisioners",
		val,
	)
}

func (j *jsiiProxy_ComputeNetwork)SetRoutingMode(val *string) {
	if err := j.validateSetRoutingModeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"routingMode",
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
func ComputeNetwork_IsConstruct(x interface{}) *bool {
	_init_.Initialize()

	if err := validateComputeNetwork_IsConstructParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"@cdktf/provider-google.computeNetwork.ComputeNetwork",
		"isConstruct",
		[]interface{}{x},
		&returns,
	)

	return returns
}

// Experimental.
func ComputeNetwork_IsTerraformElement(x interface{}) *bool {
	_init_.Initialize()

	if err := validateComputeNetwork_IsTerraformElementParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"@cdktf/provider-google.computeNetwork.ComputeNetwork",
		"isTerraformElement",
		[]interface{}{x},
		&returns,
	)

	return returns
}

// Experimental.
func ComputeNetwork_IsTerraformResource(x interface{}) *bool {
	_init_.Initialize()

	if err := validateComputeNetwork_IsTerraformResourceParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"@cdktf/provider-google.computeNetwork.ComputeNetwork",
		"isTerraformResource",
		[]interface{}{x},
		&returns,
	)

	return returns
}

func ComputeNetwork_TfResourceType() *string {
	_init_.Initialize()
	var returns *string
	_jsii_.StaticGet(
		"@cdktf/provider-google.computeNetwork.ComputeNetwork",
		"tfResourceType",
		&returns,
	)
	return returns
}

func (c *jsiiProxy_ComputeNetwork) AddOverride(path *string, value interface{}) {
	if err := c.validateAddOverrideParameters(path, value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"addOverride",
		[]interface{}{path, value},
	)
}

func (c *jsiiProxy_ComputeNetwork) GetAnyMapAttribute(terraformAttribute *string) *map[string]interface{} {
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

func (c *jsiiProxy_ComputeNetwork) GetBooleanAttribute(terraformAttribute *string) cdktf.IResolvable {
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

func (c *jsiiProxy_ComputeNetwork) GetBooleanMapAttribute(terraformAttribute *string) *map[string]*bool {
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

func (c *jsiiProxy_ComputeNetwork) GetListAttribute(terraformAttribute *string) *[]*string {
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

func (c *jsiiProxy_ComputeNetwork) GetNumberAttribute(terraformAttribute *string) *float64 {
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

func (c *jsiiProxy_ComputeNetwork) GetNumberListAttribute(terraformAttribute *string) *[]*float64 {
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

func (c *jsiiProxy_ComputeNetwork) GetNumberMapAttribute(terraformAttribute *string) *map[string]*float64 {
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

func (c *jsiiProxy_ComputeNetwork) GetStringAttribute(terraformAttribute *string) *string {
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

func (c *jsiiProxy_ComputeNetwork) GetStringMapAttribute(terraformAttribute *string) *map[string]*string {
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

func (c *jsiiProxy_ComputeNetwork) InterpolationForAttribute(terraformAttribute *string) cdktf.IResolvable {
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

func (c *jsiiProxy_ComputeNetwork) OverrideLogicalId(newLogicalId *string) {
	if err := c.validateOverrideLogicalIdParameters(newLogicalId); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"overrideLogicalId",
		[]interface{}{newLogicalId},
	)
}

func (c *jsiiProxy_ComputeNetwork) PutTimeouts(value *ComputeNetworkTimeouts) {
	if err := c.validatePutTimeoutsParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putTimeouts",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ComputeNetwork) ResetAutoCreateSubnetworks() {
	_jsii_.InvokeVoid(
		c,
		"resetAutoCreateSubnetworks",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeNetwork) ResetDeleteDefaultRoutesOnCreate() {
	_jsii_.InvokeVoid(
		c,
		"resetDeleteDefaultRoutesOnCreate",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeNetwork) ResetDescription() {
	_jsii_.InvokeVoid(
		c,
		"resetDescription",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeNetwork) ResetEnableUlaInternalIpv6() {
	_jsii_.InvokeVoid(
		c,
		"resetEnableUlaInternalIpv6",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeNetwork) ResetId() {
	_jsii_.InvokeVoid(
		c,
		"resetId",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeNetwork) ResetInternalIpv6Range() {
	_jsii_.InvokeVoid(
		c,
		"resetInternalIpv6Range",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeNetwork) ResetMtu() {
	_jsii_.InvokeVoid(
		c,
		"resetMtu",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeNetwork) ResetNetworkFirewallPolicyEnforcementOrder() {
	_jsii_.InvokeVoid(
		c,
		"resetNetworkFirewallPolicyEnforcementOrder",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeNetwork) ResetOverrideLogicalId() {
	_jsii_.InvokeVoid(
		c,
		"resetOverrideLogicalId",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeNetwork) ResetProject() {
	_jsii_.InvokeVoid(
		c,
		"resetProject",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeNetwork) ResetRoutingMode() {
	_jsii_.InvokeVoid(
		c,
		"resetRoutingMode",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeNetwork) ResetTimeouts() {
	_jsii_.InvokeVoid(
		c,
		"resetTimeouts",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeNetwork) SynthesizeAttributes() *map[string]interface{} {
	var returns *map[string]interface{}

	_jsii_.Invoke(
		c,
		"synthesizeAttributes",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComputeNetwork) ToMetadata() interface{} {
	var returns interface{}

	_jsii_.Invoke(
		c,
		"toMetadata",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComputeNetwork) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		c,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComputeNetwork) ToTerraform() interface{} {
	var returns interface{}

	_jsii_.Invoke(
		c,
		"toTerraform",
		nil, // no parameters
		&returns,
	)

	return returns
}

