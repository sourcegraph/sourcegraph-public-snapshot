package computeurlmap

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/sourcegraph/managed-services-platform-cdktf/gen/google/jsii"

	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/computeurlmap/internal"
)

type ComputeUrlMapPathMatcherRouteRulesMatchRulesOutputReference interface {
	cdktf.ComplexObject
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
	FullPathMatch() *string
	SetFullPathMatch(val *string)
	FullPathMatchInput() *string
	HeaderMatches() ComputeUrlMapPathMatcherRouteRulesMatchRulesHeaderMatchesList
	HeaderMatchesInput() interface{}
	IgnoreCase() interface{}
	SetIgnoreCase(val interface{})
	IgnoreCaseInput() interface{}
	InternalValue() interface{}
	SetInternalValue(val interface{})
	MetadataFilters() ComputeUrlMapPathMatcherRouteRulesMatchRulesMetadataFiltersList
	MetadataFiltersInput() interface{}
	PathTemplateMatch() *string
	SetPathTemplateMatch(val *string)
	PathTemplateMatchInput() *string
	PrefixMatch() *string
	SetPrefixMatch(val *string)
	PrefixMatchInput() *string
	QueryParameterMatches() ComputeUrlMapPathMatcherRouteRulesMatchRulesQueryParameterMatchesList
	QueryParameterMatchesInput() interface{}
	RegexMatch() *string
	SetRegexMatch(val *string)
	RegexMatchInput() *string
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
	PutHeaderMatches(value interface{})
	PutMetadataFilters(value interface{})
	PutQueryParameterMatches(value interface{})
	ResetFullPathMatch()
	ResetHeaderMatches()
	ResetIgnoreCase()
	ResetMetadataFilters()
	ResetPathTemplateMatch()
	ResetPrefixMatch()
	ResetQueryParameterMatches()
	ResetRegexMatch()
	// Produce the Token's value at resolution time.
	// Experimental.
	Resolve(_context cdktf.IResolveContext) interface{}
	// Return a string representation of this resolvable object.
	//
	// Returns a reversible string representation.
	// Experimental.
	ToString() *string
}

