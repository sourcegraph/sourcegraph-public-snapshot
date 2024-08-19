package redisinstance

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/sourcegraph/managed-services-platform-cdktf/gen/google/jsii"

	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/redisinstance/internal"
)

type RedisInstancePersistenceConfigOutputReference interface {
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
	InternalValue() *RedisInstancePersistenceConfig
	SetInternalValue(val *RedisInstancePersistenceConfig)
	PersistenceMode() *string
	SetPersistenceMode(val *string)
	PersistenceModeInput() *string
	RdbNextSnapshotTime() *string
	RdbSnapshotPeriod() *string
	SetRdbSnapshotPeriod(val *string)
	RdbSnapshotPeriodInput() *string
	RdbSnapshotStartTime() *string
	SetRdbSnapshotStartTime(val *string)
	RdbSnapshotStartTimeInput() *string
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
	ResetPersistenceMode()
	ResetRdbSnapshotPeriod()
	ResetRdbSnapshotStartTime()
	// Produce the Token's value at resolution time.
	// Experimental.
	Resolve(_context cdktf.IResolveContext) interface{}
	// Return a string representation of this resolvable object.
	//
	// Returns a reversible string representation.
	// Experimental.
	ToString() *string
}

// The jsii proxy struct for RedisInstancePersistenceConfigOutputReference
type jsiiProxy_RedisInstancePersistenceConfigOutputReference struct {
	internal.Type__cdktfComplexObject
}

func (j *jsiiProxy_RedisInstancePersistenceConfigOutputReference) ComplexObjectIndex() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"complexObjectIndex",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstancePersistenceConfigOutputReference) ComplexObjectIsFromSet() *bool {
	var returns *bool
	_jsii_.Get(
		j,
		"complexObjectIsFromSet",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstancePersistenceConfigOutputReference) CreationStack() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"creationStack",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstancePersistenceConfigOutputReference) Fqn() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fqn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstancePersistenceConfigOutputReference) InternalValue() *RedisInstancePersistenceConfig {
	var returns *RedisInstancePersistenceConfig
	_jsii_.Get(
		j,
		"internalValue",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstancePersistenceConfigOutputReference) PersistenceMode() *string {
	var returns *string
	_jsii_.Get(
		j,
		"persistenceMode",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstancePersistenceConfigOutputReference) PersistenceModeInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"persistenceModeInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstancePersistenceConfigOutputReference) RdbNextSnapshotTime() *string {
	var returns *string
	_jsii_.Get(
		j,
		"rdbNextSnapshotTime",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstancePersistenceConfigOutputReference) RdbSnapshotPeriod() *string {
	var returns *string
	_jsii_.Get(
		j,
		"rdbSnapshotPeriod",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstancePersistenceConfigOutputReference) RdbSnapshotPeriodInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"rdbSnapshotPeriodInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstancePersistenceConfigOutputReference) RdbSnapshotStartTime() *string {
	var returns *string
	_jsii_.Get(
		j,
		"rdbSnapshotStartTime",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstancePersistenceConfigOutputReference) RdbSnapshotStartTimeInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"rdbSnapshotStartTimeInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstancePersistenceConfigOutputReference) TerraformAttribute() *string {
	var returns *string
	_jsii_.Get(
		j,
		"terraformAttribute",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstancePersistenceConfigOutputReference) TerraformResource() cdktf.IInterpolatingParent {
	var returns cdktf.IInterpolatingParent
	_jsii_.Get(
		j,
		"terraformResource",
		&returns,
	)
	return returns
}


func NewRedisInstancePersistenceConfigOutputReference(terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) RedisInstancePersistenceConfigOutputReference {
	_init_.Initialize()

	if err := validateNewRedisInstancePersistenceConfigOutputReferenceParameters(terraformResource, terraformAttribute); err != nil {
		panic(err)
	}
	j := jsiiProxy_RedisInstancePersistenceConfigOutputReference{}

	_jsii_.Create(
		"@cdktf/provider-google.redisInstance.RedisInstancePersistenceConfigOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		&j,
	)

	return &j
}

func NewRedisInstancePersistenceConfigOutputReference_Override(r RedisInstancePersistenceConfigOutputReference, terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) {
	_init_.Initialize()

	_jsii_.Create(
		"@cdktf/provider-google.redisInstance.RedisInstancePersistenceConfigOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		r,
	)
}

func (j *jsiiProxy_RedisInstancePersistenceConfigOutputReference)SetComplexObjectIndex(val interface{}) {
	if err := j.validateSetComplexObjectIndexParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIndex",
		val,
	)
}

func (j *jsiiProxy_RedisInstancePersistenceConfigOutputReference)SetComplexObjectIsFromSet(val *bool) {
	if err := j.validateSetComplexObjectIsFromSetParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIsFromSet",
		val,
	)
}

func (j *jsiiProxy_RedisInstancePersistenceConfigOutputReference)SetInternalValue(val *RedisInstancePersistenceConfig) {
	if err := j.validateSetInternalValueParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"internalValue",
		val,
	)
}

func (j *jsiiProxy_RedisInstancePersistenceConfigOutputReference)SetPersistenceMode(val *string) {
	if err := j.validateSetPersistenceModeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"persistenceMode",
		val,
	)
}

