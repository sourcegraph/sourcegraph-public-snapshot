package clouddeploycustomtargettype

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/sourcegraph/managed-services-platform-cdktf/gen/google/jsii"

	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/clouddeploycustomtargettype/internal"
)

type ClouddeployCustomTargetTypeCustomActionsOutputReference interface {
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
	DeployAction() *string
	SetDeployAction(val *string)
	DeployActionInput() *string
	// Experimental.
	Fqn() *string
	IncludeSkaffoldModules() ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesList
	IncludeSkaffoldModulesInput() interface{}
	InternalValue() *ClouddeployCustomTargetTypeCustomActions
	SetInternalValue(val *ClouddeployCustomTargetTypeCustomActions)
	RenderAction() *string
	SetRenderAction(val *string)
	RenderActionInput() *string
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
	PutIncludeSkaffoldModules(value interface{})
	ResetIncludeSkaffoldModules()
	ResetRenderAction()
	// Produce the Token's value at resolution time.
	// Experimental.
	Resolve(_context cdktf.IResolveContext) interface{}
	// Return a string representation of this resolvable object.
	//
	// Returns a reversible string representation.
	// Experimental.
	ToString() *string
}

// The jsii proxy struct for ClouddeployCustomTargetTypeCustomActionsOutputReference
type jsiiProxy_ClouddeployCustomTargetTypeCustomActionsOutputReference struct {
	internal.Type__cdktfComplexObject
}

func (j *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsOutputReference) ComplexObjectIndex() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"complexObjectIndex",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsOutputReference) ComplexObjectIsFromSet() *bool {
	var returns *bool
	_jsii_.Get(
		j,
		"complexObjectIsFromSet",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsOutputReference) CreationStack() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"creationStack",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsOutputReference) DeployAction() *string {
	var returns *string
	_jsii_.Get(
		j,
		"deployAction",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsOutputReference) DeployActionInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"deployActionInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsOutputReference) Fqn() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fqn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsOutputReference) IncludeSkaffoldModules() ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesList {
	var returns ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesList
	_jsii_.Get(
		j,
		"includeSkaffoldModules",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsOutputReference) IncludeSkaffoldModulesInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"includeSkaffoldModulesInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsOutputReference) InternalValue() *ClouddeployCustomTargetTypeCustomActions {
	var returns *ClouddeployCustomTargetTypeCustomActions
	_jsii_.Get(
		j,
		"internalValue",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsOutputReference) RenderAction() *string {
	var returns *string
	_jsii_.Get(
		j,
		"renderAction",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsOutputReference) RenderActionInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"renderActionInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsOutputReference) TerraformAttribute() *string {
	var returns *string
	_jsii_.Get(
		j,
		"terraformAttribute",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsOutputReference) TerraformResource() cdktf.IInterpolatingParent {
	var returns cdktf.IInterpolatingParent
	_jsii_.Get(
		j,
		"terraformResource",
		&returns,
	)
	return returns
}


func NewClouddeployCustomTargetTypeCustomActionsOutputReference(terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) ClouddeployCustomTargetTypeCustomActionsOutputReference {
	_init_.Initialize()

	if err := validateNewClouddeployCustomTargetTypeCustomActionsOutputReferenceParameters(terraformResource, terraformAttribute); err != nil {
		panic(err)
	}
	j := jsiiProxy_ClouddeployCustomTargetTypeCustomActionsOutputReference{}

	_jsii_.Create(
		"@cdktf/provider-google.clouddeployCustomTargetType.ClouddeployCustomTargetTypeCustomActionsOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		&j,
	)

	return &j
}

func NewClouddeployCustomTargetTypeCustomActionsOutputReference_Override(c ClouddeployCustomTargetTypeCustomActionsOutputReference, terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) {
	_init_.Initialize()

	_jsii_.Create(
		"@cdktf/provider-google.clouddeployCustomTargetType.ClouddeployCustomTargetTypeCustomActionsOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		c,
	)
}

func (j *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsOutputReference)SetComplexObjectIndex(val interface{}) {
	if err := j.validateSetComplexObjectIndexParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIndex",
		val,
	)
}

func (j *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsOutputReference)SetComplexObjectIsFromSet(val *bool) {
	if err := j.validateSetComplexObjectIsFromSetParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIsFromSet",
		val,
	)
}

func (j *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsOutputReference)SetDeployAction(val *string) {
	if err := j.validateSetDeployActionParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"deployAction",
		val,
	)
}

func (j *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsOutputReference)SetInternalValue(val *ClouddeployCustomTargetTypeCustomActions) {
	if err := j.validateSetInternalValueParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"internalValue",
		val,
	)
}

func (j *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsOutputReference)SetRenderAction(val *string) {
	if err := j.validateSetRenderActionParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"renderAction",
		val,
	)
}

func (j *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsOutputReference)SetTerraformAttribute(val *string) {
	if err := j.validateSetTerraformAttributeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformAttribute",
		val,
	)
}

func (j *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsOutputReference)SetTerraformResource(val cdktf.IInterpolatingParent) {
	if err := j.validateSetTerraformResourceParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformResource",
		val,
	)
}

func (c *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsOutputReference) ComputeFqn() *string {
	var returns *string

	_jsii_.Invoke(
		c,
		"computeFqn",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsOutputReference) GetAnyMapAttribute(terraformAttribute *string) *map[string]interface{} {
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

func (c *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsOutputReference) GetBooleanAttribute(terraformAttribute *string) cdktf.IResolvable {
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

func (c *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsOutputReference) GetBooleanMapAttribute(terraformAttribute *string) *map[string]*bool {
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

func (c *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsOutputReference) GetListAttribute(terraformAttribute *string) *[]*string {
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

func (c *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsOutputReference) GetNumberAttribute(terraformAttribute *string) *float64 {
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

func (c *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsOutputReference) GetNumberListAttribute(terraformAttribute *string) *[]*float64 {
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

func (c *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsOutputReference) GetNumberMapAttribute(terraformAttribute *string) *map[string]*float64 {
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

func (c *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsOutputReference) GetStringAttribute(terraformAttribute *string) *string {
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

func (c *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsOutputReference) GetStringMapAttribute(terraformAttribute *string) *map[string]*string {
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

func (c *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsOutputReference) InterpolationAsList() cdktf.IResolvable {
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		c,
		"interpolationAsList",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsOutputReference) InterpolationForAttribute(property *string) cdktf.IResolvable {
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

func (c *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsOutputReference) PutIncludeSkaffoldModules(value interface{}) {
	if err := c.validatePutIncludeSkaffoldModulesParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putIncludeSkaffoldModules",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsOutputReference) ResetIncludeSkaffoldModules() {
	_jsii_.InvokeVoid(
		c,
		"resetIncludeSkaffoldModules",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsOutputReference) ResetRenderAction() {
	_jsii_.InvokeVoid(
		c,
		"resetRenderAction",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsOutputReference) Resolve(_context cdktf.IResolveContext) interface{} {
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

func (c *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsOutputReference) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		c,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

