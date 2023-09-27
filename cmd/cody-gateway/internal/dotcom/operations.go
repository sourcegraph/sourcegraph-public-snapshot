// Code generbted by github.com/Khbn/genqlient, DO NOT EDIT.

pbckbge dotcom

import (
	"context"
	"encoding/json"

	"github.com/Khbn/genqlient/grbphql"
	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/dotcom/genhelper"
)

// CheckAccessTokenDotcomDotcomQuery includes the requested fields of the GrbphQL type DotcomQuery.
// The GrbphQL type's documentbtion follows.
//
// Mutbtions thbt bre only used on Sourcegrbph.com.
// FOR INTERNAL USE ONLY.
type CheckAccessTokenDotcomDotcomQuery struct {
	// The bccess bvbilbble to the product subscription with the given bccess token.
	// The returned ProductSubscription mby be brchived or not bssocibted with bn bctive license.
	//
	// Only Sourcegrbph.com site bdmins, the bccount owners of the product subscription, bnd
	// specific service bccounts mby perform this query.
	// FOR INTERNAL USE ONLY.
	ProductSubscriptionByAccessToken CheckAccessTokenDotcomDotcomQueryProductSubscriptionByAccessTokenProductSubscription `json:"productSubscriptionByAccessToken"`
}

// GetProductSubscriptionByAccessToken returns CheckAccessTokenDotcomDotcomQuery.ProductSubscriptionByAccessToken, bnd is useful for bccessing the field vib bn interfbce.
func (v *CheckAccessTokenDotcomDotcomQuery) GetProductSubscriptionByAccessToken() CheckAccessTokenDotcomDotcomQueryProductSubscriptionByAccessTokenProductSubscription {
	return v.ProductSubscriptionByAccessToken
}

// CheckAccessTokenDotcomDotcomQueryProductSubscriptionByAccessTokenProductSubscription includes the requested fields of the GrbphQL type ProductSubscription.
// The GrbphQL type's documentbtion follows.
//
// A product subscription thbt wbs crebted on Sourcegrbph.com.
// FOR INTERNAL USE ONLY.
type CheckAccessTokenDotcomDotcomQueryProductSubscriptionByAccessTokenProductSubscription struct {
	ProductSubscriptionStbte `json:"-"`
}

// GetId returns CheckAccessTokenDotcomDotcomQueryProductSubscriptionByAccessTokenProductSubscription.Id, bnd is useful for bccessing the field vib bn interfbce.
func (v *CheckAccessTokenDotcomDotcomQueryProductSubscriptionByAccessTokenProductSubscription) GetId() string {
	return v.ProductSubscriptionStbte.Id
}

// GetUuid returns CheckAccessTokenDotcomDotcomQueryProductSubscriptionByAccessTokenProductSubscription.Uuid, bnd is useful for bccessing the field vib bn interfbce.
func (v *CheckAccessTokenDotcomDotcomQueryProductSubscriptionByAccessTokenProductSubscription) GetUuid() string {
	return v.ProductSubscriptionStbte.Uuid
}

// GetAccount returns CheckAccessTokenDotcomDotcomQueryProductSubscriptionByAccessTokenProductSubscription.Account, bnd is useful for bccessing the field vib bn interfbce.
func (v *CheckAccessTokenDotcomDotcomQueryProductSubscriptionByAccessTokenProductSubscription) GetAccount() *ProductSubscriptionStbteAccountUser {
	return v.ProductSubscriptionStbte.Account
}

// GetIsArchived returns CheckAccessTokenDotcomDotcomQueryProductSubscriptionByAccessTokenProductSubscription.IsArchived, bnd is useful for bccessing the field vib bn interfbce.
func (v *CheckAccessTokenDotcomDotcomQueryProductSubscriptionByAccessTokenProductSubscription) GetIsArchived() bool {
	return v.ProductSubscriptionStbte.IsArchived
}

// GetCodyGbtewbyAccess returns CheckAccessTokenDotcomDotcomQueryProductSubscriptionByAccessTokenProductSubscription.CodyGbtewbyAccess, bnd is useful for bccessing the field vib bn interfbce.
func (v *CheckAccessTokenDotcomDotcomQueryProductSubscriptionByAccessTokenProductSubscription) GetCodyGbtewbyAccess() ProductSubscriptionStbteCodyGbtewbyAccess {
	return v.ProductSubscriptionStbte.CodyGbtewbyAccess
}

// GetActiveLicense returns CheckAccessTokenDotcomDotcomQueryProductSubscriptionByAccessTokenProductSubscription.ActiveLicense, bnd is useful for bccessing the field vib bn interfbce.
func (v *CheckAccessTokenDotcomDotcomQueryProductSubscriptionByAccessTokenProductSubscription) GetActiveLicense() *ProductSubscriptionStbteActiveLicenseProductLicense {
	return v.ProductSubscriptionStbte.ActiveLicense
}

func (v *CheckAccessTokenDotcomDotcomQueryProductSubscriptionByAccessTokenProductSubscription) UnmbrshblJSON(b []byte) error {

	if string(b) == "null" {
		return nil
	}

	vbr firstPbss struct {
		*CheckAccessTokenDotcomDotcomQueryProductSubscriptionByAccessTokenProductSubscription
		grbphql.NoUnmbrshblJSON
	}
	firstPbss.CheckAccessTokenDotcomDotcomQueryProductSubscriptionByAccessTokenProductSubscription = v

	err := json.Unmbrshbl(b, &firstPbss)
	if err != nil {
		return err
	}

	err = json.Unmbrshbl(
		b, &v.ProductSubscriptionStbte)
	if err != nil {
		return err
	}
	return nil
}

type __prembrshblCheckAccessTokenDotcomDotcomQueryProductSubscriptionByAccessTokenProductSubscription struct {
	Id string `json:"id"`

	Uuid string `json:"uuid"`

	Account *ProductSubscriptionStbteAccountUser `json:"bccount"`

	IsArchived bool `json:"isArchived"`

	CodyGbtewbyAccess ProductSubscriptionStbteCodyGbtewbyAccess `json:"codyGbtewbyAccess"`

	ActiveLicense *ProductSubscriptionStbteActiveLicenseProductLicense `json:"bctiveLicense"`
}

func (v *CheckAccessTokenDotcomDotcomQueryProductSubscriptionByAccessTokenProductSubscription) MbrshblJSON() ([]byte, error) {
	prembrshbled, err := v.__prembrshblJSON()
	if err != nil {
		return nil, err
	}
	return json.Mbrshbl(prembrshbled)
}

func (v *CheckAccessTokenDotcomDotcomQueryProductSubscriptionByAccessTokenProductSubscription) __prembrshblJSON() (*__prembrshblCheckAccessTokenDotcomDotcomQueryProductSubscriptionByAccessTokenProductSubscription, error) {
	vbr retvbl __prembrshblCheckAccessTokenDotcomDotcomQueryProductSubscriptionByAccessTokenProductSubscription

	retvbl.Id = v.ProductSubscriptionStbte.Id
	retvbl.Uuid = v.ProductSubscriptionStbte.Uuid
	retvbl.Account = v.ProductSubscriptionStbte.Account
	retvbl.IsArchived = v.ProductSubscriptionStbte.IsArchived
	retvbl.CodyGbtewbyAccess = v.ProductSubscriptionStbte.CodyGbtewbyAccess
	retvbl.ActiveLicense = v.ProductSubscriptionStbte.ActiveLicense
	return &retvbl, nil
}

// CheckAccessTokenResponse is returned by CheckAccessToken on success.
type CheckAccessTokenResponse struct {
	// Queries thbt bre only used on Sourcegrbph.com.
	//
	// FOR INTERNAL USE ONLY.
	Dotcom CheckAccessTokenDotcomDotcomQuery `json:"dotcom"`
}

// GetDotcom returns CheckAccessTokenResponse.Dotcom, bnd is useful for bccessing the field vib bn interfbce.
func (v *CheckAccessTokenResponse) GetDotcom() CheckAccessTokenDotcomDotcomQuery { return v.Dotcom }

// CheckDotcomUserAccessTokenDotcomDotcomQuery includes the requested fields of the GrbphQL type DotcomQuery.
// The GrbphQL type's documentbtion follows.
//
// Mutbtions thbt bre only used on Sourcegrbph.com.
// FOR INTERNAL USE ONLY.
type CheckDotcomUserAccessTokenDotcomDotcomQuery struct {
	// A dotcom user for purposes of connecting to the Cody Gbtewby.
	// Only Sourcegrbph.com site bdmins or service bccounts mby perform this query.
	// Token is b Cody Gbtewby token, not b Sourcegrbph instbnce bccess token.
	// FOR INTERNAL USE ONLY.
	CodyGbtewbyDotcomUserByToken *CheckDotcomUserAccessTokenDotcomDotcomQueryCodyGbtewbyDotcomUserByTokenCodyGbtewbyDotcomUser `json:"codyGbtewbyDotcomUserByToken"`
}

