package loggingmetric

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/sourcegraph/managed-services-platform-cdktf/gen/google/jsii"

	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/loggingmetric/internal"
)

type LoggingMetricBucketOptionsOutputReference interface {
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
	ExplicitBuckets() LoggingMetricBucketOptionsExplicitBucketsOutputReference
	ExplicitBucketsInput() *LoggingMetricBucketOptionsExplicitBuckets
	ExponentialBuckets() LoggingMetricBucketOptionsExponentialBucketsOutputReference
	ExponentialBucketsInput() *LoggingMetricBucketOptionsExponentialBuckets
	// Experimental.
	Fqn() *string
	InternalValue() *LoggingMetricBucketOptions
	SetInternalValue(val *LoggingMetricBucketOptions)
	LinearBuckets() LoggingMetricBucketOptionsLinearBucketsOutputReference
	LinearBucketsInput() *LoggingMetricBucketOptionsLinearBuckets
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
	PutExplicitBuckets(value *LoggingMetricBucketOptionsExplicitBuckets)
	PutExponentialBuckets(value *LoggingMetricBucketOptionsExponentialBuckets)
	PutLinearBuckets(value *LoggingMetricBucketOptionsLinearBuckets)
	ResetExplicitBuckets()
	ResetExponentialBuckets()
	ResetLinearBuckets()
	// Produce the Token's value at resolution time.
	// Experimental.
	Resolve(_context cdktf.IResolveContext) interface{}
	// Return a string representation of this resolvable object.
	//
	// Returns a reversible string representation.
	// Experimental.
	ToString() *string
}

// The jsii proxy struct for LoggingMetricBucketOptionsOutputReference
type jsiiProxy_LoggingMetricBucketOptionsOutputReference struct {
	internal.Type__cdktfComplexObject
}

