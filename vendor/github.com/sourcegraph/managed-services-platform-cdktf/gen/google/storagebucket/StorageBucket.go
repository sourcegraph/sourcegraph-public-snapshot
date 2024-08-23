package storagebucket

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/sourcegraph/managed-services-platform-cdktf/gen/google/jsii"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/storagebucket/internal"
)

// Represents a {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/storage_bucket google_storage_bucket}.
type StorageBucket interface {
	cdktf.TerraformResource
	Autoclass() StorageBucketAutoclassOutputReference
	AutoclassInput() *StorageBucketAutoclass
	// Experimental.
	CdktfStack() cdktf.TerraformStack
	// Experimental.
	Connection() interface{}
	// Experimental.
	SetConnection(val interface{})
	// Experimental.
	ConstructNodeMetadata() *map[string]interface{}
	Cors() StorageBucketCorsList
	CorsInput() interface{}
	// Experimental.
	Count() interface{}
	// Experimental.
	SetCount(val interface{})
	CustomPlacementConfig() StorageBucketCustomPlacementConfigOutputReference
	CustomPlacementConfigInput() *StorageBucketCustomPlacementConfig
	DefaultEventBasedHold() interface{}
	SetDefaultEventBasedHold(val interface{})
	DefaultEventBasedHoldInput() interface{}
	// Experimental.
	DependsOn() *[]*string
	// Experimental.
	SetDependsOn(val *[]*string)
	EffectiveLabels() cdktf.StringMap
	EnableObjectRetention() interface{}
	SetEnableObjectRetention(val interface{})
	EnableObjectRetentionInput() interface{}
	Encryption() StorageBucketEncryptionOutputReference
	EncryptionInput() *StorageBucketEncryption
	ForceDestroy() interface{}
	SetForceDestroy(val interface{})
	ForceDestroyInput() interface{}
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
	Labels() *map[string]*string
	SetLabels(val *map[string]*string)
	LabelsInput() *map[string]*string
	// Experimental.
	Lifecycle() *cdktf.TerraformResourceLifecycle
	// Experimental.
	SetLifecycle(val *cdktf.TerraformResourceLifecycle)
	LifecycleRule() StorageBucketLifecycleRuleList
	LifecycleRuleInput() interface{}
	Location() *string
	SetLocation(val *string)
	LocationInput() *string
	Logging() StorageBucketLoggingOutputReference
	LoggingInput() *StorageBucketLogging
	Name() *string
	SetName(val *string)
	NameInput() *string
	// The tree node.
	Node() constructs.Node
	Project() *string
	SetProject(val *string)
	ProjectInput() *string
	ProjectNumber() *float64
	// Experimental.
	Provider() cdktf.TerraformProvider
	// Experimental.
	SetProvider(val cdktf.TerraformProvider)
	// Experimental.
	Provisioners() *[]interface{}
	// Experimental.
	SetProvisioners(val *[]interface{})
	PublicAccessPrevention() *string
	SetPublicAccessPrevention(val *string)
	PublicAccessPreventionInput() *string
	// Experimental.
	RawOverrides() interface{}
	RequesterPays() interface{}
	SetRequesterPays(val interface{})
	RequesterPaysInput() interface{}
	RetentionPolicy() StorageBucketRetentionPolicyOutputReference
	RetentionPolicyInput() *StorageBucketRetentionPolicy
	Rpo() *string
	SetRpo(val *string)
	RpoInput() *string
	SelfLink() *string
	SoftDeletePolicy() StorageBucketSoftDeletePolicyOutputReference
	SoftDeletePolicyInput() *StorageBucketSoftDeletePolicy
	StorageClass() *string
	SetStorageClass(val *string)
	StorageClassInput() *string
	// Experimental.
	TerraformGeneratorMetadata() *cdktf.TerraformProviderGeneratorMetadata
	TerraformLabels() cdktf.StringMap
	// Experimental.
	TerraformMetaArguments() *map[string]interface{}
	// Experimental.
	TerraformResourceType() *string
	Timeouts() StorageBucketTimeoutsOutputReference
	TimeoutsInput() interface{}
	UniformBucketLevelAccess() interface{}
	SetUniformBucketLevelAccess(val interface{})
	UniformBucketLevelAccessInput() interface{}
	Url() *string
	Versioning() StorageBucketVersioningOutputReference
	VersioningInput() *StorageBucketVersioning
	Website() StorageBucketWebsiteOutputReference
	WebsiteInput() *StorageBucketWebsite
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
	PutAutoclass(value *StorageBucketAutoclass)
	PutCors(value interface{})
	PutCustomPlacementConfig(value *StorageBucketCustomPlacementConfig)
	PutEncryption(value *StorageBucketEncryption)
	PutLifecycleRule(value interface{})
	PutLogging(value *StorageBucketLogging)
	PutRetentionPolicy(value *StorageBucketRetentionPolicy)
	PutSoftDeletePolicy(value *StorageBucketSoftDeletePolicy)
	PutTimeouts(value *StorageBucketTimeouts)
	PutVersioning(value *StorageBucketVersioning)
	PutWebsite(value *StorageBucketWebsite)
	ResetAutoclass()
	ResetCors()
	ResetCustomPlacementConfig()
	ResetDefaultEventBasedHold()
	ResetEnableObjectRetention()
	ResetEncryption()
	ResetForceDestroy()
	ResetId()
	ResetLabels()
	ResetLifecycleRule()
	ResetLogging()
	// Resets a previously passed logical Id to use the auto-generated logical id again.
	// Experimental.
	ResetOverrideLogicalId()
	ResetProject()
	ResetPublicAccessPrevention()
	ResetRequesterPays()
	ResetRetentionPolicy()
	ResetRpo()
	ResetSoftDeletePolicy()
	ResetStorageClass()
	ResetTimeouts()
	ResetUniformBucketLevelAccess()
	ResetVersioning()
	ResetWebsite()
	SynthesizeAttributes() *map[string]interface{}
	// Experimental.
	ToMetadata() interface{}
	// Returns a string representation of this construct.
	ToString() *string
	// Adds this resource to the terraform JSON output.
	// Experimental.
	ToTerraform() interface{}
}

