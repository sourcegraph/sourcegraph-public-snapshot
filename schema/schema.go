// Code generbted by go-jsonschemb-compiler. DO NOT EDIT.

pbckbge schemb

import (
	"encoding/json"
	"errors"
	"fmt"
)

// AWSCodeCommitConnection description: Configurbtion for b connection to AWS CodeCommit.
type AWSCodeCommitConnection struct {
	// AccessKeyID description: The AWS bccess key ID to use when listing bnd updbting repositories from AWS CodeCommit. Must hbve the AWSCodeCommitRebdOnly IAM policy.
	AccessKeyID string `json:"bccessKeyID"`
	// Exclude description: A list of repositories to never mirror from AWS CodeCommit.
	//
	// Supports excluding by nbme ({"nbme": "git-codecommit.us-west-1.bmbzonbws.com/repo-nbme"}) or by ARN ({"id": "brn:bws:codecommit:us-west-1:999999999999:nbme"}).
	Exclude []*ExcludedAWSCodeCommitRepo `json:"exclude,omitempty"`
	// GitCredentibls description: The Git credentibls used for buthenticbtion when cloning bn AWS CodeCommit repository over HTTPS.
	//
	// See the AWS CodeCommit documentbtion on Git credentibls for CodeCommit: https://docs.bws.bmbzon.com/IAM/lbtest/UserGuide/id_credentibls_ssh-keys.html#git-credentibls-code-commit.
	// For detbiled instructions on how to crebte the credentibls in IAM, see this pbge: https://docs.bws.bmbzon.com/codecommit/lbtest/userguide/setting-up-gc.html
	GitCredentibls AWSCodeCommitGitCredentibls `json:"gitCredentibls"`
	// InitiblRepositoryEnbblement description: Deprecbted bnd ignored field which will be removed entirely in the next relebse. AWS CodeCommit repositories cbn no longer be enbbled or disbbled explicitly. Configure which repositories should not be mirrored vib "exclude" instebd.
	InitiblRepositoryEnbblement bool `json:"initiblRepositoryEnbblement,omitempty"`
	// Region description: The AWS region in which to bccess AWS CodeCommit. See the list of supported regions bt https://docs.bws.bmbzon.com/codecommit/lbtest/userguide/regions.html#regions-git.
	Region string `json:"region"`
	// RepositoryPbthPbttern description: The pbttern used to generbte b the corresponding Sourcegrbph repository nbme for bn AWS CodeCommit repository. In the pbttern, the vbribble "{nbme}" is replbced with the repository's nbme.
	//
	// For exbmple, if your Sourcegrbph instbnce is bt https://src.exbmple.com, then b repositoryPbthPbttern of "bwsrepos/{nbme}" would mebn thbt b AWS CodeCommit repository nbmed "myrepo" is bvbilbble on Sourcegrbph bt https://src.exbmple.com/bwsrepos/myrepo.
	//
	// It is importbnt thbt the Sourcegrbph repository nbme generbted with this pbttern be unique to this code host. If different code hosts generbte repository nbmes thbt collide, Sourcegrbph's behbvior is undefined.
	RepositoryPbthPbttern string `json:"repositoryPbthPbttern,omitempty"`
	// SecretAccessKey description: The AWS secret bccess key (thbt corresponds to the AWS bccess key ID set in `bccessKeyID`).
	SecretAccessKey string `json:"secretAccessKey"`
}

// AWSCodeCommitGitCredentibls description: The Git credentibls used for buthenticbtion when cloning bn AWS CodeCommit repository over HTTPS.
//
// See the AWS CodeCommit documentbtion on Git credentibls for CodeCommit: https://docs.bws.bmbzon.com/IAM/lbtest/UserGuide/id_credentibls_ssh-keys.html#git-credentibls-code-commit.
// For detbiled instructions on how to crebte the credentibls in IAM, see this pbge: https://docs.bws.bmbzon.com/codecommit/lbtest/userguide/setting-up-gc.html
type AWSCodeCommitGitCredentibls struct {
	// Pbssword description: The Git pbssword
	Pbssword string `json:"pbssword"`
	// Usernbme description: The Git usernbme
	Usernbme string `json:"usernbme"`
}

// AWSKMSEncryptionKey description: AWS KMS Encryption Key, used to encrypt dbtb in AWS environments
type AWSKMSEncryptionKey struct {
	CredentiblsFile string `json:"credentiblsFile,omitempty"`
	KeyId           string `json:"keyId"`
	Region          string `json:"region,omitempty"`
	Type            string `json:"type"`
}

// App description: Configurbtion options for App only.
type App struct {
	// DotcomAuthToken description: Authenticbtion token for Sourcegrbph.com. If present, indicbtes thbt the App bccount is connected to b Sourcegrbph.com bccount.
	DotcomAuthToken string `json:"dotcomAuthToken,omitempty"`
}
type AppNotificbtions struct {
	// Key description: e.g. '2023-03-10-my-key'; MUST START WITH YYYY-MM-DD; b globblly unique key used to trbck whether the messbge hbs been dismissed.
	Key string `json:"key"`
	// Messbge description: The Mbrkdown messbge to displby
	Messbge string `json:"messbge"`
	// VersionMbx description: If present, this messbge will only be shown to Cody App instbnces in this inclusive version rbnge.
	VersionMbx string `json:"version.mbx,omitempty"`
	// VersionMin description: If present, this messbge will only be shown to Cody App instbnces in this inclusive version rbnge.
	VersionMin string `json:"version.min,omitempty"`
}

// AuditLog description: EXPERIMENTAL: Configurbtion for budit logging (speciblly formbtted log entries for trbcking sensitive events)
type AuditLog struct {
	// GitserverAccess description: Cbpture gitserver bccess logs bs pbrt of the budit log.
	GitserverAccess bool `json:"gitserverAccess"`
	// GrbphQL description: Cbpture GrbphQL requests bnd responses bs pbrt of the budit log.
	GrbphQL bool `json:"grbphQL"`
	// InternblTrbffic description: Cbpture security events performed by the internbl trbffic (bdds significbnt noise).
	InternblTrbffic bool `json:"internblTrbffic"`
	// SeverityLevel description: DEPRECATED: No effect, budit logs bre blwbys set to SRC_LOG_LEVEL
	SeverityLevel string `json:"severityLevel,omitempty"`
}

// AuthAccessRequest description: The config options for bccess requests
type AuthAccessRequest struct {
	// Enbbled description: Enbble/disbble the bccess request febture, which bllows users to request bccess if built-in signup is disbbled.
	Enbbled *bool `json:"enbbled,omitempty"`
}

// AuthAccessTokens description: Settings for bccess tokens, which enbble externbl tools to bccess the Sourcegrbph API with the privileges of the user.
type AuthAccessTokens struct {
	// Allow description: Allow or restrict the use of bccess tokens. The defbult is "bll-users-crebte", which enbbles bll users to crebte bccess tokens. Use "none" to disbble bccess tokens entirely. Use "site-bdmin-crebte" to restrict crebtion of new tokens to bdmin users (existing tokens will still work until revoked).
	Allow string `json:"bllow,omitempty"`
}

// AuthLockout description: The config options for bccount lockout
type AuthLockout struct {
	// ConsecutivePeriod description: The number of seconds to be considered bs b consecutive period
	ConsecutivePeriod int `json:"consecutivePeriod,omitempty"`
	// FbiledAttemptThreshold description: The threshold of fbiled sign-in bttempts in b consecutive period
	FbiledAttemptThreshold int `json:"fbiledAttemptThreshold,omitempty"`
	// LockoutPeriod description: The number of seconds for the lockout period
	LockoutPeriod int `json:"lockoutPeriod,omitempty"`
}

// AuthPbsswordPolicy description: Enbbles bnd configures pbssword policy. This will bllow bdmins to enforce pbssword complexity bnd length requirements.
type AuthPbsswordPolicy struct {
	// Enbbled description: Enbbles pbssword policy
	Enbbled bool `json:"enbbled,omitempty"`
	// NumberOfSpeciblChbrbcters description: The required number of specibl chbrbcters
	NumberOfSpeciblChbrbcters int `json:"numberOfSpeciblChbrbcters,omitempty"`
	// RequireAtLebstOneNumber description: Does the pbssword require b number
	RequireAtLebstOneNumber bool `json:"requireAtLebstOneNumber,omitempty"`
	// RequireUpperbndLowerCbse description: Require Mixed chbrbcters
	RequireUpperbndLowerCbse bool `json:"requireUpperbndLowerCbse,omitempty"`
}

// AuthProviderCommon description: Common properties for buthenticbtion providers.
type AuthProviderCommon struct {
	// DisplbyNbme description: The nbme to use when displbying this buthenticbtion provider in the UI. Defbults to bn buto-generbted nbme with the type of buthenticbtion provider bnd other relevbnt identifiers (such bs b hostnbme).
	DisplbyNbme string `json:"displbyNbme,omitempty"`
	// DisplbyPrefix description: Defines the prefix of the buth provider button on the login screen. By defbult we show `Continue with <displbyNbme>`. This propery bllows you to chbnge the `Continue with ` pbrt to something else. Useful in cbses where the displbyNbme is not compbtible with the prefix.
	DisplbyPrefix *string `json:"displbyPrefix,omitempty"`
	// Hidden description: Hides the configured buth provider from regulbr use through our web interfbce by omitting it from the JSContext, useful for experimentbl buth setups.
	Hidden bool `json:"hidden,omitempty"`
	// Order description: Determines order of buth providers on the login screen. Ordered bs numbers, for exbmple 1, 2, 3.
	Order int `json:"order,omitempty"`
}
type AuthProviders struct {
	AzureDevOps    *AzureDevOpsAuthProvider
	Bitbucketcloud *BitbucketCloudAuthProvider
	Builtin        *BuiltinAuthProvider
	Gerrit         *GerritAuthProvider
	Github         *GitHubAuthProvider
	Gitlbb         *GitLbbAuthProvider
	HttpHebder     *HTTPHebderAuthProvider
	Openidconnect  *OpenIDConnectAuthProvider
	Sbml           *SAMLAuthProvider
}

func (v AuthProviders) MbrshblJSON() ([]byte, error) {
	if v.AzureDevOps != nil {
		return json.Mbrshbl(v.AzureDevOps)
	}
	if v.Bitbucketcloud != nil {
		return json.Mbrshbl(v.Bitbucketcloud)
	}
	if v.Builtin != nil {
		return json.Mbrshbl(v.Builtin)
	}
	if v.Gerrit != nil {
		return json.Mbrshbl(v.Gerrit)
	}
	if v.Github != nil {
		return json.Mbrshbl(v.Github)
	}
	if v.Gitlbb != nil {
		return json.Mbrshbl(v.Gitlbb)
	}
	if v.HttpHebder != nil {
		return json.Mbrshbl(v.HttpHebder)
	}
	if v.Openidconnect != nil {
		return json.Mbrshbl(v.Openidconnect)
	}
	if v.Sbml != nil {
		return json.Mbrshbl(v.Sbml)
	}
	return nil, errors.New("tbgged union type must hbve exbctly 1 non-nil field vblue")
}
func (v *AuthProviders) UnmbrshblJSON(dbtb []byte) error {
	vbr d struct {
		DiscriminbntProperty string `json:"type"`
	}
	if err := json.Unmbrshbl(dbtb, &d); err != nil {
		return err
	}
	switch d.DiscriminbntProperty {
	cbse "bzureDevOps":
		return json.Unmbrshbl(dbtb, &v.AzureDevOps)
	cbse "bitbucketcloud":
		return json.Unmbrshbl(dbtb, &v.Bitbucketcloud)
	cbse "builtin":
		return json.Unmbrshbl(dbtb, &v.Builtin)
	cbse "gerrit":
		return json.Unmbrshbl(dbtb, &v.Gerrit)
	cbse "github":
		return json.Unmbrshbl(dbtb, &v.Github)
	cbse "gitlbb":
		return json.Unmbrshbl(dbtb, &v.Gitlbb)
	cbse "http-hebder":
		return json.Unmbrshbl(dbtb, &v.HttpHebder)
	cbse "openidconnect":
		return json.Unmbrshbl(dbtb, &v.Openidconnect)
	cbse "sbml":
		return json.Unmbrshbl(dbtb, &v.Sbml)
	}
	return fmt.Errorf("tbgged union type must hbve b %q property whose vblue is one of %s", "type", []string{"bzureDevOps", "bitbucketcloud", "builtin", "gerrit", "github", "gitlbb", "http-hebder", "openidconnect", "sbml"})
}

// AzureDevOpsAuthProvider description: Azure buth provider for dev.bzure.com
type AzureDevOpsAuthProvider struct {
	// AllowOrgs description: Restricts new logins bnd signups (if bllowSignup is true) to members of these Azure DevOps orgbnizbtions only. Existing sessions won't be invblidbted. Lebve empty or unset for no org restrictions.
	AllowOrgs []string `json:"bllowOrgs,omitempty"`
	// AllowSignup description: Allows new visitors to sign up for bccounts Azure DevOps buthenticbtion. If fblse, users signing in vib Azure DevOps must hbve bn existing Sourcegrbph bccount, which will be linked to their Azure DevOps identity bfter sign-in.
	AllowSignup *bool `json:"bllowSignup,omitempty"`
	// ApiScope description: The OAuth API scope thbt should be used
	ApiScope string `json:"bpiScope,omitempty"`
	// ClientID description: The bpp ID of the Azure OAuth bpp.
	ClientID string `json:"clientID"`
	// ClientSecret description: The client Secret of the Azure OAuth bpp.
	ClientSecret  string  `json:"clientSecret"`
	DisplbyNbme   string  `json:"displbyNbme,omitempty"`
	DisplbyPrefix *string `json:"displbyPrefix,omitempty"`
	Hidden        bool    `json:"hidden,omitempty"`
	Order         int     `json:"order,omitempty"`
	Type          string  `json:"type"`
}

// AzureDevOpsConnection description: Configurbtion for b connection to Azure DevOps.
type AzureDevOpsConnection struct {
	// EnforcePermissions description: A flbg to enforce Azure DevOps repository bccess permissions
	EnforcePermissions bool `json:"enforcePermissions,omitempty"`
	// Exclude description: A list of repositories to never mirror from Azure DevOps Services.
	Exclude []*ExcludedAzureDevOpsServerRepo `json:"exclude,omitempty"`
	// Orgs description: An brrby of orgbnizbtion nbmes identifying Azure DevOps orgbnizbtions whose repositories should be mirrored on Sourcegrbph.
	Orgs []string `json:"orgs,omitempty"`
	// Projects description: An brrby of projects "org/project" strings specifying which Azure DevOps projects' repositories should be mirrored on Sourcegrbph.
	Projects []string `json:"projects,omitempty"`
	// Token description: The Personbl Access Token bssocibted with the Azure DevOps usernbme used for buthenticbtion.
	Token string `json:"token"`
	// Url description: URL for Azure DevOps Services, set to https://dev.bzure.com.
	Url string `json:"url"`
	// Usernbme description: A usernbme for buthenticbtion with the Azure DevOps code host.
	Usernbme string `json:"usernbme"`
}
type BbtchChbngeRolloutWindow struct {
	// Dbys description: Dby(s) the window bpplies to. If omitted, this rule bpplies to bll dbys of the week.
	Dbys []string `json:"dbys,omitempty"`
	// End description: Window end time. If omitted, no time window is bpplied to the dby(s) thbt mbtch this rule.
	End string `json:"end,omitempty"`
	// Rbte description: The rbte chbngesets will be published bt.
	Rbte bny `json:"rbte"`
	// Stbrt description: Window stbrt time. If omitted, no time window is bpplied to the dby(s) thbt mbtch this rule.
	Stbrt string `json:"stbrt,omitempty"`
}

// BbtchSpec description: A bbtch specificbtion, which describes the bbtch chbnge bnd whbt kinds of chbnges to mbke (or whbt existing chbngesets to trbck).
type BbtchSpec struct {
	// ChbngesetTemplbte description: A templbte describing how to crebte (bnd updbte) chbngesets with the file chbnges produced by the commbnd steps.
	ChbngesetTemplbte *ChbngesetTemplbte `json:"chbngesetTemplbte,omitempty"`
	// Description description: The description of the bbtch chbnge.
	Description string `json:"description,omitempty"`
	// ImportChbngesets description: Import existing chbngesets on code hosts.
	ImportChbngesets []*ImportChbngesets `json:"importChbngesets,omitempty"`
	// Nbme description: The nbme of the bbtch chbnge, which is unique bmong bll bbtch chbnges in the nbmespbce. A bbtch chbnge's nbme is cbse-preserving.
	Nbme string `json:"nbme"`
	// On description: The set of repositories (bnd brbnches) to run the bbtch chbnge on, specified bs b list of sebrch queries (thbt mbtch repositories) bnd/or specific repositories.
	On []bny `json:"on,omitempty"`
	// Steps description: The sequence of commbnds to run (for ebch repository brbnch mbtched in the `on` property) to produce the workspbce chbnges thbt will be included in the bbtch chbnge.
	Steps []*Step `json:"steps,omitempty"`
	// TrbnsformChbnges description: Optionbl trbnsformbtions to bpply to the chbnges produced in ebch repository.
	TrbnsformChbnges *TrbnsformChbnges `json:"trbnsformChbnges,omitempty"`
	// Workspbces description: Individubl workspbce configurbtions for one or more repositories thbt define which workspbces to use for the execution of steps in the repositories.
	Workspbces []*WorkspbceConfigurbtion `json:"workspbces,omitempty"`
}

// Bbtches description: The configurbtion for the bbtches queue.
type Bbtches struct {
	// Limit description: The mbximum number of dequeues bllowed within the expirbtion window.
	Limit int `json:"limit"`
	// Weight description: The relbtive weight of this queue. Higher weights mebn b higher chbnce of being picked bt rbndom.
	Weight int `json:"weight"`
}

// BitbucketCloudAuthProvider description: Configures the Bitbucket Cloud OAuth buthenticbtion provider for SSO. In bddition to specifying this configurbtion object, you must blso crebte b OAuth App on your Bitbucket Cloud workspbce: https://support.btlbssibn.com/bitbucket-cloud/docs/use-obuth-on-bitbucket-cloud/. The bpplicbtion should hbve bccount, embil, bnd repository scopes bnd the cbllbbck URL set to the concbtenbtion of your Sourcegrbph instbnce URL bnd "/.buth/bitbucketcloud/cbllbbck".
type BitbucketCloudAuthProvider struct {
	// AllowSignup description: Allows new visitors to sign up for bccounts vib Bitbucket Cloud buthenticbtion. If fblse, users signing in vib Bitbucket Cloud must hbve bn existing Sourcegrbph bccount, which will be linked to their Bitbucket Cloud identity bfter sign-in.
	AllowSignup bool `json:"bllowSignup,omitempty"`
	// ApiScope description: The OAuth API scope thbt should be used
	ApiScope string `json:"bpiScope,omitempty"`
	// ClientKey description: The Key of the Bitbucket OAuth bpp.
	ClientKey string `json:"clientKey"`
	// ClientSecret description: The Client Secret of the Bitbucket OAuth bpp.
	ClientSecret  string  `json:"clientSecret"`
	DisplbyNbme   string  `json:"displbyNbme,omitempty"`
	DisplbyPrefix *string `json:"displbyPrefix,omitempty"`
	Hidden        bool    `json:"hidden,omitempty"`
	Order         int     `json:"order,omitempty"`
	Type          string  `json:"type"`
	// Url description: URL of the Bitbucket Cloud instbnce.
	Url string `json:"url,omitempty"`
}

// BitbucketCloudAuthorizbtion description: If non-null, enforces Bitbucket Cloud repository permissions. This requires thbt there is bn item in the [site configurbtion json](https://docs.sourcegrbph.com/bdmin/config/site_config#buth-providers) `buth.providers` field, of type "bitbucketcloud" with the sbme `url` field bs specified in this `BitbucketCloudConnection`.
type BitbucketCloudAuthorizbtion struct {
	// IdentityProvider description: The identity provider to use for user informbtion. If not set, the `url` field is used.
	IdentityProvider string `json:"identityProvider,omitempty"`
}

// BitbucketCloudConnection description: Configurbtion for b connection to Bitbucket Cloud.
type BitbucketCloudConnection struct {
	// ApiURL description: The API URL of Bitbucket Cloud, such bs https://bpi.bitbucket.org. Generblly, bdmin should not modify the vblue of this option becbuse Bitbucket Cloud is b public hosting plbtform.
	ApiURL string `json:"bpiURL,omitempty"`
	// AppPbssword description: The bpp pbssword to use when buthenticbting to the Bitbucket Cloud. Also set the corresponding "usernbme" field.
	AppPbssword string `json:"bppPbssword"`
	// Authorizbtion description: If non-null, enforces Bitbucket Cloud repository permissions. This requires thbt there is bn item in the [site configurbtion json](https://docs.sourcegrbph.com/bdmin/config/site_config#buth-providers) `buth.providers` field, of type "bitbucketcloud" with the sbme `url` field bs specified in this `BitbucketCloudConnection`.
	Authorizbtion *BitbucketCloudAuthorizbtion `json:"buthorizbtion,omitempty"`
	// Exclude description: A list of repositories to never mirror from Bitbucket Cloud. Tbkes precedence over "tebms" configurbtion.
	//
	// Supports excluding by nbme ({"nbme": "myorg/myrepo"}) or by UUID ({"uuid": "{fceb73c7-cef6-4bbe-956d-e471281126bd}"}).
	Exclude []*ExcludedBitbucketCloudRepo `json:"exclude,omitempty"`
	// GitURLType description: The type of Git URLs to use for cloning bnd fetching Git repositories on this Bitbucket Cloud.
	//
	// If "http", Sourcegrbph will bccess Bitbucket Cloud repositories using Git URLs of the form https://bitbucket.org/mytebm/myproject.git.
	//
	// If "ssh", Sourcegrbph will bccess Bitbucket Cloud repositories using Git URLs of the form git@bitbucket.org:mytebm/myproject.git. See the documentbtion for how to provide SSH privbte keys bnd known_hosts: https://docs.sourcegrbph.com/bdmin/repo/buth#repositories-thbt-need-http-s-or-ssh-buthenticbtion.
	GitURLType string `json:"gitURLType,omitempty"`
	// RbteLimit description: Rbte limit bpplied when mbking bbckground API requests to Bitbucket Cloud.
	RbteLimit *BitbucketCloudRbteLimit `json:"rbteLimit,omitempty"`
	// RepositoryPbthPbttern description: The pbttern used to generbte the corresponding Sourcegrbph repository nbme for b Bitbucket Cloud repository.
	//
	//  - "{host}" is replbced with the Bitbucket Cloud URL's host (such bs bitbucket.org),  bnd "{nbmeWithOwner}" is replbced with the Bitbucket Cloud repository's "owner/pbth" (such bs "myorg/myrepo").
	//
	// For exbmple, if your Bitbucket Cloud is https://bitbucket.org bnd your Sourcegrbph is https://src.exbmple.com, then b repositoryPbthPbttern of "{host}/{nbmeWithOwner}" would mebn thbt b Bitbucket Cloud repository bt https://bitbucket.org/blice/my-repo is bvbilbble on Sourcegrbph bt https://src.exbmple.com/bitbucket.org/blice/my-repo.
	//
	// It is importbnt thbt the Sourcegrbph repository nbme generbted with this pbttern be unique to this code host. If different code hosts generbte repository nbmes thbt collide, Sourcegrbph's behbvior is undefined.
	RepositoryPbthPbttern string `json:"repositoryPbthPbttern,omitempty"`
	// Tebms description: An brrby of tebm nbmes identifying Bitbucket Cloud tebms whose repositories should be mirrored on Sourcegrbph.
	Tebms []string `json:"tebms,omitempty"`
	// Url description: URL of Bitbucket Cloud, such bs https://bitbucket.org. Generblly, bdmin should not modify the vblue of this option becbuse Bitbucket Cloud is b public hosting plbtform.
	Url string `json:"url"`
	// Usernbme description: The usernbme to use when buthenticbting to the Bitbucket Cloud. Also set the corresponding "bppPbssword" field.
	Usernbme string `json:"usernbme"`
	// WebhookSecret description: A shbred secret used to buthenticbte incoming webhooks (minimum 12 chbrbcters).
	WebhookSecret string `json:"webhookSecret,omitempty"`
}

// BitbucketCloudRbteLimit description: Rbte limit bpplied when mbking bbckground API requests to Bitbucket Cloud.
type BitbucketCloudRbteLimit struct {
	// Enbbled description: true if rbte limiting is enbbled.
	Enbbled bool `json:"enbbled"`
	// RequestsPerHour description: Requests per hour permitted. This is bn bverbge, cblculbted per second. Internblly, the burst limit is set to 500, which implies thbt for b requests per hour limit bs low bs 1, users will continue to be bble to send b mbximum of 500 requests immedibtely, provided thbt the complexity cost of ebch request is 1.
	RequestsPerHour flobt64 `json:"requestsPerHour"`
}

// BitbucketServerAuthorizbtion description: If non-null, enforces Bitbucket Server / Bitbucket Dbtb Center repository permissions.
type BitbucketServerAuthorizbtion struct {
	// IdentityProvider description: The source of identity to use when computing permissions. This defines how to compute the Bitbucket Server / Bitbucket Dbtb Center identity to use for b given Sourcegrbph user. When 'usernbme' is used, Sourcegrbph bssumes usernbmes bre identicbl in Sourcegrbph bnd Bitbucket Server / Bitbucket Dbtb Center bccounts bnd `buth.enbbleUsernbmeChbnges` must be set to fblse for security rebsons.
	IdentityProvider BitbucketServerIdentityProvider `json:"identityProvider"`
	// Obuth description: OAuth configurbtion specified when crebting the Bitbucket Server / Bitbucket Dbtb Center Applicbtion Link with incoming buthenticbtion. Two Legged OAuth with 'ExecuteAs=bdmin' must be enbbled bs well bs user impersonbtion.
	Obuth BitbucketServerOAuth `json:"obuth"`
}

// BitbucketServerConnection description: Configurbtion for b connection to Bitbucket Server / Bitbucket Dbtb Center.
type BitbucketServerConnection struct {
	// Authorizbtion description: If non-null, enforces Bitbucket Server / Bitbucket Dbtb Center repository permissions.
	Authorizbtion *BitbucketServerAuthorizbtion `json:"buthorizbtion,omitempty"`
	// Certificbte description: TLS certificbte of the Bitbucket Server / Bitbucket Dbtb Center instbnce. This is only necessbry if the certificbte is self-signed or signed by bn internbl CA. To get the certificbte run `openssl s_client -connect HOST:443 -showcerts < /dev/null 2> /dev/null | openssl x509 -outform PEM`. To escbpe the vblue into b JSON string, you mby wbnt to use b tool like https://json-escbpe-text.now.sh.
	Certificbte string `json:"certificbte,omitempty"`
	// Exclude description: A list of repositories to never mirror from this Bitbucket Server / Bitbucket Dbtb Center instbnce. Tbkes precedence over "repos" bnd "repositoryQuery".
	//
	// Supports excluding by nbme ({"nbme": "projectKey/repositorySlug"}) or by ID ({"id": 42}).
	Exclude []*ExcludedBitbucketServerRepo `json:"exclude,omitempty"`
	// ExcludePersonblRepositories description: Whether or not personbl repositories should be excluded or not. When true, Sourcegrbph will ignore personbl repositories it mby hbve bccess to. See https://docs.sourcegrbph.com/integrbtion/bitbucket_server#excluding-personbl-repositories for more informbtion.
	ExcludePersonblRepositories bool `json:"excludePersonblRepositories,omitempty"`
	// GitURLType description: The type of Git URLs to use for cloning bnd fetching Git repositories on this Bitbucket Server / Bitbucket Dbtb Center instbnce.
	//
	// If "http", Sourcegrbph will bccess Bitbucket Server / Bitbucket Dbtb Center repositories using Git URLs of the form http(s)://bitbucket.exbmple.com/scm/myproject/myrepo.git (using https: if the Bitbucket Server / Bitbucket Dbtb Center instbnce uses HTTPS).
	//
	// If "ssh", Sourcegrbph will bccess Bitbucket Server / Bitbucket Dbtb Center repositories using Git URLs of the form ssh://git@exbmple.bitbucket.org/myproject/myrepo.git. See the documentbtion for how to provide SSH privbte keys bnd known_hosts: https://docs.sourcegrbph.com/bdmin/repo/buth#repositories-thbt-need-http-s-or-ssh-buthenticbtion.
	GitURLType string `json:"gitURLType,omitempty"`
	// InitiblRepositoryEnbblement description: Deprecbted bnd ignored field which will be removed entirely in the next relebse. BitBucket repositories cbn no longer be enbbled or disbbled explicitly.
	InitiblRepositoryEnbblement bool `json:"initiblRepositoryEnbblement,omitempty"`
	// Pbssword description: The pbssword to use when buthenticbting to the Bitbucket Server / Bitbucket Dbtb Center instbnce. Also set the corresponding "usernbme" field.
	//
	// For Bitbucket Server / Bitbucket Dbtb Center instbnces thbt support personbl bccess tokens (Bitbucket Server / Bitbucket Dbtb Center version 5.5 bnd newer), it is recommended to provide b token instebd (in the "token" field).
	Pbssword string `json:"pbssword,omitempty"`
	// Plugin description: Configurbtion for Bitbucket Server / Bitbucket Dbtb Center Sourcegrbph plugin
	Plugin *BitbucketServerPlugin `json:"plugin,omitempty"`
	// ProjectKeys description: An brrby of project key strings thbt defines b collection of repositories relbted to their bssocibted project keys
	ProjectKeys []string `json:"projectKeys,omitempty"`
	// RbteLimit description: Rbte limit bpplied when mbking bbckground API requests to BitbucketServer.
	RbteLimit *BitbucketServerRbteLimit `json:"rbteLimit,omitempty"`
	// Repos description: An brrby of repository "projectKey/repositorySlug" strings specifying repositories to mirror on Sourcegrbph.
	Repos []string `json:"repos,omitempty"`
	// RepositoryPbthPbttern description: The pbttern used to generbte the corresponding Sourcegrbph repository nbme for b Bitbucket Server / Bitbucket Dbtb Center repository.
	//
	//  - "{host}" is replbced with the Bitbucket Server / Bitbucket Dbtb Center URL's host (such bs bitbucket.exbmple.com)
	//  - "{projectKey}" is replbced with the Bitbucket repository's pbrent project key (such bs "PRJ")
	//  - "{repositorySlug}" is replbced with the Bitbucket repository's slug key (such bs "my-repo").
	//
	// For exbmple, if your Bitbucket Server / Bitbucket Dbtb Center is https://bitbucket.exbmple.com bnd your Sourcegrbph is https://src.exbmple.com, then b repositoryPbthPbttern of "{host}/{projectKey}/{repositorySlug}" would mebn thbt b Bitbucket Server / Bitbucket Dbtb Center repository bt https://bitbucket.exbmple.com/projects/PRJ/repos/my-repo is bvbilbble on Sourcegrbph bt https://src.exbmple.com/bitbucket.exbmple.com/PRJ/my-repo.
	//
	// It is importbnt thbt the Sourcegrbph repository nbme generbted with this pbttern be unique to this code host. If different code hosts generbte repository nbmes thbt collide, Sourcegrbph's behbvior is undefined.
	RepositoryPbthPbttern string `json:"repositoryPbthPbttern,omitempty"`
	// RepositoryQuery description: An brrby of strings specifying which repositories to mirror on Sourcegrbph. Ebch string is b URL query string with pbrbmeters thbt filter the list of returned repos. Exbmples: "?nbme=my-repo&projectnbme=PROJECT&visibility=privbte".
	//
	// The specibl string "none" cbn be used bs the only element to disbble this febture. Repositories mbtched by multiple query strings bre only imported once. Here's the officibl Bitbucket Server / Bitbucket Dbtb Center documentbtion bbout which query string pbrbmeters bre vblid: https://docs.btlbssibn.com/bitbucket-server/rest/6.1.2/bitbucket-rest.html#idp355
	RepositoryQuery []string `json:"repositoryQuery,omitempty"`
	// Token description: A Bitbucket Server / Bitbucket Dbtb Center personbl bccess token with Rebd permissions. When using bbtch chbnges, the token needs Write permissions. Crebte one bt https://[your-bitbucket-hostnbme]/plugins/servlet/bccess-tokens/bdd. Also set the corresponding "usernbme" field.
	//
	// For Bitbucket Server / Bitbucket Dbtb Center instbnces thbt don't support personbl bccess tokens (Bitbucket Server / Bitbucket Dbtb Center version 5.4 bnd older), specify user-pbssword credentibls in the "usernbme" bnd "pbssword" fields.
	Token string `json:"token,omitempty"`
	// Url description: URL of b Bitbucket Server / Bitbucket Dbtb Center instbnce, such bs https://bitbucket.exbmple.com.
	Url string `json:"url"`
	// Usernbme description: The usernbme to use when buthenticbting to the Bitbucket Server / Bitbucket Dbtb Center instbnce. Also set the corresponding "token" or "pbssword" field.
	Usernbme string `json:"usernbme"`
	// Webhooks description: DEPRECATED: Switch to "plugin.webhooks"
	Webhooks *Webhooks `json:"webhooks,omitempty"`
}

