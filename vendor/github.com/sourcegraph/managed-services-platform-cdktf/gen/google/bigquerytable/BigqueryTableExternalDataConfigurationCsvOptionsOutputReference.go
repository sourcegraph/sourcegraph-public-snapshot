package bigquerytable

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/sourcegraph/managed-services-platform-cdktf/gen/google/jsii"

	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/bigquerytable/internal"
)

type BigqueryTableExternalDataConfigurationCsvOptionsOutputReference interface {
	cdktf.ComplexObject
	AllowJaggedRows() interface{}
	SetAllowJaggedRows(val interface{})
	AllowJaggedRowsInput() interface{}
	AllowQuotedNewlines() interface{}
	SetAllowQuotedNewlines(val interface{})
	AllowQuotedNewlinesInput() interface{}
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
	Encoding() *string
	SetEncoding(val *string)
	EncodingInput() *string
	FieldDelimiter() *string
	SetFieldDelimiter(val *string)
	FieldDelimiterInput() *string
	// Experimental.
	Fqn() *string
	InternalValue() *BigqueryTableExternalDataConfigurationCsvOptions
	SetInternalValue(val *BigqueryTableExternalDataConfigurationCsvOptions)
	Quote() *string
	SetQuote(val *string)
	QuoteInput() *string
	SkipLeadingRows() *float64
	SetSkipLeadingRows(val *float64)
	SkipLeadingRowsInput() *float64
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
	ResetAllowJaggedRows()
	ResetAllowQuotedNewlines()
	ResetEncoding()
	ResetFieldDelimiter()
	ResetSkipLeadingRows()
	// Produce the Token's value at resolution time.
	// Experimental.
	Resolve(_context cdktf.IResolveContext) interface{}
	// Return a string representation of this resolvable object.
	//
	// Returns a reversible string representation.
	// Experimental.
	ToString() *string
}

// The jsii proxy struct for BigqueryTableExternalDataConfigurationCsvOptionsOutputReference
type jsiiProxy_BigqueryTableExternalDataConfigurationCsvOptionsOutputReference struct {
	internal.Type__cdktfComplexObject
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationCsvOptionsOutputReference) AllowJaggedRows() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"allowJaggedRows",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationCsvOptionsOutputReference) AllowJaggedRowsInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"allowJaggedRowsInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationCsvOptionsOutputReference) AllowQuotedNewlines() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"allowQuotedNewlines",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationCsvOptionsOutputReference) AllowQuotedNewlinesInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"allowQuotedNewlinesInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationCsvOptionsOutputReference) ComplexObjectIndex() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"complexObjectIndex",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationCsvOptionsOutputReference) ComplexObjectIsFromSet() *bool {
	var returns *bool
	_jsii_.Get(
		j,
		"complexObjectIsFromSet",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationCsvOptionsOutputReference) CreationStack() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"creationStack",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationCsvOptionsOutputReference) Encoding() *string {
	var returns *string
	_jsii_.Get(
		j,
		"encoding",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationCsvOptionsOutputReference) EncodingInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"encodingInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationCsvOptionsOutputReference) FieldDelimiter() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fieldDelimiter",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationCsvOptionsOutputReference) FieldDelimiterInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fieldDelimiterInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationCsvOptionsOutputReference) Fqn() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fqn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationCsvOptionsOutputReference) InternalValue() *BigqueryTableExternalDataConfigurationCsvOptions {
	var returns *BigqueryTableExternalDataConfigurationCsvOptions
	_jsii_.Get(
		j,
		"internalValue",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationCsvOptionsOutputReference) Quote() *string {
	var returns *string
	_jsii_.Get(
		j,
		"quote",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationCsvOptionsOutputReference) QuoteInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"quoteInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationCsvOptionsOutputReference) SkipLeadingRows() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"skipLeadingRows",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationCsvOptionsOutputReference) SkipLeadingRowsInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"skipLeadingRowsInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationCsvOptionsOutputReference) TerraformAttribute() *string {
	var returns *string
	_jsii_.Get(
		j,
		"terraformAttribute",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationCsvOptionsOutputReference) TerraformResource() cdktf.IInterpolatingParent {
	var returns cdktf.IInterpolatingParent
	_jsii_.Get(
		j,
		"terraformResource",
		&returns,
	)
	return returns
}


func NewBigqueryTableExternalDataConfigurationCsvOptionsOutputReference(terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) BigqueryTableExternalDataConfigurationCsvOptionsOutputReference {
	_init_.Initialize()

	if err := validateNewBigqueryTableExternalDataConfigurationCsvOptionsOutputReferenceParameters(terraformResource, terraformAttribute); err != nil {
		panic(err)
	}
	j := jsiiProxy_BigqueryTableExternalDataConfigurationCsvOptionsOutputReference{}

	_jsii_.Create(
		"@cdktf/provider-google.bigqueryTable.BigqueryTableExternalDataConfigurationCsvOptionsOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		&j,
	)

	return &j
}

func NewBigqueryTableExternalDataConfigurationCsvOptionsOutputReference_Override(b BigqueryTableExternalDataConfigurationCsvOptionsOutputReference, terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) {
	_init_.Initialize()

	_jsii_.Create(
		"@cdktf/provider-google.bigqueryTable.BigqueryTableExternalDataConfigurationCsvOptionsOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		b,
	)
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationCsvOptionsOutputReference)SetAllowJaggedRows(val interface{}) {
	if err := j.validateSetAllowJaggedRowsParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"allowJaggedRows",
		val,
	)
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationCsvOptionsOutputReference)SetAllowQuotedNewlines(val interface{}) {
	if err := j.validateSetAllowQuotedNewlinesParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"allowQuotedNewlines",
		val,
	)
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationCsvOptionsOutputReference)SetComplexObjectIndex(val interface{}) {
	if err := j.validateSetComplexObjectIndexParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIndex",
		val,
	)
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationCsvOptionsOutputReference)SetComplexObjectIsFromSet(val *bool) {
	if err := j.validateSetComplexObjectIsFromSetParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIsFromSet",
		val,
	)
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationCsvOptionsOutputReference)SetEncoding(val *string) {
	if err := j.validateSetEncodingParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"encoding",
		val,
	)
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationCsvOptionsOutputReference)SetFieldDelimiter(val *string) {
	if err := j.validateSetFieldDelimiterParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"fieldDelimiter",
		val,
	)
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationCsvOptionsOutputReference)SetInternalValue(val *BigqueryTableExternalDataConfigurationCsvOptions) {
	if err := j.validateSetInternalValueParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"internalValue",
		val,
	)
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationCsvOptionsOutputReference)SetQuote(val *string) {
	if err := j.validateSetQuoteParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"quote",
		val,
	)
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationCsvOptionsOutputReference)SetSkipLeadingRows(val *float64) {
	if err := j.validateSetSkipLeadingRowsParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"skipLeadingRows",
		val,
	)
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationCsvOptionsOutputReference)SetTerraformAttribute(val *string) {
	if err := j.validateSetTerraformAttributeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformAttribute",
		val,
	)
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationCsvOptionsOutputReference)SetTerraformResource(val cdktf.IInterpolatingParent) {
	if err := j.validateSetTerraformResourceParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformResource",
		val,
	)
}

