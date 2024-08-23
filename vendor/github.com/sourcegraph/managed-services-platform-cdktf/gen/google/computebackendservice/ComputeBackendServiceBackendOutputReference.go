package computebackendservice

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/sourcegraph/managed-services-platform-cdktf/gen/google/jsii"

	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/computebackendservice/internal"
)

type ComputeBackendServiceBackendOutputReference interface {
	cdktf.ComplexObject
	BalancingMode() *string
	SetBalancingMode(val *string)
	BalancingModeInput() *string
	CapacityScaler() *float64
	SetCapacityScaler(val *float64)
	CapacityScalerInput() *float64
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
	Description() *string
	SetDescription(val *string)
	DescriptionInput() *string
	// Experimental.
	Fqn() *string
	Group() *string
	SetGroup(val *string)
	GroupInput() *string
	InternalValue() interface{}
	SetInternalValue(val interface{})
	MaxConnections() *float64
	SetMaxConnections(val *float64)
	MaxConnectionsInput() *float64
	MaxConnectionsPerEndpoint() *float64
	SetMaxConnectionsPerEndpoint(val *float64)
	MaxConnectionsPerEndpointInput() *float64
	MaxConnectionsPerInstance() *float64
	SetMaxConnectionsPerInstance(val *float64)
	MaxConnectionsPerInstanceInput() *float64
	MaxRate() *float64
	SetMaxRate(val *float64)
	MaxRateInput() *float64
	MaxRatePerEndpoint() *float64
	SetMaxRatePerEndpoint(val *float64)
	MaxRatePerEndpointInput() *float64
	MaxRatePerInstance() *float64
	SetMaxRatePerInstance(val *float64)
	MaxRatePerInstanceInput() *float64
	MaxUtilization() *float64
	SetMaxUtilization(val *float64)
	MaxUtilizationInput() *float64
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
	ResetBalancingMode()
	ResetCapacityScaler()
	ResetDescription()
	ResetMaxConnections()
	ResetMaxConnectionsPerEndpoint()
	ResetMaxConnectionsPerInstance()
	ResetMaxRate()
	ResetMaxRatePerEndpoint()
	ResetMaxRatePerInstance()
	ResetMaxUtilization()
	// Produce the Token's value at resolution time.
	// Experimental.
	Resolve(_context cdktf.IResolveContext) interface{}
	// Return a string representation of this resolvable object.
	//
	// Returns a reversible string representation.
	// Experimental.
	ToString() *string
}

// The jsii proxy struct for ComputeBackendServiceBackendOutputReference
type jsiiProxy_ComputeBackendServiceBackendOutputReference struct {
	internal.Type__cdktfComplexObject
}

