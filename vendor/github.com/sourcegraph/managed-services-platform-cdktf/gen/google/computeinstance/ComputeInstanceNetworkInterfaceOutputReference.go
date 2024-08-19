package computeinstance

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/sourcegraph/managed-services-platform-cdktf/gen/google/jsii"

	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/computeinstance/internal"
)

type ComputeInstanceNetworkInterfaceOutputReference interface {
	cdktf.ComplexObject
	AccessConfig() ComputeInstanceNetworkInterfaceAccessConfigList
	AccessConfigInput() interface{}
	AliasIpRange() ComputeInstanceNetworkInterfaceAliasIpRangeList
	AliasIpRangeInput() interface{}
	// the index of the complex object in a list.
	// Experimental.
	ComplexObjectIndex() interface{}
	// Experimental.
	SetComplexObjectIndex(val interface{})
	// set to true if this item is from inside a set and needs tolist() for accessing it set to "0" for single list items.
	// Experimental.
	ComplexObjectIsFromSet() *bool
	// Experimental.
	SetComplexObjectIsFromSet(val *bool)
	// The creation stack of this resolvable which will be appended to errors thrown during resolution.
	//
	// If this returns an empty array the stack will not be attached.
	// Experimental.
	CreationStack() *[]*string
	// Experimental.
	Fqn() *string
	InternalIpv6PrefixLength() *float64
	SetInternalIpv6PrefixLength(val *float64)
	InternalIpv6PrefixLengthInput() *float64
	InternalValue() interface{}
	SetInternalValue(val interface{})
	Ipv6AccessConfig() ComputeInstanceNetworkInterfaceIpv6AccessConfigList
	Ipv6AccessConfigInput() interface{}
	Ipv6AccessType() *string
	Ipv6Address() *string
	SetIpv6Address(val *string)
	Ipv6AddressInput() *string
	Name() *string
	Network() *string
	SetNetwork(val *string)
	NetworkInput() *string
	NetworkIp() *string
	SetNetworkIp(val *string)
	NetworkIpInput() *string
	NicType() *string
	SetNicType(val *string)
	NicTypeInput() *string
	QueueCount() *float64
	SetQueueCount(val *float64)
	QueueCountInput() *float64
	StackType() *string
	SetStackType(val *string)
	StackTypeInput() *string
	Subnetwork() *string
	SetSubnetwork(val *string)
	SubnetworkInput() *string
	SubnetworkProject() *string
	SetSubnetworkProject(val *string)
	SubnetworkProjectInput() *string
	// Experimental.
	TerraformAttribute() *string
	// Experimental.
	SetTerraformAttribute(val *string)
	// Experimental.
	TerraformResource() cdktf.IInterpolatingParent
	// Experimental.
	SetTerraformResource(val cdktf.IInterpolatingParent)
	// Experimental.
	ComputeFqn() *string
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
	InterpolationAsList() cdktf.IResolvable
	// Experimental.
	InterpolationForAttribute(property *string) cdktf.IResolvable
	PutAccessConfig(value interface{})
	PutAliasIpRange(value interface{})
	PutIpv6AccessConfig(value interface{})
	ResetAccessConfig()
	ResetAliasIpRange()
	ResetInternalIpv6PrefixLength()
	ResetIpv6AccessConfig()
	ResetIpv6Address()
	ResetNetwork()
	ResetNetworkIp()
	ResetNicType()
	ResetQueueCount()
	ResetStackType()
	ResetSubnetwork()
	ResetSubnetworkProject()
	// Produce the Token's value at resolution time.
	// Experimental.
	Resolve(_context cdktf.IResolveContext) interface{}
	// Return a string representation of this resolvable object.
	//
	// Returns a reversible string representation.
	// Experimental.
	ToString() *string
}

// The jsii proxy struct for ComputeInstanceNetworkInterfaceOutputReference
type jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference struct {
	internal.Type__cdktfComplexObject
}