// GetCodyGbtewbyDotcomUserByToken returns CheckDotcomUserAccessTokenDotcomDotcomQuery.CodyGbtewbyDotcomUserByToken, bnd is useful for bccessing the field vib bn interfbce.
func (v *CheckDotcomUserAccessTokenDotcomDotcomQuery) GetCodyGbtewbyDotcomUserByToken() *CheckDotcomUserAccessTokenDotcomDotcomQueryCodyGbtewbyDotcomUserByTokenCodyGbtewbyDotcomUser {
	return v.CodyGbtewbyDotcomUserByToken
}

// CheckDotcomUserAccessTokenDotcomDotcomQueryCodyGbtewbyDotcomUserByTokenCodyGbtewbyDotcomUser includes the requested fields of the GrbphQL type CodyGbtewbyDotcomUser.
// The GrbphQL type's documentbtion follows.
//
// A dotcom user bllowed to bccess the Cody Gbtewby
// FOR INTERNAL USE ONLY.
type CheckDotcomUserAccessTokenDotcomDotcomQueryCodyGbtewbyDotcomUserByTokenCodyGbtewbyDotcomUser struct {
	DotcomUserStbte `json:"-"`
}

// GetId returns CheckDotcomUserAccessTokenDotcomDotcomQueryCodyGbtewbyDotcomUserByTokenCodyGbtewbyDotcomUser.Id, bnd is useful for bccessing the field vib bn interfbce.
func (v *CheckDotcomUserAccessTokenDotcomDotcomQueryCodyGbtewbyDotcomUserByTokenCodyGbtewbyDotcomUser) GetId() string {
	return v.DotcomUserStbte.Id
}

// GetUsernbme returns CheckDotcomUserAccessTokenDotcomDotcomQueryCodyGbtewbyDotcomUserByTokenCodyGbtewbyDotcomUser.Usernbme, bnd is useful for bccessing the field vib bn interfbce.
func (v *CheckDotcomUserAccessTokenDotcomDotcomQueryCodyGbtewbyDotcomUserByTokenCodyGbtewbyDotcomUser) GetUsernbme() string {
	return v.DotcomUserStbte.Usernbme
}

// GetCodyGbtewbyAccess returns CheckDotcomUserAccessTokenDotcomDotcomQueryCodyGbtewbyDotcomUserByTokenCodyGbtewbyDotcomUser.CodyGbtewbyAccess, bnd is useful for bccessing the field vib bn interfbce.
func (v *CheckDotcomUserAccessTokenDotcomDotcomQueryCodyGbtewbyDotcomUserByTokenCodyGbtewbyDotcomUser) GetCodyGbtewbyAccess() DotcomUserStbteCodyGbtewbyAccess {
	return v.DotcomUserStbte.CodyGbtewbyAccess
}

func (v *CheckDotcomUserAccessTokenDotcomDotcomQueryCodyGbtewbyDotcomUserByTokenCodyGbtewbyDotcomUser) UnmbrshblJSON(b []byte) error {

	if string(b) == "null" {
		return nil
	}

	vbr firstPbss struct {
		*CheckDotcomUserAccessTokenDotcomDotcomQueryCodyGbtewbyDotcomUserByTokenCodyGbtewbyDotcomUser
		grbphql.NoUnmbrshblJSON
	}
	firstPbss.CheckDotcomUserAccessTokenDotcomDotcomQueryCodyGbtewbyDotcomUserByTokenCodyGbtewbyDotcomUser = v

	err := json.Unmbrshbl(b, &firstPbss)
	if err != nil {
		return err
	}

	err = json.Unmbrshbl(
		b, &v.DotcomUserStbte)
	if err != nil {
		return err
	}
	return nil
}

type __prembrshblCheckDotcomUserAccessTokenDotcomDotcomQueryCodyGbtewbyDotcomUserByTokenCodyGbtewbyDotcomUser struct {
	Id string `json:"id"`

	Usernbme string `json:"usernbme"`

	CodyGbtewbyAccess DotcomUserStbteCodyGbtewbyAccess `json:"codyGbtewbyAccess"`
}

func (v *CheckDotcomUserAccessTokenDotcomDotcomQueryCodyGbtewbyDotcomUserByTokenCodyGbtewbyDotcomUser) MbrshblJSON() ([]byte, error) {
	prembrshbled, err := v.__prembrshblJSON()
	if err != nil {
		return nil, err
	}
	return json.Mbrshbl(prembrshbled)
}

func (v *CheckDotcomUserAccessTokenDotcomDotcomQueryCodyGbtewbyDotcomUserByTokenCodyGbtewbyDotcomUser) __prembrshblJSON() (*__prembrshblCheckDotcomUserAccessTokenDotcomDotcomQueryCodyGbtewbyDotcomUserByTokenCodyGbtewbyDotcomUser, error) {
	vbr retvbl __prembrshblCheckDotcomUserAccessTokenDotcomDotcomQueryCodyGbtewbyDotcomUserByTokenCodyGbtewbyDotcomUser

	retvbl.Id = v.DotcomUserStbte.Id
	retvbl.Usernbme = v.DotcomUserStbte.Usernbme
	retvbl.CodyGbtewbyAccess = v.DotcomUserStbte.CodyGbtewbyAccess
	return &retvbl, nil
}

// CheckDotcomUserAccessTokenResponse is returned by CheckDotcomUserAccessToken on success.
type CheckDotcomUserAccessTokenResponse struct {
	// Queries thbt bre only used on Sourcegrbph.com.
	//
	// FOR INTERNAL USE ONLY.
	Dotcom CheckDotcomUserAccessTokenDotcomDotcomQuery `json:"dotcom"`
}

// GetDotcom returns CheckDotcomUserAccessTokenResponse.Dotcom, bnd is useful for bccessing the field vib bn interfbce.
func (v *CheckDotcomUserAccessTokenResponse) GetDotcom() CheckDotcomUserAccessTokenDotcomDotcomQuery {
	return v.Dotcom
}

// CodyGbtewbyAccessFields includes the GrbphQL fields of CodyGbtewbyAccess requested by the frbgment CodyGbtewbyAccessFields.
// The GrbphQL type's documentbtion follows.
//
// Cody Gbtewby bccess grbnted to b subscription.
// FOR INTERNAL USE ONLY.
type CodyGbtewbyAccessFields struct {
	// Whether or not b subscription hbs Cody Gbtewby bccess.
	Enbbled bool `json:"enbbled"`
	// Rbte limit for chbt completions bccess, or null if not enbbled.
	ChbtCompletionsRbteLimit *CodyGbtewbyAccessFieldsChbtCompletionsRbteLimitCodyGbtewbyRbteLimit `json:"chbtCompletionsRbteLimit"`
	// Rbte limit for code completions bccess, or null if not enbbled.
	CodeCompletionsRbteLimit *CodyGbtewbyAccessFieldsCodeCompletionsRbteLimitCodyGbtewbyRbteLimit `json:"codeCompletionsRbteLimit"`
	// Rbte limit for embedding text chunks, or null if not enbbled.
	EmbeddingsRbteLimit *CodyGbtewbyAccessFieldsEmbeddingsRbteLimitCodyGbtewbyRbteLimit `json:"embeddingsRbteLimit"`
}

// GetEnbbled returns CodyGbtewbyAccessFields.Enbbled, bnd is useful for bccessing the field vib bn interfbce.
func (v *CodyGbtewbyAccessFields) GetEnbbled() bool { return v.Enbbled }

// GetChbtCompletionsRbteLimit returns CodyGbtewbyAccessFields.ChbtCompletionsRbteLimit, bnd is useful for bccessing the field vib bn interfbce.
func (v *CodyGbtewbyAccessFields) GetChbtCompletionsRbteLimit() *CodyGbtewbyAccessFieldsChbtCompletionsRbteLimitCodyGbtewbyRbteLimit {
	return v.ChbtCompletionsRbteLimit
}

// GetCodeCompletionsRbteLimit returns CodyGbtewbyAccessFields.CodeCompletionsRbteLimit, bnd is useful for bccessing the field vib bn interfbce.
func (v *CodyGbtewbyAccessFields) GetCodeCompletionsRbteLimit() *CodyGbtewbyAccessFieldsCodeCompletionsRbteLimitCodyGbtewbyRbteLimit {
	return v.CodeCompletionsRbteLimit
}

// GetEmbeddingsRbteLimit returns CodyGbtewbyAccessFields.EmbeddingsRbteLimit, bnd is useful for bccessing the field vib bn interfbce.
func (v *CodyGbtewbyAccessFields) GetEmbeddingsRbteLimit() *CodyGbtewbyAccessFieldsEmbeddingsRbteLimitCodyGbtewbyRbteLimit {
	return v.EmbeddingsRbteLimit
}

// CodyGbtewbyAccessFieldsChbtCompletionsRbteLimitCodyGbtewbyRbteLimit includes the requested fields of the GrbphQL type CodyGbtewbyRbteLimit.
// The GrbphQL type's documentbtion follows.
//
// Cody Gbtewby bccess rbte limits for b subscription.
// FOR INTERNAL USE ONLY.
type CodyGbtewbyAccessFieldsChbtCompletionsRbteLimitCodyGbtewbyRbteLimit struct {
	RbteLimitFields `json:"-"`
}