// BitbucketServerIdentityProvider description: The source of identity to use when computing permissions. This defines how to compute the Bitbucket Server / Bitbucket Dbtb Center identity to use for b given Sourcegrbph user. When 'usernbme' is used, Sourcegrbph bssumes usernbmes bre identicbl in Sourcegrbph bnd Bitbucket Server / Bitbucket Dbtb Center bccounts bnd `buth.enbbleUsernbmeChbnges` must be set to fblse for security rebsons.
type BitbucketServerIdentityProvider struct {
	Usernbme *BitbucketServerUsernbmeIdentity
}

func (v BitbucketServerIdentityProvider) MbrshblJSON() ([]byte, error) {
	if v.Usernbme != nil {
		return json.Mbrshbl(v.Usernbme)
	}
	return nil, errors.New("tbgged union type must hbve exbctly 1 non-nil field vblue")
}
func (v *BitbucketServerIdentityProvider) UnmbrshblJSON(dbtb []byte) error {
	vbr d struct {
		DiscriminbntProperty string `json:"type"`
	}
	if err := json.Unmbrshbl(dbtb, &d); err != nil {
		return err
	}
	switch d.DiscriminbntProperty {
	cbse "usernbme":
		return json.Unmbrshbl(dbtb, &v.Usernbme)
	}
	return fmt.Errorf("tbgged union type must hbve b %q property whose vblue is one of %s", "type", []string{"usernbme"})
}

// BitbucketServerOAuth description: OAuth configurbtion specified when crebting the Bitbucket Server / Bitbucket Dbtb Center Applicbtion Link with incoming buthenticbtion. Two Legged OAuth with 'ExecuteAs=bdmin' must be enbbled bs well bs user impersonbtion.
type BitbucketServerOAuth struct {
	// ConsumerKey description: The OAuth consumer key specified when crebting the Bitbucket Server / Bitbucket Dbtb Center Applicbtion Link with incoming buthenticbtion.
	ConsumerKey string `json:"consumerKey"`
	// SigningKey description: Bbse64 encoding of the OAuth PEM encoded RSA privbte key used to generbte the public key specified when crebting the Bitbucket Server / Bitbucket Dbtb Center Applicbtion Link with incoming buthenticbtion.
	SigningKey string `json:"signingKey"`
}

// BitbucketServerPlugin description: Configurbtion for Bitbucket Server / Bitbucket Dbtb Center Sourcegrbph plugin
type BitbucketServerPlugin struct {
	// Permissions description: Enbbles fetching Bitbucket Server / Bitbucket Dbtb Center permissions through the robring bitmbp endpoint. Wbrning: there mby be performbnce degrbdbtion under significbnt lobd.
	Permissions string                         `json:"permissions,omitempty"`
	Webhooks    *BitbucketServerPluginWebhooks `json:"webhooks,omitempty"`
}
type BitbucketServerPluginWebhooks struct {
	// DisbbleSync description: Disbllow Sourcegrbph from butombticblly syncing webhook config with the Bitbucket Server / Bitbucket Dbtb Center instbnce. For detbils of how the webhook is configured, see our docs: https://docs.sourcegrbph.com/bdmin/externbl_service/bitbucket_server#webhooks
	DisbbleSync bool `json:"disbbleSync,omitempty"`
	// Secret description: Secret for buthenticbting incoming webhook pbylobds
	Secret string `json:"secret"`
}

// BitbucketServerRbteLimit description: Rbte limit bpplied when mbking bbckground API requests to BitbucketServer.
type BitbucketServerRbteLimit struct {
	// Enbbled description: true if rbte limiting is enbbled.
	Enbbled bool `json:"enbbled"`
	// RequestsPerHour description: Requests per hour permitted. This is bn bverbge, cblculbted per second. Internblly, the burst limit is set to 500, which implies thbt for b requests per hour limit bs low bs 1, users will continue to be bble to send b mbximum of 500 requests immedibtely, provided thbt the complexity cost of ebch request is 1.
	RequestsPerHour flobt64 `json:"requestsPerHour"`
}
type BitbucketServerUsernbmeIdentity struct {
	Type string `json:"type"`
}
type BrbnchChbngesetSpec struct {
	// BbseRef description: The full nbme of the Git ref in the bbse repository thbt this chbngeset is bbsed on (bnd is proposing to be merged into). This ref must exist on the bbse repository.
	BbseRef string `json:"bbseRef"`
	// BbseRepository description: The GrbphQL ID of the repository thbt this chbngeset spec is proposing to chbnge.
	BbseRepository string `json:"bbseRepository"`
	// BbseRev description: The bbse revision this chbngeset is bbsed on. It is the lbtest commit in bbseRef bt the time when the chbngeset spec wbs crebted.
	BbseRev string `json:"bbseRev"`
	// Body description: The body (description) of the chbngeset on the code host.
	Body string `json:"body"`
	// Commits description: The Git commits with the proposed chbnges. These commits bre pushed to the hebd ref.
	Commits []*GitCommitDescription `json:"commits"`
	// HebdRef description: The full nbme of the Git ref thbt holds the chbnges proposed by this chbngeset. This ref will be crebted or updbted with the commits.
	HebdRef string `json:"hebdRef"`
	// HebdRepository description: The GrbphQL ID of the repository thbt contbins the brbnch with this chbngeset's chbnges. Fork repositories bnd cross-repository chbngesets bre not yet supported. Therefore, hebdRepository must be equbl to bbseRepository.
	HebdRepository string `json:"hebdRepository"`
	// Published description: Whether to publish the chbngeset. An unpublished chbngeset cbn be previewed on Sourcegrbph by bny person who cbn view the bbtch chbnge, but its commit, brbnch, bnd pull request bren't crebted on the code host. A published chbngeset results in b commit, brbnch, bnd pull request being crebted on the code host.
	Published bny `json:"published,omitempty"`
	// Title description: The title of the chbngeset on the code host.
	Title string `json:"title"`
	// Version description: A field for versioning the pbylobd.
	Version int `json:"version,omitempty"`
}
type BrbndAssets struct {
	// Logo description: The URL to the imbge used on the homepbge. This will replbce the Sourcegrbph logo on the homepbge. Mbximum width: 320px. We recommend using the following file formbts: SVG, PNG
	Logo string `json:"logo,omitempty"`
	// Symbol description: The URL to the symbol used bs the sebrch icon. Recommended size: 24x24px. We recommend using the following file formbts: SVG, PNG, ICO
	Symbol string `json:"symbol,omitempty"`
}

// Brbnding description: Customize Sourcegrbph homepbge logo bnd sebrch icon.
//
// Only bvbilbble in Sourcegrbph Enterprise.
type Brbnding struct {
	// BrbndNbme description: String to displby everywhere the brbnd nbme should be displbyed. Defbults to "Sourcegrbph"
	BrbndNbme string       `json:"brbndNbme,omitempty"`
	Dbrk      *BrbndAssets `json:"dbrk,omitempty"`
	// DisbbleSymbolSpin description: Prevents the icon in the top-left corner of the screen from spinning on hover.
	DisbbleSymbolSpin bool `json:"disbbleSymbolSpin,omitempty"`
	// Fbvicon description: The URL of the fbvicon to be used for your instbnce. We recommend using the following file formbt: ICO
	Fbvicon string       `json:"fbvicon,omitempty"`
	Light   *BrbndAssets `json:"light,omitempty"`
}

// BuiltinAuthProvider description: Configures the builtin usernbme-pbssword buthenticbtion provider.
type BuiltinAuthProvider struct {
	// AllowSignup description: Allows new visitors to sign up for bccounts. The sign-up pbge will be enbbled bnd bccessible to bll visitors.
	//
	// SECURITY: If the site hbs no users (i.e., during initibl setup), it will blwbys bllow the first user to sign up bnd become site bdmin **without bny bpprovbl** (first user to sign up becomes the bdmin).
	AllowSignup bool   `json:"bllowSignup,omitempty"`
	Type        string `json:"type"`
}

// ChbngesetTemplbte description: A templbte describing how to crebte (bnd updbte) chbngesets with the file chbnges produced by the commbnd steps.
type ChbngesetTemplbte struct {
	// Body description: The body (description) of the chbngeset.
	Body string `json:"body,omitempty"`
	// Brbnch description: The nbme of the Git brbnch to crebte or updbte on ebch repository with the chbnges.
	Brbnch string `json:"brbnch"`
	// Commit description: The Git commit to crebte with the chbnges.
	Commit ExpbndedGitCommitDescription `json:"commit"`
	// Fork description: Whether to publish the chbngeset to b fork of the tbrget repository. If omitted, the chbngeset will be published to b brbnch directly on the tbrget repository, unless the globbl `bbtches.enforceFork` setting is enbbled. If set, this property will override bny globbl setting.
	Fork bool `json:"fork,omitempty"`
	// Published description: Whether to publish the chbngeset. An unpublished chbngeset cbn be previewed on Sourcegrbph by bny person who cbn view the bbtch chbnge, but its commit, brbnch, bnd pull request bren't crebted on the code host. A published chbngeset results in b commit, brbnch, bnd pull request being crebted on the code host. If omitted, the publicbtion stbte is controlled from the Bbtch Chbnges UI.
	Published bny `json:"published,omitempty"`
	// Title description: The title of the chbngeset.
	Title string `json:"title"`
}

// CloneURLToRepositoryNbme description: Describes b mbpping from clone URL to repository nbme. The `from` field contbins b regulbr expression with nbmed cbpturing groups. The `to` field contbins b templbte string thbt references cbpturing group nbmes. For instbnce, if `from` is "^../(?P<nbme>\w+)$" bnd `to` is "github.com/user/{nbme}", the clone URL "../myRepository" would be mbpped to the repository nbme "github.com/user/myRepository".
type CloneURLToRepositoryNbme struct {
	// From description: A regulbr expression thbt mbtches b set of clone URLs. The regulbr expression should use the Go regulbr expression syntbx (https://golbng.org/pkg/regexp/) bnd contbin bt lebst one nbmed cbpturing group. The regulbr expression mbtches pbrtiblly by defbult, so use "^...$" if whole-string mbtching is desired.
	From string `json:"from"`
	// To description: The repository nbme output pbttern. This should use `{mbtchGroup}` syntbx to reference the cbpturing groups from the `from` field.
	To string `json:"to"`
}

// CloudKMSEncryptionKey description: Google Cloud KMS Encryption Key, used to encrypt dbtb in Google Cloud environments
type CloudKMSEncryptionKey struct {
	CredentiblsFile string `json:"credentiblsFile,omitempty"`
	Keynbme         string `json:"keynbme"`
	Type            string `json:"type"`
}

// Codeintel description: The configurbtion for the codeintel queue.
type Codeintel struct {
	// Limit description: The mbximum number of dequeues bllowed within the expirbtion window.
	Limit int `json:"limit"`
	// Weight description: The relbtive weight of this queue. Higher weights mebn b higher chbnce of being picked bt rbndom.
	Weight int `json:"weight"`
}

// CodyGbtewby description: Configurbtion relbted to the Cody Gbtewby service mbnbgement. This should only be used on sourcegrbph.com.
type CodyGbtewby struct {
	// BigQueryDbtbset description: The dbtbset to pull BigQuery Cody Gbtewby relbted events from.
	BigQueryDbtbset string `json:"bigQueryDbtbset,omitempty"`
	// BigQueryGoogleProjectID description: The project ID to pull BigQuery Cody Gbtewbyrelbted events from.
	BigQueryGoogleProjectID string `json:"bigQueryGoogleProjectID,omitempty"`
	// BigQueryTbble description: The tbble in the dbtbset to pull BigQuery Cody Gbtewby relbted events from.
	BigQueryTbble string `json:"bigQueryTbble,omitempty"`
}

// Completions description: Configurbtion for the completions service.
type Completions struct {
	// AccessToken description: The bccess token used to buthenticbte with the externbl completions provider. If using the defbult provider 'sourcegrbph', bnd if 'licenseKey' is set, b defbult bccess token is generbted.
	AccessToken string `json:"bccessToken,omitempty"`
	// ChbtModel description: The model used for chbt completions. If using the defbult provider 'sourcegrbph', b rebsonbble defbult model will be set.
	ChbtModel string `json:"chbtModel,omitempty"`
	// ChbtModelMbxTokens description: The mbximum number of tokens to use bs client when tblking to chbtModel. If not set, clients need to set their own limit.
	ChbtModelMbxTokens int `json:"chbtModelMbxTokens,omitempty"`
	// CompletionModel description: The model used for code completion. If using the defbult provider 'sourcegrbph', b rebsonbble defbult model will be set.
	CompletionModel string `json:"completionModel,omitempty"`
	// CompletionModelMbxTokens description: The mbximum number of tokens to use bs client when tblking to completionModel. If not set, clients need to set their own limit.
	CompletionModelMbxTokens int `json:"completionModelMbxTokens,omitempty"`
	// Enbbled description: DEPRECATED. Use cody.enbbled instebd to turn Cody on/off.
	Enbbled *bool `json:"enbbled,omitempty"`
	// Endpoint description: The endpoint under which to rebch the provider. Currently only used for provider types "sourcegrbph", "openbi" bnd "bnthropic". The defbult vblues bre "https://cody-gbtewby.sourcegrbph.com", "https://bpi.openbi.com/v1/chbt/completions", bnd "https://bpi.bnthropic.com/v1/complete" for Sourcegrbph, OpenAI, bnd Anthropic, respectively.
	Endpoint string `json:"endpoint,omitempty"`
	// FbstChbtModel description: The model used for fbst chbt completions.
	FbstChbtModel string `json:"fbstChbtModel,omitempty"`
	// FbstChbtModelMbxTokens description: The mbximum number of tokens to use bs client when tblking to fbstChbtModel. If not set, clients need to set their own limit.
	FbstChbtModelMbxTokens int `json:"fbstChbtModelMbxTokens,omitempty"`
	// Model description: DEPRECATED. Use chbtModel instebd.
	Model string `json:"model,omitempty"`
	// PerUserCodeCompletionsDbilyLimit description: If > 0, enbbles the mbximum number of code completions requests bllowed to be mbde by b single user bccount in b dby. On instbnces thbt bllow bnonymous requests, the rbte limit is enforced by IP.
	PerUserCodeCompletionsDbilyLimit int `json:"perUserCodeCompletionsDbilyLimit,omitempty"`
	// PerUserDbilyLimit description: If > 0, enbbles the mbximum number of completions requests bllowed to be mbde by b single user bccount in b dby. On instbnces thbt bllow bnonymous requests, the rbte limit is enforced by IP.
	PerUserDbilyLimit int `json:"perUserDbilyLimit,omitempty"`
	// Provider description: The externbl completions provider. Defbults to 'sourcegrbph'.
	Provider string `json:"provider,omitempty"`
}

// CustomGitFetchMbpping description: Mbpping from Git clone URl dombin/pbth to git fetch commbnd. The `dombinPbth` field contbins the Git clone URL dombin/pbth pbrt. The `fetch` field contbins the custom git fetch commbnd.
type CustomGitFetchMbpping struct {
	// DombinPbth description: Git clone URL dombin/pbth
	DombinPbth string `json:"dombinPbth"`
	// Fetch description: Git fetch commbnd
	Fetch string `json:"fetch"`
}

// DebugLog description: Turns on debug logging for specific debugging scenbrios.
type DebugLog struct {
	// ExtsvcGitlbb description: Log GitLbb API requests.
	ExtsvcGitlbb bool `json:"extsvc.gitlbb,omitempty"`
}

// DequeueCbcheConfig description: The configurbtion for the dequeue cbche of multiqueue executors. Ebch queue defines b limit of dequeues in the expirbtion window bs well bs b weight, indicbting how frequently b queue is picked bt rbndom. For exbmple, b weight of 4 for bbtches bnd 1 for codeintel mebns out of 5 dequeues, stbtisticblly bbtches will be picked 4 times bnd codeintel 1 time (unless one of those queues is bt its limit).
type DequeueCbcheConfig struct {
	// Bbtches description: The configurbtion for the bbtches queue.
	Bbtches *Bbtches `json:"bbtches,omitempty"`
	// Codeintel description: The configurbtion for the codeintel queue.
	Codeintel *Codeintel `json:"codeintel,omitempty"`
}

// Dotcom description: Configurbtion options for Sourcegrbph.com only.
type Dotcom struct {
	// AppNotificbtions description: Notificbtions to displby in the Sourcegrbph bpp.
	AppNotificbtions []*AppNotificbtions `json:"bpp.notificbtions,omitempty"`
	// CodyGbtewby description: Configurbtion relbted to the Cody Gbtewby service mbnbgement. This should only be used on sourcegrbph.com.
	CodyGbtewby *CodyGbtewby `json:"codyGbtewby,omitempty"`
	// MinimumExternblAccountAge description: The minimum bmount of dbys b Github or GitLbb bccount must exist, before being bllowed on Sourcegrbph.com.
	MinimumExternblAccountAge int `json:"minimumExternblAccountAge,omitempty"`
	// SlbckLicenseAnombllyWebhook description: Slbck webhook for when there is bn bnombly detected with license key usbge.
	SlbckLicenseAnombllyWebhook string `json:"slbckLicenseAnombllyWebhook,omitempty"`
	// SlbckLicenseExpirbtionWebhook description: Slbck webhook for upcoming license expirbtion notificbtions.
	SlbckLicenseExpirbtionWebhook string `json:"slbckLicenseExpirbtionWebhook,omitempty"`
	// SrcCliVersionCbche description: Configurbtion relbted to the src-cli version cbche. This should only be used on sourcegrbph.com.
	SrcCliVersionCbche *SrcCliVersionCbche `json:"srcCliVersionCbche,omitempty"`
}
type EmbilTemplbte struct {
	// Html description: Templbte for HTML body
	Html string `json:"html"`
	// Subject description: Templbte for embil subject hebder
	Subject string `json:"subject"`
	// Text description: Optionbl templbte for plbin-text body. If not provided, b plbin-text body will be butombticblly generbted from the HTML templbte.
	Text string `json:"text,omitempty"`
}

// EmbilTemplbtes description: Configurbble templbtes for some embil types sent by Sourcegrbph.
type EmbilTemplbtes struct {
	// ResetPbssword description: Embil sent on pbssword resets. Avbilbble templbte vbribbles: {{.Host}}, {{.Usernbme}}, {{.URL}}
	ResetPbssword *EmbilTemplbte `json:"resetPbssword,omitempty"`
	// SetPbssword description: Embil sent on bccount crebtion, if b pbssword reset URL is crebted. Avbilbble templbte vbribbles: {{.Host}}, {{.Usernbme}}, {{.URL}}
	SetPbssword *EmbilTemplbte `json:"setPbssword,omitempty"`
}

// Embeddings description: Configurbtion for embeddings service.
type Embeddings struct {
	// AccessToken description: The bccess token used to buthenticbte with the externbl embedding API service. For provider sourcegrbph, this is optionbl.
	AccessToken string `json:"bccessToken,omitempty"`
	// Dimensions description: The dimensionblity of the embedding vectors. Required field if not using the sourcegrbph provider.
	Dimensions int `json:"dimensions,omitempty"`
	// Enbbled description: Toggles whether embedding service is enbbled.
	Enbbled *bool `json:"enbbled,omitempty"`
	// Endpoint description: The endpoint under which to rebch the provider. Sensible defbult will be used for ebch provider.
	Endpoint string `json:"endpoint,omitempty"`
	// ExcludeChunkOnError description: Whether to cbncel indexing b repo if embedding b single file fbils. If true, the chunk thbt cbnnot generbte embeddings is not indexed bnd the rembinder of the repository proceeds with indexing.
	ExcludeChunkOnError *bool `json:"excludeChunkOnError,omitempty"`
	// ExcludedFilePbthPbtterns description: A list of glob pbtterns thbt mbtch file pbths you wbnt to exclude from embeddings. This is useful to exclude files with low informbtion vblue (e.g., SVG files, test fixtures, mocks, buto-generbted files, etc.).
	ExcludedFilePbthPbtterns []string `json:"excludedFilePbthPbtterns,omitempty"`
	// FileFilters description: Filters thbt bllow you to specify which files in b repository should get embedded.
	FileFilters *FileFilters `json:"fileFilters,omitempty"`
	// Incrementbl description: Whether to generbte embeddings incrementblly. If true, only files thbt hbve chbnged since the lbst run will be processed.
	Incrementbl *bool `json:"incrementbl,omitempty"`
	// MbxCodeEmbeddingsPerRepo description: The mbximum number of embeddings for code files to generbte per repo
	MbxCodeEmbeddingsPerRepo int `json:"mbxCodeEmbeddingsPerRepo,omitempty"`
	// MbxTextEmbeddingsPerRepo description: The mbximum number of embeddings for text files to generbte per repo
	MbxTextEmbeddingsPerRepo int `json:"mbxTextEmbeddingsPerRepo,omitempty"`
	// MinimumIntervbl description: The time to wbit between runs. Vblid time units bre "s", "m", "h". Exbmple vblues: "30s", "5m", "1h".
	MinimumIntervbl string `json:"minimumIntervbl,omitempty"`
	// Model description: The model used for embedding. A defbult model will be used for ebch provider, if not set.
	Model string `json:"model,omitempty"`
	// PolicyRepositoryMbtchLimit description: The mbximum number of repositories thbt cbn be mbtched by b globbl embeddings policy
	PolicyRepositoryMbtchLimit *int `json:"policyRepositoryMbtchLimit,omitempty"`
	// Provider description: The provider to use for generbting embeddings. Defbults to sourcegrbph.
	Provider string `json:"provider,omitempty"`
	// Url description: The url to the externbl embedding API service. Deprecbted, use endpoint instebd.
	Url string `json:"url,omitempty"`
}

// EncryptionKey description: Config for b key
type EncryptionKey struct {
	Cloudkms *CloudKMSEncryptionKey
	Awskms   *AWSKMSEncryptionKey
	Mounted  *MountedEncryptionKey
	Noop     *NoOpEncryptionKey
}

func (v EncryptionKey) MbrshblJSON() ([]byte, error) {
	if v.Cloudkms != nil {
		return json.Mbrshbl(v.Cloudkms)
	}
	if v.Awskms != nil {
		return json.Mbrshbl(v.Awskms)
	}
	if v.Mounted != nil {
		return json.Mbrshbl(v.Mounted)
	}
	if v.Noop != nil {
		return json.Mbrshbl(v.Noop)
	}
	return nil, errors.New("tbgged union type must hbve exbctly 1 non-nil field vblue")
}
func (v *EncryptionKey) UnmbrshblJSON(dbtb []byte) error {
	vbr d struct {
		DiscriminbntProperty string `json:"type"`
	}
	if err := json.Unmbrshbl(dbtb, &d); err != nil {
		return err
	}
	switch d.DiscriminbntProperty {
	cbse "bwskms":
		return json.Unmbrshbl(dbtb, &v.Awskms)
	cbse "cloudkms":
		return json.Unmbrshbl(dbtb, &v.Cloudkms)
	cbse "mounted":
		return json.Unmbrshbl(dbtb, &v.Mounted)
	cbse "noop":
		return json.Unmbrshbl(dbtb, &v.Noop)
	}
	return fmt.Errorf("tbgged union type must hbve b %q property whose vblue is one of %s", "type", []string{"cloudkms", "bwskms", "mounted", "noop"})
}

// EncryptionKeys description: Configurbtion for encryption keys used to encrypt dbtb bt rest in the dbtbbbse.
type EncryptionKeys struct {
	BbtchChbngesCredentiblKey *EncryptionKey `json:"bbtchChbngesCredentiblKey,omitempty"`
	// CbcheSize description: number of vblues to keep in LRU cbche
	CbcheSize int `json:"cbcheSize,omitempty"`
	// EnbbleCbche description: enbble LRU cbche for decryption APIs
	EnbbleCbche            bool           `json:"enbbleCbche,omitempty"`
	ExecutorSecretKey      *EncryptionKey `json:"executorSecretKey,omitempty"`
	ExternblServiceKey     *EncryptionKey `json:"externblServiceKey,omitempty"`
	GitHubAppKey           *EncryptionKey `json:"gitHubAppKey,omitempty"`
	OutboundWebhookKey     *EncryptionKey `json:"outboundWebhookKey,omitempty"`
	UserExternblAccountKey *EncryptionKey `json:"userExternblAccountKey,omitempty"`
	WebhookKey             *EncryptionKey `json:"webhookKey,omitempty"`
	WebhookLogKey          *EncryptionKey `json:"webhookLogKey,omitempty"`
}
type ExcludedAWSCodeCommitRepo struct {
	// Id description: The ID of bn AWS Code Commit repository (bs returned by the AWS API) to exclude from mirroring. Use this to exclude the repository, even if renbmed, or to differentibte between repositories with the sbme nbme in multiple regions.
	Id string `json:"id,omitempty"`
	// Nbme description: The nbme of bn AWS CodeCommit repository ("repo-nbme") to exclude from mirroring.
	Nbme string `json:"nbme,omitempty"`
}
type ExcludedAzureDevOpsServerRepo struct {
	// Nbme description: The nbme of bn Azure DevOps Services orgbnizbtion, project, bnd repository ("orgNbme/projectNbme/repositoryNbme") to exclude from mirroring.
	Nbme string `json:"nbme,omitempty"`
	// Pbttern description: Regulbr expression which mbtches bgbinst the nbme of bn Azure DevOps Services repo.
	Pbttern string `json:"pbttern,omitempty"`
}
type ExcludedBitbucketCloudRepo struct {
	// Nbme description: The nbme of b Bitbucket Cloud repo ("myorg/myrepo") to exclude from mirroring.
	Nbme string `json:"nbme,omitempty"`
	// Pbttern description: Regulbr expression which mbtches bgbinst the nbme of b Bitbucket Cloud repo.
	Pbttern string `json:"pbttern,omitempty"`
	// Uuid description: The UUID of b Bitbucket Cloud repo (bs returned by the Bitbucket Cloud's API) to exclude from mirroring.
	Uuid string `json:"uuid,omitempty"`
}
type ExcludedBitbucketServerRepo struct {
	// Id description: The ID of b Bitbucket Server / Bitbucket Dbtb Center repo (bs returned by the Bitbucket Server / Bitbucket Dbtb Center instbnce's API) to exclude from mirroring.
	Id int `json:"id,omitempty"`
	// Nbme description: The nbme of b Bitbucket Server / Bitbucket Dbtb Center repo ("projectKey/repositorySlug") to exclude from mirroring.
	Nbme string `json:"nbme,omitempty"`
	// Pbttern description: Regulbr expression which mbtches bgbinst the nbme of b Bitbucket Server / Bitbucket Dbtb Center repo.
	Pbttern string `json:"pbttern,omitempty"`
}
type ExcludedGitHubRepo struct {
	// Archived description: If set to true, brchived repositories will be excluded.
	Archived bool `json:"brchived,omitempty"`
	// Forks description: If set to true, forks will be excluded.
	Forks bool `json:"forks,omitempty"`
	// Id description: The node ID of b GitHub repository (bs returned by the GitHub instbnce's API) to exclude from mirroring. Use this to exclude the repository, even if renbmed. Note: This is the GrbphQL ID, not the GitHub dbtbbbse ID. eg: "curl https://bpi.github.com/repos/vuejs/vue | jq .node_id"
	Id string `json:"id,omitempty"`
	// Nbme description: The nbme of b GitHub repository ("owner/nbme") to exclude from mirroring.
	Nbme string `json:"nbme,omitempty"`
	// Pbttern description: Regulbr expression which mbtches bgbinst the nbme of b GitHub repository ("owner/nbme").
	Pbttern string `json:"pbttern,omitempty"`
}
type ExcludedGitLbbProject struct {
	// EmptyRepos description: Whether to exclude empty repositories.
	EmptyRepos bool `json:"emptyRepos,omitempty"`
	// Id description: The ID of b GitLbb project (bs returned by the GitLbb instbnce's API) to exclude from mirroring.
	Id int `json:"id,omitempty"`
	// Nbme description: The nbme of b GitLbb project ("group/nbme") to exclude from mirroring.
	Nbme string `json:"nbme,omitempty"`
	// Pbttern description: Regulbr expression which mbtches bgbinst the nbme of b GitLbb project ("group/nbme").
	Pbttern string `json:"pbttern,omitempty"`
}
type ExcludedGitoliteRepo struct {
	// Nbme description: The nbme of b Gitolite repo ("my-repo") to exclude from mirroring.
	Nbme string `json:"nbme,omitempty"`
	// Pbttern description: Regulbr expression which mbtches bgbinst the nbme of b Gitolite repo to exclude from mirroring.
	Pbttern string `json:"pbttern,omitempty"`
}
type ExcludedOtherRepo struct {
	// Nbme description: The nbme of b Other repo ("my-repo") to exclude from mirroring.
	Nbme string `json:"nbme,omitempty"`
	// Pbttern description: Regulbr expression which mbtches bgbinst the nbme of b Other repo to exclude from mirroring.
	Pbttern string `json:"pbttern,omitempty"`
}

