package redisinstance

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/sourcegraph/managed-services-platform-cdktf/gen/google/jsii"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/redisinstance/internal"
)

// Represents a {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/redis_instance google_redis_instance}.
type RedisInstance interface {
	cdktf.TerraformResource
	AlternativeLocationId() *string
	SetAlternativeLocationId(val *string)
	AlternativeLocationIdInput() *string
	AuthEnabled() interface{}
	SetAuthEnabled(val interface{})
	AuthEnabledInput() interface{}
	AuthorizedNetwork() *string
	SetAuthorizedNetwork(val *string)
	AuthorizedNetworkInput() *string
	AuthString() *string
	// Experimental.
	CdktfStack() cdktf.TerraformStack
	// Experimental.
	Connection() interface{}
	// Experimental.
	SetConnection(val interface{})
	ConnectMode() *string
	SetConnectMode(val *string)
	ConnectModeInput() *string
	// Experimental.
	ConstructNodeMetadata() *map[string]interface{}
	// Experimental.
	Count() interface{}
	// Experimental.
	SetCount(val interface{})
	CreateTime() *string
	CurrentLocationId() *string
	CustomerManagedKey() *string
	SetCustomerManagedKey(val *string)
	CustomerManagedKeyInput() *string
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
	// Experimental.
	Fqn() *string
	// Experimental.
	FriendlyUniqueId() *string
	Host() *string
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
	LocationId() *string
	SetLocationId(val *string)
	LocationIdInput() *string
	MaintenancePolicy() RedisInstanceMaintenancePolicyOutputReference
	MaintenancePolicyInput() *RedisInstanceMaintenancePolicy
	MaintenanceSchedule() RedisInstanceMaintenanceScheduleList
	MemorySizeGb() *float64
	SetMemorySizeGb(val *float64)
	MemorySizeGbInput() *float64
	Name() *string
	SetName(val *string)
	NameInput() *string
	// The tree node.
	Node() constructs.Node
	Nodes() RedisInstanceNodesList
	PersistenceConfig() RedisInstancePersistenceConfigOutputReference
	PersistenceConfigInput() *RedisInstancePersistenceConfig
	PersistenceIamIdentity() *string
	Port() *float64
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
	ReadEndpoint() *string
	ReadEndpointPort() *float64
	ReadReplicasMode() *string
	SetReadReplicasMode(val *string)
	ReadReplicasModeInput() *string
	RedisConfigs() *map[string]*string
	SetRedisConfigs(val *map[string]*string)
	RedisConfigsInput() *map[string]*string
	RedisVersion() *string
	SetRedisVersion(val *string)
	RedisVersionInput() *string
	Region() *string
	SetRegion(val *string)
	RegionInput() *string
	ReplicaCount() *float64
	SetReplicaCount(val *float64)
	ReplicaCountInput() *float64
	ReservedIpRange() *string
	SetReservedIpRange(val *string)
	ReservedIpRangeInput() *string
	SecondaryIpRange() *string
	SetSecondaryIpRange(val *string)
	SecondaryIpRangeInput() *string
	ServerCaCerts() RedisInstanceServerCaCertsList
	// Experimental.
	TerraformGeneratorMetadata() *cdktf.TerraformProviderGeneratorMetadata
	TerraformLabels() cdktf.StringMap
	// Experimental.
	TerraformMetaArguments() *map[string]interface{}
	// Experimental.
	TerraformResourceType() *string
	Tier() *string
	SetTier(val *string)
	TierInput() *string
	Timeouts() RedisInstanceTimeoutsOutputReference
	TimeoutsInput() interface{}
	TransitEncryptionMode() *string
	SetTransitEncryptionMode(val *string)
	TransitEncryptionModeInput() *string
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
	PutMaintenancePolicy(value *RedisInstanceMaintenancePolicy)
	PutPersistenceConfig(value *RedisInstancePersistenceConfig)
	PutTimeouts(value *RedisInstanceTimeouts)
	ResetAlternativeLocationId()
	ResetAuthEnabled()
	ResetAuthorizedNetwork()
	ResetConnectMode()
	ResetCustomerManagedKey()
	ResetDisplayName()
	ResetId()
	ResetLabels()
	ResetLocationId()
	ResetMaintenancePolicy()
	// Resets a previously passed logical Id to use the auto-generated logical id again.
	// Experimental.
	ResetOverrideLogicalId()
	ResetPersistenceConfig()
	ResetProject()
	ResetReadReplicasMode()
	ResetRedisConfigs()
	ResetRedisVersion()
	ResetRegion()
	ResetReplicaCount()
	ResetReservedIpRange()
	ResetSecondaryIpRange()
	ResetTier()
	ResetTimeouts()
	ResetTransitEncryptionMode()
	SynthesizeAttributes() *map[string]interface{}
	// Experimental.
	ToMetadata() interface{}
	// Returns a string representation of this construct.
	ToString() *string
	// Adds this resource to the terraform JSON output.
	// Experimental.
	ToTerraform() interface{}
}

