package sqldatabaseinstance

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/sourcegraph/managed-services-platform-cdktf/gen/google/jsii"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/sqldatabaseinstance/internal"
)

// Represents a {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_database_instance google_sql_database_instance}.
type SqlDatabaseInstance interface {
	cdktf.TerraformResource
	AvailableMaintenanceVersions() *[]*string
	// Experimental.
	CdktfStack() cdktf.TerraformStack
	Clone() SqlDatabaseInstanceCloneOutputReference
	CloneInput() *SqlDatabaseInstanceClone
	// Experimental.
	Connection() interface{}
	// Experimental.
	SetConnection(val interface{})
	ConnectionName() *string
	// Experimental.
	ConstructNodeMetadata() *map[string]interface{}
	// Experimental.
	Count() interface{}
	// Experimental.
	SetCount(val interface{})
	DatabaseVersion() *string
	SetDatabaseVersion(val *string)
	DatabaseVersionInput() *string
	DeletionProtection() interface{}
	SetDeletionProtection(val interface{})
	DeletionProtectionInput() interface{}
	// Experimental.
	DependsOn() *[]*string
	// Experimental.
	SetDependsOn(val *[]*string)
	DnsName() *string
	EncryptionKeyName() *string
	SetEncryptionKeyName(val *string)
	EncryptionKeyNameInput() *string
	FirstIpAddress() *string
	// Experimental.
	ForEach() cdktf.ITerraformIterator
	// Experimental.
	SetForEach(val cdktf.ITerraformIterator)
	// Experimental.
	Fqn() *string
	// Experimental.
	FriendlyUniqueId() *string
	Id() *string
	SetId(val *string)
	IdInput() *string
	InstanceType() *string
	SetInstanceType(val *string)
	InstanceTypeInput() *string
	IpAddress() SqlDatabaseInstanceIpAddressList
	// Experimental.
	Lifecycle() *cdktf.TerraformResourceLifecycle
	// Experimental.
	SetLifecycle(val *cdktf.TerraformResourceLifecycle)
	MaintenanceVersion() *string
	SetMaintenanceVersion(val *string)
	MaintenanceVersionInput() *string
	MasterInstanceName() *string
	SetMasterInstanceName(val *string)
	MasterInstanceNameInput() *string
	Name() *string
	SetName(val *string)
	NameInput() *string
	// The tree node.
	Node() constructs.Node
	PrivateIpAddress() *string
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
	PscServiceAttachmentLink() *string
	PublicIpAddress() *string
	// Experimental.
	RawOverrides() interface{}
	Region() *string
	SetRegion(val *string)
	RegionInput() *string
	ReplicaConfiguration() SqlDatabaseInstanceReplicaConfigurationOutputReference
	ReplicaConfigurationInput() *SqlDatabaseInstanceReplicaConfiguration
	RestoreBackupContext() SqlDatabaseInstanceRestoreBackupContextOutputReference
	RestoreBackupContextInput() *SqlDatabaseInstanceRestoreBackupContext
	RootPassword() *string
	SetRootPassword(val *string)
	RootPasswordInput() *string
	SelfLink() *string
	ServerCaCert() SqlDatabaseInstanceServerCaCertList
	ServiceAccountEmailAddress() *string
	Settings() SqlDatabaseInstanceSettingsOutputReference
	SettingsInput() *SqlDatabaseInstanceSettings
	// Experimental.
	TerraformGeneratorMetadata() *cdktf.TerraformProviderGeneratorMetadata
	// Experimental.
	TerraformMetaArguments() *map[string]interface{}
	// Experimental.
	TerraformResourceType() *string
	Timeouts() SqlDatabaseInstanceTimeoutsOutputReference
	TimeoutsInput() interface{}
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
	PutClone(value *SqlDatabaseInstanceClone)
	PutReplicaConfiguration(value *SqlDatabaseInstanceReplicaConfiguration)
	PutRestoreBackupContext(value *SqlDatabaseInstanceRestoreBackupContext)
	PutSettings(value *SqlDatabaseInstanceSettings)
	PutTimeouts(value *SqlDatabaseInstanceTimeouts)
	ResetClone()
	ResetDeletionProtection()
	ResetEncryptionKeyName()
	ResetId()
	ResetInstanceType()
	ResetMaintenanceVersion()
	ResetMasterInstanceName()
	ResetName()
	// Resets a previously passed logical Id to use the auto-generated logical id again.
	// Experimental.
	ResetOverrideLogicalId()
	ResetProject()
	ResetRegion()
	ResetReplicaConfiguration()
	ResetRestoreBackupContext()
	ResetRootPassword()
	ResetSettings()
	ResetTimeouts()
	SynthesizeAttributes() *map[string]interface{}
	// Experimental.
	ToMetadata() interface{}
	// Returns a string representation of this construct.
	ToString() *string
	// Adds this resource to the terraform JSON output.
	// Experimental.
	ToTerraform() interface{}
}