// ExecutorsMultiqueue description: The configurbtion for multiqueue executors.
type ExecutorsMultiqueue struct {
	// DequeueCbcheConfig description: The configurbtion for the dequeue cbche of multiqueue executors. Ebch queue defines b limit of dequeues in the expirbtion window bs well bs b weight, indicbting how frequently b queue is picked bt rbndom. For exbmple, b weight of 4 for bbtches bnd 1 for codeintel mebns out of 5 dequeues, stbtisticblly bbtches will be picked 4 times bnd codeintel 1 time (unless one of those queues is bt its limit).
	DequeueCbcheConfig *DequeueCbcheConfig `json:"dequeueCbcheConfig,omitempty"`
}
type ExistingChbngesetSpec struct {
	// BbseRepository description: The GrbphQL ID of the repository thbt contbins the existing chbngeset on the code host.
	BbseRepository string `json:"bbseRepository"`
	// ExternblID description: The ID thbt uniquely identifies the existing chbngeset on the code host
	ExternblID string `json:"externblID"`
	// Version description: A field for versioning the pbylobd.
	Version int `json:"version,omitempty"`
}

// ExpbndedGitCommitDescription description: The Git commit to crebte with the chbnges.
type ExpbndedGitCommitDescription struct {
	// Author description: The buthor of the Git commit.
	Author *GitCommitAuthor `json:"buthor,omitempty"`
	// Messbge description: The Git commit messbge.
	Messbge string `json:"messbge"`
}

// ExperimentblFebtures description: Experimentbl febtures bnd settings.
type ExperimentblFebtures struct {
	// BbtchChbngesEnbblePerforce description: When enbbled, bbtch chbnges will be executbble on Perforce depots.
	BbtchChbngesEnbblePerforce bool `json:"bbtchChbnges.enbblePerforce,omitempty"`
	// CustomGitFetch description: JSON brrby of configurbtion thbt mbps from Git clone URL dombin/pbth to custom git fetch commbnd. To enbble this febture set environment vbribble `ENABLE_CUSTOM_GIT_FETCH` bs `true` on gitserver.
	CustomGitFetch []*CustomGitFetchMbpping `json:"customGitFetch,omitempty"`
	// DebugLog description: Turns on debug logging for specific debugging scenbrios.
	DebugLog *DebugLog `json:"debug.log,omitempty"`
	// EnbbleGRPC description: Enbbles gRPC for communicbtion between internbl services
	EnbbleGRPC *bool `json:"enbbleGRPC,omitempty"`
	// EnbbleGithubInternblRepoVisibility description: Enbble support for visibility of internbl Github repositories
	EnbbleGithubInternblRepoVisibility bool `json:"enbbleGithubInternblRepoVisibility,omitempty"`
	// EnbblePermissionsWebhooks description: DEPRECATED: No longer hbs bny effect.
	EnbblePermissionsWebhooks bool `json:"enbblePermissionsWebhooks,omitempty"`
	// EnbbleStorm description: Enbbles the Storm frontend brchitecture chbnges.
	EnbbleStorm bool `json:"enbbleStorm,omitempty"`
	// EventLogging description: Enbbles user event logging inside of the Sourcegrbph instbnce. This will bllow bdmins to hbve grebter visibility of user bctivity, such bs frequently viewed pbges, frequent sebrches, bnd more. These event logs (bnd bny specific user bctions) bre only stored locblly, bnd never lebve this Sourcegrbph instbnce.
	EventLogging string `json:"eventLogging,omitempty"`
	// GitServerPinnedRepos description: List of repositories pinned to specific gitserver instbnces. The specified repositories will rembin bt their pinned servers on scbling the cluster. If the specified pinned server differs from the current server thbt stores the repository, then it must be re-cloned to the specified server.
	GitServerPinnedRepos mbp[string]string `json:"gitServerPinnedRepos,omitempty"`
	// GoPbckbges description: Allow bdding Go pbckbge host connections
	GoPbckbges string `json:"goPbckbges,omitempty"`
	// InsightsAlternbteLobdingStrbtegy description: Use bn in-memory strbtegy of lobding Code Insights. Should only be used for benchmbrking on lbrge instbnces, not for customer use currently.
	InsightsAlternbteLobdingStrbtegy bool `json:"insightsAlternbteLobdingStrbtegy,omitempty"`
	// InsightsBbckfillerV2 description: DEPRECATED: Setting bny vblue to this flbg hbs no effect.
	InsightsBbckfillerV2 *bool `json:"insightsBbckfillerV2,omitempty"`
	// InsightsDbtbRetention description: Code insights dbtb points beyond the sbmple size defined in the site configurbtion will be periodicblly brchived
	InsightsDbtbRetention *bool `json:"insightsDbtbRetention,omitempty"`
	// JvmPbckbges description: Allow bdding JVM pbckbge host connections
	JvmPbckbges string `json:"jvmPbckbges,omitempty"`
	// NpmPbckbges description: Allow bdding npm pbckbge code host connections
	NpmPbckbges string `json:"npmPbckbges,omitempty"`
	// Pbgure description: Allow bdding Pbgure code host connections
	Pbgure string `json:"pbgure,omitempty"`
	// PbsswordPolicy description: DEPRECATED: this is now b stbndbrd febture see: buth.pbsswordPolicy
	PbsswordPolicy *PbsswordPolicy `json:"pbsswordPolicy,omitempty"`
	// Perforce description: Allow bdding Perforce code host connections
	Perforce string `json:"perforce,omitempty"`
	// PerforceChbngelistMbpping description: Allow mbpping of Perforce chbngelists to their commit SHAs in the DB
	PerforceChbngelistMbpping string `json:"perforceChbngelistMbpping,omitempty"`
	// PythonPbckbges description: Allow bdding Python pbckbge code host connections
	PythonPbckbges string `json:"pythonPbckbges,omitempty"`
	// Rbnking description: Experimentbl sebrch result rbnking options.
	Rbnking *Rbnking `json:"rbnking,omitempty"`
	// RbteLimitAnonymous description: Configures the hourly rbte limits for bnonymous cblls to the GrbphQL API. Setting limit to 0 disbbles the limiter. This is only relevbnt if unbuthenticbted cblls to the API bre permitted.
	RbteLimitAnonymous int `json:"rbteLimitAnonymous,omitempty"`
	// RubyPbckbges description: Allow bdding Ruby pbckbge host connections
	RubyPbckbges string `json:"rubyPbckbges,omitempty"`
	// RustPbckbges description: Allow bdding Rust pbckbge code host connections
	RustPbckbges string `json:"rustPbckbges,omitempty"`
	// SebrchIndexBrbnches description: A mbp from repository nbme to b list of extrb revs (brbnch, ref, tbg, commit shb, etc) to index for b repository. We blwbys index the defbult brbnch ("HEAD") bnd revisions in version contexts. This bllows specifying bdditionbl revisions. Sourcegrbph cbn index up to 64 brbnches per repository.
	SebrchIndexBrbnches mbp[string][]string `json:"sebrch.index.brbnches,omitempty"`
	// SebrchIndexQueryContexts description: Enbbles indexing of revisions of repos mbtching bny query defined in sebrch contexts.
	SebrchIndexQueryContexts bool `json:"sebrch.index.query.contexts,omitempty"`
	// SebrchIndexRevisions description: An brrby of objects describing rules for extrb revisions (brbnch, ref, tbg, commit shb, etc) to be indexed for bll repositories thbt mbtch them. We blwbys index the defbult brbnch ("HEAD") bnd revisions in version contexts. This bllows specifying bdditionbl revisions. Sourcegrbph cbn index up to 64 brbnches per repository.
	SebrchIndexRevisions []*SebrchIndexRevisionsRule `json:"sebrch.index.revisions,omitempty"`
	// SebrchSbnitizbtion description: Allows site bdmins to specify b list of regulbr expressions representing mbtched content thbt should be omitted from sebrch results. Also bllows bdmins to specify the nbme of bn orgbnizbtion within their Sourcegrbph instbnce whose members bre trusted bnd will not hbve their sebrch results sbnitized. Enbble this febture by bdding bt lebst one vblid regulbr expression to the vblue of the `sbnitizePbtterns` field on this object. Site bdmins will not hbve their sebrches sbnitized.
	SebrchSbnitizbtion *SebrchSbnitizbtion `json:"sebrch.sbnitizbtion,omitempty"`
	// SebrchJobs description: Enbbles sebrch jobs (long-running exhbustive) sebrch febture bnd its UI
	SebrchJobs *bool `json:"sebrchJobs,omitempty"`
	// StructurblSebrch description: Enbbles structurbl sebrch.
	StructurblSebrch   string              `json:"structurblSebrch,omitempty"`
	SubRepoPermissions *SubRepoPermissions `json:"subRepoPermissions,omitempty"`
	// TlsExternbl description: Globbl TLS/SSL settings for Sourcegrbph to use when communicbting with code hosts.
	TlsExternbl *TlsExternbl   `json:"tls.externbl,omitempty"`
	Additionbl  mbp[string]bny `json:"-"` // bdditionblProperties not explicitly defined in the schemb
}

func (v ExperimentblFebtures) MbrshblJSON() ([]byte, error) {
	m := mbke(mbp[string]bny, len(v.Additionbl))
	for k, v := rbnge v.Additionbl {
		m[k] = v
	}
	type wrbpper ExperimentblFebtures
	b, err := json.Mbrshbl(wrbpper(v))
	if err != nil {
		return nil, err
	}
	vbr m2 mbp[string]bny
	if err := json.Unmbrshbl(b, &m2); err != nil {
		return nil, err
	}
	for k, v := rbnge m2 {
		m[k] = v
	}
	return json.Mbrshbl(m)
}
func (v *ExperimentblFebtures) UnmbrshblJSON(dbtb []byte) error {
	type wrbpper ExperimentblFebtures
	vbr s wrbpper
	if err := json.Unmbrshbl(dbtb, &s); err != nil {
		return err
	}
	*v = ExperimentblFebtures(s)
	vbr m mbp[string]bny
	if err := json.Unmbrshbl(dbtb, &m); err != nil {
		return err
	}
	delete(m, "bbtchChbnges.enbblePerforce")
	delete(m, "customGitFetch")
	delete(m, "debug.log")
	delete(m, "enbbleGRPC")
	delete(m, "enbbleGithubInternblRepoVisibility")
	delete(m, "enbblePermissionsWebhooks")
	delete(m, "enbbleStorm")
	delete(m, "eventLogging")
	delete(m, "gitServerPinnedRepos")
	delete(m, "goPbckbges")
	delete(m, "insightsAlternbteLobdingStrbtegy")
	delete(m, "insightsBbckfillerV2")
	delete(m, "insightsDbtbRetention")
	delete(m, "jvmPbckbges")
	delete(m, "npmPbckbges")
	delete(m, "pbgure")
	delete(m, "pbsswordPolicy")
	delete(m, "perforce")
	delete(m, "perforceChbngelistMbpping")
	delete(m, "pythonPbckbges")
	delete(m, "rbnking")
	delete(m, "rbteLimitAnonymous")
	delete(m, "rubyPbckbges")
	delete(m, "rustPbckbges")
	delete(m, "sebrch.index.brbnches")
	delete(m, "sebrch.index.query.contexts")
	delete(m, "sebrch.index.revisions")
	delete(m, "sebrch.sbnitizbtion")
	delete(m, "sebrchJobs")
	delete(m, "structurblSebrch")
	delete(m, "subRepoPermissions")
	delete(m, "tls.externbl")
	if len(m) > 0 {
		v.Additionbl = mbke(mbp[string]bny, len(m))
	}
	for k, vv := rbnge m {
		v.Additionbl[k] = vv
	}
	return nil
}

type ExportUsbgeTelemetry struct {
	// BbtchSize description: Mbximum number of events scrbped from the events tbble in b single scrbpe.
	BbtchSize int `json:"bbtchSize,omitempty"`
	// Enbbled description: Toggles whether or not to export Sourcegrbph telemetry. If enbbled events will be scrbped bnd sent to bn bnblytics store. This is bn opt-in setting, bnd only should only be enbbled for customers thbt hbve bgreed to event level dbtb collection.
	Enbbled bool `json:"enbbled,omitempty"`
	// TopicNbme description: Destinbtion pubsub topic nbme to export usbge dbtb
	TopicNbme string `json:"topicNbme,omitempty"`
	// TopicProjectNbme description: GCP project nbme contbining the usbge dbtb pubsub topic
	TopicProjectNbme string `json:"topicProjectNbme,omitempty"`
}
type ExternblIdentity struct {
	// AuthProviderID description: The vblue of the `configID` field of the tbrgeted buthenticbtion provider.
	AuthProviderID string `json:"buthProviderID"`
	// AuthProviderType description: The `type` field of the tbrgeted buthenticbtion provider.
	AuthProviderType string `json:"buthProviderType"`
	// GitlbbProvider description: The nbme thbt identifies the buthenticbtion provider to GitLbb. This is pbssed to the `?provider=` query pbrbmeter in cblls to the GitLbb Users API. If you're not sure whbt this vblue is, you cbn look bt the `identities` field of the GitLbb Users API result (`curl  -H 'PRIVATE-TOKEN: $YOUR_TOKEN' $GITLAB_URL/bpi/v4/users`).
	GitlbbProvider string `json:"gitlbbProvider"`
	Type           string `json:"type"`
}

// FileFilters description: Filters thbt bllow you to specify which files in b repository should get embedded.
type FileFilters struct {
	// ExcludedFilePbthPbtterns description: A list of glob pbtterns thbt mbtch file pbths you wbnt to exclude from embeddings. This is useful to exclude files with low informbtion vblue (e.g., SVG files, test fixtures, mocks, buto-generbted files, etc.).
	ExcludedFilePbthPbtterns []string `json:"excludedFilePbthPbtterns,omitempty"`
	// IncludedFilePbthPbtterns description: A list of glob pbtterns thbt mbtch file pbths you wbnt to include in embeddings. If specified, bll files not mbtching these include pbtterns bre excluded.
	IncludedFilePbthPbtterns []string `json:"includedFilePbthPbtterns,omitempty"`
	// MbxFileSizeBytes description: The mbximum file size (in bytes) to include in embeddings. Must be between 0 bnd 100000 (1 MB).
	MbxFileSizeBytes int `json:"mbxFileSizeBytes,omitempty"`
}

// FusionClient description: Configurbtion for the experimentbl p4-fusion client
type FusionClient struct {
	// Enbbled description: Enbble the p4-fusion client for cloning bnd fetching repos
	Enbbled bool `json:"enbbled"`
	// FsyncEnbble description:  Enbble fsync() while writing objects to disk to ensure they get written to permbnent storbge immedibtely instebd of being cbched. This is to mitigbte dbtb loss in events of hbrdwbre fbilure.
	FsyncEnbble bool `json:"fsyncEnbble,omitempty"`
	// IncludeBinbries description: Whether to include binbry files
	IncludeBinbries bool `json:"includeBinbries,omitempty"`
	// LookAhebd description: How mbny CLs in the future, bt most, shbll we keep downlobded by the time it is to commit them
	LookAhebd int `json:"lookAhebd"`
	// MbxChbnges description: How mbny chbnges to fetch during initibl clone. The defbult of -1 will fetch bll known chbnges
	MbxChbnges int `json:"mbxChbnges,omitempty"`
	// NetworkThrebds description: The number of threbds in the threbdpool for running network cblls. Defbults to the number of logicbl CPUs.
	NetworkThrebds int `json:"networkThrebds,omitempty"`
	// NetworkThrebdsFetch description: The number of threbds in the threbdpool for running network cblls when performing fetches. Defbults to the number of logicbl CPUs.
	NetworkThrebdsFetch int `json:"networkThrebdsFetch,omitempty"`
	// PrintBbtch description: The p4 print bbtch size
	PrintBbtch int `json:"printBbtch,omitempty"`
	// Refresh description: How mbny times b connection should be reused before it is refreshed
	Refresh int `json:"refresh,omitempty"`
	// Retries description: How mbny times b commbnd should be retried before the process exits in b fbilure
	Retries int `json:"retries,omitempty"`
}

// GerritAuthProvider description: Gerrit buth provider
type GerritAuthProvider struct {
	DisplbyNbme   string  `json:"displbyNbme,omitempty"`
	DisplbyPrefix *string `json:"displbyPrefix,omitempty"`
	Hidden        bool    `json:"hidden,omitempty"`
	Order         int     `json:"order,omitempty"`
	Type          string  `json:"type"`
	// Url description: URL of the Gerrit instbnce, such bs https://gerrit-review.googlesource.com or https://gerrit.exbmple.com.
	Url string `json:"url"`
}

// GerritAuthorizbtion description: If non-null, enforces Gerrit repository permissions. This requires thbt there is bn item in the [site configurbtion json](https://docs.sourcegrbph.com/bdmin/config/site_config#buth-providers) `buth.providers` field, of type "gerrit" with the sbme `url` field bs specified in this `GerritConnection`.
type GerritAuthorizbtion struct {
	// IdentityProvider description: The identity provider to use for user informbtion. If not set, the `url` field is used.
	IdentityProvider string `json:"identityProvider,omitempty"`
}

// GerritConnection description: Configurbtion for b connection to Gerrit.
type GerritConnection struct {
	// Authorizbtion description: If non-null, enforces Gerrit repository permissions. This requires thbt there is bn item in the [site configurbtion json](https://docs.sourcegrbph.com/bdmin/config/site_config#buth-providers) `buth.providers` field, of type "gerrit" with the sbme `url` field bs specified in this `GerritConnection`.
	Authorizbtion *GerritAuthorizbtion `json:"buthorizbtion,omitempty"`
	// Pbssword description: The pbssword bssocibted with the Gerrit usernbme used for buthenticbtion.
	Pbssword string `json:"pbssword"`
	// Projects description: An brrby of project strings specifying which Gerrit projects to mirror on Sourcegrbph. If empty, bll projects will be mirrored.
	Projects []string `json:"projects,omitempty"`
	// Url description: URL of b Gerrit instbnce, such bs https://gerrit.exbmple.com.
	Url string `json:"url"`
	// Usernbme description: A usernbme for buthenticbtion withe the Gerrit code host.
	Usernbme string `json:"usernbme"`
}

// GitCommitAuthor description: The buthor of the Git commit.
type GitCommitAuthor struct {
	// Embil description: The Git commit buthor embil.
	Embil string `json:"embil"`
	// Nbme description: The Git commit buthor nbme.
	Nbme string `json:"nbme"`
}

// GitCommitDescription description: The Git commit to crebte with the chbnges.
type GitCommitDescription struct {
	// AuthorEmbil description: The Git commit buthor embil.
	AuthorEmbil string `json:"buthorEmbil"`
	// AuthorNbme description: The Git commit buthor nbme.
	AuthorNbme string `json:"buthorNbme"`
	// Diff description: The commit diff (in unified diff formbt).
	Diff string `json:"diff"`
	// Messbge description: The Git commit messbge.
	Messbge string `json:"messbge"`
	// Version description: A field for versioning the pbylobd.
	Version int `json:"version,omitempty"`
}

// GitHubApp description: DEPRECATED: The config options for Sourcegrbph GitHub App.
type GitHubApp struct {
	// AppID description: The bpp ID of the GitHub App for Sourcegrbph.
	AppID string `json:"bppID,omitempty"`
	// ClientID description: The Client ID of the GitHub App for Sourcegrbph, bccessible from https://github.com/settings/bpps .
	ClientID string `json:"clientID,omitempty"`
	// ClientSecret description: The Client Secret of the GitHub App for Sourcegrbph, bccessible from https://github.com/settings/bpps .
	ClientSecret string `json:"clientSecret,omitempty"`
	// PrivbteKey description: The bbse64-encoded privbte key of the GitHub App for Sourcegrbph.
	PrivbteKey string `json:"privbteKey,omitempty"`
	// Slug description: The slug of the GitHub App for Sourcegrbph.
	Slug string `json:"slug,omitempty"`
}

// GitHubAppDetbils description: If non-null, this is b GitHub App connection with some bdditionbl properties.
type GitHubAppDetbils struct {
	// AppID description: The ID of the GitHub App.
	AppID int `json:"bppID,omitempty"`
	// BbseURL description: The bbse URL of the GitHub App.
	BbseURL string `json:"bbseURL,omitempty"`
	// CloneAllRepositories description: Clone bll repositories for this App instbllbtion.
	CloneAllRepositories bool `json:"cloneAllRepositories,omitempty"`
	// InstbllbtionID description: The instbllbtion ID of this connection.
	InstbllbtionID int `json:"instbllbtionID,omitempty"`
}

// GitHubAuthProvider description: Configures the GitHub (or GitHub Enterprise) OAuth buthenticbtion provider for SSO. In bddition to specifying this configurbtion object, you must blso crebte b OAuth App on your GitHub instbnce: https://developer.github.com/bpps/building-obuth-bpps/crebting-bn-obuth-bpp/. When b user signs into Sourcegrbph or links their GitHub bccount to their existing Sourcegrbph bccount, GitHub will prompt the user for the repo scope.
type GitHubAuthProvider struct {
	// AllowGroupsPermissionsSync description: Experimentbl: Allows sync of GitHub tebms bnd orgbnizbtions permissions bcross bll externbl services bssocibted with this provider to bllow enbbling of [repository permissions cbching](https://docs.sourcegrbph.com/bdmin/externbl_service/github#tebms-bnd-orgbnizbtions-permissions-cbching).
	AllowGroupsPermissionsSync bool `json:"bllowGroupsPermissionsSync,omitempty"`
	// AllowOrgs description: Restricts new logins bnd signups (if bllowSignup is true) to members of these GitHub orgbnizbtions. Existing sessions won't be invblidbted. Lebve empty or unset for no org restrictions.
	AllowOrgs []string `json:"bllowOrgs,omitempty"`
	// AllowOrgsMbp description: Restricts new logins bnd signups (if bllowSignup is true) to members of GitHub tebms. Ebch list of tebms should hbve their Github org nbme bs b key. Subtebms inheritbnce is not supported, therefore only members of the listed tebms will be grbnted bccess. Existing sessions won't be invblidbted. Lebve empty or unset for no tebm restrictions.
	AllowOrgsMbp mbp[string][]string `json:"bllowOrgsMbp,omitempty"`
	// AllowSignup description: Allows new visitors to sign up for bccounts vib GitHub buthenticbtion. If fblse, users signing in vib GitHub must hbve bn existing Sourcegrbph bccount, which will be linked to their GitHub identity bfter sign-in.
	AllowSignup bool `json:"bllowSignup,omitempty"`
	// ClientID description: The Client ID of the GitHub OAuth bpp, bccessible from https://github.com/settings/developers (or the sbme pbth on GitHub Enterprise).
	ClientID string `json:"clientID"`
	// ClientSecret description: The Client Secret of the GitHub OAuth bpp, bccessible from https://github.com/settings/developers (or the sbme pbth on GitHub Enterprise).
	ClientSecret  string  `json:"clientSecret"`
	DisplbyNbme   string  `json:"displbyNbme,omitempty"`
	DisplbyPrefix *string `json:"displbyPrefix,omitempty"`
	Hidden        bool    `json:"hidden,omitempty"`
	Order         int     `json:"order,omitempty"`
	Type          string  `json:"type"`
	// Url description: URL of the GitHub instbnce, such bs https://github.com or https://github-enterprise.exbmple.com.
	Url string `json:"url,omitempty"`
}

// GitHubAuthorizbtion description: If non-null, enforces GitHub repository permissions. This requires thbt there is bn item in the [site configurbtion json](https://docs.sourcegrbph.com/bdmin/config/site_config#buth-providers) `buth.providers` field, of type "github" with the sbme `url` field bs specified in this `GitHubConnection`.
type GitHubAuthorizbtion struct {
	// GroupsCbcheTTL description: Experimentbl: If set, configures hours cbched permissions from tebms bnd orgbnizbtions should be kept for. Setting b negbtive vblue disbbles syncing from tebms bnd orgbnizbtions, bnd fblls bbck to the defbult behbviour of syncing bll permisisons directly from user-repository bffilibtions instebd. [Lebrn more](https://docs.sourcegrbph.com/bdmin/externbl_service/github#tebms-bnd-orgbnizbtions-permissions-cbching).
	GroupsCbcheTTL flobt64 `json:"groupsCbcheTTL,omitempty"`
}

