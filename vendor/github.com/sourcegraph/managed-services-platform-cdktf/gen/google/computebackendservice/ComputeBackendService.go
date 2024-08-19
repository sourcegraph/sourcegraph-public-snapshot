package computebackendservice

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/sourcegraph/managed-services-platform-cdktf/gen/google/jsii"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/computebackendservice/internal"
)

// Represents a {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_backend_service google_compute_backend_service}.
type ComputeBackendService interface {
	cdktf.TerraformResource
	AffinityCookieTtlSec() *float64
	SetAffinityCookieTtlSec(val *float64)
	AffinityCookieTtlSecInput() *float64
	Backend() ComputeBackendServiceBackendList
	BackendInput() interface{}
	// Experimental.
	CdktfStack() cdktf.TerraformStack
	CdnPolicy() ComputeBackendServiceCdnPolicyOutputReference
	CdnPolicyInput() *ComputeBackendServiceCdnPolicy
	CircuitBreakers() ComputeBackendServiceCircuitBreakersOutputReference
	CircuitBreakersInput() *ComputeBackendServiceCircuitBreakers
	CompressionMode() *string
	SetCompressionMode(val *string)
	CompressionModeInput() *string
	// Experimental.
	Connection() interface{}
	// Experimental.
	SetConnection(val interface{})
	ConnectionDrainingTimeoutSec() *float64
	SetConnectionDrainingTimeoutSec(val *float64)
	ConnectionDrainingTimeoutSecInput() *float64
	ConsistentHash() ComputeBackendServiceConsistentHashOutputReference
	ConsistentHashInput() *ComputeBackendServiceConsistentHash
	// Experimental.
	ConstructNodeMetadata() *map[string]interface{}
	// Experimental.
	Count() interface{}
	// Experimental.
	SetCount(val interface{})
	CreationTimestamp() *string
	CustomRequestHeaders() *[]*string
	SetCustomRequestHeaders(val *[]*string)
	CustomRequestHeadersInput() *[]*string
	CustomResponseHeaders() *[]*string
	SetCustomResponseHeaders(val *[]*string)
	CustomResponseHeadersInput() *[]*string
	// Experimental.
	DependsOn() *[]*string
	// Experimental.
	SetDependsOn(val *[]*string)
	Description() *string
	SetDescription(val *string)
	DescriptionInput() *string
	EdgeSecurityPolicy() *string
	SetEdgeSecurityPolicy(val *string)
	EdgeSecurityPolicyInput() *string
	EnableCdn() interface{}
	SetEnableCdn(val interface{})
	EnableCdnInput() interface{}
	Fingerprint() *string
	// Experimental.
	ForEach() cdktf.ITerraformIterator
	// Experimental.
	SetForEach(val cdktf.ITerraformIterator)
	// Experimental.
	Fqn() *string
	// Experimental.
	FriendlyUniqueId() *string
	GeneratedId() *float64
	HealthChecks() *[]*string
	SetHealthChecks(val *[]*string)
	HealthChecksInput() *[]*string
	Iap() ComputeBackendServiceIapOutputReference
	IapInput() *ComputeBackendServiceIap
	Id() *string
	SetId(val *string)
	IdInput() *string
	// Experimental.
	Lifecycle() *cdktf.TerraformResourceLifecycle
	// Experimental.
	SetLifecycle(val *cdktf.TerraformResourceLifecycle)
	LoadBalancingScheme() *string
	SetLoadBalancingScheme(val *string)
	LoadBalancingSchemeInput() *string
	LocalityLbPolicies() ComputeBackendServiceLocalityLbPoliciesList
	LocalityLbPoliciesInput() interface{}
	LocalityLbPolicy() *string
	SetLocalityLbPolicy(val *string)
	LocalityLbPolicyInput() *string
	LogConfig() ComputeBackendServiceLogConfigOutputReference
	LogConfigInput() *ComputeBackendServiceLogConfig
	Name() *string
	SetName(val *string)
	NameInput() *string
	// The tree node.
	Node() constructs.Node
	OutlierDetection() ComputeBackendServiceOutlierDetectionOutputReference
	OutlierDetectionInput() *ComputeBackendServiceOutlierDetection
	PortName() *string
	SetPortName(val *string)
	PortNameInput() *string
	Project() *string
	SetProject(val *string)
	ProjectInput() *string
	Protocol() *string
	SetProtocol(val *string)
	ProtocolInput() *string
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
	SecurityPolicy() *string
	SetSecurityPolicy(val *string)
	SecurityPolicyInput() *string
	SecuritySettings() ComputeBackendServiceSecuritySettingsOutputReference
	SecuritySettingsInput() *ComputeBackendServiceSecuritySettings
	SelfLink() *string
	SessionAffinity() *string
	SetSessionAffinity(val *string)
	SessionAffinityInput() *string
	// Experimental.
	TerraformGeneratorMetadata() *cdktf.TerraformProviderGeneratorMetadata
	// Experimental.
	TerraformMetaArguments() *map[string]interface{}
	// Experimental.
	TerraformResourceType() *string
	Timeouts() ComputeBackendServiceTimeoutsOutputReference
	TimeoutSec() *float64
	SetTimeoutSec(val *float64)
	TimeoutSecInput() *float64
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
	PutBackend(value interface{})
	PutCdnPolicy(value *ComputeBackendServiceCdnPolicy)
	PutCircuitBreakers(value *ComputeBackendServiceCircuitBreakers)
	PutConsistentHash(value *ComputeBackendServiceConsistentHash)
	PutIap(value *ComputeBackendServiceIap)
	PutLocalityLbPolicies(value interface{})
	PutLogConfig(value *ComputeBackendServiceLogConfig)
	PutOutlierDetection(value *ComputeBackendServiceOutlierDetection)
	PutSecuritySettings(value *ComputeBackendServiceSecuritySettings)
	PutTimeouts(value *ComputeBackendServiceTimeouts)
	ResetAffinityCookieTtlSec()
	ResetBackend()
	ResetCdnPolicy()
	ResetCircuitBreakers()
	ResetCompressionMode()
	ResetConnectionDrainingTimeoutSec()
	ResetConsistentHash()
	ResetCustomRequestHeaders()
	ResetCustomResponseHeaders()
	ResetDescription()
	ResetEdgeSecurityPolicy()
	ResetEnableCdn()
	ResetHealthChecks()
	ResetIap()
	ResetId()
	ResetLoadBalancingScheme()
	ResetLocalityLbPolicies()
	ResetLocalityLbPolicy()
	ResetLogConfig()
	ResetOutlierDetection()
	// Resets a previously passed logical Id to use the auto-generated logical id again.
	// Experimental.
	ResetOverrideLogicalId()
	ResetPortName()
	ResetProject()
	ResetProtocol()
	ResetSecurityPolicy()
	ResetSecuritySettings()
	ResetSessionAffinity()
	ResetTimeouts()
	ResetTimeoutSec()
	SynthesizeAttributes() *map[string]interface{}
	// Experimental.
	ToMetadata() interface{}
	// Returns a string representation of this construct.
	ToString() *string
	// Adds this resource to the terraform JSON output.
	// Experimental.
	ToTerraform() interface{}
}

