package computesecuritypolicy

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/sourcegraph/managed-services-platform-cdktf/gen/google/jsii"

	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/computesecuritypolicy/internal"
)

type ComputeSecurityPolicyRuleRateLimitOptionsOutputReference interface {
	cdktf.ComplexObject
	BanDurationSec() *float64
	SetBanDurationSec(val *float64)
	BanDurationSecInput() *float64
	BanThreshold() ComputeSecurityPolicyRuleRateLimitOptionsBanThresholdOutputReference
	BanThresholdInput() *ComputeSecurityPolicyRuleRateLimitOptionsBanThreshold
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
	ConformAction() *string
	SetConformAction(val *string)
	ConformActionInput() *string
	// The creation stack of this resolvable which will be appended to errors thrown during resolution.
	//
	// If this returns an empty array the stack will not be attached.
	// Experimental.
	CreationStack() *[]*string
	EnforceOnKey() *string
	SetEnforceOnKey(val *string)
	EnforceOnKeyInput() *string
	EnforceOnKeyName() *string
	SetEnforceOnKeyName(val *string)
	EnforceOnKeyNameInput() *string
	ExceedAction() *string
	SetExceedAction(val *string)
	ExceedActionInput() *string
	ExceedRedirectOptions() ComputeSecurityPolicyRuleRateLimitOptionsExceedRedirectOptionsOutputReference
	ExceedRedirectOptionsInput() *ComputeSecurityPolicyRuleRateLimitOptionsExceedRedirectOptions
	// Experimental.
	Fqn() *string
	InternalValue() *ComputeSecurityPolicyRuleRateLimitOptions
	SetInternalValue(val *ComputeSecurityPolicyRuleRateLimitOptions)
	RateLimitThreshold() ComputeSecurityPolicyRuleRateLimitOptionsRateLimitThresholdOutputReference
	RateLimitThresholdInput() *ComputeSecurityPolicyRuleRateLimitOptionsRateLimitThreshold
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
	PutBanThreshold(value *ComputeSecurityPolicyRuleRateLimitOptionsBanThreshold)
	PutExceedRedirectOptions(value *ComputeSecurityPolicyRuleRateLimitOptionsExceedRedirectOptions)
	PutRateLimitThreshold(value *ComputeSecurityPolicyRuleRateLimitOptionsRateLimitThreshold)
	ResetBanDurationSec()
	ResetBanThreshold()
	ResetEnforceOnKey()
	ResetEnforceOnKeyName()
	ResetExceedRedirectOptions()
	// Produce the Token's value at resolution time.
	// Experimental.
	Resolve(_context cdktf.IResolveContext) interface{}
	// Return a string representation of this resolvable object.
	//
	// Returns a reversible string representation.
	// Experimental.
	ToString() *string
}

// The jsii proxy struct for ComputeSecurityPolicyRuleRateLimitOptionsOutputReference
type jsiiProxy_ComputeSecurityPolicyRuleRateLimitOptionsOutputReference struct {
	internal.Type__cdktfComplexObject
}