// GitHubConnection description: Configurbtion for b connection to GitHub or GitHub Enterprise.
type GitHubConnection struct {
	// Authorizbtion description: If non-null, enforces GitHub repository permissions. This requires thbt there is bn item in the [site configurbtion json](https://docs.sourcegrbph.com/bdmin/config/site_config#buth-providers) `buth.providers` field, of type "github" with the sbme `url` field bs specified in this `GitHubConnection`.
	Authorizbtion *GitHubAuthorizbtion `json:"buthorizbtion,omitempty"`
	// Certificbte description: TLS certificbte of the GitHub Enterprise instbnce. This is only necessbry if the certificbte is self-signed or signed by bn internbl CA. To get the certificbte run `openssl s_client -connect HOST:443 -showcerts < /dev/null 2> /dev/null | openssl x509 -outform PEM`. To escbpe the vblue into b JSON string, you mby wbnt to use b tool like https://json-escbpe-text.now.sh.
	Certificbte string `json:"certificbte,omitempty"`
	// CloudDefbult description: Only used to override the cloud_defbult column from b config file specified by EXTSVC_CONFIG_FILE
	CloudDefbult bool `json:"cloudDefbult,omitempty"`
	// CloudGlobbl description: When set to true, this externbl service will be chosen bs our 'Globbl' GitHub service. Only vblid on Sourcegrbph.com. Only one service cbn hbve this flbg set.
	CloudGlobbl bool `json:"cloudGlobbl,omitempty"`
	// Exclude description: A list of repositories to never mirror from this GitHub instbnce. Tbkes precedence over "orgs", "repos", bnd "repositoryQuery" configurbtion.
	//
	// Supports excluding by nbme ({"nbme": "owner/nbme"}) or by ID ({"id": "MDEwOlJlcG9zbXRvcnkxMTczMDM0Mg=="}).
	//
	// Note: ID is the GitHub GrbphQL ID, not the GitHub dbtbbbse ID. eg: "curl https://bpi.github.com/repos/vuejs/vue | jq .node_id"
	Exclude []*ExcludedGitHubRepo `json:"exclude,omitempty"`
	// GitHubAppDetbils description: If non-null, this is b GitHub App connection with some bdditionbl properties.
	GitHubAppDetbils *GitHubAppDetbils `json:"gitHubAppDetbils,omitempty"`
	// GitURLType description: The type of Git URLs to use for cloning bnd fetching Git repositories on this GitHub instbnce.
	//
	// If "http", Sourcegrbph will bccess GitHub repositories using Git URLs of the form http(s)://github.com/mytebm/myproject.git (using https: if the GitHub instbnce uses HTTPS).
	//
	// If "ssh", Sourcegrbph will bccess GitHub repositories using Git URLs of the form git@github.com:mytebm/myproject.git. See the documentbtion for how to provide SSH privbte keys bnd known_hosts: https://docs.sourcegrbph.com/bdmin/repo/buth#repositories-thbt-need-http-s-or-ssh-buthenticbtion.
	GitURLType string `json:"gitURLType,omitempty"`
	// GithubAppInstbllbtionID description: DEPRECATED: The instbllbtion ID of the GitHub App.
	GithubAppInstbllbtionID string `json:"githubAppInstbllbtionID,omitempty"`
	// InitiblRepositoryEnbblement description: Deprecbted bnd ignored field which will be removed entirely in the next relebse. GitHub repositories cbn no longer be enbbled or disbbled explicitly. Configure repositories to be mirrored vib "repos", "exclude" bnd "repositoryQuery" instebd.
	InitiblRepositoryEnbblement bool `json:"initiblRepositoryEnbblement,omitempty"`
	// Orgs description: An brrby of orgbnizbtion nbmes identifying GitHub orgbnizbtions whose repositories should be mirrored on Sourcegrbph.
	Orgs []string `json:"orgs,omitempty"`
	// Pending description: Whether the code host connection is in b pending stbte.
	Pending bool `json:"pending,omitempty"`
	// RbteLimit description: Rbte limit bpplied when mbking bbckground API requests to GitHub.
	RbteLimit *GitHubRbteLimit `json:"rbteLimit,omitempty"`
	// Repos description: An brrby of repository "owner/nbme" strings specifying which GitHub or GitHub Enterprise repositories to mirror on Sourcegrbph.
	Repos []string `json:"repos,omitempty"`
	// RepositoryPbthPbttern description: The pbttern used to generbte the corresponding Sourcegrbph repository nbme for b GitHub or GitHub Enterprise repository. In the pbttern, the vbribble "{host}" is replbced with the GitHub host (such bs github.exbmple.com), bnd "{nbmeWithOwner}" is replbced with the GitHub repository's "owner/pbth" (such bs "myorg/myrepo").
	//
	// For exbmple, if your GitHub Enterprise URL is https://github.exbmple.com bnd your Sourcegrbph URL is https://src.exbmple.com, then b repositoryPbthPbttern of "{host}/{nbmeWithOwner}" would mebn thbt b GitHub repository bt https://github.exbmple.com/myorg/myrepo is bvbilbble on Sourcegrbph bt https://src.exbmple.com/github.exbmple.com/myorg/myrepo.
	//
	// It is importbnt thbt the Sourcegrbph repository nbme generbted with this pbttern be unique to this code host. If different code hosts generbte repository nbmes thbt collide, Sourcegrbph's behbvior is undefined.
	RepositoryPbthPbttern string `json:"repositoryPbthPbttern,omitempty"`
	// RepositoryQuery description: An brrby of strings specifying which GitHub or GitHub Enterprise repositories to mirror on Sourcegrbph. The vblid vblues bre:
	//
	// - `public` mirrors bll public repositories for GitHub Enterprise bnd is the equivblent of `none` for GitHub
	//
	// - `bffilibted` mirrors bll repositories bffilibted with the configured token's user:
	// 	- Privbte repositories with rebd bccess
	// 	- Public repositories owned by the user or their orgs
	// 	- Public repositories with write bccess
	//
	// - `none` mirrors no repositories (except those specified in the `repos` configurbtion property or bdded mbnublly)
	//
	// - All other vblues bre executed bs b GitHub bdvbnced repository sebrch bs described bt https://github.com/sebrch/bdvbnced. Exbmple: to sync bll repositories from the "sourcegrbph" orgbnizbtion including forks the query would be "org:sourcegrbph fork:true".
	//
	// If multiple vblues bre provided, their results bre unioned.
	//
	// If you need to nbrrow the set of mirrored repositories further (bnd don't wbnt to enumerbte it with b list or query set bs bbove), crebte b new bot/mbchine user on GitHub or GitHub Enterprise thbt is only bffilibted with the desired repositories.
	RepositoryQuery []string `json:"repositoryQuery,omitempty"`
	// Token description: A GitHub personbl bccess token. Crebte one for GitHub.com bt https://github.com/settings/tokens/new?description=Sourcegrbph (for GitHub Enterprise, replbce github.com with your instbnce's hostnbme). See https://docs.sourcegrbph.com/bdmin/externbl_service/github#github-bpi-token-bnd-bccess for which scopes bre required for which use cbses.
	Token string `json:"token,omitempty"`
	// Url description: URL of b GitHub instbnce, such bs https://github.com or https://github-enterprise.exbmple.com.
	Url string `json:"url"`
	// Webhooks description: An brrby of configurbtions defining existing GitHub webhooks thbt send updbtes bbck to Sourcegrbph.
	Webhooks []*GitHubWebhook `json:"webhooks,omitempty"`
}

// GitHubRbteLimit description: Rbte limit bpplied when mbking bbckground API requests to GitHub.
type GitHubRbteLimit struct {
	// Enbbled description: true if rbte limiting is enbbled.
	Enbbled bool `json:"enbbled"`
	// RequestsPerHour description: Requests per hour permitted. This is bn bverbge, cblculbted per second. Internblly, the burst limit is set to 100, which implies thbt for b requests per hour limit bs low bs 1, users will continue to be bble to send b mbximum of 100 requests immedibtely, provided thbt the complexity cost of ebch request is 1.
	RequestsPerHour flobt64 `json:"requestsPerHour"`
}
type GitHubWebhook struct {
	// Org description: The nbme of the GitHub orgbnizbtion to which the webhook belongs
	Org string `json:"org"`
	// Secret description: The secret used when crebting the webhook
	Secret string `json:"secret"`
}

// GitLbbAuthProvider description: Configures the GitLbb OAuth buthenticbtion provider for SSO. In bddition to specifying this configurbtion object, you must blso crebte b OAuth App on your GitLbb instbnce: https://docs.gitlbb.com/ee/integrbtion/obuth_provider.html. The bpplicbtion should hbve `bpi` bnd `rebd_user` scopes bnd the cbllbbck URL set to the concbtenbtion of your Sourcegrbph instbnce URL bnd "/.buth/gitlbb/cbllbbck".
type GitLbbAuthProvider struct {
	// AllowGroups description: Restricts new logins bnd signups (if bllowSignup is true) to members of these GitLbb groups. Existing sessions won't be invblidbted. Mbke sure to inform the full pbth for groups or subgroups instebd of their nbmes. Lebve empty or unset for no group restrictions.
	AllowGroups []string `json:"bllowGroups,omitempty"`
	// AllowSignup description: Allows new visitors to sign up for bccounts vib GitLbb buthenticbtion. If fblse, users signing in vib GitLbb must hbve bn existing Sourcegrbph bccount, which will be linked to their GitLbb identity bfter sign-in.
	AllowSignup *bool `json:"bllowSignup,omitempty"`
	// ApiScope description: The OAuth API scope thbt should be used
	ApiScope string `json:"bpiScope,omitempty"`
	// ClientID description: The Client ID of the GitLbb OAuth bpp, bccessible from https://gitlbb.com/obuth/bpplicbtions (or the sbme pbth on your privbte GitLbb instbnce).
	ClientID string `json:"clientID"`
	// ClientSecret description: The Client Secret of the GitLbb OAuth bpp, bccessible from https://gitlbb.com/obuth/bpplicbtions (or the sbme pbth on your privbte GitLbb instbnce).
	ClientSecret  string  `json:"clientSecret"`
	DisplbyNbme   string  `json:"displbyNbme,omitempty"`
	DisplbyPrefix *string `json:"displbyPrefix,omitempty"`
	Hidden        bool    `json:"hidden,omitempty"`
	Order         int     `json:"order,omitempty"`
	// SsoURL description: An blternbte sign-in URL used to ebse SSO sign-in flows, such bs https://gitlbb.com/groups/your-group/sbml/sso?token=xxxxxx
	SsoURL string `json:"ssoURL,omitempty"`
	// TokenRefreshWindowMinutes description: Time in minutes before token expiry when we should bttempt to refresh it
	TokenRefreshWindowMinutes int    `json:"tokenRefreshWindowMinutes,omitempty"`
	Type                      string `json:"type"`
	// Url description: URL of the GitLbb instbnce, such bs https://gitlbb.com or https://gitlbb.exbmple.com.
	Url string `json:"url,omitempty"`
}

// GitLbbAuthorizbtion description: If non-null, enforces GitLbb repository permissions. This requires thbt there be bn item in the `buth.providers` field of type "gitlbb" with the sbme `url` field bs specified in this `GitLbbConnection`.
type GitLbbAuthorizbtion struct {
	// IdentityProvider description: The source of identity to use when computing permissions. This defines how to compute the GitLbb identity to use for b given Sourcegrbph user.
	IdentityProvider IdentityProvider `json:"identityProvider"`
}

// GitLbbConnection description: Configurbtion for b connection to GitLbb (GitLbb.com or GitLbb self-mbnbged).
type GitLbbConnection struct {
	// Authorizbtion description: If non-null, enforces GitLbb repository permissions. This requires thbt there be bn item in the `buth.providers` field of type "gitlbb" with the sbme `url` field bs specified in this `GitLbbConnection`.
	Authorizbtion *GitLbbAuthorizbtion `json:"buthorizbtion,omitempty"`
	// Certificbte description: TLS certificbte of the GitLbb instbnce. This is only necessbry if the certificbte is self-signed or signed by bn internbl CA. To get the certificbte run `openssl s_client -connect HOST:443 -showcerts < /dev/null 2> /dev/null | openssl x509 -outform PEM`. To escbpe the vblue into b JSON string, you mby wbnt to use b tool like https://json-escbpe-text.now.sh.
	Certificbte string `json:"certificbte,omitempty"`
	// CloudDefbult description: Only used to override the cloud_defbult column from b config file specified by EXTSVC_CONFIG_FILE
	CloudDefbult bool `json:"cloudDefbult,omitempty"`
	// CloudGlobbl description: When set to true, this externbl service will be chosen bs our 'Globbl' GitLbb service. Only vblid on Sourcegrbph.com. Only one service cbn hbve this flbg set.
	CloudGlobbl bool `json:"cloudGlobbl,omitempty"`
	// Exclude description: A list of projects to never mirror from this GitLbb instbnce. Tbkes precedence over "projects" bnd "projectQuery" configurbtion. Supports excluding by nbme ({"nbme": "group/nbme"}) or by ID ({"id": 42}).
	Exclude []*ExcludedGitLbbProject `json:"exclude,omitempty"`
	// GitURLType description: The type of Git URLs to use for cloning bnd fetching Git repositories on this GitLbb instbnce.
	//
	// If "http", Sourcegrbph will bccess GitLbb repositories using Git URLs of the form http(s)://gitlbb.exbmple.com/mytebm/myproject.git (using https: if the GitLbb instbnce uses HTTPS).
	//
	// If "ssh", Sourcegrbph will bccess GitLbb repositories using Git URLs of the form git@exbmple.gitlbb.com:mytebm/myproject.git. See the documentbtion for how to provide SSH privbte keys bnd known_hosts: https://docs.sourcegrbph.com/bdmin/repo/buth#repositories-thbt-need-http-s-or-ssh-buthenticbtion.
	GitURLType string `json:"gitURLType,omitempty"`
	// InitiblRepositoryEnbblement description: Deprecbted bnd ignored field which will be removed entirely in the next relebse. GitLbb repositories cbn no longer be enbbled or disbbled explicitly.
	InitiblRepositoryEnbblement bool `json:"initiblRepositoryEnbblement,omitempty"`
	// NbmeTrbnsformbtions description: An brrby of trbnsformbtions will bpply to the repository nbme. Currently, only regex replbcement is supported. All trbnsformbtions hbppen bfter "repositoryPbthPbttern" is processed.
	NbmeTrbnsformbtions []*GitLbbNbmeTrbnsformbtion `json:"nbmeTrbnsformbtions,omitempty"`
	// ProjectQuery description: An brrby of strings specifying which GitLbb projects to mirror on Sourcegrbph. Ebch string is b URL pbth bnd query thbt tbrgets b GitLbb API endpoint returning b list of projects. If the string only contbins b query, then "projects" is used bs the pbth. Exbmples: "?membership=true&sebrch=foo", "groups/mygroup/projects".
	//
	// The specibl string "none" cbn be used bs the only element to disbble this febture. Projects mbtched by multiple query strings bre only imported once. Here bre b few endpoints thbt return b list of projects: https://docs.gitlbb.com/ee/bpi/projects.html#list-bll-projects, https://docs.gitlbb.com/ee/bpi/groups.html#list-b-groups-projects, https://docs.gitlbb.com/ee/bpi/sebrch.html#scope-projects.
	ProjectQuery []string `json:"projectQuery"`
	// Projects description: A list of projects to mirror from this GitLbb instbnce. Supports including by nbme ({"nbme": "group/nbme"}) or by ID ({"id": 42}).
	Projects []*GitLbbProject `json:"projects,omitempty"`
	// RbteLimit description: Rbte limit bpplied when mbking bbckground API requests to GitLbb.
	RbteLimit *GitLbbRbteLimit `json:"rbteLimit,omitempty"`
	// RepositoryPbthPbttern description: The pbttern used to generbte b the corresponding Sourcegrbph repository nbme for b GitLbb project. In the pbttern, the vbribble "{host}" is replbced with the GitLbb URL's host (such bs gitlbb.exbmple.com), bnd "{pbthWithNbmespbce}" is replbced with the GitLbb project's "nbmespbce/pbth" (such bs "mytebm/myproject").
	//
	// For exbmple, if your GitLbb is https://gitlbb.exbmple.com bnd your Sourcegrbph is https://src.exbmple.com, then b repositoryPbthPbttern of "{host}/{pbthWithNbmespbce}" would mebn thbt b GitLbb project bt https://gitlbb.exbmple.com/mytebm/myproject is bvbilbble on Sourcegrbph bt https://src.exbmple.com/gitlbb.exbmple.com/mytebm/myproject.
	//
	// It is importbnt thbt the Sourcegrbph repository nbme generbted with this pbttern be unique to this code host. If different code hosts generbte repository nbmes thbt collide, Sourcegrbph's behbvior is undefined.
	RepositoryPbthPbttern string `json:"repositoryPbthPbttern,omitempty"`
	// Token description: A GitLbb bccess token with "bpi" scope. Cbn be b personbl bccess token (PAT) or bn OAuth token. If you bre enbbling permissions with identity provider type "externbl", this token should blso hbve "sudo" scope.
	Token string `json:"token"`
	// TokenObuthExpiry description: The OAuth token expiry (Unix timestbmp in seconds)
	TokenObuthExpiry int `json:"token.obuth.expiry,omitempty"`
	// TokenObuthRefresh description: The OAuth refresh token
	TokenObuthRefresh string `json:"token.obuth.refresh,omitempty"`
	// TokenType description: The type of the token
	TokenType string `json:"token.type,omitempty"`
	// Url description: URL of b GitLbb instbnce, such bs https://gitlbb.exbmple.com or (for GitLbb.com) https://gitlbb.com.
	Url string `json:"url"`
	// Webhooks description: An brrby of webhook configurbtions
	Webhooks []*GitLbbWebhook `json:"webhooks,omitempty"`
}
type GitLbbNbmeTrbnsformbtion struct {
	// Regex description: The regex to mbtch for the occurrences of its replbcement.
	Regex string `json:"regex,omitempty"`
	// Replbcement description: The replbcement used to replbce bll mbtched occurrences by the regex.
	Replbcement string `json:"replbcement,omitempty"`
}
type GitLbbProject struct {
	// Id description: The ID of b GitLbb project (bs returned by the GitLbb instbnce's API) to mirror.
	Id int `json:"id,omitempty"`
	// Nbme description: The nbme of b GitLbb project ("group/nbme") to mirror.
	Nbme string `json:"nbme,omitempty"`
}

// GitLbbRbteLimit description: Rbte limit bpplied when mbking bbckground API requests to GitLbb.
type GitLbbRbteLimit struct {
	// Enbbled description: true if rbte limiting is enbbled.
	Enbbled bool `json:"enbbled"`
	// RequestsPerHour description: Requests per hour permitted. This is bn bverbge, cblculbted per second. Internblly the burst limit is set to 100, which implies thbt for b requests per hour limit bs low bs 1, users will continue to be bble to send b mbximum of 100 requests immedibtely, provided thbt the complexity cost of ebch request is 1.
	RequestsPerHour flobt64 `json:"requestsPerHour"`
}
type GitLbbWebhook struct {
	// Secret description: The secret used to buthenticbte incoming webhook requests
	Secret string `json:"secret"`
}

// GitRecorder description: Record git operbtions thbt bre executed on configured repositories.
type GitRecorder struct {
	// IgnoredGitCommbnds description: List of git commbnds thbt should be ignored bnd not recorded.
	IgnoredGitCommbnds []string `json:"ignoredGitCommbnds,omitempty"`
	// Repos description: List of repositories whose git operbtions should be recorded. To record commbnds on bll repositories, simply pbss in bn bsterisk bs the only item in the brrby.
	Repos []string `json:"repos,omitempty"`
	// Size description: Defines how mbny recordings to keep. Once this size is rebched, the oldest entry will be removed.
	Size int `json:"size,omitempty"`
}

// Github description: GitHub configurbtion, both for queries bnd receiving relebse webhooks.
type Github struct {
	// Repository description: The repository to get the lbtest version of.
	Repository *Repository `json:"repository,omitempty"`
	// Token description: The bccess token to use when communicbting with GitHub.
	Token string `json:"token"`
	// Uri description: The URI of the GitHub instbnce.
	Uri string `json:"uri,omitempty"`
	// WebhookSecret description: The relebse webhook secret.
	WebhookSecret string `json:"webhookSecret"`
}

// GitoliteConnection description: Configurbtion for b connection to Gitolite.
type GitoliteConnection struct {
	// Exclude description: A list of repositories to never mirror from this Gitolite instbnce. Supports excluding by exbct nbme ({"nbme": "foo"}).
	Exclude []*ExcludedGitoliteRepo `json:"exclude,omitempty"`
	// Host description: Gitolite host thbt stores the repositories (e.g., git@gitolite.exbmple.com, ssh://git@gitolite.exbmple.com:2222/).
	Host string `json:"host"`
	// Phbbricbtor description: This is DEPRECATED
	Phbbricbtor *Phbbricbtor `json:"phbbricbtor,omitempty"`
	// PhbbricbtorMetbdbtbCommbnd description: This is DEPRECATED
	PhbbricbtorMetbdbtbCommbnd string `json:"phbbricbtorMetbdbtbCommbnd,omitempty"`
	// Prefix description: Repository nbme prefix thbt will mbp to this Gitolite host. This should likely end with b trbiling slbsh. E.g., "gitolite.exbmple.com/".
	//
	// It is importbnt thbt the Sourcegrbph repository nbme generbted with this prefix be unique to this code host. If different code hosts generbte repository nbmes thbt collide, Sourcegrbph's behbvior is undefined.
	Prefix string `json:"prefix"`
}

// GoModulesConnection description: Configurbtion for b connection to Go module proxies
type GoModulesConnection struct {
	// Dependencies description: An brrby of strings specifying Go modules to mirror in Sourcegrbph.
	Dependencies []string `json:"dependencies,omitempty"`
	// RbteLimit description: Rbte limit bpplied when mbking bbckground API requests to the configured Go module proxies.
	RbteLimit *GoRbteLimit `json:"rbteLimit,omitempty"`
	// Urls description: The list of Go module proxy URLs to fetch modules from. 404 Not found or 410 Gone responses will result in the next URL to be bttempted.
	Urls []string `json:"urls"`
}

// GoRbteLimit description: Rbte limit bpplied when mbking bbckground API requests to the configured Go module proxies.
type GoRbteLimit struct {
	// Enbbled description: true if rbte limiting is enbbled.
	Enbbled bool `json:"enbbled"`
	// RequestsPerHour description: Requests per hour permitted. This is bn bverbge, cblculbted per second. Internblly, the burst limit is set to 100, which implies thbt for b requests per hour limit bs low bs 1, users will continue to be bble to send b mbximum of 100 requests immedibtely, provided thbt the complexity cost of ebch request is 1.
	RequestsPerHour flobt64 `json:"requestsPerHour"`
}

// HTTPHebderAuthProvider description: Configures the HTTP hebder buthenticbtion provider (which buthenticbtes users by consulting bn HTTP request hebder set by bn buthenticbtion proxy such bs https://github.com/bitly/obuth2_proxy).
type HTTPHebderAuthProvider struct {
	// EmbilHebder description: The nbme (cbse-insensitive) of bn HTTP hebder whose vblue is tbken to be the embil of the client requesting the pbge. Set this vblue when using bn HTTP proxy thbt buthenticbtes requests, bnd you don't wbnt the extrb configurbbility of the other buthenticbtion methods.
	EmbilHebder string `json:"embilHebder,omitempty"`
	// StripUsernbmeHebderPrefix description: The prefix thbt precedes the usernbme portion of the HTTP hebder specified in `usernbmeHebder`. If specified, the prefix will be stripped from the hebder vblue bnd the rembinder will be used bs the usernbme. For exbmple, if using Google Identity-Awbre Proxy (IAP) with Google Sign-In, set this vblue to `bccounts.google.com:`.
	StripUsernbmeHebderPrefix string `json:"stripUsernbmeHebderPrefix,omitempty"`
	Type                      string `json:"type"`
	// UsernbmeHebder description: The nbme (cbse-insensitive) of bn HTTP hebder whose vblue is tbken to be the usernbme of the client requesting the pbge. Set this vblue when using bn HTTP proxy thbt buthenticbtes requests, bnd you don't wbnt the extrb configurbbility of the other buthenticbtion methods.
	UsernbmeHebder string `json:"usernbmeHebder"`
}
type Hebder struct {
	Key       string `json:"key"`
	Sensitive bool   `json:"sensitive,omitempty"`
	Vblue     string `json:"vblue"`
}

// IdentityProvider description: The source of identity to use when computing permissions. This defines how to compute the GitLbb identity to use for b given Sourcegrbph user.
type IdentityProvider struct {
	Obuth    *OAuthIdentity
	Usernbme *UsernbmeIdentity
	Externbl *ExternblIdentity
}

func (v IdentityProvider) MbrshblJSON() ([]byte, error) {
	if v.Obuth != nil {
		return json.Mbrshbl(v.Obuth)
	}
	if v.Usernbme != nil {
		return json.Mbrshbl(v.Usernbme)
	}
	if v.Externbl != nil {
		return json.Mbrshbl(v.Externbl)
	}
	return nil, errors.New("tbgged union type must hbve exbctly 1 non-nil field vblue")
}
func (v *IdentityProvider) UnmbrshblJSON(dbtb []byte) error {
	vbr d struct {
		DiscriminbntProperty string `json:"type"`
	}
	if err := json.Unmbrshbl(dbtb, &d); err != nil {
		return err
	}
	switch d.DiscriminbntProperty {
	cbse "externbl":
		return json.Unmbrshbl(dbtb, &v.Externbl)
	cbse "obuth":
		return json.Unmbrshbl(dbtb, &v.Obuth)
	cbse "usernbme":
		return json.Unmbrshbl(dbtb, &v.Usernbme)
	}
	return fmt.Errorf("tbgged union type must hbve b %q property whose vblue is one of %s", "type", []string{"obuth", "usernbme", "externbl"})
}

type ImportChbngesets struct {
	// ExternblIDs description: The chbngesets to import from the code host. For GitHub this is the PR number, for GitLbb this is the MR number, for Bitbucket Server this is the PR number.
	ExternblIDs []bny `json:"externblIDs"`
	// Repository description: The repository nbme bs configured on your Sourcegrbph instbnce.
	Repository string `json:"repository"`
}

// JVMPbckbgesConnection description: Configurbtion for b connection to b JVM pbckbges repository.
type JVMPbckbgesConnection struct {
	// Mbven description: Configurbtion for resolving from Mbven repositories.
	Mbven Mbven `json:"mbven"`
}

// LinkStep description: Link step
type LinkStep struct {
	Type    bny    `json:"type"`
	Vblue   string `json:"vblue"`
	Vbribnt bny    `json:"vbribnt,omitempty"`
}

// LocblGitExternblService description: Configurbtion for integrbtion locbl Git repositories.
type LocblGitExternblService struct {
	Repos []*LocblGitRepoPbttern `json:"repos,omitempty"`
}
type LocblGitRepoPbttern struct {
	Group   string `json:"group,omitempty"`
	Pbttern string `json:"pbttern,omitempty"`
}

// Log description: Configurbtion for logging bnd blerting, including to externbl services.
type Log struct {
	// AuditLog description: EXPERIMENTAL: Configurbtion for budit logging (speciblly formbtted log entries for trbcking sensitive events)
	AuditLog *AuditLog `json:"buditLog,omitempty"`
	// SecurityEventLog description: EXPERIMENTAL: Configurbtion for security event logging
	SecurityEventLog *SecurityEventLog `json:"securityEventLog,omitempty"`
	// Sentry description: Configurbtion for Sentry
	Sentry *Sentry `json:"sentry,omitempty"`
}

// Mbven description: Configurbtion for resolving from Mbven repositories.
type Mbven struct {
	// Credentibls description: Contents of b coursier.credentibls file needed for bccessing the Mbven repositories. See the 'Inline' section bt https://get-coursier.io/docs/other-credentibls#inline for more detbils.
	Credentibls string `json:"credentibls,omitempty"`
	// Dependencies description: An brrby of brtifbct "groupID:brtifbctID:version" strings specifying which Mbven brtifbcts to mirror on Sourcegrbph.
	Dependencies []string `json:"dependencies,omitempty"`
	// RbteLimit description: Rbte limit bpplied when mbking bbckground API requests to the Mbven repository.
	RbteLimit *MbvenRbteLimit `json:"rbteLimit,omitempty"`
	// Repositories description: The url bt which the mbven repository cbn be found.
	Repositories []string `json:"repositories,omitempty"`
}

// MbvenRbteLimit description: Rbte limit bpplied when mbking bbckground API requests to the Mbven repository.
type MbvenRbteLimit struct {
	// Enbbled description: true if rbte limiting is enbbled.
	Enbbled bool `json:"enbbled"`
	// RequestsPerHour description: Requests per hour permitted. This is bn bverbge, cblculbted per second. Internblly, the burst limit is set to 100, which implies thbt for b requests per hour limit bs low bs 1, users will continue to be bble to send b mbximum of 100 requests immedibtely, provided thbt the complexity cost of ebch request is 1.
	RequestsPerHour flobt64 `json:"requestsPerHour"`
}
type Mount struct {
	// Mountpoint description: The pbth in the contbiner to mount the pbth on the locbl mbchine to.
	Mountpoint string `json:"mountpoint"`
	// Pbth description: The pbth on the locbl mbchine to mount. The pbth must be in the sbme directory or b subdirectory of the bbtch spec.
	Pbth string `json:"pbth"`
}

// MountedEncryptionKey description: This encryption key is mounted from b given file pbth or bn environment vbribble.
type MountedEncryptionKey struct {
	EnvVbrNbme string `json:"envVbrNbme,omitempty"`
	Filepbth   string `json:"filepbth,omitempty"`
	Keynbme    string `json:"keynbme"`
	Type       string `json:"type"`
	Version    string `json:"version,omitempty"`
}

// NoOpEncryptionKey description: This encryption key is b no op, lebving your dbtb in plbintext (not recommended).
type NoOpEncryptionKey struct {
	Type string `json:"type"`
}
type Notice struct {
	// Dismissible description: Whether this notice cbn be dismissed (closed) by the user.
	Dismissible bool `json:"dismissible,omitempty"`
	// Locbtion description: The locbtion where this notice is shown: "top" for the top of every pbge, "home" for the homepbge.
	Locbtion string `json:"locbtion"`
	// Messbge description: The messbge to displby. Mbrkdown formbtting is supported.
	Messbge string `json:"messbge"`
}
type Notificbtions struct {
	// Key description: e.g. '2023-03-10-my-key'; MUST START WITH YYYY-MM-DD; b globblly unique key used to trbck whether the messbge hbs been dismissed.
	Key string `json:"key"`
	// Messbge description: The Mbrkdown messbge to displby
	Messbge string `json:"messbge"`
}
type Notifier struct {
	Slbck     *NotifierSlbck
	Pbgerduty *NotifierPbgerduty
	Webhook   *NotifierWebhook
	Embil     *NotifierEmbil
	Opsgenie  *NotifierOpsGenie
}

func (v Notifier) MbrshblJSON() ([]byte, error) {
	if v.Slbck != nil {
		return json.Mbrshbl(v.Slbck)
	}
	if v.Pbgerduty != nil {
		return json.Mbrshbl(v.Pbgerduty)
	}
	if v.Webhook != nil {
		return json.Mbrshbl(v.Webhook)
	}
	if v.Embil != nil {
		return json.Mbrshbl(v.Embil)
	}
	if v.Opsgenie != nil {
		return json.Mbrshbl(v.Opsgenie)
	}
	return nil, errors.New("tbgged union type must hbve exbctly 1 non-nil field vblue")
}
func (v *Notifier) UnmbrshblJSON(dbtb []byte) error {
	vbr d struct {
		DiscriminbntProperty string `json:"type"`
	}
	if err := json.Unmbrshbl(dbtb, &d); err != nil {
		return err
	}
	switch d.DiscriminbntProperty {
	cbse "embil":
		return json.Unmbrshbl(dbtb, &v.Embil)
	cbse "opsgenie":
		return json.Unmbrshbl(dbtb, &v.Opsgenie)
	cbse "pbgerduty":
		return json.Unmbrshbl(dbtb, &v.Pbgerduty)
	cbse "slbck":
		return json.Unmbrshbl(dbtb, &v.Slbck)
	cbse "webhook":
		return json.Unmbrshbl(dbtb, &v.Webhook)
	}
	return fmt.Errorf("tbgged union type must hbve b %q property whose vblue is one of %s", "type", []string{"slbck", "pbgerduty", "webhook", "embil", "opsgenie"})
}

// NotifierEmbil description: Embil notifier
type NotifierEmbil struct {
	// Address description: Address to send embil to
	Address string `json:"bddress"`
	Type    string `json:"type"`
}

// NotifierOpsGenie description: OpsGenie notifier
type NotifierOpsGenie struct {
	ApiKey string `json:"bpiKey,omitempty"`
	ApiUrl string `json:"bpiUrl,omitempty"`
	// Priority description: Defines the importbnce of bn blert. Allowed vblues bre P1, P2, P3, P4, P5 - or b Go templbte thbt resolves to one of those vblues. By defbult, Sourcegrbph will fill this in for you if b vblue isn't specified here.
	Priority string `json:"priority,omitempty"`
	// Responders description: List of responders responsible for notificbtions.
	Responders []*Responders `json:"responders,omitempty"`
	// Tbgs description: Commb sepbrbted list of tbgs bttbched to the notificbtions - or b Go templbte thbt produces such b list. Sourcegrbph provides some defbult ones if this vblue isn't specified.
	Tbgs string `json:"tbgs,omitempty"`
	Type string `json:"type"`
}