// The jsii proxy struct for ComputeUrlMapPathMatcherRouteRulesMatchRulesOutputReference
type jsiiProxy_ComputeUrlMapPathMatcherRouteRulesMatchRulesOutputReference struct {
	internal.Type__cdktfComplexObject
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherRouteRulesMatchRulesOutputReference) ComplexObjectIndex() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"complexObjectIndex",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherRouteRulesMatchRulesOutputReference) ComplexObjectIsFromSet() *bool {
	var returns *bool
	_jsii_.Get(
		j,
		"complexObjectIsFromSet",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherRouteRulesMatchRulesOutputReference) CreationStack() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"creationStack",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherRouteRulesMatchRulesOutputReference) Fqn() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fqn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherRouteRulesMatchRulesOutputReference) FullPathMatch() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fullPathMatch",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherRouteRulesMatchRulesOutputReference) FullPathMatchInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fullPathMatchInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherRouteRulesMatchRulesOutputReference) HeaderMatches() ComputeUrlMapPathMatcherRouteRulesMatchRulesHeaderMatchesList {
	var returns ComputeUrlMapPathMatcherRouteRulesMatchRulesHeaderMatchesList
	_jsii_.Get(
		j,
		"headerMatches",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherRouteRulesMatchRulesOutputReference) HeaderMatchesInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"headerMatchesInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherRouteRulesMatchRulesOutputReference) IgnoreCase() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"ignoreCase",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherRouteRulesMatchRulesOutputReference) IgnoreCaseInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"ignoreCaseInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherRouteRulesMatchRulesOutputReference) InternalValue() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"internalValue",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherRouteRulesMatchRulesOutputReference) MetadataFilters() ComputeUrlMapPathMatcherRouteRulesMatchRulesMetadataFiltersList {
	var returns ComputeUrlMapPathMatcherRouteRulesMatchRulesMetadataFiltersList
	_jsii_.Get(
		j,
		"metadataFilters",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherRouteRulesMatchRulesOutputReference) MetadataFiltersInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"metadataFiltersInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherRouteRulesMatchRulesOutputReference) PathTemplateMatch() *string {
	var returns *string
	_jsii_.Get(
		j,
		"pathTemplateMatch",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherRouteRulesMatchRulesOutputReference) PathTemplateMatchInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"pathTemplateMatchInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherRouteRulesMatchRulesOutputReference) PrefixMatch() *string {
	var returns *string
	_jsii_.Get(
		j,
		"prefixMatch",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherRouteRulesMatchRulesOutputReference) PrefixMatchInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"prefixMatchInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherRouteRulesMatchRulesOutputReference) QueryParameterMatches() ComputeUrlMapPathMatcherRouteRulesMatchRulesQueryParameterMatchesList {
	var returns ComputeUrlMapPathMatcherRouteRulesMatchRulesQueryParameterMatchesList
	_jsii_.Get(
		j,
		"queryParameterMatches",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherRouteRulesMatchRulesOutputReference) QueryParameterMatchesInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"queryParameterMatchesInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherRouteRulesMatchRulesOutputReference) RegexMatch() *string {
	var returns *string
	_jsii_.Get(
		j,
		"regexMatch",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherRouteRulesMatchRulesOutputReference) RegexMatchInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"regexMatchInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherRouteRulesMatchRulesOutputReference) TerraformAttribute() *string {
	var returns *string
	_jsii_.Get(
		j,
		"terraformAttribute",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherRouteRulesMatchRulesOutputReference) TerraformResource() cdktf.IInterpolatingParent {
	var returns cdktf.IInterpolatingParent
	_jsii_.Get(
		j,
		"terraformResource",
		&returns,
	)
	return returns
}


func NewComputeUrlMapPathMatcherRouteRulesMatchRulesOutputReference(terraformResource cdktf.IInterpolatingParent, terraformAttribute *string, complexObjectIndex *float64, complexObjectIsFromSet *bool) ComputeUrlMapPathMatcherRouteRulesMatchRulesOutputReference {
	_init_.Initialize()

	if err := validateNewComputeUrlMapPathMatcherRouteRulesMatchRulesOutputReferenceParameters(terraformResource, terraformAttribute, complexObjectIndex, complexObjectIsFromSet); err != nil {
		panic(err)
	}
	j := jsiiProxy_ComputeUrlMapPathMatcherRouteRulesMatchRulesOutputReference{}

	_jsii_.Create(
		"@cdktf/provider-google.computeUrlMap.ComputeUrlMapPathMatcherRouteRulesMatchRulesOutputReference",
		[]interface{}{terraformResource, terraformAttribute, complexObjectIndex, complexObjectIsFromSet},
		&j,
	)

	return &j
}

func NewComputeUrlMapPathMatcherRouteRulesMatchRulesOutputReference_Override(c ComputeUrlMapPathMatcherRouteRulesMatchRulesOutputReference, terraformResource cdktf.IInterpolatingParent, terraformAttribute *string, complexObjectIndex *float64, complexObjectIsFromSet *bool) {
	_init_.Initialize()

	_jsii_.Create(
		"@cdktf/provider-google.computeUrlMap.ComputeUrlMapPathMatcherRouteRulesMatchRulesOutputReference",
		[]interface{}{terraformResource, terraformAttribute, complexObjectIndex, complexObjectIsFromSet},
		c,
	)
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherRouteRulesMatchRulesOutputReference)SetComplexObjectIndex(val interface{}) {
	if err := j.validateSetComplexObjectIndexParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIndex",
		val,
	)
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherRouteRulesMatchRulesOutputReference)SetComplexObjectIsFromSet(val *bool) {
	if err := j.validateSetComplexObjectIsFromSetParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIsFromSet",
		val,
	)
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherRouteRulesMatchRulesOutputReference)SetFullPathMatch(val *string) {
	if err := j.validateSetFullPathMatchParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"fullPathMatch",
		val,
	)
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherRouteRulesMatchRulesOutputReference)SetIgnoreCase(val interface{}) {
	if err := j.validateSetIgnoreCaseParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"ignoreCase",
		val,
	)
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherRouteRulesMatchRulesOutputReference)SetInternalValue(val interface{}) {
	if err := j.validateSetInternalValueParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"internalValue",
		val,
	)
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherRouteRulesMatchRulesOutputReference)SetPathTemplateMatch(val *string) {
	if err := j.validateSetPathTemplateMatchParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"pathTemplateMatch",
		val,
	)
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherRouteRulesMatchRulesOutputReference)SetPrefixMatch(val *string) {
	if err := j.validateSetPrefixMatchParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"prefixMatch",
		val,
	)
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherRouteRulesMatchRulesOutputReference)SetRegexMatch(val *string) {
	if err := j.validateSetRegexMatchParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"regexMatch",
		val,
	)
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherRouteRulesMatchRulesOutputReference)SetTerraformAttribute(val *string) {
	if err := j.validateSetTerraformAttributeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformAttribute",
		val,
	)
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherRouteRulesMatchRulesOutputReference)SetTerraformResource(val cdktf.IInterpolatingParent) {
	if err := j.validateSetTerraformResourceParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformResource",
		val,
	)
}

