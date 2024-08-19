package clouddeploycustomtargettype

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/sourcegraph/managed-services-platform-cdktf/gen/google/jsii"

	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/clouddeploycustomtargettype/internal"
)

type ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesOutputReference interface {
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
	Configs() *[]*string
	SetConfigs(val *[]*string)
	ConfigsInput() *[]*string
	// The creation stack of this resolvable which will be appended to errors thrown during resolution.
	//
	// If this returns an empty array the stack will not be attached.
	// Experimental.
	CreationStack() *[]*string
	// Experimental.
	Fqn() *string
	Git() ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesGitOutputReference
	GitInput() *ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesGit
	GoogleCloudBuildRepo() ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesGoogleCloudBuildRepoOutputReference
	GoogleCloudBuildRepoInput() *ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesGoogleCloudBuildRepo
	GoogleCloudStorage() ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesGoogleCloudStorageOutputReference
	GoogleCloudStorageInput() *ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesGoogleCloudStorage
	InternalValue() interface{}
	SetInternalValue(val interface{})
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
	PutGit(value *ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesGit)
	PutGoogleCloudBuildRepo(value *ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesGoogleCloudBuildRepo)
	PutGoogleCloudStorage(value *ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesGoogleCloudStorage)
	ResetConfigs()
	ResetGit()
	ResetGoogleCloudBuildRepo()
	ResetGoogleCloudStorage()
	// Produce the Token's value at resolution time.
	// Experimental.
	Resolve(_context cdktf.IResolveContext) interface{}
	// Return a string representation of this resolvable object.
	//
	// Returns a reversible string representation.
	// Experimental.
	ToString() *string
}

// The jsii proxy struct for ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesOutputReference
type jsiiProxy_ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesOutputReference struct {
	internal.Type__cdktfComplexObject
}

func (j *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesOutputReference) ComplexObjectIndex() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"complexObjectIndex",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesOutputReference) ComplexObjectIsFromSet() *bool {
	var returns *bool
	_jsii_.Get(
		j,
		"complexObjectIsFromSet",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesOutputReference) Configs() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"configs",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesOutputReference) ConfigsInput() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"configsInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesOutputReference) CreationStack() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"creationStack",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesOutputReference) Fqn() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fqn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesOutputReference) Git() ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesGitOutputReference {
	var returns ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesGitOutputReference
	_jsii_.Get(
		j,
		"git",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesOutputReference) GitInput() *ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesGit {
	var returns *ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesGit
	_jsii_.Get(
		j,
		"gitInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesOutputReference) GoogleCloudBuildRepo() ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesGoogleCloudBuildRepoOutputReference {
	var returns ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesGoogleCloudBuildRepoOutputReference
	_jsii_.Get(
		j,
		"googleCloudBuildRepo",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesOutputReference) GoogleCloudBuildRepoInput() *ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesGoogleCloudBuildRepo {
	var returns *ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesGoogleCloudBuildRepo
	_jsii_.Get(
		j,
		"googleCloudBuildRepoInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesOutputReference) GoogleCloudStorage() ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesGoogleCloudStorageOutputReference {
	var returns ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesGoogleCloudStorageOutputReference
	_jsii_.Get(
		j,
		"googleCloudStorage",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesOutputReference) GoogleCloudStorageInput() *ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesGoogleCloudStorage {
	var returns *ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesGoogleCloudStorage
	_jsii_.Get(
		j,
		"googleCloudStorageInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesOutputReference) InternalValue() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"internalValue",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesOutputReference) TerraformAttribute() *string {
	var returns *string
	_jsii_.Get(
		j,
		"terraformAttribute",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesOutputReference) TerraformResource() cdktf.IInterpolatingParent {
	var returns cdktf.IInterpolatingParent
	_jsii_.Get(
		j,
		"terraformResource",
		&returns,
	)
	return returns
}


func NewClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesOutputReference(terraformResource cdktf.IInterpolatingParent, terraformAttribute *string, complexObjectIndex *float64, complexObjectIsFromSet *bool) ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesOutputReference {
	_init_.Initialize()

	if err := validateNewClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesOutputReferenceParameters(terraformResource, terraformAttribute, complexObjectIndex, complexObjectIsFromSet); err != nil {
		panic(err)
	}
	j := jsiiProxy_ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesOutputReference{}

	_jsii_.Create(
		"@cdktf/provider-google.clouddeployCustomTargetType.ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesOutputReference",
		[]interface{}{terraformResource, terraformAttribute, complexObjectIndex, complexObjectIsFromSet},
		&j,
	)

	return &j
}

func NewClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesOutputReference_Override(c ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesOutputReference, terraformResource cdktf.IInterpolatingParent, terraformAttribute *string, complexObjectIndex *float64, complexObjectIsFromSet *bool) {
	_init_.Initialize()

	_jsii_.Create(
		"@cdktf/provider-google.clouddeployCustomTargetType.ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesOutputReference",
		[]interface{}{terraformResource, terraformAttribute, complexObjectIndex, complexObjectIsFromSet},
		c,
	)
}

func (j *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesOutputReference)SetComplexObjectIndex(val interface{}) {
	if err := j.validateSetComplexObjectIndexParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIndex",
		val,
	)
}

func (j *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesOutputReference)SetComplexObjectIsFromSet(val *bool) {
	if err := j.validateSetComplexObjectIsFromSetParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIsFromSet",
		val,
	)
}

func (j *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesOutputReference)SetConfigs(val *[]*string) {
	if err := j.validateSetConfigsParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"configs",
		val,
	)
}

func (j *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesOutputReference)SetInternalValue(val interface{}) {
	if err := j.validateSetInternalValueParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"internalValue",
		val,
	)
}

func (j *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesOutputReference)SetTerraformAttribute(val *string) {
	if err := j.validateSetTerraformAttributeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformAttribute",
		val,
	)
}

func (j *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesOutputReference)SetTerraformResource(val cdktf.IInterpolatingParent) {
	if err := j.validateSetTerraformResourceParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformResource",
		val,
	)
}

