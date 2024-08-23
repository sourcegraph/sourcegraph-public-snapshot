package computesecuritypolicy

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/sourcegraph/managed-services-platform-cdktf/gen/google/jsii"

	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/computesecuritypolicy/internal"
)

type ComputeSecurityPolicyRuleOutputReference interface {
	cdktf.ComplexObject
	Action() *string
	SetAction(val *string)
	ActionInput() *string
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
	HeaderAction() ComputeSecurityPolicyRuleHeaderActionOutputReference
	HeaderActionInput() *ComputeSecurityPolicyRuleHeaderAction
	InternalValue() interface{}
	SetInternalValue(val interface{})
	Match() ComputeSecurityPolicyRuleMatchOutputReference
	MatchInput() *ComputeSecurityPolicyRuleMatch
	Preview() interface{}
	SetPreview(val interface{})
	PreviewInput() interface{}
	Priority() *float64
	SetPriority(val *float64)
	PriorityInput() *float64
	RateLimitOptions() ComputeSecurityPolicyRuleRateLimitOptionsOutputReference
	RateLimitOptionsInput() *ComputeSecurityPolicyRuleRateLimitOptions
	RedirectOptions() ComputeSecurityPolicyRuleRedirectOptionsOutputReference
	RedirectOptionsInput() *ComputeSecurityPolicyRuleRedirectOptions
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
	PutHeaderAction(value *ComputeSecurityPolicyRuleHeaderAction)
	PutMatch(value *ComputeSecurityPolicyRuleMatch)
	PutRateLimitOptions(value *ComputeSecurityPolicyRuleRateLimitOptions)
	PutRedirectOptions(value *ComputeSecurityPolicyRuleRedirectOptions)
	ResetDescription()
	ResetHeaderAction()
	ResetPreview()
	ResetRateLimitOptions()
	ResetRedirectOptions()
	// Produce the Token's value at resolution time.
	// Experimental.
	Resolve(_context cdktf.IResolveContext) interface{}
	// Return a string representation of this resolvable object.
	//
	// Returns a reversible string representation.
	// Experimental.
	ToString() *string
}

// The jsii proxy struct for ComputeSecurityPolicyRuleOutputReference
type jsiiProxy_ComputeSecurityPolicyRuleOutputReference struct {
	internal.Type__cdktfComplexObject
}