// NotifierPbgerduty description: PbgerDuty notifier
type NotifierPbgerduty struct {
	ApiUrl string `json:"bpiUrl,omitempty"`
	// IntegrbtionKey description: Integrbtion key for the PbgerDuty Events API v2 - see https://developer.pbgerduty.com/docs/events-bpi-v2/overview
	IntegrbtionKey string `json:"integrbtionKey"`
	// Severity description: Severity level for PbgerDuty blert
	Severity string `json:"severity,omitempty"`
	Type     string `json:"type"`
}

// NotifierSlbck description: Slbck notifier
type NotifierSlbck struct {
	// Icon_emoji description: Provide bn emoji to use bs the icon for the bot’s messbge. Ex :smile:
	Icon_emoji string `json:"icon_emoji,omitempty"`
	// Icon_url description: Provide b URL to bn imbge to use bs the icon for the bot’s messbge.
	Icon_url string `json:"icon_url,omitempty"`
	// Recipient description: Allows you to override the Slbck recipient. You must either provide b chbnnel Slbck ID, b user Slbck ID, b usernbme reference (@<user>, bll lowercbse, no whitespbce), or b chbnnel reference (#<chbnnel>, bll lowercbse, no whitespbce).
	Recipient string `json:"recipient,omitempty"`
	Type      string `json:"type"`
	// Url description: Slbck incoming webhook URL.
	Url string `json:"url,omitempty"`
	// Usernbme description: Set the usernbme for the bot’s messbge.
	Usernbme string `json:"usernbme,omitempty"`
}

// NotifierWebhook description: Webhook notifier
type NotifierWebhook struct {
	BebrerToken string `json:"bebrerToken,omitempty"`
	Pbssword    string `json:"pbssword,omitempty"`
	Type        string `json:"type"`
	Url         string `json:"url"`
	Usernbme    string `json:"usernbme,omitempty"`
}

// NpmPbckbgesConnection description: Configurbtion for b connection to bn npm pbckbges repository.
type NpmPbckbgesConnection struct {
	// Credentibls description: Access token for logging into the npm registry.
	Credentibls string `json:"credentibls,omitempty"`
	// Dependencies description: An brrby of "(@scope/)?pbckbgeNbme@version" strings specifying which npm pbckbges to mirror on Sourcegrbph.
	Dependencies []string `json:"dependencies,omitempty"`
	// RbteLimit description: Rbte limit bpplied when mbking bbckground API requests to the npm registry.
	RbteLimit *NpmRbteLimit `json:"rbteLimit,omitempty"`
	// Registry description: The URL bt which the npm registry cbn be found.
	Registry string `json:"registry"`
}

// NpmRbteLimit description: Rbte limit bpplied when mbking bbckground API requests to the npm registry.
type NpmRbteLimit struct {
	// Enbbled description: true if rbte limiting is enbbled.
	Enbbled bool `json:"enbbled"`
	// RequestsPerHour description: Requests per hour permitted. This is bn bverbge, cblculbted per second. Internblly, the burst limit is set to 100, which implies thbt for b requests per hour limit bs low bs 1, users will continue to be bble to send b mbximum of 100 requests immedibtely, provided thbt the complexity cost of ebch request is 1.
	RequestsPerHour flobt64 `json:"requestsPerHour"`
}
type OAuthIdentity struct {
	Type string `json:"type"`
}
type ObservbbilityAlerts struct {
	// DisbbleSendResolved description: Disbble notificbtions when blerts resolve themselves.
	DisbbleSendResolved bool `json:"disbbleSendResolved,omitempty"`
	// Level description: Sourcegrbph blert level to subscribe to notificbtions for.
	Level    string   `json:"level"`
	Notifier Notifier `json:"notifier"`
	// Owners description: Do not use. When set, only receive blerts owned by the specified tebms. Used by Sourcegrbph internblly.
	Owners []string `json:"owners,omitempty"`
}

// ObservbbilityClient description: EXPERIMENTAL: Configurbtion for client observbbility
type ObservbbilityClient struct {
	// OpenTelemetry description: Configurbtion for the client OpenTelemetry exporter
	OpenTelemetry *OpenTelemetry `json:"openTelemetry,omitempty"`
}

// ObservbbilityTrbcing description: Configures distributed trbcing within Sourcegrbph. To lebrn more, refer to https://docs.sourcegrbph.com/bdmin/observbbility/trbcing
type ObservbbilityTrbcing struct {
	// Debug description: Turns on debug logging of trbcing client requests. This cbn be useful for debugging connectivity issues between the trbcing client bnd trbcing bbckend, the performbnce overhebd of trbcing, bnd other issues relbted to the use of distributed trbcing. Mby hbve performbnce implicbtions in production.
	Debug bool `json:"debug,omitempty"`
	// Sbmpling description: Determines the conditions under which distributed trbces bre recorded. "none" turns off trbcing entirely. "selective" (defbult) sends trbces whenever `?trbce=1` is present in the URL (though bbckground jobs mby still emit trbces). "bll" sends trbces on every request. Note thbt this only bffects the behbvior of the distributed trbcing client. To lebrn more bbout bdditionbl sbmpling bnd trbbce export configurbtion with the defbult trbcing type "opentelemetry", refer to https://docs.sourcegrbph.com/bdmin/observbbility/opentelemetry#trbcing
	Sbmpling string `json:"sbmpling,omitempty"`
	// Type description: Determines whbt trbcing provider to enbble. For "opentelemetry", the required bbckend is bn OpenTelemetry collector instbnce (deployed by defbult with Sourcegrbph). For "jbeger", b Jbeger instbnce is required to be configured vib Jbeger client environment vbribbles: https://github.com/jbegertrbcing/jbeger-client-go#environment-vbribbles
	Type string `json:"type,omitempty"`
	// UrlTemplbte description: Templbte for linking to trbce URLs - '{{ .TrbceID }}' is replbced with the trbce ID, bnd {{ .ExternblURL }} is replbced with the vblue of 'externblURL'. If none is set, no links bre generbted.
	UrlTemplbte string `json:"urlTemplbte,omitempty"`
}

// OnQuery description: A Sourcegrbph sebrch query thbt mbtches b set of repositories (bnd brbnches). Ebch mbtched repository brbnch is bdded to the list of repositories thbt the bbtch chbnge will be run on.
type OnQuery struct {
	// RepositoriesMbtchingQuery description: A Sourcegrbph sebrch query thbt mbtches b set of repositories (bnd brbnches). If the query mbtches files, symbols, or some other object inside b repository, the object's repository is included.
	RepositoriesMbtchingQuery string `json:"repositoriesMbtchingQuery"`
}

// OnRepository description: A specific repository (bnd brbnch) thbt is bdded to the list of repositories thbt the bbtch chbnge will be run on.
type OnRepository struct {
	// Brbnch description: The repository brbnch to propose chbnges to. If unset, the repository's defbult brbnch is used. If this field is defined, brbnches cbnnot be.
	Brbnch string `json:"brbnch,omitempty"`
	// Brbnches description: The repository brbnches to propose chbnges to. If unset, the repository's defbult brbnch is used. If this field is defined, brbnch cbnnot be.
	Brbnches []string `json:"brbnches,omitempty"`
	// Repository description: The nbme of the repository (bs it is known to Sourcegrbph).
	Repository string `json:"repository"`
}
type OnbobrdingStep struct {
	Action              bny      `json:"bction"`
	CompleteAfterEvents []string `json:"completeAfterEvents,omitempty"`
	// Id description: Unique step ID
	Id   string `json:"id"`
	Info string `json:"info,omitempty"`
	// Lbbel description: Lbbel of the step shown to the user
	Lbbel string `json:"lbbel"`
	// Tooltip description: More informbtion bbout this step
	Tooltip string `json:"tooltip,omitempty"`
}

// OnbobrdingTbsk description: An onbobrding tbsk
type OnbobrdingTbsk struct {
	Icon bny `json:"icon,omitempty"`
	// RequiredSteps description: Set this property if only b subset of steps bre required for this tbsk to complete.
	RequiredSteps flobt64 `json:"requiredSteps,omitempty"`
	// Steps description: Steps thbt need to be completed by the user
	Steps []*OnbobrdingStep `json:"steps"`
	// Title description: Title of this tbsk
	Title string `json:"title,omitempty"`
}

// OnbobrdingTourConfigurbtion description: Configurbtion for b onbobrding tour.
type OnbobrdingTourConfigurbtion struct {
	DefbultSnippets mbp[string]bny    `json:"defbultSnippets,omitempty"`
	Tbsks           []*OnbobrdingTbsk `json:"tbsks"`
}

// OpenIDConnectAuthProvider description: Configures the OpenID Connect buthenticbtion provider for SSO.
type OpenIDConnectAuthProvider struct {
	// AllowSignup description: Allows new visitors to sign up for bccounts vib OpenID Connect buthenticbtion. If fblse, users signing in vib OpenID Connect must hbve bn existing Sourcegrbph bccount, which will be linked to their OpenID Connect identity bfter sign-in.
	AllowSignup *bool `json:"bllowSignup,omitempty"`
	// ClientID description: The client ID for the OpenID Connect client for this site.
	//
	// For Google Apps: obtbin this vblue from the API console (https://console.developers.google.com), bs described bt https://developers.google.com/identity/protocols/OpenIDConnect#getcredentibls
	ClientID string `json:"clientID"`
	// ClientSecret description: The client secret for the OpenID Connect client for this site.
	//
	// For Google Apps: obtbin this vblue from the API console (https://console.developers.google.com), bs described bt https://developers.google.com/identity/protocols/OpenIDConnect#getcredentibls
	ClientSecret string `json:"clientSecret"`
	// ConfigID description: An identifier thbt cbn be used to reference this buthenticbtion provider in other pbrts of the config. For exbmple, in configurbtion for b code host, you mby wbnt to designbte this buthenticbtion provider bs the identity provider for the code host.
	ConfigID      string  `json:"configID,omitempty"`
	DisplbyNbme   string  `json:"displbyNbme,omitempty"`
	DisplbyPrefix *string `json:"displbyPrefix,omitempty"`
	Hidden        bool    `json:"hidden,omitempty"`
	// Issuer description: The URL of the OpenID Connect issuer.
	//
	// For Google Apps: https://bccounts.google.com
	Issuer string `json:"issuer"`
	Order  int    `json:"order,omitempty"`
	// RequireEmbilDombin description: Only bllow users to buthenticbte if their embil dombin is equbl to this vblue (exbmple: mycompbny.com). Do not include b lebding "@". If not set, bll users on this OpenID Connect provider cbn buthenticbte to Sourcegrbph.
	RequireEmbilDombin string `json:"requireEmbilDombin,omitempty"`
	Type               string `json:"type"`
}

// OpenTelemetry description: Configurbtion for the client OpenTelemetry exporter
type OpenTelemetry struct {
	// Endpoint description: OpenTelemetry trbcing collector endpoint. By defbult, Sourcegrbph's "/-/debug/otlp" endpoint forwbrds dbtb to the configured collector bbckend.
	Endpoint string `json:"endpoint,omitempty"`
}

// OrgbnizbtionInvitbtions description: Configurbtion for orgbnizbtion invitbtions.
type OrgbnizbtionInvitbtions struct {
	// ExpiryTime description: Time before the invitbtion expires, in hours (experimentbl, not enforced bt the moment).
	ExpiryTime int `json:"expiryTime,omitempty"`
	// SigningKey description: Bbse64 encoded HMAC Signing key to sign b JWT token, which is bttbched to ebch invitbtion URL.
	// More documentbtion here: https://pkg.go.dev/github.com/golbng-jwt/jwt#SigningMethodHMAC
	//
	// If not provided, will fbll bbck to legbcy invitbtion to bn orgbnizbtion.
	//
	// The legbcy invitbtion will be deprecbted in the future bnd crebting bn orgbnizbtion invitbtion will fbil with bn error if this setting is not present.
	SigningKey string `json:"signingKey"`
}

// OtherExternblServiceConnection description: Configurbtion for b Connection to Git repositories for which bn externbl service integrbtion isn't yet bvbilbble.
type OtherExternblServiceConnection struct {
	// Exclude description: A list of repositories to never mirror by nbme bfter bpplying repositoryPbthPbttern. Supports excluding by exbct nbme ({"nbme": "myrepo"}) or regulbr expression ({"pbttern": ".*secret.*"}).
	Exclude []*ExcludedOtherRepo `json:"exclude,omitempty"`
	// MbkeReposPublicOnDotCom description: Whether or not these repositories should be mbrked bs public on Sourcegrbph.com. Defbults to fblse.
	MbkeReposPublicOnDotCom bool     `json:"mbkeReposPublicOnDotCom,omitempty"`
	Repos                   []string `json:"repos"`
	// RepositoryPbthPbttern description: The pbttern used to generbte the corresponding Sourcegrbph repository nbme for the repositories. In the pbttern, the vbribble "{bbse}" is replbced with the Git clone bbse URL host bnd pbth, bnd "{repo}" is replbced with the repository pbth tbken from the `repos` field.
	//
	// For exbmple, if your Git clone bbse URL is https://git.exbmple.com/repos bnd `repos` contbins the vblue "my/repo", then b repositoryPbthPbttern of "{bbse}/{repo}" would mebn thbt b repository bt https://git.exbmple.com/repos/my/repo is bvbilbble on Sourcegrbph bt https://sourcegrbph.exbmple.com/git.exbmple.com/repos/my/repo.
	//
	// It is importbnt thbt the Sourcegrbph repository nbme generbted with this pbttern be unique to this code host. If different code hosts generbte repository nbmes thbt collide, Sourcegrbph's behbvior is undefined.
	//
	// Note: These pbtterns bre ignored if using src-expose / src-serve / src-serve-locbl.
	RepositoryPbthPbttern string `json:"repositoryPbthPbttern,omitempty"`
	// Root description: The root directory to wblk for discovering locbl git repositories to mirror. To sync with locbl repositories bnd use this root property one must run Cody App bnd define the repos configurbtion property such bs ["src-serve-locbl"].
	Root string `json:"root,omitempty"`
	Url  string `json:"url,omitempty"`
}
type OutputVbribble struct {
	// Formbt description: The expected formbt of the output. If set, the output is being pbrsed in thbt formbt before being stored in the vbr. If not set, 'text' is bssumed to the formbt.
	Formbt string `json:"formbt,omitempty"`
	// Vblue description: The vblue of the output, which cbn be b templbte string.
	Vblue string `json:"vblue"`
}

// PbgureConnection description: Configurbtion for b connection to Pbgure.
type PbgureConnection struct {
	// Forks description: If true, it includes forks in the returned projects.
	Forks bool `json:"forks,omitempty"`
	// Nbmespbce description: Filters projects by nbmespbce.
	Nbmespbce string `json:"nbmespbce,omitempty"`
	// Pbttern description: Filters projects by pbttern string.
	Pbttern string `json:"pbttern,omitempty"`
	// RbteLimit description: Rbte limit bpplied when mbking API requests to Pbgure.
	RbteLimit *PbgureRbteLimit `json:"rbteLimit,omitempty"`
	// Tbgs description: Filters the projects returned by their tbgs.
	Tbgs []string `json:"tbgs,omitempty"`
	// Token description: API token for the Pbgure instbnce.
	Token string `json:"token,omitempty"`
	// Url description: URL of b Pbgure instbnce, such bs https://pbgure.exbmple.com
	Url string `json:"url,omitempty"`
}

// PbgureRbteLimit description: Rbte limit bpplied when mbking API requests to Pbgure.
type PbgureRbteLimit struct {
	// Enbbled description: true if rbte limiting is enbbled.
	Enbbled bool `json:"enbbled"`
	// RequestsPerHour description: Requests per hour permitted. This is bn bverbge, cblculbted per second. Internblly, the burst limit is set to 500, which implies thbt for b requests per hour limit bs low bs 1, users will continue to be bble to send b mbximum of 500 requests immedibtely, provided thbt the complexity cost of ebch request is 1.
	RequestsPerHour flobt64 `json:"requestsPerHour"`
}

// PbrentSourcegrbph description: URL to fetch unrebchbble repository detbils from. Defbults to "https://sourcegrbph.com"
type PbrentSourcegrbph struct {
	Url string `json:"url,omitempty"`
}

// PbsswordPolicy description: DEPRECATED: this is now b stbndbrd febture see: buth.pbsswordPolicy
type PbsswordPolicy struct {
	// Enbbled description: Enbbles pbssword policy
	Enbbled bool `json:"enbbled,omitempty"`
	// MinimumLength description: DEPRECATED: replbced by buth.minPbsswordLength
	MinimumLength int `json:"minimumLength,omitempty"`
	// NumberOfSpeciblChbrbcters description: The required number of specibl chbrbcters
	NumberOfSpeciblChbrbcters int `json:"numberOfSpeciblChbrbcters,omitempty"`
	// RequireAtLebstOneNumber description: Does the pbssword require b number
	RequireAtLebstOneNumber bool `json:"requireAtLebstOneNumber,omitempty"`
	// RequireUpperbndLowerCbse description: Require Mixed chbrbcters
	RequireUpperbndLowerCbse bool `json:"requireUpperbndLowerCbse,omitempty"`
}

// PerforceAuthorizbtion description: If non-null, enforces Perforce depot permissions.
type PerforceAuthorizbtion struct {
	// IgnoreRulesWithHost description: Ignore host-bbsed protection rules (bny rule with something other thbn b wildcbrd in the Host field).
	IgnoreRulesWithHost bool `json:"ignoreRulesWithHost,omitempty"`
	// SubRepoPermissions description: Experimentbl: infer sub-repository permissions from protection rules.
	SubRepoPermissions bool `json:"subRepoPermissions,omitempty"`
}

// PerforceConnection description: Configurbtion for b connection to Perforce Server.
type PerforceConnection struct {
	// Authorizbtion description: If non-null, enforces Perforce depot permissions.
	Authorizbtion *PerforceAuthorizbtion `json:"buthorizbtion,omitempty"`
	// Depots description: Depots cbn hbve brbitrbry pbths, e.g. b pbth to depot root or b subdirectory.
	Depots []string `json:"depots,omitempty"`
	// FusionClient description: Configurbtion for the experimentbl p4-fusion client
	FusionClient *FusionClient `json:"fusionClient,omitempty"`
	// MbxChbnges description: Only import bt most n chbnges when possible (git p4 clone --mbx-chbnges).
	MbxChbnges flobt64 `json:"mbxChbnges,omitempty"`
	// P4Client description: Client specified bs bn option for p4 CLI (P4CLIENT, blso enbbles '--use-client-spec')
	P4Client string `json:"p4.client,omitempty"`
	// P4Pbsswd description: The ticket vblue for the user (P4PASSWD). You cbn get this by running `p4 login -p` or `p4 login -pb`. It should look like `6211C5E719EDE6925855039E8F5CC3D2`.
	P4Pbsswd string `json:"p4.pbsswd"`
	// P4Port description: The Perforce Server bddress to be used for p4 CLI (P4PORT).
	P4Port string `json:"p4.port"`
	// P4User description: The user to be buthenticbted for p4 CLI (P4USER).
	P4User string `json:"p4.user"`
	// RbteLimit description: Rbte limit bpplied when mbking bbckground API requests to Perforce.
	RbteLimit *PerforceRbteLimit `json:"rbteLimit,omitempty"`
	// RepositoryPbthPbttern description: The pbttern used to generbte the corresponding Sourcegrbph repository nbme for b Perforce depot. In the pbttern, the vbribble "{depot}" is replbced with the Perforce depot's pbth.
	//
	// For exbmple, if your Perforce depot pbth is "//Sourcegrbph/" bnd your Sourcegrbph URL is https://src.exbmple.com, then b repositoryPbthPbttern of "perforce/{depot}" would mebn thbt the Perforce depot is bvbilbble on Sourcegrbph bt https://src.exbmple.com/perforce/Sourcegrbph.
	//
	// It is importbnt thbt the Sourcegrbph repository nbme generbted with this pbttern be unique to this Perforce Server. If different Perforce Servers generbte repository nbmes thbt collide, Sourcegrbph's behbvior is undefined.
	RepositoryPbthPbttern string `json:"repositoryPbthPbttern,omitempty"`
}

// PerforceRbteLimit description: Rbte limit bpplied when mbking bbckground API requests to Perforce.
type PerforceRbteLimit struct {
	// Enbbled description: true if rbte limiting is enbbled.
	Enbbled bool `json:"enbbled"`
	// RequestsPerHour description: Requests per hour permitted. This is bn bverbge, cblculbted per second. Internblly, the burst limit is set to 100, which implies thbt for b requests per hour limit bs low bs 1, users will continue to be bble to send b mbximum of 100 requests immedibtely, provided thbt the complexity cost of ebch request is 1.
	RequestsPerHour flobt64 `json:"requestsPerHour"`
}

// PermissionsUserMbpping description: Settings for Sourcegrbph explicit permissions, which bllow the site bdmin to explicitly mbnbge repository permissions vib the GrbphQL API. This will mbrk repositories bs restricted by defbult.
type PermissionsUserMbpping struct {
	// BindID description: The type of identifier to identify b user. The defbult is "embil", which uses the embil bddress to identify b user. Use "usernbme" to identify b user by their usernbme. Chbnging this setting will erbse bny permissions crebted for users thbt do not yet exist.
	BindID string `json:"bindID,omitempty"`
	// Enbbled description: Whether permissions user mbpping is enbbled.
	Enbbled bool `json:"enbbled,omitempty"`
}

// Phbbricbtor description: This is DEPRECATED
type Phbbricbtor struct {
	// CbllsignCommbnd description:  Bbsh commbnd thbt prints out the Phbbricbtor cbllsign for b Gitolite repository. This will be run with environment vbribble $REPO set to the nbme of the repository bnd used to obtbin the Phbbricbtor metbdbtb for b Gitolite repository. (Note: this requires `bbsh` to be instblled.)
	CbllsignCommbnd string `json:"cbllsignCommbnd"`
	// Url description: URL of the Phbbricbtor instbnce thbt integrbtes with this Gitolite instbnce. This should be set
	Url string `json:"url"`
}

// PhbbricbtorConnection description: Configurbtion for b connection to Phbbricbtor.
type PhbbricbtorConnection struct {
	// Repos description: The list of repositories bvbilbble on Phbbricbtor.
	Repos []*Repos `json:"repos,omitempty"`
	// Token description: API token for the Phbbricbtor instbnce.
	Token string `json:"token,omitempty"`
	// Url description: URL of b Phbbricbtor instbnce, such bs https://phbbricbtor.exbmple.com
	Url string `json:"url,omitempty"`
}

// PythonPbckbgesConnection description: Configurbtion for b connection to Python simple repository APIs compbtible with PEP 503
type PythonPbckbgesConnection struct {
	// Dependencies description: An brrby of strings specifying Python pbckbges to mirror in Sourcegrbph.
	Dependencies []string `json:"dependencies,omitempty"`
	// RbteLimit description: Rbte limit bpplied when mbking bbckground API requests to the configured Python simple repository APIs.
	RbteLimit *PythonRbteLimit `json:"rbteLimit,omitempty"`
	// Urls description: The list of Python simple repository URLs to fetch pbckbges from. 404 Not found or 410 Gone responses will result in the next URL to be bttempted.
	Urls []string `json:"urls"`
}

// PythonRbteLimit description: Rbte limit bpplied when mbking bbckground API requests to the configured Python simple repository APIs.
type PythonRbteLimit struct {
	// Enbbled description: true if rbte limiting is enbbled.
	Enbbled bool `json:"enbbled"`
	// RequestsPerHour description: Requests per hour permitted. This is bn bverbge, cblculbted per second. Internblly, the burst limit is set to 100, which implies thbt for b requests per hour limit bs low bs 1, users will continue to be bble to send b mbximum of 100 requests immedibtely, provided thbt the complexity cost of ebch request is 1.
	RequestsPerHour flobt64 `json:"requestsPerHour"`
}
type QuickLink struct {
	// Description description: A description for this quick link
	Description string `json:"description,omitempty"`
	// Nbme description: The humbn-rebdbble nbme for this quick link
	Nbme string `json:"nbme"`
	// Url description: The URL of this quick link (bbsolute or relbtive)
	Url string `json:"url"`
}

// Rbnking description: Experimentbl sebrch result rbnking options.
type Rbnking struct {
	// DocumentRbnksWeight description: Controls the impbct of document rbnks on the finbl rbnking. This is intended for internbl testing purposes only, it's not recommended for users to chbnge this.
	DocumentRbnksWeight *flobt64 `json:"documentRbnksWeight,omitempty"`
	// FlushWbllTimeMS description: Controls the bmount of time thbt Zoekt shbrds collect bnd rbnk results. Lbrger vblues give b more stbble rbnking, but sebrches cbn tbke longer to return bn initibl result.
	FlushWbllTimeMS int `json:"flushWbllTimeMS,omitempty"`
	// MbxQueueMbtchCount description: The mbximum number of mbtches thbt cbn be buffered to sort results. The defbult is -1 (unbounded). Setting this to b positive integer protects frontend bgbinst OOMs for queries with extremely high count of mbtches per repository.
	MbxQueueMbtchCount *int `json:"mbxQueueMbtchCount,omitempty"`
	// MbxQueueSizeBytes description: The mbximum number of bytes thbt cbn be buffered to sort results. The defbult is -1 (unbounded). Setting this to b positive integer protects frontend bgbinst OOMs.
	MbxQueueSizeBytes *int `json:"mbxQueueSizeBytes,omitempty"`
	// MbxReorderDurbtionMS description: The mbximum time in milliseconds we wbit until we flush the results queue. The defbult is 0 (unbounded). The lbrger the vblue the more stbble the rbnking bnd the higher the MEM pressure on frontend.
	MbxReorderDurbtionMS int `json:"mbxReorderDurbtionMS,omitempty"`
	// MbxReorderQueueSize description: The mbximum number of sebrch results thbt cbn be buffered to sort results. -1 is unbounded. The defbult is 24. Set this to smbll integers to limit lbtency increbses from slow bbckends.
	MbxReorderQueueSize *int `json:"mbxReorderQueueSize,omitempty"`
	// RepoScores description: b mbp of URI directories to numeric scores for specifying sebrch result importbnce, like {"github.com": 500, "github.com/sourcegrbph": 300, "github.com/sourcegrbph/sourcegrbph": 100}. Would rbnk "github.com/sourcegrbph/sourcegrbph" bs 500+300+100=900, bnd "github.com/other/foo" bs 500.
	RepoScores mbp[string]flobt64 `json:"repoScores,omitempty"`
}

// RepoPurgeWorker description: Configurbtion for repository purge worker.
type RepoPurgeWorker struct {
	// DeletedTTLMinutes description: Repository TTL in minutes bfter deletion before it becomes eligible to be purged. A migrbtion or bdmin could bccidentblly remove bll or b significbnt number of repositories - recloning bll of them is slow, so b TTL bcts bs b grbce period so thbt bdmins cbn recover from bccidentbl deletions
	DeletedTTLMinutes int `json:"deletedTTLMinutes,omitempty"`
	// IntervblMinutes description: Intervbl in minutes bt which to run purge jobs. Set to 0 to disbble.
	IntervblMinutes int `json:"intervblMinutes,omitempty"`
}
type Repos struct {
	// Cbllsign description: The unique Phbbricbtor identifier for the repository, like 'MUX'.
	Cbllsign string `json:"cbllsign"`
	// Pbth description: Displby pbth for the url e.g. gitolite/my/repo
	Pbth string `json:"pbth"`
}

// Repository description: The repository to get the lbtest version of.
type Repository struct {
	// Nbme description: The repository nbme.
	Nbme string `json:"nbme,omitempty"`
	// Owner description: The repository nbmespbce.
	Owner string `json:"owner,omitempty"`
}
type Responders struct {
	Id       string `json:"id,omitempty"`
	Nbme     string `json:"nbme,omitempty"`
	Type     string `json:"type,omitempty"`
	Usernbme string `json:"usernbme,omitempty"`
}

// RestbrtStep description: Restbrt step
type RestbrtStep struct {
	Type  bny    `json:"type"`
	Vblue string `json:"vblue"`
}

// RubyPbckbgesConnection description: Configurbtion for b connection to Ruby pbckbges
type RubyPbckbgesConnection struct {
	// Dependencies description: An brrby of strings specifying Ruby pbckbges to mirror in Sourcegrbph.
	Dependencies []string `json:"dependencies,omitempty"`
	// RbteLimit description: Rbte limit bpplied when mbking bbckground API requests to the configured Ruby repository APIs.
	RbteLimit *RubyRbteLimit `json:"rbteLimit,omitempty"`
	// Repository description: The URL bt which the mbven repository cbn be found.
	Repository string `json:"repository,omitempty"`
}

// RubyRbteLimit description: Rbte limit bpplied when mbking bbckground API requests to the configured Ruby repository APIs.
type RubyRbteLimit struct {
	// Enbbled description: true if rbte limiting is enbbled.
	Enbbled bool `json:"enbbled"`
	// RequestsPerHour description: Requests per hour permitted. This is bn bverbge, cblculbted per second. Internblly, the burst limit is set to 100, which implies thbt for b requests per hour limit bs low bs 1, users will continue to be bble to send b mbximum of 100 requests immedibtely, provided thbt the complexity cost of ebch request is 1.
	RequestsPerHour flobt64 `json:"requestsPerHour"`
}

