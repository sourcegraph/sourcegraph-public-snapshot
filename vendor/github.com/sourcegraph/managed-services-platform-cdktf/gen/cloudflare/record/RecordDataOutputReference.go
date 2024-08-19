package record

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/sourcegraph/managed-services-platform-cdktf/gen/cloudflare/jsii"

	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/cloudflare/record/internal"
)

type RecordDataOutputReference interface {
	cdktf.ComplexObject
	Algorithm() *float64
	SetAlgorithm(val *float64)
	AlgorithmInput() *float64
	Altitude() *float64
	SetAltitude(val *float64)
	AltitudeInput() *float64
	Certificate() *string
	SetCertificate(val *string)
	CertificateInput() *string
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
	Content() *string
	SetContent(val *string)
	ContentInput() *string
	// The creation stack of this resolvable which will be appended to errors thrown during resolution.
	//
	// If this returns an empty array the stack will not be attached.
	// Experimental.
	CreationStack() *[]*string
	Digest() *string
	SetDigest(val *string)
	DigestInput() *string
	DigestType() *float64
	SetDigestType(val *float64)
	DigestTypeInput() *float64
	Fingerprint() *string
	SetFingerprint(val *string)
	FingerprintInput() *string
	Flags() *string
	SetFlags(val *string)
	FlagsInput() *string
	// Experimental.
	Fqn() *string
	InternalValue() *RecordData
	SetInternalValue(val *RecordData)
	KeyTag() *float64
	SetKeyTag(val *float64)
	KeyTagInput() *float64
	LatDegrees() *float64
	SetLatDegrees(val *float64)
	LatDegreesInput() *float64
	LatDirection() *string
	SetLatDirection(val *string)
	LatDirectionInput() *string
	LatMinutes() *float64
	SetLatMinutes(val *float64)
	LatMinutesInput() *float64
	LatSeconds() *float64
	SetLatSeconds(val *float64)
	LatSecondsInput() *float64
	LongDegrees() *float64
	SetLongDegrees(val *float64)
	LongDegreesInput() *float64
	LongDirection() *string
	SetLongDirection(val *string)
	LongDirectionInput() *string
	LongMinutes() *float64
	SetLongMinutes(val *float64)
	LongMinutesInput() *float64
	LongSeconds() *float64
	SetLongSeconds(val *float64)
	LongSecondsInput() *float64
	MatchingType() *float64
	SetMatchingType(val *float64)
	MatchingTypeInput() *float64
	Name() *string
	SetName(val *string)
	NameInput() *string
	Order() *float64
	SetOrder(val *float64)
	OrderInput() *float64
	Port() *float64
	SetPort(val *float64)
	PortInput() *float64
	PrecisionHorz() *float64
	SetPrecisionHorz(val *float64)
	PrecisionHorzInput() *float64
	PrecisionVert() *float64
	SetPrecisionVert(val *float64)
	PrecisionVertInput() *float64
	Preference() *float64
	SetPreference(val *float64)
	PreferenceInput() *float64
	Priority() *float64
	SetPriority(val *float64)
	PriorityInput() *float64
	Proto() *string
	SetProto(val *string)
	Protocol() *float64
	SetProtocol(val *float64)
	ProtocolInput() *float64
	ProtoInput() *string
	PublicKey() *string
	SetPublicKey(val *string)
	PublicKeyInput() *string
	Regex() *string
	SetRegex(val *string)
	RegexInput() *string
	Replacement() *string
	SetReplacement(val *string)
	ReplacementInput() *string
	Selector() *float64
	SetSelector(val *float64)
	SelectorInput() *float64
	Service() *string
	SetService(val *string)
	ServiceInput() *string
	Size() *float64
	SetSize(val *float64)
	SizeInput() *float64
	Tag() *string
	SetTag(val *string)
	TagInput() *string
	Target() *string
	SetTarget(val *string)
	TargetInput() *string
	// Experimental.
	TerraformAttribute() *string
	// Experimental.
	SetTerraformAttribute(val *string)
	// Experimental.
	TerraformResource() cdktf.IInterpolatingParent
	// Experimental.
	SetTerraformResource(val cdktf.IInterpolatingParent)
	Type() *float64
	SetType(val *float64)
	TypeInput() *float64
	Usage() *float64
	SetUsage(val *float64)
	UsageInput() *float64
	Value() *string
	SetValue(val *string)
	ValueInput() *string
	Weight() *float64
	SetWeight(val *float64)
	WeightInput() *float64
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
	ResetAlgorithm()
	ResetAltitude()
	ResetCertificate()
	ResetContent()
	ResetDigest()
	ResetDigestType()
	ResetFingerprint()
	ResetFlags()
	ResetKeyTag()
	ResetLatDegrees()
	ResetLatDirection()
	ResetLatMinutes()
	ResetLatSeconds()
	ResetLongDegrees()
	ResetLongDirection()
	ResetLongMinutes()
	ResetLongSeconds()
	ResetMatchingType()
	ResetName()
	ResetOrder()
	ResetPort()
	ResetPrecisionHorz()
	ResetPrecisionVert()
	ResetPreference()
	ResetPriority()
	ResetProto()
	ResetProtocol()
	ResetPublicKey()
	ResetRegex()
	ResetReplacement()
	ResetSelector()
	ResetService()
	ResetSize()
	ResetTag()
	ResetTarget()
	ResetType()
	ResetUsage()
	ResetValue()
	ResetWeight()
	// Produce the Token's value at resolution time.
	// Experimental.
	Resolve(_context cdktf.IResolveContext) interface{}
	// Return a string representation of this resolvable object.
	//
	// Returns a reversible string representation.
	// Experimental.
	ToString() *string
}

