package datastreamconnectionprofile

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/sourcegraph/managed-services-platform-cdktf/gen/google/jsii"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/datastreamconnectionprofile/internal"
)

// Represents a {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/datastream_connection_profile google_datastream_connection_profile}.
type DatastreamConnectionProfile interface {
	cdktf.TerraformResource
	BigqueryProfile() DatastreamConnectionProfileBigqueryProfileOutputReference
	BigqueryProfileInput() *DatastreamConnectionProfileBigqueryProfile
	// Experimental.
	CdktfStack() cdktf.TerraformStack
	// Experimental.
	Connection() interface{}
	// Experimental.
	SetConnection(val interface{})
	ConnectionProfileId() *string
	SetConnectionProfileId(val *string)
	ConnectionProfileIdInput() *string
	// Experimental.
	ConstructNodeMetadata() *map[string]interface{}
	// Experimental.
	Count() interface{}
	// Experimental.
	SetCount(val interface{})
	// Experimental.
	DependsOn() *[]*string
	// Experimental.
	SetDependsOn(val *[]*string)
	DisplayName() *string
	SetDisplayName(val *string)
	DisplayNameInput() *string
	EffectiveLabels() cdktf.StringMap
	// Experimental.
	ForEach() cdktf.ITerraformIterator
	// Experimental.
	SetForEach(val cdktf.ITerraformIterator)
	ForwardSshConnectivity() DatastreamConnectionProfileForwardSshConnectivityOutputReference
	ForwardSshConnectivityInput() *DatastreamConnectionProfileForwardSshConnectivity
	// Experimental.
	Fqn() *string
	// Experimental.
	FriendlyUniqueId() *string
	GcsProfile() DatastreamConnectionProfileGcsProfileOutputReference
	GcsProfileInput() *DatastreamConnectionProfileGcsProfile
	Id() *string
	SetId(val *string)
	IdInput() *string
	Labels() *map[string]*string
	SetLabels(val *map[string]*string)
	LabelsInput() *map[string]*string
	// Experimental.
	Lifecycle() *cdktf.TerraformResourceLifecycle
	// Experimental.
	SetLifecycle(val *cdktf.TerraformResourceLifecycle)
	Location() *string
	SetLocation(val *string)
	LocationInput() *string
	MysqlProfile() DatastreamConnectionProfileMysqlProfileOutputReference
	MysqlProfileInput() *DatastreamConnectionProfileMysqlProfile
	Name() *string
	// The tree node.
	Node() constructs.Node
	OracleProfile() DatastreamConnectionProfileOracleProfileOutputReference
	OracleProfileInput() *DatastreamConnectionProfileOracleProfile
	PostgresqlProfile() DatastreamConnectionProfilePostgresqlProfileOutputReference
	PostgresqlProfileInput() *DatastreamConnectionProfilePostgresqlProfile
	PrivateConnectivity() DatastreamConnectionProfilePrivateConnectivityOutputReference
	PrivateConnectivityInput() *DatastreamConnectionProfilePrivateConnectivity
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
	// Experimental.
	RawOverrides() interface{}
	// Experimental.
	TerraformGeneratorMetadata() *cdktf.TerraformProviderGeneratorMetadata
	TerraformLabels() cdktf.StringMap
	// Experimental.
	TerraformMetaArguments() *map[string]interface{}
	// Experimental.
	TerraformResourceType() *string
	Timeouts() DatastreamConnectionProfileTimeoutsOutputReference
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
	PutBigqueryProfile(value *DatastreamConnectionProfileBigqueryProfile)
	PutForwardSshConnectivity(value *DatastreamConnectionProfileForwardSshConnectivity)
	PutGcsProfile(value *DatastreamConnectionProfileGcsProfile)
	PutMysqlProfile(value *DatastreamConnectionProfileMysqlProfile)
	PutOracleProfile(value *DatastreamConnectionProfileOracleProfile)
	PutPostgresqlProfile(value *DatastreamConnectionProfilePostgresqlProfile)
	PutPrivateConnectivity(value *DatastreamConnectionProfilePrivateConnectivity)
	PutTimeouts(value *DatastreamConnectionProfileTimeouts)
	ResetBigqueryProfile()
	ResetForwardSshConnectivity()
	ResetGcsProfile()
	ResetId()
	ResetLabels()
	ResetMysqlProfile()
	ResetOracleProfile()
	// Resets a previously passed logical Id to use the auto-generated logical id again.
	// Experimental.
	ResetOverrideLogicalId()
	ResetPostgresqlProfile()
	ResetPrivateConnectivity()
	ResetProject()
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

// The jsii proxy struct for DatastreamConnectionProfile
type jsiiProxy_DatastreamConnectionProfile struct {
	internal.Type__cdktfTerraformResource
}

func (j *jsiiProxy_DatastreamConnectionProfile) BigqueryProfile() DatastreamConnectionProfileBigqueryProfileOutputReference {
	var returns DatastreamConnectionProfileBigqueryProfileOutputReference
	_jsii_.Get(
		j,
		"bigqueryProfile",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_DatastreamConnectionProfile) BigqueryProfileInput() *DatastreamConnectionProfileBigqueryProfile {
	var returns *DatastreamConnectionProfileBigqueryProfile
	_jsii_.Get(
		j,
		"bigqueryProfileInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_DatastreamConnectionProfile) CdktfStack() cdktf.TerraformStack {
	var returns cdktf.TerraformStack
	_jsii_.Get(
		j,
		"cdktfStack",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_DatastreamConnectionProfile) Connection() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"connection",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_DatastreamConnectionProfile) ConnectionProfileId() *string {
	var returns *string
	_jsii_.Get(
		j,
		"connectionProfileId",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_DatastreamConnectionProfile) ConnectionProfileIdInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"connectionProfileIdInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_DatastreamConnectionProfile) ConstructNodeMetadata() *map[string]interface{} {
	var returns *map[string]interface{}
	_jsii_.Get(
		j,
		"constructNodeMetadata",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_DatastreamConnectionProfile) Count() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"count",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_DatastreamConnectionProfile) DependsOn() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"dependsOn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_DatastreamConnectionProfile) DisplayName() *string {
	var returns *string
	_jsii_.Get(
		j,
		"displayName",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_DatastreamConnectionProfile) DisplayNameInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"displayNameInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_DatastreamConnectionProfile) EffectiveLabels() cdktf.StringMap {
	var returns cdktf.StringMap
	_jsii_.Get(
		j,
		"effectiveLabels",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_DatastreamConnectionProfile) ForEach() cdktf.ITerraformIterator {
	var returns cdktf.ITerraformIterator
	_jsii_.Get(
		j,
		"forEach",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_DatastreamConnectionProfile) ForwardSshConnectivity() DatastreamConnectionProfileForwardSshConnectivityOutputReference {
	var returns DatastreamConnectionProfileForwardSshConnectivityOutputReference
	_jsii_.Get(
		j,
		"forwardSshConnectivity",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_DatastreamConnectionProfile) ForwardSshConnectivityInput() *DatastreamConnectionProfileForwardSshConnectivity {
	var returns *DatastreamConnectionProfileForwardSshConnectivity
	_jsii_.Get(
		j,
		"forwardSshConnectivityInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_DatastreamConnectionProfile) Fqn() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fqn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_DatastreamConnectionProfile) FriendlyUniqueId() *string {
	var returns *string
	_jsii_.Get(
		j,
		"friendlyUniqueId",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_DatastreamConnectionProfile) GcsProfile() DatastreamConnectionProfileGcsProfileOutputReference {
	var returns DatastreamConnectionProfileGcsProfileOutputReference
	_jsii_.Get(
		j,
		"gcsProfile",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_DatastreamConnectionProfile) GcsProfileInput() *DatastreamConnectionProfileGcsProfile {
	var returns *DatastreamConnectionProfileGcsProfile
	_jsii_.Get(
		j,
		"gcsProfileInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_DatastreamConnectionProfile) Id() *string {
	var returns *string
	_jsii_.Get(
		j,
		"id",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_DatastreamConnectionProfile) IdInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"idInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_DatastreamConnectionProfile) Labels() *map[string]*string {
	var returns *map[string]*string
	_jsii_.Get(
		j,
		"labels",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_DatastreamConnectionProfile) LabelsInput() *map[string]*string {
	var returns *map[string]*string
	_jsii_.Get(
		j,
		"labelsInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_DatastreamConnectionProfile) Lifecycle() *cdktf.TerraformResourceLifecycle {
	var returns *cdktf.TerraformResourceLifecycle
	_jsii_.Get(
		j,
		"lifecycle",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_DatastreamConnectionProfile) Location() *string {
	var returns *string
	_jsii_.Get(
		j,
		"location",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_DatastreamConnectionProfile) LocationInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"locationInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_DatastreamConnectionProfile) MysqlProfile() DatastreamConnectionProfileMysqlProfileOutputReference {
	var returns DatastreamConnectionProfileMysqlProfileOutputReference
	_jsii_.Get(
		j,
		"mysqlProfile",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_DatastreamConnectionProfile) MysqlProfileInput() *DatastreamConnectionProfileMysqlProfile {
	var returns *DatastreamConnectionProfileMysqlProfile
	_jsii_.Get(
		j,
		"mysqlProfileInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_DatastreamConnectionProfile) Name() *string {
	var returns *string
	_jsii_.Get(
		j,
		"name",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_DatastreamConnectionProfile) Node() constructs.Node {
	var returns constructs.Node
	_jsii_.Get(
		j,
		"node",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_DatastreamConnectionProfile) OracleProfile() DatastreamConnectionProfileOracleProfileOutputReference {
	var returns DatastreamConnectionProfileOracleProfileOutputReference
	_jsii_.Get(
		j,
		"oracleProfile",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_DatastreamConnectionProfile) OracleProfileInput() *DatastreamConnectionProfileOracleProfile {
	var returns *DatastreamConnectionProfileOracleProfile
	_jsii_.Get(
		j,
		"oracleProfileInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_DatastreamConnectionProfile) PostgresqlProfile() DatastreamConnectionProfilePostgresqlProfileOutputReference {
	var returns DatastreamConnectionProfilePostgresqlProfileOutputReference
	_jsii_.Get(
		j,
		"postgresqlProfile",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_DatastreamConnectionProfile) PostgresqlProfileInput() *DatastreamConnectionProfilePostgresqlProfile {
	var returns *DatastreamConnectionProfilePostgresqlProfile
	_jsii_.Get(
		j,
		"postgresqlProfileInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_DatastreamConnectionProfile) PrivateConnectivity() DatastreamConnectionProfilePrivateConnectivityOutputReference {
	var returns DatastreamConnectionProfilePrivateConnectivityOutputReference
	_jsii_.Get(
		j,
		"privateConnectivity",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_DatastreamConnectionProfile) PrivateConnectivityInput() *DatastreamConnectionProfilePrivateConnectivity {
	var returns *DatastreamConnectionProfilePrivateConnectivity
	_jsii_.Get(
		j,
		"privateConnectivityInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_DatastreamConnectionProfile) Project() *string {
	var returns *string
	_jsii_.Get(
		j,
		"project",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_DatastreamConnectionProfile) ProjectInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"projectInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_DatastreamConnectionProfile) Provider() cdktf.TerraformProvider {
	var returns cdktf.TerraformProvider
	_jsii_.Get(
		j,
		"provider",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_DatastreamConnectionProfile) Provisioners() *[]interface{} {
	var returns *[]interface{}
	_jsii_.Get(
		j,
		"provisioners",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_DatastreamConnectionProfile) RawOverrides() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"rawOverrides",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_DatastreamConnectionProfile) TerraformGeneratorMetadata() *cdktf.TerraformProviderGeneratorMetadata {
	var returns *cdktf.TerraformProviderGeneratorMetadata
	_jsii_.Get(
		j,
		"terraformGeneratorMetadata",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_DatastreamConnectionProfile) TerraformLabels() cdktf.StringMap {
	var returns cdktf.StringMap
	_jsii_.Get(
		j,
		"terraformLabels",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_DatastreamConnectionProfile) TerraformMetaArguments() *map[string]interface{} {
	var returns *map[string]interface{}
	_jsii_.Get(
		j,
		"terraformMetaArguments",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_DatastreamConnectionProfile) TerraformResourceType() *string {
	var returns *string
	_jsii_.Get(
		j,
		"terraformResourceType",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_DatastreamConnectionProfile) Timeouts() DatastreamConnectionProfileTimeoutsOutputReference {
	var returns DatastreamConnectionProfileTimeoutsOutputReference
	_jsii_.Get(
		j,
		"timeouts",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_DatastreamConnectionProfile) TimeoutsInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"timeoutsInput",
		&returns,
	)
	return returns
}


// Create a new {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/datastream_connection_profile google_datastream_connection_profile} Resource.
func NewDatastreamConnectionProfile(scope constructs.Construct, id *string, config *DatastreamConnectionProfileConfig) DatastreamConnectionProfile {
	_init_.Initialize()

	if err := validateNewDatastreamConnectionProfileParameters(scope, id, config); err != nil {
		panic(err)
	}
	j := jsiiProxy_DatastreamConnectionProfile{}

	_jsii_.Create(
		"@cdktf/provider-google.datastreamConnectionProfile.DatastreamConnectionProfile",
		[]interface{}{scope, id, config},
		&j,
	)

	return &j
}

// Create a new {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/datastream_connection_profile google_datastream_connection_profile} Resource.
func NewDatastreamConnectionProfile_Override(d DatastreamConnectionProfile, scope constructs.Construct, id *string, config *DatastreamConnectionProfileConfig) {
	_init_.Initialize()

	_jsii_.Create(
		"@cdktf/provider-google.datastreamConnectionProfile.DatastreamConnectionProfile",
		[]interface{}{scope, id, config},
		d,
	)
}

func (j *jsiiProxy_DatastreamConnectionProfile)SetConnection(val interface{}) {
	if err := j.validateSetConnectionParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"connection",
		val,
	)
}

func (j *jsiiProxy_DatastreamConnectionProfile)SetConnectionProfileId(val *string) {
	if err := j.validateSetConnectionProfileIdParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"connectionProfileId",
		val,
	)
}

func (j *jsiiProxy_DatastreamConnectionProfile)SetCount(val interface{}) {
	if err := j.validateSetCountParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"count",
		val,
	)
}

func (j *jsiiProxy_DatastreamConnectionProfile)SetDependsOn(val *[]*string) {
	_jsii_.Set(
		j,
		"dependsOn",
		val,
	)
}

func (j *jsiiProxy_DatastreamConnectionProfile)SetDisplayName(val *string) {
	if err := j.validateSetDisplayNameParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"displayName",
		val,
	)
}

func (j *jsiiProxy_DatastreamConnectionProfile)SetForEach(val cdktf.ITerraformIterator) {
	_jsii_.Set(
		j,
		"forEach",
		val,
	)
}

func (j *jsiiProxy_DatastreamConnectionProfile)SetId(val *string) {
	if err := j.validateSetIdParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"id",
		val,
	)
}

func (j *jsiiProxy_DatastreamConnectionProfile)SetLabels(val *map[string]*string) {
	if err := j.validateSetLabelsParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"labels",
		val,
	)
}

func (j *jsiiProxy_DatastreamConnectionProfile)SetLifecycle(val *cdktf.TerraformResourceLifecycle) {
	if err := j.validateSetLifecycleParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"lifecycle",
		val,
	)
}

func (j *jsiiProxy_DatastreamConnectionProfile)SetLocation(val *string) {
	if err := j.validateSetLocationParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"location",
		val,
	)
}

func (j *jsiiProxy_DatastreamConnectionProfile)SetProject(val *string) {
	if err := j.validateSetProjectParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"project",
		val,
	)
}

func (j *jsiiProxy_DatastreamConnectionProfile)SetProvider(val cdktf.TerraformProvider) {
	_jsii_.Set(
		j,
		"provider",
		val,
	)
}

func (j *jsiiProxy_DatastreamConnectionProfile)SetProvisioners(val *[]interface{}) {
	if err := j.validateSetProvisionersParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"provisioners",
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
func DatastreamConnectionProfile_IsConstruct(x interface{}) *bool {
	_init_.Initialize()

	if err := validateDatastreamConnectionProfile_IsConstructParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"@cdktf/provider-google.datastreamConnectionProfile.DatastreamConnectionProfile",
		"isConstruct",
		[]interface{}{x},
		&returns,
	)

	return returns
}

// Experimental.
func DatastreamConnectionProfile_IsTerraformElement(x interface{}) *bool {
	_init_.Initialize()

	if err := validateDatastreamConnectionProfile_IsTerraformElementParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"@cdktf/provider-google.datastreamConnectionProfile.DatastreamConnectionProfile",
		"isTerraformElement",
		[]interface{}{x},
		&returns,
	)

	return returns
}

// Experimental.
func DatastreamConnectionProfile_IsTerraformResource(x interface{}) *bool {
	_init_.Initialize()

	if err := validateDatastreamConnectionProfile_IsTerraformResourceParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"@cdktf/provider-google.datastreamConnectionProfile.DatastreamConnectionProfile",
		"isTerraformResource",
		[]interface{}{x},
		&returns,
	)

	return returns
}

func DatastreamConnectionProfile_TfResourceType() *string {
	_init_.Initialize()
	var returns *string
	_jsii_.StaticGet(
		"@cdktf/provider-google.datastreamConnectionProfile.DatastreamConnectionProfile",
		"tfResourceType",
		&returns,
	)
	return returns
}

func (d *jsiiProxy_DatastreamConnectionProfile) AddOverride(path *string, value interface{}) {
	if err := d.validateAddOverrideParameters(path, value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		d,
		"addOverride",
		[]interface{}{path, value},
	)
}

func (d *jsiiProxy_DatastreamConnectionProfile) GetAnyMapAttribute(terraformAttribute *string) *map[string]interface{} {
	if err := d.validateGetAnyMapAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *map[string]interface{}

	_jsii_.Invoke(
		d,
		"getAnyMapAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (d *jsiiProxy_DatastreamConnectionProfile) GetBooleanAttribute(terraformAttribute *string) cdktf.IResolvable {
	if err := d.validateGetBooleanAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		d,
		"getBooleanAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (d *jsiiProxy_DatastreamConnectionProfile) GetBooleanMapAttribute(terraformAttribute *string) *map[string]*bool {
	if err := d.validateGetBooleanMapAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *map[string]*bool

	_jsii_.Invoke(
		d,
		"getBooleanMapAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (d *jsiiProxy_DatastreamConnectionProfile) GetListAttribute(terraformAttribute *string) *[]*string {
	if err := d.validateGetListAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *[]*string

	_jsii_.Invoke(
		d,
		"getListAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (d *jsiiProxy_DatastreamConnectionProfile) GetNumberAttribute(terraformAttribute *string) *float64 {
	if err := d.validateGetNumberAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *float64

	_jsii_.Invoke(
		d,
		"getNumberAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (d *jsiiProxy_DatastreamConnectionProfile) GetNumberListAttribute(terraformAttribute *string) *[]*float64 {
	if err := d.validateGetNumberListAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *[]*float64

	_jsii_.Invoke(
		d,
		"getNumberListAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (d *jsiiProxy_DatastreamConnectionProfile) GetNumberMapAttribute(terraformAttribute *string) *map[string]*float64 {
	if err := d.validateGetNumberMapAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *map[string]*float64

	_jsii_.Invoke(
		d,
		"getNumberMapAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (d *jsiiProxy_DatastreamConnectionProfile) GetStringAttribute(terraformAttribute *string) *string {
	if err := d.validateGetStringAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.Invoke(
		d,
		"getStringAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (d *jsiiProxy_DatastreamConnectionProfile) GetStringMapAttribute(terraformAttribute *string) *map[string]*string {
	if err := d.validateGetStringMapAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *map[string]*string

	_jsii_.Invoke(
		d,
		"getStringMapAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (d *jsiiProxy_DatastreamConnectionProfile) InterpolationForAttribute(terraformAttribute *string) cdktf.IResolvable {
	if err := d.validateInterpolationForAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		d,
		"interpolationForAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (d *jsiiProxy_DatastreamConnectionProfile) OverrideLogicalId(newLogicalId *string) {
	if err := d.validateOverrideLogicalIdParameters(newLogicalId); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		d,
		"overrideLogicalId",
		[]interface{}{newLogicalId},
	)
}

func (d *jsiiProxy_DatastreamConnectionProfile) PutBigqueryProfile(value *DatastreamConnectionProfileBigqueryProfile) {
	if err := d.validatePutBigqueryProfileParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		d,
		"putBigqueryProfile",
		[]interface{}{value},
	)
}

func (d *jsiiProxy_DatastreamConnectionProfile) PutForwardSshConnectivity(value *DatastreamConnectionProfileForwardSshConnectivity) {
	if err := d.validatePutForwardSshConnectivityParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		d,
		"putForwardSshConnectivity",
		[]interface{}{value},
	)
}

func (d *jsiiProxy_DatastreamConnectionProfile) PutGcsProfile(value *DatastreamConnectionProfileGcsProfile) {
	if err := d.validatePutGcsProfileParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		d,
		"putGcsProfile",
		[]interface{}{value},
	)
}

func (d *jsiiProxy_DatastreamConnectionProfile) PutMysqlProfile(value *DatastreamConnectionProfileMysqlProfile) {
	if err := d.validatePutMysqlProfileParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		d,
		"putMysqlProfile",
		[]interface{}{value},
	)
}

func (d *jsiiProxy_DatastreamConnectionProfile) PutOracleProfile(value *DatastreamConnectionProfileOracleProfile) {
	if err := d.validatePutOracleProfileParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		d,
		"putOracleProfile",
		[]interface{}{value},
	)
}

func (d *jsiiProxy_DatastreamConnectionProfile) PutPostgresqlProfile(value *DatastreamConnectionProfilePostgresqlProfile) {
	if err := d.validatePutPostgresqlProfileParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		d,
		"putPostgresqlProfile",
		[]interface{}{value},
	)
}

func (d *jsiiProxy_DatastreamConnectionProfile) PutPrivateConnectivity(value *DatastreamConnectionProfilePrivateConnectivity) {
	if err := d.validatePutPrivateConnectivityParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		d,
		"putPrivateConnectivity",
		[]interface{}{value},
	)
}

func (d *jsiiProxy_DatastreamConnectionProfile) PutTimeouts(value *DatastreamConnectionProfileTimeouts) {
	if err := d.validatePutTimeoutsParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		d,
		"putTimeouts",
		[]interface{}{value},
	)
}

func (d *jsiiProxy_DatastreamConnectionProfile) ResetBigqueryProfile() {
	_jsii_.InvokeVoid(
		d,
		"resetBigqueryProfile",
		nil, // no parameters
	)
}

func (d *jsiiProxy_DatastreamConnectionProfile) ResetForwardSshConnectivity() {
	_jsii_.InvokeVoid(
		d,
		"resetForwardSshConnectivity",
		nil, // no parameters
	)
}

func (d *jsiiProxy_DatastreamConnectionProfile) ResetGcsProfile() {
	_jsii_.InvokeVoid(
		d,
		"resetGcsProfile",
		nil, // no parameters
	)
}

func (d *jsiiProxy_DatastreamConnectionProfile) ResetId() {
	_jsii_.InvokeVoid(
		d,
		"resetId",
		nil, // no parameters
	)
}

func (d *jsiiProxy_DatastreamConnectionProfile) ResetLabels() {
	_jsii_.InvokeVoid(
		d,
		"resetLabels",
		nil, // no parameters
	)
}

func (d *jsiiProxy_DatastreamConnectionProfile) ResetMysqlProfile() {
	_jsii_.InvokeVoid(
		d,
		"resetMysqlProfile",
		nil, // no parameters
	)
}

func (d *jsiiProxy_DatastreamConnectionProfile) ResetOracleProfile() {
	_jsii_.InvokeVoid(
		d,
		"resetOracleProfile",
		nil, // no parameters
	)
}

func (d *jsiiProxy_DatastreamConnectionProfile) ResetOverrideLogicalId() {
	_jsii_.InvokeVoid(
		d,
		"resetOverrideLogicalId",
		nil, // no parameters
	)
}

func (d *jsiiProxy_DatastreamConnectionProfile) ResetPostgresqlProfile() {
	_jsii_.InvokeVoid(
		d,
		"resetPostgresqlProfile",
		nil, // no parameters
	)
}

func (d *jsiiProxy_DatastreamConnectionProfile) ResetPrivateConnectivity() {
	_jsii_.InvokeVoid(
		d,
		"resetPrivateConnectivity",
		nil, // no parameters
	)
}

func (d *jsiiProxy_DatastreamConnectionProfile) ResetProject() {
	_jsii_.InvokeVoid(
		d,
		"resetProject",
		nil, // no parameters
	)
}

func (d *jsiiProxy_DatastreamConnectionProfile) ResetTimeouts() {
	_jsii_.InvokeVoid(
		d,
		"resetTimeouts",
		nil, // no parameters
	)
}

func (d *jsiiProxy_DatastreamConnectionProfile) SynthesizeAttributes() *map[string]interface{} {
	var returns *map[string]interface{}

	_jsii_.Invoke(
		d,
		"synthesizeAttributes",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (d *jsiiProxy_DatastreamConnectionProfile) ToMetadata() interface{} {
	var returns interface{}

	_jsii_.Invoke(
		d,
		"toMetadata",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (d *jsiiProxy_DatastreamConnectionProfile) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		d,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (d *jsiiProxy_DatastreamConnectionProfile) ToTerraform() interface{} {
	var returns interface{}

	_jsii_.Invoke(
		d,
		"toTerraform",
		nil, // no parameters
		&returns,
	)

	return returns
}