func (b *jsiiProxy_BigqueryTableExternalDataConfigurationCsvOptionsOutputReference) ComputeFqn() *string {
	var returns *string

	_jsii_.Invoke(
		b,
		"computeFqn",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (b *jsiiProxy_BigqueryTableExternalDataConfigurationCsvOptionsOutputReference) GetAnyMapAttribute(terraformAttribute *string) *map[string]interface{} {
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

func (b *jsiiProxy_BigqueryTableExternalDataConfigurationCsvOptionsOutputReference) GetBooleanAttribute(terraformAttribute *string) cdktf.IResolvable {
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

func (b *jsiiProxy_BigqueryTableExternalDataConfigurationCsvOptionsOutputReference) GetBooleanMapAttribute(terraformAttribute *string) *map[string]*bool {
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

func (b *jsiiProxy_BigqueryTableExternalDataConfigurationCsvOptionsOutputReference) GetListAttribute(terraformAttribute *string) *[]*string {
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

func (b *jsiiProxy_BigqueryTableExternalDataConfigurationCsvOptionsOutputReference) GetNumberAttribute(terraformAttribute *string) *float64 {
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

func (b *jsiiProxy_BigqueryTableExternalDataConfigurationCsvOptionsOutputReference) GetNumberListAttribute(terraformAttribute *string) *[]*float64 {
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

func (b *jsiiProxy_BigqueryTableExternalDataConfigurationCsvOptionsOutputReference) GetNumberMapAttribute(terraformAttribute *string) *map[string]*float64 {
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

func (b *jsiiProxy_BigqueryTableExternalDataConfigurationCsvOptionsOutputReference) GetStringAttribute(terraformAttribute *string) *string {
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

func (b *jsiiProxy_BigqueryTableExternalDataConfigurationCsvOptionsOutputReference) GetStringMapAttribute(terraformAttribute *string) *map[string]*string {
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

func (b *jsiiProxy_BigqueryTableExternalDataConfigurationCsvOptionsOutputReference) InterpolationAsList() cdktf.IResolvable {
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		b,
		"interpolationAsList",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (b *jsiiProxy_BigqueryTableExternalDataConfigurationCsvOptionsOutputReference) InterpolationForAttribute(property *string) cdktf.IResolvable {
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

func (b *jsiiProxy_BigqueryTableExternalDataConfigurationCsvOptionsOutputReference) ResetAllowJaggedRows() {
	_jsii_.InvokeVoid(
		b,
		"resetAllowJaggedRows",
		nil, // no parameters
	)
}

func (b *jsiiProxy_BigqueryTableExternalDataConfigurationCsvOptionsOutputReference) ResetAllowQuotedNewlines() {
	_jsii_.InvokeVoid(
		b,
		"resetAllowQuotedNewlines",
		nil, // no parameters
	)
}

func (b *jsiiProxy_BigqueryTableExternalDataConfigurationCsvOptionsOutputReference) ResetEncoding() {
	_jsii_.InvokeVoid(
		b,
		"resetEncoding",
		nil, // no parameters
	)
}

func (b *jsiiProxy_BigqueryTableExternalDataConfigurationCsvOptionsOutputReference) ResetFieldDelimiter() {
	_jsii_.InvokeVoid(
		b,
		"resetFieldDelimiter",
		nil, // no parameters
	)
}

func (b *jsiiProxy_BigqueryTableExternalDataConfigurationCsvOptionsOutputReference) ResetSkipLeadingRows() {
	_jsii_.InvokeVoid(
		b,
		"resetSkipLeadingRows",
		nil, // no parameters
	)
}

func (b *jsiiProxy_BigqueryTableExternalDataConfigurationCsvOptionsOutputReference) Resolve(_context cdktf.IResolveContext) interface{} {
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

func (b *jsiiProxy_BigqueryTableExternalDataConfigurationCsvOptionsOutputReference) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		b,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