// The jsii proxy struct for RecordDataOutputReference
type jsiiProxy_RecordDataOutputReference struct {
	internal.Type__cdktfComplexObject
}

func (j *jsiiProxy_RecordDataOutputReference) Algorithm() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"algorithm",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) AlgorithmInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"algorithmInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) Altitude() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"altitude",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) AltitudeInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"altitudeInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) Certificate() *string {
	var returns *string
	_jsii_.Get(
		j,
		"certificate",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) CertificateInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"certificateInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) ComplexObjectIndex() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"complexObjectIndex",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) ComplexObjectIsFromSet() *bool {
	var returns *bool
	_jsii_.Get(
		j,
		"complexObjectIsFromSet",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) Content() *string {
	var returns *string
	_jsii_.Get(
		j,
		"content",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) ContentInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"contentInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) CreationStack() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"creationStack",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) Digest() *string {
	var returns *string
	_jsii_.Get(
		j,
		"digest",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) DigestInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"digestInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) DigestType() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"digestType",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) DigestTypeInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"digestTypeInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) Fingerprint() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fingerprint",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) FingerprintInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fingerprintInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) Flags() *string {
	var returns *string
	_jsii_.Get(
		j,
		"flags",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) FlagsInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"flagsInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) Fqn() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fqn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) InternalValue() *RecordData {
	var returns *RecordData
	_jsii_.Get(
		j,
		"internalValue",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) KeyTag() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"keyTag",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) KeyTagInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"keyTagInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) LatDegrees() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"latDegrees",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) LatDegreesInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"latDegreesInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) LatDirection() *string {
	var returns *string
	_jsii_.Get(
		j,
		"latDirection",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) LatDirectionInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"latDirectionInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) LatMinutes() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"latMinutes",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) LatMinutesInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"latMinutesInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) LatSeconds() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"latSeconds",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) LatSecondsInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"latSecondsInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) LongDegrees() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"longDegrees",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) LongDegreesInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"longDegreesInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) LongDirection() *string {
	var returns *string
	_jsii_.Get(
		j,
		"longDirection",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) LongDirectionInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"longDirectionInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) LongMinutes() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"longMinutes",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) LongMinutesInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"longMinutesInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) LongSeconds() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"longSeconds",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) LongSecondsInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"longSecondsInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) MatchingType() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"matchingType",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) MatchingTypeInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"matchingTypeInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) Name() *string {
	var returns *string
	_jsii_.Get(
		j,
		"name",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) NameInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"nameInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) Order() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"order",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) OrderInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"orderInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) Port() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"port",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) PortInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"portInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) PrecisionHorz() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"precisionHorz",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) PrecisionHorzInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"precisionHorzInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) PrecisionVert() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"precisionVert",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) PrecisionVertInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"precisionVertInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) Preference() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"preference",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) PreferenceInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"preferenceInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) Priority() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"priority",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) PriorityInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"priorityInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) Proto() *string {
	var returns *string
	_jsii_.Get(
		j,
		"proto",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) Protocol() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"protocol",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) ProtocolInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"protocolInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) ProtoInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"protoInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) PublicKey() *string {
	var returns *string
	_jsii_.Get(
		j,
		"publicKey",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) PublicKeyInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"publicKeyInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) Regex() *string {
	var returns *string
	_jsii_.Get(
		j,
		"regex",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) RegexInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"regexInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) Replacement() *string {
	var returns *string
	_jsii_.Get(
		j,
		"replacement",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) ReplacementInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"replacementInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) Selector() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"selector",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) SelectorInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"selectorInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) Service() *string {
	var returns *string
	_jsii_.Get(
		j,
		"service",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) ServiceInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"serviceInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) Size() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"size",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) SizeInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"sizeInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) Tag() *string {
	var returns *string
	_jsii_.Get(
		j,
		"tag",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) TagInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"tagInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) Target() *string {
	var returns *string
	_jsii_.Get(
		j,
		"target",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) TargetInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"targetInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) TerraformAttribute() *string {
	var returns *string
	_jsii_.Get(
		j,
		"terraformAttribute",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) TerraformResource() cdktf.IInterpolatingParent {
	var returns cdktf.IInterpolatingParent
	_jsii_.Get(
		j,
		"terraformResource",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) Type() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"type",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) TypeInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"typeInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) Usage() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"usage",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) UsageInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"usageInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) Value() *string {
	var returns *string
	_jsii_.Get(
		j,
		"value",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) ValueInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"valueInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) Weight() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"weight",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_RecordDataOutputReference) WeightInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"weightInput",
		&returns,
	)
	return returns
}