func (j *jsiiProxy_ComputeSecurityPolicyRuleOutputReference) Action() *string {
	var returns *string
	_jsii_.Get(
		j,
		"action",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSecurityPolicyRuleOutputReference) ActionInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"actionInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSecurityPolicyRuleOutputReference) ComplexObjectIndex() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"complexObjectIndex",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSecurityPolicyRuleOutputReference) ComplexObjectIsFromSet() *bool {
	var returns *bool
	_jsii_.Get(
		j,
		"complexObjectIsFromSet",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSecurityPolicyRuleOutputReference) CreationStack() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"creationStack",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSecurityPolicyRuleOutputReference) Description() *string {
	var returns *string
	_jsii_.Get(
		j,
		"description",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSecurityPolicyRuleOutputReference) DescriptionInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"descriptionInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSecurityPolicyRuleOutputReference) Fqn() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fqn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSecurityPolicyRuleOutputReference) HeaderAction() ComputeSecurityPolicyRuleHeaderActionOutputReference {
	var returns ComputeSecurityPolicyRuleHeaderActionOutputReference
	_jsii_.Get(
		j,
		"headerAction",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSecurityPolicyRuleOutputReference) HeaderActionInput() *ComputeSecurityPolicyRuleHeaderAction {
	var returns *ComputeSecurityPolicyRuleHeaderAction
	_jsii_.Get(
		j,
		"headerActionInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSecurityPolicyRuleOutputReference) InternalValue() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"internalValue",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSecurityPolicyRuleOutputReference) Match() ComputeSecurityPolicyRuleMatchOutputReference {
	var returns ComputeSecurityPolicyRuleMatchOutputReference
	_jsii_.Get(
		j,
		"match",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSecurityPolicyRuleOutputReference) MatchInput() *ComputeSecurityPolicyRuleMatch {
	var returns *ComputeSecurityPolicyRuleMatch
	_jsii_.Get(
		j,
		"matchInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSecurityPolicyRuleOutputReference) Preview() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"preview",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSecurityPolicyRuleOutputReference) PreviewInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"previewInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSecurityPolicyRuleOutputReference) Priority() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"priority",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSecurityPolicyRuleOutputReference) PriorityInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"priorityInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSecurityPolicyRuleOutputReference) RateLimitOptions() ComputeSecurityPolicyRuleRateLimitOptionsOutputReference {
	var returns ComputeSecurityPolicyRuleRateLimitOptionsOutputReference
	_jsii_.Get(
		j,
		"rateLimitOptions",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSecurityPolicyRuleOutputReference) RateLimitOptionsInput() *ComputeSecurityPolicyRuleRateLimitOptions {
	var returns *ComputeSecurityPolicyRuleRateLimitOptions
	_jsii_.Get(
		j,
		"rateLimitOptionsInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSecurityPolicyRuleOutputReference) RedirectOptions() ComputeSecurityPolicyRuleRedirectOptionsOutputReference {
	var returns ComputeSecurityPolicyRuleRedirectOptionsOutputReference
	_jsii_.Get(
		j,
		"redirectOptions",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSecurityPolicyRuleOutputReference) RedirectOptionsInput() *ComputeSecurityPolicyRuleRedirectOptions {
	var returns *ComputeSecurityPolicyRuleRedirectOptions
	_jsii_.Get(
		j,
		"redirectOptionsInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSecurityPolicyRuleOutputReference) TerraformAttribute() *string {
	var returns *string
	_jsii_.Get(
		j,
		"terraformAttribute",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSecurityPolicyRuleOutputReference) TerraformResource() cdktf.IInterpolatingParent {
	var returns cdktf.IInterpolatingParent
	_jsii_.Get(
		j,
		"terraformResource",
		&returns,
	)
	return returns
}


func NewComputeSecurityPolicyRuleOutputReference(terraformResource cdktf.IInterpolatingParent, terraformAttribute *string, complexObjectIndex *float64, complexObjectIsFromSet *bool) ComputeSecurityPolicyRuleOutputReference {
	_init_.Initialize()

	if err := validateNewComputeSecurityPolicyRuleOutputReferenceParameters(terraformResource, terraformAttribute, complexObjectIndex, complexObjectIsFromSet); err != nil {
		panic(err)
	}
	j := jsiiProxy_ComputeSecurityPolicyRuleOutputReference{}

	_jsii_.Create(
		"@cdktf/provider-google.computeSecurityPolicy.ComputeSecurityPolicyRuleOutputReference",
		[]interface{}{terraformResource, terraformAttribute, complexObjectIndex, complexObjectIsFromSet},
		&j,
	)

	return &j
}

func NewComputeSecurityPolicyRuleOutputReference_Override(c ComputeSecurityPolicyRuleOutputReference, terraformResource cdktf.IInterpolatingParent, terraformAttribute *string, complexObjectIndex *float64, complexObjectIsFromSet *bool) {
	_init_.Initialize()

	_jsii_.Create(
		"@cdktf/provider-google.computeSecurityPolicy.ComputeSecurityPolicyRuleOutputReference",
		[]interface{}{terraformResource, terraformAttribute, complexObjectIndex, complexObjectIsFromSet},
		c,
	)
}

func (j *jsiiProxy_ComputeSecurityPolicyRuleOutputReference)SetAction(val *string) {
	if err := j.validateSetActionParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"action",
		val,
	)
}

func (j *jsiiProxy_ComputeSecurityPolicyRuleOutputReference)SetComplexObjectIndex(val interface{}) {
	if err := j.validateSetComplexObjectIndexParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIndex",
		val,
	)
}

func (j *jsiiProxy_ComputeSecurityPolicyRuleOutputReference)SetComplexObjectIsFromSet(val *bool) {
	if err := j.validateSetComplexObjectIsFromSetParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIsFromSet",
		val,
	)
}

func (j *jsiiProxy_ComputeSecurityPolicyRuleOutputReference)SetDescription(val *string) {
	if err := j.validateSetDescriptionParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"description",
		val,
	)
}

func (j *jsiiProxy_ComputeSecurityPolicyRuleOutputReference)SetInternalValue(val interface{}) {
	if err := j.validateSetInternalValueParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"internalValue",
		val,
	)
}

func (j *jsiiProxy_ComputeSecurityPolicyRuleOutputReference)SetPreview(val interface{}) {
	if err := j.validateSetPreviewParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"preview",
		val,
	)
}

func (j *jsiiProxy_ComputeSecurityPolicyRuleOutputReference)SetPriority(val *float64) {
	if err := j.validateSetPriorityParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"priority",
		val,
	)
}

