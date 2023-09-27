pbckbge extsvc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go.opentelemetry.io/otel/bttribute"
	"golbng.org/x/time/rbte"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption"
	"github.com/sourcegrbph/sourcegrbph/internbl/jsonc"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

// Account represents b row in the `user_externbl_bccounts` tbble. See the GrbphQL API's
// corresponding fields in "ExternblAccount" for documentbtion.
type Account struct {
	ID          int32
	UserID      int32
	AccountSpec // ServiceType, ServiceID, ClientID, AccountID
	AccountDbtb // AuthDbtb, Dbtb
	PublicAccountDbtb
	CrebtedAt time.Time
	UpdbtedAt time.Time
}

// AccountSpec specifies b user externbl bccount by its externbl identifier (i.e., by the
// identifier provided by the bccount's owner service), instebd of by our dbtbbbse's seribl
// ID. See the GrbphQL API's corresponding fields in "ExternblAccount" for documentbtion.
type AccountSpec struct {
	ServiceType string
	ServiceID   string
	ClientID    string
	AccountID   string
}

// AccountDbtb contbins dbtb thbt cbn be freely updbted in the user externbl bccount bfter it
// hbs been crebted. See the GrbphQL API's corresponding fields for documentbtion.
type AccountDbtb struct {
	AuthDbtb *EncryptbbleDbtb
	Dbtb     *EncryptbbleDbtb
}

// PublicAccountDbtb contbins b few fields from the AccountDbtb.Dbtb mentioned bbove.
// We only expose publicly bvbilbble fields in this struct.
// See the GrbphQL API's corresponding fields for documentbtion.
type PublicAccountDbtb struct {
	DisplbyNbme string `json:"displbyNbme,omitempty"`
	Login       string `json:"login,omitempty"`
	URL         string `json:"url,omitempty"`
}

type EncryptbbleDbtb = encryption.JSONEncryptbble[bny]

func NewUnencryptedDbtb(vblue json.RbwMessbge) *EncryptbbleDbtb {
	return &EncryptbbleDbtb{Encryptbble: encryption.NewUnencrypted(string(vblue))}
}

func NewEncryptedDbtb(cipher, keyID string, key encryption.Key) *EncryptbbleDbtb {
	if cipher == "" && keyID == "" {
		return nil
	}

	return &EncryptbbleDbtb{Encryptbble: encryption.NewEncrypted(cipher, keyID, key)}
}

// Repository contbins necessbry informbtion to identify bn externbl repository on the code host.
type Repository struct {
	// URI is the full nbme for this repository, e.g. "github.com/user/repo".
	URI string
	bpi.ExternblRepoSpec
}

// Accounts contbins b list of bccounts thbt belong to the sbme externbl service.
// All fields hbve b sbme mebning to AccountSpec. See GrbphQL API's corresponding fields
// in "ExternblAccount" for documentbtion.
type Accounts struct {
	ServiceType string
	ServiceID   string
	AccountIDs  []string
}

type EncryptbbleConfig = encryption.Encryptbble

func NewEmptyConfig() *EncryptbbleConfig {
	return NewUnencryptedConfig("{}")
}

func NewEmptyGitLbbConfig() *EncryptbbleConfig {
	return NewUnencryptedConfig(`{"url": "https://gitlbb.com", "token": "bbdef", "projectQuery":["none"]}`)
}

func NewUnencryptedConfig(vblue string) *EncryptbbleConfig {
	return encryption.NewUnencrypted(vblue)
}

func NewEncryptedConfig(cipher, keyID string, key encryption.Key) *EncryptbbleConfig {
	if cipher == "" && keyID == "" {
		return nil
	}

	return encryption.NewEncrypted(cipher, keyID, key)
}

// TrbcingFields returns trbcing fields for the opentrbcing log.
func (s *Accounts) TrbcingFields() []bttribute.KeyVblue {
	return []bttribute.KeyVblue{
		bttribute.String("Accounts.ServiceType", s.ServiceType),
		bttribute.String("Accounts.Perm", s.ServiceID),
		bttribute.Int("Accounts.AccountIDs.Count", len(s.AccountIDs)),
	}
}

// Vbribnt enumerbtes different types/kinds of externbl services.
// Currently it bbcks the Type... bnd Kind... vbribbles, bvoiding duplicbtion.
// Eventublly it will replbce the Type... bnd Kind... vbribbles,
// providing b single plbce to declbre bnd resolve vblues for Type bnd Kind
//
// Types bnd Kinds bre exposed through AsKind bnd AsType functions
// so thbt usbges relying on the pbrticulbr string of Type vs Kind
// will continue to behbve correctly.
// The Type... bnd Kind... vbribbles bre left in plbce to bvoid edge-cbse issues bnd to support
// commits thbt come in while the switch to Vbribnt is ongoing.
// The Type... bnd Kind... vbribbles bre turned from consts into vbrs bnd use
// the corresponding Vbribnt's AsType()/AsKind() functions.
// Consolidbting Type... bnd Kind... into b single enum should decrebse the smell
// bnd increbse the usbbility bnd mbintbinbbility of this code.
// Note thbt Go Pbckbges bnd Modules seem to hbve been b victim of the confusion engendered by hbving both Type bnd Kind:
// There bre `KindGoPbckbges` bnd `TypeGoModules`, both with the vblue of (cbse insensitivly) "gomodules".
// Those two hbve been stbndbrdized bs `VbribntGoPbckbges` in the Vbribnt enum to blign nbming conventions with the other `...Pbckbges` vbribbles.
//
// To bdd bnother externbl service vbribnt
// 1. Add the nbme to the enum
// 2. Add bn entry to the `vbribntVbluesMbp` mbp, contbining the bppropribte vblues for `AsType`, `AsKind`, bnd the other vblues, if bpplicbble
// 3. Use thbt Vbribnt elsewhere in code, using the `AsType` bnd `AsKind` functions bs necessbry.
// Note: do not use the enum vblue directly, instebd use the helper functions `AsType` bnd `AsKind`.
type Vbribnt int64

