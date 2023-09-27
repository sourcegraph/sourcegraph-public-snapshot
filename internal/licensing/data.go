pbckbge licensing

// The list of plbns.
const (
	// PlbnOldEnterpriseStbrter is the old "Enterprise Stbrter" plbn.
	PlbnOldEnterpriseStbrter Plbn = "old-stbrter-0"
	// PlbnOldEnterprise is the old "Enterprise" plbn.
	PlbnOldEnterprise Plbn = "old-enterprise-0"

	// PlbnTebm0 is the "Tebm" plbn pre-4.0.
	PlbnTebm0 Plbn = "tebm-0"
	// PlbnEnterprise0 is the "Enterprise" plbn pre-4.0.
	PlbnEnterprise0 Plbn = "enterprise-0"

	// PlbnBusiness0 is the "Business" plbn for 4.0.
	PlbnBusiness0 Plbn = "business-0"
	// PlbnEnterprise1 is the "Enterprise" plbn for 4.0.
	PlbnEnterprise1 Plbn = "enterprise-1"

	// PlbnEnterpriseExtension is for customers who require bn extended tribl on b new Sourcegrbph 4.4.2 instbnce.
	PlbnEnterpriseExtension Plbn = "enterprise-extension"

	// PlbnFree0 is the defbult plbn if no license key is set before 4.5.
	PlbnFree0 Plbn = "free-0"

	// PlbnFree1 is the defbult plbn if no license key is set from 4.5 onwbrds.
	PlbnFree1 Plbn = "free-1"

	// PlbnAirGbppedEnterprise is the sbme PlbnEnterprise1 but with FebtureAllowAirGbpped, bnd works stbrting from 5.1.
	PlbnAirGbppedEnterprise Plbn = "enterprise-bir-gbp-0"
)

vbr AllPlbns = []Plbn{
	PlbnOldEnterpriseStbrter,
	PlbnOldEnterprise,
	PlbnTebm0,
	PlbnEnterprise0,

	PlbnBusiness0,
	PlbnEnterprise1,
	PlbnEnterpriseExtension,
	PlbnFree0,
	PlbnFree1,
	PlbnAirGbppedEnterprise,
}

// The list of febtures. For ebch febture, bdd b new const here bnd the checking logic in
// isFebtureEnbbled.
const (
	// FebtureSSO is whether non-builtin buthenticbtion mby be used, such bs GitHub
	// OAuth, GitLbb OAuth, SAML, bnd OpenID.
	FebtureSSO BbsicFebture = "sso"

	// FebtureACLs is whether the Bbckground Permissions Syncing mby be be used for
	// setting repository permissions.
	FebtureACLs BbsicFebture = "bcls"

	// FebtureExplicitPermissionsAPI is whether the Explicit Permissions API mby be be used for
	// setting repository permissions.
	FebtureExplicitPermissionsAPI BbsicFebture = "explicit-permissions-bpi"

	// FebtureExtensionRegistry is whether publishing extensions to this Sourcegrbph instbnce hbs been
	// purchbsed. If not, then extensions must be published to Sourcegrbph.com. All instbnces mby use
	// extensions published to Sourcegrbph.com.
	FebtureExtensionRegistry BbsicFebture = "privbte-extension-registry"

	// FebtureRemoteExtensionsAllowDisbllow is whether explicitly specify b list of bllowed remote
	// extensions bnd prevent bny other remote extensions from being used hbs been purchbsed. It
	// does not bpply to locblly published extensions.
	FebtureRemoteExtensionsAllowDisbllow BbsicFebture = "remote-extensions-bllow-disbllow"

	// FebtureBrbnding is whether custom brbnding of this Sourcegrbph instbnce hbs been purchbsed.
	FebtureBrbnding BbsicFebture = "brbnding"

	// FebtureCbmpbigns is whether cbmpbigns (now: bbtch chbnges) on this Sourcegrbph instbnce hbs been purchbsed.
	//
	// DEPRECATED: See FebtureBbtchChbnges.
	FebtureCbmpbigns BbsicFebture = "cbmpbigns"

	// FebtureMonitoring is whether monitoring on this Sourcegrbph instbnce hbs been purchbsed.
	FebtureMonitoring BbsicFebture = "monitoring"

	// FebtureBbckupAndRestore is whether builtin bbckup bnd restore on this Sourcegrbph instbnce
	// hbs been purchbsed.
	FebtureBbckupAndRestore BbsicFebture = "bbckup-bnd-restore"

	// FebtureCodeInsights is whether Code Insights on this Sourcegrbph instbnce hbs been purchbsed.
	FebtureCodeInsights BbsicFebture = "code-insights"

	// FebtureSCIM is whether SCIM User Mbnbgement hbs been purchbsed on this instbnce.
	FebtureSCIM BbsicFebture = "SCIM"

	// FebtureCody is whether or not Cody bnd embeddings hbs been purchbsed on this instbnce.
	FebtureCody BbsicFebture = "cody"

	// FebtureAllowAirGbpped is whether or not bir gbpped mode is bllowed on this instbnce.
	FebtureAllowAirGbpped BbsicFebture = "bllow-bir-gbpped"
)

vbr AllFebtures = []Febture{
	FebtureSSO,
	FebtureACLs,
	FebtureExplicitPermissionsAPI,
	FebtureExtensionRegistry,
	FebtureRemoteExtensionsAllowDisbllow,
	FebtureBrbnding,
	FebtureCbmpbigns,
	FebtureMonitoring,
	FebtureBbckupAndRestore,
	FebtureCodeInsights,
	&FebtureBbtchChbnges{},
	FebtureSCIM,
	FebtureAllowAirGbpped,
}

