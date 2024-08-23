package computebackendservice

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/sourcegraph/managed-services-platform-cdktf/gen/google/jsii"

	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/computebackendservice/internal"
)

type ComputeBackendServiceOutlierDetectionOutputReference interface {
	cdktf.ComplexObject
	BaseEjectionTime() ComputeBackendServiceOutlierDetectionBaseEjectionTimeOutputReference
	BaseEjectionTimeInput() *ComputeBackendServiceOutlierDetectionBaseEjectionTime
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
	ConsecutiveErrors() *float64
	SetConsecutiveErrors(val *float64)
	ConsecutiveErrorsInput() *float64
	ConsecutiveGatewayFailure() *float64
	SetConsecutiveGatewayFailure(val *float64)
	ConsecutiveGatewayFailureInput() *float64
	// The creation stack of this resolvable which will be appended to errors thrown during resolution.
	//
	// If this returns an empty array the stack will not be attached.
	// Experimental.
	CreationStack() *[]*string
	EnforcingConsecutiveErrors() *float64
	SetEnforcingConsecutiveErrors(val *float64)
	EnforcingConsecutiveErrorsInput() *float64
	EnforcingConsecutiveGatewayFailure() *float64
	SetEnforcingConsecutiveGatewayFailure(val *float64)
	EnforcingConsecutiveGatewayFailureInput() *float64
	EnforcingSuccessRate() *float64
	SetEnforcingSuccessRate(val *float64)
	EnforcingSuccessRateInput() *float64
	// Experimental.
	Fqn() *string
	InternalValue() *ComputeBackendServiceOutlierDetection
	SetInternalValue(val *ComputeBackendServiceOutlierDetection)
	Interval() ComputeBackendServiceOutlierDetectionIntervalOutputReference
	IntervalInput() *ComputeBackendServiceOutlierDetectionInterval
	MaxEjectionPercent() *float64
	SetMaxEjectionPercent(val *float64)
	MaxEjectionPercentInput() *float64
	SuccessRateMinimumHosts() *float64
	SetSuccessRateMinimumHosts(val *float64)
	SuccessRateMinimumHostsInput() *float64
	SuccessRateRequestVolume() *float64
	SetSuccessRateRequestVolume(val *float64)
	SuccessRateRequestVolumeInput() *float64
	SuccessRateStdevFactor() *float64
	SetSuccessRateStdevFactor(val *float64)
	SuccessRateStdevFactorInput() *float64
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
	PutBaseEjectionTime(value *ComputeBackendServiceOutlierDetectionBaseEjectionTime)
	PutInterval(value *ComputeBackendServiceOutlierDetectionInterval)
	ResetBaseEjectionTime()
	ResetConsecutiveErrors()
	ResetConsecutiveGatewayFailure()
	ResetEnforcingConsecutiveErrors()
	ResetEnforcingConsecutiveGatewayFailure()
	ResetEnforcingSuccessRate()
	ResetInterval()
	ResetMaxEjectionPercent()
	ResetSuccessRateMinimumHosts()
	ResetSuccessRateRequestVolume()
	ResetSuccessRateStdevFactor()
	// Produce the Token's value at resolution time.
	// Experimental.
	Resolve(_context cdktf.IResolveContext) interface{}
	// Return a string representation of this resolvable object.
	//
	// Returns a reversible string representation.
	// Experimental.
	ToString() *string
}

// The jsii proxy struct for ComputeBackendServiceOutlierDetectionOutputReference
type jsiiProxy_ComputeBackendServiceOutlierDetectionOutputReference struct {
	internal.Type__cdktfComplexObject
}