func NewRecordDataOutputReference(terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) RecordDataOutputReference {
	_init_.Initialize()

	if err := validateNewRecordDataOutputReferenceParameters(terraformResource, terraformAttribute); err != nil {
		panic(err)
	}
	j := jsiiProxy_RecordDataOutputReference{}

	_jsii_.Create(
		"@cdktf/provider-cloudflare.record.RecordDataOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		&j,
	)

	return &j
}

func NewRecordDataOutputReference_Override(r RecordDataOutputReference, terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) {
	_init_.Initialize()

	_jsii_.Create(
		"@cdktf/provider-cloudflare.record.RecordDataOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		r,
	)
}

func (j *jsiiProxy_RecordDataOutputReference)SetAlgorithm(val *float64) {
	if err := j.validateSetAlgorithmParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"algorithm",
		val,
	)
}

func (j *jsiiProxy_RecordDataOutputReference)SetAltitude(val *float64) {
	if err := j.validateSetAltitudeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"altitude",
		val,
	)
}

func (j *jsiiProxy_RecordDataOutputReference)SetCertificate(val *string) {
	if err := j.validateSetCertificateParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"certificate",
		val,
	)
}

func (j *jsiiProxy_RecordDataOutputReference)SetComplexObjectIndex(val interface{}) {
	if err := j.validateSetComplexObjectIndexParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIndex",
		val,
	)
}

func (j *jsiiProxy_RecordDataOutputReference)SetComplexObjectIsFromSet(val *bool) {
	if err := j.validateSetComplexObjectIsFromSetParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIsFromSet",
		val,
	)
}

func (j *jsiiProxy_RecordDataOutputReference)SetContent(val *string) {
	if err := j.validateSetContentParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"content",
		val,
	)
}

func (j *jsiiProxy_RecordDataOutputReference)SetDigest(val *string) {
	if err := j.validateSetDigestParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"digest",
		val,
	)
}

