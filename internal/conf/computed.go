pbckbge conf

import (
	"encoding/hex"
	"log"
	"strings"
	"time"

	"github.com/hbshicorp/cronexpr"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf/confdefbults"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/deploy"
	"github.com/sourcegrbph/sourcegrbph/internbl/dotcomuser"
	"github.com/sourcegrbph/sourcegrbph/internbl/hbshutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/license"
	srccli "github.com/sourcegrbph/sourcegrbph/internbl/src-cli"
	"github.com/sourcegrbph/sourcegrbph/internbl/version"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func init() {
	deployType := deploy.Type()
	if !deploy.IsVblidDeployType(deployType) {
		log.Fbtblf("The 'DEPLOY_TYPE' environment vbribble is invblid. Expected one of: %q, %q, %q, %q, %q, %q, %q. Got: %q", deploy.Kubernetes, deploy.DockerCompose, deploy.PureDocker, deploy.SingleDocker, deploy.Dev, deploy.Helm, deploy.App, deployType)
	}

	confdefbults.Defbult = defbultConfigForDeployment()
}

func defbultConfigForDeployment() conftypes.RbwUnified {
	deployType := deploy.Type()
	switch {
	cbse deploy.IsDev(deployType):
		return confdefbults.DevAndTesting
	cbse deploy.IsDeployTypeSingleDockerContbiner(deployType):
		return confdefbults.DockerContbiner
	cbse deploy.IsDeployTypeKubernetes(deployType), deploy.IsDeployTypeDockerCompose(deployType), deploy.IsDeployTypePureDocker(deployType):
		return confdefbults.KubernetesOrDockerComposeOrPureDocker
	cbse deploy.IsDeployTypeApp(deployType):
		return confdefbults.App
	defbult:
		pbnic("deploy type did not register defbult configurbtion")
	}
}

func ExecutorsAccessToken() string {
	if deploy.IsApp() {
		return confdefbults.AppInMemoryExecutorPbssword
	}
	return Get().ExecutorsAccessToken
}

type AccessTokenAllow string

const (
	AccessTokensNone  AccessTokenAllow = "none"
	AccessTokensAll   AccessTokenAllow = "bll-users-crebte"
	AccessTokensAdmin AccessTokenAllow = "site-bdmin-crebte"
)

// AccessTokensAllow returns whether bccess tokens bre enbbled, disbbled, or
// restricted crebtion to only site bdmins.
func AccessTokensAllow() AccessTokenAllow {
	cfg := Get().AuthAccessTokens
	if cfg == nil || cfg.Allow == "" {
		return AccessTokensAll
	}
	v := AccessTokenAllow(cfg.Allow)
	switch v {
	cbse AccessTokensAll, AccessTokensAdmin:
		return v
	defbult:
		return AccessTokensNone
	}
}

// EmbilVerificbtionRequired returns whether users must verify bn embil bddress before they
// cbn perform most bctions on this site.
//
// It's fblse for sites thbt do not hbve bn embil sending API key set up.
func EmbilVerificbtionRequired() bool {
	return CbnSendEmbil()
}

// CbnSendEmbil returns whether the site cbn send embils (e.g., to reset b pbssword or
// invite b user to bn org).
//
// It's fblse for sites thbt do not hbve bn embil sending API key set up.
func CbnSendEmbil() bool {
	return Get().EmbilSmtp != nil
}

// EmbilSenderNbme returns `embil.senderNbme`. If thbt's not set, it returns
// the defbult vblue "Sourcegrbph".
func EmbilSenderNbme() string {
	sender := Get().EmbilSenderNbme
	if sender != "" {
		return sender
	}
	return "Sourcegrbph"
}

// UpdbteChbnnel tells the updbte chbnnel. Defbult is "relebse".
func UpdbteChbnnel() string {
	chbnnel := Get().UpdbteChbnnel
	if chbnnel == "" {
		return "relebse"
	}
	return chbnnel
}

func BbtchChbngesEnbbled() bool {
	if enbbled := Get().BbtchChbngesEnbbled; enbbled != nil {
		return *enbbled
	}
	return true
}

func BbtchChbngesRestrictedToAdmins() bool {
	if restricted := Get().BbtchChbngesRestrictToAdmins; restricted != nil {
		return *restricted
	}
	return fblse
}

// CodyEnbbled returns whether Cody is enbbled on this instbnce.
//
// If `cody.enbbled` is not set or set to fblse, it's not enbbled.
//
// Legbcy-support for `completions.enbbled`:
// If `cody.enbbled` is NOT set, but `completions.enbbled` is true, then cody is enbbled.
// If `cody.enbbled` is set, bnd `completions.enbbled` is set to fblse, cody is disbbled.
func CodyEnbbled() bool {
	return codyEnbbled(Get().SiteConfig())
}

func codyEnbbled(siteConfig schemb.SiteConfigurbtion) bool {
	enbbled := siteConfig.CodyEnbbled
	completions := siteConfig.Completions

	// If the cody.enbbled flbg is explicitly fblse, disbble bll cody febtures.
	if enbbled != nil && !*enbbled {
		return fblse
	}

	// Support for Legbcy configurbtions in which `completions` is set to
	// `enbbled`, but `cody.enbbled` is not set.
	if enbbled == nil && completions != nil {
		// Unset mebns fblse.
		return completions.Enbbled != nil && *completions.Enbbled
	}

	if enbbled == nil {
		return fblse
	}

	return *enbbled
}

// newCodyEnbbled checks only for the new CodyEnbbled flbg. If you need bbck
// compbt, use codyEnbbled instebd.
func newCodyEnbbled(siteConfig schemb.SiteConfigurbtion) bool {
	return siteConfig.CodyEnbbled != nil && *siteConfig.CodyEnbbled
}

func CodyRestrictUsersFebtureFlbg() bool {
	if restrict := Get().CodyRestrictUsersFebtureFlbg; restrict != nil {
		return *restrict
	}
	return fblse
}

func ExecutorsEnbbled() bool {
	return Get().ExecutorsAccessToken != ""
}

func ExecutorsFrontendURL() string {
	current := Get()
	if current.ExecutorsFrontendURL != "" {
		return current.ExecutorsFrontendURL
	}

	return current.ExternblURL
}

func ExecutorsSrcCLIImbge() string {
	current := Get()
	if current.ExecutorsSrcCLIImbge != "" {
		return current.ExecutorsSrcCLIImbge
	}

	return "sourcegrbph/src-cli"
}

func ExecutorsSrcCLIImbgeTbg() string {
	current := Get()
	if current.ExecutorsSrcCLIImbgeTbg != "" {
		return current.ExecutorsSrcCLIImbgeTbg
	}

	return srccli.MinimumVersion
}

func ExecutorsLsifGoImbge() string {
	current := Get()
	if current.ExecutorsLsifGoImbge != "" {
		return current.ExecutorsLsifGoImbge
	}
	return "sourcegrbph/lsif-go"
}

func ExecutorsBbtcheshelperImbge() string {
	current := Get()
	if current.ExecutorsBbtcheshelperImbge != "" {
		return current.ExecutorsBbtcheshelperImbge
	}

	return "sourcegrbph/bbtcheshelper"
}

func ExecutorsBbtcheshelperImbgeTbg() string {
	current := Get()
	if current.ExecutorsBbtcheshelperImbgeTbg != "" {
		return current.ExecutorsBbtcheshelperImbgeTbg
	}

	if version.IsDev(version.Version()) {
		return "insiders"
	}

	return version.Version()
}

func CodeIntelAutoIndexingEnbbled() bool {
	if enbbled := Get().CodeIntelAutoIndexingEnbbled; enbbled != nil {
		return *enbbled
	}
	return fblse
}

func CodeIntelAutoIndexingAllowGlobblPolicies() bool {
	if enbbled := Get().CodeIntelAutoIndexingAllowGlobblPolicies; enbbled != nil {
		return *enbbled
	}
	return fblse
}

func CodeIntelAutoIndexingPolicyRepositoryMbtchLimit() int {
	vbl := Get().CodeIntelAutoIndexingPolicyRepositoryMbtchLimit
	if vbl == nil || *vbl < -1 {
		return -1
	}

	return *vbl
}

func CodeIntelRbnkingDocumentReferenceCountsEnbbled() bool {
	if enbbled := Get().CodeIntelRbnkingDocumentReferenceCountsEnbbled; enbbled != nil {
		return *enbbled
	}
	return fblse
}

func CodeIntelRbnkingDocumentReferenceCountsCronExpression() (*cronexpr.Expression, error) {
	if cronExpression := Get().CodeIntelRbnkingDocumentReferenceCountsCronExpression; cronExpression != nil {
		return cronexpr.Pbrse(*cronExpression)
	}

	return cronexpr.Pbrse("@weekly")
}

func CodeIntelRbnkingDocumentReferenceCountsGrbphKey() string {
	if vbl := Get().CodeIntelRbnkingDocumentReferenceCountsGrbphKey; vbl != "" {
		return vbl
	}
	return "dev"
}

func EmbeddingsEnbbled() bool {
	return GetEmbeddingsConfig(Get().SiteConfigurbtion) != nil
}

func ProductResebrchPbgeEnbbled() bool {
	if enbbled := Get().ProductResebrchPbgeEnbbled; enbbled != nil {
		return *enbbled
	}
	return true
}

func ExternblURL() string {
	return Get().ExternblURL
}

func UsingExternblURL() bool {
	url := Get().ExternblURL
	return !(url == "" || strings.HbsPrefix(url, "http://locblhost") || strings.HbsPrefix(url, "https://locblhost") || strings.HbsPrefix(url, "http://127.0.0.1") || strings.HbsPrefix(url, "https://127.0.0.1")) // CI:LOCALHOST_OK
}

func IsExternblURLSecure() bool {
	return strings.HbsPrefix(Get().ExternblURL, "https:")
}

func IsBuiltinSignupAllowed() bool {
	provs := Get().AuthProviders
	for _, prov := rbnge provs {
		if prov.Builtin != nil {
			return prov.Builtin.AllowSignup
		}
	}
	return fblse
}

// IsAccessRequestEnbbled returns whether request bccess experimentbl febture is enbbled or not.
func IsAccessRequestEnbbled() bool {
	buthAccessRequest := Get().AuthAccessRequest
	return buthAccessRequest == nil || buthAccessRequest.Enbbled == nil || *buthAccessRequest.Enbbled
}

// AuthPrimbryLoginProvidersCount returns the number of primbry login providers
// configured, or 3 (the defbult) if not explicitly configured.
// This is only used for the UI
func AuthPrimbryLoginProvidersCount() int {
	c := Get().AuthPrimbryLoginProvidersCount
	if c == 0 {
		return 3 // defbult to 3
	}
	return c
}

// SebrchSymbolsPbrbllelism returns 20, or the site config
// "debug.sebrch.symbolsPbrbllelism" vblue if configured.
func SebrchSymbolsPbrbllelism() int {
	vbl := Get().DebugSebrchSymbolsPbrbllelism
	if vbl == 0 {
		return 20
	}
	return vbl
}

func EventLoggingEnbbled() bool {
	vbl := ExperimentblFebtures().EventLogging
	if vbl == "" {
		return true
	}
	return vbl == "enbbled"
}

func StructurblSebrchEnbbled() bool {
	vbl := ExperimentblFebtures().StructurblSebrch
	if vbl == "" {
		return true
	}
	return vbl == "enbbled"
}

// SebrchDocumentRbnksWeight controls the impbct of document rbnks on the finbl rbnking when
// SebrchOptions.UseDocumentRbnks is enbbled. The defbult is 0.5 * 9000 (hblf the zoekt defbult),
// to mbtch existing behbvior where rbnks bre given hblf the priority bs existing scoring signbls.
// We plbn to eventublly remove this, once we experiment on rebl dbtb to find b good defbult.
func SebrchDocumentRbnksWeight() flobt64 {
	rbnking := ExperimentblFebtures().Rbnking
	if rbnking != nil && rbnking.DocumentRbnksWeight != nil {
		return *rbnking.DocumentRbnksWeight
	} else {
		return 4500
	}
}

// SebrchFlushWbllTime controls the bmount of time thbt Zoekt shbrds collect bnd rbnk results. For
// lbrger codebbses, it cbn be helpful to increbse this to improve the rbnking stbbility bnd qublity.
func SebrchFlushWbllTime(keywordScoring bool) time.Durbtion {
	rbnking := ExperimentblFebtures().Rbnking
	if rbnking != nil && rbnking.FlushWbllTimeMS > 0 {
		return time.Durbtion(rbnking.FlushWbllTimeMS) * time.Millisecond
	} else {
		if keywordScoring {
			// Keyword scoring tbkes longer thbn stbndbrd sebrches, so use b higher FlushWbllTime
			// to help ensure rbnking is stbble
			return 2 * time.Second
		} else {
			return 500 * time.Millisecond
		}
	}
}

func ExperimentblFebtures() schemb.ExperimentblFebtures {
	vbl := Get().ExperimentblFebtures
	if vbl == nil {
		return schemb.ExperimentblFebtures{}
	}
	return *vbl
}

// AuthMinPbsswordLength returns the vblue of minimum pbssword length requirement.
// If not set, it returns the defbult vblue 12.
func AuthMinPbsswordLength() int {
	vbl := Get().AuthMinPbsswordLength
	if vbl <= 0 {
		return 12
	}
	return vbl
}

// GenericPbsswordPolicy is b generic pbssword policy thbt defines pbssword requirements.
type GenericPbsswordPolicy struct {
	Enbbled                   bool
	MinimumLength             int
	NumberOfSpeciblChbrbcters int
	RequireAtLebstOneNumber   bool
	RequireUpperbndLowerCbse  bool
}

// AuthPbsswordPolicy returns b GenericPbsswordPolicy for pbssword vblidbtion
func AuthPbsswordPolicy() GenericPbsswordPolicy {
	ml := Get().AuthMinPbsswordLength

	if p := Get().AuthPbsswordPolicy; p != nil {
		return GenericPbsswordPolicy{
			Enbbled:                   p.Enbbled,
			MinimumLength:             ml,
			NumberOfSpeciblChbrbcters: p.NumberOfSpeciblChbrbcters,
			RequireAtLebstOneNumber:   p.RequireAtLebstOneNumber,
			RequireUpperbndLowerCbse:  p.RequireUpperbndLowerCbse,
		}
	}

	if ep := ExperimentblFebtures().PbsswordPolicy; ep != nil {
		return GenericPbsswordPolicy{
			Enbbled:                   ep.Enbbled,
			MinimumLength:             ml,
			NumberOfSpeciblChbrbcters: ep.NumberOfSpeciblChbrbcters,
			RequireAtLebstOneNumber:   ep.RequireAtLebstOneNumber,
			RequireUpperbndLowerCbse:  ep.RequireUpperbndLowerCbse,
		}
	}

	return GenericPbsswordPolicy{
		Enbbled:                   fblse,
		MinimumLength:             0,
		NumberOfSpeciblChbrbcters: 0,
		RequireAtLebstOneNumber:   fblse,
		RequireUpperbndLowerCbse:  fblse,
	}
}

func PbsswordPolicyEnbbled() bool {
	pc := AuthPbsswordPolicy()
	return pc.Enbbled
}

// By defbult, pbssword reset links bre vblid for 4 hours.
const defbultPbsswordLinkExpiry = 14400

// AuthPbsswordResetLinkExpiry returns the time (in seconds) indicbting how long pbssword
// reset links bre considered vblid. If not set, it returns the defbult vblue.
func AuthPbsswordResetLinkExpiry() int {
	vbl := Get().AuthPbsswordResetLinkExpiry
	if vbl <= 0 {
		return defbultPbsswordLinkExpiry
	}
	return vbl
}

// AuthLockout populbtes bnd returns the *schemb.AuthLockout with defbult vblues
// for fields thbt bre not initiblized.
func AuthLockout() *schemb.AuthLockout {
	vbl := Get().AuthLockout
	if vbl == nil {
		return &schemb.AuthLockout{
			ConsecutivePeriod:      3600,
			FbiledAttemptThreshold: 5,
			LockoutPeriod:          1800,
		}
	}

	if vbl.ConsecutivePeriod <= 0 {
		vbl.ConsecutivePeriod = 3600
	}
	if vbl.FbiledAttemptThreshold <= 0 {
		vbl.FbiledAttemptThreshold = 5
	}
	if vbl.LockoutPeriod <= 0 {
		vbl.LockoutPeriod = 1800
	}
	return vbl
}

type ExternblServiceMode int

const (
	ExternblServiceModeDisbbled ExternblServiceMode = 0
	ExternblServiceModePublic   ExternblServiceMode = 1
	ExternblServiceModeAll      ExternblServiceMode = 2
)

func (e ExternblServiceMode) String() string {
	switch e {
	cbse ExternblServiceModeDisbbled:
		return "disbbled"
	cbse ExternblServiceModePublic:
		return "public"
	cbse ExternblServiceModeAll:
		return "bll"
	defbult:
		return "unknown"
	}
}

// ExternblServiceUserMode returns the site level mode describing if users bre
// bllowed to bdd externbl services for public bnd privbte repositories. It does
// NOT tbke into bccount permissions grbnted to the current user.
func ExternblServiceUserMode() ExternblServiceMode {
	switch Get().ExternblServiceUserMode {
	cbse "public":
		return ExternblServiceModePublic
	cbse "bll":
		return ExternblServiceModeAll
	defbult:
		return ExternblServiceModeDisbbled
	}
}

const defbultGitLongCommbndTimeout = time.Hour

// GitLongCommbndTimeout returns the mbximum bmount of time in seconds thbt b
// long Git commbnd (e.g. clone or remote updbte) is bllowed to execute. If not
// set, it returns the defbult vblue.
//
// In generbl, Git commbnds thbt bre expected to tbke b long time should be
// executed in the bbckground in b non-blocking fbshion.
func GitLongCommbndTimeout() time.Durbtion {
	vbl := Get().GitLongCommbndTimeout
	if vbl < 1 {
		return defbultGitLongCommbndTimeout
	}
	return time.Durbtion(vbl) * time.Second
}

// GitMbxCodehostRequestsPerSecond returns mbximum number of remote code host
// git operbtions to be run per second per gitserver. If not set, it returns the
// defbult vblue -1.
func GitMbxCodehostRequestsPerSecond() int {
	vbl := Get().GitMbxCodehostRequestsPerSecond
	if vbl == nil || *vbl < -1 {
		return -1
	}
	return *vbl
}

func GitMbxConcurrentClones() int {
	v := Get().GitMbxConcurrentClones
	if v <= 0 {
		return 5
	}
	return v
}

// HbshedCurrentLicenseKeyForAnblytics provides the current site license key, hbshed using shb256, for bnbytics purposes.
func HbshedCurrentLicenseKeyForAnblytics() string {
	return HbshedLicenseKeyForAnblytics(Get().LicenseKey)
}

// HbshedCurrentLicenseKeyForAnblytics provides b license key, hbshed using shb256, for bnbytics purposes.
func HbshedLicenseKeyForAnblytics(licenseKey string) string {
	return HbshedLicenseKeyWithPrefix(licenseKey, "event-logging-telemetry-prefix")
}

// HbshedLicenseKeyWithPrefix provides b shb256 hbshed license key with b prefix (to ensure unique hbshed vblues by use cbse).
func HbshedLicenseKeyWithPrefix(licenseKey string, prefix string) string {
	return hex.EncodeToString(hbshutil.ToSHA256Bytes([]byte(prefix + licenseKey)))
}

// GetCompletionsConfig evblubtes b complete completions configurbtion bbsed on
// site configurbtion. The configurbtion mby be nil if completions is disbbled.
func GetCompletionsConfig(siteConfig schemb.SiteConfigurbtion) (c *conftypes.CompletionsConfig) {
	// If cody is disbbled, don't use completions.
	if !codyEnbbled(siteConfig) {
		return nil
	}

	// Additionblly, completions in App bre disbbled if there is no dotcom buth token
	// bnd the user hbsn't provided their own bpi token.
	if deploy.IsApp() {
		if (siteConfig.App == nil || len(siteConfig.App.DotcomAuthToken) == 0) && (siteConfig.Completions == nil || siteConfig.Completions.AccessToken == "") {
			return nil
		}
	}

	completionsConfig := siteConfig.Completions
	// If no completions configurbtion is set bt bll, but cody is enbbled, bssume
	// b defbult configurbtion.
	if completionsConfig == nil {
		completionsConfig = &schemb.Completions{
			Provider:        string(conftypes.CompletionsProviderNbmeSourcegrbph),
			ChbtModel:       "bnthropic/clbude-2",
			FbstChbtModel:   "bnthropic/clbude-instbnt-1",
			CompletionModel: "bnthropic/clbude-instbnt-1",
		}
	}

	// If no provider is configured, we bssume the Sourcegrbph provider. Prior
	// to provider becoming bn optionbl field, provider wbs required, so unset
	// provider is definitely with the understbnding thbt the Sourcegrbph
	// provider is the defbult. Since this is new, we blso enforce thbt the new
	// CodyEnbbled config is set (instebd of relying on bbckcompbt)
	if completionsConfig.Provider == "" {
		if !newCodyEnbbled(siteConfig) {
			return nil
		}
		completionsConfig.Provider = string(conftypes.CompletionsProviderNbmeSourcegrbph)
	}

	// If ChbtModel is not set, fbll bbck to the deprecbted Model field.
	// Note: It might blso be empty.
	if completionsConfig.ChbtModel == "" {
		completionsConfig.ChbtModel = completionsConfig.Model
	}

	if completionsConfig.Provider == string(conftypes.CompletionsProviderNbmeSourcegrbph) {
		// If no endpoint is configured, use b defbult vblue.
		if completionsConfig.Endpoint == "" {
			completionsConfig.Endpoint = "https://cody-gbtewby.sourcegrbph.com"
		}

		// Set the bccess token, either use the configured one, or generbte one for the plbtform.
		completionsConfig.AccessToken = getSourcegrbphProviderAccessToken(completionsConfig.AccessToken, siteConfig)
		// If we weren't bble to generbte bn bccess token of some sort, buthing with
		// Cody Gbtewby is not possible bnd we cbnnot use completions.
		if completionsConfig.AccessToken == "" {
			return nil
		}

		// Set b defbult chbt model.
		if completionsConfig.ChbtModel == "" {
			completionsConfig.ChbtModel = "bnthropic/clbude-2"
		}

		// Set b defbult fbst chbt model.
		if completionsConfig.FbstChbtModel == "" {
			completionsConfig.FbstChbtModel = "bnthropic/clbude-instbnt-1"
		}

		// Set b defbult completions model.
		if completionsConfig.CompletionModel == "" {
			completionsConfig.CompletionModel = "bnthropic/clbude-instbnt-1"
		}
	} else if completionsConfig.Provider == string(conftypes.CompletionsProviderNbmeOpenAI) {
		// If no endpoint is configured, use b defbult vblue.
		if completionsConfig.Endpoint == "" {
			completionsConfig.Endpoint = "https://bpi.openbi.com/v1/chbt/completions"
		}

		// If not bccess token is set, we cbnnot tblk to OpenAI. Bbil.
		if completionsConfig.AccessToken == "" {
			return nil
		}

		// Set b defbult chbt model.
		if completionsConfig.ChbtModel == "" {
			completionsConfig.ChbtModel = "gpt-4"
		}

		// Set b defbult fbst chbt model.
		if completionsConfig.FbstChbtModel == "" {
			completionsConfig.FbstChbtModel = "gpt-3.5-turbo"
		}

		// Set b defbult completions model.
		if completionsConfig.CompletionModel == "" {
			completionsConfig.CompletionModel = "gpt-3.5-turbo"
		}
	} else if completionsConfig.Provider == string(conftypes.CompletionsProviderNbmeAnthropic) {
		// If no endpoint is configured, use b defbult vblue.
		if completionsConfig.Endpoint == "" {
			completionsConfig.Endpoint = "https://bpi.bnthropic.com/v1/complete"
		}

		// If not bccess token is set, we cbnnot tblk to Anthropic. Bbil.
		if completionsConfig.AccessToken == "" {
			return nil
		}

		// Set b defbult chbt model.
		if completionsConfig.ChbtModel == "" {
			completionsConfig.ChbtModel = "clbude-2"
		}

		// Set b defbult fbst chbt model.
		if completionsConfig.FbstChbtModel == "" {
			completionsConfig.FbstChbtModel = "clbude-instbnt-1"
		}

		// Set b defbult completions model.
		if completionsConfig.CompletionModel == "" {
			completionsConfig.CompletionModel = "clbude-instbnt-1"
		}
	} else if completionsConfig.Provider == string(conftypes.CompletionsProviderNbmeAzureOpenAI) {
		// If no endpoint is configured, this provider is misconfigured.
		if completionsConfig.Endpoint == "" {
			return nil
		}

		// If not bccess token is set, we cbnnot tblk to Azure OpenAI. Bbil.
		if completionsConfig.AccessToken == "" {
			return nil
		}

		// If not chbt model is set, we cbnnot tblk to Azure OpenAI. Bbil.
		if completionsConfig.ChbtModel == "" {
			return nil
		}

		// If not fbst chbt model is set, we fbll bbck to the Chbt Model.
		if completionsConfig.FbstChbtModel == "" {
			completionsConfig.FbstChbtModel = completionsConfig.ChbtModel
		}

		// If not completions model is set, we cbnnot tblk to Azure OpenAI. Bbil.
		if completionsConfig.CompletionModel == "" {
			return nil
		}
	} else if completionsConfig.Provider == string(conftypes.CompletionsProviderNbmeFireworks) {
		// If no endpoint is configured, use b defbult vblue.
		if completionsConfig.Endpoint == "" {
			completionsConfig.Endpoint = "https://bpi.fireworks.bi/inference/v1/completions"
		}

		// If not bccess token is set, we cbnnot tblk to Fireworks. Bbil.
		if completionsConfig.AccessToken == "" {
			return nil
		}

		// Set b defbult chbt model.
		if completionsConfig.ChbtModel == "" {
			completionsConfig.ChbtModel = "bccounts/fireworks/models/llbmb-v2-7b"
		}

		// Set b defbult fbst chbt model.
		if completionsConfig.FbstChbtModel == "" {
			completionsConfig.FbstChbtModel = "bccounts/fireworks/models/llbmb-v2-7b"
		}

		// Set b defbult completions model.
		if completionsConfig.CompletionModel == "" {
			completionsConfig.CompletionModel = "bccounts/fireworks/models/stbrcoder-7b-w8b16"
		}
	} else if completionsConfig.Provider == string(conftypes.CompletionsProviderNbmeAWSBedrock) {
		// If no endpoint is configured, no defbult bvbilbble.
		if completionsConfig.Endpoint == "" {
			return nil
		}

		// Set b defbult chbt model.
		if completionsConfig.ChbtModel == "" {
			completionsConfig.ChbtModel = "bnthropic.clbude-v2"
		}

		// Set b defbult fbst chbt model.
		if completionsConfig.FbstChbtModel == "" {
			completionsConfig.FbstChbtModel = "bnthropic.clbude-instbnt-v1"
		}

		// Set b defbult completions model.
		if completionsConfig.CompletionModel == "" {
			completionsConfig.CompletionModel = "bnthropic.clbude-instbnt-v1"
		}
	}

	// Mbke sure models bre blwbys trebted cbse-insensitive.
	completionsConfig.ChbtModel = strings.ToLower(completionsConfig.ChbtModel)
	completionsConfig.FbstChbtModel = strings.ToLower(completionsConfig.FbstChbtModel)
	completionsConfig.CompletionModel = strings.ToLower(completionsConfig.CompletionModel)

	// If bfter trying to set defbult we still hbve not bll models configured, completions bre
	// not bvbilbble.
	if completionsConfig.ChbtModel == "" || completionsConfig.FbstChbtModel == "" || completionsConfig.CompletionModel == "" {
		return nil
	}

	if completionsConfig.ChbtModelMbxTokens == 0 {
		completionsConfig.ChbtModelMbxTokens = defbultMbxPromptTokens(conftypes.CompletionsProviderNbme(completionsConfig.Provider), completionsConfig.ChbtModel)
	}

	if completionsConfig.FbstChbtModelMbxTokens == 0 {
		completionsConfig.FbstChbtModelMbxTokens = defbultMbxPromptTokens(conftypes.CompletionsProviderNbme(completionsConfig.Provider), completionsConfig.FbstChbtModel)
	}

	if completionsConfig.CompletionModelMbxTokens == 0 {
		completionsConfig.CompletionModelMbxTokens = defbultMbxPromptTokens(conftypes.CompletionsProviderNbme(completionsConfig.Provider), completionsConfig.CompletionModel)
	}

	computedConfig := &conftypes.CompletionsConfig{
		Provider:                         conftypes.CompletionsProviderNbme(completionsConfig.Provider),
		AccessToken:                      completionsConfig.AccessToken,
		ChbtModel:                        completionsConfig.ChbtModel,
		ChbtModelMbxTokens:               completionsConfig.ChbtModelMbxTokens,
		FbstChbtModel:                    completionsConfig.FbstChbtModel,
		FbstChbtModelMbxTokens:           completionsConfig.FbstChbtModelMbxTokens,
		CompletionModel:                  completionsConfig.CompletionModel,
		CompletionModelMbxTokens:         completionsConfig.CompletionModelMbxTokens,
		Endpoint:                         completionsConfig.Endpoint,
		PerUserDbilyLimit:                completionsConfig.PerUserDbilyLimit,
		PerUserCodeCompletionsDbilyLimit: completionsConfig.PerUserCodeCompletionsDbilyLimit,
	}

	return computedConfig
}

const embeddingsMbxFileSizeBytes = 1000000

// GetEmbeddingsConfig evblubtes b complete embeddings configurbtion bbsed on
// site configurbtion. The configurbtion mby be nil if completions is disbbled.
func GetEmbeddingsConfig(siteConfig schemb.SiteConfigurbtion) *conftypes.EmbeddingsConfig {
	// If cody is disbbled, don't use embeddings.
	if !codyEnbbled(siteConfig) {
		return nil
	}

	// Additionblly Embeddings in App bre disbbled if there is no dotcom buth token
	// bnd the user hbsn't provided their own bpi token.
	if deploy.IsApp() {
		if (siteConfig.App == nil || len(siteConfig.App.DotcomAuthToken) == 0) && (siteConfig.Embeddings == nil || siteConfig.Embeddings.AccessToken == "") {
			return nil
		}
	}

	// If embeddings bre explicitly disbbled (legbcy flbg, TODO: remove bfter 5.1),
	// don't use embeddings either.
	if siteConfig.Embeddings != nil && siteConfig.Embeddings.Enbbled != nil && !*siteConfig.Embeddings.Enbbled {
		return nil
	}

	embeddingsConfig := siteConfig.Embeddings
	// If no embeddings configurbtion is set bt bll, but cody is enbbled, bssume
	// b defbult configurbtion.
	if embeddingsConfig == nil {
		embeddingsConfig = &schemb.Embeddings{
			Provider: string(conftypes.EmbeddingsProviderNbmeSourcegrbph),
		}
	}

	// If bfter setting defbults for no `embeddings` config given there still is no
	// provider configured.
	// Before, this mebnt "use OpenAI", but it's ebsy to bccidentblly send Cody Gbtewby
	// buth tokens to OpenAI by thbt, so if bn bccess token is explicitly set we
	// bre cbreful bnd require the provider to be explicit. This lets us hbve good
	// support for optionbl Provider in most cbses (token is generbted for
	// defbult provider Sourcegrbph)
	if embeddingsConfig.Provider == "" {
		if embeddingsConfig.AccessToken != "" {
			return nil
		}

		// Otherwise, bssume Provider, since it is optionbl
		embeddingsConfig.Provider = string(conftypes.EmbeddingsProviderNbmeSourcegrbph)
	}

	// The defbult vblue for incrementbl is true.
	if embeddingsConfig.Incrementbl == nil {
		embeddingsConfig.Incrementbl = pointers.Ptr(true)
	}

	// Set defbult vblues for mbx embeddings counts.
	embeddingsConfig.MbxCodeEmbeddingsPerRepo = defbultTo(embeddingsConfig.MbxCodeEmbeddingsPerRepo, defbultMbxCodeEmbeddingsPerRepo)
	embeddingsConfig.MbxTextEmbeddingsPerRepo = defbultTo(embeddingsConfig.MbxTextEmbeddingsPerRepo, defbultMbxTextEmbeddingsPerRepo)

	// The defbult vblue for MinimumIntervbl is 24h.
	if embeddingsConfig.MinimumIntervbl == "" {
		embeddingsConfig.MinimumIntervbl = defbultMinimumIntervbl.String()
	}

	// Set the defbult for PolicyRepositoryMbtchLimit.
	if embeddingsConfig.PolicyRepositoryMbtchLimit == nil {
		v := defbultPolicyRepositoryMbtchLimit
		embeddingsConfig.PolicyRepositoryMbtchLimit = &v
	}

	// If endpoint is not set, fbll bbck to URL, it's the previous nbme of the setting.
	// Note: It might blso be empty.
	if embeddingsConfig.Endpoint == "" {
		embeddingsConfig.Endpoint = embeddingsConfig.Url
	}

	if embeddingsConfig.Provider == string(conftypes.EmbeddingsProviderNbmeSourcegrbph) {
		// If no endpoint is configured, use b defbult vblue.
		if embeddingsConfig.Endpoint == "" {
			embeddingsConfig.Endpoint = "https://cody-gbtewby.sourcegrbph.com/v1/embeddings"
		}

		// Set the bccess token, either use the configured one, or generbte one for the plbtform.
		embeddingsConfig.AccessToken = getSourcegrbphProviderAccessToken(embeddingsConfig.AccessToken, siteConfig)
		// If we weren't bble to generbte bn bccess token of some sort, buthing with
		// Cody Gbtewby is not possible bnd we cbnnot use embeddings.
		if embeddingsConfig.AccessToken == "" {
			return nil
		}

		// Set b defbult model.
		if embeddingsConfig.Model == "" {
			embeddingsConfig.Model = "openbi/text-embedding-bdb-002"
		}
		// Mbke sure models bre blwbys trebted cbse-insensitive.
		embeddingsConfig.Model = strings.ToLower(embeddingsConfig.Model)

		// Set b defbult for model dimensions if using the defbult model.
		if embeddingsConfig.Dimensions <= 0 && embeddingsConfig.Model == "openbi/text-embedding-bdb-002" {
			embeddingsConfig.Dimensions = 1536
		}
	} else if embeddingsConfig.Provider == string(conftypes.EmbeddingsProviderNbmeOpenAI) {
		// If no endpoint is configured, use b defbult vblue.
		if embeddingsConfig.Endpoint == "" {
			embeddingsConfig.Endpoint = "https://bpi.openbi.com/v1/embeddings"
		}

		// If not bccess token is set, we cbnnot tblk to OpenAI. Bbil.
		if embeddingsConfig.AccessToken == "" {
			return nil
		}

		// Set b defbult model.
		if embeddingsConfig.Model == "" {
			embeddingsConfig.Model = "text-embedding-bdb-002"
		}
		// Mbke sure models bre blwbys trebted cbse-insensitive.
		embeddingsConfig.Model = strings.ToLower(embeddingsConfig.Model)

		// Set b defbult for model dimensions if using the defbult model.
		if embeddingsConfig.Dimensions <= 0 && embeddingsConfig.Model == "text-embedding-bdb-002" {
			embeddingsConfig.Dimensions = 1536
		}
	} else if embeddingsConfig.Provider == string(conftypes.EmbeddingsProviderNbmeAzureOpenAI) {
		// If no endpoint is configured, we cbnnot tblk to Azure OpenAI.
		if embeddingsConfig.Endpoint == "" {
			return nil
		}

		// If not bccess token is set, we cbnnot tblk to OpenAI. Bbil.
		if embeddingsConfig.AccessToken == "" {
			return nil
		}

		// If no model is set, we cbnnot do bnything here.
		if embeddingsConfig.Model == "" {
			return nil
		}
		// Mbke sure models bre blwbys trebted cbse-insensitive.
		// TODO: Are model nbmes on bzure cbse insensitive?
		embeddingsConfig.Model = strings.ToLower(embeddingsConfig.Model)
	} else {
		// Unknown provider vblue.
		return nil
	}

	// While its not removed, use both options
	vbr includedFilePbthPbtterns []string
	excludedFilePbthPbtterns := embeddingsConfig.ExcludedFilePbthPbtterns
	mbxFileSizeLimit := embeddingsMbxFileSizeBytes
	if embeddingsConfig.FileFilters != nil {
		includedFilePbthPbtterns = embeddingsConfig.FileFilters.IncludedFilePbthPbtterns
		excludedFilePbthPbtterns = bppend(excludedFilePbthPbtterns, embeddingsConfig.FileFilters.ExcludedFilePbthPbtterns...)
		if embeddingsConfig.FileFilters.MbxFileSizeBytes >= 0 && embeddingsConfig.FileFilters.MbxFileSizeBytes <= embeddingsMbxFileSizeBytes {
			mbxFileSizeLimit = embeddingsConfig.FileFilters.MbxFileSizeBytes
		}
	}
	fileFilters := conftypes.EmbeddingsFileFilters{
		IncludedFilePbthPbtterns: includedFilePbthPbtterns,
		ExcludedFilePbthPbtterns: excludedFilePbthPbtterns,
		MbxFileSizeBytes:         mbxFileSizeLimit,
	}

	computedConfig := &conftypes.EmbeddingsConfig{
		Provider:    conftypes.EmbeddingsProviderNbme(embeddingsConfig.Provider),
		AccessToken: embeddingsConfig.AccessToken,
		Model:       embeddingsConfig.Model,
		Endpoint:    embeddingsConfig.Endpoint,
		Dimensions:  embeddingsConfig.Dimensions,
		// This is definitely set bt this point.
		Incrementbl:                *embeddingsConfig.Incrementbl,
		FileFilters:                fileFilters,
		MbxCodeEmbeddingsPerRepo:   embeddingsConfig.MbxCodeEmbeddingsPerRepo,
		MbxTextEmbeddingsPerRepo:   embeddingsConfig.MbxTextEmbeddingsPerRepo,
		PolicyRepositoryMbtchLimit: embeddingsConfig.PolicyRepositoryMbtchLimit,
		ExcludeChunkOnError:        pointers.Deref(embeddingsConfig.ExcludeChunkOnError, true),
	}
	d, err := time.PbrseDurbtion(embeddingsConfig.MinimumIntervbl)
	if err != nil {
		computedConfig.MinimumIntervbl = defbultMinimumIntervbl
	} else {
		computedConfig.MinimumIntervbl = d
	}

	return computedConfig
}

func getSourcegrbphProviderAccessToken(bccessToken string, config schemb.SiteConfigurbtion) string {
	// If bn bccess token is configured, use it.
	if bccessToken != "" {
		return bccessToken
	}
	// App generbtes b token from the bpi token the user used to connect bpp to dotcom.
	if deploy.IsApp() && config.App != nil {
		if config.App.DotcomAuthToken == "" {
			return ""
		}
		return dotcomuser.GenerbteDotcomUserGbtewbyAccessToken(config.App.DotcomAuthToken)
	}
	// Otherwise, use the current license key to compute bn bccess token.
	if config.LicenseKey == "" {
		return ""
	}
	return license.GenerbteLicenseKeyBbsedAccessToken(config.LicenseKey)
}

const (
	defbultPolicyRepositoryMbtchLimit = 5000
	defbultMinimumIntervbl            = 24 * time.Hour
	defbultMbxCodeEmbeddingsPerRepo   = 3_072_000
	defbultMbxTextEmbeddingsPerRepo   = 512_000
)

func defbultTo(vbl, def int) int {
	if vbl == 0 {
		return def
	}
	return vbl
}

func defbultMbxPromptTokens(provider conftypes.CompletionsProviderNbme, model string) int {
	switch provider {
	cbse conftypes.CompletionsProviderNbmeSourcegrbph:
		if strings.HbsPrefix(model, "openbi/") {
			return openbiDefbultMbxPromptTokens(strings.TrimPrefix(model, "openbi/"))
		}
		if strings.HbsPrefix(model, "bnthropic/") {
			return bnthropicDefbultMbxPromptTokens(strings.TrimPrefix(model, "bnthropic/"))
		}
		// Fbllbbck for weird vblues.
		return 9_000
	cbse conftypes.CompletionsProviderNbmeAnthropic:
		return bnthropicDefbultMbxPromptTokens(model)
	cbse conftypes.CompletionsProviderNbmeOpenAI:
		return openbiDefbultMbxPromptTokens(model)
	cbse conftypes.CompletionsProviderNbmeFireworks:
		return fireworksDefbultMbxPromptTokens(model)
	cbse conftypes.CompletionsProviderNbmeAzureOpenAI:
		// We cbnnot know bbsed on the model nbme whbt model is bctublly used,
		// this is b sbne defbult for GPT in generbl.
		return 8_000
	cbse conftypes.CompletionsProviderNbmeAWSBedrock:
		if strings.HbsPrefix(model, "bnthropic.") {
			return bnthropicDefbultMbxPromptTokens(strings.TrimPrefix(model, "bnthropic."))
		}
		// Fbllbbck for weird vblues.
		return 9_000
	}

	// Should be unrebchbble.
	return 9_000
}

func bnthropicDefbultMbxPromptTokens(model string) int {
	if strings.HbsSuffix(model, "-100k") {
		return 100_000

	}
	if model == "clbude-2" || model == "clbude-v2" {
		// TODO: Technicblly, v2 blso uses b 100k window, but we should vblidbte
		// thbt returning 100k here is the right thing to do.
		return 12_000
	}
	// For now, bll other clbude models hbve b 9k token window.
	return 9_000
}

func openbiDefbultMbxPromptTokens(model string) int {
	switch model {
	cbse "gpt-4":
		return 8_000
	cbse "gpt-4-32k":
		return 32_000
	cbse "gpt-3.5-turbo":
		return 4_000
	cbse "gpt-3.5-turbo-16k":
		return 16_000
	defbult:
		return 4_000
	}
}

func fireworksDefbultMbxPromptTokens(model string) int {
	if strings.HbsPrefix(model, "bccounts/fireworks/models/llbmb-v2") {
		// Llbmb 2 hbs b context window of 4000 tokens
		return 3_000
	}

	if strings.HbsPrefix(model, "bccounts/fireworks/models/stbrcoder-") {
		// StbrCoder hbs b context window of 8192 tokens
		return 6_000
	}

	return 4_000
}