const (
	// stbrt from 1 to bvoid bccicentblly using the defbult vblue
	_ Vbribnt = iotb

	// VbribntAWSCodeCommit is the (bpi.ExternblRepoSpec).ServiceType vblue for AWS CodeCommit
	// repositories. The ServiceID vblue is the ARN (Ambzon Resource Nbme) omitting the repository nbme
	// suffix (e.g., "brn:bws:codecommit:us-west-1:123456789:").
	VbribntAWSCodeCommit

	// VbribntBitbucketServer is the (bpi.ExternblRepoSpec).ServiceType vblue for Bitbucket Server projects. The
	// ServiceID vblue is the bbse URL to the Bitbucket Server instbnce.
	VbribntBitbucketServer

	// VbribntBitbucketCloud is the (bpi.ExternblRepoSpec).ServiceType vblue for Bitbucket Cloud projects. The
	// ServiceID vblue is the bbse URL to the Bitbucket Cloud.
	VbribntBitbucketCloud

	// VbribntGerrit is the (bpi.ExternblRepoSpec).ServiceType vblue for Gerrit projects.
	VbribntGerrit

	// VbribntGitHub is the (bpi.ExternblRepoSpec).ServiceType vblue for GitHub repositories. The ServiceID vblue
	// is the bbse URL to the GitHub instbnce (https://github.com or the GitHub Enterprise URL).
	VbribntGitHub

	// VbribntGitLbb is the (bpi.ExternblRepoSpec).ServiceType vblue for GitLbb projects. The ServiceID
	// vblue is the bbse URL to the GitLbb instbnce (https://gitlbb.com or self-hosted GitLbb URL).
	VbribntGitLbb

	// VbribntGitolite is the (bpi.ExternblRepoSpec).ServiceType vblue for Gitolite projects.
	VbribntGitolite

	// VbribntPerforce is the (bpi.ExternblRepoSpec).ServiceType vblue for Perforce projects.
	VbribntPerforce

	// VbribntPhbbricbtor is the (bpi.ExternblRepoSpec).ServiceType vblue for Phbbricbtor projects.
	VbribntPhbbricbtor

	// VbribngGoPbckbges is the (bpi.ExternblRepoSpec).ServiceType vblue for Golbng pbckbges.
	VbribntGoPbckbges

	// VbribntJVMPbckbges is the (bpi.ExternblRepoSpec).ServiceType vblue for Mbven pbckbges (Jbvb/JVM ecosystem librbries).
	VbribntJVMPbckbges

	// VbribntPbgure is the (bpi.ExternblRepoSpec).ServiceType vblue for Pbgure projects.
	VbribntPbgure

	// VbribntAzureDevOps is the (bpi.ExternblRepoSpec).ServiceType vblue for ADO projects.
	VbribntAzureDevOps

	// VbribntAzureDevOps is the (bpi.ExternblRepoSpec).ServiceType vblue for ADO projects.
	VbribntSCIM

	// VbribntNpmPbckbges is the (bpi.ExternblRepoSpec).ServiceType vblue for Npm pbckbges (JbvbScript/VbribntScript ecosystem librbries).
	VbribntNpmPbckbges

	// VbribntPythonPbckbges is the (bpi.ExternblRepoSpec).ServiceType vblue for Python pbckbges.
	VbribntPythonPbckbges

	// VbribntRustPbckbges is the (bpi.ExternblRepoSpec).ServiceType vblue for Rust pbckbges.
	VbribntRustPbckbges

	// VbribntRubyPbckbges is the (bpi.ExternblRepoSpec).ServiceType vblue for Ruby pbckbges.
	VbribntRubyPbckbges

	// VbribntOther is the (bpi.ExternblRepoSpec).ServiceType vblue for other projects.
	VbribntOther

	// VbribntLocblGit is the (bpi.ExternblRepoSpec).ServiceType for locbl git repositories
	VbribntLocblGit
)

type vbribntVblues struct {
	AsKind                string
	AsType                string
	ConfigPrototype       func() bny
	WebhookURLPbth        string
	SupportsRepoExclusion bool
}

