package computeinstance

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/sourcegraph/managed-services-platform-cdktf/gen/google/jsii"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/computeinstance/internal"
)

// Represents a {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_instance google_compute_instance}.
type ComputeInstance interface {
	cdktf.TerraformResource
	AdvancedMachineFeatures() ComputeInstanceAdvancedMachineFeaturesOutputReference
	AdvancedMachineFeaturesInput() *ComputeInstanceAdvancedMachineFeatures
	AllowStoppingForUpdate() interface{}
	SetAllowStoppingForUpdate(val interface{})
	AllowStoppingForUpdateInput() interface{}
	AttachedDisk() ComputeInstanceAttachedDiskList
	AttachedDiskInput() interface{}
	BootDisk() ComputeInstanceBootDiskOutputReference
	BootDiskInput() *ComputeInstanceBootDisk
	CanIpForward() interface{}
	SetCanIpForward(val interface{})
	CanIpForwardInput() interface{}
	// Experimental.
	CdktfStack() cdktf.TerraformStack
	ConfidentialInstanceConfig() ComputeInstanceConfidentialInstanceConfigOutputReference
	ConfidentialInstanceConfigInput() *ComputeInstanceConfidentialInstanceConfig
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
	CpuPlatform() *string
	CurrentStatus() *string
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
	DesiredStatus() *string
	SetDesiredStatus(val *string)
	DesiredStatusInput() *string
	EffectiveLabels() cdktf.StringMap
	EnableDisplay() interface{}
	SetEnableDisplay(val interface{})
	EnableDisplayInput() interface{}
	// Experimental.
	ForEach() cdktf.ITerraformIterator
	// Experimental.
	SetForEach(val cdktf.ITerraformIterator)
	// Experimental.
	Fqn() *string
	// Experimental.
	FriendlyUniqueId() *string
	GuestAccelerator() ComputeInstanceGuestAcceleratorList
	GuestAcceleratorInput() interface{}
	Hostname() *string
	SetHostname(val *string)
	HostnameInput() *string
	Id() *string
	SetId(val *string)
	IdInput() *string
	InstanceId() *string
	LabelFingerprint() *string
	Labels() *map[string]*string
	SetLabels(val *map[string]*string)
	LabelsInput() *map[string]*string
	// Experimental.
	Lifecycle() *cdktf.TerraformResourceLifecycle
	// Experimental.
	SetLifecycle(val *cdktf.TerraformResourceLifecycle)
	MachineType() *string
	SetMachineType(val *string)
	MachineTypeInput() *string
	Metadata() *map[string]*string
	SetMetadata(val *map[string]*string)
	MetadataFingerprint() *string
	MetadataInput() *map[string]*string
	MetadataStartupScript() *string
	SetMetadataStartupScript(val *string)
	MetadataStartupScriptInput() *string
	MinCpuPlatform() *string
	SetMinCpuPlatform(val *string)
	MinCpuPlatformInput() *string
	Name() *string
	SetName(val *string)
	NameInput() *string
	NetworkInterface() ComputeInstanceNetworkInterfaceList
	NetworkInterfaceInput() interface{}
	NetworkPerformanceConfig() ComputeInstanceNetworkPerformanceConfigOutputReference
	NetworkPerformanceConfigInput() *ComputeInstanceNetworkPerformanceConfig
	// The tree node.
	Node() constructs.Node
	Params() ComputeInstanceParamsOutputReference
	ParamsInput() *ComputeInstanceParams
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
	ReservationAffinity() ComputeInstanceReservationAffinityOutputReference
	ReservationAffinityInput() *ComputeInstanceReservationAffinity
	ResourcePolicies() *[]*string
	SetResourcePolicies(val *[]*string)
	ResourcePoliciesInput() *[]*string
	Scheduling() ComputeInstanceSchedulingOutputReference
	SchedulingInput() *ComputeInstanceScheduling
	ScratchDisk() ComputeInstanceScratchDiskList
	ScratchDiskInput() interface{}
	SelfLink() *string
	ServiceAccount() ComputeInstanceServiceAccountOutputReference
	ServiceAccountInput() *ComputeInstanceServiceAccount
	ShieldedInstanceConfig() ComputeInstanceShieldedInstanceConfigOutputReference
	ShieldedInstanceConfigInput() *ComputeInstanceShieldedInstanceConfig
	Tags() *[]*string
	SetTags(val *[]*string)
	TagsFingerprint() *string
	TagsInput() *[]*string
	// Experimental.
	TerraformGeneratorMetadata() *cdktf.TerraformProviderGeneratorMetadata
	TerraformLabels() cdktf.StringMap
	// Experimental.
	TerraformMetaArguments() *map[string]interface{}
	// Experimental.
	TerraformResourceType() *string
	Timeouts() ComputeInstanceTimeoutsOutputReference
	TimeoutsInput() interface{}
	Zone() *string
	SetZone(val *string)
	ZoneInput() *string
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
	PutAdvancedMachineFeatures(value *ComputeInstanceAdvancedMachineFeatures)
	PutAttachedDisk(value interface{})
	PutBootDisk(value *ComputeInstanceBootDisk)
	PutConfidentialInstanceConfig(value *ComputeInstanceConfidentialInstanceConfig)
	PutGuestAccelerator(value interface{})
	PutNetworkInterface(value interface{})
	PutNetworkPerformanceConfig(value *ComputeInstanceNetworkPerformanceConfig)
	PutParams(value *ComputeInstanceParams)
	PutReservationAffinity(value *ComputeInstanceReservationAffinity)
	PutScheduling(value *ComputeInstanceScheduling)
	PutScratchDisk(value interface{})
	PutServiceAccount(value *ComputeInstanceServiceAccount)
	PutShieldedInstanceConfig(value *ComputeInstanceShieldedInstanceConfig)
	PutTimeouts(value *ComputeInstanceTimeouts)
	ResetAdvancedMachineFeatures()
	ResetAllowStoppingForUpdate()
	ResetAttachedDisk()
	ResetCanIpForward()
	ResetConfidentialInstanceConfig()
	ResetDeletionProtection()
	ResetDescription()
	ResetDesiredStatus()
	ResetEnableDisplay()
	ResetGuestAccelerator()
	ResetHostname()
	ResetId()
	ResetLabels()
	ResetMetadata()
	ResetMetadataStartupScript()
	ResetMinCpuPlatform()
	ResetNetworkPerformanceConfig()
	// Resets a previously passed logical Id to use the auto-generated logical id again.
	// Experimental.
	ResetOverrideLogicalId()
	ResetParams()
	ResetProject()
	ResetReservationAffinity()
	ResetResourcePolicies()
	ResetScheduling()
	ResetScratchDisk()
	ResetServiceAccount()
	ResetShieldedInstanceConfig()
	ResetTags()
	ResetTimeouts()
	ResetZone()
	SynthesizeAttributes() *map[string]interface{}
	// Experimental.
	ToMetadata() interface{}
	// Returns a string representation of this construct.
	ToString() *string
	// Adds this resource to the terraform JSON output.
	// Experimental.
	ToTerraform() interface{}
}