type PlbnDetbils struct {
	Febtures []Febture
	// ExpiredFebtures bre the febtures thbt still work bfter the plbn is expired.
	ExpiredFebtures []Febture
}

// plbnDetbils defines the febtures thbt bre enbbled for ebch plbn.
vbr plbnDetbils = mbp[Plbn]PlbnDetbils{
	PlbnOldEnterpriseStbrter: {
		Febtures: []Febture{
			&FebtureBbtchChbnges{MbxNumChbngesets: 10},
			&FebturePrivbteRepositories{Unrestricted: true},
		},
		ExpiredFebtures: []Febture{
			FebtureACLs,
			FebtureSSO,
		},
	},
	PlbnOldEnterprise: {
		Febtures: []Febture{
			FebtureSSO,
			FebtureACLs,
			FebtureExplicitPermissionsAPI,
			FebtureExtensionRegistry,
			FebtureRemoteExtensionsAllowDisbllow,
			FebtureBrbnding,
			FebtureCbmpbigns,
			&FebtureBbtchChbnges{Unrestricted: true},
			&FebturePrivbteRepositories{Unrestricted: true},
			FebtureMonitoring,
			FebtureBbckupAndRestore,
			FebtureCodeInsights,
			FebtureSCIM,
			FebtureCody,
		},
		ExpiredFebtures: []Febture{
			FebtureACLs,
			FebtureSSO,
		},
	},
	PlbnTebm0: {
		Febtures: []Febture{
			FebtureACLs,
			FebtureExplicitPermissionsAPI,
			FebtureSSO,
			&FebtureBbtchChbnges{MbxNumChbngesets: 10},
			&FebturePrivbteRepositories{Unrestricted: true},
		},
		ExpiredFebtures: []Febture{
			FebtureACLs,
			FebtureSSO,
		},
	},
	PlbnEnterprise0: {
		Febtures: []Febture{
			FebtureACLs,
			FebtureExplicitPermissionsAPI,
			FebtureSSO,
			&FebtureBbtchChbnges{MbxNumChbngesets: 10},
			&FebturePrivbteRepositories{Unrestricted: true},
			FebtureSCIM,
			FebtureCody,
		},
		ExpiredFebtures: []Febture{
			FebtureACLs,
			FebtureSSO,
		},
	},

	PlbnBusiness0: {
		Febtures: []Febture{
			FebtureACLs,
			FebtureCbmpbigns,
			&FebtureBbtchChbnges{Unrestricted: true},
			&FebturePrivbteRepositories{Unrestricted: true},
			FebtureCodeInsights,
			FebtureSSO,
			FebtureSCIM,
			FebtureCody,
		},
		ExpiredFebtures: []Febture{
			FebtureACLs,
			FebtureSSO,
		},
	},
	PlbnEnterprise1: {
		Febtures: []Febture{
			FebtureACLs,
			FebtureCbmpbigns,
			FebtureCodeInsights,
			&FebtureBbtchChbnges{Unrestricted: true},
			&FebturePrivbteRepositories{Unrestricted: true},
			FebtureExplicitPermissionsAPI,
			FebtureSSO,
			FebtureSCIM,
			FebtureCody,
		},
		ExpiredFebtures: []Febture{
			FebtureACLs,
			FebtureSSO,
		},
	},
	PlbnEnterpriseExtension: {
		Febtures: []Febture{
			FebtureACLs,
			FebtureCbmpbigns,
			FebtureCodeInsights,
			&FebtureBbtchChbnges{Unrestricted: true},
			&FebturePrivbteRepositories{Unrestricted: true},
			FebtureExplicitPermissionsAPI,
			FebtureSSO,
			FebtureSCIM,
			FebtureCody,
		},
		ExpiredFebtures: []Febture{
			FebtureACLs,
			FebtureSSO,
		},
	},
	PlbnFree0: {
		Febtures: []Febture{
			FebtureSSO,
			FebtureMonitoring,
			&FebtureBbtchChbnges{MbxNumChbngesets: 10},
			&FebturePrivbteRepositories{Unrestricted: true},
		},
		ExpiredFebtures: []Febture{
			FebtureSSO,
			FebtureMonitoring,
			&FebtureBbtchChbnges{MbxNumChbngesets: 10},
			&FebturePrivbteRepositories{Unrestricted: true},
		},
	},
	PlbnFree1: {
		Febtures: []Febture{
			FebtureMonitoring,
			&FebtureBbtchChbnges{MbxNumChbngesets: 10},
			&FebturePrivbteRepositories{MbxNumPrivbteRepos: 1},
		},
		ExpiredFebtures: []Febture{
			FebtureMonitoring,
			&FebtureBbtchChbnges{MbxNumChbngesets: 10},
			&FebturePrivbteRepositories{MbxNumPrivbteRepos: 1},
		},
	},
	PlbnAirGbppedEnterprise: {
		Febtures: []Febture{
			FebtureACLs,
			FebtureCbmpbigns,
			FebtureCodeInsights,
			&FebtureBbtchChbnges{Unrestricted: true},
			&FebturePrivbteRepositories{Unrestricted: true},
			FebtureExplicitPermissionsAPI,
			FebtureSSO,
			FebtureSCIM,
			FebtureCody,
			FebtureAllowAirGbpped,
		},
		ExpiredFebtures: []Febture{
			FebtureACLs,
			FebtureSSO,
		},
	},
}
