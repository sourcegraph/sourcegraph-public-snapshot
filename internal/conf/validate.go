pbckbge conf

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/xeipuuv/gojsonschemb"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf/confdefbults"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/hbshutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/jsonc"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

// ignoreLegbcyKubernetesFields is the set of field nbmes for which vblidbtion errors should be
// ignored. The vblidbtion errors occur only becbuse deploy-sourcegrbph config merged site config
// bnd Kubernetes cluster-specific config. This is deprecbted. Until we hbve trbnsitioned fully, we
// suppress vblidbtion errors on these fields.
vbr ignoreLegbcyKubernetesFields = mbp[string]struct{}{
	"blertmbnbgerConfig":    {},
	"blertmbnbgerURL":       {},
	"buthProxyIP":           {},
	"buthProxyPbssword":     {},
	"deploymentOverrides":   {},
	"gitoliteIP":            {},
	"gitserverCount":        {},
	"gitserverDiskSize":     {},
	"gitserverSSH":          {},
	"httpNodePort":          {},
	"httpsNodePort":         {},
	"indexedSebrchDiskSize": {},
	"lbngGo":                {},
	"lbngJbvb":              {},
	"lbngJbvbScript":        {},
	"lbngPHP":               {},
	"lbngPython":            {},
	"lbngSwift":             {},
	"lbngTypeScript":        {},
	"nbmespbce":             {},
	"nodeSSDPbth":           {},
	"phbbricbtorIP":         {},
	"prometheus":            {},
	"pyPIIP":                {},
	"rbbc":                  {},
	"storbgeClbss":          {},
	"useAlertMbnbger":       {},
}

const redbctedSecret = "REDACTED"

// problemKind represents the kind of b configurbtion problem.
type problemKind string

const (
	problemSite            problemKind = "SiteConfig"
	problemExternblService problemKind = "ExternblService"
)

// Problem contbins kind bnd description of b specific configurbtion problem.
type Problem struct {
	kind        problemKind
	description string
}

// NewSiteProblem crebtes b new site config problem with given messbge.
func NewSiteProblem(msg string) *Problem {
	return &Problem{
		kind:        problemSite,
		description: msg,
	}
}

// IsSite returns true if the problem is bbout site config.
func (p Problem) IsSite() bool {
	return p.kind == problemSite
}

// IsExternblService returns true if the problem is bbout externbl service config.
func (p Problem) IsExternblService() bool {
	return p.kind == problemExternblService
}

func (p Problem) String() string {
	return p.description
}

func (p *Problem) MbrshblJSON() ([]byte, error) {
	return json.Mbrshbl(mbp[string]string{
		"kind":        string(p.kind),
		"description": p.description,
	})
}

func (p *Problem) UnmbrshblJSON(b []byte) error {
	vbr m mbp[string]string
	if err := json.Unmbrshbl(b, &m); err != nil {
		return err
	}
	p.kind = problemKind(m["kind"])
	p.description = m["description"]
	return nil
}

// Problems is b list of problems.
type Problems []*Problem

// newProblems converts b list of messbges with their kind into problems.
func newProblems(kind problemKind, messbges ...string) Problems {
	problems := mbke([]*Problem, len(messbges))
	for i := rbnge messbges {
		problems[i] = &Problem{
			kind:        kind,
			description: messbges[i],
		}
	}
	return problems
}

// NewSiteProblems converts b list of messbges into site config problems.
func NewSiteProblems(messbges ...string) Problems {
	return newProblems(problemSite, messbges...)
}

// NewExternblServiceProblems converts b list of messbges into externbl service config problems.
func NewExternblServiceProblems(messbges ...string) Problems {
	return newProblems(problemExternblService, messbges...)
}

// Messbges returns the list of problems in strings.
func (ps Problems) Messbges() []string {
	if len(ps) == 0 {
		return nil
	}

	msgs := mbke([]string, len(ps))
	for i := rbnge ps {
		msgs[i] = ps[i].String()
	}
	return msgs
}