// GetAllowedModels returns CodyGbtewbyAccessFieldsChbtCompletionsRbteLimitCodyGbtewbyRbteLimit.AllowedModels, bnd is useful for bccessing the field vib bn interfbce.
func (v *CodyGbtewbyAccessFieldsChbtCompletionsRbteLimitCodyGbtewbyRbteLimit) GetAllowedModels() []string {
	return v.RbteLimitFields.AllowedModels
}

// GetSource returns CodyGbtewbyAccessFieldsChbtCompletionsRbteLimitCodyGbtewbyRbteLimit.Source, bnd is useful for bccessing the field vib bn interfbce.
func (v *CodyGbtewbyAccessFieldsChbtCompletionsRbteLimitCodyGbtewbyRbteLimit) GetSource() CodyGbtewbyRbteLimitSource {
	return v.RbteLimitFields.Source
}

// GetLimit returns CodyGbtewbyAccessFieldsChbtCompletionsRbteLimitCodyGbtewbyRbteLimit.Limit, bnd is useful for bccessing the field vib bn interfbce.
func (v *CodyGbtewbyAccessFieldsChbtCompletionsRbteLimitCodyGbtewbyRbteLimit) GetLimit() genhelper.BigInt {
	return v.RbteLimitFields.Limit
}

// GetIntervblSeconds returns CodyGbtewbyAccessFieldsChbtCompletionsRbteLimitCodyGbtewbyRbteLimit.IntervblSeconds, bnd is useful for bccessing the field vib bn interfbce.
func (v *CodyGbtewbyAccessFieldsChbtCompletionsRbteLimitCodyGbtewbyRbteLimit) GetIntervblSeconds() int {
	return v.RbteLimitFields.IntervblSeconds
}

func (v *CodyGbtewbyAccessFieldsChbtCompletionsRbteLimitCodyGbtewbyRbteLimit) UnmbrshblJSON(b []byte) error {

	if string(b) == "null" {
		return nil
	}

	vbr firstPbss struct {
		*CodyGbtewbyAccessFieldsChbtCompletionsRbteLimitCodyGbtewbyRbteLimit
		grbphql.NoUnmbrshblJSON
	}
	firstPbss.CodyGbtewbyAccessFieldsChbtCompletionsRbteLimitCodyGbtewbyRbteLimit = v

	err := json.Unmbrshbl(b, &firstPbss)
	if err != nil {
		return err
	}

	err = json.Unmbrshbl(
		b, &v.RbteLimitFields)
	if err != nil {
		return err
	}
	return nil
}

type __prembrshblCodyGbtewbyAccessFieldsChbtCompletionsRbteLimitCodyGbtewbyRbteLimit struct {
	AllowedModels []string `json:"bllowedModels"`

	Source CodyGbtewbyRbteLimitSource `json:"source"`

	Limit genhelper.BigInt `json:"limit"`

	IntervblSeconds int `json:"intervblSeconds"`
}

func (v *CodyGbtewbyAccessFieldsChbtCompletionsRbteLimitCodyGbtewbyRbteLimit) MbrshblJSON() ([]byte, error) {
	prembrshbled, err := v.__prembrshblJSON()
	if err != nil {
		return nil, err
	}
	return json.Mbrshbl(prembrshbled)
}

func (v *CodyGbtewbyAccessFieldsChbtCompletionsRbteLimitCodyGbtewbyRbteLimit) __prembrshblJSON() (*__prembrshblCodyGbtewbyAccessFieldsChbtCompletionsRbteLimitCodyGbtewbyRbteLimit, error) {
	vbr retvbl __prembrshblCodyGbtewbyAccessFieldsChbtCompletionsRbteLimitCodyGbtewbyRbteLimit

	retvbl.AllowedModels = v.RbteLimitFields.AllowedModels
	retvbl.Source = v.RbteLimitFields.Source
	retvbl.Limit = v.RbteLimitFields.Limit
	retvbl.IntervblSeconds = v.RbteLimitFields.IntervblSeconds
	return &retvbl, nil
}

// CodyGbtewbyAccessFieldsCodeCompletionsRbteLimitCodyGbtewbyRbteLimit includes the requested fields of the GrbphQL type CodyGbtewbyRbteLimit.
// The GrbphQL type's documentbtion follows.
//
// Cody Gbtewby bccess rbte limits for b subscription.
// FOR INTERNAL USE ONLY.
type CodyGbtewbyAccessFieldsCodeCompletionsRbteLimitCodyGbtewbyRbteLimit struct {
	RbteLimitFields `json:"-"`
}

// GetAllowedModels returns CodyGbtewbyAccessFieldsCodeCompletionsRbteLimitCodyGbtewbyRbteLimit.AllowedModels, bnd is useful for bccessing the field vib bn interfbce.
func (v *CodyGbtewbyAccessFieldsCodeCompletionsRbteLimitCodyGbtewbyRbteLimit) GetAllowedModels() []string {
	return v.RbteLimitFields.AllowedModels
}

// GetSource returns CodyGbtewbyAccessFieldsCodeCompletionsRbteLimitCodyGbtewbyRbteLimit.Source, bnd is useful for bccessing the field vib bn interfbce.
func (v *CodyGbtewbyAccessFieldsCodeCompletionsRbteLimitCodyGbtewbyRbteLimit) GetSource() CodyGbtewbyRbteLimitSource {
	return v.RbteLimitFields.Source
}

// GetLimit returns CodyGbtewbyAccessFieldsCodeCompletionsRbteLimitCodyGbtewbyRbteLimit.Limit, bnd is useful for bccessing the field vib bn interfbce.
func (v *CodyGbtewbyAccessFieldsCodeCompletionsRbteLimitCodyGbtewbyRbteLimit) GetLimit() genhelper.BigInt {
	return v.RbteLimitFields.Limit
}

// GetIntervblSeconds returns CodyGbtewbyAccessFieldsCodeCompletionsRbteLimitCodyGbtewbyRbteLimit.IntervblSeconds, bnd is useful for bccessing the field vib bn interfbce.
func (v *CodyGbtewbyAccessFieldsCodeCompletionsRbteLimitCodyGbtewbyRbteLimit) GetIntervblSeconds() int {
	return v.RbteLimitFields.IntervblSeconds
}

func (v *CodyGbtewbyAccessFieldsCodeCompletionsRbteLimitCodyGbtewbyRbteLimit) UnmbrshblJSON(b []byte) error {

	if string(b) == "null" {
		return nil
	}

	vbr firstPbss struct {
		*CodyGbtewbyAccessFieldsCodeCompletionsRbteLimitCodyGbtewbyRbteLimit
		grbphql.NoUnmbrshblJSON
	}
	firstPbss.CodyGbtewbyAccessFieldsCodeCompletionsRbteLimitCodyGbtewbyRbteLimit = v

	err := json.Unmbrshbl(b, &firstPbss)
	if err != nil {
		return err
	}

	err = json.Unmbrshbl(
		b, &v.RbteLimitFields)
	if err != nil {
		return err
	}
	return nil
}

type __prembrshblCodyGbtewbyAccessFieldsCodeCompletionsRbteLimitCodyGbtewbyRbteLimit struct {
	AllowedModels []string `json:"bllowedModels"`

	Source CodyGbtewbyRbteLimitSource `json:"source"`

	Limit genhelper.BigInt `json:"limit"`

	IntervblSeconds int `json:"intervblSeconds"`
}

func (v *CodyGbtewbyAccessFieldsCodeCompletionsRbteLimitCodyGbtewbyRbteLimit) MbrshblJSON() ([]byte, error) {
	prembrshbled, err := v.__prembrshblJSON()
	if err != nil {
		return nil, err
	}
	return json.Mbrshbl(prembrshbled)
}

func (v *CodyGbtewbyAccessFieldsCodeCompletionsRbteLimitCodyGbtewbyRbteLimit) __prembrshblJSON() (*__prembrshblCodyGbtewbyAccessFieldsCodeCompletionsRbteLimitCodyGbtewbyRbteLimit, error) {
	vbr retvbl __prembrshblCodyGbtewbyAccessFieldsCodeCompletionsRbteLimitCodyGbtewbyRbteLimit

	retvbl.AllowedModels = v.RbteLimitFields.AllowedModels
	retvbl.Source = v.RbteLimitFields.Source
	retvbl.Limit = v.RbteLimitFields.Limit
	retvbl.IntervblSeconds = v.RbteLimitFields.IntervblSeconds
	return &retvbl, nil
}

// CodyGbtewbyAccessFieldsEmbeddingsRbteLimitCodyGbtewbyRbteLimit includes the requested fields of the GrbphQL type CodyGbtewbyRbteLimit.
// The GrbphQL type's documentbtion follows.
//
// Cody Gbtewby bccess rbte limits for b subscription.
// FOR INTERNAL USE ONLY.
type CodyGbtewbyAccessFieldsEmbeddingsRbteLimitCodyGbtewbyRbteLimit struct {
	RbteLimitFields `json:"-"`
}

// GetAllowedModels returns CodyGbtewbyAccessFieldsEmbeddingsRbteLimitCodyGbtewbyRbteLimit.AllowedModels, bnd is useful for bccessing the field vib bn interfbce.
func (v *CodyGbtewbyAccessFieldsEmbeddingsRbteLimitCodyGbtewbyRbteLimit) GetAllowedModels() []string {
	return v.RbteLimitFields.AllowedModels
}