// The jsii proxy struct for StorageBucket
type jsiiProxy_StorageBucket struct {
	internal.Type__cdktfTerraformResource
}

func (j *jsiiProxy_StorageBucket) Autoclass() StorageBucketAutoclassOutputReference {
	var returns StorageBucketAutoclassOutputReference
	_jsii_.Get(
		j,
		"autoclass",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucket) AutoclassInput() *StorageBucketAutoclass {
	var returns *StorageBucketAutoclass
	_jsii_.Get(
		j,
		"autoclassInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucket) CdktfStack() cdktf.TerraformStack {
	var returns cdktf.TerraformStack
	_jsii_.Get(
		j,
		"cdktfStack",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucket) Connection() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"connection",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucket) ConstructNodeMetadata() *map[string]interface{} {
	var returns *map[string]interface{}
	_jsii_.Get(
		j,
		"constructNodeMetadata",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucket) Cors() StorageBucketCorsList {
	var returns StorageBucketCorsList
	_jsii_.Get(
		j,
		"cors",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucket) CorsInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"corsInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucket) Count() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"count",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucket) CustomPlacementConfig() StorageBucketCustomPlacementConfigOutputReference {
	var returns StorageBucketCustomPlacementConfigOutputReference
	_jsii_.Get(
		j,
		"customPlacementConfig",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucket) CustomPlacementConfigInput() *StorageBucketCustomPlacementConfig {
	var returns *StorageBucketCustomPlacementConfig
	_jsii_.Get(
		j,
		"customPlacementConfigInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucket) DefaultEventBasedHold() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"defaultEventBasedHold",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucket) DefaultEventBasedHoldInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"defaultEventBasedHoldInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucket) DependsOn() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"dependsOn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucket) EffectiveLabels() cdktf.StringMap {
	var returns cdktf.StringMap
	_jsii_.Get(
		j,
		"effectiveLabels",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucket) EnableObjectRetention() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"enableObjectRetention",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucket) EnableObjectRetentionInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"enableObjectRetentionInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucket) Encryption() StorageBucketEncryptionOutputReference {
	var returns StorageBucketEncryptionOutputReference
	_jsii_.Get(
		j,
		"encryption",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucket) EncryptionInput() *StorageBucketEncryption {
	var returns *StorageBucketEncryption
	_jsii_.Get(
		j,
		"encryptionInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucket) ForceDestroy() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"forceDestroy",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucket) ForceDestroyInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"forceDestroyInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucket) ForEach() cdktf.ITerraformIterator {
	var returns cdktf.ITerraformIterator
	_jsii_.Get(
		j,
		"forEach",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucket) Fqn() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fqn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucket) FriendlyUniqueId() *string {
	var returns *string
	_jsii_.Get(
		j,
		"friendlyUniqueId",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucket) Id() *string {
	var returns *string
	_jsii_.Get(
		j,
		"id",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucket) IdInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"idInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucket) Labels() *map[string]*string {
	var returns *map[string]*string
	_jsii_.Get(
		j,
		"labels",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucket) LabelsInput() *map[string]*string {
	var returns *map[string]*string
	_jsii_.Get(
		j,
		"labelsInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucket) Lifecycle() *cdktf.TerraformResourceLifecycle {
	var returns *cdktf.TerraformResourceLifecycle
	_jsii_.Get(
		j,
		"lifecycle",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucket) LifecycleRule() StorageBucketLifecycleRuleList {
	var returns StorageBucketLifecycleRuleList
	_jsii_.Get(
		j,
		"lifecycleRule",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucket) LifecycleRuleInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"lifecycleRuleInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucket) Location() *string {
	var returns *string
	_jsii_.Get(
		j,
		"location",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucket) LocationInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"locationInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucket) Logging() StorageBucketLoggingOutputReference {
	var returns StorageBucketLoggingOutputReference
	_jsii_.Get(
		j,
		"logging",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucket) LoggingInput() *StorageBucketLogging {
	var returns *StorageBucketLogging
	_jsii_.Get(
		j,
		"loggingInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucket) Name() *string {
	var returns *string
	_jsii_.Get(
		j,
		"name",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucket) NameInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"nameInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucket) Node() constructs.Node {
	var returns constructs.Node
	_jsii_.Get(
		j,
		"node",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucket) Project() *string {
	var returns *string
	_jsii_.Get(
		j,
		"project",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucket) ProjectInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"projectInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucket) ProjectNumber() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"projectNumber",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucket) Provider() cdktf.TerraformProvider {
	var returns cdktf.TerraformProvider
	_jsii_.Get(
		j,
		"provider",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucket) Provisioners() *[]interface{} {
	var returns *[]interface{}
	_jsii_.Get(
		j,
		"provisioners",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucket) PublicAccessPrevention() *string {
	var returns *string
	_jsii_.Get(
		j,
		"publicAccessPrevention",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucket) PublicAccessPreventionInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"publicAccessPreventionInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucket) RawOverrides() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"rawOverrides",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucket) RequesterPays() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"requesterPays",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucket) RequesterPaysInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"requesterPaysInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucket) RetentionPolicy() StorageBucketRetentionPolicyOutputReference {
	var returns StorageBucketRetentionPolicyOutputReference
	_jsii_.Get(
		j,
		"retentionPolicy",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucket) RetentionPolicyInput() *StorageBucketRetentionPolicy {
	var returns *StorageBucketRetentionPolicy
	_jsii_.Get(
		j,
		"retentionPolicyInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucket) Rpo() *string {
	var returns *string
	_jsii_.Get(
		j,
		"rpo",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucket) RpoInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"rpoInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucket) SelfLink() *string {
	var returns *string
	_jsii_.Get(
		j,
		"selfLink",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucket) SoftDeletePolicy() StorageBucketSoftDeletePolicyOutputReference {
	var returns StorageBucketSoftDeletePolicyOutputReference
	_jsii_.Get(
		j,
		"softDeletePolicy",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucket) SoftDeletePolicyInput() *StorageBucketSoftDeletePolicy {
	var returns *StorageBucketSoftDeletePolicy
	_jsii_.Get(
		j,
		"softDeletePolicyInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucket) StorageClass() *string {
	var returns *string
	_jsii_.Get(
		j,
		"storageClass",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucket) StorageClassInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"storageClassInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucket) TerraformGeneratorMetadata() *cdktf.TerraformProviderGeneratorMetadata {
	var returns *cdktf.TerraformProviderGeneratorMetadata
	_jsii_.Get(
		j,
		"terraformGeneratorMetadata",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucket) TerraformLabels() cdktf.StringMap {
	var returns cdktf.StringMap
	_jsii_.Get(
		j,
		"terraformLabels",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucket) TerraformMetaArguments() *map[string]interface{} {
	var returns *map[string]interface{}
	_jsii_.Get(
		j,
		"terraformMetaArguments",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucket) TerraformResourceType() *string {
	var returns *string
	_jsii_.Get(
		j,
		"terraformResourceType",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucket) Timeouts() StorageBucketTimeoutsOutputReference {
	var returns StorageBucketTimeoutsOutputReference
	_jsii_.Get(
		j,
		"timeouts",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucket) TimeoutsInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"timeoutsInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucket) UniformBucketLevelAccess() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"uniformBucketLevelAccess",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucket) UniformBucketLevelAccessInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"uniformBucketLevelAccessInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucket) Url() *string {
	var returns *string
	_jsii_.Get(
		j,
		"url",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucket) Versioning() StorageBucketVersioningOutputReference {
	var returns StorageBucketVersioningOutputReference
	_jsii_.Get(
		j,
		"versioning",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucket) VersioningInput() *StorageBucketVersioning {
	var returns *StorageBucketVersioning
	_jsii_.Get(
		j,
		"versioningInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucket) Website() StorageBucketWebsiteOutputReference {
	var returns StorageBucketWebsiteOutputReference
	_jsii_.Get(
		j,
		"website",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucket) WebsiteInput() *StorageBucketWebsite {
	var returns *StorageBucketWebsite
	_jsii_.Get(
		j,
		"websiteInput",
		&returns,
	)
	return returns
}


// Create a new {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/storage_bucket google_storage_bucket} Resource.
func NewStorageBucket(scope constructs.Construct, id *string, config *StorageBucketConfig) StorageBucket {
	_init_.Initialize()

	if err := validateNewStorageBucketParameters(scope, id, config); err != nil {
		panic(err)
	}
	j := jsiiProxy_StorageBucket{}

	_jsii_.Create(
		"@cdktf/provider-google.storageBucket.StorageBucket",
		[]interface{}{scope, id, config},
		&j,
	)

	return &j
}

// Create a new {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/storage_bucket google_storage_bucket} Resource.
func NewStorageBucket_Override(s StorageBucket, scope constructs.Construct, id *string, config *StorageBucketConfig) {
	_init_.Initialize()

	_jsii_.Create(
		"@cdktf/provider-google.storageBucket.StorageBucket",
		[]interface{}{scope, id, config},
		s,
	)
}

func (j *jsiiProxy_StorageBucket)SetConnection(val interface{}) {
	if err := j.validateSetConnectionParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"connection",
		val,
	)
}

func (j *jsiiProxy_StorageBucket)SetCount(val interface{}) {
	if err := j.validateSetCountParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"count",
		val,
	)
}

func (j *jsiiProxy_StorageBucket)SetDefaultEventBasedHold(val interface{}) {
	if err := j.validateSetDefaultEventBasedHoldParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"defaultEventBasedHold",
		val,
	)
}

func (j *jsiiProxy_StorageBucket)SetDependsOn(val *[]*string) {
	_jsii_.Set(
		j,
		"dependsOn",
		val,
	)
}

func (j *jsiiProxy_StorageBucket)SetEnableObjectRetention(val interface{}) {
	if err := j.validateSetEnableObjectRetentionParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"enableObjectRetention",
		val,
	)
}

func (j *jsiiProxy_StorageBucket)SetForceDestroy(val interface{}) {
	if err := j.validateSetForceDestroyParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"forceDestroy",
		val,
	)
}

func (j *jsiiProxy_StorageBucket)SetForEach(val cdktf.ITerraformIterator) {
	_jsii_.Set(
		j,
		"forEach",
		val,
	)
}

func (j *jsiiProxy_StorageBucket)SetId(val *string) {
	if err := j.validateSetIdParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"id",
		val,
	)
}

func (j *jsiiProxy_StorageBucket)SetLabels(val *map[string]*string) {
	if err := j.validateSetLabelsParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"labels",
		val,
	)
}

func (j *jsiiProxy_StorageBucket)SetLifecycle(val *cdktf.TerraformResourceLifecycle) {
	if err := j.validateSetLifecycleParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"lifecycle",
		val,
	)
}

func (j *jsiiProxy_StorageBucket)SetLocation(val *string) {
	if err := j.validateSetLocationParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"location",
		val,
	)
}

func (j *jsiiProxy_StorageBucket)SetName(val *string) {
	if err := j.validateSetNameParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"name",
		val,
	)
}

func (j *jsiiProxy_StorageBucket)SetProject(val *string) {
	if err := j.validateSetProjectParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"project",
		val,
	)
}

func (j *jsiiProxy_StorageBucket)SetProvider(val cdktf.TerraformProvider) {
	_jsii_.Set(
		j,
		"provider",
		val,
	)
}

func (j *jsiiProxy_StorageBucket)SetProvisioners(val *[]interface{}) {
	if err := j.validateSetProvisionersParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"provisioners",
		val,
	)
}

func (j *jsiiProxy_StorageBucket)SetPublicAccessPrevention(val *string) {
	if err := j.validateSetPublicAccessPreventionParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"publicAccessPrevention",
		val,
	)
}

func (j *jsiiProxy_StorageBucket)SetRequesterPays(val interface{}) {
	if err := j.validateSetRequesterPaysParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"requesterPays",
		val,
	)
}

func (j *jsiiProxy_StorageBucket)SetRpo(val *string) {
	if err := j.validateSetRpoParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"rpo",
		val,
	)
}

func (j *jsiiProxy_StorageBucket)SetStorageClass(val *string) {
	if err := j.validateSetStorageClassParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"storageClass",
		val,
	)
}

func (j *jsiiProxy_StorageBucket)SetUniformBucketLevelAccess(val interface{}) {
	if err := j.validateSetUniformBucketLevelAccessParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"uniformBucketLevelAccess",
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
func StorageBucket_IsConstruct(x interface{}) *bool {
	_init_.Initialize()

	if err := validateStorageBucket_IsConstructParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"@cdktf/provider-google.storageBucket.StorageBucket",
		"isConstruct",
		[]interface{}{x},
		&returns,
	)

	return returns
}

// Experimental.
func StorageBucket_IsTerraformElement(x interface{}) *bool {
	_init_.Initialize()

	if err := validateStorageBucket_IsTerraformElementParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"@cdktf/provider-google.storageBucket.StorageBucket",
		"isTerraformElement",
		[]interface{}{x},
		&returns,
	)

	return returns
}

// Experimental.
func StorageBucket_IsTerraformResource(x interface{}) *bool {
	_init_.Initialize()

	if err := validateStorageBucket_IsTerraformResourceParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"@cdktf/provider-google.storageBucket.StorageBucket",
		"isTerraformResource",
		[]interface{}{x},
		&returns,
	)

	return returns
}

func StorageBucket_TfResourceType() *string {
	_init_.Initialize()
	var returns *string
	_jsii_.StaticGet(
		"@cdktf/provider-google.storageBucket.StorageBucket",
		"tfResourceType",
		&returns,
	)
	return returns
}

func (s *jsiiProxy_StorageBucket) AddOverride(path *string, value interface{}) {
	if err := s.validateAddOverrideParameters(path, value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		s,
		"addOverride",
		[]interface{}{path, value},
	)
}

func (s *jsiiProxy_StorageBucket) GetAnyMapAttribute(terraformAttribute *string) *map[string]interface{} {
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

func (s *jsiiProxy_StorageBucket) GetBooleanAttribute(terraformAttribute *string) cdktf.IResolvable {
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

func (s *jsiiProxy_StorageBucket) GetBooleanMapAttribute(terraformAttribute *string) *map[string]*bool {
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

func (s *jsiiProxy_StorageBucket) GetListAttribute(terraformAttribute *string) *[]*string {
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

func (s *jsiiProxy_StorageBucket) GetNumberAttribute(terraformAttribute *string) *float64 {
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

func (s *jsiiProxy_StorageBucket) GetNumberListAttribute(terraformAttribute *string) *[]*float64 {
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

func (s *jsiiProxy_StorageBucket) GetNumberMapAttribute(terraformAttribute *string) *map[string]*float64 {
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

func (s *jsiiProxy_StorageBucket) GetStringAttribute(terraformAttribute *string) *string {
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

func (s *jsiiProxy_StorageBucket) GetStringMapAttribute(terraformAttribute *string) *map[string]*string {
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

func (s *jsiiProxy_StorageBucket) InterpolationForAttribute(terraformAttribute *string) cdktf.IResolvable {
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

func (s *jsiiProxy_StorageBucket) OverrideLogicalId(newLogicalId *string) {
	if err := s.validateOverrideLogicalIdParameters(newLogicalId); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		s,
		"overrideLogicalId",
		[]interface{}{newLogicalId},
	)
}

func (s *jsiiProxy_StorageBucket) PutAutoclass(value *StorageBucketAutoclass) {
	if err := s.validatePutAutoclassParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		s,
		"putAutoclass",
		[]interface{}{value},
	)
}

func (s *jsiiProxy_StorageBucket) PutCors(value interface{}) {
	if err := s.validatePutCorsParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		s,
		"putCors",
		[]interface{}{value},
	)
}

func (s *jsiiProxy_StorageBucket) PutCustomPlacementConfig(value *StorageBucketCustomPlacementConfig) {
	if err := s.validatePutCustomPlacementConfigParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		s,
		"putCustomPlacementConfig",
		[]interface{}{value},
	)
}

func (s *jsiiProxy_StorageBucket) PutEncryption(value *StorageBucketEncryption) {
	if err := s.validatePutEncryptionParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		s,
		"putEncryption",
		[]interface{}{value},
	)
}

func (s *jsiiProxy_StorageBucket) PutLifecycleRule(value interface{}) {
	if err := s.validatePutLifecycleRuleParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		s,
		"putLifecycleRule",
		[]interface{}{value},
	)
}

func (s *jsiiProxy_StorageBucket) PutLogging(value *StorageBucketLogging) {
	if err := s.validatePutLoggingParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		s,
		"putLogging",
		[]interface{}{value},
	)
}

func (s *jsiiProxy_StorageBucket) PutRetentionPolicy(value *StorageBucketRetentionPolicy) {
	if err := s.validatePutRetentionPolicyParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		s,
		"putRetentionPolicy",
		[]interface{}{value},
	)
}

func (s *jsiiProxy_StorageBucket) PutSoftDeletePolicy(value *StorageBucketSoftDeletePolicy) {
	if err := s.validatePutSoftDeletePolicyParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		s,
		"putSoftDeletePolicy",
		[]interface{}{value},
	)
}

func (s *jsiiProxy_StorageBucket) PutTimeouts(value *StorageBucketTimeouts) {
	if err := s.validatePutTimeoutsParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		s,
		"putTimeouts",
		[]interface{}{value},
	)
}

func (s *jsiiProxy_StorageBucket) PutVersioning(value *StorageBucketVersioning) {
	if err := s.validatePutVersioningParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		s,
		"putVersioning",
		[]interface{}{value},
	)
}

func (s *jsiiProxy_StorageBucket) PutWebsite(value *StorageBucketWebsite) {
	if err := s.validatePutWebsiteParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		s,
		"putWebsite",
		[]interface{}{value},
	)
}

func (s *jsiiProxy_StorageBucket) ResetAutoclass() {
	_jsii_.InvokeVoid(
		s,
		"resetAutoclass",
		nil, // no parameters
	)
}

func (s *jsiiProxy_StorageBucket) ResetCors() {
	_jsii_.InvokeVoid(
		s,
		"resetCors",
		nil, // no parameters
	)
}

func (s *jsiiProxy_StorageBucket) ResetCustomPlacementConfig() {
	_jsii_.InvokeVoid(
		s,
		"resetCustomPlacementConfig",
		nil, // no parameters
	)
}

func (s *jsiiProxy_StorageBucket) ResetDefaultEventBasedHold() {
	_jsii_.InvokeVoid(
		s,
		"resetDefaultEventBasedHold",
		nil, // no parameters
	)
}

func (s *jsiiProxy_StorageBucket) ResetEnableObjectRetention() {
	_jsii_.InvokeVoid(
		s,
		"resetEnableObjectRetention",
		nil, // no parameters
	)
}

func (s *jsiiProxy_StorageBucket) ResetEncryption() {
	_jsii_.InvokeVoid(
		s,
		"resetEncryption",
		nil, // no parameters
	)
}

func (s *jsiiProxy_StorageBucket) ResetForceDestroy() {
	_jsii_.InvokeVoid(
		s,
		"resetForceDestroy",
		nil, // no parameters
	)
}

func (s *jsiiProxy_StorageBucket) ResetId() {
	_jsii_.InvokeVoid(
		s,
		"resetId",
		nil, // no parameters
	)
}

func (s *jsiiProxy_StorageBucket) ResetLabels() {
	_jsii_.InvokeVoid(
		s,
		"resetLabels",
		nil, // no parameters
	)
}

func (s *jsiiProxy_StorageBucket) ResetLifecycleRule() {
	_jsii_.InvokeVoid(
		s,
		"resetLifecycleRule",
		nil, // no parameters
	)
}

func (s *jsiiProxy_StorageBucket) ResetLogging() {
	_jsii_.InvokeVoid(
		s,
		"resetLogging",
		nil, // no parameters
	)
}

func (s *jsiiProxy_StorageBucket) ResetOverrideLogicalId() {
	_jsii_.InvokeVoid(
		s,
		"resetOverrideLogicalId",
		nil, // no parameters
	)
}

func (s *jsiiProxy_StorageBucket) ResetProject() {
	_jsii_.InvokeVoid(
		s,
		"resetProject",
		nil, // no parameters
	)
}

func (s *jsiiProxy_StorageBucket) ResetPublicAccessPrevention() {
	_jsii_.InvokeVoid(
		s,
		"resetPublicAccessPrevention",
		nil, // no parameters
	)
}

func (s *jsiiProxy_StorageBucket) ResetRequesterPays() {
	_jsii_.InvokeVoid(
		s,
		"resetRequesterPays",
		nil, // no parameters
	)
}

func (s *jsiiProxy_StorageBucket) ResetRetentionPolicy() {
	_jsii_.InvokeVoid(
		s,
		"resetRetentionPolicy",
		nil, // no parameters
	)
}

func (s *jsiiProxy_StorageBucket) ResetRpo() {
	_jsii_.InvokeVoid(
		s,
		"resetRpo",
		nil, // no parameters
	)
}

func (s *jsiiProxy_StorageBucket) ResetSoftDeletePolicy() {
	_jsii_.InvokeVoid(
		s,
		"resetSoftDeletePolicy",
		nil, // no parameters
	)
}

func (s *jsiiProxy_StorageBucket) ResetStorageClass() {
	_jsii_.InvokeVoid(
		s,
		"resetStorageClass",
		nil, // no parameters
	)
}

func (s *jsiiProxy_StorageBucket) ResetTimeouts() {
	_jsii_.InvokeVoid(
		s,
		"resetTimeouts",
		nil, // no parameters
	)
}

func (s *jsiiProxy_StorageBucket) ResetUniformBucketLevelAccess() {
	_jsii_.InvokeVoid(
		s,
		"resetUniformBucketLevelAccess",
		nil, // no parameters
	)
}

func (s *jsiiProxy_StorageBucket) ResetVersioning() {
	_jsii_.InvokeVoid(
		s,
		"resetVersioning",
		nil, // no parameters
	)
}

func (s *jsiiProxy_StorageBucket) ResetWebsite() {
	_jsii_.InvokeVoid(
		s,
		"resetWebsite",
		nil, // no parameters
	)
}

func (s *jsiiProxy_StorageBucket) SynthesizeAttributes() *map[string]interface{} {
	var returns *map[string]interface{}

	_jsii_.Invoke(
		s,
		"synthesizeAttributes",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (s *jsiiProxy_StorageBucket) ToMetadata() interface{} {
	var returns interface{}

	_jsii_.Invoke(
		s,
		"toMetadata",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (s *jsiiProxy_StorageBucket) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		s,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (s *jsiiProxy_StorageBucket) ToTerraform() interface{} {
	var returns interface{}

	_jsii_.Invoke(
		s,
		"toTerraform",
		nil, // no parameters
		&returns,
	)

	return returns
}