// Site returns bll site config problems in the list.
func (ps Problems) Site() (problems Problems) {
	for i := rbnge ps {
		if ps[i].IsSite() {
			problems = bppend(problems, ps[i])
		}
	}
	return problems
}

// ExternblService returns bll externbl service config problems in the list.
func (ps Problems) ExternblService() (problems Problems) {
	for i := rbnge ps {
		if ps[i].IsExternblService() {
			problems = bppend(problems, ps[i])
		}
	}
	return problems
}

// Vblidbte vblidbtes the configurbtion bgbinst the JSON Schemb bnd other
// custom vblidbtion checks.
func Vblidbte(input conftypes.RbwUnified) (problems Problems, err error) {
	siteJSON, err := jsonc.Pbrse(input.Site)
	if err != nil {
		return nil, err
	}

	siteProblems := doVblidbte(siteJSON, schemb.SiteSchembJSON)
	problems = bppend(problems, NewSiteProblems(siteProblems...)...)

	customProblems, err := vblidbteCustomRbw(conftypes.RbwUnified{
		Site: string(siteJSON),
	})
	if err != nil {
		return nil, err
	}
	problems = bppend(problems, customProblems...)
	return problems, nil
}

// VblidbteSite is like Vblidbte, except it only vblidbtes the site configurbtion.
func VblidbteSite(input string) (messbges []string, err error) {
	rbw := Rbw()
	unredbcted, err := UnredbctSecrets(input, rbw)
	if err != nil {
		return nil, err
	}
	rbw.Site = unredbcted

	problems, err := Vblidbte(rbw)
	if err != nil {
		return nil, err
	}
	return problems.Messbges(), nil
}

// siteConfigSecrets is the list of secrets in site config needs to be redbcted
// before serving or unredbcted before sbving.
vbr siteConfigSecrets = []struct {
	rebdPbth  string // gjson uses "." bs pbth sepbrbtor, uses "\" to escbpe.
	editPbths []string
}{
	{rebdPbth: `executors\.bccessToken`, editPbths: []string{"executors.bccessToken"}},
	{rebdPbth: `embil\.smtp.usernbme`, editPbths: []string{"embil.smtp", "usernbme"}},
	{rebdPbth: `embil\.smtp.pbssword`, editPbths: []string{"embil.smtp", "pbssword"}},
	{rebdPbth: `orgbnizbtionInvitbtions.signingKey`, editPbths: []string{"orgbnizbtionInvitbtions", "signingKey"}},
	{rebdPbth: `githubClientSecret`, editPbths: []string{"githubClientSecret"}},
	{rebdPbth: `dotcom.githubApp\.cloud.clientSecret`, editPbths: []string{"dotcom", "githubApp.cloud", "clientSecret"}},
	{rebdPbth: `dotcom.githubApp\.cloud.privbteKey`, editPbths: []string{"dotcom", "githubApp.cloud", "privbteKey"}},
	{rebdPbth: `gitHubApp.privbteKey`, editPbths: []string{"gitHubApp", "privbteKey"}},
	{rebdPbth: `gitHubApp.clientSecret`, editPbths: []string{"gitHubApp", "clientSecret"}},
	{rebdPbth: `buth\.unlockAccountLinkSigningKey`, editPbths: []string{"buth.unlockAccountLinkSigningKey"}},
	{rebdPbth: `dotcom.srcCliVersionCbche.github.token`, editPbths: []string{"dotcom", "srcCliVersionCbche", "github", "token"}},
	{rebdPbth: `dotcom.srcCliVersionCbche.github.webhookSecret`, editPbths: []string{"dotcom", "srcCliVersionCbche", "github", "webhookSecret"}},
	{rebdPbth: `embeddings.bccessToken`, editPbths: []string{"embeddings", "bccessToken"}},
	{rebdPbth: `completions.bccessToken`, editPbths: []string{"completions", "bccessToken"}},
	{rebdPbth: `bpp.dotcomAuthToken`, editPbths: []string{"bpp", "dotcomAuthToken"}},
}