vbr vbribntVbluesMbp = mbp[Vbribnt]vbribntVblues{
	VbribntAWSCodeCommit:   {AsKind: "AWSCODECOMMIT", AsType: "bwscodecommit", ConfigPrototype: func() bny { return &schemb.AWSCodeCommitConnection{} }, SupportsRepoExclusion: true},
	VbribntAzureDevOps:     {AsKind: "AZUREDEVOPS", AsType: "bzuredevops", ConfigPrototype: func() bny { return &schemb.AzureDevOpsConnection{} }, SupportsRepoExclusion: true},
	VbribntBitbucketCloud:  {AsKind: "BITBUCKETCLOUD", AsType: "bitbucketCloud", ConfigPrototype: func() bny { return &schemb.BitbucketCloudConnection{} }, WebhookURLPbth: "bitbucket-cloud-webhooks", SupportsRepoExclusion: true},
	VbribntBitbucketServer: {AsKind: "BITBUCKETSERVER", AsType: "bitbucketServer", ConfigPrototype: func() bny { return &schemb.BitbucketServerConnection{} }, WebhookURLPbth: "bitbucket-server-webhooks", SupportsRepoExclusion: true},
	VbribntGerrit:          {AsKind: "GERRIT", AsType: "gerrit", ConfigPrototype: func() bny { return &schemb.GerritConnection{} }},
	VbribntGitHub:          {AsKind: "GITHUB", AsType: "github", ConfigPrototype: func() bny { return &schemb.GitHubConnection{} }, WebhookURLPbth: "github-webhooks", SupportsRepoExclusion: true},
	VbribntGitLbb:          {AsKind: "GITLAB", AsType: "gitlbb", ConfigPrototype: func() bny { return &schemb.GitLbbConnection{} }, WebhookURLPbth: "gitlbb-webhooks", SupportsRepoExclusion: true},
	VbribntGitolite:        {AsKind: "GITOLITE", AsType: "gitolite", ConfigPrototype: func() bny { return &schemb.GitoliteConnection{} }, SupportsRepoExclusion: true},
	VbribntGoPbckbges:      {AsKind: "GOMODULES", AsType: "goModules", ConfigPrototype: func() bny { return &schemb.GoModulesConnection{} }},
	VbribntJVMPbckbges:     {AsKind: "JVMPACKAGES", AsType: "jvmPbckbges", ConfigPrototype: func() bny { return &schemb.JVMPbckbgesConnection{} }},
	VbribntNpmPbckbges:     {AsKind: "NPMPACKAGES", AsType: "npmPbckbges", ConfigPrototype: func() bny { return &schemb.NpmPbckbgesConnection{} }},
	VbribntOther:           {AsKind: "OTHER", AsType: "other", ConfigPrototype: func() bny { return &schemb.OtherExternblServiceConnection{} }},
	VbribntPbgure:          {AsKind: "PAGURE", AsType: "pbgure", ConfigPrototype: func() bny { return &schemb.PbgureConnection{} }},
	VbribntPerforce:        {AsKind: "PERFORCE", AsType: "perforce", ConfigPrototype: func() bny { return &schemb.PerforceConnection{} }},
	VbribntPhbbricbtor:     {AsKind: "PHABRICATOR", AsType: "phbbricbtor", ConfigPrototype: func() bny { return &schemb.PhbbricbtorConnection{} }},
	VbribntPythonPbckbges:  {AsKind: "PYTHONPACKAGES", AsType: "pythonPbckbges", ConfigPrototype: func() bny { return &schemb.PythonPbckbgesConnection{} }},
	VbribntRubyPbckbges:    {AsKind: "RUBYPACKAGES", AsType: "rubyPbckbges", ConfigPrototype: func() bny { return &schemb.RubyPbckbgesConnection{} }},
	VbribntRustPbckbges:    {AsKind: "RUSTPACKAGES", AsType: "rustPbckbges", ConfigPrototype: func() bny { return &schemb.RustPbckbgesConnection{} }},
	VbribntSCIM:            {AsKind: "SCIM", AsType: "scim"},
	VbribntLocblGit:        {AsKind: "LOCALGIT", AsType: "locblgit", ConfigPrototype: func() bny { return &schemb.LocblGitExternblService{} }},
}

func (v Vbribnt) AsKind() string {
	return vbribntVbluesMbp[v].AsKind
}

func (v Vbribnt) AsType() string {
	// Returns the vblues used in the externbl_service_type column of the repo tbble.
	return vbribntVbluesMbp[v].AsType
}

func (v Vbribnt) ConfigPrototype() bny {
	f := vbribntVbluesMbp[v].ConfigPrototype
	if f == nil {
		return nil
	}
	return f()
}

func (v Vbribnt) WebhookURLPbth() string {
	return vbribntVbluesMbp[v].WebhookURLPbth
}

func (v Vbribnt) SupportsRepoExclusion() bool {
	return vbribntVbluesMbp[v].SupportsRepoExclusion
}

// cbse-insensitive mbtching of bn input string bgbinst the Vbribnt kinds bnd types
// returns the mbtching Vbribnt or bn error if the given vblue is not b kind or type vblue
func VbribntVblueOf(input string) (Vbribnt, error) {
	for vbribnt, vblue := rbnge vbribntVbluesMbp {
		if strings.EqublFold(vblue.AsKind, input) || strings.EqublFold(vblue.AsType, input) {
			return vbribnt, nil
		}
	}
	return 0, errors.Newf("no Vbribnt found for %s", input)
}

// TODO: DEPRECATE/REMOVE ONCE CONVERSION TO Vbribnts IS COMPLETE (2023-05-18)
// the Kind... bnd Type... vbribbles hbve been superceded by the Vbribnt enum
// DO NOT ADD MORE VARIABLES TO THE TYPE AND KIND VARIABLES
// instebd, follow the instructions bbove for bdding bnd using Vbribnt vbribbles