func (j *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference) AccessConfig() ComputeInstanceNetworkInterfaceAccessConfigList {
	var returns ComputeInstanceNetworkInterfaceAccessConfigList
	_jsii_.Get(
		j,
		"accessConfig",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference) AccessConfigInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"accessConfigInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference) AliasIpRange() ComputeInstanceNetworkInterfaceAliasIpRangeList {
	var returns ComputeInstanceNetworkInterfaceAliasIpRangeList
	_jsii_.Get(
		j,
		"aliasIpRange",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference) AliasIpRangeInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"aliasIpRangeInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference) ComplexObjectIndex() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"complexObjectIndex",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference) ComplexObjectIsFromSet() *bool {
	var returns *bool
	_jsii_.Get(
		j,
		"complexObjectIsFromSet",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference) CreationStack() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"creationStack",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference) Fqn() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fqn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference) InternalIpv6PrefixLength() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"internalIpv6PrefixLength",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference) InternalIpv6PrefixLengthInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"internalIpv6PrefixLengthInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference) InternalValue() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"internalValue",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference) Ipv6AccessConfig() ComputeInstanceNetworkInterfaceIpv6AccessConfigList {
	var returns ComputeInstanceNetworkInterfaceIpv6AccessConfigList
	_jsii_.Get(
		j,
		"ipv6AccessConfig",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference) Ipv6AccessConfigInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"ipv6AccessConfigInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference) Ipv6AccessType() *string {
	var returns *string
	_jsii_.Get(
		j,
		"ipv6AccessType",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference) Ipv6Address() *string {
	var returns *string
	_jsii_.Get(
		j,
		"ipv6Address",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference) Ipv6AddressInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"ipv6AddressInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference) Name() *string {
	var returns *string
	_jsii_.Get(
		j,
		"name",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference) Network() *string {
	var returns *string
	_jsii_.Get(
		j,
		"network",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference) NetworkInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"networkInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference) NetworkIp() *string {
	var returns *string
	_jsii_.Get(
		j,
		"networkIp",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference) NetworkIpInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"networkIpInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference) NicType() *string {
	var returns *string
	_jsii_.Get(
		j,
		"nicType",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference) NicTypeInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"nicTypeInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference) QueueCount() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"queueCount",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference) QueueCountInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"queueCountInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference) StackType() *string {
	var returns *string
	_jsii_.Get(
		j,
		"stackType",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference) StackTypeInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"stackTypeInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference) Subnetwork() *string {
	var returns *string
	_jsii_.Get(
		j,
		"subnetwork",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference) SubnetworkInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"subnetworkInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference) SubnetworkProject() *string {
	var returns *string
	_jsii_.Get(
		j,
		"subnetworkProject",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference) SubnetworkProjectInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"subnetworkProjectInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference) TerraformAttribute() *string {
	var returns *string
	_jsii_.Get(
		j,
		"terraformAttribute",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference) TerraformResource() cdktf.IInterpolatingParent {
	var returns cdktf.IInterpolatingParent
	_jsii_.Get(
		j,
		"terraformResource",
		&returns,
	)
	return returns
}


func NewComputeInstanceNetworkInterfaceOutputReference(terraformResource cdktf.IInterpolatingParent, terraformAttribute *string, complexObjectIndex *float64, complexObjectIsFromSet *bool) ComputeInstanceNetworkInterfaceOutputReference {
	_init_.Initialize()

	if err := validateNewComputeInstanceNetworkInterfaceOutputReferenceParameters(terraformResource, terraformAttribute, complexObjectIndex, complexObjectIsFromSet); err != nil {
		panic(err)
	}
	j := jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference{}

	_jsii_.Create(
		"@cdktf/provider-google.computeInstance.ComputeInstanceNetworkInterfaceOutputReference",
		[]interface{}{terraformResource, terraformAttribute, complexObjectIndex, complexObjectIsFromSet},
		&j,
	)

	return &j
}

func NewComputeInstanceNetworkInterfaceOutputReference_Override(c ComputeInstanceNetworkInterfaceOutputReference, terraformResource cdktf.IInterpolatingParent, terraformAttribute *string, complexObjectIndex *float64, complexObjectIsFromSet *bool) {
	_init_.Initialize()

	_jsii_.Create(
		"@cdktf/provider-google.computeInstance.ComputeInstanceNetworkInterfaceOutputReference",
		[]interface{}{terraformResource, terraformAttribute, complexObjectIndex, complexObjectIsFromSet},
		c,
	)
}

func (j *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference)SetComplexObjectIndex(val interface{}) {
	if err := j.validateSetComplexObjectIndexParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIndex",
		val,
	)
}

func (j *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference)SetComplexObjectIsFromSet(val *bool) {
	if err := j.validateSetComplexObjectIsFromSetParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIsFromSet",
		val,
	)
}

func (j *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference)SetInternalIpv6PrefixLength(val *float64) {
	if err := j.validateSetInternalIpv6PrefixLengthParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"internalIpv6PrefixLength",
		val,
	)
}

func (j *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference)SetInternalValue(val interface{}) {
	if err := j.validateSetInternalValueParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"internalValue",
		val,
	)
}

func (j *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference)SetIpv6Address(val *string) {
	if err := j.validateSetIpv6AddressParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"ipv6Address",
		val,
	)
}

func (j *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference)SetNetwork(val *string) {
	if err := j.validateSetNetworkParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"network",
		val,
	)
}