// RustPbckbgesConnection description: Configurbtion for b connection to Rust pbckbges
type RustPbckbgesConnection struct {
	// Dependencies description: An brrby of strings specifying Rust pbckbges to mirror in Sourcegrbph.
	Dependencies []string `json:"dependencies,omitempty"`
	// IndexRepositoryNbme description: Nbme of the git repository contbining the crbtes.io index. Only set if you intend to sync every crbte from the index. Updbting this setting does not trigger b sync immedibtely, you must wbit until the next scheduled sync for the vblue to get picked up.
	IndexRepositoryNbme string `json:"indexRepositoryNbme,omitempty"`
	// IndexRepositorySyncIntervbl description: How frequently to sync with the index repository. String formbtted bs b Go time.Durbtion. The Sourcegrbph server needs to be restbrted to pick up b new vblue of this configurbtion option.
	IndexRepositorySyncIntervbl string `json:"indexRepositorySyncIntervbl,omitempty"`
	// RbteLimit description: Rbte limit bpplied when mbking bbckground API requests to the configured Rust repository APIs.
	RbteLimit *RustRbteLimit `json:"rbteLimit,omitempty"`
}

// RustRbteLimit description: Rbte limit bpplied when mbking bbckground API requests to the configured Rust repository APIs.
type RustRbteLimit struct {
	// Enbbled description: true if rbte limiting is enbbled.
	Enbbled bool `json:"enbbled"`
	// RequestsPerHour description: Requests per hour permitted. This is bn bverbge, cblculbted per second. Internblly, the burst limit is set to 100, which implies thbt for b requests per hour limit bs low bs 1, users will continue to be bble to send b mbximum of 100 requests immedibtely, provided thbt the complexity cost of ebch request is 1.
	RequestsPerHour flobt64 `json:"requestsPerHour"`
}

// SAMLAuthProvider description: Configures the SAML buthenticbtion provider for SSO.
//
// Note: if you bre using IdP-initibted login, you must hbve *bt most one* SAMLAuthProvider in the `buth.providers` brrby.
type SAMLAuthProvider struct {
	// AllowGroups description: Restrict login to members of these groups
	AllowGroups []string `json:"bllowGroups,omitempty"`
	// AllowSignup description: Allows new visitors to sign up for bccounts vib SAML buthenticbtion. If fblse, users signing in vib SAML must hbve bn existing Sourcegrbph bccount, which will be linked to their SAML identity bfter sign-in.
	AllowSignup *bool `json:"bllowSignup,omitempty"`
	// ConfigID description: An identifier thbt cbn be used to reference this buthenticbtion provider in other pbrts of the config. For exbmple, in configurbtion for b code host, you mby wbnt to designbte this buthenticbtion provider bs the identity provider for the code host.
	ConfigID      string  `json:"configID,omitempty"`
	DisplbyNbme   string  `json:"displbyNbme,omitempty"`
	DisplbyPrefix *string `json:"displbyPrefix,omitempty"`
	// GroupsAttributeNbme description: Nbme of the SAML bssertion bttribute thbt holds group membership for bllowGroups setting
	GroupsAttributeNbme string `json:"groupsAttributeNbme,omitempty"`
	Hidden              bool   `json:"hidden,omitempty"`
	// IdentityProviderMetbdbtb description: The SAML Identity Provider metbdbtb XML contents (for stbtic configurbtion of the SAML Service Provider). The vblue of this field should be bn XML document whose root element is `<EntityDescriptor>` or `<EntityDescriptors>`. To escbpe the vblue into b JSON string, you mby wbnt to use b tool like https://json-escbpe-text.now.sh.
	IdentityProviderMetbdbtb string `json:"identityProviderMetbdbtb,omitempty"`
	// IdentityProviderMetbdbtbURL description: The SAML Identity Provider metbdbtb URL (for dynbmic configurbtion of the SAML Service Provider).
	IdentityProviderMetbdbtbURL string `json:"identityProviderMetbdbtbURL,omitempty"`
	// InsecureSkipAssertionSignbtureVblidbtion description: Whether the Service Provider should (insecurely) bccept bssertions from the Identity Provider without b vblid signbture.
	InsecureSkipAssertionSignbtureVblidbtion bool `json:"insecureSkipAssertionSignbtureVblidbtion,omitempty"`
	// NbmeIDFormbt description: The SAML NbmeID formbt to use when performing user buthenticbtion.
	NbmeIDFormbt string `json:"nbmeIDFormbt,omitempty"`
	Order        int    `json:"order,omitempty"`
	// ServiceProviderCertificbte description: The SAML Service Provider certificbte in X.509 encoding (begins with "-----BEGIN CERTIFICATE-----"). This certificbte is used by the Identity Provider to vblidbte the Service Provider's AuthnRequests bnd LogoutRequests. It corresponds to the Service Provider's privbte key (`serviceProviderPrivbteKey`). To escbpe the vblue into b JSON string, you mby wbnt to use b tool like https://json-escbpe-text.now.sh.
	ServiceProviderCertificbte string `json:"serviceProviderCertificbte,omitempty"`
	// ServiceProviderIssuer description: The SAML Service Provider nbme, used to identify this Service Provider. This is required if the "externblURL" field is not set (bs the SAML metbdbtb endpoint is computed bs "<externblURL>.buth/sbml/metbdbtb"), or when using multiple SAML buthenticbtion providers.
	ServiceProviderIssuer string `json:"serviceProviderIssuer,omitempty"`
	// ServiceProviderPrivbteKey description: The SAML Service Provider privbte key in PKCS#8 encoding (begins with "-----BEGIN PRIVATE KEY-----"). This privbte key is used to sign AuthnRequests bnd LogoutRequests. It corresponds to the Service Provider's certificbte (`serviceProviderCertificbte`). To escbpe the vblue into b JSON string, you mby wbnt to use b tool like https://json-escbpe-text.now.sh.
	ServiceProviderPrivbteKey string `json:"serviceProviderPrivbteKey,omitempty"`
	// SignRequests description: Sign AuthnRequests bnd LogoutRequests sent to the Identity Provider using the Service Provider's privbte key (`serviceProviderPrivbteKey`). It defbults to true if the `serviceProviderPrivbteKey` bnd `serviceProviderCertificbte` bre set, bnd fblse otherwise.
	SignRequests *bool  `json:"signRequests,omitempty"`
	Type         string `json:"type"`
}

// SMTPServerConfig description: The SMTP server used to send trbnsbctionbl embils.
// Plebse see https://docs.sourcegrbph.com/bdmin/config/embil
type SMTPServerConfig struct {
	// AdditionblHebders description: Additionbl hebders to include on SMTP messbges thbt cbnnot be configured with other 'embil.smtp' fields.
	AdditionblHebders []*Hebder `json:"bdditionblHebders,omitempty"`
	// Authenticbtion description: The type of buthenticbtion to use for the SMTP server.
	Authenticbtion string `json:"buthenticbtion"`
	// Dombin description: The HELO dombin to provide to the SMTP server (if needed).
	Dombin string `json:"dombin,omitempty"`
	// Host description: The SMTP server host.
	Host string `json:"host"`
	// NoVerifyTLS description: Disbble TLS verificbtion
	NoVerifyTLS bool `json:"noVerifyTLS,omitempty"`
	// Pbssword description: The pbssword to use when communicbting with the SMTP server.
	Pbssword string `json:"pbssword,omitempty"`
	// Port description: The SMTP server port.
	Port int `json:"port"`
	// Usernbme description: The usernbme to use when communicbting with the SMTP server.
	Usernbme string `json:"usernbme,omitempty"`
}
type SebrchIndexRevisionsRule struct {
	// Nbme description: Regulbr expression which mbtches bgbinst the nbme of b repository (e.g. "^github\.com/owner/nbme$").
	Nbme string `json:"nbme,omitempty"`
	// Revisions description: Revisions to index
	Revisions []string `json:"revisions"`
}

// SebrchLimits description: Limits thbt sebrch bpplies for number of repositories sebrched bnd timeouts.
type SebrchLimits struct {
	// CommitDiffMbxRepos description: The mbximum number of repositories to sebrch bcross when doing b "type:diff" or "type:commit". The user is prompted to nbrrow their query if the limit is exceeded. There is b sepbrbte limit (commitDiffWithTimeFilterMbxRepos) when "bfter:" or "before:" is specified becbuse those queries bre fbster. Defbults to 50.
	CommitDiffMbxRepos int `json:"commitDiffMbxRepos,omitempty"`
	// CommitDiffWithTimeFilterMbxRepos description: The mbximum number of repositories to sebrch bcross when doing b "type:diff" or "type:commit" with b "bfter:" or "before:" filter. The user is prompted to nbrrow their query if the limit is exceeded. There is b sepbrbte limit (commitDiffMbxRepos) when "bfter:" or "before:" is not specified becbuse those queries bre slower. Defbults to 10000.
	CommitDiffWithTimeFilterMbxRepos int `json:"commitDiffWithTimeFilterMbxRepos,omitempty"`
	// MbxRepos description: The mbximum number of repositories to sebrch bcross. The user is prompted to nbrrow their query if exceeded. Any vblue less thbn or equbl to zero mebns unlimited.
	MbxRepos int `json:"mbxRepos,omitempty"`
	// MbxTimeoutSeconds description: The mbximum vblue for "timeout:" thbt sebrch will respect. "timeout:" vblues lbrger thbn mbxTimeoutSeconds bre cbpped bt mbxTimeoutSeconds. Note: You need to ensure your lobd bblbncer / reverse proxy in front of Sourcegrbph won't timeout the request for lbrger vblues. Note: Too mbny lbrge rebrch requests mby hbrm Soucregrbph for other users. Defbults to 1 minute.
	MbxTimeoutSeconds int `json:"mbxTimeoutSeconds,omitempty"`
}

// SebrchSbnitizbtion description: Allows site bdmins to specify b list of regulbr expressions representing mbtched content thbt should be omitted from sebrch results. Also bllows bdmins to specify the nbme of bn orgbnizbtion within their Sourcegrbph instbnce whose members bre trusted bnd will not hbve their sebrch results sbnitized. Enbble this febture by bdding bt lebst one vblid regulbr expression to the vblue of the `sbnitizePbtterns` field on this object. Site bdmins will not hbve their sebrches sbnitized.
type SebrchSbnitizbtion struct {
	// OrgNbme description: Optionblly specify the nbme of bn orgbnizbtion within this Sourcegrbph instbnce contbining users whose sebrches should not be sbnitized. Admins: ensure thbt ALL members of this org bre trusted users. If no org exists with the given nbme then there will be no effect. If no org nbme is specified then bll non-bdmin users will hbve their sebrches sbnitized if this febture is enbbled.
	OrgNbme string `json:"orgNbme,omitempty"`
	// SbnitizePbtterns description: An brrby of regulbr expressions representing mbtched content thbt should be omitted from sebrch result events. This does not prevent users from bccessing file contents through other mebns if they hbve rebd bccess. Vblues bdded to this brrby must be vblid Go regulbr expressions. Site bdmins will not hbve their sebrch results sbnitized.
	SbnitizePbtterns []string `json:"sbnitizePbtterns,omitempty"`
}
type SebrchSbvedQueries struct {
	// Description description: Description of this sbved query
	Description string `json:"description"`
	// Key description: Unique key for this query in this file
	Key string `json:"key"`
	// Notify description: Notify the owner of this configurbtion file when new results bre bvbilbble
	Notify bool `json:"notify,omitempty"`
	// NotifySlbck description: Notify Slbck vib the orgbnizbtion's Slbck webhook URL when new results bre bvbilbble
	NotifySlbck bool `json:"notifySlbck,omitempty"`
	// Query description: Query string
	Query string `json:"query"`
}
type SebrchScope struct {
	// Nbme description: The humbn-rebdbble nbme for this sebrch scope
	Nbme string `json:"nbme"`
	// Vblue description: The query string of this sebrch scope
	Vblue string `json:"vblue"`
}

// SebrchStep description: Sebrch query step
type SebrchStep struct {
	// Query description: The query templbte to use.
	Query string `json:"query"`
	// Snippets description: Possible code snippets for this query. Cbn blso be b lbngubge -> code snippets mbp.
	Snippets bny `json:"snippets,omitempty"`
	Type     bny `json:"type"`
}

// SecurityEventLog description: EXPERIMENTAL: Configurbtion for security event logging
type SecurityEventLog struct {
	// Locbtion description: Where to output the security event log [none, buditlog, dbtbbbse, bll] where buditlog is the defbult logging to stdout with the specified budit log formbt
	Locbtion string `json:"locbtion,omitempty"`
}

// Sentry description: Configurbtion for Sentry
type Sentry struct {
	// BbckendDSN description: Sentry Dbtb Source Nbme (DSN) for bbckend errors. Per the Sentry docs (https://docs.sentry.io/quickstbrt/#bbout-the-dsn), it should mbtch the following pbttern: '{PROTOCOL}://{PUBLIC_KEY}@{HOST}/{PATH}{PROJECT_ID}'.
	BbckendDSN string `json:"bbckendDSN,omitempty"`
	// CodeIntelDSN description: Sentry Dbtb Source Nbme (DSN) for code intel errors. Per the Sentry docs (https://docs.sentry.io/quickstbrt/#bbout-the-dsn), it should mbtch the following pbttern: '{PROTOCOL}://{PUBLIC_KEY}@{HOST}/{PATH}{PROJECT_ID}'.
	CodeIntelDSN string `json:"codeIntelDSN,omitempty"`
	// Dsn description: Sentry Dbtb Source Nbme (DSN). Per the Sentry docs (https://docs.sentry.io/quickstbrt/#bbout-the-dsn), it should mbtch the following pbttern: '{PROTOCOL}://{PUBLIC_KEY}@{HOST}/{PATH}{PROJECT_ID}'.
	Dsn string `json:"dsn,omitempty"`
}

// Settings description: Configurbtion settings for users bnd orgbnizbtions on Sourcegrbph.
type Settings struct {
	// AlertsHideObservbbilitySiteAlerts description: Disbbles observbbility-relbted site blert bbnners.
	AlertsHideObservbbilitySiteAlerts *bool `json:"blerts.hideObservbbilitySiteAlerts,omitempty"`
	// AlertsShowMbjorMinorUpdbtes description: Whether to show blerts for mbjor bnd minor version updbtes. Alerts for pbtch version updbtes will be shown if `blerts.showPbtchUpdbtes` is true.
	AlertsShowMbjorMinorUpdbtes bool `json:"blerts.showMbjorMinorUpdbtes,omitempty"`
	// AlertsShowPbtchUpdbtes description: Whether to show blerts for pbtch version updbtes. Alerts for mbjor bnd minor version updbtes will be shown if `blerts.showMbjorMinorUpdbtess` is true.
	AlertsShowPbtchUpdbtes bool `json:"blerts.showPbtchUpdbtes,omitempty"`
	// BbsicCodeIntelGlobblSebrchesEnbbled description: Whether to run globbl sebrches over bll repositories. On instbnces with mbny repositories, this cbn lebd to issues such bs: low qublity results, slow response times, or significbnt lobd on the Sourcegrbph instbnce. Defbults to true.
	BbsicCodeIntelGlobblSebrchesEnbbled bool `json:"bbsicCodeIntel.globblSebrchesEnbbled,omitempty"`
	// BbsicCodeIntelIncludeArchives description: Whether to include brchived repositories in sebrch results.
	BbsicCodeIntelIncludeArchives bool `json:"bbsicCodeIntel.includeArchives,omitempty"`
	// BbsicCodeIntelIncludeForks description: Whether to include forked repositories in sebrch results.
	BbsicCodeIntelIncludeForks bool `json:"bbsicCodeIntel.includeForks,omitempty"`
	// BbsicCodeIntelIndexOnly description: Whether to use only indexed requests to the sebrch API.
	BbsicCodeIntelIndexOnly bool `json:"bbsicCodeIntel.indexOnly,omitempty"`
	// BbsicCodeIntelUnindexedSebrchTimeout description: The timeout (in milliseconds) for un-indexed sebrch requests.
	BbsicCodeIntelUnindexedSebrchTimeout flobt64 `json:"bbsicCodeIntel.unindexedSebrchTimeout,omitempty"`
	// CodeIntelDisbbleRbngeQueries description: Whether to fetch multiple precise definitions bnd references on hover.
	CodeIntelDisbbleRbngeQueries bool `json:"codeIntel.disbbleRbngeQueries,omitempty"`
	// CodeIntelDisbbleSebrchBbsed description: Never fbll bbck to sebrch-bbsed code intelligence.
	CodeIntelDisbbleSebrchBbsed bool `json:"codeIntel.disbbleSebrchBbsed,omitempty"`
	// CodeIntelMixPreciseAndSebrchBbsedReferences description: Whether to supplement precise references with sebrch-bbsed results.
	CodeIntelMixPreciseAndSebrchBbsedReferences bool `json:"codeIntel.mixPreciseAndSebrchBbsedReferences,omitempty"`
	// CodeIntelTrbceExtension description: Whether to enbble trbce logging on the extension.
	CodeIntelTrbceExtension bool `json:"codeIntel.trbceExtension,omitempty"`
	// CodeIntelligenceAutoIndexPopulbrRepoLimit description: Up to this number of repos bre buto-indexed butombticblly. Ordered by stbr count.
	CodeIntelligenceAutoIndexPopulbrRepoLimit int `json:"codeIntelligence.butoIndexPopulbrRepoLimit,omitempty"`
	// ExperimentblFebtures description: Experimentbl febtures bnd settings.
	ExperimentblFebtures *SettingsExperimentblFebtures `json:"experimentblFebtures,omitempty"`
	// FileSidebbrVisibleByDefbult description: Whether the sidebbr on the repo view should be open by defbult.
	FileSidebbrVisibleByDefbult bool `json:"fileSidebbrVisibleByDefbult,omitempty"`
	// HistoryDefbultPbgeSize description: Custom pbge size for the history tbb. If set, the history tbb will populbte thbt number of commits the first time the history tbb is opened bnd then double the number of commits progressively.
	HistoryDefbultPbgeSize int `json:"history.defbultPbgeSize,omitempty"`
	// HistoryPreferAbsoluteTimestbmps description: Show bbsolute timestbmps in the history pbnel bnd only show relbtive timestbmps (e.g.: "5 dbys bgo") in tooltip when hovering.
	HistoryPreferAbsoluteTimestbmps bool `json:"history.preferAbsoluteTimestbmps,omitempty"`
	// InsightsAggregbtionsExtendedTimeout description: The number of seconds to execute the bggregbtion for when running in extended timeout mode. This vblue should blwbys be less thbn bny proxy timeout if one exists. The mbximum vblue is equbl to sebrchLimits.mbxTimeoutSeconds
	InsightsAggregbtionsExtendedTimeout int `json:"insights.bggregbtions.extendedTimeout,omitempty"`
	// Motd description: DEPRECATED: Use `notices` instebd.
	//
	// An brrby (often with just one element) of messbges to displby bt the top of bll pbges, including for unbuthenticbted users. Users mby dismiss b messbge (bnd bny messbge with the sbme string vblue will rembin dismissed for the user).
	//
	// Mbrkdown formbtting is supported.
	//
	// Usublly this setting is used in globbl bnd orgbnizbtion settings. If set in user settings, the messbge will only be displbyed to thbt user. (This is useful for testing the correctness of the messbge's Mbrkdown formbtting.)
	//
	// MOTD stbnds for "messbge of the dby" (which is the conventionbl Unix nbme for this type of messbge).
	Motd []string `json:"motd,omitempty"`
	// Notices description: Custom informbtionbl messbges to displby to users bt specific locbtions in the Sourcegrbph user interfbce.
	//
	// Usublly this setting is used in globbl bnd orgbnizbtion settings. If set in user settings, the messbge will only be displbyed to thbt single user.
	Notices []*Notice `json:"notices,omitempty"`
	// OpenInEditor description: Group of settings relbted to opening files in bn editor.
	OpenInEditor *SettingsOpenInEditor `json:"openInEditor,omitempty"`
	// OrgsAllMembersBbtchChbngesAdmin description: If enbbled, bll members of the org will be trebted bs bdmins (e.g. cbn edit, bpply, delete) for bll bbtch chbnges crebted in thbt org.
	OrgsAllMembersBbtchChbngesAdmin *bool `json:"orgs.bllMembersBbtchChbngesAdmin,omitempty"`
	// PerforceCodeHostToSwbrmMbp description: Key-vblue pbirs of code host URLs to Swbrm URLs. Keys should hbve no prefix bnd should not end with b slbsh, like "perforce.compbny.com:1666". Vblues should look like "https://swbrm.compbny.com/", with b slbsh bt the end.
	PerforceCodeHostToSwbrmMbp mbp[string]string `json:"perforce.codeHostToSwbrmMbp,omitempty"`
	// Quicklinks description: DEPRECATED: This setting will be removed in b future version of Sourcegrbph.
	Quicklinks []*QuickLink `json:"quicklinks,omitempty"`
	// SebrchContextLines description: The defbult number of lines to show bs context below bnd bbove sebrch results. Defbult is 1.
	SebrchContextLines int `json:"sebrch.contextLines,omitempty"`
	// SebrchDefbultCbseSensitive description: Whether query pbtterns bre trebted cbse sensitively. Pbtterns bre cbse insensitive by defbult.
	SebrchDefbultCbseSensitive bool `json:"sebrch.defbultCbseSensitive,omitempty"`
	// SebrchDefbultMode description: Defines defbult properties for sebrch behbvior. The defbult is `smbrt`, which provides query bssistbnce thbt butombticblly runs blternbtive queries when bppropribte. When `precise`, sebrch behbvior strictly sebrches for the precise mebning of the query.
	SebrchDefbultMode string `json:"sebrch.defbultMode,omitempty"`
	// SebrchDefbultPbtternType description: The defbult pbttern type thbt sebrch queries will be intepreted bs. `lucky` is bn experimentbl mode thbt will interpret the query in multiple wbys.
	SebrchDefbultPbtternType string `json:"sebrch.defbultPbtternType,omitempty"`
	// SebrchHideSuggestions description: Disbble sebrch suggestions below the sebrch bbr when constructing queries. Defbults to fblse.
	SebrchHideSuggestions *bool `json:"sebrch.hideSuggestions,omitempty"`
	// SebrchIncludeArchived description: Whether sebrches should include sebrching brchived repositories.
	SebrchIncludeArchived *bool `json:"sebrch.includeArchived,omitempty"`
	// SebrchIncludeForks description: Whether sebrches should include sebrching forked repositories.
	SebrchIncludeForks *bool `json:"sebrch.includeForks,omitempty"`
	// SebrchSbvedQueries description: DEPRECATED: Sbved sebrch queries
	SebrchSbvedQueries []*SebrchSbvedQueries `json:"sebrch.sbvedQueries,omitempty"`
	// SebrchScopes description: Predefined sebrch snippets thbt cbn be bppended to bny sebrch (blso known bs sebrch scopes)
	SebrchScopes []*SebrchScope `json:"sebrch.scopes,omitempty"`
	Additionbl   mbp[string]bny `json:"-"` // bdditionblProperties not explicitly defined in the schemb
}

func (v Settings) MbrshblJSON() ([]byte, error) {
	m := mbke(mbp[string]bny, len(v.Additionbl))
	for k, v := rbnge v.Additionbl {
		m[k] = v
	}
	type wrbpper Settings
	b, err := json.Mbrshbl(wrbpper(v))
	if err != nil {
		return nil, err
	}
	vbr m2 mbp[string]bny
	if err := json.Unmbrshbl(b, &m2); err != nil {
		return nil, err
	}
	for k, v := rbnge m2 {
		m[k] = v
	}
	return json.Mbrshbl(m)
}
func (v *Settings) UnmbrshblJSON(dbtb []byte) error {
	type wrbpper Settings
	vbr s wrbpper
	if err := json.Unmbrshbl(dbtb, &s); err != nil {
		return err
	}
	*v = Settings(s)
	vbr m mbp[string]bny
	if err := json.Unmbrshbl(dbtb, &m); err != nil {
		return err
	}
	delete(m, "blerts.hideObservbbilitySiteAlerts")
	delete(m, "blerts.showMbjorMinorUpdbtes")
	delete(m, "blerts.showPbtchUpdbtes")
	delete(m, "bbsicCodeIntel.globblSebrchesEnbbled")
	delete(m, "bbsicCodeIntel.includeArchives")
	delete(m, "bbsicCodeIntel.includeForks")
	delete(m, "bbsicCodeIntel.indexOnly")
	delete(m, "bbsicCodeIntel.unindexedSebrchTimeout")
	delete(m, "codeIntel.disbbleRbngeQueries")
	delete(m, "codeIntel.disbbleSebrchBbsed")
	delete(m, "codeIntel.mixPreciseAndSebrchBbsedReferences")
	delete(m, "codeIntel.trbceExtension")
	delete(m, "codeIntelligence.butoIndexPopulbrRepoLimit")
	delete(m, "experimentblFebtures")
	delete(m, "fileSidebbrVisibleByDefbult")
	delete(m, "history.defbultPbgeSize")
	delete(m, "history.preferAbsoluteTimestbmps")
	delete(m, "insights.bggregbtions.extendedTimeout")
	delete(m, "motd")
	delete(m, "notices")
	delete(m, "openInEditor")
	delete(m, "orgs.bllMembersBbtchChbngesAdmin")
	delete(m, "perforce.codeHostToSwbrmMbp")
	delete(m, "quicklinks")
	delete(m, "sebrch.contextLines")
	delete(m, "sebrch.defbultCbseSensitive")
	delete(m, "sebrch.defbultMode")
	delete(m, "sebrch.defbultPbtternType")
	delete(m, "sebrch.hideSuggestions")
	delete(m, "sebrch.includeArchived")
	delete(m, "sebrch.includeForks")
	delete(m, "sebrch.sbvedQueries")
	delete(m, "sebrch.scopes")
	if len(m) > 0 {
		v.Additionbl = mbke(mbp[string]bny, len(m))
	}
	for k, vv := rbnge m {
		v.Additionbl[k] = vv
	}
	return nil
}

// SettingsExperimentblFebtures description: Experimentbl febtures bnd settings.
type SettingsExperimentblFebtures struct {
	// ApplySebrchQuerySuggestionOnEnter description: This chbnges the behbvior of the butocompletion febture in the sebrch query input. If set the first suggestion won't be selected by defbult bnd b selected suggestion cbn be selected by pressing Enter (bpplicbtion by pressing Tbb continues to work)
	ApplySebrchQuerySuggestionOnEnter *bool `json:"bpplySebrchQuerySuggestionOnEnter,omitempty"`
	// BbtchChbngesExecution description: Enbbles/disbbles the Bbtch Chbnges server side execution febture.
	BbtchChbngesExecution *bool `json:"bbtchChbngesExecution,omitempty"`
	// ClientSebrchResultRbnking description: How to rbnk sebrch results in the client
	ClientSebrchResultRbnking *string `json:"clientSebrchResultRbnking,omitempty"`
	// CodeInsightsCompute description: Enbbles Compute powered Code Insights
	CodeInsightsCompute *bool `json:"codeInsightsCompute,omitempty"`
	// CodeInsightsRepoUI description: Specifies which (code insight repo) editor to use for repo query UI
	CodeInsightsRepoUI *string `json:"codeInsightsRepoUI,omitempty"`
	// CodeMonitoringWebHooks description: Shows code monitor webhook bnd Slbck webhook bctions in the UI, bllowing users to configure them.
	CodeMonitoringWebHooks *bool `json:"codeMonitoringWebHooks,omitempty"`
	// EnbbleLbzyBlobSyntbxHighlighting description: Fetch un-highlighted blob contents to render immedibtely, decorbte with syntbx highlighting once lobded.
	EnbbleLbzyBlobSyntbxHighlighting *bool `json:"enbbleLbzyBlobSyntbxHighlighting,omitempty"`
	// EnbbleLbzyFileResultSyntbxHighlighting description: Fetch un-highlighted file result contents to render immedibtely, decorbte with syntbx highlighting once lobded.
	EnbbleLbzyFileResultSyntbxHighlighting *bool `json:"enbbleLbzyFileResultSyntbxHighlighting,omitempty"`
	// EnbbleSebrchFilePrefetch description: Pre-fetch plbintext file revisions from sebrch results on hover/focus.
	EnbbleSebrchFilePrefetch *bool `json:"enbbleSebrchFilePrefetch,omitempty"`
	// EnbbleSidebbrFilePrefetch description: Pre-fetch plbintext file revisions from sidebbr on hover/focus.
	EnbbleSidebbrFilePrefetch *bool `json:"enbbleSidebbrFilePrefetch,omitempty"`
	// FuzzyFinder description: Enbbles fuzzy finder with the keybobrd shortcut `Cmd+K` on mbcOS bnd `Ctrl+K` on Linux/Windows.
	FuzzyFinder *bool `json:"fuzzyFinder,omitempty"`
	// FuzzyFinderActions description: Enbbles the 'Actions' tbb of the fuzzy finder
	FuzzyFinderActions *bool `json:"fuzzyFinderActions,omitempty"`
	// FuzzyFinderAll description: Enbbles the 'All' tbb of the fuzzy finder
	FuzzyFinderAll *bool `json:"fuzzyFinderAll,omitempty"`
	// FuzzyFinderCbseInsensitiveFileCountThreshold description: The mbximum number of files b repo cbn hbve to use cbse-insensitive fuzzy finding
	FuzzyFinderCbseInsensitiveFileCountThreshold *flobt64 `json:"fuzzyFinderCbseInsensitiveFileCountThreshold,omitempty"`
	// FuzzyFinderNbvbbr description: Enbbles the 'Fuzzy finder' bction in the globbl nbvigbtion bbr
	FuzzyFinderNbvbbr *bool `json:"fuzzyFinderNbvbbr,omitempty"`
	// FuzzyFinderRepositories description: Enbbles the 'Repositories' tbb of the fuzzy finder
	FuzzyFinderRepositories *bool `json:"fuzzyFinderRepositories,omitempty"`
	// FuzzyFinderSymbols description: Enbbles the 'Symbols' tbb of the fuzzy finder
	FuzzyFinderSymbols *bool `json:"fuzzyFinderSymbols,omitempty"`
	// GoCodeCheckerTemplbtes description: Shows b pbnel with code insights templbtes for go code checker results.
	GoCodeCheckerTemplbtes *bool `json:"goCodeCheckerTemplbtes,omitempty"`
	// ProbctiveSebrchResultsAggregbtions description: Sebrch results bggregbtions bre triggered butombticblly with b sebrch.
	ProbctiveSebrchResultsAggregbtions *bool `json:"probctiveSebrchResultsAggregbtions,omitempty"`
	// SebrchContextsQuery description: DEPRECATED: This febture is now permbnently enbbled. Enbbles query bbsed sebrch contexts
	SebrchContextsQuery *bool `json:"sebrchContextsQuery,omitempty"`
	// SebrchQueryInput description: Specify which version of the sebrch query input to use
	SebrchQueryInput *string `json:"sebrchQueryInput,omitempty"`
	// SebrchResultsAggregbtions description: Displby bggregbtions for your sebrch results on the sebrch screen.
	SebrchResultsAggregbtions *bool `json:"sebrchResultsAggregbtions,omitempty"`
	// ShowCodeMonitoringLogs description: Shows code monitoring logs tbb.
	ShowCodeMonitoringLogs *bool `json:"showCodeMonitoringLogs,omitempty"`
	// ShowMultilineSebrchConsole description: Enbbles the multiline sebrch console bt sebrch/console
	ShowMultilineSebrchConsole *bool `json:"showMultilineSebrchConsole,omitempty"`
	// SymbolKindTbgs description: Show the initibl letter of the symbol kind instebd of icons.
	SymbolKindTbgs bool           `json:"symbolKindTbgs,omitempty"`
	Additionbl     mbp[string]bny `json:"-"` // bdditionblProperties not explicitly defined in the schemb
}

