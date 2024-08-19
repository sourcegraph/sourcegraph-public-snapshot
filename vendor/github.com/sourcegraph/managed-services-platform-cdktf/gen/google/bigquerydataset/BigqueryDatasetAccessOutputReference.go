package bigquerydataset

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/sourcegraph/managed-services-platform-cdktf/gen/google/jsii"

	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/bigquerydataset/internal"
)

type BigqueryDatasetAccessOutputReference interface {
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
	Dataset() BigqueryDatasetAccessDatasetOutputReference
	DatasetInput() *BigqueryDatasetAccessDataset
	Domain() *string
	SetDomain(val *string)
	DomainInput() *string
	// Experimental.
	Fqn() *string
	GroupByEmail() *string
	SetGroupByEmail(val *string)
	GroupByEmailInput() *string
	IamMember() *string
	SetIamMember(val *string)
	IamMemberInput() *string
	InternalValue() interface{}
	SetInternalValue(val interface{})
	Role() *string
	SetRole(val *string)
	RoleInput() *string
	Routine() BigqueryDatasetAccessRoutineOutputReference
	RoutineInput() *BigqueryDatasetAccessRoutine
	SpecialGroup() *string
	SetSpecialGroup(val *string)
	SpecialGroupInput() *string
	// Experimental.
	TerraformAttribute() *string
	// Experimental.
	SetTerraformAttribute(val *string)
	// Experimental.
	TerraformResource() cdktf.IInterpolatingParent
	// Experimental.
	SetTerraformResource(val cdktf.IInterpolatingParent)
	UserByEmail() *string
	SetUserByEmail(val *string)
	UserByEmailInput() *string
	View() BigqueryDatasetAccessViewOutputReference
	ViewInput() *BigqueryDatasetAccessView
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
	PutDataset(value *BigqueryDatasetAccessDataset)
	PutRoutine(value *BigqueryDatasetAccessRoutine)
	PutView(value *BigqueryDatasetAccessView)
	ResetDataset()
	ResetDomain()
	ResetGroupByEmail()
	ResetIamMember()
	ResetRole()
	ResetRoutine()
	ResetSpecialGroup()
	ResetUserByEmail()
	ResetView()
	// Produce the Token's value at resolution time.
	// Experimental.
	Resolve(_context cdktf.IResolveContext) interface{}
	// Return a string representation of this resolvable object.
	//
	// Returns a reversible string representation.
	// Experimental.
	ToString() *string
}

// The jsii proxy struct for BigqueryDatasetAccessOutputReference
type jsiiProxy_BigqueryDatasetAccessOutputReference struct {
	internal.Type__cdktfComplexObject
}