func (j *jsiiProxy_ComputeBackendServiceOutlierDetectionOutputReference) BaseEjectionTime() ComputeBackendServiceOutlierDetectionBaseEjectionTimeOutputReference {
	var returns ComputeBackendServiceOutlierDetectionBaseEjectionTimeOutputReference
	_jsii_.Get(
		j,
		"baseEjectionTime",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceOutlierDetectionOutputReference) BaseEjectionTimeInput() *ComputeBackendServiceOutlierDetectionBaseEjectionTime {
	var returns *ComputeBackendServiceOutlierDetectionBaseEjectionTime
	_jsii_.Get(
		j,
		"baseEjectionTimeInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceOutlierDetectionOutputReference) ComplexObjectIndex() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"complexObjectIndex",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceOutlierDetectionOutputReference) ComplexObjectIsFromSet() *bool {
	var returns *bool
	_jsii_.Get(
		j,
		"complexObjectIsFromSet",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceOutlierDetectionOutputReference) ConsecutiveErrors() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"consecutiveErrors",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceOutlierDetectionOutputReference) ConsecutiveErrorsInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"consecutiveErrorsInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceOutlierDetectionOutputReference) ConsecutiveGatewayFailure() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"consecutiveGatewayFailure",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceOutlierDetectionOutputReference) ConsecutiveGatewayFailureInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"consecutiveGatewayFailureInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceOutlierDetectionOutputReference) CreationStack() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"creationStack",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceOutlierDetectionOutputReference) EnforcingConsecutiveErrors() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"enforcingConsecutiveErrors",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceOutlierDetectionOutputReference) EnforcingConsecutiveErrorsInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"enforcingConsecutiveErrorsInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceOutlierDetectionOutputReference) EnforcingConsecutiveGatewayFailure() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"enforcingConsecutiveGatewayFailure",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceOutlierDetectionOutputReference) EnforcingConsecutiveGatewayFailureInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"enforcingConsecutiveGatewayFailureInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceOutlierDetectionOutputReference) EnforcingSuccessRate() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"enforcingSuccessRate",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceOutlierDetectionOutputReference) EnforcingSuccessRateInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"enforcingSuccessRateInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceOutlierDetectionOutputReference) Fqn() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fqn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceOutlierDetectionOutputReference) InternalValue() *ComputeBackendServiceOutlierDetection {
	var returns *ComputeBackendServiceOutlierDetection
	_jsii_.Get(
		j,
		"internalValue",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceOutlierDetectionOutputReference) Interval() ComputeBackendServiceOutlierDetectionIntervalOutputReference {
	var returns ComputeBackendServiceOutlierDetectionIntervalOutputReference
	_jsii_.Get(
		j,
		"interval",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceOutlierDetectionOutputReference) IntervalInput() *ComputeBackendServiceOutlierDetectionInterval {
	var returns *ComputeBackendServiceOutlierDetectionInterval
	_jsii_.Get(
		j,
		"intervalInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceOutlierDetectionOutputReference) MaxEjectionPercent() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"maxEjectionPercent",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceOutlierDetectionOutputReference) MaxEjectionPercentInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"maxEjectionPercentInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceOutlierDetectionOutputReference) SuccessRateMinimumHosts() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"successRateMinimumHosts",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceOutlierDetectionOutputReference) SuccessRateMinimumHostsInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"successRateMinimumHostsInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceOutlierDetectionOutputReference) SuccessRateRequestVolume() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"successRateRequestVolume",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceOutlierDetectionOutputReference) SuccessRateRequestVolumeInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"successRateRequestVolumeInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceOutlierDetectionOutputReference) SuccessRateStdevFactor() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"successRateStdevFactor",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceOutlierDetectionOutputReference) SuccessRateStdevFactorInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"successRateStdevFactorInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceOutlierDetectionOutputReference) TerraformAttribute() *string {
	var returns *string
	_jsii_.Get(
		j,
		"terraformAttribute",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceOutlierDetectionOutputReference) TerraformResource() cdktf.IInterpolatingParent {
	var returns cdktf.IInterpolatingParent
	_jsii_.Get(
		j,
		"terraformResource",
		&returns,
	)
	return returns
}


func NewComputeBackendServiceOutlierDetectionOutputReference(terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) ComputeBackendServiceOutlierDetectionOutputReference {
	_init_.Initialize()

	if err := validateNewComputeBackendServiceOutlierDetectionOutputReferenceParameters(terraformResource, terraformAttribute); err != nil {
		panic(err)
	}
	j := jsiiProxy_ComputeBackendServiceOutlierDetectionOutputReference{}

	_jsii_.Create(
		"@cdktf/provider-google.computeBackendService.ComputeBackendServiceOutlierDetectionOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		&j,
	)

	return &j
}

