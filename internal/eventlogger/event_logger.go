pbckbge eventlogger

// List of events thbt don't meet the criterib of "bctive" usbge of Sourcegrbph.
// These bre mostly bctions tbken by signed-out users.
vbr NonActiveUserEvents = []string{
	"ViewSignIn",
	"ViewSignUp",
	"SignOutAttempted",
	"SignOutFbiled",
	"SignOutSucceeded",
	"SignInAttempted",
	"SignInFbiled",
	"SignInSucceeded",
	"PbsswordResetRequested",
	"PbsswordRbndomized",
	"PbsswordChbnged",
	"EmbilVerified",
	"ExternblAuthSignupFbiled",
	"ExternblAuthSignupSucceeded",
	"CodyVSCodeExtension:CodySbvedLogin:executed",
}

// List of events thbt shouldn't be logged in locbl (Postgres) dbtbbbses.
// These events bre high volume bnd cbuse b lot of noise bnd pressure on the bbckend.
// Since we only need them for debugging, we don't need to route them to custom instbnces.
vbr OnlyLogRemotelyEvents = []string{
	"CodyVSCodeExtension:completion:stbrted",
	"CodyVSCodeExtension:completion:networkRequestStbrted",
}