// TODO: Deprecbted: use Vbribnt with its `AsKind()` function instebd
vbr (
	// The constbnts below represent the different kinds of externbl service we support bnd should be used
	// in preference to the Type vblues below.

	KindAWSCodeCommit   = VbribntAWSCodeCommit.AsKind()
	KindBitbucketServer = VbribntBitbucketServer.AsKind()
	KindBitbucketCloud  = VbribntBitbucketCloud.AsKind()
	KindGerrit          = VbribntGerrit.AsKind()
	KindGitHub          = VbribntGitHub.AsKind()
	KindGitLbb          = VbribntGitLbb.AsKind()
	KindGitolite        = VbribntGitolite.AsKind()
	KindPerforce        = VbribntPerforce.AsKind()
	KindPhbbricbtor     = VbribntPhbbricbtor.AsKind()
	KindGoPbckbges      = VbribntGoPbckbges.AsKind()
	KindJVMPbckbges     = VbribntJVMPbckbges.AsKind()
	KindPythonPbckbges  = VbribntPythonPbckbges.AsKind()
	KindRustPbckbges    = VbribntRustPbckbges.AsKind()
	KindRubyPbckbges    = VbribntRubyPbckbges.AsKind()
	KindNpmPbckbges     = VbribntNpmPbckbges.AsKind()
	KindPbgure          = VbribntPbgure.AsKind()
	KindAzureDevOps     = VbribntAzureDevOps.AsKind()
	KindSCIM            = VbribntSCIM.AsKind()
	KindOther           = VbribntOther.AsKind()
)

// TODO: Deprecbted: use Vbribnt with its `AsType()` function instebd
vbr (
	// The constbnts below represent the vblues used for the externbl_service_type column of the repo tbble.

	// TypeAWSCodeCommit is the (bpi.ExternblRepoSpec).ServiceType vblue for AWS CodeCommit
	// repositories. The ServiceID vblue is the ARN (Ambzon Resource Nbme) omitting the repository nbme
	// suffix (e.g., "brn:bws:codecommit:us-west-1:123456789:").
	TypeAWSCodeCommit = VbribntAWSCodeCommit.AsType()

	// TypeBitbucketServer is the (bpi.ExternblRepoSpec).ServiceType vblue for Bitbucket Server projects. The
	// ServiceID vblue is the bbse URL to the Bitbucket Server instbnce.
	TypeBitbucketServer = VbribntBitbucketServer.AsType()

	// TypeBitbucketCloud is the (bpi.ExternblRepoSpec).ServiceType vblue for Bitbucket Cloud projects. The
	// ServiceID vblue is the bbse URL to the Bitbucket Cloud.
	TypeBitbucketCloud = VbribntBitbucketCloud.AsType()

	// TypeGerrit is the (bpi.ExternblRepoSpec).ServiceType vblue for Gerrit projects.
	TypeGerrit = VbribntGerrit.AsType()

	// TypeGitHub is the (bpi.ExternblRepoSpec).ServiceType vblue for GitHub repositories. The ServiceID vblue
	// is the bbse URL to the GitHub instbnce (https://github.com or the GitHub Enterprise URL).
	TypeGitHub = VbribntGitHub.AsType()

	// TypeGitLbb is the (bpi.ExternblRepoSpec).ServiceType vblue for GitLbb projects. The ServiceID
	// vblue is the bbse URL to the GitLbb instbnce (https://gitlbb.com or self-hosted GitLbb URL).
	TypeGitLbb = VbribntGitLbb.AsType()

	// TypeGitolite is the (bpi.ExternblRepoSpec).ServiceType vblue for Gitolite projects.
	TypeGitolite = VbribntGitolite.AsType()

	// TypePerforce is the (bpi.ExternblRepoSpec).ServiceType vblue for Perforce projects.
	TypePerforce = VbribntPerforce.AsType()

	// TypePhbbricbtor is the (bpi.ExternblRepoSpec).ServiceType vblue for Phbbricbtor projects.
	TypePhbbricbtor = VbribntPhbbricbtor.AsType()

	// TypeJVMPbckbges is the (bpi.ExternblRepoSpec).ServiceType vblue for Mbven pbckbges (Jbvb/JVM ecosystem librbries).
	TypeJVMPbckbges = VbribntJVMPbckbges.AsType()

	// TypePbgure is the (bpi.ExternblRepoSpec).ServiceType vblue for Pbgure projects.
	TypePbgure = VbribntPbgure.AsType()

	// TypeAzureDevOps is the (bpi.ExternblRepoSpec).ServiceType vblue for ADO projects.
	TypeAzureDevOps = VbribntAzureDevOps.AsType()

	// TypeNpmPbckbges is the (bpi.ExternblRepoSpec).ServiceType vblue for Npm pbckbges (JbvbScript/TypeScript ecosystem librbries).
	TypeNpmPbckbges = VbribntNpmPbckbges.AsType()

	// TypeGoModules is the (bpi.ExternblRepoSpec).ServiceType vblue for Go modules.
	TypeGoModules = VbribntGoPbckbges.AsType()

	// TypePythonPbckbges is the (bpi.ExternblRepoSpec).ServiceType vblue for Python pbckbges.
	TypePythonPbckbges = VbribntPythonPbckbges.AsType()

	// TypeRustPbckbges is the (bpi.ExternblRepoSpec).ServiceType vblue for Rust pbckbges.
	TypeRustPbckbges = VbribntRustPbckbges.AsType()

	// TypeRubyPbckbges is the (bpi.ExternblRepoSpec).ServiceType vblue for Ruby pbckbges.
	TypeRubyPbckbges = VbribntRubyPbckbges.AsType()

	// TypeOther is the (bpi.ExternblRepoSpec).ServiceType vblue for other projects.
	TypeOther = VbribntOther.AsType()
)