func (j *jsiiProxy_ComputeSecurityPolicyRuleOutputReference)SetTerraformAttribute(val *string) {
	if err := j.validateSetTerraformAttributeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformAttribute",
		val,
	)
}

func (j *jsiiProxy_ComputeSecurityPolicyRuleOutputReference)SetTerraformResource(val cdktf.IInterpolatingParent) {
	if err := j.validateSetTerraformResourceParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformResource",
		val,
	)
}

func (c *jsiiProxy_ComputeSecurityPolicyRuleOutputReference) ComputeFqn() *string {
	var returns *string

	_jsii_.Invoke(
		c,
		"computeFqn",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComputeSecurityPolicyRuleOutputReference) GetAnyMapAttribute(terraformAttribute *string) *map[string]interface{} {
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

func (c *jsiiProxy_ComputeSecurityPolicyRuleOutputReference) GetBooleanAttribute(terraformAttribute *string) cdktf.IResolvable {
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

func (c *jsiiProxy_ComputeSecurityPolicyRuleOutputReference) GetBooleanMapAttribute(terraformAttribute *string) *map[string]*bool {
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

func (c *jsiiProxy_ComputeSecurityPolicyRuleOutputReference) GetListAttribute(terraformAttribute *string) *[]*string {
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

func (c *jsiiProxy_ComputeSecurityPolicyRuleOutputReference) GetNumberAttribute(terraformAttribute *string) *float64 {
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

func (c *jsiiProxy_ComputeSecurityPolicyRuleOutputReference) GetNumberListAttribute(terraformAttribute *string) *[]*float64 {
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

func (c *jsiiProxy_ComputeSecurityPolicyRuleOutputReference) GetNumberMapAttribute(terraformAttribute *string) *map[string]*float64 {
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

func (c *jsiiProxy_ComputeSecurityPolicyRuleOutputReference) GetStringAttribute(terraformAttribute *string) *string {
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

func (c *jsiiProxy_ComputeSecurityPolicyRuleOutputReference) GetStringMapAttribute(terraformAttribute *string) *map[string]*string {
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

func (c *jsiiProxy_ComputeSecurityPolicyRuleOutputReference) InterpolationAsList() cdktf.IResolvable {
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		c,
		"interpolationAsList",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComputeSecurityPolicyRuleOutputReference) InterpolationForAttribute(property *string) cdktf.IResolvable {
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

func (c *jsiiProxy_ComputeSecurityPolicyRuleOutputReference) PutHeaderAction(value *ComputeSecurityPolicyRuleHeaderAction) {
	if err := c.validatePutHeaderActionParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putHeaderAction",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ComputeSecurityPolicyRuleOutputReference) PutMatch(value *ComputeSecurityPolicyRuleMatch) {
	if err := c.validatePutMatchParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putMatch",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ComputeSecurityPolicyRuleOutputReference) PutRateLimitOptions(value *ComputeSecurityPolicyRuleRateLimitOptions) {
	if err := c.validatePutRateLimitOptionsParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putRateLimitOptions",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ComputeSecurityPolicyRuleOutputReference) PutRedirectOptions(value *ComputeSecurityPolicyRuleRedirectOptions) {
	if err := c.validatePutRedirectOptionsParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putRedirectOptions",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ComputeSecurityPolicyRuleOutputReference) ResetDescription() {
	_jsii_.InvokeVoid(
		c,
		"resetDescription",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeSecurityPolicyRuleOutputReference) ResetHeaderAction() {
	_jsii_.InvokeVoid(
		c,
		"resetHeaderAction",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeSecurityPolicyRuleOutputReference) ResetPreview() {
	_jsii_.InvokeVoid(
		c,
		"resetPreview",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeSecurityPolicyRuleOutputReference) ResetRateLimitOptions() {
	_jsii_.InvokeVoid(
		c,
		"resetRateLimitOptions",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeSecurityPolicyRuleOutputReference) ResetRedirectOptions() {
	_jsii_.InvokeVoid(
		c,
		"resetRedirectOptions",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeSecurityPolicyRuleOutputReference) Resolve(_context cdktf.IResolveContext) interface{} {
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

func (c *jsiiProxy_ComputeSecurityPolicyRuleOutputReference) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		c,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