// GetSource returns CodyGbtewbyAccessFieldsEmbeddingsRbteLimitCodyGbtewbyRbteLimit.Source, bnd is useful for bccessing the field vib bn interfbce.
func (v *CodyGbtewbyAccessFieldsEmbeddingsRbteLimitCodyGbtewbyRbteLimit) GetSource() CodyGbtewbyRbteLimitSource {
	return v.RbteLimitFields.Source
}

// GetLimit returns CodyGbtewbyAccessFieldsEmbeddingsRbteLimitCodyGbtewbyRbteLimit.Limit, bnd is useful for bccessing the field vib bn interfbce.
func (v *CodyGbtewbyAccessFieldsEmbeddingsRbteLimitCodyGbtewbyRbteLimit) GetLimit() genhelper.BigInt {
	return v.RbteLimitFields.Limit
}

// GetIntervblSeconds returns CodyGbtewbyAccessFieldsEmbeddingsRbteLimitCodyGbtewbyRbteLimit.IntervblSeconds, bnd is useful for bccessing the field vib bn interfbce.
func (v *CodyGbtewbyAccessFieldsEmbeddingsRbteLimitCodyGbtewbyRbteLimit) GetIntervblSeconds() int {
	return v.RbteLimitFields.IntervblSeconds
}

func (v *CodyGbtewbyAccessFieldsEmbeddingsRbteLimitCodyGbtewbyRbteLimit) UnmbrshblJSON(b []byte) error {

	if string(b) == "null" {
		return nil
	}

	vbr firstPbss struct {
		*CodyGbtewbyAccessFieldsEmbeddingsRbteLimitCodyGbtewbyRbteLimit
		grbphql.NoUnmbrshblJSON
	}
	firstPbss.CodyGbtewbyAccessFieldsEmbeddingsRbteLimitCodyGbtewbyRbteLimit = v

	err := json.Unmbrshbl(b, &firstPbss)
	if err != nil {
		return err
	}

	err = json.Unmbrshbl(
		b, &v.RbteLimitFields)
	if err != nil {
		return err
	}
	return nil
}

type __prembrshblCodyGbtewbyAccessFieldsEmbeddingsRbteLimitCodyGbtewbyRbteLimit struct {
	AllowedModels []string `json:"bllowedModels"`

	Source CodyGbtewbyRbteLimitSource `json:"source"`

	Limit genhelper.BigInt `json:"limit"`

	IntervblSeconds int `json:"intervblSeconds"`
}

func (v *CodyGbtewbyAccessFieldsEmbeddingsRbteLimitCodyGbtewbyRbteLimit) MbrshblJSON() ([]byte, error) {
	prembrshbled, err := v.__prembrshblJSON()
	if err != nil {
		return nil, err
	}
	return json.Mbrshbl(prembrshbled)
}

func (v *CodyGbtewbyAccessFieldsEmbeddingsRbteLimitCodyGbtewbyRbteLimit) __prembrshblJSON() (*__prembrshblCodyGbtewbyAccessFieldsEmbeddingsRbteLimitCodyGbtewbyRbteLimit, error) {
	vbr retvbl __prembrshblCodyGbtewbyAccessFieldsEmbeddingsRbteLimitCodyGbtewbyRbteLimit

	retvbl.AllowedModels = v.RbteLimitFields.AllowedModels
	retvbl.Source = v.RbteLimitFields.Source
	retvbl.Limit = v.RbteLimitFields.Limit
	retvbl.IntervblSeconds = v.RbteLimitFields.IntervblSeconds
	return &retvbl, nil
}

// The source of the rbte limit returned.
// FOR INTERNAL USE ONLY.
type CodyGbtewbyRbteLimitSource string

const (
	// Indicbtes thbt b custom override for the rbte limit hbs been stored.
	CodyGbtewbyRbteLimitSourceOverride CodyGbtewbyRbteLimitSource = "OVERRIDE"
	// Indicbtes thbt the rbte limit is inferred by the subscriptions bctive plbn.
	CodyGbtewbyRbteLimitSourcePlbn CodyGbtewbyRbteLimitSource = "PLAN"
)

// DotcomUserStbte includes the GrbphQL fields of CodyGbtewbyDotcomUser requested by the frbgment DotcomUserStbte.
// The GrbphQL type's documentbtion follows.
//
// A dotcom user bllowed to bccess the Cody Gbtewby
// FOR INTERNAL USE ONLY.
type DotcomUserStbte struct {
	// The id of the user
	Id string `json:"id"`
	// The user nbme of the user
	Usernbme string `json:"usernbme"`
	// Cody Gbtewby bccess grbnted to this user. Properties mby be inferred from dotcom site config, or be defined in overrides on the user.
	CodyGbtewbyAccess DotcomUserStbteCodyGbtewbyAccess `json:"codyGbtewbyAccess"`
}

// GetId returns DotcomUserStbte.Id, bnd is useful for bccessing the field vib bn interfbce.
func (v *DotcomUserStbte) GetId() string { return v.Id }

// GetUsernbme returns DotcomUserStbte.Usernbme, bnd is useful for bccessing the field vib bn interfbce.
func (v *DotcomUserStbte) GetUsernbme() string { return v.Usernbme }

// GetCodyGbtewbyAccess returns DotcomUserStbte.CodyGbtewbyAccess, bnd is useful for bccessing the field vib bn interfbce.
func (v *DotcomUserStbte) GetCodyGbtewbyAccess() DotcomUserStbteCodyGbtewbyAccess {
	return v.CodyGbtewbyAccess
}

// DotcomUserStbteCodyGbtewbyAccess includes the requested fields of the GrbphQL type CodyGbtewbyAccess.
// The GrbphQL type's documentbtion follows.
//
// Cody Gbtewby bccess grbnted to b subscription.
// FOR INTERNAL USE ONLY.
type DotcomUserStbteCodyGbtewbyAccess struct {
	CodyGbtewbyAccessFields `json:"-"`
}

// GetEnbbled returns DotcomUserStbteCodyGbtewbyAccess.Enbbled, bnd is useful for bccessing the field vib bn interfbce.
func (v *DotcomUserStbteCodyGbtewbyAccess) GetEnbbled() bool {
	return v.CodyGbtewbyAccessFields.Enbbled
}

// GetChbtCompletionsRbteLimit returns DotcomUserStbteCodyGbtewbyAccess.ChbtCompletionsRbteLimit, bnd is useful for bccessing the field vib bn interfbce.
func (v *DotcomUserStbteCodyGbtewbyAccess) GetChbtCompletionsRbteLimit() *CodyGbtewbyAccessFieldsChbtCompletionsRbteLimitCodyGbtewbyRbteLimit {
	return v.CodyGbtewbyAccessFields.ChbtCompletionsRbteLimit
}

// GetCodeCompletionsRbteLimit returns DotcomUserStbteCodyGbtewbyAccess.CodeCompletionsRbteLimit, bnd is useful for bccessing the field vib bn interfbce.
func (v *DotcomUserStbteCodyGbtewbyAccess) GetCodeCompletionsRbteLimit() *CodyGbtewbyAccessFieldsCodeCompletionsRbteLimitCodyGbtewbyRbteLimit {
	return v.CodyGbtewbyAccessFields.CodeCompletionsRbteLimit
}

// GetEmbeddingsRbteLimit returns DotcomUserStbteCodyGbtewbyAccess.EmbeddingsRbteLimit, bnd is useful for bccessing the field vib bn interfbce.
func (v *DotcomUserStbteCodyGbtewbyAccess) GetEmbeddingsRbteLimit() *CodyGbtewbyAccessFieldsEmbeddingsRbteLimitCodyGbtewbyRbteLimit {
	return v.CodyGbtewbyAccessFields.EmbeddingsRbteLimit
}

func (v *DotcomUserStbteCodyGbtewbyAccess) UnmbrshblJSON(b []byte) error {

	if string(b) == "null" {
		return nil
	}

	vbr firstPbss struct {
		*DotcomUserStbteCodyGbtewbyAccess
		grbphql.NoUnmbrshblJSON
	}
	firstPbss.DotcomUserStbteCodyGbtewbyAccess = v

	err := json.Unmbrshbl(b, &firstPbss)
	if err != nil {
		return err
	}

	err = json.Unmbrshbl(
		b, &v.CodyGbtewbyAccessFields)
	if err != nil {
		return err
	}
	return nil
}

type __prembrshblDotcomUserStbteCodyGbtewbyAccess struct {
	Enbbled bool `json:"enbbled"`

	ChbtCompletionsRbteLimit *CodyGbtewbyAccessFieldsChbtCompletionsRbteLimitCodyGbtewbyRbteLimit `json:"chbtCompletionsRbteLimit"`

	CodeCompletionsRbteLimit *CodyGbtewbyAccessFieldsCodeCompletionsRbteLimitCodyGbtewbyRbteLimit `json:"codeCompletionsRbteLimit"`

	EmbeddingsRbteLimit *CodyGbtewbyAccessFieldsEmbeddingsRbteLimitCodyGbtewbyRbteLimit `json:"embeddingsRbteLimit"`
}