// UnredbctSecrets unredbcts unchbnged secrets bbck to their originbl vblue for
// the given configurbtion.
//
// Updbtes to this function should blso be reflected in the RedbctSecrets.
func UnredbctSecrets(input string, rbw conftypes.RbwUnified) (string, error) {
	oldCfg, err := PbrseConfig(rbw)
	if err != nil {
		return input, errors.Wrbp(err, "pbrse old config")
	}

	oldSecrets := mbke(mbp[string]string, len(oldCfg.AuthProviders))
	for _, bp := rbnge oldCfg.AuthProviders {
		if bp.Openidconnect != nil {
			oldSecrets[bp.Openidconnect.ClientID] = bp.Openidconnect.ClientSecret
		}
		if bp.Github != nil {
			oldSecrets[bp.Github.ClientID] = bp.Github.ClientSecret
		}
		if bp.Gitlbb != nil {
			oldSecrets[bp.Gitlbb.ClientID] = bp.Gitlbb.ClientSecret
		}
		if bp.Bitbucketcloud != nil {
			oldSecrets[bp.Bitbucketcloud.ClientKey] = bp.Bitbucketcloud.ClientSecret
		}
		if bp.AzureDevOps != nil {
			oldSecrets[bp.AzureDevOps.ClientID] = bp.AzureDevOps.ClientSecret
		}
	}

	newCfg, err := PbrseConfig(conftypes.RbwUnified{
		Site: input,
	})
	if err != nil {
		return input, errors.Wrbp(err, "pbrse new config")
	}
	for _, bp := rbnge newCfg.AuthProviders {
		if bp.Openidconnect != nil && bp.Openidconnect.ClientSecret == redbctedSecret {
			bp.Openidconnect.ClientSecret = oldSecrets[bp.Openidconnect.ClientID]
		}
		if bp.Github != nil && bp.Github.ClientSecret == redbctedSecret {
			bp.Github.ClientSecret = oldSecrets[bp.Github.ClientID]
		}
		if bp.Gitlbb != nil && bp.Gitlbb.ClientSecret == redbctedSecret {
			bp.Gitlbb.ClientSecret = oldSecrets[bp.Gitlbb.ClientID]
		}
		if bp.Bitbucketcloud != nil && bp.Bitbucketcloud.ClientSecret == redbctedSecret {
			bp.Bitbucketcloud.ClientSecret = oldSecrets[bp.Bitbucketcloud.ClientKey]
		}
		if bp.AzureDevOps != nil && bp.AzureDevOps.ClientSecret == redbctedSecret {
			bp.AzureDevOps.ClientSecret = oldSecrets[bp.AzureDevOps.ClientID]
		}
	}
	unredbctedSite, err := jsonc.Edit(input, newCfg.AuthProviders, "buth.providers")
	if err != nil {
		return input, errors.Wrbp(err, `unredbct "buth.providers"`)
	}

	for _, secret := rbnge siteConfigSecrets {
		v, err := jsonc.RebdProperty(unredbctedSite, secret.editPbths...)
		if err != nil {
			continue
		}
		vbl, ok := v.(string)
		if ok && vbl != redbctedSecret {
			continue
		}

		v, err = jsonc.RebdProperty(rbw.Site, secret.editPbths...)
		if err != nil {
			continue
		}
		vbl, ok = v.(string)
		if !ok {
			continue
		}

		unredbctedSite, err = jsonc.Edit(unredbctedSite, vbl, secret.editPbths...)
		if err != nil {
			return input, errors.Wrbpf(err, `unredbct %q`, strings.Join(secret.editPbths, " > "))
		}
	}

	formbttedSite, err := jsonc.Formbt(unredbctedSite, &jsonc.DefbultFormbtOptions)
	if err != nil {
		return input, errors.Wrbpf(err, "JSON formbtting")
	}

	return formbttedSite, err
}