// The jsii proxy struct for ComputeBackendService
type jsiiProxy_ComputeBackendService struct {
	internal.Type__cdktfTerraformResource
}

func (j *jsiiProxy_ComputeBackendService) AffinityCookieTtlSec() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"affinityCookieTtlSec",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) AffinityCookieTtlSecInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"affinityCookieTtlSecInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) Backend() ComputeBackendServiceBackendList {
	var returns ComputeBackendServiceBackendList
	_jsii_.Get(
		j,
		"backend",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) BackendInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"backendInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) CdktfStack() cdktf.TerraformStack {
	var returns cdktf.TerraformStack
	_jsii_.Get(
		j,
		"cdktfStack",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) CdnPolicy() ComputeBackendServiceCdnPolicyOutputReference {
	var returns ComputeBackendServiceCdnPolicyOutputReference
	_jsii_.Get(
		j,
		"cdnPolicy",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) CdnPolicyInput() *ComputeBackendServiceCdnPolicy {
	var returns *ComputeBackendServiceCdnPolicy
	_jsii_.Get(
		j,
		"cdnPolicyInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) CircuitBreakers() ComputeBackendServiceCircuitBreakersOutputReference {
	var returns ComputeBackendServiceCircuitBreakersOutputReference
	_jsii_.Get(
		j,
		"circuitBreakers",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) CircuitBreakersInput() *ComputeBackendServiceCircuitBreakers {
	var returns *ComputeBackendServiceCircuitBreakers
	_jsii_.Get(
		j,
		"circuitBreakersInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) CompressionMode() *string {
	var returns *string
	_jsii_.Get(
		j,
		"compressionMode",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) CompressionModeInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"compressionModeInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) Connection() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"connection",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) ConnectionDrainingTimeoutSec() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"connectionDrainingTimeoutSec",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) ConnectionDrainingTimeoutSecInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"connectionDrainingTimeoutSecInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) ConsistentHash() ComputeBackendServiceConsistentHashOutputReference {
	var returns ComputeBackendServiceConsistentHashOutputReference
	_jsii_.Get(
		j,
		"consistentHash",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) ConsistentHashInput() *ComputeBackendServiceConsistentHash {
	var returns *ComputeBackendServiceConsistentHash
	_jsii_.Get(
		j,
		"consistentHashInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) ConstructNodeMetadata() *map[string]interface{} {
	var returns *map[string]interface{}
	_jsii_.Get(
		j,
		"constructNodeMetadata",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) Count() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"count",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) CreationTimestamp() *string {
	var returns *string
	_jsii_.Get(
		j,
		"creationTimestamp",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) CustomRequestHeaders() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"customRequestHeaders",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) CustomRequestHeadersInput() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"customRequestHeadersInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) CustomResponseHeaders() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"customResponseHeaders",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) CustomResponseHeadersInput() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"customResponseHeadersInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) DependsOn() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"dependsOn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) Description() *string {
	var returns *string
	_jsii_.Get(
		j,
		"description",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) DescriptionInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"descriptionInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) EdgeSecurityPolicy() *string {
	var returns *string
	_jsii_.Get(
		j,
		"edgeSecurityPolicy",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) EdgeSecurityPolicyInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"edgeSecurityPolicyInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) EnableCdn() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"enableCdn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) EnableCdnInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"enableCdnInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) Fingerprint() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fingerprint",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) ForEach() cdktf.ITerraformIterator {
	var returns cdktf.ITerraformIterator
	_jsii_.Get(
		j,
		"forEach",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) Fqn() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fqn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) FriendlyUniqueId() *string {
	var returns *string
	_jsii_.Get(
		j,
		"friendlyUniqueId",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) GeneratedId() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"generatedId",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) HealthChecks() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"healthChecks",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) HealthChecksInput() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"healthChecksInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) Iap() ComputeBackendServiceIapOutputReference {
	var returns ComputeBackendServiceIapOutputReference
	_jsii_.Get(
		j,
		"iap",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) IapInput() *ComputeBackendServiceIap {
	var returns *ComputeBackendServiceIap
	_jsii_.Get(
		j,
		"iapInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) Id() *string {
	var returns *string
	_jsii_.Get(
		j,
		"id",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) IdInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"idInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) Lifecycle() *cdktf.TerraformResourceLifecycle {
	var returns *cdktf.TerraformResourceLifecycle
	_jsii_.Get(
		j,
		"lifecycle",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) LoadBalancingScheme() *string {
	var returns *string
	_jsii_.Get(
		j,
		"loadBalancingScheme",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) LoadBalancingSchemeInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"loadBalancingSchemeInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) LocalityLbPolicies() ComputeBackendServiceLocalityLbPoliciesList {
	var returns ComputeBackendServiceLocalityLbPoliciesList
	_jsii_.Get(
		j,
		"localityLbPolicies",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) LocalityLbPoliciesInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"localityLbPoliciesInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) LocalityLbPolicy() *string {
	var returns *string
	_jsii_.Get(
		j,
		"localityLbPolicy",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) LocalityLbPolicyInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"localityLbPolicyInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) LogConfig() ComputeBackendServiceLogConfigOutputReference {
	var returns ComputeBackendServiceLogConfigOutputReference
	_jsii_.Get(
		j,
		"logConfig",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) LogConfigInput() *ComputeBackendServiceLogConfig {
	var returns *ComputeBackendServiceLogConfig
	_jsii_.Get(
		j,
		"logConfigInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) Name() *string {
	var returns *string
	_jsii_.Get(
		j,
		"name",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) NameInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"nameInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) Node() constructs.Node {
	var returns constructs.Node
	_jsii_.Get(
		j,
		"node",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) OutlierDetection() ComputeBackendServiceOutlierDetectionOutputReference {
	var returns ComputeBackendServiceOutlierDetectionOutputReference
	_jsii_.Get(
		j,
		"outlierDetection",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) OutlierDetectionInput() *ComputeBackendServiceOutlierDetection {
	var returns *ComputeBackendServiceOutlierDetection
	_jsii_.Get(
		j,
		"outlierDetectionInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) PortName() *string {
	var returns *string
	_jsii_.Get(
		j,
		"portName",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) PortNameInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"portNameInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) Project() *string {
	var returns *string
	_jsii_.Get(
		j,
		"project",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) ProjectInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"projectInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) Protocol() *string {
	var returns *string
	_jsii_.Get(
		j,
		"protocol",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) ProtocolInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"protocolInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) Provider() cdktf.TerraformProvider {
	var returns cdktf.TerraformProvider
	_jsii_.Get(
		j,
		"provider",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) Provisioners() *[]interface{} {
	var returns *[]interface{}
	_jsii_.Get(
		j,
		"provisioners",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) RawOverrides() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"rawOverrides",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) SecurityPolicy() *string {
	var returns *string
	_jsii_.Get(
		j,
		"securityPolicy",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) SecurityPolicyInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"securityPolicyInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) SecuritySettings() ComputeBackendServiceSecuritySettingsOutputReference {
	var returns ComputeBackendServiceSecuritySettingsOutputReference
	_jsii_.Get(
		j,
		"securitySettings",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) SecuritySettingsInput() *ComputeBackendServiceSecuritySettings {
	var returns *ComputeBackendServiceSecuritySettings
	_jsii_.Get(
		j,
		"securitySettingsInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) SelfLink() *string {
	var returns *string
	_jsii_.Get(
		j,
		"selfLink",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) SessionAffinity() *string {
	var returns *string
	_jsii_.Get(
		j,
		"sessionAffinity",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) SessionAffinityInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"sessionAffinityInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) TerraformGeneratorMetadata() *cdktf.TerraformProviderGeneratorMetadata {
	var returns *cdktf.TerraformProviderGeneratorMetadata
	_jsii_.Get(
		j,
		"terraformGeneratorMetadata",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) TerraformMetaArguments() *map[string]interface{} {
	var returns *map[string]interface{}
	_jsii_.Get(
		j,
		"terraformMetaArguments",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) TerraformResourceType() *string {
	var returns *string
	_jsii_.Get(
		j,
		"terraformResourceType",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) Timeouts() ComputeBackendServiceTimeoutsOutputReference {
	var returns ComputeBackendServiceTimeoutsOutputReference
	_jsii_.Get(
		j,
		"timeouts",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) TimeoutSec() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"timeoutSec",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) TimeoutSecInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"timeoutSecInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendService) TimeoutsInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"timeoutsInput",
		&returns,
	)
	return returns
}


// Create a new {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_backend_service google_compute_backend_service} Resource.
func NewComputeBackendService(scope constructs.Construct, id *string, config *ComputeBackendServiceConfig) ComputeBackendService {
	_init_.Initialize()

	if err := validateNewComputeBackendServiceParameters(scope, id, config); err != nil {
		panic(err)
	}
	j := jsiiProxy_ComputeBackendService{}

	_jsii_.Create(
		"@cdktf/provider-google.computeBackendService.ComputeBackendService",
		[]interface{}{scope, id, config},
		&j,
	)

	return &j
}

// Create a new {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_backend_service google_compute_backend_service} Resource.
func NewComputeBackendService_Override(c ComputeBackendService, scope constructs.Construct, id *string, config *ComputeBackendServiceConfig) {
	_init_.Initialize()

	_jsii_.Create(
		"@cdktf/provider-google.computeBackendService.ComputeBackendService",
		[]interface{}{scope, id, config},
		c,
	)
}

func (j *jsiiProxy_ComputeBackendService)SetAffinityCookieTtlSec(val *float64) {
	if err := j.validateSetAffinityCookieTtlSecParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"affinityCookieTtlSec",
		val,
	)
}

func (j *jsiiProxy_ComputeBackendService)SetCompressionMode(val *string) {
	if err := j.validateSetCompressionModeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"compressionMode",
		val,
	)
}

func (j *jsiiProxy_ComputeBackendService)SetConnection(val interface{}) {
	if err := j.validateSetConnectionParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"connection",
		val,
	)
}

func (j *jsiiProxy_ComputeBackendService)SetConnectionDrainingTimeoutSec(val *float64) {
	if err := j.validateSetConnectionDrainingTimeoutSecParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"connectionDrainingTimeoutSec",
		val,
	)
}

func (j *jsiiProxy_ComputeBackendService)SetCount(val interface{}) {
	if err := j.validateSetCountParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"count",
		val,
	)
}

func (j *jsiiProxy_ComputeBackendService)SetCustomRequestHeaders(val *[]*string) {
	if err := j.validateSetCustomRequestHeadersParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"customRequestHeaders",
		val,
	)
}

func (j *jsiiProxy_ComputeBackendService)SetCustomResponseHeaders(val *[]*string) {
	if err := j.validateSetCustomResponseHeadersParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"customResponseHeaders",
		val,
	)
}

func (j *jsiiProxy_ComputeBackendService)SetDependsOn(val *[]*string) {
	_jsii_.Set(
		j,
		"dependsOn",
		val,
	)
}

func (j *jsiiProxy_ComputeBackendService)SetDescription(val *string) {
	if err := j.validateSetDescriptionParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"description",
		val,
	)
}

func (j *jsiiProxy_ComputeBackendService)SetEdgeSecurityPolicy(val *string) {
	if err := j.validateSetEdgeSecurityPolicyParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"edgeSecurityPolicy",
		val,
	)
}

func (j *jsiiProxy_ComputeBackendService)SetEnableCdn(val interface{}) {
	if err := j.validateSetEnableCdnParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"enableCdn",
		val,
	)
}

func (j *jsiiProxy_ComputeBackendService)SetForEach(val cdktf.ITerraformIterator) {
	_jsii_.Set(
		j,
		"forEach",
		val,
	)
}

func (j *jsiiProxy_ComputeBackendService)SetHealthChecks(val *[]*string) {
	if err := j.validateSetHealthChecksParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"healthChecks",
		val,
	)
}

func (j *jsiiProxy_ComputeBackendService)SetId(val *string) {
	if err := j.validateSetIdParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"id",
		val,
	)
}

func (j *jsiiProxy_ComputeBackendService)SetLifecycle(val *cdktf.TerraformResourceLifecycle) {
	if err := j.validateSetLifecycleParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"lifecycle",
		val,
	)
}

func (j *jsiiProxy_ComputeBackendService)SetLoadBalancingScheme(val *string) {
	if err := j.validateSetLoadBalancingSchemeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"loadBalancingScheme",
		val,
	)
}

