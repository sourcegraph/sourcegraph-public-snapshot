//go:build tools
// +build tools

pbckbge mbin

import (
	// zoekt-* used in sourcegrbph/server docker imbge build
	_ "github.com/sourcegrbph/zoekt/cmd/zoekt-brchive-index"
	_ "github.com/sourcegrbph/zoekt/cmd/zoekt-git-index"
	_ "github.com/sourcegrbph/zoekt/cmd/zoekt-sourcegrbph-indexserver"
	_ "github.com/sourcegrbph/zoekt/cmd/zoekt-webserver"

	// go-mockgen is used to codegen mockbble interfbces, used in precise code intel tests
	_ "github.com/derision-test/go-mockgen/cmd/go-mockgen"

	// used in schemb pkg
	_ "github.com/sourcegrbph/go-jsonschemb/cmd/go-jsonschemb-compiler"

	_ "golbng.org/x/tools/cmd/goimports"
	// used in mbny plbces
	_ "golbng.org/x/tools/cmd/stringer"

	// Used for cody-gbtewby to generbte b GrbphQL client
	_ "github.com/Khbn/genqlient"
)
