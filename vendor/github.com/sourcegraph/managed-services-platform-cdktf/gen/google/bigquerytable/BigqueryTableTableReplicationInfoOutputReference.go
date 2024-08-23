package bigquerytable

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/sourcegraph/managed-services-platform-cdktf/gen/google/jsii"

	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/bigquerytable/internal"
)

type BigqueryTableTableReplicationInfoOutputReference interface {
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
	InternalValue() *BigqueryTableTableReplicationInfo
	SetInternalValue(val *BigqueryTableTableReplicationInfo)
	ReplicationIntervalMs() *float64
	SetReplicationIntervalMs(val *float64)
	ReplicationIntervalMsInput() *float64
	SourceDatasetId() *string
	SetSourceDatasetId(val *string)
	SourceDatasetIdInput() *string
	SourceProjectId() *string
	SetSourceProjectId(val *string)
	SourceProjectIdInput() *string
	SourceTableId() *string
	SetSourceTableId(val *string)
	SourceTableIdInput() *string
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
	ResetReplicationIntervalMs()
	// Produce the Token's value at resolution time.
	// Experimental.
	Resolve(_context cdktf.IResolveContext) interface{}
	// Return a string representation of this resolvable object.
	//
	// Returns a reversible string representation.
	// Experimental.
	ToString() *string
}

// The jsii proxy struct for BigqueryTableTableReplicationInfoOutputReference
type jsiiProxy_BigqueryTableTableReplicationInfoOutputReference struct {
	internal.Type__cdktfComplexObject
}

func (j *jsiiProxy_BigqueryTableTableReplicationInfoOutputReference) ComplexObjectIndex() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"complexObjectIndex",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableTableReplicationInfoOutputReference) ComplexObjectIsFromSet() *bool {
	var returns *bool
	_jsii_.Get(
		j,
		"complexObjectIsFromSet",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableTableReplicationInfoOutputReference) CreationStack() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"creationStack",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableTableReplicationInfoOutputReference) Fqn() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fqn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableTableReplicationInfoOutputReference) InternalValue() *BigqueryTableTableReplicationInfo {
	var returns *BigqueryTableTableReplicationInfo
	_jsii_.Get(
		j,
		"internalValue",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableTableReplicationInfoOutputReference) ReplicationIntervalMs() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"replicationIntervalMs",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableTableReplicationInfoOutputReference) ReplicationIntervalMsInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"replicationIntervalMsInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableTableReplicationInfoOutputReference) SourceDatasetId() *string {
	var returns *string
	_jsii_.Get(
		j,
		"sourceDatasetId",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableTableReplicationInfoOutputReference) SourceDatasetIdInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"sourceDatasetIdInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableTableReplicationInfoOutputReference) SourceProjectId() *string {
	var returns *string
	_jsii_.Get(
		j,
		"sourceProjectId",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableTableReplicationInfoOutputReference) SourceProjectIdInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"sourceProjectIdInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableTableReplicationInfoOutputReference) SourceTableId() *string {
	var returns *string
	_jsii_.Get(
		j,
		"sourceTableId",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableTableReplicationInfoOutputReference) SourceTableIdInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"sourceTableIdInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableTableReplicationInfoOutputReference) TerraformAttribute() *string {
	var returns *string
	_jsii_.Get(
		j,
		"terraformAttribute",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableTableReplicationInfoOutputReference) TerraformResource() cdktf.IInterpolatingParent {
	var returns cdktf.IInterpolatingParent
	_jsii_.Get(
		j,
		"terraformResource",
		&returns,
	)
	return returns
}


func NewBigqueryTableTableReplicationInfoOutputReference(terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) BigqueryTableTableReplicationInfoOutputReference {
	_init_.Initialize()

	if err := validateNewBigqueryTableTableReplicationInfoOutputReferenceParameters(terraformResource, terraformAttribute); err != nil {
		panic(err)
	}
	j := jsiiProxy_BigqueryTableTableReplicationInfoOutputReference{}

	_jsii_.Create(
		"@cdktf/provider-google.bigqueryTable.BigqueryTableTableReplicationInfoOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		&j,
	)

	return &j
}

func NewBigqueryTableTableReplicationInfoOutputReference_Override(b BigqueryTableTableReplicationInfoOutputReference, terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) {
	_init_.Initialize()

	_jsii_.Create(
		"@cdktf/provider-google.bigqueryTable.BigqueryTableTableReplicationInfoOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		b,
	)
}

func (j *jsiiProxy_BigqueryTableTableReplicationInfoOutputReference)SetComplexObjectIndex(val interface{}) {
	if err := j.validateSetComplexObjectIndexParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIndex",
		val,
	)
}

func (j *jsiiProxy_BigqueryTableTableReplicationInfoOutputReference)SetComplexObjectIsFromSet(val *bool) {
	if err := j.validateSetComplexObjectIsFromSetParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIsFromSet",
		val,
	)
}

func (j *jsiiProxy_BigqueryTableTableReplicationInfoOutputReference)SetInternalValue(val *BigqueryTableTableReplicationInfo) {
	if err := j.validateSetInternalValueParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"internalValue",
		val,
	)
}

