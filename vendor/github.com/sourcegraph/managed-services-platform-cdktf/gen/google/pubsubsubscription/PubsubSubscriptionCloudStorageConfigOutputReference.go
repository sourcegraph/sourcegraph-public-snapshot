package pubsubsubscription

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/sourcegraph/managed-services-platform-cdktf/gen/google/jsii"

	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/pubsubsubscription/internal"
)

type PubsubSubscriptionCloudStorageConfigOutputReference interface {
	cdktf.ComplexObject
	AvroConfig() PubsubSubscriptionCloudStorageConfigAvroConfigOutputReference
	AvroConfigInput() *PubsubSubscriptionCloudStorageConfigAvroConfig
	Bucket() *string
	SetBucket(val *string)
	BucketInput() *string
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
	FilenamePrefix() *string
	SetFilenamePrefix(val *string)
	FilenamePrefixInput() *string
	FilenameSuffix() *string
	SetFilenameSuffix(val *string)
	FilenameSuffixInput() *string
	// Experimental.
	Fqn() *string
	InternalValue() *PubsubSubscriptionCloudStorageConfig
	SetInternalValue(val *PubsubSubscriptionCloudStorageConfig)
	MaxBytes() *float64
	SetMaxBytes(val *float64)
	MaxBytesInput() *float64
	MaxDuration() *string
	SetMaxDuration(val *string)
	MaxDurationInput() *string
	State() *string
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
	PutAvroConfig(value *PubsubSubscriptionCloudStorageConfigAvroConfig)
	ResetAvroConfig()
	ResetFilenamePrefix()
	ResetFilenameSuffix()
	ResetMaxBytes()
	ResetMaxDuration()
	// Produce the Token's value at resolution time.
	// Experimental.
	Resolve(_context cdktf.IResolveContext) interface{}
	// Return a string representation of this resolvable object.
	//
	// Returns a reversible string representation.
	// Experimental.
	ToString() *string
}

// The jsii proxy struct for PubsubSubscriptionCloudStorageConfigOutputReference
type jsiiProxy_PubsubSubscriptionCloudStorageConfigOutputReference struct {
	internal.Type__cdktfComplexObject
}

func (j *jsiiProxy_PubsubSubscriptionCloudStorageConfigOutputReference) AvroConfig() PubsubSubscriptionCloudStorageConfigAvroConfigOutputReference {
	var returns PubsubSubscriptionCloudStorageConfigAvroConfigOutputReference
	_jsii_.Get(
		j,
		"avroConfig",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscriptionCloudStorageConfigOutputReference) AvroConfigInput() *PubsubSubscriptionCloudStorageConfigAvroConfig {
	var returns *PubsubSubscriptionCloudStorageConfigAvroConfig
	_jsii_.Get(
		j,
		"avroConfigInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscriptionCloudStorageConfigOutputReference) Bucket() *string {
	var returns *string
	_jsii_.Get(
		j,
		"bucket",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscriptionCloudStorageConfigOutputReference) BucketInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"bucketInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscriptionCloudStorageConfigOutputReference) ComplexObjectIndex() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"complexObjectIndex",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscriptionCloudStorageConfigOutputReference) ComplexObjectIsFromSet() *bool {
	var returns *bool
	_jsii_.Get(
		j,
		"complexObjectIsFromSet",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscriptionCloudStorageConfigOutputReference) CreationStack() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"creationStack",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscriptionCloudStorageConfigOutputReference) FilenamePrefix() *string {
	var returns *string
	_jsii_.Get(
		j,
		"filenamePrefix",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscriptionCloudStorageConfigOutputReference) FilenamePrefixInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"filenamePrefixInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscriptionCloudStorageConfigOutputReference) FilenameSuffix() *string {
	var returns *string
	_jsii_.Get(
		j,
		"filenameSuffix",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscriptionCloudStorageConfigOutputReference) FilenameSuffixInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"filenameSuffixInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscriptionCloudStorageConfigOutputReference) Fqn() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fqn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscriptionCloudStorageConfigOutputReference) InternalValue() *PubsubSubscriptionCloudStorageConfig {
	var returns *PubsubSubscriptionCloudStorageConfig
	_jsii_.Get(
		j,
		"internalValue",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscriptionCloudStorageConfigOutputReference) MaxBytes() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"maxBytes",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscriptionCloudStorageConfigOutputReference) MaxBytesInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"maxBytesInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscriptionCloudStorageConfigOutputReference) MaxDuration() *string {
	var returns *string
	_jsii_.Get(
		j,
		"maxDuration",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscriptionCloudStorageConfigOutputReference) MaxDurationInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"maxDurationInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscriptionCloudStorageConfigOutputReference) State() *string {
	var returns *string
	_jsii_.Get(
		j,
		"state",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscriptionCloudStorageConfigOutputReference) TerraformAttribute() *string {
	var returns *string
	_jsii_.Get(
		j,
		"terraformAttribute",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscriptionCloudStorageConfigOutputReference) TerraformResource() cdktf.IInterpolatingParent {
	var returns cdktf.IInterpolatingParent
	_jsii_.Get(
		j,
		"terraformResource",
		&returns,
	)
	return returns
}


func NewPubsubSubscriptionCloudStorageConfigOutputReference(terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) PubsubSubscriptionCloudStorageConfigOutputReference {
	_init_.Initialize()

	if err := validateNewPubsubSubscriptionCloudStorageConfigOutputReferenceParameters(terraformResource, terraformAttribute); err != nil {
		panic(err)
	}
	j := jsiiProxy_PubsubSubscriptionCloudStorageConfigOutputReference{}

	_jsii_.Create(
		"@cdktf/provider-google.pubsubSubscription.PubsubSubscriptionCloudStorageConfigOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		&j,
	)

	return &j
}

func NewPubsubSubscriptionCloudStorageConfigOutputReference_Override(p PubsubSubscriptionCloudStorageConfigOutputReference, terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) {
	_init_.Initialize()

	_jsii_.Create(
		"@cdktf/provider-google.pubsubSubscription.PubsubSubscriptionCloudStorageConfigOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		p,
	)
}

func (j *jsiiProxy_PubsubSubscriptionCloudStorageConfigOutputReference)SetBucket(val *string) {
	if err := j.validateSetBucketParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"bucket",
		val,
	)
}

func (j *jsiiProxy_PubsubSubscriptionCloudStorageConfigOutputReference)SetComplexObjectIndex(val interface{}) {
	if err := j.validateSetComplexObjectIndexParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIndex",
		val,
	)
}

func (j *jsiiProxy_PubsubSubscriptionCloudStorageConfigOutputReference)SetComplexObjectIsFromSet(val *bool) {
	if err := j.validateSetComplexObjectIsFromSetParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIsFromSet",
		val,
	)
}