// The jsii proxy struct for ComputeInstance
type jsiiProxy_ComputeInstance struct {
	internal.Type__cdktfTerraformResource
}

func (j *jsiiProxy_ComputeInstance) AdvancedMachineFeatures() ComputeInstanceAdvancedMachineFeaturesOutputReference {
	var returns ComputeInstanceAdvancedMachineFeaturesOutputReference
	_jsii_.Get(
		j,
		"advancedMachineFeatures",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) AdvancedMachineFeaturesInput() *ComputeInstanceAdvancedMachineFeatures {
	var returns *ComputeInstanceAdvancedMachineFeatures
	_jsii_.Get(
		j,
		"advancedMachineFeaturesInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) AllowStoppingForUpdate() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"allowStoppingForUpdate",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) AllowStoppingForUpdateInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"allowStoppingForUpdateInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) AttachedDisk() ComputeInstanceAttachedDiskList {
	var returns ComputeInstanceAttachedDiskList
	_jsii_.Get(
		j,
		"attachedDisk",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) AttachedDiskInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"attachedDiskInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) BootDisk() ComputeInstanceBootDiskOutputReference {
	var returns ComputeInstanceBootDiskOutputReference
	_jsii_.Get(
		j,
		"bootDisk",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) BootDiskInput() *ComputeInstanceBootDisk {
	var returns *ComputeInstanceBootDisk
	_jsii_.Get(
		j,
		"bootDiskInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) CanIpForward() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"canIpForward",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) CanIpForwardInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"canIpForwardInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) CdktfStack() cdktf.TerraformStack {
	var returns cdktf.TerraformStack
	_jsii_.Get(
		j,
		"cdktfStack",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) ConfidentialInstanceConfig() ComputeInstanceConfidentialInstanceConfigOutputReference {
	var returns ComputeInstanceConfidentialInstanceConfigOutputReference
	_jsii_.Get(
		j,
		"confidentialInstanceConfig",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) ConfidentialInstanceConfigInput() *ComputeInstanceConfidentialInstanceConfig {
	var returns *ComputeInstanceConfidentialInstanceConfig
	_jsii_.Get(
		j,
		"confidentialInstanceConfigInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) Connection() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"connection",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) ConstructNodeMetadata() *map[string]interface{} {
	var returns *map[string]interface{}
	_jsii_.Get(
		j,
		"constructNodeMetadata",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) Count() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"count",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) CpuPlatform() *string {
	var returns *string
	_jsii_.Get(
		j,
		"cpuPlatform",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) CurrentStatus() *string {
	var returns *string
	_jsii_.Get(
		j,
		"currentStatus",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) DeletionProtection() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"deletionProtection",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) DeletionProtectionInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"deletionProtectionInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) DependsOn() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"dependsOn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) Description() *string {
	var returns *string
	_jsii_.Get(
		j,
		"description",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) DescriptionInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"descriptionInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) DesiredStatus() *string {
	var returns *string
	_jsii_.Get(
		j,
		"desiredStatus",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) DesiredStatusInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"desiredStatusInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) EffectiveLabels() cdktf.StringMap {
	var returns cdktf.StringMap
	_jsii_.Get(
		j,
		"effectiveLabels",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) EnableDisplay() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"enableDisplay",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) EnableDisplayInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"enableDisplayInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) ForEach() cdktf.ITerraformIterator {
	var returns cdktf.ITerraformIterator
	_jsii_.Get(
		j,
		"forEach",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) Fqn() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fqn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) FriendlyUniqueId() *string {
	var returns *string
	_jsii_.Get(
		j,
		"friendlyUniqueId",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) GuestAccelerator() ComputeInstanceGuestAcceleratorList {
	var returns ComputeInstanceGuestAcceleratorList
	_jsii_.Get(
		j,
		"guestAccelerator",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) GuestAcceleratorInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"guestAcceleratorInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) Hostname() *string {
	var returns *string
	_jsii_.Get(
		j,
		"hostname",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) HostnameInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"hostnameInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) Id() *string {
	var returns *string
	_jsii_.Get(
		j,
		"id",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) IdInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"idInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) InstanceId() *string {
	var returns *string
	_jsii_.Get(
		j,
		"instanceId",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) LabelFingerprint() *string {
	var returns *string
	_jsii_.Get(
		j,
		"labelFingerprint",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) Labels() *map[string]*string {
	var returns *map[string]*string
	_jsii_.Get(
		j,
		"labels",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) LabelsInput() *map[string]*string {
	var returns *map[string]*string
	_jsii_.Get(
		j,
		"labelsInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) Lifecycle() *cdktf.TerraformResourceLifecycle {
	var returns *cdktf.TerraformResourceLifecycle
	_jsii_.Get(
		j,
		"lifecycle",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) MachineType() *string {
	var returns *string
	_jsii_.Get(
		j,
		"machineType",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) MachineTypeInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"machineTypeInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) Metadata() *map[string]*string {
	var returns *map[string]*string
	_jsii_.Get(
		j,
		"metadata",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) MetadataFingerprint() *string {
	var returns *string
	_jsii_.Get(
		j,
		"metadataFingerprint",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) MetadataInput() *map[string]*string {
	var returns *map[string]*string
	_jsii_.Get(
		j,
		"metadataInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) MetadataStartupScript() *string {
	var returns *string
	_jsii_.Get(
		j,
		"metadataStartupScript",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) MetadataStartupScriptInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"metadataStartupScriptInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) MinCpuPlatform() *string {
	var returns *string
	_jsii_.Get(
		j,
		"minCpuPlatform",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) MinCpuPlatformInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"minCpuPlatformInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) Name() *string {
	var returns *string
	_jsii_.Get(
		j,
		"name",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) NameInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"nameInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) NetworkInterface() ComputeInstanceNetworkInterfaceList {
	var returns ComputeInstanceNetworkInterfaceList
	_jsii_.Get(
		j,
		"networkInterface",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) NetworkInterfaceInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"networkInterfaceInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) NetworkPerformanceConfig() ComputeInstanceNetworkPerformanceConfigOutputReference {
	var returns ComputeInstanceNetworkPerformanceConfigOutputReference
	_jsii_.Get(
		j,
		"networkPerformanceConfig",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) NetworkPerformanceConfigInput() *ComputeInstanceNetworkPerformanceConfig {
	var returns *ComputeInstanceNetworkPerformanceConfig
	_jsii_.Get(
		j,
		"networkPerformanceConfigInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) Node() constructs.Node {
	var returns constructs.Node
	_jsii_.Get(
		j,
		"node",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) Params() ComputeInstanceParamsOutputReference {
	var returns ComputeInstanceParamsOutputReference
	_jsii_.Get(
		j,
		"params",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) ParamsInput() *ComputeInstanceParams {
	var returns *ComputeInstanceParams
	_jsii_.Get(
		j,
		"paramsInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) Project() *string {
	var returns *string
	_jsii_.Get(
		j,
		"project",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) ProjectInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"projectInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) Provider() cdktf.TerraformProvider {
	var returns cdktf.TerraformProvider
	_jsii_.Get(
		j,
		"provider",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) Provisioners() *[]interface{} {
	var returns *[]interface{}
	_jsii_.Get(
		j,
		"provisioners",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) RawOverrides() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"rawOverrides",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) ReservationAffinity() ComputeInstanceReservationAffinityOutputReference {
	var returns ComputeInstanceReservationAffinityOutputReference
	_jsii_.Get(
		j,
		"reservationAffinity",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) ReservationAffinityInput() *ComputeInstanceReservationAffinity {
	var returns *ComputeInstanceReservationAffinity
	_jsii_.Get(
		j,
		"reservationAffinityInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) ResourcePolicies() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"resourcePolicies",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) ResourcePoliciesInput() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"resourcePoliciesInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) Scheduling() ComputeInstanceSchedulingOutputReference {
	var returns ComputeInstanceSchedulingOutputReference
	_jsii_.Get(
		j,
		"scheduling",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) SchedulingInput() *ComputeInstanceScheduling {
	var returns *ComputeInstanceScheduling
	_jsii_.Get(
		j,
		"schedulingInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) ScratchDisk() ComputeInstanceScratchDiskList {
	var returns ComputeInstanceScratchDiskList
	_jsii_.Get(
		j,
		"scratchDisk",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) ScratchDiskInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"scratchDiskInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) SelfLink() *string {
	var returns *string
	_jsii_.Get(
		j,
		"selfLink",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) ServiceAccount() ComputeInstanceServiceAccountOutputReference {
	var returns ComputeInstanceServiceAccountOutputReference
	_jsii_.Get(
		j,
		"serviceAccount",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) ServiceAccountInput() *ComputeInstanceServiceAccount {
	var returns *ComputeInstanceServiceAccount
	_jsii_.Get(
		j,
		"serviceAccountInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) ShieldedInstanceConfig() ComputeInstanceShieldedInstanceConfigOutputReference {
	var returns ComputeInstanceShieldedInstanceConfigOutputReference
	_jsii_.Get(
		j,
		"shieldedInstanceConfig",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) ShieldedInstanceConfigInput() *ComputeInstanceShieldedInstanceConfig {
	var returns *ComputeInstanceShieldedInstanceConfig
	_jsii_.Get(
		j,
		"shieldedInstanceConfigInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) Tags() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"tags",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) TagsFingerprint() *string {
	var returns *string
	_jsii_.Get(
		j,
		"tagsFingerprint",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) TagsInput() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"tagsInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) TerraformGeneratorMetadata() *cdktf.TerraformProviderGeneratorMetadata {
	var returns *cdktf.TerraformProviderGeneratorMetadata
	_jsii_.Get(
		j,
		"terraformGeneratorMetadata",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) TerraformLabels() cdktf.StringMap {
	var returns cdktf.StringMap
	_jsii_.Get(
		j,
		"terraformLabels",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) TerraformMetaArguments() *map[string]interface{} {
	var returns *map[string]interface{}
	_jsii_.Get(
		j,
		"terraformMetaArguments",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) TerraformResourceType() *string {
	var returns *string
	_jsii_.Get(
		j,
		"terraformResourceType",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) Timeouts() ComputeInstanceTimeoutsOutputReference {
	var returns ComputeInstanceTimeoutsOutputReference
	_jsii_.Get(
		j,
		"timeouts",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) TimeoutsInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"timeoutsInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) Zone() *string {
	var returns *string
	_jsii_.Get(
		j,
		"zone",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeInstance) ZoneInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"zoneInput",
		&returns,
	)
	return returns
}


// Create a new {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_instance google_compute_instance} Resource.
func NewComputeInstance(scope constructs.Construct, id *string, config *ComputeInstanceConfig) ComputeInstance {
	_init_.Initialize()

	if err := validateNewComputeInstanceParameters(scope, id, config); err != nil {
		panic(err)
	}
	j := jsiiProxy_ComputeInstance{}

	_jsii_.Create(
		"@cdktf/provider-google.computeInstance.ComputeInstance",
		[]interface{}{scope, id, config},
		&j,
	)

	return &j
}

// Create a new {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_instance google_compute_instance} Resource.
func NewComputeInstance_Override(c ComputeInstance, scope constructs.Construct, id *string, config *ComputeInstanceConfig) {
	_init_.Initialize()

	_jsii_.Create(
		"@cdktf/provider-google.computeInstance.ComputeInstance",
		[]interface{}{scope, id, config},
		c,
	)
}

func (j *jsiiProxy_ComputeInstance)SetAllowStoppingForUpdate(val interface{}) {
	if err := j.validateSetAllowStoppingForUpdateParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"allowStoppingForUpdate",
		val,
	)
}

func (j *jsiiProxy_ComputeInstance)SetCanIpForward(val interface{}) {
	if err := j.validateSetCanIpForwardParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"canIpForward",
		val,
	)
}

func (j *jsiiProxy_ComputeInstance)SetConnection(val interface{}) {
	if err := j.validateSetConnectionParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"connection",
		val,
	)
}

func (j *jsiiProxy_ComputeInstance)SetCount(val interface{}) {
	if err := j.validateSetCountParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"count",
		val,
	)
}

func (j *jsiiProxy_ComputeInstance)SetDeletionProtection(val interface{}) {
	if err := j.validateSetDeletionProtectionParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"deletionProtection",
		val,
	)
}

func (j *jsiiProxy_ComputeInstance)SetDependsOn(val *[]*string) {
	_jsii_.Set(
		j,
		"dependsOn",
		val,
	)
}

func (j *jsiiProxy_ComputeInstance)SetDescription(val *string) {
	if err := j.validateSetDescriptionParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"description",
		val,
	)
}

func (j *jsiiProxy_ComputeInstance)SetDesiredStatus(val *string) {
	if err := j.validateSetDesiredStatusParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"desiredStatus",
		val,
	)
}

func (j *jsiiProxy_ComputeInstance)SetEnableDisplay(val interface{}) {
	if err := j.validateSetEnableDisplayParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"enableDisplay",
		val,
	)
}

func (j *jsiiProxy_ComputeInstance)SetForEach(val cdktf.ITerraformIterator) {
	_jsii_.Set(
		j,
		"forEach",
		val,
	)
}

func (j *jsiiProxy_ComputeInstance)SetHostname(val *string) {
	if err := j.validateSetHostnameParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"hostname",
		val,
	)
}

func (j *jsiiProxy_ComputeInstance)SetId(val *string) {
	if err := j.validateSetIdParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"id",
		val,
	)
}

func (j *jsiiProxy_ComputeInstance)SetLabels(val *map[string]*string) {
	if err := j.validateSetLabelsParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"labels",
		val,
	)
}

func (j *jsiiProxy_ComputeInstance)SetLifecycle(val *cdktf.TerraformResourceLifecycle) {
	if err := j.validateSetLifecycleParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"lifecycle",
		val,
	)
}

func (j *jsiiProxy_ComputeInstance)SetMachineType(val *string) {
	if err := j.validateSetMachineTypeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"machineType",
		val,
	)
}

func (j *jsiiProxy_ComputeInstance)SetMetadata(val *map[string]*string) {
	if err := j.validateSetMetadataParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"metadata",
		val,
	)
}

func (j *jsiiProxy_ComputeInstance)SetMetadataStartupScript(val *string) {
	if err := j.validateSetMetadataStartupScriptParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"metadataStartupScript",
		val,
	)
}

func (j *jsiiProxy_ComputeInstance)SetMinCpuPlatform(val *string) {
	if err := j.validateSetMinCpuPlatformParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"minCpuPlatform",
		val,
	)
}

func (j *jsiiProxy_ComputeInstance)SetName(val *string) {
	if err := j.validateSetNameParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"name",
		val,
	)
}

func (j *jsiiProxy_ComputeInstance)SetProject(val *string) {
	if err := j.validateSetProjectParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"project",
		val,
	)
}

func (j *jsiiProxy_ComputeInstance)SetProvider(val cdktf.TerraformProvider) {
	_jsii_.Set(
		j,
		"provider",
		val,
	)
}

func (j *jsiiProxy_ComputeInstance)SetProvisioners(val *[]interface{}) {
	if err := j.validateSetProvisionersParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"provisioners",
		val,
	)
}

func (j *jsiiProxy_ComputeInstance)SetResourcePolicies(val *[]*string) {
	if err := j.validateSetResourcePoliciesParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"resourcePolicies",
		val,
	)
}

func (j *jsiiProxy_ComputeInstance)SetTags(val *[]*string) {
	if err := j.validateSetTagsParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"tags",
		val,
	)
}

func (j *jsiiProxy_ComputeInstance)SetZone(val *string) {
	if err := j.validateSetZoneParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"zone",
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
func ComputeInstance_IsConstruct(x interface{}) *bool {
	_init_.Initialize()

	if err := validateComputeInstance_IsConstructParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"@cdktf/provider-google.computeInstance.ComputeInstance",
		"isConstruct",
		[]interface{}{x},
		&returns,
	)

	return returns
}

// Experimental.
func ComputeInstance_IsTerraformElement(x interface{}) *bool {
	_init_.Initialize()

	if err := validateComputeInstance_IsTerraformElementParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"@cdktf/provider-google.computeInstance.ComputeInstance",
		"isTerraformElement",
		[]interface{}{x},
		&returns,
	)

	return returns
}

// Experimental.
func ComputeInstance_IsTerraformResource(x interface{}) *bool {
	_init_.Initialize()

	if err := validateComputeInstance_IsTerraformResourceParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"@cdktf/provider-google.computeInstance.ComputeInstance",
		"isTerraformResource",
		[]interface{}{x},
		&returns,
	)

	return returns
}

func ComputeInstance_TfResourceType() *string {
	_init_.Initialize()
	var returns *string
	_jsii_.StaticGet(
		"@cdktf/provider-google.computeInstance.ComputeInstance",
		"tfResourceType",
		&returns,
	)
	return returns
}

func (c *jsiiProxy_ComputeInstance) AddOverride(path *string, value interface{}) {
	if err := c.validateAddOverrideParameters(path, value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"addOverride",
		[]interface{}{path, value},
	)
}

func (c *jsiiProxy_ComputeInstance) GetAnyMapAttribute(terraformAttribute *string) *map[string]interface{} {
	if err := c.validateGetAnyMapAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *map[string]interface{}

	_jsii_.Invoke(
		c,
		"getAnyMapAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComputeInstance) GetBooleanAttribute(terraformAttribute *string) cdktf.IResolvable {
	if err := c.validateGetBooleanAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		c,
		"getBooleanAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComputeInstance) GetBooleanMapAttribute(terraformAttribute *string) *map[string]*bool {
	if err := c.validateGetBooleanMapAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *map[string]*bool

	_jsii_.Invoke(
		c,
		"getBooleanMapAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComputeInstance) GetListAttribute(terraformAttribute *string) *[]*string {
	if err := c.validateGetListAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *[]*string

	_jsii_.Invoke(
		c,
		"getListAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComputeInstance) GetNumberAttribute(terraformAttribute *string) *float64 {
	if err := c.validateGetNumberAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *float64

	_jsii_.Invoke(
		c,
		"getNumberAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComputeInstance) GetNumberListAttribute(terraformAttribute *string) *[]*float64 {
	if err := c.validateGetNumberListAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *[]*float64

	_jsii_.Invoke(
		c,
		"getNumberListAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComputeInstance) GetNumberMapAttribute(terraformAttribute *string) *map[string]*float64 {
	if err := c.validateGetNumberMapAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *map[string]*float64

	_jsii_.Invoke(
		c,
		"getNumberMapAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComputeInstance) GetStringAttribute(terraformAttribute *string) *string {
	if err := c.validateGetStringAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.Invoke(
		c,
		"getStringAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComputeInstance) GetStringMapAttribute(terraformAttribute *string) *map[string]*string {
	if err := c.validateGetStringMapAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *map[string]*string

	_jsii_.Invoke(
		c,
		"getStringMapAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComputeInstance) InterpolationForAttribute(terraformAttribute *string) cdktf.IResolvable {
	if err := c.validateInterpolationForAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		c,
		"interpolationForAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComputeInstance) OverrideLogicalId(newLogicalId *string) {
	if err := c.validateOverrideLogicalIdParameters(newLogicalId); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"overrideLogicalId",
		[]interface{}{newLogicalId},
	)
}

func (c *jsiiProxy_ComputeInstance) PutAdvancedMachineFeatures(value *ComputeInstanceAdvancedMachineFeatures) {
	if err := c.validatePutAdvancedMachineFeaturesParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putAdvancedMachineFeatures",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ComputeInstance) PutAttachedDisk(value interface{}) {
	if err := c.validatePutAttachedDiskParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putAttachedDisk",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ComputeInstance) PutBootDisk(value *ComputeInstanceBootDisk) {
	if err := c.validatePutBootDiskParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putBootDisk",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ComputeInstance) PutConfidentialInstanceConfig(value *ComputeInstanceConfidentialInstanceConfig) {
	if err := c.validatePutConfidentialInstanceConfigParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putConfidentialInstanceConfig",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ComputeInstance) PutGuestAccelerator(value interface{}) {
	if err := c.validatePutGuestAcceleratorParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putGuestAccelerator",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ComputeInstance) PutNetworkInterface(value interface{}) {
	if err := c.validatePutNetworkInterfaceParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putNetworkInterface",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ComputeInstance) PutNetworkPerformanceConfig(value *ComputeInstanceNetworkPerformanceConfig) {
	if err := c.validatePutNetworkPerformanceConfigParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putNetworkPerformanceConfig",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ComputeInstance) PutParams(value *ComputeInstanceParams) {
	if err := c.validatePutParamsParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putParams",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ComputeInstance) PutReservationAffinity(value *ComputeInstanceReservationAffinity) {
	if err := c.validatePutReservationAffinityParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putReservationAffinity",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ComputeInstance) PutScheduling(value *ComputeInstanceScheduling) {
	if err := c.validatePutSchedulingParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putScheduling",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ComputeInstance) PutScratchDisk(value interface{}) {
	if err := c.validatePutScratchDiskParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putScratchDisk",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ComputeInstance) PutServiceAccount(value *ComputeInstanceServiceAccount) {
	if err := c.validatePutServiceAccountParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putServiceAccount",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ComputeInstance) PutShieldedInstanceConfig(value *ComputeInstanceShieldedInstanceConfig) {
	if err := c.validatePutShieldedInstanceConfigParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putShieldedInstanceConfig",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ComputeInstance) PutTimeouts(value *ComputeInstanceTimeouts) {
	if err := c.validatePutTimeoutsParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putTimeouts",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ComputeInstance) ResetAdvancedMachineFeatures() {
	_jsii_.InvokeVoid(
		c,
		"resetAdvancedMachineFeatures",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeInstance) ResetAllowStoppingForUpdate() {
	_jsii_.InvokeVoid(
		c,
		"resetAllowStoppingForUpdate",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeInstance) ResetAttachedDisk() {
	_jsii_.InvokeVoid(
		c,
		"resetAttachedDisk",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeInstance) ResetCanIpForward() {
	_jsii_.InvokeVoid(
		c,
		"resetCanIpForward",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeInstance) ResetConfidentialInstanceConfig() {
	_jsii_.InvokeVoid(
		c,
		"resetConfidentialInstanceConfig",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeInstance) ResetDeletionProtection() {
	_jsii_.InvokeVoid(
		c,
		"resetDeletionProtection",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeInstance) ResetDescription() {
	_jsii_.InvokeVoid(
		c,
		"resetDescription",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeInstance) ResetDesiredStatus() {
	_jsii_.InvokeVoid(
		c,
		"resetDesiredStatus",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeInstance) ResetEnableDisplay() {
	_jsii_.InvokeVoid(
		c,
		"resetEnableDisplay",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeInstance) ResetGuestAccelerator() {
	_jsii_.InvokeVoid(
		c,
		"resetGuestAccelerator",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeInstance) ResetHostname() {
	_jsii_.InvokeVoid(
		c,
		"resetHostname",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeInstance) ResetId() {
	_jsii_.InvokeVoid(
		c,
		"resetId",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeInstance) ResetLabels() {
	_jsii_.InvokeVoid(
		c,
		"resetLabels",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeInstance) ResetMetadata() {
	_jsii_.InvokeVoid(
		c,
		"resetMetadata",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeInstance) ResetMetadataStartupScript() {
	_jsii_.InvokeVoid(
		c,
		"resetMetadataStartupScript",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeInstance) ResetMinCpuPlatform() {
	_jsii_.InvokeVoid(
		c,
		"resetMinCpuPlatform",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeInstance) ResetNetworkPerformanceConfig() {
	_jsii_.InvokeVoid(
		c,
		"resetNetworkPerformanceConfig",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeInstance) ResetOverrideLogicalId() {
	_jsii_.InvokeVoid(
		c,
		"resetOverrideLogicalId",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeInstance) ResetParams() {
	_jsii_.InvokeVoid(
		c,
		"resetParams",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeInstance) ResetProject() {
	_jsii_.InvokeVoid(
		c,
		"resetProject",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeInstance) ResetReservationAffinity() {
	_jsii_.InvokeVoid(
		c,
		"resetReservationAffinity",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeInstance) ResetResourcePolicies() {
	_jsii_.InvokeVoid(
		c,
		"resetResourcePolicies",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeInstance) ResetScheduling() {
	_jsii_.InvokeVoid(
		c,
		"resetScheduling",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeInstance) ResetScratchDisk() {
	_jsii_.InvokeVoid(
		c,
		"resetScratchDisk",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeInstance) ResetServiceAccount() {
	_jsii_.InvokeVoid(
		c,
		"resetServiceAccount",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeInstance) ResetShieldedInstanceConfig() {
	_jsii_.InvokeVoid(
		c,
		"resetShieldedInstanceConfig",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeInstance) ResetTags() {
	_jsii_.InvokeVoid(
		c,
		"resetTags",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeInstance) ResetTimeouts() {
	_jsii_.InvokeVoid(
		c,
		"resetTimeouts",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeInstance) ResetZone() {
	_jsii_.InvokeVoid(
		c,
		"resetZone",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeInstance) SynthesizeAttributes() *map[string]interface{} {
	var returns *map[string]interface{}

	_jsii_.Invoke(
		c,
		"synthesizeAttributes",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComputeInstance) ToMetadata() interface{} {
	var returns interface{}

	_jsii_.Invoke(
		c,
		"toMetadata",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComputeInstance) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		c,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComputeInstance) ToTerraform() interface{} {
	var returns interface{}

	_jsii_.Invoke(
		c,
		"toTerraform",
		nil, // no parameters
		&returns,
	)

	return returns
}

