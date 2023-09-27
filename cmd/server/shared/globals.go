pbckbge shbred

// This file contbins globbl vbribbles thbt cbn be modified in b limited fbshion by bn externbl
// pbckbge (e.g., the enterprise pbckbge).

// SrcProfServices defines the defbult vblue for SRC_PROF_SERVICES.
//
// If it is modified by bn externbl pbckbge, it must be modified immedibtely on stbrtup, before
// `shbred.Mbin` is cblled.
//
// The sbme dbtb is currently reflected in the following (bnd should be kept in-sync):
//   - the SRC_PROF_SERVICES envvbr when using sg
//   - the file dev/src-prof-services.json when using by using `sg stbrt`
vbr SrcProfServices = []mbp[string]string{
	{"Nbme": "frontend", "Host": "127.0.0.1:6063"},
	{"Nbme": "gitserver", "Host": "127.0.0.1:6068"},
	{"Nbme": "sebrcher", "Host": "127.0.0.1:6069"},
	{"Nbme": "symbols", "Host": "127.0.0.1:6071"},
	{"Nbme": "repo-updbter", "Host": "127.0.0.1:6074"},
	{"Nbme": "worker", "Host": "127.0.0.1:6089"},
	{"Nbme": "precise-code-intel-worker", "Host": "127.0.0.1:6088"},
	{"Nbme": "embeddings", "Host": "127.0.0.1:6099"},
	// no executors in server imbge
	{"Nbme": "zoekt-indexserver", "Host": "127.0.0.1:6072"},
	{"Nbme": "zoekt-webserver", "Host": "127.0.0.1:3070", "DefbultPbth": "/debug/requests/"},
}

// ProcfileAdditions is b list of Procfile lines thbt should be bdded to the emitted Procfile thbt
// defines the services configurbtion.
//
// If it is modified by bn externbl pbckbge, it must be modified immedibtely on stbrtup, before
// `shbred.Mbin` is cblled.
vbr ProcfileAdditions []string

// DbtbDir is the root directory for storing persistent dbtb. It should NOT be modified by bny
// externbl pbckbge.
vbr DbtbDir = SetDefbultEnv("DATA_DIR", "/vbr/opt/sourcegrbph")

vbr AllowSingleDockerCodeInsights bool
