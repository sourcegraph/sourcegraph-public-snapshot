package bigquerytable

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/sourcegraph/managed-services-platform-cdktf/gen/google/jsii"

	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/bigquerytable/internal"
)

type BigqueryTableExternalDataConfigurationOutputReference interface {
	cdktf.ComplexObject
	Autodetect() interface{}
	SetAutodetect(val interface{})
	AutodetectInput() interface{}
	AvroOptions() BigqueryTableExternalDataConfigurationAvroOptionsOutputReference
	AvroOptionsInput() *BigqueryTableExternalDataConfigurationAvroOptions
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
	Compression() *string
	SetCompression(val *string)
	CompressionInput() *string
	ConnectionId() *string
	SetConnectionId(val *string)
	ConnectionIdInput() *string
	// The creation stack of this resolvable which will be appended to errors thrown during resolution.
	//
	// If this returns an empty array the stack will not be attached.
	// Experimental.
	CreationStack() *[]*string
	CsvOptions() BigqueryTableExternalDataConfigurationCsvOptionsOutputReference
	CsvOptionsInput() *BigqueryTableExternalDataConfigurationCsvOptions
	FileSetSpecType() *string
	SetFileSetSpecType(val *string)
	FileSetSpecTypeInput() *string
	// Experimental.
	Fqn() *string
	GoogleSheetsOptions() BigqueryTableExternalDataConfigurationGoogleSheetsOptionsOutputReference
	GoogleSheetsOptionsInput() *BigqueryTableExternalDataConfigurationGoogleSheetsOptions
	HivePartitioningOptions() BigqueryTableExternalDataConfigurationHivePartitioningOptionsOutputReference
	HivePartitioningOptionsInput() *BigqueryTableExternalDataConfigurationHivePartitioningOptions
	IgnoreUnknownValues() interface{}
	SetIgnoreUnknownValues(val interface{})
	IgnoreUnknownValuesInput() interface{}
	InternalValue() *BigqueryTableExternalDataConfiguration
	SetInternalValue(val *BigqueryTableExternalDataConfiguration)
	JsonExtension() *string
	SetJsonExtension(val *string)
	JsonExtensionInput() *string
	JsonOptions() BigqueryTableExternalDataConfigurationJsonOptionsOutputReference
	JsonOptionsInput() *BigqueryTableExternalDataConfigurationJsonOptions
	MaxBadRecords() *float64
	SetMaxBadRecords(val *float64)
	MaxBadRecordsInput() *float64
	MetadataCacheMode() *string
	SetMetadataCacheMode(val *string)
	MetadataCacheModeInput() *string
	ObjectMetadata() *string
	SetObjectMetadata(val *string)
	ObjectMetadataInput() *string
	ParquetOptions() BigqueryTableExternalDataConfigurationParquetOptionsOutputReference
	ParquetOptionsInput() *BigqueryTableExternalDataConfigurationParquetOptions
	ReferenceFileSchemaUri() *string
	SetReferenceFileSchemaUri(val *string)
	ReferenceFileSchemaUriInput() *string
	Schema() *string
	SetSchema(val *string)
	SchemaInput() *string
	SourceFormat() *string
	SetSourceFormat(val *string)
	SourceFormatInput() *string
	SourceUris() *[]*string
	SetSourceUris(val *[]*string)
	SourceUrisInput() *[]*string
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
	PutAvroOptions(value *BigqueryTableExternalDataConfigurationAvroOptions)
	PutCsvOptions(value *BigqueryTableExternalDataConfigurationCsvOptions)
	PutGoogleSheetsOptions(value *BigqueryTableExternalDataConfigurationGoogleSheetsOptions)
	PutHivePartitioningOptions(value *BigqueryTableExternalDataConfigurationHivePartitioningOptions)
	PutJsonOptions(value *BigqueryTableExternalDataConfigurationJsonOptions)
	PutParquetOptions(value *BigqueryTableExternalDataConfigurationParquetOptions)
	ResetAvroOptions()
	ResetCompression()
	ResetConnectionId()
	ResetCsvOptions()
	ResetFileSetSpecType()
	ResetGoogleSheetsOptions()
	ResetHivePartitioningOptions()
	ResetIgnoreUnknownValues()
	ResetJsonExtension()
	ResetJsonOptions()
	ResetMaxBadRecords()
	ResetMetadataCacheMode()
	ResetObjectMetadata()
	ResetParquetOptions()
	ResetReferenceFileSchemaUri()
	ResetSchema()
	ResetSourceFormat()
	// Produce the Token's value at resolution time.
	// Experimental.
	Resolve(_context cdktf.IResolveContext) interface{}
	// Return a string representation of this resolvable object.
	//
	// Returns a reversible string representation.
	// Experimental.
	ToString() *string
}

