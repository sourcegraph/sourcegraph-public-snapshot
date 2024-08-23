package sqldatabaseinstance

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/sourcegraph/managed-services-platform-cdktf/gen/google/jsii"

	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/sqldatabaseinstance/internal"
)

type SqlDatabaseInstanceSettingsOutputReference interface {
	cdktf.ComplexObject
	ActivationPolicy() *string
	SetActivationPolicy(val *string)
	ActivationPolicyInput() *string
	ActiveDirectoryConfig() SqlDatabaseInstanceSettingsActiveDirectoryConfigOutputReference
	ActiveDirectoryConfigInput() *SqlDatabaseInstanceSettingsActiveDirectoryConfig
	AdvancedMachineFeatures() SqlDatabaseInstanceSettingsAdvancedMachineFeaturesOutputReference
	AdvancedMachineFeaturesInput() *SqlDatabaseInstanceSettingsAdvancedMachineFeatures
	AvailabilityType() *string
	SetAvailabilityType(val *string)
	AvailabilityTypeInput() *string
	BackupConfiguration() SqlDatabaseInstanceSettingsBackupConfigurationOutputReference
	BackupConfigurationInput() *SqlDatabaseInstanceSettingsBackupConfiguration
	Collation() *string
	SetCollation(val *string)
	CollationInput() *string
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
	ConnectorEnforcement() *string
	SetConnectorEnforcement(val *string)
	ConnectorEnforcementInput() *string
	// The creation stack of this resolvable which will be appended to errors thrown during resolution.
	//
	// If this returns an empty array the stack will not be attached.
	// Experimental.
	CreationStack() *[]*string
	DatabaseFlags() SqlDatabaseInstanceSettingsDatabaseFlagsList
	DatabaseFlagsInput() interface{}
	DataCacheConfig() SqlDatabaseInstanceSettingsDataCacheConfigOutputReference
	DataCacheConfigInput() *SqlDatabaseInstanceSettingsDataCacheConfig
	DeletionProtectionEnabled() interface{}
	SetDeletionProtectionEnabled(val interface{})
	DeletionProtectionEnabledInput() interface{}
	DenyMaintenancePeriod() SqlDatabaseInstanceSettingsDenyMaintenancePeriodOutputReference
	DenyMaintenancePeriodInput() *SqlDatabaseInstanceSettingsDenyMaintenancePeriod
	DiskAutoresize() interface{}
	SetDiskAutoresize(val interface{})
	DiskAutoresizeInput() interface{}
	DiskAutoresizeLimit() *float64
	SetDiskAutoresizeLimit(val *float64)
	DiskAutoresizeLimitInput() *float64
	DiskSize() *float64
	SetDiskSize(val *float64)
	DiskSizeInput() *float64
	DiskType() *string
	SetDiskType(val *string)
	DiskTypeInput() *string
	Edition() *string
	SetEdition(val *string)
	EditionInput() *string
	EnableGoogleMlIntegration() interface{}
	SetEnableGoogleMlIntegration(val interface{})
	EnableGoogleMlIntegrationInput() interface{}
	// Experimental.
	Fqn() *string
	InsightsConfig() SqlDatabaseInstanceSettingsInsightsConfigOutputReference
	InsightsConfigInput() *SqlDatabaseInstanceSettingsInsightsConfig
	InternalValue() *SqlDatabaseInstanceSettings
	SetInternalValue(val *SqlDatabaseInstanceSettings)
	IpConfiguration() SqlDatabaseInstanceSettingsIpConfigurationOutputReference
	IpConfigurationInput() *SqlDatabaseInstanceSettingsIpConfiguration
	LocationPreference() SqlDatabaseInstanceSettingsLocationPreferenceOutputReference
	LocationPreferenceInput() *SqlDatabaseInstanceSettingsLocationPreference
	MaintenanceWindow() SqlDatabaseInstanceSettingsMaintenanceWindowOutputReference
	MaintenanceWindowInput() *SqlDatabaseInstanceSettingsMaintenanceWindow
	PasswordValidationPolicy() SqlDatabaseInstanceSettingsPasswordValidationPolicyOutputReference
	PasswordValidationPolicyInput() *SqlDatabaseInstanceSettingsPasswordValidationPolicy
	PricingPlan() *string
	SetPricingPlan(val *string)
	PricingPlanInput() *string
	SqlServerAuditConfig() SqlDatabaseInstanceSettingsSqlServerAuditConfigOutputReference
	SqlServerAuditConfigInput() *SqlDatabaseInstanceSettingsSqlServerAuditConfig
	// Experimental.
	TerraformAttribute() *string
	// Experimental.
	SetTerraformAttribute(val *string)
	// Experimental.
	TerraformResource() cdktf.IInterpolatingParent
	// Experimental.
	SetTerraformResource(val cdktf.IInterpolatingParent)
	Tier() *string
	SetTier(val *string)
	TierInput() *string
	TimeZone() *string
	SetTimeZone(val *string)
	TimeZoneInput() *string
	UserLabels() *map[string]*string
	SetUserLabels(val *map[string]*string)
	UserLabelsInput() *map[string]*string
	Version() *float64
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
	PutActiveDirectoryConfig(value *SqlDatabaseInstanceSettingsActiveDirectoryConfig)
	PutAdvancedMachineFeatures(value *SqlDatabaseInstanceSettingsAdvancedMachineFeatures)
	PutBackupConfiguration(value *SqlDatabaseInstanceSettingsBackupConfiguration)
	PutDatabaseFlags(value interface{})
	PutDataCacheConfig(value *SqlDatabaseInstanceSettingsDataCacheConfig)
	PutDenyMaintenancePeriod(value *SqlDatabaseInstanceSettingsDenyMaintenancePeriod)
	PutInsightsConfig(value *SqlDatabaseInstanceSettingsInsightsConfig)
	PutIpConfiguration(value *SqlDatabaseInstanceSettingsIpConfiguration)
	PutLocationPreference(value *SqlDatabaseInstanceSettingsLocationPreference)
	PutMaintenanceWindow(value *SqlDatabaseInstanceSettingsMaintenanceWindow)
	PutPasswordValidationPolicy(value *SqlDatabaseInstanceSettingsPasswordValidationPolicy)
	PutSqlServerAuditConfig(value *SqlDatabaseInstanceSettingsSqlServerAuditConfig)
	ResetActivationPolicy()
	ResetActiveDirectoryConfig()
	ResetAdvancedMachineFeatures()
	ResetAvailabilityType()
	ResetBackupConfiguration()
	ResetCollation()
	ResetConnectorEnforcement()
	ResetDatabaseFlags()
	ResetDataCacheConfig()
	ResetDeletionProtectionEnabled()
	ResetDenyMaintenancePeriod()
	ResetDiskAutoresize()
	ResetDiskAutoresizeLimit()
	ResetDiskSize()
	ResetDiskType()
	ResetEdition()
	ResetEnableGoogleMlIntegration()
	ResetInsightsConfig()
	ResetIpConfiguration()
	ResetLocationPreference()
	ResetMaintenanceWindow()
	ResetPasswordValidationPolicy()
	ResetPricingPlan()
	ResetSqlServerAuditConfig()
	ResetTimeZone()
	ResetUserLabels()
	// Produce the Token's value at resolution time.
	// Experimental.
	Resolve(_context cdktf.IResolveContext) interface{}
	// Return a string representation of this resolvable object.
	//
	// Returns a reversible string representation.
	// Experimental.
	ToString() *string
}