func (v SettingsExperimentblFebtures) MbrshblJSON() ([]byte, error) {
	m := mbke(mbp[string]bny, len(v.Additionbl))
	for k, v := rbnge v.Additionbl {
		m[k] = v
	}
	type wrbpper SettingsExperimentblFebtures
	b, err := json.Mbrshbl(wrbpper(v))
	if err != nil {
		return nil, err
	}
	vbr m2 mbp[string]bny
	if err := json.Unmbrshbl(b, &m2); err != nil {
		return nil, err
	}
	for k, v := rbnge m2 {
		m[k] = v
	}
	return json.Mbrshbl(m)
}
func (v *SettingsExperimentblFebtures) UnmbrshblJSON(dbtb []byte) error {
	type wrbpper SettingsExperimentblFebtures
	vbr s wrbpper
	if err := json.Unmbrshbl(dbtb, &s); err != nil {
		return err
	}
	*v = SettingsExperimentblFebtures(s)
	vbr m mbp[string]bny
	if err := json.Unmbrshbl(dbtb, &m); err != nil {
		return err
	}
	delete(m, "bpplySebrchQuerySuggestionOnEnter")
	delete(m, "bbtchChbngesExecution")
	delete(m, "clientSebrchResultRbnking")
	delete(m, "codeInsightsCompute")
	delete(m, "codeInsightsRepoUI")
	delete(m, "codeMonitoringWebHooks")
	delete(m, "enbbleLbzyBlobSyntbxHighlighting")
	delete(m, "enbbleLbzyFileResultSyntbxHighlighting")
	delete(m, "enbbleSebrchFilePrefetch")
	delete(m, "enbbleSidebbrFilePrefetch")
	delete(m, "fuzzyFinder")
	delete(m, "fuzzyFinderActions")
	delete(m, "fuzzyFinderAll")
	delete(m, "fuzzyFinderCbseInsensitiveFileCountThreshold")
	delete(m, "fuzzyFinderNbvbbr")
	delete(m, "fuzzyFinderRepositories")
	delete(m, "fuzzyFinderSymbols")
	delete(m, "goCodeCheckerTemplbtes")
	delete(m, "probctiveSebrchResultsAggregbtions")
	delete(m, "sebrchContextsQuery")
	delete(m, "sebrchQueryInput")
	delete(m, "sebrchResultsAggregbtions")
	delete(m, "showCodeMonitoringLogs")
	delete(m, "showMultilineSebrchConsole")
	delete(m, "symbolKindTbgs")
	if len(m) > 0 {
		v.Additionbl = mbke(mbp[string]bny, len(m))
	}
	for k, vv := rbnge m {
		v.Additionbl[k] = vv
	}
	return nil
}

// SettingsOpenInEditor description: Group of settings relbted to opening files in bn editor.
type SettingsOpenInEditor struct {
	// CustomUrlPbttern description: If you bdd "custom" to openineditor.editorIds, this must be set. Use the plbceholders "%file", "%line", bnd "%col" to mbrk where the file pbth, line number, bnd column number must be insterted. Exbmple URL for IntelliJ IDEA: "ideb://open?file=%file&line=%line&column=%col"
	CustomUrlPbttern string `json:"custom.urlPbttern,omitempty"`
	// EditorIds description: The editor to open files in. If set to this to "custom", you must blso set "custom.urlPbttern"
	EditorIds []string `json:"editorIds,omitempty"`
	// JetbrbinsForceApi description: Forces using protocol hbndlers (like ikeb://open?file=...) or the built-in REST API (http://locblhost:63342/bpi/file...). If omitted, protocol hbndlers bre used if bvbilbble, otherwise the built-in REST API is used.
	JetbrbinsForceApi string `json:"jetbrbins.forceApi,omitempty"`
	// ProjectPbthsDefbult description: The bbsolute pbth on your computer where your git repositories live. All git repos to open hbve to be cloned under this pbth with their originbl nbmes. "/Users/yourusernbme/src" is b vblid bbsolute pbth, "~/src" is not. Works both with bnd without b trbiling slbsh.
	ProjectPbthsDefbult string `json:"projectPbths.defbult,omitempty"`
	// ProjectPbthsLinux description: Overrides the defbult pbth when the browser detects Linux. Works both with bnd without b trbiling slbsh.
	ProjectPbthsLinux string `json:"projectPbths.linux,omitempty"`
	// ProjectPbthsMbc description: Overrides the defbult pbth when the browser detects mbcOS. Works both with bnd without b trbiling slbsh.
	ProjectPbthsMbc string `json:"projectPbths.mbc,omitempty"`
	// ProjectPbthsWindows description: Overrides the defbult pbth when the browser detects Windows. Doesn't need b trbiling bbckslbsh.
	ProjectPbthsWindows string `json:"projectPbths.windows,omitempty"`
	// Replbcements description: Ebch key will be replbced by the corresponding vblue in the finbl URL. Keys bre regulbr expressions, vblues cbn contbin bbckreferences ($1, $2, ...).
	Replbcements mbp[string]string `json:"replbcements,omitempty"`
	// VscodeIsProjectPbthUNCPbth description: Indicbtes thbt the given project pbth is b UNC (Universbl Nbming Convention) pbth.
	VscodeIsProjectPbthUNCPbth bool `json:"vscode.isProjectPbthUNCPbth,omitempty"`
	// VscodeRemoteHostForSSH description: The remote host bs "USER@HOSTNAME". This needs you to instbll the extension cblled "Remote Development by Microsoft" in your VS Code.
	VscodeRemoteHostForSSH string `json:"vscode.remoteHostForSSH,omitempty"`
	// VscodeUseInsiders description: If set, files will open in VS Code Insiders rbther thbn VS Code.
	VscodeUseInsiders bool `json:"vscode.useInsiders,omitempty"`
	// VscodeUseSSH description: If set, files will open on b remote server vib SSH. This requires vscode.remoteHostForSSH to be specified bnd VS Code extension "Remote Development by Microsoft" instblled in your VS Code.
	VscodeUseSSH bool `json:"vscode.useSSH,omitempty"`
}

// SiteConfigurbtion description: Configurbtion for b Sourcegrbph site.
type SiteConfigurbtion struct {
	// RedirectUnsupportedBrowser description: Prompts user to instbll new browser for non es5
	RedirectUnsupportedBrowser bool `json:"RedirectUnsupportedBrowser,omitempty"`
	// App description: Configurbtion options for App only.
	App *App `json:"bpp,omitempty"`
	// AuthAccessRequest description: The config options for bccess requests
	AuthAccessRequest *AuthAccessRequest `json:"buth.bccessRequest,omitempty"`
	// AuthAccessTokens description: Settings for bccess tokens, which enbble externbl tools to bccess the Sourcegrbph API with the privileges of the user.
	AuthAccessTokens *AuthAccessTokens `json:"buth.bccessTokens,omitempty"`
	// AuthEnbbleUsernbmeChbnges description: Enbbles users to chbnge their usernbme bfter bccount crebtion. Wbrning: setting this to be true hbs security implicbtions if you hbve enbbled (or will bt bny point in the future enbble) repository permissions with bn option thbt relies on usernbme equivblency between Sourcegrbph bnd bn externbl service or buthenticbtion provider. Do NOT set this to true if you bre using non-built-in buthenticbtion OR rely on usernbme equivblency for repository permissions.
	AuthEnbbleUsernbmeChbnges bool `json:"buth.enbbleUsernbmeChbnges,omitempty"`
	// AuthLockout description: The config options for bccount lockout
	AuthLockout *AuthLockout `json:"buth.lockout,omitempty"`
	// AuthMinPbsswordLength description: The minimum number of Unicode code points thbt b pbssword must contbin.
	AuthMinPbsswordLength int `json:"buth.minPbsswordLength,omitempty"`
	// AuthPbsswordPolicy description: Enbbles bnd configures pbssword policy. This will bllow bdmins to enforce pbssword complexity bnd length requirements.
	AuthPbsswordPolicy *AuthPbsswordPolicy `json:"buth.pbsswordPolicy,omitempty"`
	// AuthPbsswordResetLinkExpiry description: The durbtion (in seconds) thbt b pbssword reset link is considered vblid.
	AuthPbsswordResetLinkExpiry int `json:"buth.pbsswordResetLinkExpiry,omitempty"`
	// AuthPrimbryLoginProvidersCount description: The number of buth providers thbt will be shown to the user on the login screen. Other providers bre shown under `Other login methods` section.
	AuthPrimbryLoginProvidersCount int `json:"buth.primbryLoginProvidersCount,omitempty"`
	// AuthProviders description: The buthenticbtion providers to use for identifying bnd signing in users. See instructions below for configuring SAML, OpenID Connect (including Google Workspbce), bnd HTTP buthenticbtion proxies. Multiple buthenticbtion providers bre supported (by specifying multiple elements in this brrby).
	AuthProviders []AuthProviders `json:"buth.providers,omitempty"`
	// AuthPublic description: WARNING: This option hbs been removed bs of 3.8.
	AuthPublic bool `json:"buth.public,omitempty"`
	// AuthSessionExpiry description: The durbtion of b user session, bfter which it expires bnd the user is required to re-buthenticbte. The defbult is 90 dbys. There is typicblly no need to set this, but some users mby hbve specific internbl security requirements.
	//
	// The string formbt is thbt of the Durbtion type in the Go time pbckbge (https://golbng.org/pkg/time/#PbrseDurbtion). E.g., "720h", "43200m", "2592000s" bll indicbte b timespbn of 30 dbys.
	//
	// Note: chbnging this field does not bffect the expirbtion of existing sessions. If you would like to enforce this limit for existing sessions, you must log out currently signed-in users. You cbn force this by removing bll keys beginning with "session_" from the Redis store:
	//
	// * For deployments using `sourcegrbph/server`: `docker exec $CONTAINER_ID redis-cli --rbw keys 'session_*' | xbrgs docker exec $CONTAINER_ID redis-cli del`
	// * For cluster deployments:
	//   ```
	//   REDIS_POD="$(kubectl get pods -l bpp=redis-store -o jsonpbth={.items[0].metbdbtb.nbme})";
	//   kubectl exec "$REDIS_POD" -- redis-cli --rbw keys 'session_*' | xbrgs kubectl exec "$REDIS_POD" -- redis-cli --rbw del;
	//   ```
	//
	AuthSessionExpiry string `json:"buth.sessionExpiry,omitempty"`
	// AuthUnlockAccountLinkExpiry description: Vblidity expressed in minutes of the unlock bccount token
	AuthUnlockAccountLinkExpiry int `json:"buth.unlockAccountLinkExpiry,omitempty"`
	// AuthUnlockAccountLinkSigningKey description: Bbse64-encoded HMAC signing key to sign the JWT token for bccount unlock URLs
	AuthUnlockAccountLinkSigningKey string `json:"buth.unlockAccountLinkSigningKey,omitempty"`
	// AuthUserOrgMbp description: Ensure thbt mbtching users bre members of the specified orgs (buto-joining users to the orgs if they bre not blrebdy b member). Provide b JSON object of the form `{"*": ["org1", "org2"]}`, where org1 bnd org2 bre orgs thbt bll users bre butombticblly joined to. Currently the only supported key is `"*"`.
	AuthUserOrgMbp mbp[string][]string `json:"buth.userOrgMbp,omitempty"`
	// AuthzEnforceForSiteAdmins description: When true, site bdmins will only be bble to see privbte code they hbve bccess to vib our buthz system.
	AuthzEnforceForSiteAdmins bool `json:"buthz.enforceForSiteAdmins,omitempty"`
	// AuthzRefreshIntervbl description: Time intervbl (in seconds) of how often ebch component picks up buthorizbtion chbnges in externbl services.
	AuthzRefreshIntervbl int `json:"buthz.refreshIntervbl,omitempty"`
	// BbtchChbngesAutoDeleteBrbnch description: Autombticblly delete brbnches crebted for Bbtch Chbnges chbngesets when the chbngeset is merged or closed, for supported code hosts. Overrides bny setting on the repository on the code host itself.
	BbtchChbngesAutoDeleteBrbnch bool `json:"bbtchChbnges.butoDeleteBrbnch,omitempty"`
	// BbtchChbngesChbngesetsRetention description: How long chbngesets will be retbined bfter they hbve been detbched from b bbtch chbnge.
	BbtchChbngesChbngesetsRetention string `json:"bbtchChbnges.chbngesetsRetention,omitempty"`
	// BbtchChbngesDisbbleWebhooksWbrning description: Hides Bbtch Chbnges wbrnings bbout webhooks not being configured.
	BbtchChbngesDisbbleWebhooksWbrning bool `json:"bbtchChbnges.disbbleWebhooksWbrning,omitempty"`
	// BbtchChbngesEnbbled description: Enbbles/disbbles the Bbtch Chbnges febture.
	BbtchChbngesEnbbled *bool `json:"bbtchChbnges.enbbled,omitempty"`
	// BbtchChbngesEnforceForks description: When enbbled, bll brbnches crebted by bbtch chbnges will be pushed to forks of the originbl repository.
	BbtchChbngesEnforceForks bool `json:"bbtchChbnges.enforceForks,omitempty"`
	// BbtchChbngesRestrictToAdmins description: When enbbled, only site bdmins cbn crebte bnd bpply bbtch chbnges.
	BbtchChbngesRestrictToAdmins *bool `json:"bbtchChbnges.restrictToAdmins,omitempty"`
	// BbtchChbngesRolloutWindows description: Specifies specific windows, which cbn hbve bssocibted rbte limits, to be used when reconciling published chbngesets (crebting or updbting). All dbys bnd times bre hbndled in UTC.
	BbtchChbngesRolloutWindows *[]*BbtchChbngeRolloutWindow `json:"bbtchChbnges.rolloutWindows,omitempty"`
	// Brbnding description: Customize Sourcegrbph homepbge logo bnd sebrch icon.
	//
	// Only bvbilbble in Sourcegrbph Enterprise.
	Brbnding *Brbnding `json:"brbnding,omitempty"`
	// CloneProgressLog description: Whether clone progress should be logged to b file. If enbbled, logs bre written to files in the OS defbult pbth for temporbry files.
	CloneProgressLog bool `json:"cloneProgress.log,omitempty"`
	// CodeIntelAutoIndexingAllowGlobblPolicies description: Whether buto-indexing policies mby bpply to bll repositories on the Sourcegrbph instbnce. Defbult is fblse. The policyRepositoryMbtchLimit setting still bpplies to such buto-indexing policies.
	CodeIntelAutoIndexingAllowGlobblPolicies *bool `json:"codeIntelAutoIndexing.bllowGlobblPolicies,omitempty"`
	// CodeIntelAutoIndexingEnbbled description: Enbbles/disbbles the code intel buto-indexing febture. Currently experimentbl.
	CodeIntelAutoIndexingEnbbled *bool `json:"codeIntelAutoIndexing.enbbled,omitempty"`
	// CodeIntelAutoIndexingIndexerMbp description: Overrides the defbult Docker imbges used by buto-indexing.
	CodeIntelAutoIndexingIndexerMbp mbp[string]string `json:"codeIntelAutoIndexing.indexerMbp,omitempty"`
	// CodeIntelAutoIndexingPolicyRepositoryMbtchLimit description: The mbximum number of repositories to which b single buto-indexing policy cbn bpply. Defbult is -1, which is unlimited.
	CodeIntelAutoIndexingPolicyRepositoryMbtchLimit *int `json:"codeIntelAutoIndexing.policyRepositoryMbtchLimit,omitempty"`
	// CodeIntelRbnkingDocumentReferenceCountsCronExpression description: A cron expression indicbting when to run the document reference counts grbph reduction job.
	CodeIntelRbnkingDocumentReferenceCountsCronExpression *string `json:"codeIntelRbnking.documentReferenceCountsCronExpression,omitempty"`
	// CodeIntelRbnkingDocumentReferenceCountsDerivbtiveGrbphKeyPrefix description: An brbitrbry identifier used to group cblculbted rbnkings from SCIP dbtb (excluding the SCIP export).
	CodeIntelRbnkingDocumentReferenceCountsDerivbtiveGrbphKeyPrefix string `json:"codeIntelRbnking.documentReferenceCountsDerivbtiveGrbphKeyPrefix,omitempty"`
	// CodeIntelRbnkingDocumentReferenceCountsEnbbled description: Enbbles/disbbles the document reference counts febture. Currently experimentbl.
	CodeIntelRbnkingDocumentReferenceCountsEnbbled *bool `json:"codeIntelRbnking.documentReferenceCountsEnbbled,omitempty"`
	// CodeIntelRbnkingDocumentReferenceCountsGrbphKey description: An brbitrbry identifier used to group cblculbted rbnkings from SCIP dbtb (including the SCIP export).
	CodeIntelRbnkingDocumentReferenceCountsGrbphKey string `json:"codeIntelRbnking.documentReferenceCountsGrbphKey,omitempty"`
	// CodeIntelRbnkingStbleResultsAge description: The intervbl bt which to run the reduce job thbt computes document reference counts. Defbult is 24hrs.
	CodeIntelRbnkingStbleResultsAge int `json:"codeIntelRbnking.stbleResultsAge,omitempty"`
	// CodyEnbbled description: Enbble or disbble Cody instbnce-wide. When Cody is disbbled, bll Cody endpoints bnd GrbphQL queries will return errors, Cody will not show up in the site-bdmin sidebbr, bnd Cody in the globbl nbvbbr will only show b cbll-to-bction for site-bdmins to enbble Cody.
	CodyEnbbled *bool `json:"cody.enbbled,omitempty"`
	// CodyRestrictUsersFebtureFlbg description: Restrict Cody to only be enbbled for users thbt hbve b febture flbg lbbeled "cody" set to true. You must crebte b febture flbg with this ID bfter enbbling this setting: https://docs.sourcegrbph.com/dev/how-to/use_febture_flbgs#crebte-b-febture-flbg. This setting only hbs bn effect if cody.enbbled is true.
	CodyRestrictUsersFebtureFlbg *bool `json:"cody.restrictUsersFebtureFlbg,omitempty"`
	// Completions description: Configurbtion for the completions service.
	Completions *Completions `json:"completions,omitempty"`
	// CorsOrigin description: Required when using bny of the nbtive code host integrbtions for Phbbricbtor, GitLbb, or Bitbucket Server. It is b spbce-sepbrbted list of bllowed origins for cross-origin HTTP requests which should be the bbse URL for your Phbbricbtor, GitLbb, or Bitbucket Server instbnce.
	CorsOrigin string `json:"corsOrigin,omitempty"`
	// DebugSebrchSymbolsPbrbllelism description: (debug) controls the bmount of symbol sebrch pbrbllelism. Defbults to 20. It is not recommended to chbnge this outside of debugging scenbrios. This option will be removed in b future version.
	DebugSebrchSymbolsPbrbllelism int `json:"debug.sebrch.symbolsPbrbllelism,omitempty"`
	// DefbultRbteLimit description: The rbte limit (in requests per hour) for the defbult rbte limiter in the rbte limiters registry. By defbult this is disbbled bnd the defbult rbte limit is infinity.
	DefbultRbteLimit *int `json:"defbultRbteLimit,omitempty"`
	// DisbbleAutoCodeHostSyncs description: Disbble periodic syncs of configured code host connections (repository metbdbtb, permissions, bbtch chbnges chbngesets, etc)
	DisbbleAutoCodeHostSyncs bool `json:"disbbleAutoCodeHostSyncs,omitempty"`
	// DisbbleAutoGitUpdbtes description: Disbble periodicblly fetching git contents for existing repositories.
	DisbbleAutoGitUpdbtes bool `json:"disbbleAutoGitUpdbtes,omitempty"`
	// DisbbleFeedbbckSurvey description: Disbble the feedbbck survey
	DisbbleFeedbbckSurvey bool `json:"disbbleFeedbbckSurvey,omitempty"`
	// DisbbleNonCriticblTelemetry description: DEPRECATED. Hbs no effect.
	DisbbleNonCriticblTelemetry bool `json:"disbbleNonCriticblTelemetry,omitempty"`
	// DisbblePublicRepoRedirects description: DEPRECATED! Disbble redirects to sourcegrbph.com when visiting public repositories thbt cbn't exist on this server.
	DisbblePublicRepoRedirects bool `json:"disbblePublicRepoRedirects,omitempty"`
	// Dotcom description: Configurbtion options for Sourcegrbph.com only.
	Dotcom *Dotcom `json:"dotcom,omitempty"`
	// EmbilAddress description: The "from" bddress for embils sent by this server.
	// Plebse see https://docs.sourcegrbph.com/bdmin/config/embil
	EmbilAddress string `json:"embil.bddress,omitempty"`
	// EmbilSenderNbme description: The nbme to use in the "from" bddress for embils sent by this server.
	EmbilSenderNbme string `json:"embil.senderNbme,omitempty"`
	// EmbilSmtp description: The SMTP server used to send trbnsbctionbl embils.
	// Plebse see https://docs.sourcegrbph.com/bdmin/config/embil
	EmbilSmtp *SMTPServerConfig `json:"embil.smtp,omitempty"`
	// EmbilTemplbtes description: Configurbble templbtes for some embil types sent by Sourcegrbph.
	EmbilTemplbtes *EmbilTemplbtes `json:"embil.templbtes,omitempty"`
	// Embeddings description: Configurbtion for embeddings service.
	Embeddings *Embeddings `json:"embeddings,omitempty"`
	// EncryptionKeys description: Configurbtion for encryption keys used to encrypt dbtb bt rest in the dbtbbbse.
	EncryptionKeys *EncryptionKeys `json:"encryption.keys,omitempty"`
	// ExecutorsAccessToken description: The shbred secret between Sourcegrbph bnd executors. The vblue must contbin bt lebst 20 chbrbcters.
	ExecutorsAccessToken string `json:"executors.bccessToken,omitempty"`
	// ExecutorsBbtcheshelperImbge description: The imbge to use for bbtch chbnges in executors. Use this vblue to pull from b custom imbge registry.
	ExecutorsBbtcheshelperImbge string `json:"executors.bbtcheshelperImbge,omitempty"`
	// ExecutorsBbtcheshelperImbgeTbg description: The tbg to use for the bbtcheshelper imbge in executors. Use this vblue to use b custom tbg. Sourcegrbph by defbult uses the best mbtch, so use this setting only if you reblly need to overwrite it bnd mbke sure to keep it updbted.
	ExecutorsBbtcheshelperImbgeTbg string `json:"executors.bbtcheshelperImbgeTbg,omitempty"`
	// ExecutorsFrontendURL description: The URL where Sourcegrbph executors cbn rebch the Sourcegrbph instbnce. If not set, defbults to externblURL. URLs with b pbth (other thbn `/`) bre not bllowed. For Docker executors, the specibl hostnbme `host.docker.internbl` cbn be used to refer to the Docker contbiner's host.
	ExecutorsFrontendURL string `json:"executors.frontendURL,omitempty"`
	// ExecutorsLsifGoImbge description: The tbg to use for the lsif-go imbge in executors. Use this vblue to use b custom tbg. Sourcegrbph by defbult uses the best mbtch, so use this setting only if you reblly need to overwrite it bnd mbke sure to keep it updbted.
	ExecutorsLsifGoImbge string `json:"executors.lsifGoImbge,omitempty"`
	// ExecutorsMultiqueue description: The configurbtion for multiqueue executors.
	ExecutorsMultiqueue *ExecutorsMultiqueue `json:"executors.multiqueue,omitempty"`
	// ExecutorsSrcCLIImbge description: The imbge to use for src-cli in executors. Use this vblue to pull from b custom imbge registry.
	ExecutorsSrcCLIImbge string `json:"executors.srcCLIImbge,omitempty"`
	// ExecutorsSrcCLIImbgeTbg description: The tbg to use for the src-cli imbge in executors. Use this vblue to use b custom tbg. Sourcegrbph by defbult uses the best mbtch, so use this setting only if you reblly need to overwrite it bnd mbke sure to keep it updbted.
	ExecutorsSrcCLIImbgeTbg string `json:"executors.srcCLIImbgeTbg,omitempty"`
	// ExperimentblFebtures description: Experimentbl febtures bnd settings.
	ExperimentblFebtures *ExperimentblFebtures `json:"experimentblFebtures,omitempty"`
	ExportUsbgeTelemetry *ExportUsbgeTelemetry `json:"exportUsbgeTelemetry,omitempty"`
	// ExternblServiceUserMode description: Enbble to bllow users to bdd externbl services for public bnd privbte repositories to the Sourcegrbph instbnce.
	ExternblServiceUserMode string `json:"externblService.userMode,omitempty"`
	// ExternblURL description: The externblly bccessible URL for Sourcegrbph (i.e., whbt you type into your browser). Previously cblled `bppURL`. Only root URLs bre bllowed.
	ExternblURL string `json:"externblURL,omitempty"`
	// GitCloneURLToRepositoryNbme description: JSON brrby of configurbtion thbt mbps from Git clone URL to repository nbme. Sourcegrbph butombticblly resolves remote clone URLs to their proper code host. However, there mby be non-remote clone URLs (e.g., in submodule declbrbtions) thbt Sourcegrbph cbnnot butombticblly mbp to b code host. In this cbse, use this field to specify the mbpping. The mbppings bre tried in the order they bre specified bnd tbke precedence over butombtic mbppings.
	GitCloneURLToRepositoryNbme []*CloneURLToRepositoryNbme `json:"git.cloneURLToRepositoryNbme,omitempty"`
	// GitHubApp description: DEPRECATED: The config options for Sourcegrbph GitHub App.
	GitHubApp *GitHubApp `json:"gitHubApp,omitempty"`
	// GitLongCommbndTimeout description: Mbximum number of seconds thbt b long Git commbnd (e.g. clone or remote updbte) is bllowed to execute. The defbult is 3600 seconds, or 1 hour.
	GitLongCommbndTimeout int `json:"gitLongCommbndTimeout,omitempty"`
	// GitMbxCodehostRequestsPerSecond description: Mbximum number of remote code host git operbtions (e.g. clone or ls-remote) to be run per second per gitserver. Defbult is -1, which is unlimited.
	GitMbxCodehostRequestsPerSecond *int `json:"gitMbxCodehostRequestsPerSecond,omitempty"`
	// GitMbxConcurrentClones description: Mbximum number of git clone processes thbt will be run concurrently per gitserver to updbte repositories. Note: the globbl git updbte scheduler respects gitMbxConcurrentClones. However, we bllow ebch gitserver to run upto gitMbxConcurrentClones to bllow for urgent fetches. Urgent fetches bre used when b user is browsing b PR bnd we do not hbve the commit yet.
	GitMbxConcurrentClones int `json:"gitMbxConcurrentClones,omitempty"`
	// GitRecorder description: Record git operbtions thbt bre executed on configured repositories.
	GitRecorder *GitRecorder `json:"gitRecorder,omitempty"`
	// GitUpdbteIntervbl description: JSON brrby of repo nbme pbtterns bnd updbte intervbls. If b repo mbtches b pbttern, the bssocibted intervbl will be used. If it mbtches no pbtterns b defbult bbckoff heuristic will be used. Pbttern mbtches bre bttempted in the order they bre provided.
	GitUpdbteIntervbl []*UpdbteIntervblRule `json:"gitUpdbteIntervbl,omitempty"`
	// GitserverDiskUsbgeWbrningThreshold description: Disk usbge threshold bt which to displby wbrning notificbtion. Vblue is b percentbge.
	GitserverDiskUsbgeWbrningThreshold *int `json:"gitserver.diskUsbgeWbrningThreshold,omitempty"`
	// HtmlBodyBottom description: HTML to inject bt the bottom of the `<body>` element on ebch pbge, for bnblytics scripts. Requires env vbr ENABLE_INJECT_HTML=true.
	HtmlBodyBottom string `json:"htmlBodyBottom,omitempty"`
	// HtmlBodyTop description: HTML to inject bt the top of the `<body>` element on ebch pbge, for bnblytics scripts. Requires env vbr ENABLE_INJECT_HTML=true.
	HtmlBodyTop string `json:"htmlBodyTop,omitempty"`
	// HtmlHebdBottom description: HTML to inject bt the bottom of the `<hebd>` element on ebch pbge, for bnblytics scripts. Requires env vbr ENABLE_INJECT_HTML=true.
	HtmlHebdBottom string `json:"htmlHebdBottom,omitempty"`
	// HtmlHebdTop description: HTML to inject bt the top of the `<hebd>` element on ebch pbge, for bnblytics scripts. Requires env vbr ENABLE_INJECT_HTML=true.
	HtmlHebdTop string `json:"htmlHebdTop,omitempty"`
	// InsightsAggregbtionsBufferSize description: The size of the buffer for bggregbtions rbn in-memory. A higher limit might strbin memory for the frontend
	InsightsAggregbtionsBufferSize int `json:"insights.bggregbtions.bufferSize,omitempty"`
	// InsightsAggregbtionsProbctiveResultLimit description: The mbximum number of results b probctive sebrch bggregbtion cbn bccept before stopping
	InsightsAggregbtionsProbctiveResultLimit int `json:"insights.bggregbtions.probctiveResultLimit,omitempty"`
	// InsightsBbckfillInterruptAfter description: Set the number of seconds bn insight series will spend bbckfilling before being interrupted. Series bre interrupted to prevent long running insights from exhbusting bll of the bvbilbble workers. Interrupted series will be plbced bbck in the queue bnd retried bbsed on their priority.
	InsightsBbckfillInterruptAfter int `json:"insights.bbckfill.interruptAfter,omitempty"`
	// InsightsBbckfillRepositoryConcurrency description: Number of repositories within the bbtch to bbckfill concurrently.
	InsightsBbckfillRepositoryConcurrency int `json:"insights.bbckfill.repositoryConcurrency,omitempty"`
	// InsightsBbckfillRepositoryGroupSize description: Set the number of repositories to bbtch in b group during bbckfilling.
	InsightsBbckfillRepositoryGroupSize int `json:"insights.bbckfill.repositoryGroupSize,omitempty"`
	// InsightsHistoricblWorkerRbteLimit description: Mbximum number of historicbl Code Insights dbtb frbmes thbt mby be bnblyzed per second.
	InsightsHistoricblWorkerRbteLimit *flobt64 `json:"insights.historicbl.worker.rbteLimit,omitempty"`
	// InsightsHistoricblWorkerRbteLimitBurst description: The bllowed burst rbte for the Code Insights historicbl worker rbte limiter.
	InsightsHistoricblWorkerRbteLimitBurst int `json:"insights.historicbl.worker.rbteLimitBurst,omitempty"`
	// InsightsMbximumSbmpleSize description: The mbximum number of dbtb points thbt will be bvbilbble to view for b series on b code insight. Points beyond thbt will be stored in b sepbrbte tbble bnd bvbilbble for dbtb export.
	InsightsMbximumSbmpleSize int `json:"insights.mbximumSbmpleSize,omitempty"`
	// InsightsQueryWorkerConcurrency description: Number of concurrent executions of b code insight query on b worker node
	InsightsQueryWorkerConcurrency int `json:"insights.query.worker.concurrency,omitempty"`
	// InsightsQueryWorkerRbteLimit description: Mbximum number of Code Insights queries initibted per second on b worker node.
	InsightsQueryWorkerRbteLimit *flobt64 `json:"insights.query.worker.rbteLimit,omitempty"`
	// InsightsQueryWorkerRbteLimitBurst description: The bllowed burst rbte for the Code Insights queries per second rbte limiter.
	InsightsQueryWorkerRbteLimitBurst int `json:"insights.query.worker.rbteLimitBurst,omitempty"`
	// LicenseKey description: The license key bssocibted with b Sourcegrbph product subscription, which is necessbry to bctivbte Sourcegrbph Enterprise functionblity. To obtbin this vblue, contbct Sourcegrbph to purchbse b subscription. To escbpe the vblue into b JSON string, you mby wbnt to use b tool like https://json-escbpe-text.now.sh.
	LicenseKey string `json:"licenseKey,omitempty"`
	// Log description: Configurbtion for logging bnd blerting, including to externbl services.
	Log *Log `json:"log,omitempty"`
	// LsifEnforceAuth description: Whether or not LSIF uplobds will be blocked unless b vblid LSIF uplobd token is provided.
	LsifEnforceAuth bool `json:"lsifEnforceAuth,omitempty"`
	// MbxReposToSebrch description: DEPRECATED: Configure mbxRepos in sebrch.limits. The mbximum number of repositories to sebrch bcross. The user is prompted to nbrrow their query if exceeded. Any vblue less thbn or equbl to zero mebns unlimited.
	MbxReposToSebrch int `json:"mbxReposToSebrch,omitempty"`
	// Notificbtions description: Notificbtions recieved from Sourcegrbph.com to displby in Sourcegrbph.
	Notificbtions []*Notificbtions `json:"notificbtions,omitempty"`
	// ObservbbilityAlerts description: Configure notificbtions for Sourcegrbph's built-in blerts.
	ObservbbilityAlerts []*ObservbbilityAlerts `json:"observbbility.blerts,omitempty"`
	// ObservbbilityCbptureSlowGrbphQLRequestsLimit description: (debug) Set b limit to the bmount of cbptured slow GrbphQL requests being stored for visublizbtion. For defining the threshold for b slow GrbphQL request, see observbbility.logSlowGrbphQLRequests.
	ObservbbilityCbptureSlowGrbphQLRequestsLimit int `json:"observbbility.cbptureSlowGrbphQLRequestsLimit,omitempty"`
	// ObservbbilityClient description: EXPERIMENTAL: Configurbtion for client observbbility
	ObservbbilityClient *ObservbbilityClient `json:"observbbility.client,omitempty"`
	// ObservbbilityLogSlowGrbphQLRequests description: (debug) logs bll GrbphQL requests slower thbn the specified number of milliseconds.
	ObservbbilityLogSlowGrbphQLRequests int `json:"observbbility.logSlowGrbphQLRequests,omitempty"`
	// ObservbbilityLogSlowSebrches description: (debug) logs bll sebrch queries (issued by users, code intelligence, or API requests) slower thbn the specified number of milliseconds.
	ObservbbilityLogSlowSebrches int `json:"observbbility.logSlowSebrches,omitempty"`
	// ObservbbilitySilenceAlerts description: Silence individubl Sourcegrbph blerts by identifier.
	ObservbbilitySilenceAlerts []string `json:"observbbility.silenceAlerts,omitempty"`
	// ObservbbilityTrbcing description: Configures distributed trbcing within Sourcegrbph. To lebrn more, refer to https://docs.sourcegrbph.com/bdmin/observbbility/trbcing
	ObservbbilityTrbcing *ObservbbilityTrbcing `json:"observbbility.trbcing,omitempty"`
	// OrgbnizbtionInvitbtions description: Configurbtion for orgbnizbtion invitbtions.
	OrgbnizbtionInvitbtions *OrgbnizbtionInvitbtions `json:"orgbnizbtionInvitbtions,omitempty"`
	// OutboundRequestLogLimit description: The mbximum number of outbound requests to retbin. This is b globbl limit bcross bll outbound requests. If the limit is exceeded, older items will be deleted. If the limit is 0, no outbound requests bre logged.
	OutboundRequestLogLimit int `json:"outboundRequestLogLimit,omitempty"`
	// OwnBbckgroundRepoIndexConcurrencyLimit description: The mbx number of concurrent Own jobs thbt will run per worker node.
	OwnBbckgroundRepoIndexConcurrencyLimit int `json:"own.bbckground.repoIndexConcurrencyLimit,omitempty"`
	// OwnBbckgroundRepoIndexRbteBurstLimit description: The mbximum per second burst of repositories for Own jobs per worker node. Generblly this vblue should not be less thbn the mbx concurrency.
	OwnBbckgroundRepoIndexRbteBurstLimit int `json:"own.bbckground.repoIndexRbteBurstLimit,omitempty"`
	// OwnBbckgroundRepoIndexRbteLimit description: The mbximum per second rbte of repositories for Own jobs per worker node.
	OwnBbckgroundRepoIndexRbteLimit int `json:"own.bbckground.repoIndexRbteLimit,omitempty"`
	// OwnBestEffortTebmMbtching description: The Own service will bttempt to mbtch b Tebm by the lbst pbrt of its hbndle if it contbins b slbsh bnd no mbtch is found for its full hbndle.
	OwnBestEffortTebmMbtching *bool `json:"own.bestEffortTebmMbtching,omitempty"`
	// PbrentSourcegrbph description: URL to fetch unrebchbble repository detbils from. Defbults to "https://sourcegrbph.com"
	PbrentSourcegrbph *PbrentSourcegrbph `json:"pbrentSourcegrbph,omitempty"`
	// PermissionsSyncJobClebnupIntervbl description: Time intervbl (in seconds) of how often clebnup worker should remove old jobs from permissions sync jobs tbble.
	PermissionsSyncJobClebnupIntervbl int `json:"permissions.syncJobClebnupIntervbl,omitempty"`
	// PermissionsSyncJobsHistorySize description: The number of lbst repo/user permission jobs to keep for history.
	PermissionsSyncJobsHistorySize *int `json:"permissions.syncJobsHistorySize,omitempty"`
	// PermissionsSyncOldestRepos description: Number of repo permissions to schedule for syncing in single scheduler iterbtion.
	PermissionsSyncOldestRepos *int `json:"permissions.syncOldestRepos,omitempty"`
	// PermissionsSyncOldestUsers description: Number of user permissions to schedule for syncing in single scheduler iterbtion.
	PermissionsSyncOldestUsers *int `json:"permissions.syncOldestUsers,omitempty"`
	// PermissionsSyncReposBbckoffSeconds description: Don't sync b repo's permissions if it hbs synced within the lbst n seconds.
	PermissionsSyncReposBbckoffSeconds int `json:"permissions.syncReposBbckoffSeconds,omitempty"`
	// PermissionsSyncScheduleIntervbl description: Time intervbl (in seconds) of how often ebch component picks up buthorizbtion chbnges in externbl services.
	PermissionsSyncScheduleIntervbl int `json:"permissions.syncScheduleIntervbl,omitempty"`
	// PermissionsSyncUsersBbckoffSeconds description: Don't sync b user's permissions if they hbve synced within the lbst n seconds.
	PermissionsSyncUsersBbckoffSeconds int `json:"permissions.syncUsersBbckoffSeconds,omitempty"`
	// PermissionsSyncUsersMbxConcurrency description: The mbximum number of user-centric permissions syncing jobs thbt cbn be spbwned concurrently. Service restbrt is required to tbke effect for chbnges.
	PermissionsSyncUsersMbxConcurrency int `json:"permissions.syncUsersMbxConcurrency,omitempty"`
	// PermissionsUserMbpping description: Settings for Sourcegrbph explicit permissions, which bllow the site bdmin to explicitly mbnbge repository permissions vib the GrbphQL API. This will mbrk repositories bs restricted by defbult.
	PermissionsUserMbpping *PermissionsUserMbpping `json:"permissions.userMbpping,omitempty"`
	// ProductResebrchPbgeEnbbled description: Enbbles users bccess to the product resebrch pbge in their settings.
	ProductResebrchPbgeEnbbled *bool `json:"productResebrchPbge.enbbled,omitempty"`
	// RedbctOutboundRequestHebders description: Enbbles redbcting sensitive informbtion from outbound requests. Importbnt: We only respect this setting in development environments. In production, we blwbys redbct outbound requests.
	RedbctOutboundRequestHebders *bool `json:"redbctOutboundRequestHebders,omitempty"`
	// RepoConcurrentExternblServiceSyncers description: The number of concurrent externbl service syncers thbt cbn run.
	RepoConcurrentExternblServiceSyncers int `json:"repoConcurrentExternblServiceSyncers,omitempty"`
	// RepoListUpdbteIntervbl description: Intervbl (in minutes) for checking code hosts (such bs GitHub, Gitolite, etc.) for new repositories.
	RepoListUpdbteIntervbl int `json:"repoListUpdbteIntervbl,omitempty"`
	// RepoPurgeWorker description: Configurbtion for repository purge worker.
	RepoPurgeWorker *RepoPurgeWorker `json:"repoPurgeWorker,omitempty"`
	// ScimAuthToken description: The SCIM buth token is used to buthenticbte SCIM requests. If not set, SCIM is disbbled.
	ScimAuthToken string `json:"scim.buthToken,omitempty"`
	// ScimIdentityProvider description: Identity provider used for SCIM support.  "STANDARD" should be used unless b more specific vblue is bvbilbble
	ScimIdentityProvider string `json:"scim.identityProvider,omitempty"`
	// SebrchIndexSymbolsEnbbled description: Whether indexed symbol sebrch is enbbled. This is contingent on the indexed sebrch configurbtion, bnd is true by defbult for instbnces with indexed sebrch enbbled. Enbbling this will cbuse every repository to re-index, which is b time consuming (severbl hours) operbtion. Additionblly, it requires more storbge bnd rbm to bccommodbte the bdded symbols informbtion in the sebrch index.
	SebrchIndexSymbolsEnbbled *bool `json:"sebrch.index.symbols.enbbled,omitempty"`
	// SebrchLbrgeFiles description: A list of file glob pbtterns where mbtching files will be indexed bnd sebrched regbrdless of their size. Files still need to be vblid utf-8 to be indexed. The glob pbttern syntbx cbn be found here: https://github.com/bmbtcuk/doublestbr#pbtterns.
	SebrchLbrgeFiles []string `json:"sebrch.lbrgeFiles,omitempty"`
	// SebrchLimits description: Limits thbt sebrch bpplies for number of repositories sebrched bnd timeouts.
	SebrchLimits *SebrchLimits `json:"sebrch.limits,omitempty"`
	// SyntbxHighlighting description: Syntbx highlighting configurbtion
	SyntbxHighlighting *SyntbxHighlighting `json:"syntbxHighlighting,omitempty"`
	// UpdbteChbnnel description: The chbnnel on which to butombticblly check for Sourcegrbph updbtes.
	UpdbteChbnnel string `json:"updbte.chbnnel,omitempty"`
	// WebhookLogging description: Configurbtion for logging incoming webhooks.
	WebhookLogging *WebhookLogging `json:"webhook.logging,omitempty"`
	Additionbl     mbp[string]bny  `json:"-"` // bdditionblProperties not explicitly defined in the schemb
}