func (j *jsiiProxy_PubsubSubscriptionCloudStorageConfigOutputReference)SetFilenamePrefix(val *string) {
	if err := j.validateSetFilenamePrefixParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"filenamePrefix",
		val,
	)
}

func (j *jsiiProxy_PubsubSubscriptionCloudStorageConfigOutputReference)SetFilenameSuffix(val *string) {
	if err := j.validateSetFilenameSuffixParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"filenameSuffix",
		val,
	)
}

func (j *jsiiProxy_PubsubSubscriptionCloudStorageConfigOutputReference)SetInternalValue(val *PubsubSubscriptionCloudStorageConfig) {
	if err := j.validateSetInternalValueParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"internalValue",
		val,
	)
}

func (j *jsiiProxy_PubsubSubscriptionCloudStorageConfigOutputReference)SetMaxBytes(val *float64) {
	if err := j.validateSetMaxBytesParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"maxBytes",
		val,
	)
}

func (j *jsiiProxy_PubsubSubscriptionCloudStorageConfigOutputReference)SetMaxDuration(val *string) {
	if err := j.validateSetMaxDurationParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"maxDuration",
		val,
	)
}

func (j *jsiiProxy_PubsubSubscriptionCloudStorageConfigOutputReference)SetTerraformAttribute(val *string) {
	if err := j.validateSetTerraformAttributeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformAttribute",
		val,
	)
}

func (j *jsiiProxy_PubsubSubscriptionCloudStorageConfigOutputReference)SetTerraformResource(val cdktf.IInterpolatingParent) {
	if err := j.validateSetTerraformResourceParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformResource",
		val,
	)
}