func (j *jsiiProxy_LoggingMetricBucketOptionsOutputReference) ComplexObjectIndex() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"complexObjectIndex",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_LoggingMetricBucketOptionsOutputReference) ComplexObjectIsFromSet() *bool {
	var returns *bool
	_jsii_.Get(
		j,
		"complexObjectIsFromSet",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_LoggingMetricBucketOptionsOutputReference) CreationStack() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"creationStack",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_LoggingMetricBucketOptionsOutputReference) ExplicitBuckets() LoggingMetricBucketOptionsExplicitBucketsOutputReference {
	var returns LoggingMetricBucketOptionsExplicitBucketsOutputReference
	_jsii_.Get(
		j,
		"explicitBuckets",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_LoggingMetricBucketOptionsOutputReference) ExplicitBucketsInput() *LoggingMetricBucketOptionsExplicitBuckets {
	var returns *LoggingMetricBucketOptionsExplicitBuckets
	_jsii_.Get(
		j,
		"explicitBucketsInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_LoggingMetricBucketOptionsOutputReference) ExponentialBuckets() LoggingMetricBucketOptionsExponentialBucketsOutputReference {
	var returns LoggingMetricBucketOptionsExponentialBucketsOutputReference
	_jsii_.Get(
		j,
		"exponentialBuckets",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_LoggingMetricBucketOptionsOutputReference) ExponentialBucketsInput() *LoggingMetricBucketOptionsExponentialBuckets {
	var returns *LoggingMetricBucketOptionsExponentialBuckets
	_jsii_.Get(
		j,
		"exponentialBucketsInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_LoggingMetricBucketOptionsOutputReference) Fqn() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fqn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_LoggingMetricBucketOptionsOutputReference) InternalValue() *LoggingMetricBucketOptions {
	var returns *LoggingMetricBucketOptions
	_jsii_.Get(
		j,
		"internalValue",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_LoggingMetricBucketOptionsOutputReference) LinearBuckets() LoggingMetricBucketOptionsLinearBucketsOutputReference {
	var returns LoggingMetricBucketOptionsLinearBucketsOutputReference
	_jsii_.Get(
		j,
		"linearBuckets",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_LoggingMetricBucketOptionsOutputReference) LinearBucketsInput() *LoggingMetricBucketOptionsLinearBuckets {
	var returns *LoggingMetricBucketOptionsLinearBuckets
	_jsii_.Get(
		j,
		"linearBucketsInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_LoggingMetricBucketOptionsOutputReference) TerraformAttribute() *string {
	var returns *string
	_jsii_.Get(
		j,
		"terraformAttribute",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_LoggingMetricBucketOptionsOutputReference) TerraformResource() cdktf.IInterpolatingParent {
	var returns cdktf.IInterpolatingParent
	_jsii_.Get(
		j,
		"terraformResource",
		&returns,
	)
	return returns
}


func NewLoggingMetricBucketOptionsOutputReference(terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) LoggingMetricBucketOptionsOutputReference {
	_init_.Initialize()

	if err := validateNewLoggingMetricBucketOptionsOutputReferenceParameters(terraformResource, terraformAttribute); err != nil {
		panic(err)
	}
	j := jsiiProxy_LoggingMetricBucketOptionsOutputReference{}

	_jsii_.Create(
		"@cdktf/provider-google.loggingMetric.LoggingMetricBucketOptionsOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		&j,
	)

	return &j
}

func NewLoggingMetricBucketOptionsOutputReference_Override(l LoggingMetricBucketOptionsOutputReference, terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) {
	_init_.Initialize()

	_jsii_.Create(
		"@cdktf/provider-google.loggingMetric.LoggingMetricBucketOptionsOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		l,
	)
}

func (j *jsiiProxy_LoggingMetricBucketOptionsOutputReference)SetComplexObjectIndex(val interface{}) {
	if err := j.validateSetComplexObjectIndexParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIndex",
		val,
	)
}

func (j *jsiiProxy_LoggingMetricBucketOptionsOutputReference)SetComplexObjectIsFromSet(val *bool) {
	if err := j.validateSetComplexObjectIsFromSetParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIsFromSet",
		val,
	)
}

func (j *jsiiProxy_LoggingMetricBucketOptionsOutputReference)SetInternalValue(val *LoggingMetricBucketOptions) {
	if err := j.validateSetInternalValueParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"internalValue",
		val,
	)
}

func (j *jsiiProxy_LoggingMetricBucketOptionsOutputReference)SetTerraformAttribute(val *string) {
	if err := j.validateSetTerraformAttributeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformAttribute",
		val,
	)
}

func (j *jsiiProxy_LoggingMetricBucketOptionsOutputReference)SetTerraformResource(val cdktf.IInterpolatingParent) {
	if err := j.validateSetTerraformResourceParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformResource",
		val,
	)
}

func (l *jsiiProxy_LoggingMetricBucketOptionsOutputReference) ComputeFqn() *string {
	var returns *string

	_jsii_.Invoke(
		l,
		"computeFqn",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (l *jsiiProxy_LoggingMetricBucketOptionsOutputReference) GetAnyMapAttribute(terraformAttribute *string) *map[string]interface{} {
	if err := l.validateGetAnyMapAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *map[string]interface{}

	_jsii_.Invoke(
		l,
		"getAnyMapAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (l *jsiiProxy_LoggingMetricBucketOptionsOutputReference) GetBooleanAttribute(terraformAttribute *string) cdktf.IResolvable {
	if err := l.validateGetBooleanAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		l,
		"getBooleanAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (l *jsiiProxy_LoggingMetricBucketOptionsOutputReference) GetBooleanMapAttribute(terraformAttribute *string) *map[string]*bool {
	if err := l.validateGetBooleanMapAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *map[string]*bool

	_jsii_.Invoke(
		l,
		"getBooleanMapAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (l *jsiiProxy_LoggingMetricBucketOptionsOutputReference) GetListAttribute(terraformAttribute *string) *[]*string {
	if err := l.validateGetListAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *[]*string

	_jsii_.Invoke(
		l,
		"getListAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (l *jsiiProxy_LoggingMetricBucketOptionsOutputReference) GetNumberAttribute(terraformAttribute *string) *float64 {
	if err := l.validateGetNumberAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *float64

	_jsii_.Invoke(
		l,
		"getNumberAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (l *jsiiProxy_LoggingMetricBucketOptionsOutputReference) GetNumberListAttribute(terraformAttribute *string) *[]*float64 {
	if err := l.validateGetNumberListAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *[]*float64

	_jsii_.Invoke(
		l,
		"getNumberListAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (l *jsiiProxy_LoggingMetricBucketOptionsOutputReference) GetNumberMapAttribute(terraformAttribute *string) *map[string]*float64 {
	if err := l.validateGetNumberMapAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *map[string]*float64

	_jsii_.Invoke(
		l,
		"getNumberMapAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (l *jsiiProxy_LoggingMetricBucketOptionsOutputReference) GetStringAttribute(terraformAttribute *string) *string {
	if err := l.validateGetStringAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.Invoke(
		l,
		"getStringAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (l *jsiiProxy_LoggingMetricBucketOptionsOutputReference) GetStringMapAttribute(terraformAttribute *string) *map[string]*string {
	if err := l.validateGetStringMapAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *map[string]*string

	_jsii_.Invoke(
		l,
		"getStringMapAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (l *jsiiProxy_LoggingMetricBucketOptionsOutputReference) InterpolationAsList() cdktf.IResolvable {
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		l,
		"interpolationAsList",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (l *jsiiProxy_LoggingMetricBucketOptionsOutputReference) InterpolationForAttribute(property *string) cdktf.IResolvable {
	if err := l.validateInterpolationForAttributeParameters(property); err != nil {
		panic(err)
	}
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		l,
		"interpolationForAttribute",
		[]interface{}{property},
		&returns,
	)

	return returns
}

func (l *jsiiProxy_LoggingMetricBucketOptionsOutputReference) PutExplicitBuckets(value *LoggingMetricBucketOptionsExplicitBuckets) {
	if err := l.validatePutExplicitBucketsParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		l,
		"putExplicitBuckets",
		[]interface{}{value},
	)
}

func (l *jsiiProxy_LoggingMetricBucketOptionsOutputReference) PutExponentialBuckets(value *LoggingMetricBucketOptionsExponentialBuckets) {
	if err := l.validatePutExponentialBucketsParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		l,
		"putExponentialBuckets",
		[]interface{}{value},
	)
}

func (l *jsiiProxy_LoggingMetricBucketOptionsOutputReference) PutLinearBuckets(value *LoggingMetricBucketOptionsLinearBuckets) {
	if err := l.validatePutLinearBucketsParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		l,
		"putLinearBuckets",
		[]interface{}{value},
	)
}

func (l *jsiiProxy_LoggingMetricBucketOptionsOutputReference) ResetExplicitBuckets() {
	_jsii_.InvokeVoid(
		l,
		"resetExplicitBuckets",
		nil, // no parameters
	)
}

func (l *jsiiProxy_LoggingMetricBucketOptionsOutputReference) ResetExponentialBuckets() {
	_jsii_.InvokeVoid(
		l,
		"resetExponentialBuckets",
		nil, // no parameters
	)
}

func (l *jsiiProxy_LoggingMetricBucketOptionsOutputReference) ResetLinearBuckets() {
	_jsii_.InvokeVoid(
		l,
		"resetLinearBuckets",
		nil, // no parameters
	)
}

func (l *jsiiProxy_LoggingMetricBucketOptionsOutputReference) Resolve(_context cdktf.IResolveContext) interface{} {
	if err := l.validateResolveParameters(_context); err != nil {
		panic(err)
	}
	var returns interface{}

	_jsii_.Invoke(
		l,
		"resolve",
		[]interface{}{_context},
		&returns,
	)

	return returns
}

func (l *jsiiProxy_LoggingMetricBucketOptionsOutputReference) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		l,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