func (j *jsiiProxy_RedisInstancePersistenceConfigOutputReference)SetRdbSnapshotPeriod(val *string) {
	if err := j.validateSetRdbSnapshotPeriodParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"rdbSnapshotPeriod",
		val,
	)
}

func (j *jsiiProxy_RedisInstancePersistenceConfigOutputReference)SetRdbSnapshotStartTime(val *string) {
	if err := j.validateSetRdbSnapshotStartTimeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"rdbSnapshotStartTime",
		val,
	)
}

func (j *jsiiProxy_RedisInstancePersistenceConfigOutputReference)SetTerraformAttribute(val *string) {
	if err := j.validateSetTerraformAttributeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformAttribute",
		val,
	)
}

func (j *jsiiProxy_RedisInstancePersistenceConfigOutputReference)SetTerraformResource(val cdktf.IInterpolatingParent) {
	if err := j.validateSetTerraformResourceParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformResource",
		val,
	)
}

func (r *jsiiProxy_RedisInstancePersistenceConfigOutputReference) ComputeFqn() *string {
	var returns *string

	_jsii_.Invoke(
		r,
		"computeFqn",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (r *jsiiProxy_RedisInstancePersistenceConfigOutputReference) GetAnyMapAttribute(terraformAttribute *string) *map[string]interface{} {
	if err := r.validateGetAnyMapAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *map[string]interface{}

	_jsii_.Invoke(
		r,
		"getAnyMapAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (r *jsiiProxy_RedisInstancePersistenceConfigOutputReference) GetBooleanAttribute(terraformAttribute *string) cdktf.IResolvable {
	if err := r.validateGetBooleanAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		r,
		"getBooleanAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (r *jsiiProxy_RedisInstancePersistenceConfigOutputReference) GetBooleanMapAttribute(terraformAttribute *string) *map[string]*bool {
	if err := r.validateGetBooleanMapAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *map[string]*bool

	_jsii_.Invoke(
		r,
		"getBooleanMapAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (r *jsiiProxy_RedisInstancePersistenceConfigOutputReference) GetListAttribute(terraformAttribute *string) *[]*string {
	if err := r.validateGetListAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *[]*string

	_jsii_.Invoke(
		r,
		"getListAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (r *jsiiProxy_RedisInstancePersistenceConfigOutputReference) GetNumberAttribute(terraformAttribute *string) *float64 {
	if err := r.validateGetNumberAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *float64

	_jsii_.Invoke(
		r,
		"getNumberAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (r *jsiiProxy_RedisInstancePersistenceConfigOutputReference) GetNumberListAttribute(terraformAttribute *string) *[]*float64 {
	if err := r.validateGetNumberListAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *[]*float64

	_jsii_.Invoke(
		r,
		"getNumberListAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (r *jsiiProxy_RedisInstancePersistenceConfigOutputReference) GetNumberMapAttribute(terraformAttribute *string) *map[string]*float64 {
	if err := r.validateGetNumberMapAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *map[string]*float64

	_jsii_.Invoke(
		r,
		"getNumberMapAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (r *jsiiProxy_RedisInstancePersistenceConfigOutputReference) GetStringAttribute(terraformAttribute *string) *string {
	if err := r.validateGetStringAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.Invoke(
		r,
		"getStringAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (r *jsiiProxy_RedisInstancePersistenceConfigOutputReference) GetStringMapAttribute(terraformAttribute *string) *map[string]*string {
	if err := r.validateGetStringMapAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *map[string]*string

	_jsii_.Invoke(
		r,
		"getStringMapAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (r *jsiiProxy_RedisInstancePersistenceConfigOutputReference) InterpolationAsList() cdktf.IResolvable {
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		r,
		"interpolationAsList",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (r *jsiiProxy_RedisInstancePersistenceConfigOutputReference) InterpolationForAttribute(property *string) cdktf.IResolvable {
	if err := r.validateInterpolationForAttributeParameters(property); err != nil {
		panic(err)
	}
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		r,
		"interpolationForAttribute",
		[]interface{}{property},
		&returns,
	)

	return returns
}

func (r *jsiiProxy_RedisInstancePersistenceConfigOutputReference) ResetPersistenceMode() {
	_jsii_.InvokeVoid(
		r,
		"resetPersistenceMode",
		nil, // no parameters
	)
}

func (r *jsiiProxy_RedisInstancePersistenceConfigOutputReference) ResetRdbSnapshotPeriod() {
	_jsii_.InvokeVoid(
		r,
		"resetRdbSnapshotPeriod",
		nil, // no parameters
	)
}

func (r *jsiiProxy_RedisInstancePersistenceConfigOutputReference) ResetRdbSnapshotStartTime() {
	_jsii_.InvokeVoid(
		r,
		"resetRdbSnapshotStartTime",
		nil, // no parameters
	)
}

func (r *jsiiProxy_RedisInstancePersistenceConfigOutputReference) Resolve(_context cdktf.IResolveContext) interface{} {
	if err := r.validateResolveParameters(_context); err != nil {
		panic(err)
	}
	var returns interface{}

	_jsii_.Invoke(
		r,
		"resolve",
		[]interface{}{_context},
		&returns,
	)

	return returns
}

func (r *jsiiProxy_RedisInstancePersistenceConfigOutputReference) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		r,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