func (c *jsiiProxy_ComputeUrlMapPathMatcherRouteRulesMatchRulesOutputReference) ComputeFqn() *string {
	var returns *string

	_jsii_.Invoke(
		c,
		"computeFqn",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComputeUrlMapPathMatcherRouteRulesMatchRulesOutputReference) GetAnyMapAttribute(terraformAttribute *string) *map[string]interface{} {
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

func (c *jsiiProxy_ComputeUrlMapPathMatcherRouteRulesMatchRulesOutputReference) GetBooleanAttribute(terraformAttribute *string) cdktf.IResolvable {
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

func (c *jsiiProxy_ComputeUrlMapPathMatcherRouteRulesMatchRulesOutputReference) GetBooleanMapAttribute(terraformAttribute *string) *map[string]*bool {
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

func (c *jsiiProxy_ComputeUrlMapPathMatcherRouteRulesMatchRulesOutputReference) GetListAttribute(terraformAttribute *string) *[]*string {
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

func (c *jsiiProxy_ComputeUrlMapPathMatcherRouteRulesMatchRulesOutputReference) GetNumberAttribute(terraformAttribute *string) *float64 {
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

func (c *jsiiProxy_ComputeUrlMapPathMatcherRouteRulesMatchRulesOutputReference) GetNumberListAttribute(terraformAttribute *string) *[]*float64 {
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

func (c *jsiiProxy_ComputeUrlMapPathMatcherRouteRulesMatchRulesOutputReference) GetNumberMapAttribute(terraformAttribute *string) *map[string]*float64 {
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

func (c *jsiiProxy_ComputeUrlMapPathMatcherRouteRulesMatchRulesOutputReference) GetStringAttribute(terraformAttribute *string) *string {
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

func (c *jsiiProxy_ComputeUrlMapPathMatcherRouteRulesMatchRulesOutputReference) GetStringMapAttribute(terraformAttribute *string) *map[string]*string {
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

func (c *jsiiProxy_ComputeUrlMapPathMatcherRouteRulesMatchRulesOutputReference) InterpolationAsList() cdktf.IResolvable {
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		c,
		"interpolationAsList",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComputeUrlMapPathMatcherRouteRulesMatchRulesOutputReference) InterpolationForAttribute(property *string) cdktf.IResolvable {
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

func (c *jsiiProxy_ComputeUrlMapPathMatcherRouteRulesMatchRulesOutputReference) PutHeaderMatches(value interface{}) {
	if err := c.validatePutHeaderMatchesParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putHeaderMatches",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ComputeUrlMapPathMatcherRouteRulesMatchRulesOutputReference) PutMetadataFilters(value interface{}) {
	if err := c.validatePutMetadataFiltersParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putMetadataFilters",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ComputeUrlMapPathMatcherRouteRulesMatchRulesOutputReference) PutQueryParameterMatches(value interface{}) {
	if err := c.validatePutQueryParameterMatchesParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putQueryParameterMatches",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ComputeUrlMapPathMatcherRouteRulesMatchRulesOutputReference) ResetFullPathMatch() {
	_jsii_.InvokeVoid(
		c,
		"resetFullPathMatch",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeUrlMapPathMatcherRouteRulesMatchRulesOutputReference) ResetHeaderMatches() {
	_jsii_.InvokeVoid(
		c,
		"resetHeaderMatches",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeUrlMapPathMatcherRouteRulesMatchRulesOutputReference) ResetIgnoreCase() {
	_jsii_.InvokeVoid(
		c,
		"resetIgnoreCase",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeUrlMapPathMatcherRouteRulesMatchRulesOutputReference) ResetMetadataFilters() {
	_jsii_.InvokeVoid(
		c,
		"resetMetadataFilters",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeUrlMapPathMatcherRouteRulesMatchRulesOutputReference) ResetPathTemplateMatch() {
	_jsii_.InvokeVoid(
		c,
		"resetPathTemplateMatch",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeUrlMapPathMatcherRouteRulesMatchRulesOutputReference) ResetPrefixMatch() {
	_jsii_.InvokeVoid(
		c,
		"resetPrefixMatch",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeUrlMapPathMatcherRouteRulesMatchRulesOutputReference) ResetQueryParameterMatches() {
	_jsii_.InvokeVoid(
		c,
		"resetQueryParameterMatches",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeUrlMapPathMatcherRouteRulesMatchRulesOutputReference) ResetRegexMatch() {
	_jsii_.InvokeVoid(
		c,
		"resetRegexMatch",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeUrlMapPathMatcherRouteRulesMatchRulesOutputReference) Resolve(_context cdktf.IResolveContext) interface{} {
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

func (c *jsiiProxy_ComputeUrlMapPathMatcherRouteRulesMatchRulesOutputReference) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		c,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

