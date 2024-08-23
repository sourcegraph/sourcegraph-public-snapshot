package cloudrunv2service

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/sourcegraph/managed-services-platform-cdktf/gen/google/jsii"

	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/cloudrunv2service/internal"
)

type CloudRunV2ServiceTemplateOutputReference interface {
	cdktf.ComplexObject
	Annotations() *map[string]*string
	SetAnnotations(val *map[string]*string)
	AnnotationsInput() *map[string]*string
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
	Containers() CloudRunV2ServiceTemplateContainersList
	ContainersInput() interface{}
	// The creation stack of this resolvable which will be appended to errors thrown during resolution.
	//
	// If this returns an empty array the stack will not be attached.
	// Experimental.
	CreationStack() *[]*string
	EncryptionKey() *string
	SetEncryptionKey(val *string)
	EncryptionKeyInput() *string
	ExecutionEnvironment() *string
	SetExecutionEnvironment(val *string)
	ExecutionEnvironmentInput() *string
	// Experimental.
	Fqn() *string
	InternalValue() *CloudRunV2ServiceTemplate
	SetInternalValue(val *CloudRunV2ServiceTemplate)
	Labels() *map[string]*string
	SetLabels(val *map[string]*string)
	LabelsInput() *map[string]*string
	MaxInstanceRequestConcurrency() *float64
	SetMaxInstanceRequestConcurrency(val *float64)
	MaxInstanceRequestConcurrencyInput() *float64
	Revision() *string
	SetRevision(val *string)
	RevisionInput() *string
	Scaling() CloudRunV2ServiceTemplateScalingOutputReference
	ScalingInput() *CloudRunV2ServiceTemplateScaling
	ServiceAccount() *string
	SetServiceAccount(val *string)
	ServiceAccountInput() *string
	SessionAffinity() interface{}
	SetSessionAffinity(val interface{})
	SessionAffinityInput() interface{}
	// Experimental.
	TerraformAttribute() *string
	// Experimental.
	SetTerraformAttribute(val *string)
	// Experimental.
	TerraformResource() cdktf.IInterpolatingParent
	// Experimental.
	SetTerraformResource(val cdktf.IInterpolatingParent)
	Timeout() *string
	SetTimeout(val *string)
	TimeoutInput() *string
	Volumes() CloudRunV2ServiceTemplateVolumesList
	VolumesInput() interface{}
	VpcAccess() CloudRunV2ServiceTemplateVpcAccessOutputReference
	VpcAccessInput() *CloudRunV2ServiceTemplateVpcAccess
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
	PutContainers(value interface{})
	PutScaling(value *CloudRunV2ServiceTemplateScaling)
	PutVolumes(value interface{})
	PutVpcAccess(value *CloudRunV2ServiceTemplateVpcAccess)
	ResetAnnotations()
	ResetContainers()
	ResetEncryptionKey()
	ResetExecutionEnvironment()
	ResetLabels()
	ResetMaxInstanceRequestConcurrency()
	ResetRevision()
	ResetScaling()
	ResetServiceAccount()
	ResetSessionAffinity()
	ResetTimeout()
	ResetVolumes()
	ResetVpcAccess()
	// Produce the Token's value at resolution time.
	// Experimental.
	Resolve(_context cdktf.IResolveContext) interface{}
	// Return a string representation of this resolvable object.
	//
	// Returns a reversible string representation.
	// Experimental.
	ToString() *string
}

// The jsii proxy struct for CloudRunV2ServiceTemplateOutputReference
type jsiiProxy_CloudRunV2ServiceTemplateOutputReference struct {
	internal.Type__cdktfComplexObject
}