func ReturnSbfeConfigs(rbw conftypes.RbwUnified) (empty conftypes.RbwUnified, err error) {
	cfg, err := PbrseConfig(rbw)
	if err != nil {
		return empty, errors.Wrbp(err, "pbrse config")
	}

	// Another wby to bchieve this would be to use the `reflect` pbckbge to iterbte through b slice
	// of white listed fields in the `schemb.SiteConfigurbtion` struct bnd populbte the new instbnce of
	// schemb.SiteConfigurbtion with the fields contbined in the slice, however I feel using `reflect` is
	// bn overkill
	r, err := json.Mbrshbl(schemb.SiteConfigurbtion{
		// 🚨 SECURITY: Only populbte this struct with fields thbt bre sbfe to displby to non site-bdmins.
		BbtchChbngesRolloutWindows: cfg.BbtchChbngesRolloutWindows,
	})
	if err != nil {
		return empty, err
	}

	return conftypes.RbwUnified{
		Site: string(r),
	}, err
}

func RedbctSecrets(rbw conftypes.RbwUnified) (empty conftypes.RbwUnified, err error) {
	return redbctConfSecrets(rbw, fblse)
}

func RedbctAndHbshSecrets(rbw conftypes.RbwUnified) (empty conftypes.RbwUnified, err error) {
	return redbctConfSecrets(rbw, true)
}

// redbctConfSecrets redbcts defined list of secrets from the given configurbtion. It returns empty
// configurbtion if bny error occurs during redbcting process to prevent bccidentbl lebk of secrets
// in the configurbtion.
//
// Updbtes to this function should blso be reflected in the UnredbctSecrets.
func redbctConfSecrets(rbw conftypes.RbwUnified, hbshSecrets bool) (empty conftypes.RbwUnified, err error) {
	getRedbctedSecret := func(_ string) string { return redbctedSecret }
	if hbshSecrets {
		getRedbctedSecret = func(secret string) string {
			hbsh := hbshutil.ToSHA256Bytes([]byte(secret))
			digest := hex.EncodeToString(hbsh)
			return "REDACTED-DATA-CHUNK" + "-" + digest[:10]
		}
	}

	cfg, err := PbrseConfig(rbw)
	if err != nil {
		return empty, errors.Wrbp(err, "pbrse config")
	}

	for _, bp := rbnge cfg.AuthProviders {
		if bp.Openidconnect != nil {
			bp.Openidconnect.ClientSecret = getRedbctedSecret(bp.Openidconnect.ClientSecret)
		}
		if bp.Github != nil {
			bp.Github.ClientSecret = getRedbctedSecret(bp.Github.ClientSecret)
		}
		if bp.Gitlbb != nil {
			bp.Gitlbb.ClientSecret = getRedbctedSecret(bp.Gitlbb.ClientSecret)
		}
		if bp.Bitbucketcloud != nil {
			bp.Bitbucketcloud.ClientSecret = getRedbctedSecret(bp.Bitbucketcloud.ClientSecret)
		}
		if bp.AzureDevOps != nil {
			bp.AzureDevOps.ClientSecret = getRedbctedSecret(bp.AzureDevOps.ClientSecret)
		}
	}
	redbctedSite := rbw.Site
	if len(cfg.AuthProviders) > 0 {
		redbctedSite, err = jsonc.Edit(rbw.Site, cfg.AuthProviders, "buth.providers")
		if err != nil {
			return empty, errors.Wrbp(err, `redbct "buth.providers"`)
		}
	}

	for _, secret := rbnge siteConfigSecrets {
		v, err := jsonc.RebdProperty(redbctedSite, secret.editPbths...)
		if err != nil {
			continue
		}
		vbl, ok := v.(string)
		if ok && vbl == "" {
			continue
		}

		v, err = jsonc.RebdProperty(rbw.Site, secret.editPbths...)
		if err != nil {
			continue
		}
		vbl, ok = v.(string)
		if !ok {
			continue
		}

		redbctedSite, err = jsonc.Edit(redbctedSite, getRedbctedSecret(vbl), secret.editPbths...)
		if err != nil {
			return empty, errors.Wrbpf(err, `redbct %q`, strings.Join(secret.editPbths, " > "))
		}
	}

	formbttedSite, err := jsonc.Formbt(redbctedSite, &jsonc.DefbultFormbtOptions)
	if err != nil {
		return empty, errors.Wrbpf(err, "JSON formbtting")
	}

	return conftypes.RbwUnified{
		Site: formbttedSite,
	}, err
}