// The jsii proxy struct for SqlDatabaseInstanceSettingsOutputReference
type jsiiProxy_SqlDatabaseInstanceSettingsOutputReference struct {
	internal.Type__cdktfComplexObject
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) ActivationPolicy() *string {
	var returns *string
	_jsii_.Get(
		j,
		"activationPolicy",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) ActivationPolicyInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"activationPolicyInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) ActiveDirectoryConfig() SqlDatabaseInstanceSettingsActiveDirectoryConfigOutputReference {
	var returns SqlDatabaseInstanceSettingsActiveDirectoryConfigOutputReference
	_jsii_.Get(
		j,
		"activeDirectoryConfig",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) ActiveDirectoryConfigInput() *SqlDatabaseInstanceSettingsActiveDirectoryConfig {
	var returns *SqlDatabaseInstanceSettingsActiveDirectoryConfig
	_jsii_.Get(
		j,
		"activeDirectoryConfigInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) AdvancedMachineFeatures() SqlDatabaseInstanceSettingsAdvancedMachineFeaturesOutputReference {
	var returns SqlDatabaseInstanceSettingsAdvancedMachineFeaturesOutputReference
	_jsii_.Get(
		j,
		"advancedMachineFeatures",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) AdvancedMachineFeaturesInput() *SqlDatabaseInstanceSettingsAdvancedMachineFeatures {
	var returns *SqlDatabaseInstanceSettingsAdvancedMachineFeatures
	_jsii_.Get(
		j,
		"advancedMachineFeaturesInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) AvailabilityType() *string {
	var returns *string
	_jsii_.Get(
		j,
		"availabilityType",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) AvailabilityTypeInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"availabilityTypeInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) BackupConfiguration() SqlDatabaseInstanceSettingsBackupConfigurationOutputReference {
	var returns SqlDatabaseInstanceSettingsBackupConfigurationOutputReference
	_jsii_.Get(
		j,
		"backupConfiguration",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) BackupConfigurationInput() *SqlDatabaseInstanceSettingsBackupConfiguration {
	var returns *SqlDatabaseInstanceSettingsBackupConfiguration
	_jsii_.Get(
		j,
		"backupConfigurationInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) Collation() *string {
	var returns *string
	_jsii_.Get(
		j,
		"collation",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) CollationInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"collationInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) ComplexObjectIndex() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"complexObjectIndex",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) ComplexObjectIsFromSet() *bool {
	var returns *bool
	_jsii_.Get(
		j,
		"complexObjectIsFromSet",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) ConnectorEnforcement() *string {
	var returns *string
	_jsii_.Get(
		j,
		"connectorEnforcement",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) ConnectorEnforcementInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"connectorEnforcementInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) CreationStack() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"creationStack",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) DatabaseFlags() SqlDatabaseInstanceSettingsDatabaseFlagsList {
	var returns SqlDatabaseInstanceSettingsDatabaseFlagsList
	_jsii_.Get(
		j,
		"databaseFlags",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) DatabaseFlagsInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"databaseFlagsInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) DataCacheConfig() SqlDatabaseInstanceSettingsDataCacheConfigOutputReference {
	var returns SqlDatabaseInstanceSettingsDataCacheConfigOutputReference
	_jsii_.Get(
		j,
		"dataCacheConfig",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) DataCacheConfigInput() *SqlDatabaseInstanceSettingsDataCacheConfig {
	var returns *SqlDatabaseInstanceSettingsDataCacheConfig
	_jsii_.Get(
		j,
		"dataCacheConfigInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) DeletionProtectionEnabled() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"deletionProtectionEnabled",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) DeletionProtectionEnabledInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"deletionProtectionEnabledInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) DenyMaintenancePeriod() SqlDatabaseInstanceSettingsDenyMaintenancePeriodOutputReference {
	var returns SqlDatabaseInstanceSettingsDenyMaintenancePeriodOutputReference
	_jsii_.Get(
		j,
		"denyMaintenancePeriod",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) DenyMaintenancePeriodInput() *SqlDatabaseInstanceSettingsDenyMaintenancePeriod {
	var returns *SqlDatabaseInstanceSettingsDenyMaintenancePeriod
	_jsii_.Get(
		j,
		"denyMaintenancePeriodInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) DiskAutoresize() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"diskAutoresize",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) DiskAutoresizeInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"diskAutoresizeInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) DiskAutoresizeLimit() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"diskAutoresizeLimit",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) DiskAutoresizeLimitInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"diskAutoresizeLimitInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) DiskSize() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"diskSize",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) DiskSizeInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"diskSizeInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) DiskType() *string {
	var returns *string
	_jsii_.Get(
		j,
		"diskType",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) DiskTypeInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"diskTypeInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) Edition() *string {
	var returns *string
	_jsii_.Get(
		j,
		"edition",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) EditionInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"editionInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) EnableGoogleMlIntegration() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"enableGoogleMlIntegration",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) EnableGoogleMlIntegrationInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"enableGoogleMlIntegrationInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) Fqn() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fqn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) InsightsConfig() SqlDatabaseInstanceSettingsInsightsConfigOutputReference {
	var returns SqlDatabaseInstanceSettingsInsightsConfigOutputReference
	_jsii_.Get(
		j,
		"insightsConfig",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) InsightsConfigInput() *SqlDatabaseInstanceSettingsInsightsConfig {
	var returns *SqlDatabaseInstanceSettingsInsightsConfig
	_jsii_.Get(
		j,
		"insightsConfigInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) InternalValue() *SqlDatabaseInstanceSettings {
	var returns *SqlDatabaseInstanceSettings
	_jsii_.Get(
		j,
		"internalValue",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) IpConfiguration() SqlDatabaseInstanceSettingsIpConfigurationOutputReference {
	var returns SqlDatabaseInstanceSettingsIpConfigurationOutputReference
	_jsii_.Get(
		j,
		"ipConfiguration",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) IpConfigurationInput() *SqlDatabaseInstanceSettingsIpConfiguration {
	var returns *SqlDatabaseInstanceSettingsIpConfiguration
	_jsii_.Get(
		j,
		"ipConfigurationInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) LocationPreference() SqlDatabaseInstanceSettingsLocationPreferenceOutputReference {
	var returns SqlDatabaseInstanceSettingsLocationPreferenceOutputReference
	_jsii_.Get(
		j,
		"locationPreference",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) LocationPreferenceInput() *SqlDatabaseInstanceSettingsLocationPreference {
	var returns *SqlDatabaseInstanceSettingsLocationPreference
	_jsii_.Get(
		j,
		"locationPreferenceInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) MaintenanceWindow() SqlDatabaseInstanceSettingsMaintenanceWindowOutputReference {
	var returns SqlDatabaseInstanceSettingsMaintenanceWindowOutputReference
	_jsii_.Get(
		j,
		"maintenanceWindow",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) MaintenanceWindowInput() *SqlDatabaseInstanceSettingsMaintenanceWindow {
	var returns *SqlDatabaseInstanceSettingsMaintenanceWindow
	_jsii_.Get(
		j,
		"maintenanceWindowInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) PasswordValidationPolicy() SqlDatabaseInstanceSettingsPasswordValidationPolicyOutputReference {
	var returns SqlDatabaseInstanceSettingsPasswordValidationPolicyOutputReference
	_jsii_.Get(
		j,
		"passwordValidationPolicy",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) PasswordValidationPolicyInput() *SqlDatabaseInstanceSettingsPasswordValidationPolicy {
	var returns *SqlDatabaseInstanceSettingsPasswordValidationPolicy
	_jsii_.Get(
		j,
		"passwordValidationPolicyInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) PricingPlan() *string {
	var returns *string
	_jsii_.Get(
		j,
		"pricingPlan",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) PricingPlanInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"pricingPlanInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) SqlServerAuditConfig() SqlDatabaseInstanceSettingsSqlServerAuditConfigOutputReference {
	var returns SqlDatabaseInstanceSettingsSqlServerAuditConfigOutputReference
	_jsii_.Get(
		j,
		"sqlServerAuditConfig",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) SqlServerAuditConfigInput() *SqlDatabaseInstanceSettingsSqlServerAuditConfig {
	var returns *SqlDatabaseInstanceSettingsSqlServerAuditConfig
	_jsii_.Get(
		j,
		"sqlServerAuditConfigInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) TerraformAttribute() *string {
	var returns *string
	_jsii_.Get(
		j,
		"terraformAttribute",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) TerraformResource() cdktf.IInterpolatingParent {
	var returns cdktf.IInterpolatingParent
	_jsii_.Get(
		j,
		"terraformResource",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) Tier() *string {
	var returns *string
	_jsii_.Get(
		j,
		"tier",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) TierInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"tierInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) TimeZone() *string {
	var returns *string
	_jsii_.Get(
		j,
		"timeZone",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) TimeZoneInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"timeZoneInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) UserLabels() *map[string]*string {
	var returns *map[string]*string
	_jsii_.Get(
		j,
		"userLabels",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) UserLabelsInput() *map[string]*string {
	var returns *map[string]*string
	_jsii_.Get(
		j,
		"userLabelsInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) Version() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"version",
		&returns,
	)
	return returns
}


func NewSqlDatabaseInstanceSettingsOutputReference(terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) SqlDatabaseInstanceSettingsOutputReference {
	_init_.Initialize()

	if err := validateNewSqlDatabaseInstanceSettingsOutputReferenceParameters(terraformResource, terraformAttribute); err != nil {
		panic(err)
	}
	j := jsiiProxy_SqlDatabaseInstanceSettingsOutputReference{}

	_jsii_.Create(
		"@cdktf/provider-google.sqlDatabaseInstance.SqlDatabaseInstanceSettingsOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		&j,
	)

	return &j
}

func NewSqlDatabaseInstanceSettingsOutputReference_Override(s SqlDatabaseInstanceSettingsOutputReference, terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) {
	_init_.Initialize()

	_jsii_.Create(
		"@cdktf/provider-google.sqlDatabaseInstance.SqlDatabaseInstanceSettingsOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		s,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference)SetActivationPolicy(val *string) {
	if err := j.validateSetActivationPolicyParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"activationPolicy",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference)SetAvailabilityType(val *string) {
	if err := j.validateSetAvailabilityTypeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"availabilityType",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference)SetCollation(val *string) {
	if err := j.validateSetCollationParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"collation",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference)SetComplexObjectIndex(val interface{}) {
	if err := j.validateSetComplexObjectIndexParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIndex",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference)SetComplexObjectIsFromSet(val *bool) {
	if err := j.validateSetComplexObjectIsFromSetParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIsFromSet",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference)SetConnectorEnforcement(val *string) {
	if err := j.validateSetConnectorEnforcementParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"connectorEnforcement",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference)SetDeletionProtectionEnabled(val interface{}) {
	if err := j.validateSetDeletionProtectionEnabledParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"deletionProtectionEnabled",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference)SetDiskAutoresize(val interface{}) {
	if err := j.validateSetDiskAutoresizeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"diskAutoresize",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference)SetDiskAutoresizeLimit(val *float64) {
	if err := j.validateSetDiskAutoresizeLimitParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"diskAutoresizeLimit",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference)SetDiskSize(val *float64) {
	if err := j.validateSetDiskSizeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"diskSize",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference)SetDiskType(val *string) {
	if err := j.validateSetDiskTypeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"diskType",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference)SetEdition(val *string) {
	if err := j.validateSetEditionParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"edition",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference)SetEnableGoogleMlIntegration(val interface{}) {
	if err := j.validateSetEnableGoogleMlIntegrationParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"enableGoogleMlIntegration",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference)SetInternalValue(val *SqlDatabaseInstanceSettings) {
	if err := j.validateSetInternalValueParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"internalValue",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference)SetPricingPlan(val *string) {
	if err := j.validateSetPricingPlanParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"pricingPlan",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference)SetTerraformAttribute(val *string) {
	if err := j.validateSetTerraformAttributeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformAttribute",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference)SetTerraformResource(val cdktf.IInterpolatingParent) {
	if err := j.validateSetTerraformResourceParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformResource",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference)SetTier(val *string) {
	if err := j.validateSetTierParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"tier",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference)SetTimeZone(val *string) {
	if err := j.validateSetTimeZoneParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"timeZone",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference)SetUserLabels(val *map[string]*string) {
	if err := j.validateSetUserLabelsParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"userLabels",
		val,
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) ComputeFqn() *string {
	var returns *string

	_jsii_.Invoke(
		s,
		"computeFqn",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) GetAnyMapAttribute(terraformAttribute *string) *map[string]interface{} {
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

func (s *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) GetBooleanAttribute(terraformAttribute *string) cdktf.IResolvable {
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

func (s *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) GetBooleanMapAttribute(terraformAttribute *string) *map[string]*bool {
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

func (s *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) GetListAttribute(terraformAttribute *string) *[]*string {
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

func (s *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) GetNumberAttribute(terraformAttribute *string) *float64 {
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

func (s *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) GetNumberListAttribute(terraformAttribute *string) *[]*float64 {
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

func (s *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) GetNumberMapAttribute(terraformAttribute *string) *map[string]*float64 {
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

func (s *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) GetStringAttribute(terraformAttribute *string) *string {
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

func (s *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) GetStringMapAttribute(terraformAttribute *string) *map[string]*string {
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

func (s *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) InterpolationAsList() cdktf.IResolvable {
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		s,
		"interpolationAsList",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) InterpolationForAttribute(property *string) cdktf.IResolvable {
	if err := s.validateInterpolationForAttributeParameters(property); err != nil {
		panic(err)
	}
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		s,
		"interpolationForAttribute",
		[]interface{}{property},
		&returns,
	)

	return returns
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) PutActiveDirectoryConfig(value *SqlDatabaseInstanceSettingsActiveDirectoryConfig) {
	if err := s.validatePutActiveDirectoryConfigParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		s,
		"putActiveDirectoryConfig",
		[]interface{}{value},
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) PutAdvancedMachineFeatures(value *SqlDatabaseInstanceSettingsAdvancedMachineFeatures) {
	if err := s.validatePutAdvancedMachineFeaturesParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		s,
		"putAdvancedMachineFeatures",
		[]interface{}{value},
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) PutBackupConfiguration(value *SqlDatabaseInstanceSettingsBackupConfiguration) {
	if err := s.validatePutBackupConfigurationParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		s,
		"putBackupConfiguration",
		[]interface{}{value},
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) PutDatabaseFlags(value interface{}) {
	if err := s.validatePutDatabaseFlagsParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		s,
		"putDatabaseFlags",
		[]interface{}{value},
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) PutDataCacheConfig(value *SqlDatabaseInstanceSettingsDataCacheConfig) {
	if err := s.validatePutDataCacheConfigParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		s,
		"putDataCacheConfig",
		[]interface{}{value},
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) PutDenyMaintenancePeriod(value *SqlDatabaseInstanceSettingsDenyMaintenancePeriod) {
	if err := s.validatePutDenyMaintenancePeriodParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		s,
		"putDenyMaintenancePeriod",
		[]interface{}{value},
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) PutInsightsConfig(value *SqlDatabaseInstanceSettingsInsightsConfig) {
	if err := s.validatePutInsightsConfigParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		s,
		"putInsightsConfig",
		[]interface{}{value},
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) PutIpConfiguration(value *SqlDatabaseInstanceSettingsIpConfiguration) {
	if err := s.validatePutIpConfigurationParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		s,
		"putIpConfiguration",
		[]interface{}{value},
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) PutLocationPreference(value *SqlDatabaseInstanceSettingsLocationPreference) {
	if err := s.validatePutLocationPreferenceParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		s,
		"putLocationPreference",
		[]interface{}{value},
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) PutMaintenanceWindow(value *SqlDatabaseInstanceSettingsMaintenanceWindow) {
	if err := s.validatePutMaintenanceWindowParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		s,
		"putMaintenanceWindow",
		[]interface{}{value},
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) PutPasswordValidationPolicy(value *SqlDatabaseInstanceSettingsPasswordValidationPolicy) {
	if err := s.validatePutPasswordValidationPolicyParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		s,
		"putPasswordValidationPolicy",
		[]interface{}{value},
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) PutSqlServerAuditConfig(value *SqlDatabaseInstanceSettingsSqlServerAuditConfig) {
	if err := s.validatePutSqlServerAuditConfigParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		s,
		"putSqlServerAuditConfig",
		[]interface{}{value},
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) ResetActivationPolicy() {
	_jsii_.InvokeVoid(
		s,
		"resetActivationPolicy",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) ResetActiveDirectoryConfig() {
	_jsii_.InvokeVoid(
		s,
		"resetActiveDirectoryConfig",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) ResetAdvancedMachineFeatures() {
	_jsii_.InvokeVoid(
		s,
		"resetAdvancedMachineFeatures",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) ResetAvailabilityType() {
	_jsii_.InvokeVoid(
		s,
		"resetAvailabilityType",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) ResetBackupConfiguration() {
	_jsii_.InvokeVoid(
		s,
		"resetBackupConfiguration",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) ResetCollation() {
	_jsii_.InvokeVoid(
		s,
		"resetCollation",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) ResetConnectorEnforcement() {
	_jsii_.InvokeVoid(
		s,
		"resetConnectorEnforcement",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) ResetDatabaseFlags() {
	_jsii_.InvokeVoid(
		s,
		"resetDatabaseFlags",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) ResetDataCacheConfig() {
	_jsii_.InvokeVoid(
		s,
		"resetDataCacheConfig",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) ResetDeletionProtectionEnabled() {
	_jsii_.InvokeVoid(
		s,
		"resetDeletionProtectionEnabled",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) ResetDenyMaintenancePeriod() {
	_jsii_.InvokeVoid(
		s,
		"resetDenyMaintenancePeriod",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) ResetDiskAutoresize() {
	_jsii_.InvokeVoid(
		s,
		"resetDiskAutoresize",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) ResetDiskAutoresizeLimit() {
	_jsii_.InvokeVoid(
		s,
		"resetDiskAutoresizeLimit",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) ResetDiskSize() {
	_jsii_.InvokeVoid(
		s,
		"resetDiskSize",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) ResetDiskType() {
	_jsii_.InvokeVoid(
		s,
		"resetDiskType",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) ResetEdition() {
	_jsii_.InvokeVoid(
		s,
		"resetEdition",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) ResetEnableGoogleMlIntegration() {
	_jsii_.InvokeVoid(
		s,
		"resetEnableGoogleMlIntegration",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) ResetInsightsConfig() {
	_jsii_.InvokeVoid(
		s,
		"resetInsightsConfig",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) ResetIpConfiguration() {
	_jsii_.InvokeVoid(
		s,
		"resetIpConfiguration",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) ResetLocationPreference() {
	_jsii_.InvokeVoid(
		s,
		"resetLocationPreference",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) ResetMaintenanceWindow() {
	_jsii_.InvokeVoid(
		s,
		"resetMaintenanceWindow",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) ResetPasswordValidationPolicy() {
	_jsii_.InvokeVoid(
		s,
		"resetPasswordValidationPolicy",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) ResetPricingPlan() {
	_jsii_.InvokeVoid(
		s,
		"resetPricingPlan",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) ResetSqlServerAuditConfig() {
	_jsii_.InvokeVoid(
		s,
		"resetSqlServerAuditConfig",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) ResetTimeZone() {
	_jsii_.InvokeVoid(
		s,
		"resetTimeZone",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) ResetUserLabels() {
	_jsii_.InvokeVoid(
		s,
		"resetUserLabels",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) Resolve(_context cdktf.IResolveContext) interface{} {
	if err := s.validateResolveParameters(_context); err != nil {
		panic(err)
	}
	var returns interface{}

	_jsii_.Invoke(
		s,
		"resolve",
		[]interface{}{_context},
		&returns,
	)

	return returns
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsOutputReference) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		s,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