func (j *jsiiProxy_ComputeSecurityPolicyRuleRateLimitOptionsOutputReference) BanDurationSec() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"banDurationSec",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSecurityPolicyRuleRateLimitOptionsOutputReference) BanDurationSecInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"banDurationSecInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSecurityPolicyRuleRateLimitOptionsOutputReference) BanThreshold() ComputeSecurityPolicyRuleRateLimitOptionsBanThresholdOutputReference {
	var returns ComputeSecurityPolicyRuleRateLimitOptionsBanThresholdOutputReference
	_jsii_.Get(
		j,
		"banThreshold",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSecurityPolicyRuleRateLimitOptionsOutputReference) BanThresholdInput() *ComputeSecurityPolicyRuleRateLimitOptionsBanThreshold {
	var returns *ComputeSecurityPolicyRuleRateLimitOptionsBanThreshold
	_jsii_.Get(
		j,
		"banThresholdInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSecurityPolicyRuleRateLimitOptionsOutputReference) ComplexObjectIndex() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"complexObjectIndex",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSecurityPolicyRuleRateLimitOptionsOutputReference) ComplexObjectIsFromSet() *bool {
	var returns *bool
	_jsii_.Get(
		j,
		"complexObjectIsFromSet",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSecurityPolicyRuleRateLimitOptionsOutputReference) ConformAction() *string {
	var returns *string
	_jsii_.Get(
		j,
		"conformAction",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSecurityPolicyRuleRateLimitOptionsOutputReference) ConformActionInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"conformActionInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSecurityPolicyRuleRateLimitOptionsOutputReference) CreationStack() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"creationStack",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSecurityPolicyRuleRateLimitOptionsOutputReference) EnforceOnKey() *string {
	var returns *string
	_jsii_.Get(
		j,
		"enforceOnKey",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSecurityPolicyRuleRateLimitOptionsOutputReference) EnforceOnKeyInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"enforceOnKeyInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSecurityPolicyRuleRateLimitOptionsOutputReference) EnforceOnKeyName() *string {
	var returns *string
	_jsii_.Get(
		j,
		"enforceOnKeyName",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSecurityPolicyRuleRateLimitOptionsOutputReference) EnforceOnKeyNameInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"enforceOnKeyNameInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSecurityPolicyRuleRateLimitOptionsOutputReference) ExceedAction() *string {
	var returns *string
	_jsii_.Get(
		j,
		"exceedAction",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSecurityPolicyRuleRateLimitOptionsOutputReference) ExceedActionInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"exceedActionInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSecurityPolicyRuleRateLimitOptionsOutputReference) ExceedRedirectOptions() ComputeSecurityPolicyRuleRateLimitOptionsExceedRedirectOptionsOutputReference {
	var returns ComputeSecurityPolicyRuleRateLimitOptionsExceedRedirectOptionsOutputReference
	_jsii_.Get(
		j,
		"exceedRedirectOptions",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSecurityPolicyRuleRateLimitOptionsOutputReference) ExceedRedirectOptionsInput() *ComputeSecurityPolicyRuleRateLimitOptionsExceedRedirectOptions {
	var returns *ComputeSecurityPolicyRuleRateLimitOptionsExceedRedirectOptions
	_jsii_.Get(
		j,
		"exceedRedirectOptionsInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSecurityPolicyRuleRateLimitOptionsOutputReference) Fqn() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fqn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSecurityPolicyRuleRateLimitOptionsOutputReference) InternalValue() *ComputeSecurityPolicyRuleRateLimitOptions {
	var returns *ComputeSecurityPolicyRuleRateLimitOptions
	_jsii_.Get(
		j,
		"internalValue",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSecurityPolicyRuleRateLimitOptionsOutputReference) RateLimitThreshold() ComputeSecurityPolicyRuleRateLimitOptionsRateLimitThresholdOutputReference {
	var returns ComputeSecurityPolicyRuleRateLimitOptionsRateLimitThresholdOutputReference
	_jsii_.Get(
		j,
		"rateLimitThreshold",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSecurityPolicyRuleRateLimitOptionsOutputReference) RateLimitThresholdInput() *ComputeSecurityPolicyRuleRateLimitOptionsRateLimitThreshold {
	var returns *ComputeSecurityPolicyRuleRateLimitOptionsRateLimitThreshold
	_jsii_.Get(
		j,
		"rateLimitThresholdInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSecurityPolicyRuleRateLimitOptionsOutputReference) TerraformAttribute() *string {
	var returns *string
	_jsii_.Get(
		j,
		"terraformAttribute",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSecurityPolicyRuleRateLimitOptionsOutputReference) TerraformResource() cdktf.IInterpolatingParent {
	var returns cdktf.IInterpolatingParent
	_jsii_.Get(
		j,
		"terraformResource",
		&returns,
	)
	return returns
}


func NewComputeSecurityPolicyRuleRateLimitOptionsOutputReference(terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) ComputeSecurityPolicyRuleRateLimitOptionsOutputReference {
	_init_.Initialize()

	if err := validateNewComputeSecurityPolicyRuleRateLimitOptionsOutputReferenceParameters(terraformResource, terraformAttribute); err != nil {
		panic(err)
	}
	j := jsiiProxy_ComputeSecurityPolicyRuleRateLimitOptionsOutputReference{}

	_jsii_.Create(
		"@cdktf/provider-google.computeSecurityPolicy.ComputeSecurityPolicyRuleRateLimitOptionsOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		&j,
	)

	return &j
}

func NewComputeSecurityPolicyRuleRateLimitOptionsOutputReference_Override(c ComputeSecurityPolicyRuleRateLimitOptionsOutputReference, terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) {
	_init_.Initialize()

	_jsii_.Create(
		"@cdktf/provider-google.computeSecurityPolicy.ComputeSecurityPolicyRuleRateLimitOptionsOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		c,
	)
}

func (j *jsiiProxy_ComputeSecurityPolicyRuleRateLimitOptionsOutputReference)SetBanDurationSec(val *float64) {
	if err := j.validateSetBanDurationSecParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"banDurationSec",
		val,
	)
}

func (j *jsiiProxy_ComputeSecurityPolicyRuleRateLimitOptionsOutputReference)SetComplexObjectIndex(val interface{}) {
	if err := j.validateSetComplexObjectIndexParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIndex",
		val,
	)
}

func (j *jsiiProxy_ComputeSecurityPolicyRuleRateLimitOptionsOutputReference)SetComplexObjectIsFromSet(val *bool) {
	if err := j.validateSetComplexObjectIsFromSetParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIsFromSet",
		val,
	)
}

func (j *jsiiProxy_ComputeSecurityPolicyRuleRateLimitOptionsOutputReference)SetConformAction(val *string) {
	if err := j.validateSetConformActionParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"conformAction",
		val,
	)
}

func (j *jsiiProxy_ComputeSecurityPolicyRuleRateLimitOptionsOutputReference)SetEnforceOnKey(val *string) {
	if err := j.validateSetEnforceOnKeyParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"enforceOnKey",
		val,
	)
}

func (j *jsiiProxy_ComputeSecurityPolicyRuleRateLimitOptionsOutputReference)SetEnforceOnKeyName(val *string) {
	if err := j.validateSetEnforceOnKeyNameParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"enforceOnKeyName",
		val,
	)
}

func (j *jsiiProxy_ComputeSecurityPolicyRuleRateLimitOptionsOutputReference)SetExceedAction(val *string) {
	if err := j.validateSetExceedActionParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"exceedAction",
		val,
	)
}

func (j *jsiiProxy_ComputeSecurityPolicyRuleRateLimitOptionsOutputReference)SetInternalValue(val *ComputeSecurityPolicyRuleRateLimitOptions) {
	if err := j.validateSetInternalValueParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"internalValue",
		val,
	)
}