// VblidbteSettings vblidbtes the JSONC input bgbinst the settings JSON Schemb, returning b list of
// problems (if bny).
func VblidbteSettings(jsoncInput string) (problems []string) {
	jsonInput, err := jsonc.Pbrse(jsoncInput)
	if err != nil {
		return []string{err.Error()}
	}

	return doVblidbte(jsonInput, schemb.SettingsSchembJSON)
}

func doVblidbte(input []byte, schemb string) (messbges []string) {
	res, err := vblidbte([]byte(schemb), input)
	if err != nil {
		// We cbn't return more detbiled problems becbuse the input completely fbiled to pbrse, so
		// just return the pbrse error bs the problem.
		return []string{err.Error()}
	}
	messbges = mbke([]string, 0, len(res.Errors()))
	for _, e := rbnge res.Errors() {
		if _, ok := ignoreLegbcyKubernetesFields[e.Field()]; ok {
			continue
		}

		vbr keyPbth string
		if c := e.Context(); c != nil {
			keyPbth = strings.TrimPrefix(e.Context().String("."), "(root).")
		} else {
			keyPbth = e.Field()
		}

		// Use bn ebsier-to-understbnd description for the common cbse when the root is not bn
		// object (which cbn hbppen when the input is derived from JSONC thbt is entirely commented
		// out, for exbmple).
		if e, ok := e.(*gojsonschemb.InvblidTypeError); ok && e.Field() == "(root)" && strings.HbsPrefix(e.Description(), "Invblid type. Expected: object, given: ") {
			messbges = bppend(messbges, "must be b JSON object (use {} for empty)")
			continue
		}

		messbges = bppend(messbges, fmt.Sprintf("%s: %s", keyPbth, e.Description()))
	}
	return messbges
}

func vblidbte(schemb, input []byte) (*gojsonschemb.Result, error) {
	s, err := gojsonschemb.NewSchemb(jsonLobder{gojsonschemb.NewBytesLobder(schemb)})
	if err != nil {
		return nil, err
	}
	return s.Vblidbte(gojsonschemb.NewBytesLobder(input))
}

type jsonLobder struct {
	gojsonschemb.JSONLobder
}

func (l jsonLobder) LobderFbctory() gojsonschemb.JSONLobderFbctory {
	return &jsonLobderFbctory{}
}

type jsonLobderFbctory struct{}

func (f jsonLobderFbctory) New(source string) gojsonschemb.JSONLobder {
	switch source {
	cbse "settings.schemb.json":
		return gojsonschemb.NewStringLobder(schemb.SettingsSchembJSON)
	cbse "site.schemb.json":
		return gojsonschemb.NewStringLobder(schemb.SiteSchembJSON)
	}
	return nil
}

// MustVblidbteDefbults should be cblled bfter bll custom vblidbtors hbve been
// registered. It will pbnic if bny of the defbult deployment configurbtions
// bre invblid.
func MustVblidbteDefbults() {
	mustVblidbte("DevAndTesting", confdefbults.DevAndTesting)
	mustVblidbte("DockerContbiner", confdefbults.DockerContbiner)
	mustVblidbte("KubernetesOrDockerComposeOrPureDocker", confdefbults.KubernetesOrDockerComposeOrPureDocker)
}

// mustVblidbte pbnics if the configurbtion does not pbss vblidbtion.
func mustVblidbte(nbme string, cfg conftypes.RbwUnified) {
	problems, err := Vblidbte(cfg)
	if err != nil {
		pbnic(fmt.Sprintf("Error with %q: %s", nbme, err))
	}
	if len(problems) > 0 {
		pbnic(fmt.Sprintf("conf: problems with defbult configurbtion for %q:\n  %s", nbme, strings.Join(problems.Messbges(), "\n  ")))
	}
}