func (j *jsiiProxy_ComputeBackendService)SetLocalityLbPolicy(val *string) {
	if err := j.validateSetLocalityLbPolicyParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"localityLbPolicy",
		val,
	)
}

func (j *jsiiProxy_ComputeBackendService)SetName(val *string) {
	if err := j.validateSetNameParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"name",
		val,
	)
}

func (j *jsiiProxy_ComputeBackendService)SetPortName(val *string) {
	if err := j.validateSetPortNameParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"portName",
		val,
	)
}

func (j *jsiiProxy_ComputeBackendService)SetProject(val *string) {
	if err := j.validateSetProjectParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"project",
		val,
	)
}

func (j *jsiiProxy_ComputeBackendService)SetProtocol(val *string) {
	if err := j.validateSetProtocolParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"protocol",
		val,
	)
}

func (j *jsiiProxy_ComputeBackendService)SetProvider(val cdktf.TerraformProvider) {
	_jsii_.Set(
		j,
		"provider",
		val,
	)
}

func (j *jsiiProxy_ComputeBackendService)SetProvisioners(val *[]interface{}) {
	if err := j.validateSetProvisionersParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"provisioners",
		val,
	)
}

func (j *jsiiProxy_ComputeBackendService)SetSecurityPolicy(val *string) {
	if err := j.validateSetSecurityPolicyParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"securityPolicy",
		val,
	)
}

