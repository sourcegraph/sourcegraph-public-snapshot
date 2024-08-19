package bigquerytable

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/sourcegraph/managed-services-platform-cdktf/gen/google/jsii"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/bigquerytable/internal"
)

// Represents a {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_table google_bigquery_table}.
type BigqueryTable interface {
	cdktf.TerraformResource
	// Experimental.
	CdktfStack() cdktf.TerraformStack
	Clustering() *[]*string
	SetClustering(val *[]*string)
	ClusteringInput() *[]*string
	// Experimental.
	Connection() interface{}
	// Experimental.
	SetConnection(val interface{})
	// Experimental.
	ConstructNodeMetadata() *map[string]interface{}
	// Experimental.
	Count() interface{}
	// Experimental.
	SetCount(val interface{})
	CreationTime() *float64
	DatasetId() *string
	SetDatasetId(val *string)
	DatasetIdInput() *string
	DeletionProtection() interface{}
	SetDeletionProtection(val interface{})
	DeletionProtectionInput() interface{}
	// Experimental.
	DependsOn() *[]*string
	// Experimental.
	SetDependsOn(val *[]*string)
	Description() *string
	SetDescription(val *string)
	DescriptionInput() *string
	EffectiveLabels() cdktf.StringMap
	EncryptionConfiguration() BigqueryTableEncryptionConfigurationOutputReference
	EncryptionConfigurationInput() *BigqueryTableEncryptionConfiguration
	Etag() *string
	ExpirationTime() *float64
	SetExpirationTime(val *float64)
	ExpirationTimeInput() *float64
	ExternalDataConfiguration() BigqueryTableExternalDataConfigurationOutputReference
	ExternalDataConfigurationInput() *BigqueryTableExternalDataConfiguration
	// Experimental.
	ForEach() cdktf.ITerraformIterator
	// Experimental.
	SetForEach(val cdktf.ITerraformIterator)
	// Experimental.
	Fqn() *string
	FriendlyName() *string
	SetFriendlyName(val *string)
	FriendlyNameInput() *string
	// Experimental.
	FriendlyUniqueId() *string
	Id() *string
	SetId(val *string)
	IdInput() *string
	Labels() *map[string]*string
	SetLabels(val *map[string]*string)
	LabelsInput() *map[string]*string
	LastModifiedTime() *float64
	// Experimental.
	Lifecycle() *cdktf.TerraformResourceLifecycle
	// Experimental.
	SetLifecycle(val *cdktf.TerraformResourceLifecycle)
	Location() *string
	MaterializedView() BigqueryTableMaterializedViewOutputReference
	MaterializedViewInput() *BigqueryTableMaterializedView
	MaxStaleness() *string
	SetMaxStaleness(val *string)
	MaxStalenessInput() *string
	// The tree node.
	Node() constructs.Node
	NumBytes() *float64
	NumLongTermBytes() *float64
	NumRows() *float64
	Project() *string
	SetProject(val *string)
	ProjectInput() *string
	// Experimental.
	Provider() cdktf.TerraformProvider
	// Experimental.
	SetProvider(val cdktf.TerraformProvider)
	// Experimental.
	Provisioners() *[]interface{}
	// Experimental.
	SetProvisioners(val *[]interface{})
	RangePartitioning() BigqueryTableRangePartitioningOutputReference
	RangePartitioningInput() *BigqueryTableRangePartitioning
	// Experimental.
	RawOverrides() interface{}
	RequirePartitionFilter() interface{}
	SetRequirePartitionFilter(val interface{})
	RequirePartitionFilterInput() interface{}
	Schema() *string
	SetSchema(val *string)
	SchemaInput() *string
	SelfLink() *string
	TableConstraints() BigqueryTableTableConstraintsOutputReference
	TableConstraintsInput() *BigqueryTableTableConstraints
	TableId() *string
	SetTableId(val *string)
	TableIdInput() *string
	TableReplicationInfo() BigqueryTableTableReplicationInfoOutputReference
	TableReplicationInfoInput() *BigqueryTableTableReplicationInfo
	// Experimental.
	TerraformGeneratorMetadata() *cdktf.TerraformProviderGeneratorMetadata
	TerraformLabels() cdktf.StringMap
	// Experimental.
	TerraformMetaArguments() *map[string]interface{}
	// Experimental.
	TerraformResourceType() *string
	TimePartitioning() BigqueryTableTimePartitioningOutputReference
	TimePartitioningInput() *BigqueryTableTimePartitioning
	Type() *string
	View() BigqueryTableViewOutputReference
	ViewInput() *BigqueryTableView
	// Experimental.
	AddOverride(path *string, value interface{})
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
	InterpolationForAttribute(terraformAttribute *string) cdktf.IResolvable
	// Overrides the auto-generated logical ID with a specific ID.
	// Experimental.
	OverrideLogicalId(newLogicalId *string)
	PutEncryptionConfiguration(value *BigqueryTableEncryptionConfiguration)
	PutExternalDataConfiguration(value *BigqueryTableExternalDataConfiguration)
	PutMaterializedView(value *BigqueryTableMaterializedView)
	PutRangePartitioning(value *BigqueryTableRangePartitioning)
	PutTableConstraints(value *BigqueryTableTableConstraints)
	PutTableReplicationInfo(value *BigqueryTableTableReplicationInfo)
	PutTimePartitioning(value *BigqueryTableTimePartitioning)
	PutView(value *BigqueryTableView)
	ResetClustering()
	ResetDeletionProtection()
	ResetDescription()
	ResetEncryptionConfiguration()
	ResetExpirationTime()
	ResetExternalDataConfiguration()
	ResetFriendlyName()
	ResetId()
	ResetLabels()
	ResetMaterializedView()
	ResetMaxStaleness()
	// Resets a previously passed logical Id to use the auto-generated logical id again.
	// Experimental.
	ResetOverrideLogicalId()
	ResetProject()
	ResetRangePartitioning()
	ResetRequirePartitionFilter()
	ResetSchema()
	ResetTableConstraints()
	ResetTableReplicationInfo()
	ResetTimePartitioning()
	ResetView()
	SynthesizeAttributes() *map[string]interface{}
	// Experimental.
	ToMetadata() interface{}
	// Returns a string representation of this construct.
	ToString() *string
	// Adds this resource to the terraform JSON output.
	// Experimental.
	ToTerraform() interface{}
}