func NewComputeBackendServiceOutlierDetectionOutputReference_Override(c ComputeBackendServiceOutlierDetectionOutputReference, terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) {
	_init_.Initialize()

	_jsii_.Create(
		"@cdktf/provider-google.computeBackendService.ComputeBackendServiceOutlierDetectionOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		c,
	)
}

func (j *jsiiProxy_ComputeBackendServiceOutlierDetectionOutputReference)SetComplexObjectIndex(val interface{}) {
	if err := j.validateSetComplexObjectIndexParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIndex",
		val,
	)
}

func (j *jsiiProxy_ComputeBackendServiceOutlierDetectionOutputReference)SetComplexObjectIsFromSet(val *bool) {
	if err := j.validateSetComplexObjectIsFromSetParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIsFromSet",
		val,
	)
}

func (j *jsiiProxy_ComputeBackendServiceOutlierDetectionOutputReference)SetConsecutiveErrors(val *float64) {
	if err := j.validateSetConsecutiveErrorsParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"consecutiveErrors",
		val,
	)
}

func (j *jsiiProxy_ComputeBackendServiceOutlierDetectionOutputReference)SetConsecutiveGatewayFailure(val *float64) {
	if err := j.validateSetConsecutiveGatewayFailureParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"consecutiveGatewayFailure",
		val,
	)
}

func (j *jsiiProxy_ComputeBackendServiceOutlierDetectionOutputReference)SetEnforcingConsecutiveErrors(val *float64) {
	if err := j.validateSetEnforcingConsecutiveErrorsParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"enforcingConsecutiveErrors",
		val,
	)
}

func (j *jsiiProxy_ComputeBackendServiceOutlierDetectionOutputReference)SetEnforcingConsecutiveGatewayFailure(val *float64) {
	if err := j.validateSetEnforcingConsecutiveGatewayFailureParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"enforcingConsecutiveGatewayFailure",
		val,
	)
}

func (j *jsiiProxy_ComputeBackendServiceOutlierDetectionOutputReference)SetEnforcingSuccessRate(val *float64) {
	if err := j.validateSetEnforcingSuccessRateParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"enforcingSuccessRate",
		val,
	)
}

func (j *jsiiProxy_ComputeBackendServiceOutlierDetectionOutputReference)SetInternalValue(val *ComputeBackendServiceOutlierDetection) {
	if err := j.validateSetInternalValueParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"internalValue",
		val,
	)
}

func (j *jsiiProxy_ComputeBackendServiceOutlierDetectionOutputReference)SetMaxEjectionPercent(val *float64) {
	if err := j.validateSetMaxEjectionPercentParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"maxEjectionPercent",
		val,
	)
}

func (j *jsiiProxy_ComputeBackendServiceOutlierDetectionOutputReference)SetSuccessRateMinimumHosts(val *float64) {
	if err := j.validateSetSuccessRateMinimumHostsParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"successRateMinimumHosts",
		val,
	)
}

func (j *jsiiProxy_ComputeBackendServiceOutlierDetectionOutputReference)SetSuccessRateRequestVolume(val *float64) {
	if err := j.validateSetSuccessRateRequestVolumeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"successRateRequestVolume",
		val,
	)
}

func (j *jsiiProxy_ComputeBackendServiceOutlierDetectionOutputReference)SetSuccessRateStdevFactor(val *float64) {
	if err := j.validateSetSuccessRateStdevFactorParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"successRateStdevFactor",
		val,
	)
}

func (j *jsiiProxy_ComputeBackendServiceOutlierDetectionOutputReference)SetTerraformAttribute(val *string) {
	if err := j.validateSetTerraformAttributeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformAttribute",
		val,
	)
}

func (j *jsiiProxy_ComputeBackendServiceOutlierDetectionOutputReference)SetTerraformResource(val cdktf.IInterpolatingParent) {
	if err := j.validateSetTerraformResourceParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformResource",
		val,
	)
}