func (j *jsiiProxy_RecordDataOutputReference)SetDigestType(val *float64) {
	if err := j.validateSetDigestTypeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"digestType",
		val,
	)
}

func (j *jsiiProxy_RecordDataOutputReference)SetFingerprint(val *string) {
	if err := j.validateSetFingerprintParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"fingerprint",
		val,
	)
}

func (j *jsiiProxy_RecordDataOutputReference)SetFlags(val *string) {
	if err := j.validateSetFlagsParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"flags",
		val,
	)
}

func (j *jsiiProxy_RecordDataOutputReference)SetInternalValue(val *RecordData) {
	if err := j.validateSetInternalValueParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"internalValue",
		val,
	)
}

func (j *jsiiProxy_RecordDataOutputReference)SetKeyTag(val *float64) {
	if err := j.validateSetKeyTagParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"keyTag",
		val,
	)
}

func (j *jsiiProxy_RecordDataOutputReference)SetLatDegrees(val *float64) {
	if err := j.validateSetLatDegreesParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"latDegrees",
		val,
	)
}

func (j *jsiiProxy_RecordDataOutputReference)SetLatDirection(val *string) {
	if err := j.validateSetLatDirectionParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"latDirection",
		val,
	)
}

func (j *jsiiProxy_RecordDataOutputReference)SetLatMinutes(val *float64) {
	if err := j.validateSetLatMinutesParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"latMinutes",
		val,
	)
}

func (j *jsiiProxy_RecordDataOutputReference)SetLatSeconds(val *float64) {
	if err := j.validateSetLatSecondsParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"latSeconds",
		val,
	)
}

func (j *jsiiProxy_RecordDataOutputReference)SetLongDegrees(val *float64) {
	if err := j.validateSetLongDegreesParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"longDegrees",
		val,
	)
}

func (j *jsiiProxy_RecordDataOutputReference)SetLongDirection(val *string) {
	if err := j.validateSetLongDirectionParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"longDirection",
		val,
	)
}

func (j *jsiiProxy_RecordDataOutputReference)SetLongMinutes(val *float64) {
	if err := j.validateSetLongMinutesParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"longMinutes",
		val,
	)
}

func (j *jsiiProxy_RecordDataOutputReference)SetLongSeconds(val *float64) {
	if err := j.validateSetLongSecondsParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"longSeconds",
		val,
	)
}

func (j *jsiiProxy_RecordDataOutputReference)SetMatchingType(val *float64) {
	if err := j.validateSetMatchingTypeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"matchingType",
		val,
	)
}

func (j *jsiiProxy_RecordDataOutputReference)SetName(val *string) {
	if err := j.validateSetNameParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"name",
		val,
	)
}

func (j *jsiiProxy_RecordDataOutputReference)SetOrder(val *float64) {
	if err := j.validateSetOrderParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"order",
		val,
	)
}

func (j *jsiiProxy_RecordDataOutputReference)SetPort(val *float64) {
	if err := j.validateSetPortParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"port",
		val,
	)
}

func (j *jsiiProxy_RecordDataOutputReference)SetPrecisionHorz(val *float64) {
	if err := j.validateSetPrecisionHorzParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"precisionHorz",
		val,
	)
}

func (j *jsiiProxy_RecordDataOutputReference)SetPrecisionVert(val *float64) {
	if err := j.validateSetPrecisionVertParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"precisionVert",
		val,
	)
}

func (j *jsiiProxy_RecordDataOutputReference)SetPreference(val *float64) {
	if err := j.validateSetPreferenceParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"preference",
		val,
	)
}

func (j *jsiiProxy_RecordDataOutputReference)SetPriority(val *float64) {
	if err := j.validateSetPriorityParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"priority",
		val,
	)
}

func (j *jsiiProxy_RecordDataOutputReference)SetProto(val *string) {
	if err := j.validateSetProtoParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"proto",
		val,
	)
}

func (j *jsiiProxy_RecordDataOutputReference)SetProtocol(val *float64) {
	if err := j.validateSetProtocolParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"protocol",
		val,
	)
}