func (j *jsiiProxy_BigqueryTableTableReplicationInfoOutputReference)SetReplicationIntervalMs(val *float64) {
	if err := j.validateSetReplicationIntervalMsParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"replicationIntervalMs",
		val,
	)
}

func (j *jsiiProxy_BigqueryTableTableReplicationInfoOutputReference)SetSourceDatasetId(val *string) {
	if err := j.validateSetSourceDatasetIdParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"sourceDatasetId",
		val,
	)
}

func (j *jsiiProxy_BigqueryTableTableReplicationInfoOutputReference)SetSourceProjectId(val *string) {
	if err := j.validateSetSourceProjectIdParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"sourceProjectId",
		val,
	)
}

func (j *jsiiProxy_BigqueryTableTableReplicationInfoOutputReference)SetSourceTableId(val *string) {
	if err := j.validateSetSourceTableIdParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"sourceTableId",
		val,
	)
}

func (j *jsiiProxy_BigqueryTableTableReplicationInfoOutputReference)SetTerraformAttribute(val *string) {
	if err := j.validateSetTerraformAttributeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformAttribute",
		val,
	)
}

func (j *jsiiProxy_BigqueryTableTableReplicationInfoOutputReference)SetTerraformResource(val cdktf.IInterpolatingParent) {
	if err := j.validateSetTerraformResourceParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformResource",
		val,
	)
}

func (b *jsiiProxy_BigqueryTableTableReplicationInfoOutputReference) ComputeFqn() *string {
	var returns *string

	_jsii_.Invoke(
		b,
		"computeFqn",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (b *jsiiProxy_BigqueryTableTableReplicationInfoOutputReference) GetAnyMapAttribute(terraformAttribute *string) *map[string]interface{} {
	if err := b.validateGetAnyMapAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *map[string]interface{}

	_jsii_.Invoke(
		b,
		"getAnyMapAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (b *jsiiProxy_BigqueryTableTableReplicationInfoOutputReference) GetBooleanAttribute(terraformAttribute *string) cdktf.IResolvable {
	if err := b.validateGetBooleanAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		b,
		"getBooleanAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (b *jsiiProxy_BigqueryTableTableReplicationInfoOutputReference) GetBooleanMapAttribute(terraformAttribute *string) *map[string]*bool {
	if err := b.validateGetBooleanMapAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *map[string]*bool

	_jsii_.Invoke(
		b,
		"getBooleanMapAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (b *jsiiProxy_BigqueryTableTableReplicationInfoOutputReference) GetListAttribute(terraformAttribute *string) *[]*string {
	if err := b.validateGetListAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *[]*string

	_jsii_.Invoke(
		b,
		"getListAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (b *jsiiProxy_BigqueryTableTableReplicationInfoOutputReference) GetNumberAttribute(terraformAttribute *string) *float64 {
	if err := b.validateGetNumberAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *float64

	_jsii_.Invoke(
		b,
		"getNumberAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (b *jsiiProxy_BigqueryTableTableReplicationInfoOutputReference) GetNumberListAttribute(terraformAttribute *string) *[]*float64 {
	if err := b.validateGetNumberListAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *[]*float64

	_jsii_.Invoke(
		b,
		"getNumberListAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (b *jsiiProxy_BigqueryTableTableReplicationInfoOutputReference) GetNumberMapAttribute(terraformAttribute *string) *map[string]*float64 {
	if err := b.validateGetNumberMapAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *map[string]*float64

	_jsii_.Invoke(
		b,
		"getNumberMapAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (b *jsiiProxy_BigqueryTableTableReplicationInfoOutputReference) GetStringAttribute(terraformAttribute *string) *string {
	if err := b.validateGetStringAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.Invoke(
		b,
		"getStringAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (b *jsiiProxy_BigqueryTableTableReplicationInfoOutputReference) GetStringMapAttribute(terraformAttribute *string) *map[string]*string {
	if err := b.validateGetStringMapAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *map[string]*string

	_jsii_.Invoke(
		b,
		"getStringMapAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (b *jsiiProxy_BigqueryTableTableReplicationInfoOutputReference) InterpolationAsList() cdktf.IResolvable {
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		b,
		"interpolationAsList",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (b *jsiiProxy_BigqueryTableTableReplicationInfoOutputReference) InterpolationForAttribute(property *string) cdktf.IResolvable {
	if err := b.validateInterpolationForAttributeParameters(property); err != nil {
		panic(err)
	}
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		b,
		"interpolationForAttribute",
		[]interface{}{property},
		&returns,
	)

	return returns
}

func (b *jsiiProxy_BigqueryTableTableReplicationInfoOutputReference) ResetReplicationIntervalMs() {
	_jsii_.InvokeVoid(
		b,
		"resetReplicationIntervalMs",
		nil, // no parameters
	)
}

func (b *jsiiProxy_BigqueryTableTableReplicationInfoOutputReference) Resolve(_context cdktf.IResolveContext) interface{} {
	if err := b.validateResolveParameters(_context); err != nil {
		panic(err)
	}
	var returns interface{}

	_jsii_.Invoke(
		b,
		"resolve",
		[]interface{}{_context},
		&returns,
	)

	return returns
}

func (b *jsiiProxy_BigqueryTableTableReplicationInfoOutputReference) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		b,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