func (c *jsiiProxy_ComputeBackendServiceOutlierDetectionOutputReference) ComputeFqn() *string {
	var returns *string

	_jsii_.Invoke(
		c,
		"computeFqn",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComputeBackendServiceOutlierDetectionOutputReference) GetAnyMapAttribute(terraformAttribute *string) *map[string]interface{} {
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

func (c *jsiiProxy_ComputeBackendServiceOutlierDetectionOutputReference) GetBooleanAttribute(terraformAttribute *string) cdktf.IResolvable {
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

func (c *jsiiProxy_ComputeBackendServiceOutlierDetectionOutputReference) GetBooleanMapAttribute(terraformAttribute *string) *map[string]*bool {
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

func (c *jsiiProxy_ComputeBackendServiceOutlierDetectionOutputReference) GetListAttribute(terraformAttribute *string) *[]*string {
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

func (c *jsiiProxy_ComputeBackendServiceOutlierDetectionOutputReference) GetNumberAttribute(terraformAttribute *string) *float64 {
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

func (c *jsiiProxy_ComputeBackendServiceOutlierDetectionOutputReference) GetNumberListAttribute(terraformAttribute *string) *[]*float64 {
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

func (c *jsiiProxy_ComputeBackendServiceOutlierDetectionOutputReference) GetNumberMapAttribute(terraformAttribute *string) *map[string]*float64 {
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

func (c *jsiiProxy_ComputeBackendServiceOutlierDetectionOutputReference) GetStringAttribute(terraformAttribute *string) *string {
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

func (c *jsiiProxy_ComputeBackendServiceOutlierDetectionOutputReference) GetStringMapAttribute(terraformAttribute *string) *map[string]*string {
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

func (c *jsiiProxy_ComputeBackendServiceOutlierDetectionOutputReference) InterpolationAsList() cdktf.IResolvable {
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		c,
		"interpolationAsList",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComputeBackendServiceOutlierDetectionOutputReference) InterpolationForAttribute(property *string) cdktf.IResolvable {
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

func (c *jsiiProxy_ComputeBackendServiceOutlierDetectionOutputReference) PutBaseEjectionTime(value *ComputeBackendServiceOutlierDetectionBaseEjectionTime) {
	if err := c.validatePutBaseEjectionTimeParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putBaseEjectionTime",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ComputeBackendServiceOutlierDetectionOutputReference) PutInterval(value *ComputeBackendServiceOutlierDetectionInterval) {
	if err := c.validatePutIntervalParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putInterval",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ComputeBackendServiceOutlierDetectionOutputReference) ResetBaseEjectionTime() {
	_jsii_.InvokeVoid(
		c,
		"resetBaseEjectionTime",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeBackendServiceOutlierDetectionOutputReference) ResetConsecutiveErrors() {
	_jsii_.InvokeVoid(
		c,
		"resetConsecutiveErrors",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeBackendServiceOutlierDetectionOutputReference) ResetConsecutiveGatewayFailure() {
	_jsii_.InvokeVoid(
		c,
		"resetConsecutiveGatewayFailure",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeBackendServiceOutlierDetectionOutputReference) ResetEnforcingConsecutiveErrors() {
	_jsii_.InvokeVoid(
		c,
		"resetEnforcingConsecutiveErrors",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeBackendServiceOutlierDetectionOutputReference) ResetEnforcingConsecutiveGatewayFailure() {
	_jsii_.InvokeVoid(
		c,
		"resetEnforcingConsecutiveGatewayFailure",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeBackendServiceOutlierDetectionOutputReference) ResetEnforcingSuccessRate() {
	_jsii_.InvokeVoid(
		c,
		"resetEnforcingSuccessRate",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeBackendServiceOutlierDetectionOutputReference) ResetInterval() {
	_jsii_.InvokeVoid(
		c,
		"resetInterval",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeBackendServiceOutlierDetectionOutputReference) ResetMaxEjectionPercent() {
	_jsii_.InvokeVoid(
		c,
		"resetMaxEjectionPercent",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeBackendServiceOutlierDetectionOutputReference) ResetSuccessRateMinimumHosts() {
	_jsii_.InvokeVoid(
		c,
		"resetSuccessRateMinimumHosts",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeBackendServiceOutlierDetectionOutputReference) ResetSuccessRateRequestVolume() {
	_jsii_.InvokeVoid(
		c,
		"resetSuccessRateRequestVolume",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeBackendServiceOutlierDetectionOutputReference) ResetSuccessRateStdevFactor() {
	_jsii_.InvokeVoid(
		c,
		"resetSuccessRateStdevFactor",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeBackendServiceOutlierDetectionOutputReference) Resolve(_context cdktf.IResolveContext) interface{} {
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

func (c *jsiiProxy_ComputeBackendServiceOutlierDetectionOutputReference) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		c,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