// The jsii proxy struct for RedisInstance
type jsiiProxy_RedisInstance struct {
	internal.Type__cdktfTerraformResource
}

func (j *jsiiProxy_RedisInstance) AlternativeLocationId() *string {
	var returns *string
	_jsii_.Get(
		j,
		"alternativeLocationId",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) AlternativeLocationIdInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"alternativeLocationIdInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) AuthEnabled() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"authEnabled",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) AuthEnabledInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"authEnabledInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) AuthorizedNetwork() *string {
	var returns *string
	_jsii_.Get(
		j,
		"authorizedNetwork",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) AuthorizedNetworkInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"authorizedNetworkInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) AuthString() *string {
	var returns *string
	_jsii_.Get(
		j,
		"authString",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) CdktfStack() cdktf.TerraformStack {
	var returns cdktf.TerraformStack
	_jsii_.Get(
		j,
		"cdktfStack",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) Connection() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"connection",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) ConnectMode() *string {
	var returns *string
	_jsii_.Get(
		j,
		"connectMode",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) ConnectModeInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"connectModeInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) ConstructNodeMetadata() *map[string]interface{} {
	var returns *map[string]interface{}
	_jsii_.Get(
		j,
		"constructNodeMetadata",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) Count() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"count",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) CreateTime() *string {
	var returns *string
	_jsii_.Get(
		j,
		"createTime",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) CurrentLocationId() *string {
	var returns *string
	_jsii_.Get(
		j,
		"currentLocationId",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) CustomerManagedKey() *string {
	var returns *string
	_jsii_.Get(
		j,
		"customerManagedKey",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) CustomerManagedKeyInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"customerManagedKeyInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) DependsOn() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"dependsOn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) DisplayName() *string {
	var returns *string
	_jsii_.Get(
		j,
		"displayName",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) DisplayNameInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"displayNameInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) EffectiveLabels() cdktf.StringMap {
	var returns cdktf.StringMap
	_jsii_.Get(
		j,
		"effectiveLabels",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) ForEach() cdktf.ITerraformIterator {
	var returns cdktf.ITerraformIterator
	_jsii_.Get(
		j,
		"forEach",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) Fqn() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fqn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) FriendlyUniqueId() *string {
	var returns *string
	_jsii_.Get(
		j,
		"friendlyUniqueId",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) Host() *string {
	var returns *string
	_jsii_.Get(
		j,
		"host",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) Id() *string {
	var returns *string
	_jsii_.Get(
		j,
		"id",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) IdInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"idInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) Labels() *map[string]*string {
	var returns *map[string]*string
	_jsii_.Get(
		j,
		"labels",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) LabelsInput() *map[string]*string {
	var returns *map[string]*string
	_jsii_.Get(
		j,
		"labelsInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) Lifecycle() *cdktf.TerraformResourceLifecycle {
	var returns *cdktf.TerraformResourceLifecycle
	_jsii_.Get(
		j,
		"lifecycle",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) LocationId() *string {
	var returns *string
	_jsii_.Get(
		j,
		"locationId",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) LocationIdInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"locationIdInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) MaintenancePolicy() RedisInstanceMaintenancePolicyOutputReference {
	var returns RedisInstanceMaintenancePolicyOutputReference
	_jsii_.Get(
		j,
		"maintenancePolicy",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) MaintenancePolicyInput() *RedisInstanceMaintenancePolicy {
	var returns *RedisInstanceMaintenancePolicy
	_jsii_.Get(
		j,
		"maintenancePolicyInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) MaintenanceSchedule() RedisInstanceMaintenanceScheduleList {
	var returns RedisInstanceMaintenanceScheduleList
	_jsii_.Get(
		j,
		"maintenanceSchedule",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) MemorySizeGb() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"memorySizeGb",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) MemorySizeGbInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"memorySizeGbInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) Name() *string {
	var returns *string
	_jsii_.Get(
		j,
		"name",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) NameInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"nameInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) Node() constructs.Node {
	var returns constructs.Node
	_jsii_.Get(
		j,
		"node",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) Nodes() RedisInstanceNodesList {
	var returns RedisInstanceNodesList
	_jsii_.Get(
		j,
		"nodes",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) PersistenceConfig() RedisInstancePersistenceConfigOutputReference {
	var returns RedisInstancePersistenceConfigOutputReference
	_jsii_.Get(
		j,
		"persistenceConfig",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) PersistenceConfigInput() *RedisInstancePersistenceConfig {
	var returns *RedisInstancePersistenceConfig
	_jsii_.Get(
		j,
		"persistenceConfigInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) PersistenceIamIdentity() *string {
	var returns *string
	_jsii_.Get(
		j,
		"persistenceIamIdentity",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) Port() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"port",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) Project() *string {
	var returns *string
	_jsii_.Get(
		j,
		"project",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) ProjectInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"projectInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) Provider() cdktf.TerraformProvider {
	var returns cdktf.TerraformProvider
	_jsii_.Get(
		j,
		"provider",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) Provisioners() *[]interface{} {
	var returns *[]interface{}
	_jsii_.Get(
		j,
		"provisioners",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) RawOverrides() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"rawOverrides",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) ReadEndpoint() *string {
	var returns *string
	_jsii_.Get(
		j,
		"readEndpoint",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) ReadEndpointPort() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"readEndpointPort",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) ReadReplicasMode() *string {
	var returns *string
	_jsii_.Get(
		j,
		"readReplicasMode",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) ReadReplicasModeInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"readReplicasModeInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) RedisConfigs() *map[string]*string {
	var returns *map[string]*string
	_jsii_.Get(
		j,
		"redisConfigs",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) RedisConfigsInput() *map[string]*string {
	var returns *map[string]*string
	_jsii_.Get(
		j,
		"redisConfigsInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) RedisVersion() *string {
	var returns *string
	_jsii_.Get(
		j,
		"redisVersion",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) RedisVersionInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"redisVersionInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) Region() *string {
	var returns *string
	_jsii_.Get(
		j,
		"region",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) RegionInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"regionInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) ReplicaCount() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"replicaCount",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) ReplicaCountInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"replicaCountInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) ReservedIpRange() *string {
	var returns *string
	_jsii_.Get(
		j,
		"reservedIpRange",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) ReservedIpRangeInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"reservedIpRangeInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) SecondaryIpRange() *string {
	var returns *string
	_jsii_.Get(
		j,
		"secondaryIpRange",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) SecondaryIpRangeInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"secondaryIpRangeInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) ServerCaCerts() RedisInstanceServerCaCertsList {
	var returns RedisInstanceServerCaCertsList
	_jsii_.Get(
		j,
		"serverCaCerts",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) TerraformGeneratorMetadata() *cdktf.TerraformProviderGeneratorMetadata {
	var returns *cdktf.TerraformProviderGeneratorMetadata
	_jsii_.Get(
		j,
		"terraformGeneratorMetadata",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) TerraformLabels() cdktf.StringMap {
	var returns cdktf.StringMap
	_jsii_.Get(
		j,
		"terraformLabels",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) TerraformMetaArguments() *map[string]interface{} {
	var returns *map[string]interface{}
	_jsii_.Get(
		j,
		"terraformMetaArguments",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) TerraformResourceType() *string {
	var returns *string
	_jsii_.Get(
		j,
		"terraformResourceType",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) Tier() *string {
	var returns *string
	_jsii_.Get(
		j,
		"tier",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) TierInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"tierInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) Timeouts() RedisInstanceTimeoutsOutputReference {
	var returns RedisInstanceTimeoutsOutputReference
	_jsii_.Get(
		j,
		"timeouts",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) TimeoutsInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"timeoutsInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) TransitEncryptionMode() *string {
	var returns *string
	_jsii_.Get(
		j,
		"transitEncryptionMode",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RedisInstance) TransitEncryptionModeInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"transitEncryptionModeInput",
		&returns,
	)
	return returns
}


// Create a new {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/redis_instance google_redis_instance} Resource.
func NewRedisInstance(scope constructs.Construct, id *string, config *RedisInstanceConfig) RedisInstance {
	_init_.Initialize()

	if err := validateNewRedisInstanceParameters(scope, id, config); err != nil {
		panic(err)
	}
	j := jsiiProxy_RedisInstance{}

	_jsii_.Create(
		"@cdktf/provider-google.redisInstance.RedisInstance",
		[]interface{}{scope, id, config},
		&j,
	)

	return &j
}

// Create a new {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/redis_instance google_redis_instance} Resource.
func NewRedisInstance_Override(r RedisInstance, scope constructs.Construct, id *string, config *RedisInstanceConfig) {
	_init_.Initialize()

	_jsii_.Create(
		"@cdktf/provider-google.redisInstance.RedisInstance",
		[]interface{}{scope, id, config},
		r,
	)
}

func (j *jsiiProxy_RedisInstance)SetAlternativeLocationId(val *string) {
	if err := j.validateSetAlternativeLocationIdParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"alternativeLocationId",
		val,
	)
}

func (j *jsiiProxy_RedisInstance)SetAuthEnabled(val interface{}) {
	if err := j.validateSetAuthEnabledParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"authEnabled",
		val,
	)
}

func (j *jsiiProxy_RedisInstance)SetAuthorizedNetwork(val *string) {
	if err := j.validateSetAuthorizedNetworkParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"authorizedNetwork",
		val,
	)
}

func (j *jsiiProxy_RedisInstance)SetConnection(val interface{}) {
	if err := j.validateSetConnectionParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"connection",
		val,
	)
}

func (j *jsiiProxy_RedisInstance)SetConnectMode(val *string) {
	if err := j.validateSetConnectModeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"connectMode",
		val,
	)
}

func (j *jsiiProxy_RedisInstance)SetCount(val interface{}) {
	if err := j.validateSetCountParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"count",
		val,
	)
}

func (j *jsiiProxy_RedisInstance)SetCustomerManagedKey(val *string) {
	if err := j.validateSetCustomerManagedKeyParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"customerManagedKey",
		val,
	)
}

func (j *jsiiProxy_RedisInstance)SetDependsOn(val *[]*string) {
	_jsii_.Set(
		j,
		"dependsOn",
		val,
	)
}

func (j *jsiiProxy_RedisInstance)SetDisplayName(val *string) {
	if err := j.validateSetDisplayNameParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"displayName",
		val,
	)
}

func (j *jsiiProxy_RedisInstance)SetForEach(val cdktf.ITerraformIterator) {
	_jsii_.Set(
		j,
		"forEach",
		val,
	)
}

func (j *jsiiProxy_RedisInstance)SetId(val *string) {
	if err := j.validateSetIdParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"id",
		val,
	)
}

func (j *jsiiProxy_RedisInstance)SetLabels(val *map[string]*string) {
	if err := j.validateSetLabelsParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"labels",
		val,
	)
}

func (j *jsiiProxy_RedisInstance)SetLifecycle(val *cdktf.TerraformResourceLifecycle) {
	if err := j.validateSetLifecycleParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"lifecycle",
		val,
	)
}

func (j *jsiiProxy_RedisInstance)SetLocationId(val *string) {
	if err := j.validateSetLocationIdParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"locationId",
		val,
	)
}

func (j *jsiiProxy_RedisInstance)SetMemorySizeGb(val *float64) {
	if err := j.validateSetMemorySizeGbParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"memorySizeGb",
		val,
	)
}

func (j *jsiiProxy_RedisInstance)SetName(val *string) {
	if err := j.validateSetNameParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"name",
		val,
	)
}

func (j *jsiiProxy_RedisInstance)SetProject(val *string) {
	if err := j.validateSetProjectParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"project",
		val,
	)
}

func (j *jsiiProxy_RedisInstance)SetProvider(val cdktf.TerraformProvider) {
	_jsii_.Set(
		j,
		"provider",
		val,
	)
}

func (j *jsiiProxy_RedisInstance)SetProvisioners(val *[]interface{}) {
	if err := j.validateSetProvisionersParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"provisioners",
		val,
	)
}

func (j *jsiiProxy_RedisInstance)SetReadReplicasMode(val *string) {
	if err := j.validateSetReadReplicasModeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"readReplicasMode",
		val,
	)
}

func (j *jsiiProxy_RedisInstance)SetRedisConfigs(val *map[string]*string) {
	if err := j.validateSetRedisConfigsParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"redisConfigs",
		val,
	)
}

func (j *jsiiProxy_RedisInstance)SetRedisVersion(val *string) {
	if err := j.validateSetRedisVersionParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"redisVersion",
		val,
	)
}

func (j *jsiiProxy_RedisInstance)SetRegion(val *string) {
	if err := j.validateSetRegionParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"region",
		val,
	)
}

func (j *jsiiProxy_RedisInstance)SetReplicaCount(val *float64) {
	if err := j.validateSetReplicaCountParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"replicaCount",
		val,
	)
}