func (j *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference)SetNetworkIp(val *string) {
	if err := j.validateSetNetworkIpParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"networkIp",
		val,
	)
}

func (j *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference)SetNicType(val *string) {
	if err := j.validateSetNicTypeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"nicType",
		val,
	)
}

func (j *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference)SetQueueCount(val *float64) {
	if err := j.validateSetQueueCountParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"queueCount",
		val,
	)
}

func (j *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference)SetStackType(val *string) {
	if err := j.validateSetStackTypeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"stackType",
		val,
	)
}

func (j *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference)SetSubnetwork(val *string) {
	if err := j.validateSetSubnetworkParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"subnetwork",
		val,
	)
}

func (j *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference)SetSubnetworkProject(val *string) {
	if err := j.validateSetSubnetworkProjectParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"subnetworkProject",
		val,
	)
}

func (j *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference)SetTerraformAttribute(val *string) {
	if err := j.validateSetTerraformAttributeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformAttribute",
		val,
	)
}

func (j *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference)SetTerraformResource(val cdktf.IInterpolatingParent) {
	if err := j.validateSetTerraformResourceParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformResource",
		val,
	)
}

func (c *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference) ComputeFqn() *string {
	var returns *string

	_jsii_.Invoke(
		c,
		"computeFqn",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference) GetAnyMapAttribute(terraformAttribute *string) *map[string]interface{} {
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

func (c *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference) GetBooleanAttribute(terraformAttribute *string) cdktf.IResolvable {
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

func (c *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference) GetBooleanMapAttribute(terraformAttribute *string) *map[string]*bool {
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

func (c *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference) GetListAttribute(terraformAttribute *string) *[]*string {
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

func (c *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference) GetNumberAttribute(terraformAttribute *string) *float64 {
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

func (c *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference) GetNumberListAttribute(terraformAttribute *string) *[]*float64 {
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

func (c *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference) GetNumberMapAttribute(terraformAttribute *string) *map[string]*float64 {
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

func (c *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference) GetStringAttribute(terraformAttribute *string) *string {
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

func (c *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference) GetStringMapAttribute(terraformAttribute *string) *map[string]*string {
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

func (c *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference) InterpolationAsList() cdktf.IResolvable {
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		c,
		"interpolationAsList",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference) InterpolationForAttribute(property *string) cdktf.IResolvable {
	if err := c.validateInterpolationForAttributeParameters(property); err != nil {
		panic(err)
	}
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		c,
		"interpolationForAttribute",
		[]interface{}{property},
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference) PutAccessConfig(value interface{}) {
	if err := c.validatePutAccessConfigParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putAccessConfig",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference) PutAliasIpRange(value interface{}) {
	if err := c.validatePutAliasIpRangeParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putAliasIpRange",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference) PutIpv6AccessConfig(value interface{}) {
	if err := c.validatePutIpv6AccessConfigParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putIpv6AccessConfig",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference) ResetAccessConfig() {
	_jsii_.InvokeVoid(
		c,
		"resetAccessConfig",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference) ResetAliasIpRange() {
	_jsii_.InvokeVoid(
		c,
		"resetAliasIpRange",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference) ResetInternalIpv6PrefixLength() {
	_jsii_.InvokeVoid(
		c,
		"resetInternalIpv6PrefixLength",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference) ResetIpv6AccessConfig() {
	_jsii_.InvokeVoid(
		c,
		"resetIpv6AccessConfig",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference) ResetIpv6Address() {
	_jsii_.InvokeVoid(
		c,
		"resetIpv6Address",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference) ResetNetwork() {
	_jsii_.InvokeVoid(
		c,
		"resetNetwork",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference) ResetNetworkIp() {
	_jsii_.InvokeVoid(
		c,
		"resetNetworkIp",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference) ResetNicType() {
	_jsii_.InvokeVoid(
		c,
		"resetNicType",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference) ResetQueueCount() {
	_jsii_.InvokeVoid(
		c,
		"resetQueueCount",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference) ResetStackType() {
	_jsii_.InvokeVoid(
		c,
		"resetStackType",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference) ResetSubnetwork() {
	_jsii_.InvokeVoid(
		c,
		"resetSubnetwork",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference) ResetSubnetworkProject() {
	_jsii_.InvokeVoid(
		c,
		"resetSubnetworkProject",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference) Resolve(_context cdktf.IResolveContext) interface{} {
	if err := c.validateResolveParameters(_context); err != nil {
		panic(err)
	}
	var returns interface{}

	_jsii_.Invoke(
		c,
		"resolve",
		[]interface{}{_context},
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComputeInstanceNetworkInterfaceOutputReference) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		c,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