func (j *jsiiProxy_CloudRunV2ServiceTemplateOutputReference) Annotations() *map[string]*string {
	var returns *map[string]*string
	_jsii_.Get(
		j,
		"annotations",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2ServiceTemplateOutputReference) AnnotationsInput() *map[string]*string {
	var returns *map[string]*string
	_jsii_.Get(
		j,
		"annotationsInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2ServiceTemplateOutputReference) ComplexObjectIndex() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"complexObjectIndex",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2ServiceTemplateOutputReference) ComplexObjectIsFromSet() *bool {
	var returns *bool
	_jsii_.Get(
		j,
		"complexObjectIsFromSet",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2ServiceTemplateOutputReference) Containers() CloudRunV2ServiceTemplateContainersList {
	var returns CloudRunV2ServiceTemplateContainersList
	_jsii_.Get(
		j,
		"containers",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2ServiceTemplateOutputReference) ContainersInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"containersInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2ServiceTemplateOutputReference) CreationStack() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"creationStack",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2ServiceTemplateOutputReference) EncryptionKey() *string {
	var returns *string
	_jsii_.Get(
		j,
		"encryptionKey",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2ServiceTemplateOutputReference) EncryptionKeyInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"encryptionKeyInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2ServiceTemplateOutputReference) ExecutionEnvironment() *string {
	var returns *string
	_jsii_.Get(
		j,
		"executionEnvironment",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2ServiceTemplateOutputReference) ExecutionEnvironmentInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"executionEnvironmentInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2ServiceTemplateOutputReference) Fqn() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fqn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2ServiceTemplateOutputReference) InternalValue() *CloudRunV2ServiceTemplate {
	var returns *CloudRunV2ServiceTemplate
	_jsii_.Get(
		j,
		"internalValue",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2ServiceTemplateOutputReference) Labels() *map[string]*string {
	var returns *map[string]*string
	_jsii_.Get(
		j,
		"labels",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2ServiceTemplateOutputReference) LabelsInput() *map[string]*string {
	var returns *map[string]*string
	_jsii_.Get(
		j,
		"labelsInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2ServiceTemplateOutputReference) MaxInstanceRequestConcurrency() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"maxInstanceRequestConcurrency",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2ServiceTemplateOutputReference) MaxInstanceRequestConcurrencyInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"maxInstanceRequestConcurrencyInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2ServiceTemplateOutputReference) Revision() *string {
	var returns *string
	_jsii_.Get(
		j,
		"revision",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2ServiceTemplateOutputReference) RevisionInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"revisionInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2ServiceTemplateOutputReference) Scaling() CloudRunV2ServiceTemplateScalingOutputReference {
	var returns CloudRunV2ServiceTemplateScalingOutputReference
	_jsii_.Get(
		j,
		"scaling",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2ServiceTemplateOutputReference) ScalingInput() *CloudRunV2ServiceTemplateScaling {
	var returns *CloudRunV2ServiceTemplateScaling
	_jsii_.Get(
		j,
		"scalingInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2ServiceTemplateOutputReference) ServiceAccount() *string {
	var returns *string
	_jsii_.Get(
		j,
		"serviceAccount",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2ServiceTemplateOutputReference) ServiceAccountInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"serviceAccountInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2ServiceTemplateOutputReference) SessionAffinity() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"sessionAffinity",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2ServiceTemplateOutputReference) SessionAffinityInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"sessionAffinityInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2ServiceTemplateOutputReference) TerraformAttribute() *string {
	var returns *string
	_jsii_.Get(
		j,
		"terraformAttribute",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2ServiceTemplateOutputReference) TerraformResource() cdktf.IInterpolatingParent {
	var returns cdktf.IInterpolatingParent
	_jsii_.Get(
		j,
		"terraformResource",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2ServiceTemplateOutputReference) Timeout() *string {
	var returns *string
	_jsii_.Get(
		j,
		"timeout",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2ServiceTemplateOutputReference) TimeoutInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"timeoutInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2ServiceTemplateOutputReference) Volumes() CloudRunV2ServiceTemplateVolumesList {
	var returns CloudRunV2ServiceTemplateVolumesList
	_jsii_.Get(
		j,
		"volumes",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2ServiceTemplateOutputReference) VolumesInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"volumesInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2ServiceTemplateOutputReference) VpcAccess() CloudRunV2ServiceTemplateVpcAccessOutputReference {
	var returns CloudRunV2ServiceTemplateVpcAccessOutputReference
	_jsii_.Get(
		j,
		"vpcAccess",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudRunV2ServiceTemplateOutputReference) VpcAccessInput() *CloudRunV2ServiceTemplateVpcAccess {
	var returns *CloudRunV2ServiceTemplateVpcAccess
	_jsii_.Get(
		j,
		"vpcAccessInput",
		&returns,
	)
	return returns
}


func NewCloudRunV2ServiceTemplateOutputReference(terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) CloudRunV2ServiceTemplateOutputReference {
	_init_.Initialize()

	if err := validateNewCloudRunV2ServiceTemplateOutputReferenceParameters(terraformResource, terraformAttribute); err != nil {
		panic(err)
	}
	j := jsiiProxy_CloudRunV2ServiceTemplateOutputReference{}

	_jsii_.Create(
		"@cdktf/provider-google.cloudRunV2Service.CloudRunV2ServiceTemplateOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		&j,
	)

	return &j
}

func NewCloudRunV2ServiceTemplateOutputReference_Override(c CloudRunV2ServiceTemplateOutputReference, terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) {
	_init_.Initialize()

	_jsii_.Create(
		"@cdktf/provider-google.cloudRunV2Service.CloudRunV2ServiceTemplateOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		c,
	)
}

func (j *jsiiProxy_CloudRunV2ServiceTemplateOutputReference)SetAnnotations(val *map[string]*string) {
	if err := j.validateSetAnnotationsParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"annotations",
		val,
	)
}

func (j *jsiiProxy_CloudRunV2ServiceTemplateOutputReference)SetComplexObjectIndex(val interface{}) {
	if err := j.validateSetComplexObjectIndexParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIndex",
		val,
	)
}

func (j *jsiiProxy_CloudRunV2ServiceTemplateOutputReference)SetComplexObjectIsFromSet(val *bool) {
	if err := j.validateSetComplexObjectIsFromSetParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIsFromSet",
		val,
	)
}

func (j *jsiiProxy_CloudRunV2ServiceTemplateOutputReference)SetEncryptionKey(val *string) {
	if err := j.validateSetEncryptionKeyParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"encryptionKey",
		val,
	)
}

func (j *jsiiProxy_CloudRunV2ServiceTemplateOutputReference)SetExecutionEnvironment(val *string) {
	if err := j.validateSetExecutionEnvironmentParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"executionEnvironment",
		val,
	)
}

func (j *jsiiProxy_CloudRunV2ServiceTemplateOutputReference)SetInternalValue(val *CloudRunV2ServiceTemplate) {
	if err := j.validateSetInternalValueParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"internalValue",
		val,
	)
}

func (j *jsiiProxy_CloudRunV2ServiceTemplateOutputReference)SetLabels(val *map[string]*string) {
	if err := j.validateSetLabelsParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"labels",
		val,
	)
}

func (j *jsiiProxy_CloudRunV2ServiceTemplateOutputReference)SetMaxInstanceRequestConcurrency(val *float64) {
	if err := j.validateSetMaxInstanceRequestConcurrencyParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"maxInstanceRequestConcurrency",
		val,
	)
}

func (j *jsiiProxy_CloudRunV2ServiceTemplateOutputReference)SetRevision(val *string) {
	if err := j.validateSetRevisionParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"revision",
		val,
	)
}

func (j *jsiiProxy_CloudRunV2ServiceTemplateOutputReference)SetServiceAccount(val *string) {
	if err := j.validateSetServiceAccountParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"serviceAccount",
		val,
	)
}

func (j *jsiiProxy_CloudRunV2ServiceTemplateOutputReference)SetSessionAffinity(val interface{}) {
	if err := j.validateSetSessionAffinityParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"sessionAffinity",
		val,
	)
}