// The jsii proxy struct for SqlDatabaseInstance
type jsiiProxy_SqlDatabaseInstance struct {
	internal.Type__cdktfTerraformResource
}

func (j *jsiiProxy_SqlDatabaseInstance) AvailableMaintenanceVersions() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"availableMaintenanceVersions",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstance) CdktfStack() cdktf.TerraformStack {
	var returns cdktf.TerraformStack
	_jsii_.Get(
		j,
		"cdktfStack",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstance) Clone() SqlDatabaseInstanceCloneOutputReference {
	var returns SqlDatabaseInstanceCloneOutputReference
	_jsii_.Get(
		j,
		"clone",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstance) CloneInput() *SqlDatabaseInstanceClone {
	var returns *SqlDatabaseInstanceClone
	_jsii_.Get(
		j,
		"cloneInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstance) Connection() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"connection",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstance) ConnectionName() *string {
	var returns *string
	_jsii_.Get(
		j,
		"connectionName",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstance) ConstructNodeMetadata() *map[string]interface{} {
	var returns *map[string]interface{}
	_jsii_.Get(
		j,
		"constructNodeMetadata",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstance) Count() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"count",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstance) DatabaseVersion() *string {
	var returns *string
	_jsii_.Get(
		j,
		"databaseVersion",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstance) DatabaseVersionInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"databaseVersionInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstance) DeletionProtection() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"deletionProtection",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstance) DeletionProtectionInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"deletionProtectionInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstance) DependsOn() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"dependsOn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstance) DnsName() *string {
	var returns *string
	_jsii_.Get(
		j,
		"dnsName",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstance) EncryptionKeyName() *string {
	var returns *string
	_jsii_.Get(
		j,
		"encryptionKeyName",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstance) EncryptionKeyNameInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"encryptionKeyNameInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstance) FirstIpAddress() *string {
	var returns *string
	_jsii_.Get(
		j,
		"firstIpAddress",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstance) ForEach() cdktf.ITerraformIterator {
	var returns cdktf.ITerraformIterator
	_jsii_.Get(
		j,
		"forEach",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstance) Fqn() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fqn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstance) FriendlyUniqueId() *string {
	var returns *string
	_jsii_.Get(
		j,
		"friendlyUniqueId",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstance) Id() *string {
	var returns *string
	_jsii_.Get(
		j,
		"id",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstance) IdInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"idInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstance) InstanceType() *string {
	var returns *string
	_jsii_.Get(
		j,
		"instanceType",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstance) InstanceTypeInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"instanceTypeInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstance) IpAddress() SqlDatabaseInstanceIpAddressList {
	var returns SqlDatabaseInstanceIpAddressList
	_jsii_.Get(
		j,
		"ipAddress",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstance) Lifecycle() *cdktf.TerraformResourceLifecycle {
	var returns *cdktf.TerraformResourceLifecycle
	_jsii_.Get(
		j,
		"lifecycle",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstance) MaintenanceVersion() *string {
	var returns *string
	_jsii_.Get(
		j,
		"maintenanceVersion",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstance) MaintenanceVersionInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"maintenanceVersionInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstance) MasterInstanceName() *string {
	var returns *string
	_jsii_.Get(
		j,
		"masterInstanceName",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstance) MasterInstanceNameInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"masterInstanceNameInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstance) Name() *string {
	var returns *string
	_jsii_.Get(
		j,
		"name",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstance) NameInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"nameInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstance) Node() constructs.Node {
	var returns constructs.Node
	_jsii_.Get(
		j,
		"node",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstance) PrivateIpAddress() *string {
	var returns *string
	_jsii_.Get(
		j,
		"privateIpAddress",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstance) Project() *string {
	var returns *string
	_jsii_.Get(
		j,
		"project",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstance) ProjectInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"projectInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstance) Provider() cdktf.TerraformProvider {
	var returns cdktf.TerraformProvider
	_jsii_.Get(
		j,
		"provider",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstance) Provisioners() *[]interface{} {
	var returns *[]interface{}
	_jsii_.Get(
		j,
		"provisioners",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstance) PscServiceAttachmentLink() *string {
	var returns *string
	_jsii_.Get(
		j,
		"pscServiceAttachmentLink",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstance) PublicIpAddress() *string {
	var returns *string
	_jsii_.Get(
		j,
		"publicIpAddress",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstance) RawOverrides() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"rawOverrides",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstance) Region() *string {
	var returns *string
	_jsii_.Get(
		j,
		"region",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstance) RegionInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"regionInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstance) ReplicaConfiguration() SqlDatabaseInstanceReplicaConfigurationOutputReference {
	var returns SqlDatabaseInstanceReplicaConfigurationOutputReference
	_jsii_.Get(
		j,
		"replicaConfiguration",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstance) ReplicaConfigurationInput() *SqlDatabaseInstanceReplicaConfiguration {
	var returns *SqlDatabaseInstanceReplicaConfiguration
	_jsii_.Get(
		j,
		"replicaConfigurationInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstance) RestoreBackupContext() SqlDatabaseInstanceRestoreBackupContextOutputReference {
	var returns SqlDatabaseInstanceRestoreBackupContextOutputReference
	_jsii_.Get(
		j,
		"restoreBackupContext",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstance) RestoreBackupContextInput() *SqlDatabaseInstanceRestoreBackupContext {
	var returns *SqlDatabaseInstanceRestoreBackupContext
	_jsii_.Get(
		j,
		"restoreBackupContextInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstance) RootPassword() *string {
	var returns *string
	_jsii_.Get(
		j,
		"rootPassword",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstance) RootPasswordInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"rootPasswordInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstance) SelfLink() *string {
	var returns *string
	_jsii_.Get(
		j,
		"selfLink",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstance) ServerCaCert() SqlDatabaseInstanceServerCaCertList {
	var returns SqlDatabaseInstanceServerCaCertList
	_jsii_.Get(
		j,
		"serverCaCert",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstance) ServiceAccountEmailAddress() *string {
	var returns *string
	_jsii_.Get(
		j,
		"serviceAccountEmailAddress",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstance) Settings() SqlDatabaseInstanceSettingsOutputReference {
	var returns SqlDatabaseInstanceSettingsOutputReference
	_jsii_.Get(
		j,
		"settings",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstance) SettingsInput() *SqlDatabaseInstanceSettings {
	var returns *SqlDatabaseInstanceSettings
	_jsii_.Get(
		j,
		"settingsInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstance) TerraformGeneratorMetadata() *cdktf.TerraformProviderGeneratorMetadata {
	var returns *cdktf.TerraformProviderGeneratorMetadata
	_jsii_.Get(
		j,
		"terraformGeneratorMetadata",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstance) TerraformMetaArguments() *map[string]interface{} {
	var returns *map[string]interface{}
	_jsii_.Get(
		j,
		"terraformMetaArguments",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstance) TerraformResourceType() *string {
	var returns *string
	_jsii_.Get(
		j,
		"terraformResourceType",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstance) Timeouts() SqlDatabaseInstanceTimeoutsOutputReference {
	var returns SqlDatabaseInstanceTimeoutsOutputReference
	_jsii_.Get(
		j,
		"timeouts",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstance) TimeoutsInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"timeoutsInput",
		&returns,
	)
	return returns
}


// Create a new {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_database_instance google_sql_database_instance} Resource.
func NewSqlDatabaseInstance(scope constructs.Construct, id *string, config *SqlDatabaseInstanceConfig) SqlDatabaseInstance {
	_init_.Initialize()

	if err := validateNewSqlDatabaseInstanceParameters(scope, id, config); err != nil {
		panic(err)
	}
	j := jsiiProxy_SqlDatabaseInstance{}

	_jsii_.Create(
		"@cdktf/provider-google.sqlDatabaseInstance.SqlDatabaseInstance",
		[]interface{}{scope, id, config},
		&j,
	)

	return &j
}

// Create a new {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_database_instance google_sql_database_instance} Resource.
func NewSqlDatabaseInstance_Override(s SqlDatabaseInstance, scope constructs.Construct, id *string, config *SqlDatabaseInstanceConfig) {
	_init_.Initialize()

	_jsii_.Create(
		"@cdktf/provider-google.sqlDatabaseInstance.SqlDatabaseInstance",
		[]interface{}{scope, id, config},
		s,
	)
}

func (j *jsiiProxy_SqlDatabaseInstance)SetConnection(val interface{}) {
	if err := j.validateSetConnectionParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"connection",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstance)SetCount(val interface{}) {
	if err := j.validateSetCountParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"count",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstance)SetDatabaseVersion(val *string) {
	if err := j.validateSetDatabaseVersionParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"databaseVersion",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstance)SetDeletionProtection(val interface{}) {
	if err := j.validateSetDeletionProtectionParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"deletionProtection",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstance)SetDependsOn(val *[]*string) {
	_jsii_.Set(
		j,
		"dependsOn",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstance)SetEncryptionKeyName(val *string) {
	if err := j.validateSetEncryptionKeyNameParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"encryptionKeyName",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstance)SetForEach(val cdktf.ITerraformIterator) {
	_jsii_.Set(
		j,
		"forEach",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstance)SetId(val *string) {
	if err := j.validateSetIdParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"id",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstance)SetInstanceType(val *string) {
	if err := j.validateSetInstanceTypeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"instanceType",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstance)SetLifecycle(val *cdktf.TerraformResourceLifecycle) {
	if err := j.validateSetLifecycleParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"lifecycle",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstance)SetMaintenanceVersion(val *string) {
	if err := j.validateSetMaintenanceVersionParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"maintenanceVersion",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstance)SetMasterInstanceName(val *string) {
	if err := j.validateSetMasterInstanceNameParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"masterInstanceName",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstance)SetName(val *string) {
	if err := j.validateSetNameParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"name",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstance)SetProject(val *string) {
	if err := j.validateSetProjectParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"project",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstance)SetProvider(val cdktf.TerraformProvider) {
	_jsii_.Set(
		j,
		"provider",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstance)SetProvisioners(val *[]interface{}) {
	if err := j.validateSetProvisionersParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"provisioners",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstance)SetRegion(val *string) {
	if err := j.validateSetRegionParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"region",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstance)SetRootPassword(val *string) {
	if err := j.validateSetRootPasswordParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"rootPassword",
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
func SqlDatabaseInstance_IsConstruct(x interface{}) *bool {
	_init_.Initialize()

	if err := validateSqlDatabaseInstance_IsConstructParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"@cdktf/provider-google.sqlDatabaseInstance.SqlDatabaseInstance",
		"isConstruct",
		[]interface{}{x},
		&returns,
	)

	return returns
}

// Experimental.
func SqlDatabaseInstance_IsTerraformElement(x interface{}) *bool {
	_init_.Initialize()

	if err := validateSqlDatabaseInstance_IsTerraformElementParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"@cdktf/provider-google.sqlDatabaseInstance.SqlDatabaseInstance",
		"isTerraformElement",
		[]interface{}{x},
		&returns,
	)

	return returns
}

// Experimental.
func SqlDatabaseInstance_IsTerraformResource(x interface{}) *bool {
	_init_.Initialize()

	if err := validateSqlDatabaseInstance_IsTerraformResourceParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"@cdktf/provider-google.sqlDatabaseInstance.SqlDatabaseInstance",
		"isTerraformResource",
		[]interface{}{x},
		&returns,
	)

	return returns
}

func SqlDatabaseInstance_TfResourceType() *string {
	_init_.Initialize()
	var returns *string
	_jsii_.StaticGet(
		"@cdktf/provider-google.sqlDatabaseInstance.SqlDatabaseInstance",
		"tfResourceType",
		&returns,
	)
	return returns
}

func (s *jsiiProxy_SqlDatabaseInstance) AddOverride(path *string, value interface{}) {
	if err := s.validateAddOverrideParameters(path, value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		s,
		"addOverride",
		[]interface{}{path, value},
	)
}

func (s *jsiiProxy_SqlDatabaseInstance) GetAnyMapAttribute(terraformAttribute *string) *map[string]interface{} {
	if err := s.validateGetAnyMapAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *map[string]interface{}

	_jsii_.Invoke(
		s,
		"getAnyMapAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (s *jsiiProxy_SqlDatabaseInstance) GetBooleanAttribute(terraformAttribute *string) cdktf.IResolvable {
	if err := s.validateGetBooleanAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		s,
		"getBooleanAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (s *jsiiProxy_SqlDatabaseInstance) GetBooleanMapAttribute(terraformAttribute *string) *map[string]*bool {
	if err := s.validateGetBooleanMapAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *map[string]*bool

	_jsii_.Invoke(
		s,
		"getBooleanMapAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (s *jsiiProxy_SqlDatabaseInstance) GetListAttribute(terraformAttribute *string) *[]*string {
	if err := s.validateGetListAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *[]*string

	_jsii_.Invoke(
		s,
		"getListAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (s *jsiiProxy_SqlDatabaseInstance) GetNumberAttribute(terraformAttribute *string) *float64 {
	if err := s.validateGetNumberAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *float64

	_jsii_.Invoke(
		s,
		"getNumberAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (s *jsiiProxy_SqlDatabaseInstance) GetNumberListAttribute(terraformAttribute *string) *[]*float64 {
	if err := s.validateGetNumberListAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *[]*float64

	_jsii_.Invoke(
		s,
		"getNumberListAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (s *jsiiProxy_SqlDatabaseInstance) GetNumberMapAttribute(terraformAttribute *string) *map[string]*float64 {
	if err := s.validateGetNumberMapAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *map[string]*float64

	_jsii_.Invoke(
		s,
		"getNumberMapAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (s *jsiiProxy_SqlDatabaseInstance) GetStringAttribute(terraformAttribute *string) *string {
	if err := s.validateGetStringAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.Invoke(
		s,
		"getStringAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (s *jsiiProxy_SqlDatabaseInstance) GetStringMapAttribute(terraformAttribute *string) *map[string]*string {
	if err := s.validateGetStringMapAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *map[string]*string

	_jsii_.Invoke(
		s,
		"getStringMapAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (s *jsiiProxy_SqlDatabaseInstance) InterpolationForAttribute(terraformAttribute *string) cdktf.IResolvable {
	if err := s.validateInterpolationForAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		s,
		"interpolationForAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (s *jsiiProxy_SqlDatabaseInstance) OverrideLogicalId(newLogicalId *string) {
	if err := s.validateOverrideLogicalIdParameters(newLogicalId); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		s,
		"overrideLogicalId",
		[]interface{}{newLogicalId},
	)
}

func (s *jsiiProxy_SqlDatabaseInstance) PutClone(value *SqlDatabaseInstanceClone) {
	if err := s.validatePutCloneParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		s,
		"putClone",
		[]interface{}{value},
	)
}

func (s *jsiiProxy_SqlDatabaseInstance) PutReplicaConfiguration(value *SqlDatabaseInstanceReplicaConfiguration) {
	if err := s.validatePutReplicaConfigurationParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		s,
		"putReplicaConfiguration",
		[]interface{}{value},
	)
}

func (s *jsiiProxy_SqlDatabaseInstance) PutRestoreBackupContext(value *SqlDatabaseInstanceRestoreBackupContext) {
	if err := s.validatePutRestoreBackupContextParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		s,
		"putRestoreBackupContext",
		[]interface{}{value},
	)
}

func (s *jsiiProxy_SqlDatabaseInstance) PutSettings(value *SqlDatabaseInstanceSettings) {
	if err := s.validatePutSettingsParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		s,
		"putSettings",
		[]interface{}{value},
	)
}

func (s *jsiiProxy_SqlDatabaseInstance) PutTimeouts(value *SqlDatabaseInstanceTimeouts) {
	if err := s.validatePutTimeoutsParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		s,
		"putTimeouts",
		[]interface{}{value},
	)
}

func (s *jsiiProxy_SqlDatabaseInstance) ResetClone() {
	_jsii_.InvokeVoid(
		s,
		"resetClone",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstance) ResetDeletionProtection() {
	_jsii_.InvokeVoid(
		s,
		"resetDeletionProtection",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstance) ResetEncryptionKeyName() {
	_jsii_.InvokeVoid(
		s,
		"resetEncryptionKeyName",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstance) ResetId() {
	_jsii_.InvokeVoid(
		s,
		"resetId",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstance) ResetInstanceType() {
	_jsii_.InvokeVoid(
		s,
		"resetInstanceType",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstance) ResetMaintenanceVersion() {
	_jsii_.InvokeVoid(
		s,
		"resetMaintenanceVersion",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstance) ResetMasterInstanceName() {
	_jsii_.InvokeVoid(
		s,
		"resetMasterInstanceName",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstance) ResetName() {
	_jsii_.InvokeVoid(
		s,
		"resetName",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstance) ResetOverrideLogicalId() {
	_jsii_.InvokeVoid(
		s,
		"resetOverrideLogicalId",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstance) ResetProject() {
	_jsii_.InvokeVoid(
		s,
		"resetProject",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstance) ResetRegion() {
	_jsii_.InvokeVoid(
		s,
		"resetRegion",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstance) ResetReplicaConfiguration() {
	_jsii_.InvokeVoid(
		s,
		"resetReplicaConfiguration",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstance) ResetRestoreBackupContext() {
	_jsii_.InvokeVoid(
		s,
		"resetRestoreBackupContext",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstance) ResetRootPassword() {
	_jsii_.InvokeVoid(
		s,
		"resetRootPassword",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstance) ResetSettings() {
	_jsii_.InvokeVoid(
		s,
		"resetSettings",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstance) ResetTimeouts() {
	_jsii_.InvokeVoid(
		s,
		"resetTimeouts",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstance) SynthesizeAttributes() *map[string]interface{} {
	var returns *map[string]interface{}

	_jsii_.Invoke(
		s,
		"synthesizeAttributes",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (s *jsiiProxy_SqlDatabaseInstance) ToMetadata() interface{} {
	var returns interface{}

	_jsii_.Invoke(
		s,
		"toMetadata",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (s *jsiiProxy_SqlDatabaseInstance) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		s,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (s *jsiiProxy_SqlDatabaseInstance) ToTerraform() interface{} {
	var returns interface{}

	_jsii_.Invoke(
		s,
		"toTerraform",
		nil, // no parameters
		&returns,
	)

	return returns
}