func (j *jsiiProxy_BigqueryDatasetAccessOutputReference) ComplexObjectIndex() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"complexObjectIndex",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryDatasetAccessOutputReference) ComplexObjectIsFromSet() *bool {
	var returns *bool
	_jsii_.Get(
		j,
		"complexObjectIsFromSet",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryDatasetAccessOutputReference) CreationStack() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"creationStack",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryDatasetAccessOutputReference) Dataset() BigqueryDatasetAccessDatasetOutputReference {
	var returns BigqueryDatasetAccessDatasetOutputReference
	_jsii_.Get(
		j,
		"dataset",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryDatasetAccessOutputReference) DatasetInput() *BigqueryDatasetAccessDataset {
	var returns *BigqueryDatasetAccessDataset
	_jsii_.Get(
		j,
		"datasetInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryDatasetAccessOutputReference) Domain() *string {
	var returns *string
	_jsii_.Get(
		j,
		"domain",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryDatasetAccessOutputReference) DomainInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"domainInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryDatasetAccessOutputReference) Fqn() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fqn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryDatasetAccessOutputReference) GroupByEmail() *string {
	var returns *string
	_jsii_.Get(
		j,
		"groupByEmail",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryDatasetAccessOutputReference) GroupByEmailInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"groupByEmailInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryDatasetAccessOutputReference) IamMember() *string {
	var returns *string
	_jsii_.Get(
		j,
		"iamMember",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryDatasetAccessOutputReference) IamMemberInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"iamMemberInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryDatasetAccessOutputReference) InternalValue() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"internalValue",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryDatasetAccessOutputReference) Role() *string {
	var returns *string
	_jsii_.Get(
		j,
		"role",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryDatasetAccessOutputReference) RoleInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"roleInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryDatasetAccessOutputReference) Routine() BigqueryDatasetAccessRoutineOutputReference {
	var returns BigqueryDatasetAccessRoutineOutputReference
	_jsii_.Get(
		j,
		"routine",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryDatasetAccessOutputReference) RoutineInput() *BigqueryDatasetAccessRoutine {
	var returns *BigqueryDatasetAccessRoutine
	_jsii_.Get(
		j,
		"routineInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryDatasetAccessOutputReference) SpecialGroup() *string {
	var returns *string
	_jsii_.Get(
		j,
		"specialGroup",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryDatasetAccessOutputReference) SpecialGroupInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"specialGroupInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryDatasetAccessOutputReference) TerraformAttribute() *string {
	var returns *string
	_jsii_.Get(
		j,
		"terraformAttribute",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryDatasetAccessOutputReference) TerraformResource() cdktf.IInterpolatingParent {
	var returns cdktf.IInterpolatingParent
	_jsii_.Get(
		j,
		"terraformResource",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryDatasetAccessOutputReference) UserByEmail() *string {
	var returns *string
	_jsii_.Get(
		j,
		"userByEmail",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryDatasetAccessOutputReference) UserByEmailInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"userByEmailInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryDatasetAccessOutputReference) View() BigqueryDatasetAccessViewOutputReference {
	var returns BigqueryDatasetAccessViewOutputReference
	_jsii_.Get(
		j,
		"view",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryDatasetAccessOutputReference) ViewInput() *BigqueryDatasetAccessView {
	var returns *BigqueryDatasetAccessView
	_jsii_.Get(
		j,
		"viewInput",
		&returns,
	)
	return returns
}


func NewBigqueryDatasetAccessOutputReference(terraformResource cdktf.IInterpolatingParent, terraformAttribute *string, complexObjectIndex *float64, complexObjectIsFromSet *bool) BigqueryDatasetAccessOutputReference {
	_init_.Initialize()

	if err := validateNewBigqueryDatasetAccessOutputReferenceParameters(terraformResource, terraformAttribute, complexObjectIndex, complexObjectIsFromSet); err != nil {
		panic(err)
	}
	j := jsiiProxy_BigqueryDatasetAccessOutputReference{}

	_jsii_.Create(
		"@cdktf/provider-google.bigqueryDataset.BigqueryDatasetAccessOutputReference",
		[]interface{}{terraformResource, terraformAttribute, complexObjectIndex, complexObjectIsFromSet},
		&j,
	)

	return &j
}

func NewBigqueryDatasetAccessOutputReference_Override(b BigqueryDatasetAccessOutputReference, terraformResource cdktf.IInterpolatingParent, terraformAttribute *string, complexObjectIndex *float64, complexObjectIsFromSet *bool) {
	_init_.Initialize()

	_jsii_.Create(
		"@cdktf/provider-google.bigqueryDataset.BigqueryDatasetAccessOutputReference",
		[]interface{}{terraformResource, terraformAttribute, complexObjectIndex, complexObjectIsFromSet},
		b,
	)
}

func (j *jsiiProxy_BigqueryDatasetAccessOutputReference)SetComplexObjectIndex(val interface{}) {
	if err := j.validateSetComplexObjectIndexParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIndex",
		val,
	)
}

func (j *jsiiProxy_BigqueryDatasetAccessOutputReference)SetComplexObjectIsFromSet(val *bool) {
	if err := j.validateSetComplexObjectIsFromSetParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIsFromSet",
		val,
	)
}

func (j *jsiiProxy_BigqueryDatasetAccessOutputReference)SetDomain(val *string) {
	if err := j.validateSetDomainParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"domain",
		val,
	)
}

func (j *jsiiProxy_BigqueryDatasetAccessOutputReference)SetGroupByEmail(val *string) {
	if err := j.validateSetGroupByEmailParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"groupByEmail",
		val,
	)
}

func (j *jsiiProxy_BigqueryDatasetAccessOutputReference)SetIamMember(val *string) {
	if err := j.validateSetIamMemberParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"iamMember",
		val,
	)
}

func (j *jsiiProxy_BigqueryDatasetAccessOutputReference)SetInternalValue(val interface{}) {
	if err := j.validateSetInternalValueParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"internalValue",
		val,
	)
}

func (j *jsiiProxy_BigqueryDatasetAccessOutputReference)SetRole(val *string) {
	if err := j.validateSetRoleParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"role",
		val,
	)
}

func (j *jsiiProxy_BigqueryDatasetAccessOutputReference)SetSpecialGroup(val *string) {
	if err := j.validateSetSpecialGroupParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"specialGroup",
		val,
	)
}

func (j *jsiiProxy_BigqueryDatasetAccessOutputReference)SetTerraformAttribute(val *string) {
	if err := j.validateSetTerraformAttributeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformAttribute",
		val,
	)
}

func (j *jsiiProxy_BigqueryDatasetAccessOutputReference)SetTerraformResource(val cdktf.IInterpolatingParent) {
	if err := j.validateSetTerraformResourceParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformResource",
		val,
	)
}

func (j *jsiiProxy_BigqueryDatasetAccessOutputReference)SetUserByEmail(val *string) {
	if err := j.validateSetUserByEmailParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"userByEmail",
		val,
	)
}