// The jsii proxy struct for BigqueryTableExternalDataConfigurationOutputReference
type jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference struct {
	internal.Type__cdktfComplexObject
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) Autodetect() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"autodetect",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) AutodetectInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"autodetectInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) AvroOptions() BigqueryTableExternalDataConfigurationAvroOptionsOutputReference {
	var returns BigqueryTableExternalDataConfigurationAvroOptionsOutputReference
	_jsii_.Get(
		j,
		"avroOptions",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) AvroOptionsInput() *BigqueryTableExternalDataConfigurationAvroOptions {
	var returns *BigqueryTableExternalDataConfigurationAvroOptions
	_jsii_.Get(
		j,
		"avroOptionsInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) ComplexObjectIndex() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"complexObjectIndex",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) ComplexObjectIsFromSet() *bool {
	var returns *bool
	_jsii_.Get(
		j,
		"complexObjectIsFromSet",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) Compression() *string {
	var returns *string
	_jsii_.Get(
		j,
		"compression",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) CompressionInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"compressionInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) ConnectionId() *string {
	var returns *string
	_jsii_.Get(
		j,
		"connectionId",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) ConnectionIdInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"connectionIdInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) CreationStack() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"creationStack",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) CsvOptions() BigqueryTableExternalDataConfigurationCsvOptionsOutputReference {
	var returns BigqueryTableExternalDataConfigurationCsvOptionsOutputReference
	_jsii_.Get(
		j,
		"csvOptions",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) CsvOptionsInput() *BigqueryTableExternalDataConfigurationCsvOptions {
	var returns *BigqueryTableExternalDataConfigurationCsvOptions
	_jsii_.Get(
		j,
		"csvOptionsInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) FileSetSpecType() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fileSetSpecType",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) FileSetSpecTypeInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fileSetSpecTypeInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) Fqn() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fqn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) GoogleSheetsOptions() BigqueryTableExternalDataConfigurationGoogleSheetsOptionsOutputReference {
	var returns BigqueryTableExternalDataConfigurationGoogleSheetsOptionsOutputReference
	_jsii_.Get(
		j,
		"googleSheetsOptions",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) GoogleSheetsOptionsInput() *BigqueryTableExternalDataConfigurationGoogleSheetsOptions {
	var returns *BigqueryTableExternalDataConfigurationGoogleSheetsOptions
	_jsii_.Get(
		j,
		"googleSheetsOptionsInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) HivePartitioningOptions() BigqueryTableExternalDataConfigurationHivePartitioningOptionsOutputReference {
	var returns BigqueryTableExternalDataConfigurationHivePartitioningOptionsOutputReference
	_jsii_.Get(
		j,
		"hivePartitioningOptions",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) HivePartitioningOptionsInput() *BigqueryTableExternalDataConfigurationHivePartitioningOptions {
	var returns *BigqueryTableExternalDataConfigurationHivePartitioningOptions
	_jsii_.Get(
		j,
		"hivePartitioningOptionsInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) IgnoreUnknownValues() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"ignoreUnknownValues",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) IgnoreUnknownValuesInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"ignoreUnknownValuesInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) InternalValue() *BigqueryTableExternalDataConfiguration {
	var returns *BigqueryTableExternalDataConfiguration
	_jsii_.Get(
		j,
		"internalValue",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) JsonExtension() *string {
	var returns *string
	_jsii_.Get(
		j,
		"jsonExtension",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) JsonExtensionInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"jsonExtensionInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) JsonOptions() BigqueryTableExternalDataConfigurationJsonOptionsOutputReference {
	var returns BigqueryTableExternalDataConfigurationJsonOptionsOutputReference
	_jsii_.Get(
		j,
		"jsonOptions",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) JsonOptionsInput() *BigqueryTableExternalDataConfigurationJsonOptions {
	var returns *BigqueryTableExternalDataConfigurationJsonOptions
	_jsii_.Get(
		j,
		"jsonOptionsInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) MaxBadRecords() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"maxBadRecords",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) MaxBadRecordsInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"maxBadRecordsInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) MetadataCacheMode() *string {
	var returns *string
	_jsii_.Get(
		j,
		"metadataCacheMode",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) MetadataCacheModeInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"metadataCacheModeInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) ObjectMetadata() *string {
	var returns *string
	_jsii_.Get(
		j,
		"objectMetadata",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) ObjectMetadataInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"objectMetadataInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) ParquetOptions() BigqueryTableExternalDataConfigurationParquetOptionsOutputReference {
	var returns BigqueryTableExternalDataConfigurationParquetOptionsOutputReference
	_jsii_.Get(
		j,
		"parquetOptions",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) ParquetOptionsInput() *BigqueryTableExternalDataConfigurationParquetOptions {
	var returns *BigqueryTableExternalDataConfigurationParquetOptions
	_jsii_.Get(
		j,
		"parquetOptionsInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) ReferenceFileSchemaUri() *string {
	var returns *string
	_jsii_.Get(
		j,
		"referenceFileSchemaUri",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) ReferenceFileSchemaUriInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"referenceFileSchemaUriInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) Schema() *string {
	var returns *string
	_jsii_.Get(
		j,
		"schema",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) SchemaInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"schemaInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) SourceFormat() *string {
	var returns *string
	_jsii_.Get(
		j,
		"sourceFormat",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) SourceFormatInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"sourceFormatInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) SourceUris() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"sourceUris",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) SourceUrisInput() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"sourceUrisInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) TerraformAttribute() *string {
	var returns *string
	_jsii_.Get(
		j,
		"terraformAttribute",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) TerraformResource() cdktf.IInterpolatingParent {
	var returns cdktf.IInterpolatingParent
	_jsii_.Get(
		j,
		"terraformResource",
		&returns,
	)
	return returns
}


func NewBigqueryTableExternalDataConfigurationOutputReference(terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) BigqueryTableExternalDataConfigurationOutputReference {
	_init_.Initialize()

	if err := validateNewBigqueryTableExternalDataConfigurationOutputReferenceParameters(terraformResource, terraformAttribute); err != nil {
		panic(err)
	}
	j := jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference{}

	_jsii_.Create(
		"@cdktf/provider-google.bigqueryTable.BigqueryTableExternalDataConfigurationOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		&j,
	)

	return &j
}

func NewBigqueryTableExternalDataConfigurationOutputReference_Override(b BigqueryTableExternalDataConfigurationOutputReference, terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) {
	_init_.Initialize()

	_jsii_.Create(
		"@cdktf/provider-google.bigqueryTable.BigqueryTableExternalDataConfigurationOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		b,
	)
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference)SetAutodetect(val interface{}) {
	if err := j.validateSetAutodetectParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"autodetect",
		val,
	)
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference)SetComplexObjectIndex(val interface{}) {
	if err := j.validateSetComplexObjectIndexParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIndex",
		val,
	)
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference)SetComplexObjectIsFromSet(val *bool) {
	if err := j.validateSetComplexObjectIsFromSetParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIsFromSet",
		val,
	)
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference)SetCompression(val *string) {
	if err := j.validateSetCompressionParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"compression",
		val,
	)
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference)SetConnectionId(val *string) {
	if err := j.validateSetConnectionIdParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"connectionId",
		val,
	)
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference)SetFileSetSpecType(val *string) {
	if err := j.validateSetFileSetSpecTypeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"fileSetSpecType",
		val,
	)
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference)SetIgnoreUnknownValues(val interface{}) {
	if err := j.validateSetIgnoreUnknownValuesParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"ignoreUnknownValues",
		val,
	)
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference)SetInternalValue(val *BigqueryTableExternalDataConfiguration) {
	if err := j.validateSetInternalValueParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"internalValue",
		val,
	)
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference)SetJsonExtension(val *string) {
	if err := j.validateSetJsonExtensionParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"jsonExtension",
		val,
	)
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference)SetMaxBadRecords(val *float64) {
	if err := j.validateSetMaxBadRecordsParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"maxBadRecords",
		val,
	)
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference)SetMetadataCacheMode(val *string) {
	if err := j.validateSetMetadataCacheModeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"metadataCacheMode",
		val,
	)
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference)SetObjectMetadata(val *string) {
	if err := j.validateSetObjectMetadataParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"objectMetadata",
		val,
	)
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference)SetReferenceFileSchemaUri(val *string) {
	if err := j.validateSetReferenceFileSchemaUriParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"referenceFileSchemaUri",
		val,
	)
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference)SetSchema(val *string) {
	if err := j.validateSetSchemaParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"schema",
		val,
	)
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference)SetSourceFormat(val *string) {
	if err := j.validateSetSourceFormatParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"sourceFormat",
		val,
	)
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference)SetSourceUris(val *[]*string) {
	if err := j.validateSetSourceUrisParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"sourceUris",
		val,
	)
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference)SetTerraformAttribute(val *string) {
	if err := j.validateSetTerraformAttributeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformAttribute",
		val,
	)
}