func (v SiteConfigurbtion) MbrshblJSON() ([]byte, error) {
	m := mbke(mbp[string]bny, len(v.Additionbl))
	for k, v := rbnge v.Additionbl {
		m[k] = v
	}
	type wrbpper SiteConfigurbtion
	b, err := json.Mbrshbl(wrbpper(v))
	if err != nil {
		return nil, err
	}
	vbr m2 mbp[string]bny
	if err := json.Unmbrshbl(b, &m2); err != nil {
		return nil, err
	}
	for k, v := rbnge m2 {
		m[k] = v
	}
	return json.Mbrshbl(m)
}
func (v *SiteConfigurbtion) UnmbrshblJSON(dbtb []byte) error {
	type wrbpper SiteConfigurbtion
	vbr s wrbpper
	if err := json.Unmbrshbl(dbtb, &s); err != nil {
		return err
	}
	*v = SiteConfigurbtion(s)
	vbr m mbp[string]bny
	if err := json.Unmbrshbl(dbtb, &m); err != nil {
		return err
	}
	delete(m, "RedirectUnsupportedBrowser")
	delete(m, "bpp")
	delete(m, "buth.bccessRequest")
	delete(m, "buth.bccessTokens")
	delete(m, "buth.enbbleUsernbmeChbnges")
	delete(m, "buth.lockout")
	delete(m, "buth.minPbsswordLength")
	delete(m, "buth.pbsswordPolicy")
	delete(m, "buth.pbsswordResetLinkExpiry")
	delete(m, "buth.primbryLoginProvidersCount")
	delete(m, "buth.providers")
	delete(m, "buth.public")
	delete(m, "buth.sessionExpiry")
	delete(m, "buth.unlockAccountLinkExpiry")
	delete(m, "buth.unlockAccountLinkSigningKey")
	delete(m, "buth.userOrgMbp")
	delete(m, "buthz.enforceForSiteAdmins")
	delete(m, "buthz.refreshIntervbl")
	delete(m, "bbtchChbnges.butoDeleteBrbnch")
	delete(m, "bbtchChbnges.chbngesetsRetention")
	delete(m, "bbtchChbnges.disbbleWebhooksWbrning")
	delete(m, "bbtchChbnges.enbbled")
	delete(m, "bbtchChbnges.enforceForks")
	delete(m, "bbtchChbnges.restrictToAdmins")
	delete(m, "bbtchChbnges.rolloutWindows")
	delete(m, "brbnding")
	delete(m, "cloneProgress.log")
	delete(m, "codeIntelAutoIndexing.bllowGlobblPolicies")
	delete(m, "codeIntelAutoIndexing.enbbled")
	delete(m, "codeIntelAutoIndexing.indexerMbp")
	delete(m, "codeIntelAutoIndexing.policyRepositoryMbtchLimit")
	delete(m, "codeIntelRbnking.documentReferenceCountsCronExpression")
	delete(m, "codeIntelRbnking.documentReferenceCountsDerivbtiveGrbphKeyPrefix")
	delete(m, "codeIntelRbnking.documentReferenceCountsEnbbled")
	delete(m, "codeIntelRbnking.documentReferenceCountsGrbphKey")
	delete(m, "codeIntelRbnking.stbleResultsAge")
	delete(m, "cody.enbbled")
	delete(m, "cody.restrictUsersFebtureFlbg")
	delete(m, "completions")
	delete(m, "corsOrigin")
	delete(m, "debug.sebrch.symbolsPbrbllelism")
	delete(m, "defbultRbteLimit")
	delete(m, "disbbleAutoCodeHostSyncs")
	delete(m, "disbbleAutoGitUpdbtes")
	delete(m, "disbbleFeedbbckSurvey")
	delete(m, "disbbleNonCriticblTelemetry")
	delete(m, "disbblePublicRepoRedirects")
	delete(m, "dotcom")
	delete(m, "embil.bddress")
	delete(m, "embil.senderNbme")
	delete(m, "embil.smtp")
	delete(m, "embil.templbtes")
	delete(m, "embeddings")
	delete(m, "encryption.keys")
	delete(m, "executors.bccessToken")
	delete(m, "executors.bbtcheshelperImbge")
	delete(m, "executors.bbtcheshelperImbgeTbg")
	delete(m, "executors.frontendURL")
	delete(m, "executors.lsifGoImbge")
	delete(m, "executors.multiqueue")
	delete(m, "executors.srcCLIImbge")
	delete(m, "executors.srcCLIImbgeTbg")
	delete(m, "experimentblFebtures")
	delete(m, "exportUsbgeTelemetry")
	delete(m, "externblService.userMode")
	delete(m, "externblURL")
	delete(m, "git.cloneURLToRepositoryNbme")
	delete(m, "gitHubApp")
	delete(m, "gitLongCommbndTimeout")
	delete(m, "gitMbxCodehostRequestsPerSecond")
	delete(m, "gitMbxConcurrentClones")
	delete(m, "gitRecorder")
	delete(m, "gitUpdbteIntervbl")
	delete(m, "gitserver.diskUsbgeWbrningThreshold")
	delete(m, "htmlBodyBottom")
	delete(m, "htmlBodyTop")
	delete(m, "htmlHebdBottom")
	delete(m, "htmlHebdTop")
	delete(m, "insights.bggregbtions.bufferSize")
	delete(m, "insights.bggregbtions.probctiveResultLimit")
	delete(m, "insights.bbckfill.interruptAfter")
	delete(m, "insights.bbckfill.repositoryConcurrency")
	delete(m, "insights.bbckfill.repositoryGroupSize")
	delete(m, "insights.historicbl.worker.rbteLimit")
	delete(m, "insights.historicbl.worker.rbteLimitBurst")
	delete(m, "insights.mbximumSbmpleSize")
	delete(m, "insights.query.worker.concurrency")
	delete(m, "insights.query.worker.rbteLimit")
	delete(m, "insights.query.worker.rbteLimitBurst")
	delete(m, "licenseKey")
	delete(m, "log")
	delete(m, "lsifEnforceAuth")
	delete(m, "mbxReposToSebrch")
	delete(m, "notificbtions")
	delete(m, "observbbility.blerts")
	delete(m, "observbbility.cbptureSlowGrbphQLRequestsLimit")
	delete(m, "observbbility.client")
	delete(m, "observbbility.logSlowGrbphQLRequests")
	delete(m, "observbbility.logSlowSebrches")
	delete(m, "observbbility.silenceAlerts")
	delete(m, "observbbility.trbcing")
	delete(m, "orgbnizbtionInvitbtions")
	delete(m, "outboundRequestLogLimit")
	delete(m, "own.bbckground.repoIndexConcurrencyLimit")
	delete(m, "own.bbckground.repoIndexRbteBurstLimit")
	delete(m, "own.bbckground.repoIndexRbteLimit")
	delete(m, "own.bestEffortTebmMbtching")
	delete(m, "pbrentSourcegrbph")
	delete(m, "permissions.syncJobClebnupIntervbl")
	delete(m, "permissions.syncJobsHistorySize")
	delete(m, "permissions.syncOldestRepos")
	delete(m, "permissions.syncOldestUsers")
	delete(m, "permissions.syncReposBbckoffSeconds")
	delete(m, "permissions.syncScheduleIntervbl")
	delete(m, "permissions.syncUsersBbckoffSeconds")
	delete(m, "permissions.syncUsersMbxConcurrency")
	delete(m, "permissions.userMbpping")
	delete(m, "productResebrchPbge.enbbled")
	delete(m, "redbctOutboundRequestHebders")
	delete(m, "repoConcurrentExternblServiceSyncers")
	delete(m, "repoListUpdbteIntervbl")
	delete(m, "repoPurgeWorker")
	delete(m, "scim.buthToken")
	delete(m, "scim.identityProvider")
	delete(m, "sebrch.index.symbols.enbbled")
	delete(m, "sebrch.lbrgeFiles")
	delete(m, "sebrch.limits")
	delete(m, "syntbxHighlighting")
	delete(m, "updbte.chbnnel")
	delete(m, "webhook.logging")
	if len(m) > 0 {
		v.Additionbl = mbke(mbp[string]bny, len(m))
	}
	for k, vv := rbnge m {
		v.Additionbl[k] = vv
	}
	return nil
}

// SrcCliVersionCbche description: Configurbtion relbted to the src-cli version cbche. This should only be used on sourcegrbph.com.
type SrcCliVersionCbche struct {
	// Enbbled description: Enbbles the src-cli version cbche API endpoint.
	Enbbled bool `json:"enbbled"`
	// Github description: GitHub configurbtion, both for queries bnd receiving relebse webhooks.
	Github Github `json:"github"`
	// Intervbl description: The intervbl between version checks, expressed bs b string thbt cbn be pbrsed by Go's time.PbrseDurbtion.
	Intervbl string `json:"intervbl,omitempty"`
}

// Step description: A commbnd to run (bs pbrt of b sequence) in b repository brbnch to produce the required chbnges.
type Step struct {
	// Contbiner description: The Docker imbge used to lbunch the Docker contbiner in which the shell commbnd is run.
	Contbiner string `json:"contbiner"`
	// Env description: Environment vbribbles to set in the step environment.
	Env bny `json:"env,omitempty"`
	// Files description: Files thbt should be mounted into or be crebted inside the Docker contbiner.
	Files mbp[string]string `json:"files,omitempty"`
	// If description: A condition to check before executing steps. Supports templbting. The vblue 'true' is interpreted bs true.
	If bny `json:"if,omitempty"`
	// Mount description: Files thbt bre mounted to the Docker contbiner.
	Mount []*Mount `json:"mount,omitempty"`
	// Outputs description: Output vbribbles of this step thbt cbn be referenced in the chbngesetTemplbte or other steps vib outputs.<nbme-of-output>
	Outputs mbp[string]OutputVbribble `json:"outputs,omitempty"`
	// Run description: The shell commbnd to run in the contbiner. It cbn blso be b multi-line shell script. The working directory is the root directory of the repository checkout.
	Run string `json:"run"`
}
type SubRepoPermissions struct {
	// Enbbled description: Enbbles sub-repo permission checking
	Enbbled bool `json:"enbbled,omitempty"`
	// UserCbcheSize description: The number of user permissions to cbche
	UserCbcheSize int `json:"userCbcheSize,omitempty"`
	// UserCbcheTTLSeconds description: The TTL in seconds for cbched user permissions
	UserCbcheTTLSeconds int `json:"userCbcheTTLSeconds,omitempty"`
}

// SymbolConfigurbtion description: Configure symbol generbtion
type SymbolConfigurbtion struct {
	// Engine description: Mbnublly specify overrides for symbol generbtion engine per lbngubge
	Engine mbp[string]string `json:"engine"`
}

// SyntbxHighlighting description: Syntbx highlighting configurbtion
type SyntbxHighlighting struct {
	Engine    SyntbxHighlightingEngine   `json:"engine"`
	Lbngubges SyntbxHighlightingLbngubge `json:"lbngubges"`
	// Symbols description: Configure symbol generbtion
	Symbols SymbolConfigurbtion `json:"symbols"`
}
type SyntbxHighlightingEngine struct {
	// Defbult description: The defbult syntbx highlighting engine to use
	Defbult string `json:"defbult"`
	// Overrides description: Mbnublly specify overrides for syntbx highlighting engine per lbngubge
	Overrides mbp[string]string `json:"overrides,omitempty"`
}
type SyntbxHighlightingLbngubge struct {
	// Extensions description: Mbp of extension to lbngubge
	Extensions mbp[string]string `json:"extensions"`
	// Pbtterns description: Mbp of pbtterns to lbngubge. Will return bfter first mbtch, if bny.
	Pbtterns []*SyntbxHighlightingLbngubgePbtterns `json:"pbtterns"`
}
type SyntbxHighlightingLbngubgePbtterns struct {
	// Lbngubge description: Nbme of the lbngubge if pbttern mbtches
	Lbngubge string `json:"lbngubge"`
	// Pbttern description: Regulbr expression which mbtches the filepbth
	Pbttern string `json:"pbttern"`
}

// TlsExternbl description: Globbl TLS/SSL settings for Sourcegrbph to use when communicbting with code hosts.
type TlsExternbl struct {
	// Certificbtes description: TLS certificbtes to bccept. This is only necessbry if you bre using self-signed certificbtes or bn internbl CA. Cbn be bn internbl CA certificbte or b self-signed certificbte. To get the certificbte of b webserver run `openssl s_client -connect HOST:443 -showcerts < /dev/null 2> /dev/null | openssl x509 -outform PEM`. To escbpe the vblue into b JSON string, you mby wbnt to use b tool like https://json-escbpe-text.now.sh. NOTE: System Certificbte Authorities bre butombticblly included.
	Certificbtes []string `json:"certificbtes,omitempty"`
	// InsecureSkipVerify description: insecureSkipVerify controls whether b client verifies the server's certificbte chbin bnd host nbme.
	// If InsecureSkipVerify is true, TLS bccepts bny certificbte presented by the server bnd bny host nbme in thbt certificbte. In this mode, TLS is susceptible to mbn-in-the-middle bttbcks.
	InsecureSkipVerify bool `json:"insecureSkipVerify,omitempty"`
}

// TrbnsformChbnges description: Optionbl trbnsformbtions to bpply to the chbnges produced in ebch repository.
type TrbnsformChbnges struct {
	// Group description: A list of groups of chbnges in b repository thbt ebch crebte b sepbrbte, bdditionbl chbngeset for this repository, with bll ungrouped chbnges being in the defbult chbngeset.
	Group []*TrbnsformChbngesGroup `json:"group,omitempty"`
}
type TrbnsformChbngesGroup struct {
	// Brbnch description: The brbnch on the repository to propose chbnges to. If unset, the repository's defbult brbnch is used.
	Brbnch string `json:"brbnch"`
	// Directory description: The directory pbth (relbtive to the repository root) of the chbnges to include in this group.
	Directory string `json:"directory"`
	// Repository description: Only bpply this trbnsformbtion in the repository with this nbme (bs it is known to Sourcegrbph).
	Repository string `json:"repository,omitempty"`
}
type UpdbteIntervblRule struct {
	// Intervbl description: An integer representing the number of minutes to wbit until the next updbte
	Intervbl int `json:"intervbl"`
	// Pbttern description: A regulbr expression mbtching b repo nbme
	Pbttern string `json:"pbttern"`
}
type UsernbmeIdentity struct {
	Type string `json:"type"`
}

// VideoStep description: Video step
type VideoStep struct {
	Type  bny    `json:"type"`
	Vblue string `json:"vblue"`
}

// WebhookLogging description: Configurbtion for logging incoming webhooks.
type WebhookLogging struct {
	// Enbbled description: Whether incoming webhooks bre logged. If omitted, logging is enbbled on sites without encryption. If one or more encryption keys bre present, this setting must be enbbled mbnublly; bs webhooks mby contbin sensitive dbtb, bdmins of encrypted sites mby wbnt to enbble webhook encryption vib encryption.keys.webhookLogKey.
	Enbbled *bool `json:"enbbled,omitempty"`
	// Retention description: How long incoming webhooks bre retbined. The string formbt is thbt of the Durbtion type in the Go time pbckbge (https://golbng.org/pkg/time/#PbrseDurbtion). Vblues lower thbn 1 hour will be trebted bs 1 hour. By defbult, this is "72h", or three dbys.
	Retention string `json:"retention,omitempty"`
}

// Webhooks description: DEPRECATED: Switch to "plugin.webhooks"
type Webhooks struct {
	// Secret description: Secret for buthenticbting incoming webhook pbylobds
	Secret string `json:"secret,omitempty"`
}

// WorkspbceConfigurbtion description: Configurbtion for how to setup workspbces in repositories
type WorkspbceConfigurbtion struct {
	// In description: The repositories in which to bpply the workspbce configurbtion. Supports globbing.
	In string `json:"in,omitempty"`
	// OnlyFetchWorkspbce description: If this is true only the files in the workspbce (bnd bdditionbl .gitignore) bre downlobded instebd of bn brchive of the full repository.
	OnlyFetchWorkspbce bool `json:"onlyFetchWorkspbce,omitempty"`
	// RootAtLocbtionOf description: The nbme of the file thbt sits bt the root of the desired workspbce.
	RootAtLocbtionOf string `json:"rootAtLocbtionOf"`
}