func (j *jsiiProxy_RedisInstance)SetReservedIpRange(val *string) {
	if err := j.validateSetReservedIpRangeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"reservedIpRange",
		val,
	)
}

func (j *jsiiProxy_RedisInstance)SetSecondaryIpRange(val *string) {
	if err := j.validateSetSecondaryIpRangeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"secondaryIpRange",
		val,
	)
}

func (j *jsiiProxy_RedisInstance)SetTier(val *string) {
	if err := j.validateSetTierParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"tier",
		val,
	)
}

func (j *jsiiProxy_RedisInstance)SetTransitEncryptionMode(val *string) {
	if err := j.validateSetTransitEncryptionModeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"transitEncryptionMode",
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
func RedisInstance_IsConstruct(x interface{}) *bool {
	_init_.Initialize()

	if err := validateRedisInstance_IsConstructParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"@cdktf/provider-google.redisInstance.RedisInstance",
		"isConstruct",
		[]interface{}{x},
		&returns,
	)

	return returns
}

// Experimental.
func RedisInstance_IsTerraformElement(x interface{}) *bool {
	_init_.Initialize()

	if err := validateRedisInstance_IsTerraformElementParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"@cdktf/provider-google.redisInstance.RedisInstance",
		"isTerraformElement",
		[]interface{}{x},
		&returns,
	)

	return returns
}

