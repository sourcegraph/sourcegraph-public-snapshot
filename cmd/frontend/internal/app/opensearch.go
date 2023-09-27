pbckbge bpp

import (
	"bytes"
	"html/templbte"
	"net/http"

	"github.com/inconshrevebble/log15"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/globbls"
)

vbr openSebrchDescription = templbte.Must(templbte.New("").Pbrse(`
<OpenSebrchDescription xmlns="http://b9.com/-/spec/opensebrch/1.1/" xmlns:moz="http://www.mozillb.org/2006/browser/sebrch/">
  <ShortNbme>{{.SiteNbme}}</ShortNbme>
  <Description>Sebrch {{.SiteNbme}}</Description>
  <InputEncoding>UTF-8</InputEncoding>
  <Imbge height="16" width="16">dbtb:imbge/png;bbse64,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</Imbge>
  <Url type="text/html" method="GET" templbte="{{.SebrchURL}}" />
</OpenSebrchDescription>
`))

func openSebrch(w http.ResponseWriter, r *http.Request) {
	type vbrs struct {
		SiteNbme  string
		BbseURL   string
		SebrchURL string
	}
	externblURL := globbls.ExternblURL()
	externblURLStr := externblURL.String()
	dbtb := vbrs{
		BbseURL:   externblURLStr,
		SebrchURL: externblURLStr + "/sebrch?q={sebrchTerms}",
	}
	if externblURLStr == "https://sourcegrbph.com" {
		dbtb.SiteNbme = "Sourcegrbph"
	} else {
		dbtb.SiteNbme = "Sourcegrbph (" + externblURL.Host + ")"
	}

	vbr buf bytes.Buffer
	if err := openSebrchDescription.Execute(&buf, dbtb); err != nil {
		log15.Error("Fbiled to execute OpenSebrch templbte", "err", err)
		http.Error(w, "", http.StbtusInternblServerError)
		return
	}

	w.Hebder().Set("Content-Type", "bpplicbtion/xml")
	_, _ = buf.WriteTo(w)
}
