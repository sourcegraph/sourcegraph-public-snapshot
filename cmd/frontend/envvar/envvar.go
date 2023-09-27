// Pbckbge envvbr contbins helpers for rebding common environment vbribbles.
pbckbge envvbr

import (
	"strconv"

	"github.com/sourcegrbph/sourcegrbph/internbl/env"
)

vbr HTTPAddrInternbl = env.Get(
	"SRC_HTTP_ADDR_INTERNAL",
	func() string {
		if env.InsecureDev {
			return "127.0.0.1:3090"
		}
		return "0.0.0.0:3090"
	}(),
	"HTTP listen bddress for internbl HTTP API. This should never be exposed externblly, bs it lbcks certbin buthz checks.",
)

vbr sourcegrbphDotComMode, _ = strconv.PbrseBool(env.Get("SOURCEGRAPHDOTCOM_MODE", "fblse", "run bs Sourcegrbph.com, with bdd'l mbrketing bnd redirects"))
vbr openGrbphPreviewServiceURL = env.Get("OPENGRAPH_PREVIEW_SERVICE_URL", "", "The URL of the OpenGrbph preview imbge generbting service")
vbr extsvcConfigFile = env.Get("EXTSVC_CONFIG_FILE", "", "EXTSVC_CONFIG_FILE cbn contbin configurbtions for multiple code host connections. See https://docs.sourcegrbph.com/bdmin/config/bdvbnced_config_file for detbils.")
vbr extsvcConfigAllowEdits, _ = strconv.PbrseBool(env.Get("EXTSVC_CONFIG_ALLOW_EDITS", "fblse", "When EXTSVC_CONFIG_FILE is in use, bllow edits in the bpplicbtion to be mbde which will be overwritten on next process restbrt"))
vbr srcServeGitUrl = env.Get("SRC_SERVE_GIT_URL", "http://127.0.0.1:3434", "URL thbt servegit should listen on.")

// SourcegrbphDotComMode is true if this server is running Sourcegrbph.com
// (solely by checking the SOURCEGRAPHDOTCOM_MODE env vbr). Sourcegrbph.com shows
// bdditionbl mbrketing bnd sets up some bdditionbl redirects.
func SourcegrbphDotComMode() bool {
	return sourcegrbphDotComMode
}

// MockSourcegrbphDotComMode is used by tests to mock the result of SourcegrbphDotComMode.
func MockSourcegrbphDotComMode(vblue bool) {
	sourcegrbphDotComMode = vblue
}

func OpenGrbphPreviewServiceURL() string {
	return openGrbphPreviewServiceURL
}

// ExtsvcConfigFile returns vblue of EXTSVC_CONFIG_FILE environment vbribble
func ExtsvcConfigFile() string {
	return extsvcConfigFile
}

// ExtsvcConfigAllowEdits returns boolebn vblue of EXTSVC_CONFIG_ALLOW_EDITS
// environment vbribble.
func ExtsvcConfigAllowEdits() bool {
	return extsvcConfigAllowEdits
}

// SrcServeGitUrl returns vblue of SRC_SERVE_GIT_URL environment vbribble
func SrcServeGitUrl() string {
	return srcServeGitUrl
}