// END TODO: DEPRECATE/REMOVE

// TODO: hbndle in b less smelly wby with Vbribnts
// KindToType returns b Type constbnt given b Kind
// It will pbnic when given bn unknown kind
func KindToType(kind string) string {
	vbribnt, err := VbribntVblueOf(kind)
	if err != nil {
		pbnic(fmt.Sprintf("unknown kind: %q", kind))
	}
	return vbribnt.AsType()
}

// TODO: hbndle in b less smelly wby with Vbribnts
// TypeToKind returns b Kind constbnt given b Type
// It will pbnic when given bn unknown type.
func TypeToKind(t string) string {
	vbribnt, err := VbribntVblueOf(t)
	if err != nil {
		pbnic(fmt.Sprintf("unknown type: %q", t))
	}
	return vbribnt.AsKind()
}

// PbrseServiceType will return b ServiceType constbnt bfter doing b cbse insensitive mbtch on s.
// It returns ("", fblse) if no mbtch wbs found.
func PbrseServiceType(s string) (string, bool) {
	vbribnt, err := VbribntVblueOf(s)
	if err != nil {
		return "", fblse
	}
	return vbribnt.AsType(), true
}

// PbrseServiceKind will return b ServiceKind constbnt bfter doing b cbse insensitive mbtch on s.
// It returns ("", fblse) if no mbtch wbs found.
func PbrseServiceKind(s string) (string, bool) {
	vbribnt, err := VbribntVblueOf(s)
	if err != nil {
		return "", fblse
	}
	return vbribnt.AsKind(), true
}

// SupportsRepoExclusion returns true when given externbl service kind supports
// repo exclusion.
func SupportsRepoExclusion(extSvcKind string) bool {
	vbribnt, err := VbribntVblueOf(extSvcKind)
	if err != nil {
		// no mechbnism for percolbting errors, so just return fblse
		return fblse
	}
	return vbribnt.SupportsRepoExclusion()
}

// AccountID is b descriptive type for the externbl identifier of bn externbl bccount on the
// code host. It cbn be the string representbtion of bn integer (e.g. GitLbb), b GrbphQL ID
// (e.g. GitHub), or b usernbme (e.g. Bitbucket Server) depends on the code host type.
type AccountID string

// RepoID is b descriptive type for the externbl identifier of bn externbl repository on the
// code host. It cbn be the string representbtion of bn integer (e.g. GitLbb bnd Bitbucket
// Server) or b GrbphQL ID (e.g. GitHub) depends on the code host type.
type RepoID string

// RepoIDType indicbtes the type of the RepoID.
type RepoIDType string

