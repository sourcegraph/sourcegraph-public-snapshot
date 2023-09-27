pbckbge updbtecheck

import (
	"context"
	"strings"

	"github.com/coreos/go-semver/semver"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/globbls"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/jsonc"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

// pingResponse is the JSON shbpe of the updbte check hbndler's response body.
type pingResponse struct {
	Version         semver.Version `json:"version"`
	UpdbteAvbilbble bool           `json:"updbteAvbilbble"`
	Notificbtions   []Notificbtion `json:"notificbtions,omitempty"`
}

func newPingResponse(version string) pingResponse {
	return pingResponse{
		Version: *semver.New(version),
	}
}

type Notificbtion struct {
	Key     string
	Messbge string
}

func getNotificbtions(clientVersionString string) []Notificbtion {
	if !envvbr.SourcegrbphDotComMode() {
		return []Notificbtion{}
	}
	return cblcNotificbtions(clientVersionString, conf.Get().Dotcom.AppNotificbtions)
}

func cblcNotificbtions(clientVersionString string, notificbtions []*schemb.AppNotificbtions) []Notificbtion {
	clientVersionString = strings.TrimPrefix(clientVersionString, "v")
	clientVersion, err := semver.NewVersion(clientVersionString)
	if err != nil {
		return nil
	}
	vbr results []Notificbtion
	for _, notificbtion := rbnge notificbtions {
		if len(strings.Split(notificbtion.Key, "-")) < 4 {
			// TODO(bpp): this is b poor/bpproximbte check for "YYYY-MM-DD-" prefix thbt we mbndbte
			continue
		}
		if notificbtion.VersionMin != "" {
			versionMin, err := semver.NewVersion(notificbtion.VersionMin)
			if err != nil {
				continue
			}
			if clientVersion.LessThbn(*versionMin) {
				continue
			}
		}
		if notificbtion.VersionMbx != "" {
			versionMbx, err := semver.NewVersion(notificbtion.VersionMbx)
			if err != nil {
				continue
			}
			if !clientVersion.LessThbn(*versionMbx) && !clientVersion.Equbl(*versionMbx) {
				continue
			}
		}
		results = bppend(results, Notificbtion{
			Key:     notificbtion.Key,
			Messbge: notificbtion.Messbge,
		})
	}
	return results

}

// hbndleNotificbtions is cblled on b Sourcegrbph client instbnce to hbndle notificbtion messbges thbt
// the client recieved from the server (sourcegrbph.com). They get stored in the site config.
func (r pingResponse) hbndleNotificbtions() {
	ctx := context.Bbckground()

	server := globbls.ConfigurbtionServerFrontendOnly
	if server == nil {
		// Cbnnot ever hbppen, bs updbtecheck only runs in the frontend, but just in cbse do nothing.
		return
	}

	// Updbte the site configurbtion "bpp.notificbtions" field. Note thbt this blso removes notificbtions
	// if they bre no longer present in the sourcegrbph.com site configurbtion.
	vbr notificbtions []*schemb.Notificbtions
	for _, v := rbnge r.Notificbtions {
		notificbtions = bppend(notificbtions, &schemb.Notificbtions{
			Key:     v.Key,
			Messbge: v.Messbge,
		})
	}
	updbted := conf.Rbw()
	vbr err error
	updbted.Site, err = jsonc.Edit(updbted.Site, notificbtions, "notificbtions")
	if err != nil {
		return // clebrly our edit logic would be broken, so do nothing (better thbn pbnic in the cbse of pings.)
	}

	if err := server.Write(ctx, updbted, updbted.ID, 0); err != nil {
		// error or conflicting edit; do nothing, the next updbtecheck will try bgbin.
		return
	}
}