func (p *jsiiProxy_PubsubSubscriptionCloudStorageConfigOutputReference) ComputeFqn() *string {
	var returns *string

	_jsii_.Invoke(
		p,
		"computeFqn",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (p *jsiiProxy_PubsubSubscriptionCloudStorageConfigOutputReference) GetAnyMapAttribute(terraformAttribute *string) *map[string]interface{} {
	if err := p.validateGetAnyMapAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *map[string]interface{}

	_jsii_.Invoke(
		p,
		"getAnyMapAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (p *jsiiProxy_PubsubSubscriptionCloudStorageConfigOutputReference) GetBooleanAttribute(terraformAttribute *string) cdktf.IResolvable {
	if err := p.validateGetBooleanAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		p,
		"getBooleanAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (p *jsiiProxy_PubsubSubscriptionCloudStorageConfigOutputReference) GetBooleanMapAttribute(terraformAttribute *string) *map[string]*bool {
	if err := p.validateGetBooleanMapAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *map[string]*bool

	_jsii_.Invoke(
		p,
		"getBooleanMapAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (p *jsiiProxy_PubsubSubscriptionCloudStorageConfigOutputReference) GetListAttribute(terraformAttribute *string) *[]*string {
	if err := p.validateGetListAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *[]*string

	_jsii_.Invoke(
		p,
		"getListAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (p *jsiiProxy_PubsubSubscriptionCloudStorageConfigOutputReference) GetNumberAttribute(terraformAttribute *string) *float64 {
	if err := p.validateGetNumberAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *float64

	_jsii_.Invoke(
		p,
		"getNumberAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (p *jsiiProxy_PubsubSubscriptionCloudStorageConfigOutputReference) GetNumberListAttribute(terraformAttribute *string) *[]*float64 {
	if err := p.validateGetNumberListAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *[]*float64

	_jsii_.Invoke(
		p,
		"getNumberListAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (p *jsiiProxy_PubsubSubscriptionCloudStorageConfigOutputReference) GetNumberMapAttribute(terraformAttribute *string) *map[string]*float64 {
	if err := p.validateGetNumberMapAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *map[string]*float64

	_jsii_.Invoke(
		p,
		"getNumberMapAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (p *jsiiProxy_PubsubSubscriptionCloudStorageConfigOutputReference) GetStringAttribute(terraformAttribute *string) *string {
	if err := p.validateGetStringAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.Invoke(
		p,
		"getStringAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (p *jsiiProxy_PubsubSubscriptionCloudStorageConfigOutputReference) GetStringMapAttribute(terraformAttribute *string) *map[string]*string {
	if err := p.validateGetStringMapAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *map[string]*string

	_jsii_.Invoke(
		p,
		"getStringMapAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (p *jsiiProxy_PubsubSubscriptionCloudStorageConfigOutputReference) InterpolationAsList() cdktf.IResolvable {
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		p,
		"interpolationAsList",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (p *jsiiProxy_PubsubSubscriptionCloudStorageConfigOutputReference) InterpolationForAttribute(property *string) cdktf.IResolvable {
	if err := p.validateInterpolationForAttributeParameters(property); err != nil {
		panic(err)
	}
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		p,
		"interpolationForAttribute",
		[]interface{}{property},
		&returns,
	)

	return returns
}

func (p *jsiiProxy_PubsubSubscriptionCloudStorageConfigOutputReference) PutAvroConfig(value *PubsubSubscriptionCloudStorageConfigAvroConfig) {
	if err := p.validatePutAvroConfigParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		p,
		"putAvroConfig",
		[]interface{}{value},
	)
}

func (p *jsiiProxy_PubsubSubscriptionCloudStorageConfigOutputReference) ResetAvroConfig() {
	_jsii_.InvokeVoid(
		p,
		"resetAvroConfig",
		nil, // no parameters
	)
}

func (p *jsiiProxy_PubsubSubscriptionCloudStorageConfigOutputReference) ResetFilenamePrefix() {
	_jsii_.InvokeVoid(
		p,
		"resetFilenamePrefix",
		nil, // no parameters
	)
}

func (p *jsiiProxy_PubsubSubscriptionCloudStorageConfigOutputReference) ResetFilenameSuffix() {
	_jsii_.InvokeVoid(
		p,
		"resetFilenameSuffix",
		nil, // no parameters
	)
}

func (p *jsiiProxy_PubsubSubscriptionCloudStorageConfigOutputReference) ResetMaxBytes() {
	_jsii_.InvokeVoid(
		p,
		"resetMaxBytes",
		nil, // no parameters
	)
}

func (p *jsiiProxy_PubsubSubscriptionCloudStorageConfigOutputReference) ResetMaxDuration() {
	_jsii_.InvokeVoid(
		p,
		"resetMaxDuration",
		nil, // no parameters
	)
}

func (p *jsiiProxy_PubsubSubscriptionCloudStorageConfigOutputReference) Resolve(_context cdktf.IResolveContext) interface{} {
	if err := p.validateResolveParameters(_context); err != nil {
		panic(err)
	}
	var returns interface{}

	_jsii_.Invoke(
		p,
		"resolve",
		[]interface{}{_context},
		&returns,
	)

	return returns
}

func (p *jsiiProxy_PubsubSubscriptionCloudStorageConfigOutputReference) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		p,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