func (j *jsiiProxy_ComputeSecurityPolicyRuleRateLimitOptionsOutputReference)SetTerraformAttribute(val *string) {
	if err := j.validateSetTerraformAttributeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformAttribute",
		val,
	)
}

func (j *jsiiProxy_ComputeSecurityPolicyRuleRateLimitOptionsOutputReference)SetTerraformResource(val cdktf.IInterpolatingParent) {
	if err := j.validateSetTerraformResourceParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformResource",
		val,
	)
}

func (c *jsiiProxy_ComputeSecurityPolicyRuleRateLimitOptionsOutputReference) ComputeFqn() *string {
	var returns *string

	_jsii_.Invoke(
		c,
		"computeFqn",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComputeSecurityPolicyRuleRateLimitOptionsOutputReference) GetAnyMapAttribute(terraformAttribute *string) *map[string]interface{} {
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

func (c *jsiiProxy_ComputeSecurityPolicyRuleRateLimitOptionsOutputReference) GetBooleanAttribute(terraformAttribute *string) cdktf.IResolvable {
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

func (c *jsiiProxy_ComputeSecurityPolicyRuleRateLimitOptionsOutputReference) GetBooleanMapAttribute(terraformAttribute *string) *map[string]*bool {
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

func (c *jsiiProxy_ComputeSecurityPolicyRuleRateLimitOptionsOutputReference) GetListAttribute(terraformAttribute *string) *[]*string {
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

func (c *jsiiProxy_ComputeSecurityPolicyRuleRateLimitOptionsOutputReference) GetNumberAttribute(terraformAttribute *string) *float64 {
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

func (c *jsiiProxy_ComputeSecurityPolicyRuleRateLimitOptionsOutputReference) GetNumberListAttribute(terraformAttribute *string) *[]*float64 {
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

func (c *jsiiProxy_ComputeSecurityPolicyRuleRateLimitOptionsOutputReference) GetNumberMapAttribute(terraformAttribute *string) *map[string]*float64 {
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

func (c *jsiiProxy_ComputeSecurityPolicyRuleRateLimitOptionsOutputReference) GetStringAttribute(terraformAttribute *string) *string {
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

func (c *jsiiProxy_ComputeSecurityPolicyRuleRateLimitOptionsOutputReference) GetStringMapAttribute(terraformAttribute *string) *map[string]*string {
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

func (c *jsiiProxy_ComputeSecurityPolicyRuleRateLimitOptionsOutputReference) InterpolationAsList() cdktf.IResolvable {
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		c,
		"interpolationAsList",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComputeSecurityPolicyRuleRateLimitOptionsOutputReference) InterpolationForAttribute(property *string) cdktf.IResolvable {
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

func (c *jsiiProxy_ComputeSecurityPolicyRuleRateLimitOptionsOutputReference) PutBanThreshold(value *ComputeSecurityPolicyRuleRateLimitOptionsBanThreshold) {
	if err := c.validatePutBanThresholdParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putBanThreshold",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ComputeSecurityPolicyRuleRateLimitOptionsOutputReference) PutExceedRedirectOptions(value *ComputeSecurityPolicyRuleRateLimitOptionsExceedRedirectOptions) {
	if err := c.validatePutExceedRedirectOptionsParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putExceedRedirectOptions",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ComputeSecurityPolicyRuleRateLimitOptionsOutputReference) PutRateLimitThreshold(value *ComputeSecurityPolicyRuleRateLimitOptionsRateLimitThreshold) {
	if err := c.validatePutRateLimitThresholdParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putRateLimitThreshold",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ComputeSecurityPolicyRuleRateLimitOptionsOutputReference) ResetBanDurationSec() {
	_jsii_.InvokeVoid(
		c,
		"resetBanDurationSec",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeSecurityPolicyRuleRateLimitOptionsOutputReference) ResetBanThreshold() {
	_jsii_.InvokeVoid(
		c,
		"resetBanThreshold",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeSecurityPolicyRuleRateLimitOptionsOutputReference) ResetEnforceOnKey() {
	_jsii_.InvokeVoid(
		c,
		"resetEnforceOnKey",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeSecurityPolicyRuleRateLimitOptionsOutputReference) ResetEnforceOnKeyName() {
	_jsii_.InvokeVoid(
		c,
		"resetEnforceOnKeyName",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeSecurityPolicyRuleRateLimitOptionsOutputReference) ResetExceedRedirectOptions() {
	_jsii_.InvokeVoid(
		c,
		"resetExceedRedirectOptions",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeSecurityPolicyRuleRateLimitOptionsOutputReference) Resolve(_context cdktf.IResolveContext) interface{} {
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

func (c *jsiiProxy_ComputeSecurityPolicyRuleRateLimitOptionsOutputReference) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		c,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