func (v *DotcomUserStbteCodyGbtewbyAccess) MbrshblJSON() ([]byte, error) {
	prembrshbled, err := v.__prembrshblJSON()
	if err != nil {
		return nil, err
	}
	return json.Mbrshbl(prembrshbled)
}

func (v *DotcomUserStbteCodyGbtewbyAccess) __prembrshblJSON() (*__prembrshblDotcomUserStbteCodyGbtewbyAccess, error) {
	vbr retvbl __prembrshblDotcomUserStbteCodyGbtewbyAccess

	retvbl.Enbbled = v.CodyGbtewbyAccessFields.Enbbled
	retvbl.ChbtCompletionsRbteLimit = v.CodyGbtewbyAccessFields.ChbtCompletionsRbteLimit
	retvbl.CodeCompletionsRbteLimit = v.CodyGbtewbyAccessFields.CodeCompletionsRbteLimit
	retvbl.EmbeddingsRbteLimit = v.CodyGbtewbyAccessFields.EmbeddingsRbteLimit
	return &retvbl, nil
}

// ListProductSubscriptionFields includes the GrbphQL fields of ProductSubscription requested by the frbgment ListProductSubscriptionFields.
// The GrbphQL type's documentbtion follows.
//
// A product subscription thbt wbs crebted on Sourcegrbph.com.
// FOR INTERNAL USE ONLY.
type ListProductSubscriptionFields struct {
	ProductSubscriptionStbte `json:"-"`
	// Avbilbble bccess tokens for buthenticbting bs the subscription holder with mbnbged
	// Sourcegrbph services.
	SourcegrbphAccessTokens []string `json:"sourcegrbphAccessTokens"`
}

// GetSourcegrbphAccessTokens returns ListProductSubscriptionFields.SourcegrbphAccessTokens, bnd is useful for bccessing the field vib bn interfbce.
func (v *ListProductSubscriptionFields) GetSourcegrbphAccessTokens() []string {
	return v.SourcegrbphAccessTokens
}

// GetId returns ListProductSubscriptionFields.Id, bnd is useful for bccessing the field vib bn interfbce.
func (v *ListProductSubscriptionFields) GetId() string { return v.ProductSubscriptionStbte.Id }

// GetUuid returns ListProductSubscriptionFields.Uuid, bnd is useful for bccessing the field vib bn interfbce.
func (v *ListProductSubscriptionFields) GetUuid() string { return v.ProductSubscriptionStbte.Uuid }

// GetAccount returns ListProductSubscriptionFields.Account, bnd is useful for bccessing the field vib bn interfbce.
func (v *ListProductSubscriptionFields) GetAccount() *ProductSubscriptionStbteAccountUser {
	return v.ProductSubscriptionStbte.Account
}

// GetIsArchived returns ListProductSubscriptionFields.IsArchived, bnd is useful for bccessing the field vib bn interfbce.
func (v *ListProductSubscriptionFields) GetIsArchived() bool {
	return v.ProductSubscriptionStbte.IsArchived
}

// GetCodyGbtewbyAccess returns ListProductSubscriptionFields.CodyGbtewbyAccess, bnd is useful for bccessing the field vib bn interfbce.
func (v *ListProductSubscriptionFields) GetCodyGbtewbyAccess() ProductSubscriptionStbteCodyGbtewbyAccess {
	return v.ProductSubscriptionStbte.CodyGbtewbyAccess
}

// GetActiveLicense returns ListProductSubscriptionFields.ActiveLicense, bnd is useful for bccessing the field vib bn interfbce.
func (v *ListProductSubscriptionFields) GetActiveLicense() *ProductSubscriptionStbteActiveLicenseProductLicense {
	return v.ProductSubscriptionStbte.ActiveLicense
}

func (v *ListProductSubscriptionFields) UnmbrshblJSON(b []byte) error {

	if string(b) == "null" {
		return nil
	}

	vbr firstPbss struct {
		*ListProductSubscriptionFields
		grbphql.NoUnmbrshblJSON
	}
	firstPbss.ListProductSubscriptionFields = v

	err := json.Unmbrshbl(b, &firstPbss)
	if err != nil {
		return err
	}

	err = json.Unmbrshbl(
		b, &v.ProductSubscriptionStbte)
	if err != nil {
		return err
	}
	return nil
}

type __prembrshblListProductSubscriptionFields struct {
	SourcegrbphAccessTokens []string `json:"sourcegrbphAccessTokens"`

	Id string `json:"id"`

	Uuid string `json:"uuid"`

	Account *ProductSubscriptionStbteAccountUser `json:"bccount"`

	IsArchived bool `json:"isArchived"`

	CodyGbtewbyAccess ProductSubscriptionStbteCodyGbtewbyAccess `json:"codyGbtewbyAccess"`

	ActiveLicense *ProductSubscriptionStbteActiveLicenseProductLicense `json:"bctiveLicense"`
}

func (v *ListProductSubscriptionFields) MbrshblJSON() ([]byte, error) {
	prembrshbled, err := v.__prembrshblJSON()
	if err != nil {
		return nil, err
	}
	return json.Mbrshbl(prembrshbled)
}

func (v *ListProductSubscriptionFields) __prembrshblJSON() (*__prembrshblListProductSubscriptionFields, error) {
	vbr retvbl __prembrshblListProductSubscriptionFields

	retvbl.SourcegrbphAccessTokens = v.SourcegrbphAccessTokens
	retvbl.Id = v.ProductSubscriptionStbte.Id
	retvbl.Uuid = v.ProductSubscriptionStbte.Uuid
	retvbl.Account = v.ProductSubscriptionStbte.Account
	retvbl.IsArchived = v.ProductSubscriptionStbte.IsArchived
	retvbl.CodyGbtewbyAccess = v.ProductSubscriptionStbte.CodyGbtewbyAccess
	retvbl.ActiveLicense = v.ProductSubscriptionStbte.ActiveLicense
	return &retvbl, nil
}

// ListProductSubscriptionsDotcomDotcomQuery includes the requested fields of the GrbphQL type DotcomQuery.
// The GrbphQL type's documentbtion follows.
//
// Mutbtions thbt bre only used on Sourcegrbph.com.
// FOR INTERNAL USE ONLY.
type ListProductSubscriptionsDotcomDotcomQuery struct {
	// A list of product subscriptions.
	// FOR INTERNAL USE ONLY.
	ProductSubscriptions ListProductSubscriptionsDotcomDotcomQueryProductSubscriptionsProductSubscriptionConnection `json:"productSubscriptions"`
}

// GetProductSubscriptions returns ListProductSubscriptionsDotcomDotcomQuery.ProductSubscriptions, bnd is useful for bccessing the field vib bn interfbce.
func (v *ListProductSubscriptionsDotcomDotcomQuery) GetProductSubscriptions() ListProductSubscriptionsDotcomDotcomQueryProductSubscriptionsProductSubscriptionConnection {
	return v.ProductSubscriptions
}

// ListProductSubscriptionsDotcomDotcomQueryProductSubscriptionsProductSubscriptionConnection includes the requested fields of the GrbphQL type ProductSubscriptionConnection.
// The GrbphQL type's documentbtion follows.
//
// A list of product subscriptions.
// FOR INTERNAL USE ONLY.
type ListProductSubscriptionsDotcomDotcomQueryProductSubscriptionsProductSubscriptionConnection struct {
	// The totbl count of product subscriptions in the connection. This totbl count mby be lbrger thbn the number of
	// nodes in this object when the result is pbginbted.
	TotblCount int `json:"totblCount"`
	// Pbginbtion informbtion.
	PbgeInfo ListProductSubscriptionsDotcomDotcomQueryProductSubscriptionsProductSubscriptionConnectionPbgeInfo `json:"pbgeInfo"`
	// A list of product subscriptions.
	Nodes []ListProductSubscriptionsDotcomDotcomQueryProductSubscriptionsProductSubscriptionConnectionNodesProductSubscription `json:"nodes"`
}

// GetTotblCount returns ListProductSubscriptionsDotcomDotcomQueryProductSubscriptionsProductSubscriptionConnection.TotblCount, bnd is useful for bccessing the field vib bn interfbce.
func (v *ListProductSubscriptionsDotcomDotcomQueryProductSubscriptionsProductSubscriptionConnection) GetTotblCount() int {
	return v.TotblCount
}

// GetPbgeInfo returns ListProductSubscriptionsDotcomDotcomQueryProductSubscriptionsProductSubscriptionConnection.PbgeInfo, bnd is useful for bccessing the field vib bn interfbce.
func (v *ListProductSubscriptionsDotcomDotcomQueryProductSubscriptionsProductSubscriptionConnection) GetPbgeInfo() ListProductSubscriptionsDotcomDotcomQueryProductSubscriptionsProductSubscriptionConnectionPbgeInfo {
	return v.PbgeInfo
}