func (b *jsiiProxy_BigqueryDatasetAccessOutputReference) ComputeFqn() *string {
	var returns *string

	_jsii_.Invoke(
		b,
		"computeFqn",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (b *jsiiProxy_BigqueryDatasetAccessOutputReference) GetAnyMapAttribute(terraformAttribute *string) *map[string]interface{} {
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

func (b *jsiiProxy_BigqueryDatasetAccessOutputReference) GetBooleanAttribute(terraformAttribute *string) cdktf.IResolvable {
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

func (b *jsiiProxy_BigqueryDatasetAccessOutputReference) GetBooleanMapAttribute(terraformAttribute *string) *map[string]*bool {
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

func (b *jsiiProxy_BigqueryDatasetAccessOutputReference) GetListAttribute(terraformAttribute *string) *[]*string {
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

func (b *jsiiProxy_BigqueryDatasetAccessOutputReference) GetNumberAttribute(terraformAttribute *string) *float64 {
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

func (b *jsiiProxy_BigqueryDatasetAccessOutputReference) GetNumberListAttribute(terraformAttribute *string) *[]*float64 {
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

func (b *jsiiProxy_BigqueryDatasetAccessOutputReference) GetNumberMapAttribute(terraformAttribute *string) *map[string]*float64 {
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

func (b *jsiiProxy_BigqueryDatasetAccessOutputReference) GetStringAttribute(terraformAttribute *string) *string {
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

func (b *jsiiProxy_BigqueryDatasetAccessOutputReference) GetStringMapAttribute(terraformAttribute *string) *map[string]*string {
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

func (b *jsiiProxy_BigqueryDatasetAccessOutputReference) InterpolationAsList() cdktf.IResolvable {
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		b,
		"interpolationAsList",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (b *jsiiProxy_BigqueryDatasetAccessOutputReference) InterpolationForAttribute(property *string) cdktf.IResolvable {
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

func (b *jsiiProxy_BigqueryDatasetAccessOutputReference) PutDataset(value *BigqueryDatasetAccessDataset) {
	if err := b.validatePutDatasetParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		b,
		"putDataset",
		[]interface{}{value},
	)
}

func (b *jsiiProxy_BigqueryDatasetAccessOutputReference) PutRoutine(value *BigqueryDatasetAccessRoutine) {
	if err := b.validatePutRoutineParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		b,
		"putRoutine",
		[]interface{}{value},
	)
}

func (b *jsiiProxy_BigqueryDatasetAccessOutputReference) PutView(value *BigqueryDatasetAccessView) {
	if err := b.validatePutViewParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		b,
		"putView",
		[]interface{}{value},
	)
}

func (b *jsiiProxy_BigqueryDatasetAccessOutputReference) ResetDataset() {
	_jsii_.InvokeVoid(
		b,
		"resetDataset",
		nil, // no parameters
	)
}

func (b *jsiiProxy_BigqueryDatasetAccessOutputReference) ResetDomain() {
	_jsii_.InvokeVoid(
		b,
		"resetDomain",
		nil, // no parameters
	)
}

func (b *jsiiProxy_BigqueryDatasetAccessOutputReference) ResetGroupByEmail() {
	_jsii_.InvokeVoid(
		b,
		"resetGroupByEmail",
		nil, // no parameters
	)
}

func (b *jsiiProxy_BigqueryDatasetAccessOutputReference) ResetIamMember() {
	_jsii_.InvokeVoid(
		b,
		"resetIamMember",
		nil, // no parameters
	)
}

func (b *jsiiProxy_BigqueryDatasetAccessOutputReference) ResetRole() {
	_jsii_.InvokeVoid(
		b,
		"resetRole",
		nil, // no parameters
	)
}

func (b *jsiiProxy_BigqueryDatasetAccessOutputReference) ResetRoutine() {
	_jsii_.InvokeVoid(
		b,
		"resetRoutine",
		nil, // no parameters
	)
}

func (b *jsiiProxy_BigqueryDatasetAccessOutputReference) ResetSpecialGroup() {
	_jsii_.InvokeVoid(
		b,
		"resetSpecialGroup",
		nil, // no parameters
	)
}

func (b *jsiiProxy_BigqueryDatasetAccessOutputReference) ResetUserByEmail() {
	_jsii_.InvokeVoid(
		b,
		"resetUserByEmail",
		nil, // no parameters
	)
}

func (b *jsiiProxy_BigqueryDatasetAccessOutputReference) ResetView() {
	_jsii_.InvokeVoid(
		b,
		"resetView",
		nil, // no parameters
	)
}

func (b *jsiiProxy_BigqueryDatasetAccessOutputReference) Resolve(_context cdktf.IResolveContext) interface{} {
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

func (b *jsiiProxy_BigqueryDatasetAccessOutputReference) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		b,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