func (j *jsiiProxy_CloudRunV2ServiceTemplateOutputReference)SetTerraformAttribute(val *string) {
	if err := j.validateSetTerraformAttributeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformAttribute",
		val,
	)
}

func (j *jsiiProxy_CloudRunV2ServiceTemplateOutputReference)SetTerraformResource(val cdktf.IInterpolatingParent) {
	if err := j.validateSetTerraformResourceParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformResource",
		val,
	)
}

func (j *jsiiProxy_CloudRunV2ServiceTemplateOutputReference)SetTimeout(val *string) {
	if err := j.validateSetTimeoutParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"timeout",
		val,
	)
}

func (c *jsiiProxy_CloudRunV2ServiceTemplateOutputReference) ComputeFqn() *string {
	var returns *string

	_jsii_.Invoke(
		c,
		"computeFqn",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (c *jsiiProxy_CloudRunV2ServiceTemplateOutputReference) GetAnyMapAttribute(terraformAttribute *string) *map[string]interface{} {
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

func (c *jsiiProxy_CloudRunV2ServiceTemplateOutputReference) GetBooleanAttribute(terraformAttribute *string) cdktf.IResolvable {
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

func (c *jsiiProxy_CloudRunV2ServiceTemplateOutputReference) GetBooleanMapAttribute(terraformAttribute *string) *map[string]*bool {
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

func (c *jsiiProxy_CloudRunV2ServiceTemplateOutputReference) GetListAttribute(terraformAttribute *string) *[]*string {
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

func (c *jsiiProxy_CloudRunV2ServiceTemplateOutputReference) GetNumberAttribute(terraformAttribute *string) *float64 {
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

func (c *jsiiProxy_CloudRunV2ServiceTemplateOutputReference) GetNumberListAttribute(terraformAttribute *string) *[]*float64 {
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

func (c *jsiiProxy_CloudRunV2ServiceTemplateOutputReference) GetNumberMapAttribute(terraformAttribute *string) *map[string]*float64 {
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

func (c *jsiiProxy_CloudRunV2ServiceTemplateOutputReference) GetStringAttribute(terraformAttribute *string) *string {
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

func (c *jsiiProxy_CloudRunV2ServiceTemplateOutputReference) GetStringMapAttribute(terraformAttribute *string) *map[string]*string {
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

func (c *jsiiProxy_CloudRunV2ServiceTemplateOutputReference) InterpolationAsList() cdktf.IResolvable {
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		c,
		"interpolationAsList",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (c *jsiiProxy_CloudRunV2ServiceTemplateOutputReference) InterpolationForAttribute(property *string) cdktf.IResolvable {
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

func (c *jsiiProxy_CloudRunV2ServiceTemplateOutputReference) PutContainers(value interface{}) {
	if err := c.validatePutContainersParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putContainers",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_CloudRunV2ServiceTemplateOutputReference) PutScaling(value *CloudRunV2ServiceTemplateScaling) {
	if err := c.validatePutScalingParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putScaling",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_CloudRunV2ServiceTemplateOutputReference) PutVolumes(value interface{}) {
	if err := c.validatePutVolumesParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putVolumes",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_CloudRunV2ServiceTemplateOutputReference) PutVpcAccess(value *CloudRunV2ServiceTemplateVpcAccess) {
	if err := c.validatePutVpcAccessParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putVpcAccess",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_CloudRunV2ServiceTemplateOutputReference) ResetAnnotations() {
	_jsii_.InvokeVoid(
		c,
		"resetAnnotations",
		nil, // no parameters
	)
}

func (c *jsiiProxy_CloudRunV2ServiceTemplateOutputReference) ResetContainers() {
	_jsii_.InvokeVoid(
		c,
		"resetContainers",
		nil, // no parameters
	)
}

func (c *jsiiProxy_CloudRunV2ServiceTemplateOutputReference) ResetEncryptionKey() {
	_jsii_.InvokeVoid(
		c,
		"resetEncryptionKey",
		nil, // no parameters
	)
}

func (c *jsiiProxy_CloudRunV2ServiceTemplateOutputReference) ResetExecutionEnvironment() {
	_jsii_.InvokeVoid(
		c,
		"resetExecutionEnvironment",
		nil, // no parameters
	)
}

func (c *jsiiProxy_CloudRunV2ServiceTemplateOutputReference) ResetLabels() {
	_jsii_.InvokeVoid(
		c,
		"resetLabels",
		nil, // no parameters
	)
}

func (c *jsiiProxy_CloudRunV2ServiceTemplateOutputReference) ResetMaxInstanceRequestConcurrency() {
	_jsii_.InvokeVoid(
		c,
		"resetMaxInstanceRequestConcurrency",
		nil, // no parameters
	)
}

func (c *jsiiProxy_CloudRunV2ServiceTemplateOutputReference) ResetRevision() {
	_jsii_.InvokeVoid(
		c,
		"resetRevision",
		nil, // no parameters
	)
}

func (c *jsiiProxy_CloudRunV2ServiceTemplateOutputReference) ResetScaling() {
	_jsii_.InvokeVoid(
		c,
		"resetScaling",
		nil, // no parameters
	)
}

func (c *jsiiProxy_CloudRunV2ServiceTemplateOutputReference) ResetServiceAccount() {
	_jsii_.InvokeVoid(
		c,
		"resetServiceAccount",
		nil, // no parameters
	)
}

func (c *jsiiProxy_CloudRunV2ServiceTemplateOutputReference) ResetSessionAffinity() {
	_jsii_.InvokeVoid(
		c,
		"resetSessionAffinity",
		nil, // no parameters
	)
}

func (c *jsiiProxy_CloudRunV2ServiceTemplateOutputReference) ResetTimeout() {
	_jsii_.InvokeVoid(
		c,
		"resetTimeout",
		nil, // no parameters
	)
}

func (c *jsiiProxy_CloudRunV2ServiceTemplateOutputReference) ResetVolumes() {
	_jsii_.InvokeVoid(
		c,
		"resetVolumes",
		nil, // no parameters
	)
}

func (c *jsiiProxy_CloudRunV2ServiceTemplateOutputReference) ResetVpcAccess() {
	_jsii_.InvokeVoid(
		c,
		"resetVpcAccess",
		nil, // no parameters
	)
}

func (c *jsiiProxy_CloudRunV2ServiceTemplateOutputReference) Resolve(_context cdktf.IResolveContext) interface{} {
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

func (c *jsiiProxy_CloudRunV2ServiceTemplateOutputReference) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		c,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