func (c *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesOutputReference) ComputeFqn() *string {
	var returns *string

	_jsii_.Invoke(
		c,
		"computeFqn",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesOutputReference) GetAnyMapAttribute(terraformAttribute *string) *map[string]interface{} {
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

func (c *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesOutputReference) GetBooleanAttribute(terraformAttribute *string) cdktf.IResolvable {
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

func (c *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesOutputReference) GetBooleanMapAttribute(terraformAttribute *string) *map[string]*bool {
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

func (c *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesOutputReference) GetListAttribute(terraformAttribute *string) *[]*string {
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

func (c *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesOutputReference) GetNumberAttribute(terraformAttribute *string) *float64 {
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

func (c *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesOutputReference) GetNumberListAttribute(terraformAttribute *string) *[]*float64 {
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

func (c *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesOutputReference) GetNumberMapAttribute(terraformAttribute *string) *map[string]*float64 {
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

func (c *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesOutputReference) GetStringAttribute(terraformAttribute *string) *string {
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

func (c *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesOutputReference) GetStringMapAttribute(terraformAttribute *string) *map[string]*string {
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

func (c *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesOutputReference) InterpolationAsList() cdktf.IResolvable {
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		c,
		"interpolationAsList",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesOutputReference) InterpolationForAttribute(property *string) cdktf.IResolvable {
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

func (c *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesOutputReference) PutGit(value *ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesGit) {
	if err := c.validatePutGitParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putGit",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesOutputReference) PutGoogleCloudBuildRepo(value *ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesGoogleCloudBuildRepo) {
	if err := c.validatePutGoogleCloudBuildRepoParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putGoogleCloudBuildRepo",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesOutputReference) PutGoogleCloudStorage(value *ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesGoogleCloudStorage) {
	if err := c.validatePutGoogleCloudStorageParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putGoogleCloudStorage",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesOutputReference) ResetConfigs() {
	_jsii_.InvokeVoid(
		c,
		"resetConfigs",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesOutputReference) ResetGit() {
	_jsii_.InvokeVoid(
		c,
		"resetGit",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesOutputReference) ResetGoogleCloudBuildRepo() {
	_jsii_.InvokeVoid(
		c,
		"resetGoogleCloudBuildRepo",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesOutputReference) ResetGoogleCloudStorage() {
	_jsii_.InvokeVoid(
		c,
		"resetGoogleCloudStorage",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesOutputReference) Resolve(_context cdktf.IResolveContext) interface{} {
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

func (c *jsiiProxy_ClouddeployCustomTargetTypeCustomActionsIncludeSkaffoldModulesOutputReference) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		c,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