func (j *jsiiProxy_RecordDataOutputReference)SetPublicKey(val *string) {
	if err := j.validateSetPublicKeyParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"publicKey",
		val,
	)
}

func (j *jsiiProxy_RecordDataOutputReference)SetRegex(val *string) {
	if err := j.validateSetRegexParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"regex",
		val,
	)
}

func (j *jsiiProxy_RecordDataOutputReference)SetReplacement(val *string) {
	if err := j.validateSetReplacementParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"replacement",
		val,
	)
}

func (j *jsiiProxy_RecordDataOutputReference)SetSelector(val *float64) {
	if err := j.validateSetSelectorParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"selector",
		val,
	)
}

func (j *jsiiProxy_RecordDataOutputReference)SetService(val *string) {
	if err := j.validateSetServiceParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"service",
		val,
	)
}

func (j *jsiiProxy_RecordDataOutputReference)SetSize(val *float64) {
	if err := j.validateSetSizeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"size",
		val,
	)
}

func (j *jsiiProxy_RecordDataOutputReference)SetTag(val *string) {
	if err := j.validateSetTagParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"tag",
		val,
	)
}

func (j *jsiiProxy_RecordDataOutputReference)SetTarget(val *string) {
	if err := j.validateSetTargetParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"target",
		val,
	)
}

func (j *jsiiProxy_RecordDataOutputReference)SetTerraformAttribute(val *string) {
	if err := j.validateSetTerraformAttributeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformAttribute",
		val,
	)
}

func (j *jsiiProxy_RecordDataOutputReference)SetTerraformResource(val cdktf.IInterpolatingParent) {
	if err := j.validateSetTerraformResourceParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformResource",
		val,
	)
}

func (j *jsiiProxy_RecordDataOutputReference)SetType(val *float64) {
	if err := j.validateSetTypeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"type",
		val,
	)
}

func (j *jsiiProxy_RecordDataOutputReference)SetUsage(val *float64) {
	if err := j.validateSetUsageParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"usage",
		val,
	)
}

func (j *jsiiProxy_RecordDataOutputReference)SetValue(val *string) {
	if err := j.validateSetValueParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"value",
		val,
	)
}

func (j *jsiiProxy_RecordDataOutputReference)SetWeight(val *float64) {
	if err := j.validateSetWeightParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"weight",
		val,
	)
}