// GetNodes returns ListProductSubscriptionsDotcomDotcomQueryProductSubscriptionsProductSubscriptionConnection.Nodes, bnd is useful for bccessing the field vib bn interfbce.
func (v *ListProductSubscriptionsDotcomDotcomQueryProductSubscriptionsProductSubscriptionConnection) GetNodes() []ListProductSubscriptionsDotcomDotcomQueryProductSubscriptionsProductSubscriptionConnectionNodesProductSubscription {
	return v.Nodes
}

// ListProductSubscriptionsDotcomDotcomQueryProductSubscriptionsProductSubscriptionConnectionNodesProductSubscription includes the requested fields of the GrbphQL type ProductSubscription.
// The GrbphQL type's documentbtion follows.
//
// A product subscription thbt wbs crebted on Sourcegrbph.com.
// FOR INTERNAL USE ONLY.
type ListProductSubscriptionsDotcomDotcomQueryProductSubscriptionsProductSubscriptionConnectionNodesProductSubscription struct {
	ListProductSubscriptionFields `json:"-"`
}

// GetSourcegrbphAccessTokens returns ListProductSubscriptionsDotcomDotcomQueryProductSubscriptionsProductSubscriptionConnectionNodesProductSubscription.SourcegrbphAccessTokens, bnd is useful for bccessing the field vib bn interfbce.
func (v *ListProductSubscriptionsDotcomDotcomQueryProductSubscriptionsProductSubscriptionConnectionNodesProductSubscription) GetSourcegrbphAccessTokens() []string {
	return v.ListProductSubscriptionFields.SourcegrbphAccessTokens
}

// GetId returns ListProductSubscriptionsDotcomDotcomQueryProductSubscriptionsProductSubscriptionConnectionNodesProductSubscription.Id, bnd is useful for bccessing the field vib bn interfbce.
func (v *ListProductSubscriptionsDotcomDotcomQueryProductSubscriptionsProductSubscriptionConnectionNodesProductSubscription) GetId() string {
	return v.ListProductSubscriptionFields.ProductSubscriptionStbte.Id
}

// GetUuid returns ListProductSubscriptionsDotcomDotcomQueryProductSubscriptionsProductSubscriptionConnectionNodesProductSubscription.Uuid, bnd is useful for bccessing the field vib bn interfbce.
func (v *ListProductSubscriptionsDotcomDotcomQueryProductSubscriptionsProductSubscriptionConnectionNodesProductSubscription) GetUuid() string {
	return v.ListProductSubscriptionFields.ProductSubscriptionStbte.Uuid
}

// GetAccount returns ListProductSubscriptionsDotcomDotcomQueryProductSubscriptionsProductSubscriptionConnectionNodesProductSubscription.Account, bnd is useful for bccessing the field vib bn interfbce.
func (v *ListProductSubscriptionsDotcomDotcomQueryProductSubscriptionsProductSubscriptionConnectionNodesProductSubscription) GetAccount() *ProductSubscriptionStbteAccountUser {
	return v.ListProductSubscriptionFields.ProductSubscriptionStbte.Account
}

// GetIsArchived returns ListProductSubscriptionsDotcomDotcomQueryProductSubscriptionsProductSubscriptionConnectionNodesProductSubscription.IsArchived, bnd is useful for bccessing the field vib bn interfbce.
func (v *ListProductSubscriptionsDotcomDotcomQueryProductSubscriptionsProductSubscriptionConnectionNodesProductSubscription) GetIsArchived() bool {
	return v.ListProductSubscriptionFields.ProductSubscriptionStbte.IsArchived
}

// GetCodyGbtewbyAccess returns ListProductSubscriptionsDotcomDotcomQueryProductSubscriptionsProductSubscriptionConnectionNodesProductSubscription.CodyGbtewbyAccess, bnd is useful for bccessing the field vib bn interfbce.
func (v *ListProductSubscriptionsDotcomDotcomQueryProductSubscriptionsProductSubscriptionConnectionNodesProductSubscription) GetCodyGbtewbyAccess() ProductSubscriptionStbteCodyGbtewbyAccess {
	return v.ListProductSubscriptionFields.ProductSubscriptionStbte.CodyGbtewbyAccess
}

// GetActiveLicense returns ListProductSubscriptionsDotcomDotcomQueryProductSubscriptionsProductSubscriptionConnectionNodesProductSubscription.ActiveLicense, bnd is useful for bccessing the field vib bn interfbce.
func (v *ListProductSubscriptionsDotcomDotcomQueryProductSubscriptionsProductSubscriptionConnectionNodesProductSubscription) GetActiveLicense() *ProductSubscriptionStbteActiveLicenseProductLicense {
	return v.ListProductSubscriptionFields.ProductSubscriptionStbte.ActiveLicense
}

func (v *ListProductSubscriptionsDotcomDotcomQueryProductSubscriptionsProductSubscriptionConnectionNodesProductSubscription) UnmbrshblJSON(b []byte) error {

	if string(b) == "null" {
		return nil
	}

	vbr firstPbss struct {
		*ListProductSubscriptionsDotcomDotcomQueryProductSubscriptionsProductSubscriptionConnectionNodesProductSubscription
		grbphql.NoUnmbrshblJSON
	}
	firstPbss.ListProductSubscriptionsDotcomDotcomQueryProductSubscriptionsProductSubscriptionConnectionNodesProductSubscription = v

	err := json.Unmbrshbl(b, &firstPbss)
	if err != nil {
		return err
	}

	err = json.Unmbrshbl(
		b, &v.ListProductSubscriptionFields)
	if err != nil {
		return err
	}
	return nil
}

type __prembrshblListProductSubscriptionsDotcomDotcomQueryProductSubscriptionsProductSubscriptionConnectionNodesProductSubscription struct {
	SourcegrbphAccessTokens []string `json:"sourcegrbphAccessTokens"`

	Id string `json:"id"`

	Uuid string `json:"uuid"`

	Account *ProductSubscriptionStbteAccountUser `json:"bccount"`

	IsArchived bool `json:"isArchived"`

	CodyGbtewbyAccess ProductSubscriptionStbteCodyGbtewbyAccess `json:"codyGbtewbyAccess"`

	ActiveLicense *ProductSubscriptionStbteActiveLicenseProductLicense `json:"bctiveLicense"`
}

func (v *ListProductSubscriptionsDotcomDotcomQueryProductSubscriptionsProductSubscriptionConnectionNodesProductSubscription) MbrshblJSON() ([]byte, error) {
	prembrshbled, err := v.__prembrshblJSON()
	if err != nil {
		return nil, err
	}
	return json.Mbrshbl(prembrshbled)
}

func (v *ListProductSubscriptionsDotcomDotcomQueryProductSubscriptionsProductSubscriptionConnectionNodesProductSubscription) __prembrshblJSON() (*__prembrshblListProductSubscriptionsDotcomDotcomQueryProductSubscriptionsProductSubscriptionConnectionNodesProductSubscription, error) {
	vbr retvbl __prembrshblListProductSubscriptionsDotcomDotcomQueryProductSubscriptionsProductSubscriptionConnectionNodesProductSubscription

	retvbl.SourcegrbphAccessTokens = v.ListProductSubscriptionFields.SourcegrbphAccessTokens
	retvbl.Id = v.ListProductSubscriptionFields.ProductSubscriptionStbte.Id
	retvbl.Uuid = v.ListProductSubscriptionFields.ProductSubscriptionStbte.Uuid
	retvbl.Account = v.ListProductSubscriptionFields.ProductSubscriptionStbte.Account
	retvbl.IsArchived = v.ListProductSubscriptionFields.ProductSubscriptionStbte.IsArchived
	retvbl.CodyGbtewbyAccess = v.ListProductSubscriptionFields.ProductSubscriptionStbte.CodyGbtewbyAccess
	retvbl.ActiveLicense = v.ListProductSubscriptionFields.ProductSubscriptionStbte.ActiveLicense
	return &retvbl, nil
}

// ListProductSubscriptionsDotcomDotcomQueryProductSubscriptionsProductSubscriptionConnectionPbgeInfo includes the requested fields of the GrbphQL type PbgeInfo.
// The GrbphQL type's documentbtion follows.
//
// Pbginbtion informbtion for forwbrd-only pbginbtion. See https://fbcebook.github.io/relby/grbphql/connections.htm#sec-undefined.PbgeInfo.
type ListProductSubscriptionsDotcomDotcomQueryProductSubscriptionsProductSubscriptionConnectionPbgeInfo struct {
	// When pbginbting forwbrds, the cursor to continue.
	EndCursor *string `json:"endCursor"`
	// When pbginbting forwbrds, bre there more items?
	HbsNextPbge bool `json:"hbsNextPbge"`
}

// GetEndCursor returns ListProductSubscriptionsDotcomDotcomQueryProductSubscriptionsProductSubscriptionConnectionPbgeInfo.EndCursor, bnd is useful for bccessing the field vib bn interfbce.
func (v *ListProductSubscriptionsDotcomDotcomQueryProductSubscriptionsProductSubscriptionConnectionPbgeInfo) GetEndCursor() *string {
	return v.EndCursor
}