// Experimental.
func RedisInstance_IsTerraformResource(x interface{}) *bool {
	_init_.Initialize()

	if err := validateRedisInstance_IsTerraformResourceParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"@cdktf/provider-google.redisInstance.RedisInstance",
		"isTerraformResource",
		[]interface{}{x},
		&returns,
	)

	return returns
}

func RedisInstance_TfResourceType() *string {
	_init_.Initialize()
	var returns *string
	_jsii_.StaticGet(
		"@cdktf/provider-google.redisInstance.RedisInstance",
		"tfResourceType",
		&returns,
	)
	return returns
}

func (r *jsiiProxy_RedisInstance) AddOverride(path *string, value interface{}) {
	if err := r.validateAddOverrideParameters(path, value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		r,
		"addOverride",
		[]interface{}{path, value},
	)
}

func (r *jsiiProxy_RedisInstance) GetAnyMapAttribute(terraformAttribute *string) *map[string]interface{} {
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

func (r *jsiiProxy_RedisInstance) GetBooleanAttribute(terraformAttribute *string) cdktf.IResolvable {
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

func (r *jsiiProxy_RedisInstance) GetBooleanMapAttribute(terraformAttribute *string) *map[string]*bool {
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

func (r *jsiiProxy_RedisInstance) GetListAttribute(terraformAttribute *string) *[]*string {
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

func (r *jsiiProxy_RedisInstance) GetNumberAttribute(terraformAttribute *string) *float64 {
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

func (r *jsiiProxy_RedisInstance) GetNumberListAttribute(terraformAttribute *string) *[]*float64 {
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

func (r *jsiiProxy_RedisInstance) GetNumberMapAttribute(terraformAttribute *string) *map[string]*float64 {
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

func (r *jsiiProxy_RedisInstance) GetStringAttribute(terraformAttribute *string) *string {
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

func (r *jsiiProxy_RedisInstance) GetStringMapAttribute(terraformAttribute *string) *map[string]*string {
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

func (r *jsiiProxy_RedisInstance) InterpolationForAttribute(terraformAttribute *string) cdktf.IResolvable {
	if err := r.validateInterpolationForAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		r,
		"interpolationForAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (r *jsiiProxy_RedisInstance) OverrideLogicalId(newLogicalId *string) {
	if err := r.validateOverrideLogicalIdParameters(newLogicalId); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		r,
		"overrideLogicalId",
		[]interface{}{newLogicalId},
	)
}

func (r *jsiiProxy_RedisInstance) PutMaintenancePolicy(value *RedisInstanceMaintenancePolicy) {
	if err := r.validatePutMaintenancePolicyParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		r,
		"putMaintenancePolicy",
		[]interface{}{value},
	)
}

func (r *jsiiProxy_RedisInstance) PutPersistenceConfig(value *RedisInstancePersistenceConfig) {
	if err := r.validatePutPersistenceConfigParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		r,
		"putPersistenceConfig",
		[]interface{}{value},
	)
}

func (r *jsiiProxy_RedisInstance) PutTimeouts(value *RedisInstanceTimeouts) {
	if err := r.validatePutTimeoutsParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		r,
		"putTimeouts",
		[]interface{}{value},
	)
}

func (r *jsiiProxy_RedisInstance) ResetAlternativeLocationId() {
	_jsii_.InvokeVoid(
		r,
		"resetAlternativeLocationId",
		nil, // no parameters
	)
}

func (r *jsiiProxy_RedisInstance) ResetAuthEnabled() {
	_jsii_.InvokeVoid(
		r,
		"resetAuthEnabled",
		nil, // no parameters
	)
}

func (r *jsiiProxy_RedisInstance) ResetAuthorizedNetwork() {
	_jsii_.InvokeVoid(
		r,
		"resetAuthorizedNetwork",
		nil, // no parameters
	)
}

func (r *jsiiProxy_RedisInstance) ResetConnectMode() {
	_jsii_.InvokeVoid(
		r,
		"resetConnectMode",
		nil, // no parameters
	)
}

func (r *jsiiProxy_RedisInstance) ResetCustomerManagedKey() {
	_jsii_.InvokeVoid(
		r,
		"resetCustomerManagedKey",
		nil, // no parameters
	)
}

func (r *jsiiProxy_RedisInstance) ResetDisplayName() {
	_jsii_.InvokeVoid(
		r,
		"resetDisplayName",
		nil, // no parameters
	)
}

func (r *jsiiProxy_RedisInstance) ResetId() {
	_jsii_.InvokeVoid(
		r,
		"resetId",
		nil, // no parameters
	)
}

func (r *jsiiProxy_RedisInstance) ResetLabels() {
	_jsii_.InvokeVoid(
		r,
		"resetLabels",
		nil, // no parameters
	)
}

func (r *jsiiProxy_RedisInstance) ResetLocationId() {
	_jsii_.InvokeVoid(
		r,
		"resetLocationId",
		nil, // no parameters
	)
}

func (r *jsiiProxy_RedisInstance) ResetMaintenancePolicy() {
	_jsii_.InvokeVoid(
		r,
		"resetMaintenancePolicy",
		nil, // no parameters
	)
}

func (r *jsiiProxy_RedisInstance) ResetOverrideLogicalId() {
	_jsii_.InvokeVoid(
		r,
		"resetOverrideLogicalId",
		nil, // no parameters
	)
}

func (r *jsiiProxy_RedisInstance) ResetPersistenceConfig() {
	_jsii_.InvokeVoid(
		r,
		"resetPersistenceConfig",
		nil, // no parameters
	)
}

func (r *jsiiProxy_RedisInstance) ResetProject() {
	_jsii_.InvokeVoid(
		r,
		"resetProject",
		nil, // no parameters
	)
}

func (r *jsiiProxy_RedisInstance) ResetReadReplicasMode() {
	_jsii_.InvokeVoid(
		r,
		"resetReadReplicasMode",
		nil, // no parameters
	)
}

func (r *jsiiProxy_RedisInstance) ResetRedisConfigs() {
	_jsii_.InvokeVoid(
		r,
		"resetRedisConfigs",
		nil, // no parameters
	)
}

func (r *jsiiProxy_RedisInstance) ResetRedisVersion() {
	_jsii_.InvokeVoid(
		r,
		"resetRedisVersion",
		nil, // no parameters
	)
}

func (r *jsiiProxy_RedisInstance) ResetRegion() {
	_jsii_.InvokeVoid(
		r,
		"resetRegion",
		nil, // no parameters
	)
}

func (r *jsiiProxy_RedisInstance) ResetReplicaCount() {
	_jsii_.InvokeVoid(
		r,
		"resetReplicaCount",
		nil, // no parameters
	)
}

func (r *jsiiProxy_RedisInstance) ResetReservedIpRange() {
	_jsii_.InvokeVoid(
		r,
		"resetReservedIpRange",
		nil, // no parameters
	)
}

func (r *jsiiProxy_RedisInstance) ResetSecondaryIpRange() {
	_jsii_.InvokeVoid(
		r,
		"resetSecondaryIpRange",
		nil, // no parameters
	)
}

func (r *jsiiProxy_RedisInstance) ResetTier() {
	_jsii_.InvokeVoid(
		r,
		"resetTier",
		nil, // no parameters
	)
}

func (r *jsiiProxy_RedisInstance) ResetTimeouts() {
	_jsii_.InvokeVoid(
		r,
		"resetTimeouts",
		nil, // no parameters
	)
}

func (r *jsiiProxy_RedisInstance) ResetTransitEncryptionMode() {
	_jsii_.InvokeVoid(
		r,
		"resetTransitEncryptionMode",
		nil, // no parameters
	)
}

func (r *jsiiProxy_RedisInstance) SynthesizeAttributes() *map[string]interface{} {
	var returns *map[string]interface{}

	_jsii_.Invoke(
		r,
		"synthesizeAttributes",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (r *jsiiProxy_RedisInstance) ToMetadata() interface{} {
	var returns interface{}

	_jsii_.Invoke(
		r,
		"toMetadata",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (r *jsiiProxy_RedisInstance) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		r,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (r *jsiiProxy_RedisInstance) ToTerraform() interface{} {
	var returns interface{}

	_jsii_.Invoke(
		r,
		"toTerraform",
		nil, // no parameters
		&returns,
	)

	return returns
}