func (j *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference)SetTerraformResource(val cdktf.IInterpolatingParent) {
	if err := j.validateSetTerraformResourceParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformResource",
		val,
	)
}

func (b *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) ComputeFqn() *string {
	var returns *string

	_jsii_.Invoke(
		b,
		"computeFqn",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (b *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) GetAnyMapAttribute(terraformAttribute *string) *map[string]interface{} {
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

func (b *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) GetBooleanAttribute(terraformAttribute *string) cdktf.IResolvable {
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

func (b *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) GetBooleanMapAttribute(terraformAttribute *string) *map[string]*bool {
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

func (b *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) GetListAttribute(terraformAttribute *string) *[]*string {
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

func (b *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) GetNumberAttribute(terraformAttribute *string) *float64 {
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

func (b *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) GetNumberListAttribute(terraformAttribute *string) *[]*float64 {
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

func (b *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) GetNumberMapAttribute(terraformAttribute *string) *map[string]*float64 {
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

func (b *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) GetStringAttribute(terraformAttribute *string) *string {
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

func (b *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) GetStringMapAttribute(terraformAttribute *string) *map[string]*string {
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

func (b *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) InterpolationAsList() cdktf.IResolvable {
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		b,
		"interpolationAsList",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (b *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) InterpolationForAttribute(property *string) cdktf.IResolvable {
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

func (b *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) PutAvroOptions(value *BigqueryTableExternalDataConfigurationAvroOptions) {
	if err := b.validatePutAvroOptionsParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		b,
		"putAvroOptions",
		[]interface{}{value},
	)
}

func (b *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) PutCsvOptions(value *BigqueryTableExternalDataConfigurationCsvOptions) {
	if err := b.validatePutCsvOptionsParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		b,
		"putCsvOptions",
		[]interface{}{value},
	)
}

func (b *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) PutGoogleSheetsOptions(value *BigqueryTableExternalDataConfigurationGoogleSheetsOptions) {
	if err := b.validatePutGoogleSheetsOptionsParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		b,
		"putGoogleSheetsOptions",
		[]interface{}{value},
	)
}

func (b *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) PutHivePartitioningOptions(value *BigqueryTableExternalDataConfigurationHivePartitioningOptions) {
	if err := b.validatePutHivePartitioningOptionsParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		b,
		"putHivePartitioningOptions",
		[]interface{}{value},
	)
}

func (b *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) PutJsonOptions(value *BigqueryTableExternalDataConfigurationJsonOptions) {
	if err := b.validatePutJsonOptionsParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		b,
		"putJsonOptions",
		[]interface{}{value},
	)
}

func (b *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) PutParquetOptions(value *BigqueryTableExternalDataConfigurationParquetOptions) {
	if err := b.validatePutParquetOptionsParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		b,
		"putParquetOptions",
		[]interface{}{value},
	)
}

func (b *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) ResetAvroOptions() {
	_jsii_.InvokeVoid(
		b,
		"resetAvroOptions",
		nil, // no parameters
	)
}

func (b *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) ResetCompression() {
	_jsii_.InvokeVoid(
		b,
		"resetCompression",
		nil, // no parameters
	)
}

func (b *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) ResetConnectionId() {
	_jsii_.InvokeVoid(
		b,
		"resetConnectionId",
		nil, // no parameters
	)
}

func (b *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) ResetCsvOptions() {
	_jsii_.InvokeVoid(
		b,
		"resetCsvOptions",
		nil, // no parameters
	)
}

func (b *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) ResetFileSetSpecType() {
	_jsii_.InvokeVoid(
		b,
		"resetFileSetSpecType",
		nil, // no parameters
	)
}

func (b *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) ResetGoogleSheetsOptions() {
	_jsii_.InvokeVoid(
		b,
		"resetGoogleSheetsOptions",
		nil, // no parameters
	)
}

func (b *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) ResetHivePartitioningOptions() {
	_jsii_.InvokeVoid(
		b,
		"resetHivePartitioningOptions",
		nil, // no parameters
	)
}

func (b *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) ResetIgnoreUnknownValues() {
	_jsii_.InvokeVoid(
		b,
		"resetIgnoreUnknownValues",
		nil, // no parameters
	)
}

func (b *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) ResetJsonExtension() {
	_jsii_.InvokeVoid(
		b,
		"resetJsonExtension",
		nil, // no parameters
	)
}

func (b *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) ResetJsonOptions() {
	_jsii_.InvokeVoid(
		b,
		"resetJsonOptions",
		nil, // no parameters
	)
}

func (b *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) ResetMaxBadRecords() {
	_jsii_.InvokeVoid(
		b,
		"resetMaxBadRecords",
		nil, // no parameters
	)
}

func (b *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) ResetMetadataCacheMode() {
	_jsii_.InvokeVoid(
		b,
		"resetMetadataCacheMode",
		nil, // no parameters
	)
}

func (b *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) ResetObjectMetadata() {
	_jsii_.InvokeVoid(
		b,
		"resetObjectMetadata",
		nil, // no parameters
	)
}

func (b *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) ResetParquetOptions() {
	_jsii_.InvokeVoid(
		b,
		"resetParquetOptions",
		nil, // no parameters
	)
}

func (b *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) ResetReferenceFileSchemaUri() {
	_jsii_.InvokeVoid(
		b,
		"resetReferenceFileSchemaUri",
		nil, // no parameters
	)
}

func (b *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) ResetSchema() {
	_jsii_.InvokeVoid(
		b,
		"resetSchema",
		nil, // no parameters
	)
}

func (b *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) ResetSourceFormat() {
	_jsii_.InvokeVoid(
		b,
		"resetSourceFormat",
		nil, // no parameters
	)
}

func (b *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) Resolve(_context cdktf.IResolveContext) interface{} {
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

func (b *jsiiProxy_BigqueryTableExternalDataConfigurationOutputReference) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		b,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