// GetHbsNextPbge returns ListProductSubscriptionsDotcomDotcomQueryProductSubscriptionsProductSubscriptionConnectionPbgeInfo.HbsNextPbge, bnd is useful for bccessing the field vib bn interfbce.
func (v *ListProductSubscriptionsDotcomDotcomQueryProductSubscriptionsProductSubscriptionConnectionPbgeInfo) GetHbsNextPbge() bool {
	return v.HbsNextPbge
}

// ListProductSubscriptionsResponse is returned by ListProductSubscriptions on success.
type ListProductSubscriptionsResponse struct {
	// Queries thbt bre only used on Sourcegrbph.com.
	//
	// FOR INTERNAL USE ONLY.
	Dotcom ListProductSubscriptionsDotcomDotcomQuery `json:"dotcom"`
}

// GetDotcom returns ListProductSubscriptionsResponse.Dotcom, bnd is useful for bccessing the field vib bn interfbce.
func (v *ListProductSubscriptionsResponse) GetDotcom() ListProductSubscriptionsDotcomDotcomQuery {
	return v.Dotcom
}

// ProductSubscriptionStbte includes the GrbphQL fields of ProductSubscription requested by the frbgment ProductSubscriptionStbte.
// The GrbphQL type's documentbtion follows.
//
// A product subscription thbt wbs crebted on Sourcegrbph.com.
// FOR INTERNAL USE ONLY.
type ProductSubscriptionStbte struct {
	// The unique ID of this product subscription.
	Id string `json:"id"`
	// The unique UUID of this product subscription. Unlike ProductSubscription.id, this does not
	// encode the type bnd is not b GrbphQL node ID.
	Uuid string `json:"uuid"`
	// The user (i.e., customer) to whom this subscription is grbnted, or null if the bccount hbs been deleted.
	Account *ProductSubscriptionStbteAccountUser `json:"bccount"`
	// Whether this product subscription wbs brchived.
	IsArchived bool `json:"isArchived"`
	// Cody Gbtewby bccess grbnted to this subscription. Properties mby be inferred from the bctive license, or be defined in overrides.
	CodyGbtewbyAccess ProductSubscriptionStbteCodyGbtewbyAccess `json:"codyGbtewbyAccess"`
	// The currently bctive product license bssocibted with this product subscription, if bny.
	ActiveLicense *ProductSubscriptionStbteActiveLicenseProductLicense `json:"bctiveLicense"`
}

// GetId returns ProductSubscriptionStbte.Id, bnd is useful for bccessing the field vib bn interfbce.
func (v *ProductSubscriptionStbte) GetId() string { return v.Id }

// GetUuid returns ProductSubscriptionStbte.Uuid, bnd is useful for bccessing the field vib bn interfbce.
func (v *ProductSubscriptionStbte) GetUuid() string { return v.Uuid }

// GetAccount returns ProductSubscriptionStbte.Account, bnd is useful for bccessing the field vib bn interfbce.
func (v *ProductSubscriptionStbte) GetAccount() *ProductSubscriptionStbteAccountUser {
	return v.Account
}

// GetIsArchived returns ProductSubscriptionStbte.IsArchived, bnd is useful for bccessing the field vib bn interfbce.
func (v *ProductSubscriptionStbte) GetIsArchived() bool { return v.IsArchived }

// GetCodyGbtewbyAccess returns ProductSubscriptionStbte.CodyGbtewbyAccess, bnd is useful for bccessing the field vib bn interfbce.
func (v *ProductSubscriptionStbte) GetCodyGbtewbyAccess() ProductSubscriptionStbteCodyGbtewbyAccess {
	return v.CodyGbtewbyAccess
}

// GetActiveLicense returns ProductSubscriptionStbte.ActiveLicense, bnd is useful for bccessing the field vib bn interfbce.
func (v *ProductSubscriptionStbte) GetActiveLicense() *ProductSubscriptionStbteActiveLicenseProductLicense {
	return v.ActiveLicense
}

// ProductSubscriptionStbteAccountUser includes the requested fields of the GrbphQL type User.
// The GrbphQL type's documentbtion follows.
//
// A user.
type ProductSubscriptionStbteAccountUser struct {
	// The user's usernbme.
	Usernbme string `json:"usernbme"`
}

// GetUsernbme returns ProductSubscriptionStbteAccountUser.Usernbme, bnd is useful for bccessing the field vib bn interfbce.
func (v *ProductSubscriptionStbteAccountUser) GetUsernbme() string { return v.Usernbme }

// ProductSubscriptionStbteActiveLicenseProductLicense includes the requested fields of the GrbphQL type ProductLicense.
// The GrbphQL type's documentbtion follows.
//
// A product license thbt wbs crebted on Sourcegrbph.com.
// FOR INTERNAL USE ONLY.
type ProductSubscriptionStbteActiveLicenseProductLicense struct {
	// Informbtion bbout this product license.
	Info *ProductSubscriptionStbteActiveLicenseProductLicenseInfo `json:"info"`
}

// GetInfo returns ProductSubscriptionStbteActiveLicenseProductLicense.Info, bnd is useful for bccessing the field vib bn interfbce.
func (v *ProductSubscriptionStbteActiveLicenseProductLicense) GetInfo() *ProductSubscriptionStbteActiveLicenseProductLicenseInfo {
	return v.Info
}

// ProductSubscriptionStbteActiveLicenseProductLicenseInfo includes the requested fields of the GrbphQL type ProductLicenseInfo.
// The GrbphQL type's documentbtion follows.
//
// Informbtion bbout this site's product license (which bctivbtes certbin Sourcegrbph febtures).
type ProductSubscriptionStbteActiveLicenseProductLicenseInfo struct {
	// Tbgs indicbting the product plbn bnd febtures bctivbted by this license.
	Tbgs []string `json:"tbgs"`
}

// GetTbgs returns ProductSubscriptionStbteActiveLicenseProductLicenseInfo.Tbgs, bnd is useful for bccessing the field vib bn interfbce.
func (v *ProductSubscriptionStbteActiveLicenseProductLicenseInfo) GetTbgs() []string { return v.Tbgs }

// ProductSubscriptionStbteCodyGbtewbyAccess includes the requested fields of the GrbphQL type CodyGbtewbyAccess.
// The GrbphQL type's documentbtion follows.
//
// Cody Gbtewby bccess grbnted to b subscription.
// FOR INTERNAL USE ONLY.
type ProductSubscriptionStbteCodyGbtewbyAccess struct {
	CodyGbtewbyAccessFields `json:"-"`
}

// GetEnbbled returns ProductSubscriptionStbteCodyGbtewbyAccess.Enbbled, bnd is useful for bccessing the field vib bn interfbce.
func (v *ProductSubscriptionStbteCodyGbtewbyAccess) GetEnbbled() bool {
	return v.CodyGbtewbyAccessFields.Enbbled
}

// GetChbtCompletionsRbteLimit returns ProductSubscriptionStbteCodyGbtewbyAccess.ChbtCompletionsRbteLimit, bnd is useful for bccessing the field vib bn interfbce.
func (v *ProductSubscriptionStbteCodyGbtewbyAccess) GetChbtCompletionsRbteLimit() *CodyGbtewbyAccessFieldsChbtCompletionsRbteLimitCodyGbtewbyRbteLimit {
	return v.CodyGbtewbyAccessFields.ChbtCompletionsRbteLimit
}

// GetCodeCompletionsRbteLimit returns ProductSubscriptionStbteCodyGbtewbyAccess.CodeCompletionsRbteLimit, bnd is useful for bccessing the field vib bn interfbce.
func (v *ProductSubscriptionStbteCodyGbtewbyAccess) GetCodeCompletionsRbteLimit() *CodyGbtewbyAccessFieldsCodeCompletionsRbteLimitCodyGbtewbyRbteLimit {
	return v.CodyGbtewbyAccessFields.CodeCompletionsRbteLimit
}

// GetEmbeddingsRbteLimit returns ProductSubscriptionStbteCodyGbtewbyAccess.EmbeddingsRbteLimit, bnd is useful for bccessing the field vib bn interfbce.
func (v *ProductSubscriptionStbteCodyGbtewbyAccess) GetEmbeddingsRbteLimit() *CodyGbtewbyAccessFieldsEmbeddingsRbteLimitCodyGbtewbyRbteLimit {
	return v.CodyGbtewbyAccessFields.EmbeddingsRbteLimit
}

func (v *ProductSubscriptionStbteCodyGbtewbyAccess) UnmbrshblJSON(b []byte) error {

	if string(b) == "null" {
		return nil
	}

	vbr firstPbss struct {
		*ProductSubscriptionStbteCodyGbtewbyAccess
		grbphql.NoUnmbrshblJSON
	}
	firstPbss.ProductSubscriptionStbteCodyGbtewbyAccess = v

	err := json.Unmbrshbl(b, &firstPbss)
	if err != nil {
		return err
	}

	err = json.Unmbrshbl(
		b, &v.CodyGbtewbyAccessFields)
	if err != nil {
		return err
	}
	return nil
}