// The jsii proxy struct for BigqueryTable
type jsiiProxy_BigqueryTable struct {
	internal.Type__cdktfTerraformResource
}

func (j *jsiiProxy_BigqueryTable) CdktfStack() cdktf.TerraformStack {
	var returns cdktf.TerraformStack
	_jsii_.Get(
		j,
		"cdktfStack",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTable) Clustering() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"clustering",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTable) ClusteringInput() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"clusteringInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTable) Connection() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"connection",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTable) ConstructNodeMetadata() *map[string]interface{} {
	var returns *map[string]interface{}
	_jsii_.Get(
		j,
		"constructNodeMetadata",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTable) Count() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"count",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTable) CreationTime() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"creationTime",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTable) DatasetId() *string {
	var returns *string
	_jsii_.Get(
		j,
		"datasetId",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTable) DatasetIdInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"datasetIdInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTable) DeletionProtection() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"deletionProtection",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTable) DeletionProtectionInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"deletionProtectionInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTable) DependsOn() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"dependsOn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTable) Description() *string {
	var returns *string
	_jsii_.Get(
		j,
		"description",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTable) DescriptionInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"descriptionInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTable) EffectiveLabels() cdktf.StringMap {
	var returns cdktf.StringMap
	_jsii_.Get(
		j,
		"effectiveLabels",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTable) EncryptionConfiguration() BigqueryTableEncryptionConfigurationOutputReference {
	var returns BigqueryTableEncryptionConfigurationOutputReference
	_jsii_.Get(
		j,
		"encryptionConfiguration",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTable) EncryptionConfigurationInput() *BigqueryTableEncryptionConfiguration {
	var returns *BigqueryTableEncryptionConfiguration
	_jsii_.Get(
		j,
		"encryptionConfigurationInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTable) Etag() *string {
	var returns *string
	_jsii_.Get(
		j,
		"etag",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTable) ExpirationTime() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"expirationTime",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTable) ExpirationTimeInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"expirationTimeInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTable) ExternalDataConfiguration() BigqueryTableExternalDataConfigurationOutputReference {
	var returns BigqueryTableExternalDataConfigurationOutputReference
	_jsii_.Get(
		j,
		"externalDataConfiguration",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTable) ExternalDataConfigurationInput() *BigqueryTableExternalDataConfiguration {
	var returns *BigqueryTableExternalDataConfiguration
	_jsii_.Get(
		j,
		"externalDataConfigurationInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTable) ForEach() cdktf.ITerraformIterator {
	var returns cdktf.ITerraformIterator
	_jsii_.Get(
		j,
		"forEach",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTable) Fqn() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fqn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTable) FriendlyName() *string {
	var returns *string
	_jsii_.Get(
		j,
		"friendlyName",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTable) FriendlyNameInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"friendlyNameInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTable) FriendlyUniqueId() *string {
	var returns *string
	_jsii_.Get(
		j,
		"friendlyUniqueId",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTable) Id() *string {
	var returns *string
	_jsii_.Get(
		j,
		"id",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTable) IdInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"idInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTable) Labels() *map[string]*string {
	var returns *map[string]*string
	_jsii_.Get(
		j,
		"labels",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTable) LabelsInput() *map[string]*string {
	var returns *map[string]*string
	_jsii_.Get(
		j,
		"labelsInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTable) LastModifiedTime() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"lastModifiedTime",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTable) Lifecycle() *cdktf.TerraformResourceLifecycle {
	var returns *cdktf.TerraformResourceLifecycle
	_jsii_.Get(
		j,
		"lifecycle",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTable) Location() *string {
	var returns *string
	_jsii_.Get(
		j,
		"location",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTable) MaterializedView() BigqueryTableMaterializedViewOutputReference {
	var returns BigqueryTableMaterializedViewOutputReference
	_jsii_.Get(
		j,
		"materializedView",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTable) MaterializedViewInput() *BigqueryTableMaterializedView {
	var returns *BigqueryTableMaterializedView
	_jsii_.Get(
		j,
		"materializedViewInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTable) MaxStaleness() *string {
	var returns *string
	_jsii_.Get(
		j,
		"maxStaleness",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTable) MaxStalenessInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"maxStalenessInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTable) Node() constructs.Node {
	var returns constructs.Node
	_jsii_.Get(
		j,
		"node",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTable) NumBytes() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"numBytes",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTable) NumLongTermBytes() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"numLongTermBytes",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTable) NumRows() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"numRows",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTable) Project() *string {
	var returns *string
	_jsii_.Get(
		j,
		"project",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTable) ProjectInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"projectInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTable) Provider() cdktf.TerraformProvider {
	var returns cdktf.TerraformProvider
	_jsii_.Get(
		j,
		"provider",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTable) Provisioners() *[]interface{} {
	var returns *[]interface{}
	_jsii_.Get(
		j,
		"provisioners",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTable) RangePartitioning() BigqueryTableRangePartitioningOutputReference {
	var returns BigqueryTableRangePartitioningOutputReference
	_jsii_.Get(
		j,
		"rangePartitioning",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTable) RangePartitioningInput() *BigqueryTableRangePartitioning {
	var returns *BigqueryTableRangePartitioning
	_jsii_.Get(
		j,
		"rangePartitioningInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTable) RawOverrides() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"rawOverrides",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTable) RequirePartitionFilter() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"requirePartitionFilter",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTable) RequirePartitionFilterInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"requirePartitionFilterInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTable) Schema() *string {
	var returns *string
	_jsii_.Get(
		j,
		"schema",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTable) SchemaInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"schemaInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTable) SelfLink() *string {
	var returns *string
	_jsii_.Get(
		j,
		"selfLink",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTable) TableConstraints() BigqueryTableTableConstraintsOutputReference {
	var returns BigqueryTableTableConstraintsOutputReference
	_jsii_.Get(
		j,
		"tableConstraints",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTable) TableConstraintsInput() *BigqueryTableTableConstraints {
	var returns *BigqueryTableTableConstraints
	_jsii_.Get(
		j,
		"tableConstraintsInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTable) TableId() *string {
	var returns *string
	_jsii_.Get(
		j,
		"tableId",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTable) TableIdInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"tableIdInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTable) TableReplicationInfo() BigqueryTableTableReplicationInfoOutputReference {
	var returns BigqueryTableTableReplicationInfoOutputReference
	_jsii_.Get(
		j,
		"tableReplicationInfo",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTable) TableReplicationInfoInput() *BigqueryTableTableReplicationInfo {
	var returns *BigqueryTableTableReplicationInfo
	_jsii_.Get(
		j,
		"tableReplicationInfoInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTable) TerraformGeneratorMetadata() *cdktf.TerraformProviderGeneratorMetadata {
	var returns *cdktf.TerraformProviderGeneratorMetadata
	_jsii_.Get(
		j,
		"terraformGeneratorMetadata",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTable) TerraformLabels() cdktf.StringMap {
	var returns cdktf.StringMap
	_jsii_.Get(
		j,
		"terraformLabels",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTable) TerraformMetaArguments() *map[string]interface{} {
	var returns *map[string]interface{}
	_jsii_.Get(
		j,
		"terraformMetaArguments",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTable) TerraformResourceType() *string {
	var returns *string
	_jsii_.Get(
		j,
		"terraformResourceType",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTable) TimePartitioning() BigqueryTableTimePartitioningOutputReference {
	var returns BigqueryTableTimePartitioningOutputReference
	_jsii_.Get(
		j,
		"timePartitioning",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTable) TimePartitioningInput() *BigqueryTableTimePartitioning {
	var returns *BigqueryTableTimePartitioning
	_jsii_.Get(
		j,
		"timePartitioningInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTable) Type() *string {
	var returns *string
	_jsii_.Get(
		j,
		"type",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTable) View() BigqueryTableViewOutputReference {
	var returns BigqueryTableViewOutputReference
	_jsii_.Get(
		j,
		"view",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_BigqueryTable) ViewInput() *BigqueryTableView {
	var returns *BigqueryTableView
	_jsii_.Get(
		j,
		"viewInput",
		&returns,
	)
	return returns
}


// Create a new {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_table google_bigquery_table} Resource.
func NewBigqueryTable(scope constructs.Construct, id *string, config *BigqueryTableConfig) BigqueryTable {
	_init_.Initialize()

	if err := validateNewBigqueryTableParameters(scope, id, config); err != nil {
		panic(err)
	}
	j := jsiiProxy_BigqueryTable{}

	_jsii_.Create(
		"@cdktf/provider-google.bigqueryTable.BigqueryTable",
		[]interface{}{scope, id, config},
		&j,
	)

	return &j
}

// Create a new {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/bigquery_table google_bigquery_table} Resource.
func NewBigqueryTable_Override(b BigqueryTable, scope constructs.Construct, id *string, config *BigqueryTableConfig) {
	_init_.Initialize()

	_jsii_.Create(
		"@cdktf/provider-google.bigqueryTable.BigqueryTable",
		[]interface{}{scope, id, config},
		b,
	)
}

func (j *jsiiProxy_BigqueryTable)SetClustering(val *[]*string) {
	if err := j.validateSetClusteringParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"clustering",
		val,
	)
}

func (j *jsiiProxy_BigqueryTable)SetConnection(val interface{}) {
	if err := j.validateSetConnectionParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"connection",
		val,
	)
}

func (j *jsiiProxy_BigqueryTable)SetCount(val interface{}) {
	if err := j.validateSetCountParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"count",
		val,
	)
}

func (j *jsiiProxy_BigqueryTable)SetDatasetId(val *string) {
	if err := j.validateSetDatasetIdParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"datasetId",
		val,
	)
}

func (j *jsiiProxy_BigqueryTable)SetDeletionProtection(val interface{}) {
	if err := j.validateSetDeletionProtectionParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"deletionProtection",
		val,
	)
}

func (j *jsiiProxy_BigqueryTable)SetDependsOn(val *[]*string) {
	_jsii_.Set(
		j,
		"dependsOn",
		val,
	)
}

func (j *jsiiProxy_BigqueryTable)SetDescription(val *string) {
	if err := j.validateSetDescriptionParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"description",
		val,
	)
}

func (j *jsiiProxy_BigqueryTable)SetExpirationTime(val *float64) {
	if err := j.validateSetExpirationTimeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"expirationTime",
		val,
	)
}

func (j *jsiiProxy_BigqueryTable)SetForEach(val cdktf.ITerraformIterator) {
	_jsii_.Set(
		j,
		"forEach",
		val,
	)
}

func (j *jsiiProxy_BigqueryTable)SetFriendlyName(val *string) {
	if err := j.validateSetFriendlyNameParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"friendlyName",
		val,
	)
}

func (j *jsiiProxy_BigqueryTable)SetId(val *string) {
	if err := j.validateSetIdParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"id",
		val,
	)
}

func (j *jsiiProxy_BigqueryTable)SetLabels(val *map[string]*string) {
	if err := j.validateSetLabelsParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"labels",
		val,
	)
}

func (j *jsiiProxy_BigqueryTable)SetLifecycle(val *cdktf.TerraformResourceLifecycle) {
	if err := j.validateSetLifecycleParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"lifecycle",
		val,
	)
}

func (j *jsiiProxy_BigqueryTable)SetMaxStaleness(val *string) {
	if err := j.validateSetMaxStalenessParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"maxStaleness",
		val,
	)
}

func (j *jsiiProxy_BigqueryTable)SetProject(val *string) {
	if err := j.validateSetProjectParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"project",
		val,
	)
}

func (j *jsiiProxy_BigqueryTable)SetProvider(val cdktf.TerraformProvider) {
	_jsii_.Set(
		j,
		"provider",
		val,
	)
}

func (j *jsiiProxy_BigqueryTable)SetProvisioners(val *[]interface{}) {
	if err := j.validateSetProvisionersParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"provisioners",
		val,
	)
}

func (j *jsiiProxy_BigqueryTable)SetRequirePartitionFilter(val interface{}) {
	if err := j.validateSetRequirePartitionFilterParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"requirePartitionFilter",
		val,
	)
}

func (j *jsiiProxy_BigqueryTable)SetSchema(val *string) {
	if err := j.validateSetSchemaParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"schema",
		val,
	)
}

func (j *jsiiProxy_BigqueryTable)SetTableId(val *string) {
	if err := j.validateSetTableIdParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"tableId",
		val,
	)
}

// Checks if `x` is a construct.
//
// Use this method instead of `instanceof` to properly detect `Construct`
// instances, even when the construct library is symlinked.
//
// Explanation: in JavaScript, multiple copies of the `constructs` library on
// disk are seen as independent, completely different libraries. As a
// consequence, the class `Construct` in each copy of the `constructs` library
// is seen as a different class, and an instance of one class will not test as
// `instanceof` the other class. `npm install` will not create installations
// like this, but users may manually symlink construct libraries together or
// use a monorepo tool: in those cases, multiple copies of the `constructs`
// library can be accidentally installed, and `instanceof` will behave
// unpredictably. It is safest to avoid using `instanceof`, and using
// this type-testing method instead.
//
// Returns: true if `x` is an object created from a class which extends `Construct`.
func BigqueryTable_IsConstruct(x interface{}) *bool {
	_init_.Initialize()

	if err := validateBigqueryTable_IsConstructParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"@cdktf/provider-google.bigqueryTable.BigqueryTable",
		"isConstruct",
		[]interface{}{x},
		&returns,
	)

	return returns
}

// Experimental.
func BigqueryTable_IsTerraformElement(x interface{}) *bool {
	_init_.Initialize()

	if err := validateBigqueryTable_IsTerraformElementParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"@cdktf/provider-google.bigqueryTable.BigqueryTable",
		"isTerraformElement",
		[]interface{}{x},
		&returns,
	)

	return returns
}

// Experimental.
func BigqueryTable_IsTerraformResource(x interface{}) *bool {
	_init_.Initialize()

	if err := validateBigqueryTable_IsTerraformResourceParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"@cdktf/provider-google.bigqueryTable.BigqueryTable",
		"isTerraformResource",
		[]interface{}{x},
		&returns,
	)

	return returns
}

func BigqueryTable_TfResourceType() *string {
	_init_.Initialize()
	var returns *string
	_jsii_.StaticGet(
		"@cdktf/provider-google.bigqueryTable.BigqueryTable",
		"tfResourceType",
		&returns,
	)
	return returns
}

func (b *jsiiProxy_BigqueryTable) AddOverride(path *string, value interface{}) {
	if err := b.validateAddOverrideParameters(path, value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		b,
		"addOverride",
		[]interface{}{path, value},
	)
}

func (b *jsiiProxy_BigqueryTable) GetAnyMapAttribute(terraformAttribute *string) *map[string]interface{} {
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

func (b *jsiiProxy_BigqueryTable) GetBooleanAttribute(terraformAttribute *string) cdktf.IResolvable {
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

func (b *jsiiProxy_BigqueryTable) GetBooleanMapAttribute(terraformAttribute *string) *map[string]*bool {
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

func (b *jsiiProxy_BigqueryTable) GetListAttribute(terraformAttribute *string) *[]*string {
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

func (b *jsiiProxy_BigqueryTable) GetNumberAttribute(terraformAttribute *string) *float64 {
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

func (b *jsiiProxy_BigqueryTable) GetNumberListAttribute(terraformAttribute *string) *[]*float64 {
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

func (b *jsiiProxy_BigqueryTable) GetNumberMapAttribute(terraformAttribute *string) *map[string]*float64 {
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

func (b *jsiiProxy_BigqueryTable) GetStringAttribute(terraformAttribute *string) *string {
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

func (b *jsiiProxy_BigqueryTable) GetStringMapAttribute(terraformAttribute *string) *map[string]*string {
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

func (b *jsiiProxy_BigqueryTable) InterpolationForAttribute(terraformAttribute *string) cdktf.IResolvable {
	if err := b.validateInterpolationForAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		b,
		"interpolationForAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (b *jsiiProxy_BigqueryTable) OverrideLogicalId(newLogicalId *string) {
	if err := b.validateOverrideLogicalIdParameters(newLogicalId); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		b,
		"overrideLogicalId",
		[]interface{}{newLogicalId},
	)
}

func (b *jsiiProxy_BigqueryTable) PutEncryptionConfiguration(value *BigqueryTableEncryptionConfiguration) {
	if err := b.validatePutEncryptionConfigurationParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		b,
		"putEncryptionConfiguration",
		[]interface{}{value},
	)
}

func (b *jsiiProxy_BigqueryTable) PutExternalDataConfiguration(value *BigqueryTableExternalDataConfiguration) {
	if err := b.validatePutExternalDataConfigurationParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		b,
		"putExternalDataConfiguration",
		[]interface{}{value},
	)
}

func (b *jsiiProxy_BigqueryTable) PutMaterializedView(value *BigqueryTableMaterializedView) {
	if err := b.validatePutMaterializedViewParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		b,
		"putMaterializedView",
		[]interface{}{value},
	)
}

func (b *jsiiProxy_BigqueryTable) PutRangePartitioning(value *BigqueryTableRangePartitioning) {
	if err := b.validatePutRangePartitioningParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		b,
		"putRangePartitioning",
		[]interface{}{value},
	)
}

func (b *jsiiProxy_BigqueryTable) PutTableConstraints(value *BigqueryTableTableConstraints) {
	if err := b.validatePutTableConstraintsParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		b,
		"putTableConstraints",
		[]interface{}{value},
	)
}

func (b *jsiiProxy_BigqueryTable) PutTableReplicationInfo(value *BigqueryTableTableReplicationInfo) {
	if err := b.validatePutTableReplicationInfoParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		b,
		"putTableReplicationInfo",
		[]interface{}{value},
	)
}

func (b *jsiiProxy_BigqueryTable) PutTimePartitioning(value *BigqueryTableTimePartitioning) {
	if err := b.validatePutTimePartitioningParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		b,
		"putTimePartitioning",
		[]interface{}{value},
	)
}

func (b *jsiiProxy_BigqueryTable) PutView(value *BigqueryTableView) {
	if err := b.validatePutViewParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		b,
		"putView",
		[]interface{}{value},
	)
}

func (b *jsiiProxy_BigqueryTable) ResetClustering() {
	_jsii_.InvokeVoid(
		b,
		"resetClustering",
		nil, // no parameters
	)
}

func (b *jsiiProxy_BigqueryTable) ResetDeletionProtection() {
	_jsii_.InvokeVoid(
		b,
		"resetDeletionProtection",
		nil, // no parameters
	)
}

func (b *jsiiProxy_BigqueryTable) ResetDescription() {
	_jsii_.InvokeVoid(
		b,
		"resetDescription",
		nil, // no parameters
	)
}

func (b *jsiiProxy_BigqueryTable) ResetEncryptionConfiguration() {
	_jsii_.InvokeVoid(
		b,
		"resetEncryptionConfiguration",
		nil, // no parameters
	)
}

func (b *jsiiProxy_BigqueryTable) ResetExpirationTime() {
	_jsii_.InvokeVoid(
		b,
		"resetExpirationTime",
		nil, // no parameters
	)
}

func (b *jsiiProxy_BigqueryTable) ResetExternalDataConfiguration() {
	_jsii_.InvokeVoid(
		b,
		"resetExternalDataConfiguration",
		nil, // no parameters
	)
}

func (b *jsiiProxy_BigqueryTable) ResetFriendlyName() {
	_jsii_.InvokeVoid(
		b,
		"resetFriendlyName",
		nil, // no parameters
	)
}

func (b *jsiiProxy_BigqueryTable) ResetId() {
	_jsii_.InvokeVoid(
		b,
		"resetId",
		nil, // no parameters
	)
}

func (b *jsiiProxy_BigqueryTable) ResetLabels() {
	_jsii_.InvokeVoid(
		b,
		"resetLabels",
		nil, // no parameters
	)
}

func (b *jsiiProxy_BigqueryTable) ResetMaterializedView() {
	_jsii_.InvokeVoid(
		b,
		"resetMaterializedView",
		nil, // no parameters
	)
}

func (b *jsiiProxy_BigqueryTable) ResetMaxStaleness() {
	_jsii_.InvokeVoid(
		b,
		"resetMaxStaleness",
		nil, // no parameters
	)
}

func (b *jsiiProxy_BigqueryTable) ResetOverrideLogicalId() {
	_jsii_.InvokeVoid(
		b,
		"resetOverrideLogicalId",
		nil, // no parameters
	)
}

func (b *jsiiProxy_BigqueryTable) ResetProject() {
	_jsii_.InvokeVoid(
		b,
		"resetProject",
		nil, // no parameters
	)
}

func (b *jsiiProxy_BigqueryTable) ResetRangePartitioning() {
	_jsii_.InvokeVoid(
		b,
		"resetRangePartitioning",
		nil, // no parameters
	)
}

func (b *jsiiProxy_BigqueryTable) ResetRequirePartitionFilter() {
	_jsii_.InvokeVoid(
		b,
		"resetRequirePartitionFilter",
		nil, // no parameters
	)
}

func (b *jsiiProxy_BigqueryTable) ResetSchema() {
	_jsii_.InvokeVoid(
		b,
		"resetSchema",
		nil, // no parameters
	)
}

func (b *jsiiProxy_BigqueryTable) ResetTableConstraints() {
	_jsii_.InvokeVoid(
		b,
		"resetTableConstraints",
		nil, // no parameters
	)
}

func (b *jsiiProxy_BigqueryTable) ResetTableReplicationInfo() {
	_jsii_.InvokeVoid(
		b,
		"resetTableReplicationInfo",
		nil, // no parameters
	)
}

func (b *jsiiProxy_BigqueryTable) ResetTimePartitioning() {
	_jsii_.InvokeVoid(
		b,
		"resetTimePartitioning",
		nil, // no parameters
	)
}

func (b *jsiiProxy_BigqueryTable) ResetView() {
	_jsii_.InvokeVoid(
		b,
		"resetView",
		nil, // no parameters
	)
}

func (b *jsiiProxy_BigqueryTable) SynthesizeAttributes() *map[string]interface{} {
	var returns *map[string]interface{}

	_jsii_.Invoke(
		b,
		"synthesizeAttributes",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (b *jsiiProxy_BigqueryTable) ToMetadata() interface{} {
	var returns interface{}

	_jsii_.Invoke(
		b,
		"toMetadata",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (b *jsiiProxy_BigqueryTable) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		b,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (b *jsiiProxy_BigqueryTable) ToTerraform() interface{} {
	var returns interface{}

	_jsii_.Invoke(
		b,
		"toTerraform",
		nil, // no parameters
		&returns,
	)

	return returns
}