func (j *jsiiProxy_ComputeBackendServiceBackendOutputReference) BalancingMode() *string {
	var returns *string
	_jsii_.Get(
		j,
		"balancingMode",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceBackendOutputReference) BalancingModeInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"balancingModeInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceBackendOutputReference) CapacityScaler() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"capacityScaler",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceBackendOutputReference) CapacityScalerInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"capacityScalerInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceBackendOutputReference) ComplexObjectIndex() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"complexObjectIndex",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceBackendOutputReference) ComplexObjectIsFromSet() *bool {
	var returns *bool
	_jsii_.Get(
		j,
		"complexObjectIsFromSet",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceBackendOutputReference) CreationStack() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"creationStack",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceBackendOutputReference) Description() *string {
	var returns *string
	_jsii_.Get(
		j,
		"description",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceBackendOutputReference) DescriptionInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"descriptionInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceBackendOutputReference) Fqn() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fqn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceBackendOutputReference) Group() *string {
	var returns *string
	_jsii_.Get(
		j,
		"group",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceBackendOutputReference) GroupInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"groupInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceBackendOutputReference) InternalValue() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"internalValue",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceBackendOutputReference) MaxConnections() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"maxConnections",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceBackendOutputReference) MaxConnectionsInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"maxConnectionsInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceBackendOutputReference) MaxConnectionsPerEndpoint() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"maxConnectionsPerEndpoint",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceBackendOutputReference) MaxConnectionsPerEndpointInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"maxConnectionsPerEndpointInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceBackendOutputReference) MaxConnectionsPerInstance() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"maxConnectionsPerInstance",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceBackendOutputReference) MaxConnectionsPerInstanceInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"maxConnectionsPerInstanceInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceBackendOutputReference) MaxRate() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"maxRate",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceBackendOutputReference) MaxRateInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"maxRateInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceBackendOutputReference) MaxRatePerEndpoint() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"maxRatePerEndpoint",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceBackendOutputReference) MaxRatePerEndpointInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"maxRatePerEndpointInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceBackendOutputReference) MaxRatePerInstance() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"maxRatePerInstance",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceBackendOutputReference) MaxRatePerInstanceInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"maxRatePerInstanceInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceBackendOutputReference) MaxUtilization() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"maxUtilization",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceBackendOutputReference) MaxUtilizationInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"maxUtilizationInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceBackendOutputReference) TerraformAttribute() *string {
	var returns *string
	_jsii_.Get(
		j,
		"terraformAttribute",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceBackendOutputReference) TerraformResource() cdktf.IInterpolatingParent {
	var returns cdktf.IInterpolatingParent
	_jsii_.Get(
		j,
		"terraformResource",
		&returns,
	)
	return returns
}


func NewComputeBackendServiceBackendOutputReference(terraformResource cdktf.IInterpolatingParent, terraformAttribute *string, complexObjectIndex *float64, complexObjectIsFromSet *bool) ComputeBackendServiceBackendOutputReference {
	_init_.Initialize()

	if err := validateNewComputeBackendServiceBackendOutputReferenceParameters(terraformResource, terraformAttribute, complexObjectIndex, complexObjectIsFromSet); err != nil {
		panic(err)
	}
	j := jsiiProxy_ComputeBackendServiceBackendOutputReference{}

	_jsii_.Create(
		"@cdktf/provider-google.computeBackendService.ComputeBackendServiceBackendOutputReference",
		[]interface{}{terraformResource, terraformAttribute, complexObjectIndex, complexObjectIsFromSet},
		&j,
	)

	return &j
}

func NewComputeBackendServiceBackendOutputReference_Override(c ComputeBackendServiceBackendOutputReference, terraformResource cdktf.IInterpolatingParent, terraformAttribute *string, complexObjectIndex *float64, complexObjectIsFromSet *bool) {
	_init_.Initialize()

	_jsii_.Create(
		"@cdktf/provider-google.computeBackendService.ComputeBackendServiceBackendOutputReference",
		[]interface{}{terraformResource, terraformAttribute, complexObjectIndex, complexObjectIsFromSet},
		c,
	)
}

func (j *jsiiProxy_ComputeBackendServiceBackendOutputReference)SetBalancingMode(val *string) {
	if err := j.validateSetBalancingModeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"balancingMode",
		val,
	)
}

func (j *jsiiProxy_ComputeBackendServiceBackendOutputReference)SetCapacityScaler(val *float64) {
	if err := j.validateSetCapacityScalerParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"capacityScaler",
		val,
	)
}

func (j *jsiiProxy_ComputeBackendServiceBackendOutputReference)SetComplexObjectIndex(val interface{}) {
	if err := j.validateSetComplexObjectIndexParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIndex",
		val,
	)
}

func (j *jsiiProxy_ComputeBackendServiceBackendOutputReference)SetComplexObjectIsFromSet(val *bool) {
	if err := j.validateSetComplexObjectIsFromSetParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIsFromSet",
		val,
	)
}

func (j *jsiiProxy_ComputeBackendServiceBackendOutputReference)SetDescription(val *string) {
	if err := j.validateSetDescriptionParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"description",
		val,
	)
}

func (j *jsiiProxy_ComputeBackendServiceBackendOutputReference)SetGroup(val *string) {
	if err := j.validateSetGroupParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"group",
		val,
	)
}

func (j *jsiiProxy_ComputeBackendServiceBackendOutputReference)SetInternalValue(val interface{}) {
	if err := j.validateSetInternalValueParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"internalValue",
		val,
	)
}

func (j *jsiiProxy_ComputeBackendServiceBackendOutputReference)SetMaxConnections(val *float64) {
	if err := j.validateSetMaxConnectionsParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"maxConnections",
		val,
	)
}

func (j *jsiiProxy_ComputeBackendServiceBackendOutputReference)SetMaxConnectionsPerEndpoint(val *float64) {
	if err := j.validateSetMaxConnectionsPerEndpointParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"maxConnectionsPerEndpoint",
		val,
	)
}

func (j *jsiiProxy_ComputeBackendServiceBackendOutputReference)SetMaxConnectionsPerInstance(val *float64) {
	if err := j.validateSetMaxConnectionsPerInstanceParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"maxConnectionsPerInstance",
		val,
	)
}

func (j *jsiiProxy_ComputeBackendServiceBackendOutputReference)SetMaxRate(val *float64) {
	if err := j.validateSetMaxRateParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"maxRate",
		val,
	)
}

func (j *jsiiProxy_ComputeBackendServiceBackendOutputReference)SetMaxRatePerEndpoint(val *float64) {
	if err := j.validateSetMaxRatePerEndpointParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"maxRatePerEndpoint",
		val,
	)
}