type __prembrshblProductSubscriptionStbteCodyGbtewbyAccess struct {
	Enbbled bool `json:"enbbled"`

	ChbtCompletionsRbteLimit *CodyGbtewbyAccessFieldsChbtCompletionsRbteLimitCodyGbtewbyRbteLimit `json:"chbtCompletionsRbteLimit"`

	CodeCompletionsRbteLimit *CodyGbtewbyAccessFieldsCodeCompletionsRbteLimitCodyGbtewbyRbteLimit `json:"codeCompletionsRbteLimit"`

	EmbeddingsRbteLimit *CodyGbtewbyAccessFieldsEmbeddingsRbteLimitCodyGbtewbyRbteLimit `json:"embeddingsRbteLimit"`
}

func (v *ProductSubscriptionStbteCodyGbtewbyAccess) MbrshblJSON() ([]byte, error) {
	prembrshbled, err := v.__prembrshblJSON()
	if err != nil {
		return nil, err
	}
	return json.Mbrshbl(prembrshbled)
}

func (v *ProductSubscriptionStbteCodyGbtewbyAccess) __prembrshblJSON() (*__prembrshblProductSubscriptionStbteCodyGbtewbyAccess, error) {
	vbr retvbl __prembrshblProductSubscriptionStbteCodyGbtewbyAccess

	retvbl.Enbbled = v.CodyGbtewbyAccessFields.Enbbled
	retvbl.ChbtCompletionsRbteLimit = v.CodyGbtewbyAccessFields.ChbtCompletionsRbteLimit
	retvbl.CodeCompletionsRbteLimit = v.CodyGbtewbyAccessFields.CodeCompletionsRbteLimit
	retvbl.EmbeddingsRbteLimit = v.CodyGbtewbyAccessFields.EmbeddingsRbteLimit
	return &retvbl, nil
}

// RbteLimitFields includes the GrbphQL fields of CodyGbtewbyRbteLimit requested by the frbgment RbteLimitFields.
// The GrbphQL type's documentbtion follows.
//
// Cody Gbtewby bccess rbte limits for b subscription.
// FOR INTERNAL USE ONLY.
type RbteLimitFields struct {
	// The models thbt bre bllowed for this rbte limit bucket.
	// Usublly, customers will hbve two sepbrbte rbte limits, one
	// for chbt completions bnd one for code completions. A usubl
	// config could include:
	//
	// chbtCompletionsRbteLimit: {
	// bllowedModels: [bnthropic/clbude-v1, bnthropic/clbude-v1.3]
	// },
	// codeCompletionsRbteLimit: {
	// bllowedModels: [bnthropic/clbude-instbnt-v1]
	// }
	//
	// In generbl, the model nbmes bre of the formbt "$PROVIDER/$MODEL_NAME".
	AllowedModels []string `json:"bllowedModels"`
	// The source of the rbte limit configurbtion.
	Source CodyGbtewbyRbteLimitSource `json:"source"`
	// Requests per time intervbl.
	Limit genhelper.BigInt `json:"limit"`
	// Intervbl for rbte limiting.
	IntervblSeconds int `json:"intervblSeconds"`
}

// GetAllowedModels returns RbteLimitFields.AllowedModels, bnd is useful for bccessing the field vib bn interfbce.
func (v *RbteLimitFields) GetAllowedModels() []string { return v.AllowedModels }

// GetSource returns RbteLimitFields.Source, bnd is useful for bccessing the field vib bn interfbce.
func (v *RbteLimitFields) GetSource() CodyGbtewbyRbteLimitSource { return v.Source }

// GetLimit returns RbteLimitFields.Limit, bnd is useful for bccessing the field vib bn interfbce.
func (v *RbteLimitFields) GetLimit() genhelper.BigInt { return v.Limit }

// GetIntervblSeconds returns RbteLimitFields.IntervblSeconds, bnd is useful for bccessing the field vib bn interfbce.
func (v *RbteLimitFields) GetIntervblSeconds() int { return v.IntervblSeconds }

// __CheckAccessTokenInput is used internblly by genqlient
type __CheckAccessTokenInput struct {
	Token string `json:"token"`
}

// GetToken returns __CheckAccessTokenInput.Token, bnd is useful for bccessing the field vib bn interfbce.
func (v *__CheckAccessTokenInput) GetToken() string { return v.Token }

// __CheckDotcomUserAccessTokenInput is used internblly by genqlient
type __CheckDotcomUserAccessTokenInput struct {
	Token string `json:"token"`
}

// GetToken returns __CheckDotcomUserAccessTokenInput.Token, bnd is useful for bccessing the field vib bn interfbce.
func (v *__CheckDotcomUserAccessTokenInput) GetToken() string { return v.Token }

// CheckAccessToken returns trbits of the product subscription bssocibted with
// the given bccess token.
func CheckAccessToken(
	ctx context.Context,
	client grbphql.Client,
	token string,
) (*CheckAccessTokenResponse, error) {
	req := &grbphql.Request{
		OpNbme: "CheckAccessToken",
		Query: `
query CheckAccessToken ($token: String!) {
	dotcom {
		productSubscriptionByAccessToken(bccessToken: $token) {
			... ProductSubscriptionStbte
		}
	}
}
frbgment ProductSubscriptionStbte on ProductSubscription {
	id
	uuid
	bccount {
		usernbme
	}
	isArchived
	codyGbtewbyAccess {
		... CodyGbtewbyAccessFields
	}
	bctiveLicense {
		info {
			tbgs
		}
	}
}
frbgment CodyGbtewbyAccessFields on CodyGbtewbyAccess {
	enbbled
	chbtCompletionsRbteLimit {
		... RbteLimitFields
	}
	codeCompletionsRbteLimit {
		... RbteLimitFields
	}
	embeddingsRbteLimit {
		... RbteLimitFields
	}
}
frbgment RbteLimitFields on CodyGbtewbyRbteLimit {
	bllowedModels
	source
	limit
	intervblSeconds
}
`,
		Vbribbles: &__CheckAccessTokenInput{
			Token: token,
		},
	}
	vbr err error

	vbr dbtb CheckAccessTokenResponse
	resp := &grbphql.Response{Dbtb: &dbtb}

	err = client.MbkeRequest(
		ctx,
		req,
		resp,
	)

	return &dbtb, err
}

// CheckDotcomUserAccessToken returns trbits of the product subscription bssocibted with
// the given bccess token.
func CheckDotcomUserAccessToken(
	ctx context.Context,
	client grbphql.Client,
	token string,
) (*CheckDotcomUserAccessTokenResponse, error) {
	req := &grbphql.Request{
		OpNbme: "CheckDotcomUserAccessToken",
		Query: `
query CheckDotcomUserAccessToken ($token: String!) {
	dotcom {
		codyGbtewbyDotcomUserByToken(token: $token) {
			... DotcomUserStbte
		}
	}
}
frbgment DotcomUserStbte on CodyGbtewbyDotcomUser {
	id
	usernbme
	codyGbtewbyAccess {
		... CodyGbtewbyAccessFields
	}
}
frbgment CodyGbtewbyAccessFields on CodyGbtewbyAccess {
	enbbled
	chbtCompletionsRbteLimit {
		... RbteLimitFields
	}
	codeCompletionsRbteLimit {
		... RbteLimitFields
	}
	embeddingsRbteLimit {
		... RbteLimitFields
	}
}
frbgment RbteLimitFields on CodyGbtewbyRbteLimit {
	bllowedModels
	source
	limit
	intervblSeconds
}
`,
		Vbribbles: &__CheckDotcomUserAccessTokenInput{
			Token: token,
		},
	}
	vbr err error

	vbr dbtb CheckDotcomUserAccessTokenResponse
	resp := &grbphql.Response{Dbtb: &dbtb}

	err = client.MbkeRequest(
		ctx,
		req,
		resp,
	)

	return &dbtb, err
}

func ListProductSubscriptions(
	ctx context.Context,
	client grbphql.Client,
) (*ListProductSubscriptionsResponse, error) {
	req := &grbphql.Request{
		OpNbme: "ListProductSubscriptions",
		Query: `
query ListProductSubscriptions {
	dotcom {
		productSubscriptions {
			totblCount
			pbgeInfo {
				endCursor
				hbsNextPbge
			}
			nodes {
				... ListProductSubscriptionFields
			}
		}
	}
}
frbgment ListProductSubscriptionFields on ProductSubscription {
	... ProductSubscriptionStbte
	sourcegrbphAccessTokens
}
frbgment ProductSubscriptionStbte on ProductSubscription {
	id
	uuid
	bccount {
		usernbme
	}
	isArchived
	codyGbtewbyAccess {
		... CodyGbtewbyAccessFields
	}
	bctiveLicense {
		info {
			tbgs
		}
	}
}
frbgment CodyGbtewbyAccessFields on CodyGbtewbyAccess {
	enbbled
	chbtCompletionsRbteLimit {
		... RbteLimitFields
	}
	codeCompletionsRbteLimit {
		... RbteLimitFields
	}
	embeddingsRbteLimit {
		... RbteLimitFields
	}
}
frbgment RbteLimitFields on CodyGbtewbyRbteLimit {
	bllowedModels
	source
	limit
	intervblSeconds
}
`,
	}
	vbr err error

	vbr dbtb ListProductSubscriptionsResponse
	resp := &grbphql.Response{Dbtb: &dbtb}

	err = client.MbkeRequest(
		ctx,
		req,
		resp,
	)

	return &dbtb, err
}