func (j *jsiiProxy_ComputeBackendService)SetSessionAffinity(val *string) {
	if err := j.validateSetSessionAffinityParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"sessionAffinity",
		val,
	)
}

func (j *jsiiProxy_ComputeBackendService)SetTimeoutSec(val *float64) {
	if err := j.validateSetTimeoutSecParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"timeoutSec",
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
func ComputeBackendService_IsConstruct(x interface{}) *bool {
	_init_.Initialize()

	if err := validateComputeBackendService_IsConstructParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"@cdktf/provider-google.computeBackendService.ComputeBackendService",
		"isConstruct",
		[]interface{}{x},
		&returns,
	)

	return returns
}

// Experimental.
func ComputeBackendService_IsTerraformElement(x interface{}) *bool {
	_init_.Initialize()

	if err := validateComputeBackendService_IsTerraformElementParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"@cdktf/provider-google.computeBackendService.ComputeBackendService",
		"isTerraformElement",
		[]interface{}{x},
		&returns,
	)

	return returns
}

// Experimental.
func ComputeBackendService_IsTerraformResource(x interface{}) *bool {
	_init_.Initialize()

	if err := validateComputeBackendService_IsTerraformResourceParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"@cdktf/provider-google.computeBackendService.ComputeBackendService",
		"isTerraformResource",
		[]interface{}{x},
		&returns,
	)

	return returns
}