func (j *jsiiProxy_ComputeBackendServiceBackendOutputReference)SetMaxRatePerInstance(val *float64) {
	if err := j.validateSetMaxRatePerInstanceParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"maxRatePerInstance",
		val,
	)
}

func (j *jsiiProxy_ComputeBackendServiceBackendOutputReference)SetMaxUtilization(val *float64) {
	if err := j.validateSetMaxUtilizationParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"maxUtilization",
		val,
	)
}

func (j *jsiiProxy_ComputeBackendServiceBackendOutputReference)SetTerraformAttribute(val *string) {
	if err := j.validateSetTerraformAttributeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformAttribute",
		val,
	)
}

func (j *jsiiProxy_ComputeBackendServiceBackendOutputReference)SetTerraformResource(val cdktf.IInterpolatingParent) {
	if err := j.validateSetTerraformResourceParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformResource",
		val,
	)
}

func (c *jsiiProxy_ComputeBackendServiceBackendOutputReference) ComputeFqn() *string {
	var returns *string

	_jsii_.Invoke(
		c,
		"computeFqn",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComputeBackendServiceBackendOutputReference) GetAnyMapAttribute(terraformAttribute *string) *map[string]interface{} {
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

func (c *jsiiProxy_ComputeBackendServiceBackendOutputReference) GetBooleanAttribute(terraformAttribute *string) cdktf.IResolvable {
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

func (c *jsiiProxy_ComputeBackendServiceBackendOutputReference) GetBooleanMapAttribute(terraformAttribute *string) *map[string]*bool {
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

func (c *jsiiProxy_ComputeBackendServiceBackendOutputReference) GetListAttribute(terraformAttribute *string) *[]*string {
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

func (c *jsiiProxy_ComputeBackendServiceBackendOutputReference) GetNumberAttribute(terraformAttribute *string) *float64 {
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

func (c *jsiiProxy_ComputeBackendServiceBackendOutputReference) GetNumberListAttribute(terraformAttribute *string) *[]*float64 {
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

func (c *jsiiProxy_ComputeBackendServiceBackendOutputReference) GetNumberMapAttribute(terraformAttribute *string) *map[string]*float64 {
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

func (c *jsiiProxy_ComputeBackendServiceBackendOutputReference) GetStringAttribute(terraformAttribute *string) *string {
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

func (c *jsiiProxy_ComputeBackendServiceBackendOutputReference) GetStringMapAttribute(terraformAttribute *string) *map[string]*string {
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

func (c *jsiiProxy_ComputeBackendServiceBackendOutputReference) InterpolationAsList() cdktf.IResolvable {
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		c,
		"interpolationAsList",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComputeBackendServiceBackendOutputReference) InterpolationForAttribute(property *string) cdktf.IResolvable {
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

func (c *jsiiProxy_ComputeBackendServiceBackendOutputReference) ResetBalancingMode() {
	_jsii_.InvokeVoid(
		c,
		"resetBalancingMode",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeBackendServiceBackendOutputReference) ResetCapacityScaler() {
	_jsii_.InvokeVoid(
		c,
		"resetCapacityScaler",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeBackendServiceBackendOutputReference) ResetDescription() {
	_jsii_.InvokeVoid(
		c,
		"resetDescription",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeBackendServiceBackendOutputReference) ResetMaxConnections() {
	_jsii_.InvokeVoid(
		c,
		"resetMaxConnections",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeBackendServiceBackendOutputReference) ResetMaxConnectionsPerEndpoint() {
	_jsii_.InvokeVoid(
		c,
		"resetMaxConnectionsPerEndpoint",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeBackendServiceBackendOutputReference) ResetMaxConnectionsPerInstance() {
	_jsii_.InvokeVoid(
		c,
		"resetMaxConnectionsPerInstance",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeBackendServiceBackendOutputReference) ResetMaxRate() {
	_jsii_.InvokeVoid(
		c,
		"resetMaxRate",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeBackendServiceBackendOutputReference) ResetMaxRatePerEndpoint() {
	_jsii_.InvokeVoid(
		c,
		"resetMaxRatePerEndpoint",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeBackendServiceBackendOutputReference) ResetMaxRatePerInstance() {
	_jsii_.InvokeVoid(
		c,
		"resetMaxRatePerInstance",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeBackendServiceBackendOutputReference) ResetMaxUtilization() {
	_jsii_.InvokeVoid(
		c,
		"resetMaxUtilization",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeBackendServiceBackendOutputReference) Resolve(_context cdktf.IResolveContext) interface{} {
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

func (c *jsiiProxy_ComputeBackendServiceBackendOutputReference) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		c,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