func (r *jsiiProxy_RecordDataOutputReference) ComputeFqn() *string {
	var returns *string

	_jsii_.Invoke(
		r,
		"computeFqn",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (r *jsiiProxy_RecordDataOutputReference) GetAnyMapAttribute(terraformAttribute *string) *map[string]interface{} {
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

func (r *jsiiProxy_RecordDataOutputReference) GetBooleanAttribute(terraformAttribute *string) cdktf.IResolvable {
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

func (r *jsiiProxy_RecordDataOutputReference) GetBooleanMapAttribute(terraformAttribute *string) *map[string]*bool {
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

func (r *jsiiProxy_RecordDataOutputReference) GetListAttribute(terraformAttribute *string) *[]*string {
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

func (r *jsiiProxy_RecordDataOutputReference) GetNumberAttribute(terraformAttribute *string) *float64 {
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

func (r *jsiiProxy_RecordDataOutputReference) GetNumberListAttribute(terraformAttribute *string) *[]*float64 {
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

func (r *jsiiProxy_RecordDataOutputReference) GetNumberMapAttribute(terraformAttribute *string) *map[string]*float64 {
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

func (r *jsiiProxy_RecordDataOutputReference) GetStringAttribute(terraformAttribute *string) *string {
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

func (r *jsiiProxy_RecordDataOutputReference) GetStringMapAttribute(terraformAttribute *string) *map[string]*string {
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

func (r *jsiiProxy_RecordDataOutputReference) InterpolationAsList() cdktf.IResolvable {
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		r,
		"interpolationAsList",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (r *jsiiProxy_RecordDataOutputReference) InterpolationForAttribute(property *string) cdktf.IResolvable {
	if err := r.validateInterpolationForAttributeParameters(property); err != nil {
		panic(err)
	}
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		r,
		"interpolationForAttribute",
		[]interface{}{property},
		&returns,
	)

	return returns
}

func (r *jsiiProxy_RecordDataOutputReference) ResetAlgorithm() {
	_jsii_.InvokeVoid(
		r,
		"resetAlgorithm",
		nil, // no parameters
	)
}

func (r *jsiiProxy_RecordDataOutputReference) ResetAltitude() {
	_jsii_.InvokeVoid(
		r,
		"resetAltitude",
		nil, // no parameters
	)
}

func (r *jsiiProxy_RecordDataOutputReference) ResetCertificate() {
	_jsii_.InvokeVoid(
		r,
		"resetCertificate",
		nil, // no parameters
	)
}

func (r *jsiiProxy_RecordDataOutputReference) ResetContent() {
	_jsii_.InvokeVoid(
		r,
		"resetContent",
		nil, // no parameters
	)
}

func (r *jsiiProxy_RecordDataOutputReference) ResetDigest() {
	_jsii_.InvokeVoid(
		r,
		"resetDigest",
		nil, // no parameters
	)
}

func (r *jsiiProxy_RecordDataOutputReference) ResetDigestType() {
	_jsii_.InvokeVoid(
		r,
		"resetDigestType",
		nil, // no parameters
	)
}

func (r *jsiiProxy_RecordDataOutputReference) ResetFingerprint() {
	_jsii_.InvokeVoid(
		r,
		"resetFingerprint",
		nil, // no parameters
	)
}

func (r *jsiiProxy_RecordDataOutputReference) ResetFlags() {
	_jsii_.InvokeVoid(
		r,
		"resetFlags",
		nil, // no parameters
	)
}

func (r *jsiiProxy_RecordDataOutputReference) ResetKeyTag() {
	_jsii_.InvokeVoid(
		r,
		"resetKeyTag",
		nil, // no parameters
	)
}

func (r *jsiiProxy_RecordDataOutputReference) ResetLatDegrees() {
	_jsii_.InvokeVoid(
		r,
		"resetLatDegrees",
		nil, // no parameters
	)
}

func (r *jsiiProxy_RecordDataOutputReference) ResetLatDirection() {
	_jsii_.InvokeVoid(
		r,
		"resetLatDirection",
		nil, // no parameters
	)
}

func (r *jsiiProxy_RecordDataOutputReference) ResetLatMinutes() {
	_jsii_.InvokeVoid(
		r,
		"resetLatMinutes",
		nil, // no parameters
	)
}

func (r *jsiiProxy_RecordDataOutputReference) ResetLatSeconds() {
	_jsii_.InvokeVoid(
		r,
		"resetLatSeconds",
		nil, // no parameters
	)
}

func (r *jsiiProxy_RecordDataOutputReference) ResetLongDegrees() {
	_jsii_.InvokeVoid(
		r,
		"resetLongDegrees",
		nil, // no parameters
	)
}

func (r *jsiiProxy_RecordDataOutputReference) ResetLongDirection() {
	_jsii_.InvokeVoid(
		r,
		"resetLongDirection",
		nil, // no parameters
	)
}

func (r *jsiiProxy_RecordDataOutputReference) ResetLongMinutes() {
	_jsii_.InvokeVoid(
		r,
		"resetLongMinutes",
		nil, // no parameters
	)
}

func (r *jsiiProxy_RecordDataOutputReference) ResetLongSeconds() {
	_jsii_.InvokeVoid(
		r,
		"resetLongSeconds",
		nil, // no parameters
	)
}

func (r *jsiiProxy_RecordDataOutputReference) ResetMatchingType() {
	_jsii_.InvokeVoid(
		r,
		"resetMatchingType",
		nil, // no parameters
	)
}

func (r *jsiiProxy_RecordDataOutputReference) ResetName() {
	_jsii_.InvokeVoid(
		r,
		"resetName",
		nil, // no parameters
	)
}

func (r *jsiiProxy_RecordDataOutputReference) ResetOrder() {
	_jsii_.InvokeVoid(
		r,
		"resetOrder",
		nil, // no parameters
	)
}

func (r *jsiiProxy_RecordDataOutputReference) ResetPort() {
	_jsii_.InvokeVoid(
		r,
		"resetPort",
		nil, // no parameters
	)
}

func (r *jsiiProxy_RecordDataOutputReference) ResetPrecisionHorz() {
	_jsii_.InvokeVoid(
		r,
		"resetPrecisionHorz",
		nil, // no parameters
	)
}

func (r *jsiiProxy_RecordDataOutputReference) ResetPrecisionVert() {
	_jsii_.InvokeVoid(
		r,
		"resetPrecisionVert",
		nil, // no parameters
	)
}

func (r *jsiiProxy_RecordDataOutputReference) ResetPreference() {
	_jsii_.InvokeVoid(
		r,
		"resetPreference",
		nil, // no parameters
	)
}

func (r *jsiiProxy_RecordDataOutputReference) ResetPriority() {
	_jsii_.InvokeVoid(
		r,
		"resetPriority",
		nil, // no parameters
	)
}

func (r *jsiiProxy_RecordDataOutputReference) ResetProto() {
	_jsii_.InvokeVoid(
		r,
		"resetProto",
		nil, // no parameters
	)
}

func (r *jsiiProxy_RecordDataOutputReference) ResetProtocol() {
	_jsii_.InvokeVoid(
		r,
		"resetProtocol",
		nil, // no parameters
	)
}

func (r *jsiiProxy_RecordDataOutputReference) ResetPublicKey() {
	_jsii_.InvokeVoid(
		r,
		"resetPublicKey",
		nil, // no parameters
	)
}

func (r *jsiiProxy_RecordDataOutputReference) ResetRegex() {
	_jsii_.InvokeVoid(
		r,
		"resetRegex",
		nil, // no parameters
	)
}

func (r *jsiiProxy_RecordDataOutputReference) ResetReplacement() {
	_jsii_.InvokeVoid(
		r,
		"resetReplacement",
		nil, // no parameters
	)
}

func (r *jsiiProxy_RecordDataOutputReference) ResetSelector() {
	_jsii_.InvokeVoid(
		r,
		"resetSelector",
		nil, // no parameters
	)
}

func (r *jsiiProxy_RecordDataOutputReference) ResetService() {
	_jsii_.InvokeVoid(
		r,
		"resetService",
		nil, // no parameters
	)
}

func (r *jsiiProxy_RecordDataOutputReference) ResetSize() {
	_jsii_.InvokeVoid(
		r,
		"resetSize",
		nil, // no parameters
	)
}

func (r *jsiiProxy_RecordDataOutputReference) ResetTag() {
	_jsii_.InvokeVoid(
		r,
		"resetTag",
		nil, // no parameters
	)
}

func (r *jsiiProxy_RecordDataOutputReference) ResetTarget() {
	_jsii_.InvokeVoid(
		r,
		"resetTarget",
		nil, // no parameters
	)
}

func (r *jsiiProxy_RecordDataOutputReference) ResetType() {
	_jsii_.InvokeVoid(
		r,
		"resetType",
		nil, // no parameters
	)
}

func (r *jsiiProxy_RecordDataOutputReference) ResetUsage() {
	_jsii_.InvokeVoid(
		r,
		"resetUsage",
		nil, // no parameters
	)
}

func (r *jsiiProxy_RecordDataOutputReference) ResetValue() {
	_jsii_.InvokeVoid(
		r,
		"resetValue",
		nil, // no parameters
	)
}

func (r *jsiiProxy_RecordDataOutputReference) ResetWeight() {
	_jsii_.InvokeVoid(
		r,
		"resetWeight",
		nil, // no parameters
	)
}

func (r *jsiiProxy_RecordDataOutputReference) Resolve(_context cdktf.IResolveContext) interface{} {
	if err := r.validateResolveParameters(_context); err != nil {
		panic(err)
	}
	var returns interface{}

	_jsii_.Invoke(
		r,
		"resolve",
		[]interface{}{_context},
		&returns,
	)

	return returns
}

func (r *jsiiProxy_RecordDataOutputReference) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		r,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