func ComputeBackendService_TfResourceType() *string {
	_init_.Initialize()
	var returns *string
	_jsii_.StaticGet(
		"@cdktf/provider-google.computeBackendService.ComputeBackendService",
		"tfResourceType",
		&returns,
	)
	return returns
}

func (c *jsiiProxy_ComputeBackendService) AddOverride(path *string, value interface{}) {
	if err := c.validateAddOverrideParameters(path, value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"addOverride",
		[]interface{}{path, value},
	)
}

func (c *jsiiProxy_ComputeBackendService) GetAnyMapAttribute(terraformAttribute *string) *map[string]interface{} {
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

func (c *jsiiProxy_ComputeBackendService) GetBooleanAttribute(terraformAttribute *string) cdktf.IResolvable {
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

func (c *jsiiProxy_ComputeBackendService) GetBooleanMapAttribute(terraformAttribute *string) *map[string]*bool {
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

func (c *jsiiProxy_ComputeBackendService) GetListAttribute(terraformAttribute *string) *[]*string {
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

func (c *jsiiProxy_ComputeBackendService) GetNumberAttribute(terraformAttribute *string) *float64 {
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

func (c *jsiiProxy_ComputeBackendService) GetNumberListAttribute(terraformAttribute *string) *[]*float64 {
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

func (c *jsiiProxy_ComputeBackendService) GetNumberMapAttribute(terraformAttribute *string) *map[string]*float64 {
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

func (c *jsiiProxy_ComputeBackendService) GetStringAttribute(terraformAttribute *string) *string {
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

func (c *jsiiProxy_ComputeBackendService) GetStringMapAttribute(terraformAttribute *string) *map[string]*string {
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

func (c *jsiiProxy_ComputeBackendService) InterpolationForAttribute(terraformAttribute *string) cdktf.IResolvable {
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

func (c *jsiiProxy_ComputeBackendService) OverrideLogicalId(newLogicalId *string) {
	if err := c.validateOverrideLogicalIdParameters(newLogicalId); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"overrideLogicalId",
		[]interface{}{newLogicalId},
	)
}

func (c *jsiiProxy_ComputeBackendService) PutBackend(value interface{}) {
	if err := c.validatePutBackendParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putBackend",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ComputeBackendService) PutCdnPolicy(value *ComputeBackendServiceCdnPolicy) {
	if err := c.validatePutCdnPolicyParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putCdnPolicy",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ComputeBackendService) PutCircuitBreakers(value *ComputeBackendServiceCircuitBreakers) {
	if err := c.validatePutCircuitBreakersParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putCircuitBreakers",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ComputeBackendService) PutConsistentHash(value *ComputeBackendServiceConsistentHash) {
	if err := c.validatePutConsistentHashParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putConsistentHash",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ComputeBackendService) PutIap(value *ComputeBackendServiceIap) {
	if err := c.validatePutIapParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putIap",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ComputeBackendService) PutLocalityLbPolicies(value interface{}) {
	if err := c.validatePutLocalityLbPoliciesParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putLocalityLbPolicies",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ComputeBackendService) PutLogConfig(value *ComputeBackendServiceLogConfig) {
	if err := c.validatePutLogConfigParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putLogConfig",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ComputeBackendService) PutOutlierDetection(value *ComputeBackendServiceOutlierDetection) {
	if err := c.validatePutOutlierDetectionParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putOutlierDetection",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ComputeBackendService) PutSecuritySettings(value *ComputeBackendServiceSecuritySettings) {
	if err := c.validatePutSecuritySettingsParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putSecuritySettings",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ComputeBackendService) PutTimeouts(value *ComputeBackendServiceTimeouts) {
	if err := c.validatePutTimeoutsParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putTimeouts",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ComputeBackendService) ResetAffinityCookieTtlSec() {
	_jsii_.InvokeVoid(
		c,
		"resetAffinityCookieTtlSec",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeBackendService) ResetBackend() {
	_jsii_.InvokeVoid(
		c,
		"resetBackend",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeBackendService) ResetCdnPolicy() {
	_jsii_.InvokeVoid(
		c,
		"resetCdnPolicy",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeBackendService) ResetCircuitBreakers() {
	_jsii_.InvokeVoid(
		c,
		"resetCircuitBreakers",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeBackendService) ResetCompressionMode() {
	_jsii_.InvokeVoid(
		c,
		"resetCompressionMode",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeBackendService) ResetConnectionDrainingTimeoutSec() {
	_jsii_.InvokeVoid(
		c,
		"resetConnectionDrainingTimeoutSec",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeBackendService) ResetConsistentHash() {
	_jsii_.InvokeVoid(
		c,
		"resetConsistentHash",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeBackendService) ResetCustomRequestHeaders() {
	_jsii_.InvokeVoid(
		c,
		"resetCustomRequestHeaders",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeBackendService) ResetCustomResponseHeaders() {
	_jsii_.InvokeVoid(
		c,
		"resetCustomResponseHeaders",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeBackendService) ResetDescription() {
	_jsii_.InvokeVoid(
		c,
		"resetDescription",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeBackendService) ResetEdgeSecurityPolicy() {
	_jsii_.InvokeVoid(
		c,
		"resetEdgeSecurityPolicy",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeBackendService) ResetEnableCdn() {
	_jsii_.InvokeVoid(
		c,
		"resetEnableCdn",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeBackendService) ResetHealthChecks() {
	_jsii_.InvokeVoid(
		c,
		"resetHealthChecks",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeBackendService) ResetIap() {
	_jsii_.InvokeVoid(
		c,
		"resetIap",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeBackendService) ResetId() {
	_jsii_.InvokeVoid(
		c,
		"resetId",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeBackendService) ResetLoadBalancingScheme() {
	_jsii_.InvokeVoid(
		c,
		"resetLoadBalancingScheme",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeBackendService) ResetLocalityLbPolicies() {
	_jsii_.InvokeVoid(
		c,
		"resetLocalityLbPolicies",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeBackendService) ResetLocalityLbPolicy() {
	_jsii_.InvokeVoid(
		c,
		"resetLocalityLbPolicy",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeBackendService) ResetLogConfig() {
	_jsii_.InvokeVoid(
		c,
		"resetLogConfig",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeBackendService) ResetOutlierDetection() {
	_jsii_.InvokeVoid(
		c,
		"resetOutlierDetection",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeBackendService) ResetOverrideLogicalId() {
	_jsii_.InvokeVoid(
		c,
		"resetOverrideLogicalId",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeBackendService) ResetPortName() {
	_jsii_.InvokeVoid(
		c,
		"resetPortName",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeBackendService) ResetProject() {
	_jsii_.InvokeVoid(
		c,
		"resetProject",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeBackendService) ResetProtocol() {
	_jsii_.InvokeVoid(
		c,
		"resetProtocol",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeBackendService) ResetSecurityPolicy() {
	_jsii_.InvokeVoid(
		c,
		"resetSecurityPolicy",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeBackendService) ResetSecuritySettings() {
	_jsii_.InvokeVoid(
		c,
		"resetSecuritySettings",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeBackendService) ResetSessionAffinity() {
	_jsii_.InvokeVoid(
		c,
		"resetSessionAffinity",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeBackendService) ResetTimeouts() {
	_jsii_.InvokeVoid(
		c,
		"resetTimeouts",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeBackendService) ResetTimeoutSec() {
	_jsii_.InvokeVoid(
		c,
		"resetTimeoutSec",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeBackendService) SynthesizeAttributes() *map[string]interface{} {
	var returns *map[string]interface{}

	_jsii_.Invoke(
		c,
		"synthesizeAttributes",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComputeBackendService) ToMetadata() interface{} {
	var returns interface{}

	_jsii_.Invoke(
		c,
		"toMetadata",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComputeBackendService) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		c,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComputeBackendService) ToTerraform() interface{} {
	var returns interface{}

	_jsii_.Invoke(
		c,
		"toTerraform",
		nil, // no parameters
		&returns,
	)

	return returns
}

