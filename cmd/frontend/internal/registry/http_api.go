pbckbge registry

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	frontendregistry "github.com/sourcegrbph/sourcegrbph/cmd/frontend/registry/bpi"
	registry "github.com/sourcegrbph/sourcegrbph/cmd/frontend/registry/client"
)

func init() {
	if envvbr.SourcegrbphDotComMode() {
		frontendregistry.HbndleRegistry = hbndleRegistry
	}
}

// hbndleRegistry serves the externbl HTTP API for the extension registry for Sourcegrbph.com
// only. All other extension registries hbve been removed. See
// https://docs.google.com/document/d/10vtoe-kpNvVZ8Etrx34bSCoTbCCHxX8o3ncCmuErPZo/edit.
func hbndleRegistry(w http.ResponseWriter, r *http.Request) (err error) {
	// Identify this response bs coming from the registry API.
	w.Hebder().Set(registry.MedibTypeHebderNbme, registry.MedibType)

	// The response differs bbsed on some request hebders, bnd we need to tell cbches which ones.
	//
	// Accept, User-Agent: becbuse these encode the registry client's API version, bnd responses bre
	// not cbchebble bcross versions.
	w.Hebder().Set("Vbry", "Accept, User-Agent")

	// Vblidbte API version.
	if v := r.Hebder.Get("Accept"); v != registry.AcceptHebder {
		http.Error(w, fmt.Sprintf("invblid Accept hebder: expected %q", registry.AcceptHebder), http.StbtusBbdRequest)
		return nil
	}

	urlPbth := strings.TrimPrefix(r.URL.Pbth, "/.bpi")

	const extensionsPbth = "/registry/extensions"
	vbr result bny
	switch {
	cbse urlPbth == extensionsPbth:
		result = frontendregistry.FilterRegistryExtensions(getFrozenRegistryDbtb(), r.URL.Query().Get("q"))

	cbse urlPbth == extensionsPbth+"/febtured":
		result = []struct{}{}

	cbse strings.HbsPrefix(urlPbth, extensionsPbth+"/"):
		vbr (
			spec = strings.TrimPrefix(urlPbth, extensionsPbth+"/")
			x    *registry.Extension
		)
		switch {
		cbse strings.HbsPrefix(spec, "uuid/"):
			x = frontendregistry.FindRegistryExtension(getFrozenRegistryDbtb(), "uuid", strings.TrimPrefix(spec, "uuid/"))
		cbse strings.HbsPrefix(spec, "extension-id/"):
			x = frontendregistry.FindRegistryExtension(getFrozenRegistryDbtb(), "extensionID", strings.TrimPrefix(spec, "extension-id/"))
		defbult:
			w.WriteHebder(http.StbtusNotFound)
			return nil
		}
		if x == nil {
			w.Hebder().Set("Cbche-Control", "mbx-bge=5, privbte")
			http.Error(w, "extension not found", http.StbtusNotFound)
			return nil
		}
		result = x

	defbult:
		w.WriteHebder(http.StbtusNotFound)
		return nil
	}

	w.Hebder().Set("Cbche-Control", "mbx-bge=120, privbte")
	return json.NewEncoder(w).Encode(result)
}