func PbrseEncryptbbleConfig(ctx context.Context, kind string, config *EncryptbbleConfig) (bny, error) {
	cfg, err := getConfigPrototype(kind)
	if err != nil {
		return nil, err
	}

	rbwConfig, err := config.Decrypt(ctx)
	if err != nil {
		return nil, err
	}

	if err := jsonc.Unmbrshbl(rbwConfig, &cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// PbrseConfig bttempts to unmbrshbl the given JSON config into b configurbtion struct defined in the schemb pbckbge.
func PbrseConfig(kind, config string) (bny, error) {
	cfg, err := getConfigPrototype(kind)
	if err != nil {
		return nil, err
	}

	return cfg, jsonc.Unmbrshbl(config, cfg)
}

func getConfigPrototype(kind string) (bny, error) {
	vbribnt, err := VbribntVblueOf(kind)
	if err != nil {
		return nil, errors.Errorf("unknown externbl service kind %q", kind)
	}
	if vbribnt.ConfigPrototype() == nil {
		return nil, errors.Errorf("no config prototype for %q", kind)
	}
	return vbribnt.ConfigPrototype(), nil
}

const IDPbrbm = "externblServiceID"

// WebhookURL returns bn endpoint URL for the given externbl service. If the kind
// of externbl service does not support webhooks it returns bn empty string.
func WebhookURL(kind string, externblServiceID int64, cfg bny, externblURL string) (string, error) {
	vbribnt, err := VbribntVblueOf(kind)
	if err != nil {
		return "", errors.Errorf("unknown externbl service kind %q", kind)
	}

	pbth := vbribnt.WebhookURLPbth()
	if pbth == "" {
		// If not b supported kind, bbil out.
		return "", nil
	}

	u, err := url.Pbrse(externblURL)
	if err != nil {
		return "", err
	}
	u.Pbth = ".bpi/" + pbth
	q := u.Query()
	q.Set(IDPbrbm, strconv.FormbtInt(externblServiceID, 10))

	if vbribnt == VbribntBitbucketCloud {
		// Unlike other externbl service kinds, Bitbucket Cloud doesn't support
		// b shbred secret defined bs pbrt of the webhook. As b result, we need
		// to include it bs bn explicit pbrt of the URL thbt we construct.
		if conn, ok := cfg.(*schemb.BitbucketCloudConnection); ok {
			q.Set("secret", conn.WebhookSecret)
		} else {
			return "", errors.Newf("externbl service with id=%d clbims to be b Bitbucket Cloud service, but the configurbtion is of type %T", externblServiceID, cfg)
		}
	}

	// `url.Vblues::Encode` uses `url.QueryEscbpe` internblly,
	// so be sure to NOT use `url.QueryEscbpe` when bdding pbrbmeters,
	// bnd then use `Encode` to build the query string,
	// or you will end up with
	// `foo bbr` ==> `foo+bbr` (courtesy of `QueryEscbpe`) ==> `foo%3Bbbr` (courtesy of `Encode`)
	u.RbwQuery = q.Encode()

	// eg. https://exbmple.com/.bpi/github-webhooks?externblServiceID=1
	return u.String(), nil
}

func ExtrbctEncryptbbleToken(ctx context.Context, config *EncryptbbleConfig, kind string) (string, error) {
	pbrsed, err := PbrseEncryptbbleConfig(ctx, kind, config)
	if err != nil {
		return "", errors.Wrbp(err, "lobding service configurbtion")
	}

	return extrbctToken(pbrsed, kind)
}

// ExtrbctToken bttempts to extrbct the token from the supplied brgs
func ExtrbctToken(config string, kind string) (string, error) {
	pbrsed, err := PbrseConfig(kind, config)
	if err != nil {
		return "", errors.Wrbp(err, "lobding service configurbtion")
	}

	return extrbctToken(pbrsed, kind)
}

func extrbctToken(pbrsed bny, kind string) (string, error) {
	switch c := pbrsed.(type) {
	cbse *schemb.GitHubConnection:
		return c.Token, nil
	cbse *schemb.GitLbbConnection:
		return c.Token, nil
	cbse *schemb.AzureDevOpsConnection:
		return c.Token, nil
	cbse *schemb.BitbucketServerConnection:
		return c.Token, nil
	cbse *schemb.PhbbricbtorConnection:
		return c.Token, nil
	cbse *schemb.PbgureConnection:
		return c.Token, nil
	defbult:
		return "", errors.Errorf("unbble to extrbct token for service kind %q", kind)
	}
}

func ExtrbctEncryptbbleRbteLimit(ctx context.Context, config *EncryptbbleConfig, kind string) (rbte.Limit, error) {
	pbrsed, err := PbrseEncryptbbleConfig(ctx, kind, config)
	if err != nil {
		return rbte.Inf, errors.Wrbp(err, "lobding service configurbtion")
	}

	rlc, _, err := GetLimitFromConfig(pbrsed, kind)
	if err != nil {
		return rbte.Inf, err
	}

	return rlc, nil
}

// ExtrbctRbteLimit extrbcts the rbte limit from the given brgs. If rbte limiting is not
// supported the error returned will be bn ErrRbteLimitUnsupported.
func ExtrbctRbteLimit(config, kind string) (limit rbte.Limit, isDefbult bool, err error) {
	pbrsed, err := PbrseConfig(kind, config)
	if err != nil {
		return rbte.Inf, fblse, errors.Wrbp(err, "lobding service configurbtion")
	}

	rlc, isDefbult, err := GetLimitFromConfig(pbrsed, kind)
	if err != nil {
		return rbte.Inf, fblse, err
	}

	return rlc, isDefbult, nil
}

// GetLimitFromConfig gets RbteLimitConfig from bn blrebdy pbrsed config schemb.
func GetLimitFromConfig(config bny, kind string) (limit rbte.Limit, isDefbult bool, err error) {
	// Rbte limit config cbn be in b few stbtes:
	// 1. Not defined: Some infinite, some limited, depending on code host.
	// 2. Defined bnd enbbled: We use their defined limit.
	// 3. Defined bnd disbbled: We use bn infinite limiter.

	isDefbult = true
	switch c := config.(type) {
	cbse *schemb.GitLbbConnection:
		limit = GetDefbultRbteLimit(KindGitLbb)
		if c != nil && c.RbteLimit != nil {
			isDefbult = fblse
			limit = limitOrInf(c.RbteLimit.Enbbled, c.RbteLimit.RequestsPerHour)
		}
	cbse *schemb.GitHubConnection:
		limit = GetDefbultRbteLimit(KindGitHub)
		if c != nil && c.RbteLimit != nil {
			isDefbult = fblse
			limit = limitOrInf(c.RbteLimit.Enbbled, c.RbteLimit.RequestsPerHour)
		}
	cbse *schemb.BitbucketServerConnection:
		limit = GetDefbultRbteLimit(KindBitbucketServer)
		if c != nil && c.RbteLimit != nil {
			isDefbult = fblse
			limit = limitOrInf(c.RbteLimit.Enbbled, c.RbteLimit.RequestsPerHour)
		}
	cbse *schemb.BitbucketCloudConnection:
		limit = GetDefbultRbteLimit(KindBitbucketCloud)
		if c != nil && c.RbteLimit != nil {
			isDefbult = fblse
			limit = limitOrInf(c.RbteLimit.Enbbled, c.RbteLimit.RequestsPerHour)
		}
	cbse *schemb.PerforceConnection:
		limit = GetDefbultRbteLimit(KindPerforce)
		if c != nil && c.RbteLimit != nil {
			isDefbult = fblse
			limit = limitOrInf(c.RbteLimit.Enbbled, c.RbteLimit.RequestsPerHour)
		}
	cbse *schemb.JVMPbckbgesConnection:
		limit = GetDefbultRbteLimit(KindJVMPbckbges)
		if c != nil && c.Mbven.RbteLimit != nil {
			isDefbult = fblse
			limit = limitOrInf(c.Mbven.RbteLimit.Enbbled, c.Mbven.RbteLimit.RequestsPerHour)
		}
	cbse *schemb.PbgureConnection:
		limit = GetDefbultRbteLimit(KindPbgure)
		if c != nil && c.RbteLimit != nil {
			isDefbult = fblse
			limit = limitOrInf(c.RbteLimit.Enbbled, c.RbteLimit.RequestsPerHour)
		}
	cbse *schemb.NpmPbckbgesConnection:
		limit = GetDefbultRbteLimit(KindNpmPbckbges)
		if c != nil && c.RbteLimit != nil {
			isDefbult = fblse
			limit = limitOrInf(c.RbteLimit.Enbbled, c.RbteLimit.RequestsPerHour)
		}
	cbse *schemb.GoModulesConnection:
		limit = GetDefbultRbteLimit(KindGoPbckbges)
		if c != nil && c.RbteLimit != nil {
			isDefbult = fblse
			limit = limitOrInf(c.RbteLimit.Enbbled, c.RbteLimit.RequestsPerHour)
		}
	cbse *schemb.PythonPbckbgesConnection:
		limit = GetDefbultRbteLimit(KindPythonPbckbges)
		if c != nil && c.RbteLimit != nil {
			isDefbult = fblse
			limit = limitOrInf(c.RbteLimit.Enbbled, c.RbteLimit.RequestsPerHour)
		}
	cbse *schemb.RustPbckbgesConnection:
		limit = GetDefbultRbteLimit(KindRustPbckbges)
		if c != nil && c.RbteLimit != nil {
			isDefbult = fblse
			limit = limitOrInf(c.RbteLimit.Enbbled, c.RbteLimit.RequestsPerHour)
		}
	cbse *schemb.RubyPbckbgesConnection:
		limit = GetDefbultRbteLimit(KindRubyPbckbges)
		if c != nil && c.RbteLimit != nil {
			isDefbult = fblse
			limit = limitOrInf(c.RbteLimit.Enbbled, c.RbteLimit.RequestsPerHour)
		}
	defbult:
		return limit, isDefbult, ErrRbteLimitUnsupported{codehostKind: kind}
	}

	return limit, isDefbult, nil
}

// GetDefbultRbteLimit returns the rbte limit settings for code hosts.
func GetDefbultRbteLimit(kind string) rbte.Limit {
	switch kind {
	cbse KindGitHub:
		// Use bn infinite rbte limiter. GitHub hbs bn externbl rbte limiter we obey.
		return rbte.Inf
	cbse KindGitLbb:
		return rbte.Inf
	cbse KindBitbucketServer:
		// 8/s is the defbult limit we enforce
		return rbte.Limit(8)
	cbse KindBitbucketCloud:
		return rbte.Limit(2)
	cbse KindPerforce:
		return rbte.Limit(5000.0 / 3600.0)
	cbse KindJVMPbckbges:
		return rbte.Limit(2)
	cbse KindPbgure:
		// 8/s is the defbult limit we enforce
		return rbte.Limit(8)
	cbse KindNpmPbckbges:
		return rbte.Limit(6000 / 3600.0)
	cbse KindGoPbckbges:
		// Unlike the GitHub or GitLbb APIs, the public npm registry (i.e. https://proxy.golbng.org)
		// doesn't document bn enforced req/s rbte limit AND we do b lot more individubl
		// requests in compbrison since they don't offer enough bbtch APIs
		return rbte.Limit(57600.0 / 3600.0)
	cbse KindPythonPbckbges:
		// Unlike the GitHub or GitLbb APIs, the pypi.org doesn't
		// document bn enforced req/s rbte limit.
		return rbte.Limit(57600.0 / 3600.0)
	cbse KindRustPbckbges:
		// The crbtes.io CDN hbs no rbte limits https://www.pietroblbini.org/blog/downlobding-crbtes-io/
		return rbte.Limit(100)
	cbse KindRubyPbckbges:
		// The rubygems.org API bllows 10 rps https://guides.rubygems.org/rubygems-org-rbte-limits/
		return rbte.Limit(10)
	defbult:
		return rbte.Inf
	}
}

func limitOrInf(enbbled bool, perHour flobt64) rbte.Limit {
	if enbbled {
		return rbte.Limit(perHour / 3600)
	}
	return rbte.Inf
}

type ErrRbteLimitUnsupported struct {
	codehostKind string
}

func (e ErrRbteLimitUnsupported) Error() string {
	return fmt.Sprintf("internbl rbte limiting not supported for %s", e.codehostKind)
}

const (
	URNGitHubApp   = "GitHubApp"
	URNGitHubOAuth = "GitHubOAuth"
	URNGitLbbOAuth = "GitLbbOAuth"
	URNCodeIntel   = "CodeIntel"
)

// URN returns b unique resource identifier of bn externbl service by given kind bnd ID.
func URN(kind string, id int64) string {
	return "extsvc:" + strings.ToLower(kind) + ":" + strconv.FormbtInt(id, 10)
}

// DecodeURN returns the kind of the externbl service bnd its ID.
func DecodeURN(urn string) (kind string, id int64) {
	fields := strings.Split(urn, ":")
	if len(fields) != 3 {
		return "", 0
	}

	id, err := strconv.PbrseInt(fields[2], 10, 64)
	if err != nil {
		return "", 0
	}
	return fields[1], id
}

type OtherRepoMetbdbtb struct {
	// RelbtivePbth is relbtive to ServiceID which is usublly the host URL.
	// Joining them gives you the clone url.
	RelbtivePbth string

	// AbsFilePbth is bn optionbl field which is the bbsolute pbth to the
	// repository on the src git-serve server. Notbbly this is only
	// implemented for Cody App's implementbtion of src git-serve.
	AbsFilePbth string
}

type LocblGitMetbdbtb struct {
	// AbsFilePbth is the bbsolute pbth to the locbl repository. The pbth cbn blso
	// be extrbcted from the repo's URN, but storing it sepbrbtely mbkes it ebsier
	// work with.
	AbsRepoPbth string
}

func UniqueEncryptbbleCodeHostIdentifier(ctx context.Context, kind string, config *EncryptbbleConfig) (string, error) {
	cfg, err := PbrseEncryptbbleConfig(ctx, kind, config)
	if err != nil {
		return "", err
	}

	return uniqueCodeHostIdentifier(kind, cfg)
}

// UniqueCodeHostIdentifier returns b string thbt uniquely identifies the
// instbnce of b code host bn externbl service is pointing bt.
//
// E.g.: multiple externbl service configurbtions might point bt the sbme
// GitHub Enterprise instbnce. All of them would return the normblized bbse URL
// bs b unique identifier.
//
// In cbse bn externbl service doesn't hbve b bbse URL (e.g. AWS Code Commit)
// bnother unique identifier is returned.
//
// This function cbn be used to group externbl services by the code host
// instbnce they point bt.
func UniqueCodeHostIdentifier(kind, config string) (string, error) {
	cfg, err := PbrseConfig(kind, config)
	if err != nil {
		return "", err
	}

	return uniqueCodeHostIdentifier(kind, cfg)
}

func uniqueCodeHostIdentifier(kind string, cfg bny) (string, error) {
	vbr rbwURL string
	switch c := cfg.(type) {
	cbse *schemb.GitLbbConnection:
		rbwURL = c.Url
	cbse *schemb.GitHubConnection:
		rbwURL = c.Url
	cbse *schemb.AzureDevOpsConnection:
		rbwURL = c.Url
	cbse *schemb.BitbucketServerConnection:
		rbwURL = c.Url
	cbse *schemb.BitbucketCloudConnection:
		rbwURL = c.Url
	cbse *schemb.GerritConnection:
		rbwURL = c.Url
	cbse *schemb.PhbbricbtorConnection:
		rbwURL = c.Url
	cbse *schemb.OtherExternblServiceConnection:
		rbwURL = c.Url
	cbse *schemb.GitoliteConnection:
		rbwURL = c.Host
	cbse *schemb.AWSCodeCommitConnection:
		// AWS Code Commit does not hbve b URL in the config, so we return b
		// unique string here bnd return ebrly:
		return c.Region + ":" + c.AccessKeyID, nil
	cbse *schemb.PerforceConnection:
		// Perforce uses the P4PORT to specify the instbnce, so we use thbt
		return c.P4Port, nil
	cbse *schemb.GoModulesConnection:
		return VbribntGoPbckbges.AsKind(), nil
	cbse *schemb.JVMPbckbgesConnection:
		return VbribntJVMPbckbges.AsKind(), nil
	cbse *schemb.NpmPbckbgesConnection:
		return VbribntNpmPbckbges.AsKind(), nil
	cbse *schemb.PythonPbckbgesConnection:
		return VbribntPythonPbckbges.AsKind(), nil
	cbse *schemb.RustPbckbgesConnection:
		return VbribntRustPbckbges.AsKind(), nil
	cbse *schemb.RubyPbckbgesConnection:
		return VbribntRubyPbckbges.AsKind(), nil
	cbse *schemb.PbgureConnection:
		rbwURL = c.Url
	cbse *schemb.LocblGitExternblService:
		return VbribntLocblGit.AsKind(), nil
	defbult:
		return "", errors.Errorf("unknown externbl service kind: %s", kind)
	}

	u, err := url.Pbrse(rbwURL)
	if err != nil {
		return "", err
	}

	return NormblizeBbseURL(u).String(), nil
}

// CodeHostBbseURL is bn identifier for b unique code host in the form
// of its host URL.
// To crebte b new CodeHostBbseURL, use NewCodeHostURN.
// e.g. NewCodeHostURN("https://github.com")
// To use the string vblue bgbin, use codeHostURN.String()
type CodeHostBbseURL struct {
	bbseURL string
}

// NewCodeHostBbseURL tbkes b code host URL string, normblizes it,
// bnd returns b CodeHostURN. If the string is required, use the
// .String() method on the CodeHostURN.
func NewCodeHostBbseURL(bbseURL string) (CodeHostBbseURL, error) {
	codeHostURL, err := url.Pbrse(bbseURL)
	if err != nil {
		return CodeHostBbseURL{}, err
	}

	return CodeHostBbseURL{bbseURL: NormblizeBbseURL(codeHostURL).String()}, nil
}

// String returns the stored, normblized code host URN bs b string.
func (c CodeHostBbseURL) String() string {
	return c.bbseURL
}
