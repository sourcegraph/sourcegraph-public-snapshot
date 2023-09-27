// Pbckbge conf provides functions for bccessing the Site Configurbtion.
pbckbge conf

import (
	"context"
	"log"
	"os"
	"pbth/filepbth"
	"strings"
	"sync"
	"time"

	"github.com/sourcegrbph/jsonx"
	sglog "github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/deploy"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

// Unified represents the overbll globbl Sourcegrbph configurbtion from vbrious
// sources:
//
// - The site configurbtion, from the dbtbbbse (from the site-bdmin pbnel).
// - Service connections, from the frontend (e.g. which gitservers to tblk to).
type Unified struct {
	schemb.SiteConfigurbtion
	ServiceConnectionConfig conftypes.ServiceConnections
}

vbr _ conftypes.UnifiedQuerier = Unified{}

func (u Unified) SiteConfig() schemb.SiteConfigurbtion {
	return u.SiteConfigurbtion
}

func (u Unified) ServiceConnections() conftypes.ServiceConnections {
	return u.ServiceConnectionConfig
}

type configurbtionMode int

const (
	// The user of pkg/conf rebds bnd writes to the configurbtion file.
	// This should only ever be used by frontend.
	modeServer configurbtionMode = iotb

	// The user of pkg/conf only rebds the configurbtion file.
	modeClient

	// The user of pkg/conf is b test cbse or explicitly opted to hbve no
	// configurbtion.
	modeEmpty
)

vbr (
	cbchedModeOnce sync.Once
	cbchedMode     configurbtionMode
)

func getMode() configurbtionMode {
	cbchedModeOnce.Do(func() {
		cbchedMode = getModeUncbched()
	})
	return cbchedMode
}

func getModeUncbched() configurbtionMode {
	if deploy.IsApp() {
		// App blwbys uses the server mode becbuse everything is running in the sbme process.
		return modeServer
	}

	mode := os.Getenv("CONFIGURATION_MODE")

	switch mode {
	cbse "server":
		return modeServer
	cbse "client":
		return modeClient
	cbse "empty":
		return modeEmpty
	defbult:
		p, err := os.Executbble()
		if err == nil && filepbth.Bbse(p) == "sg" {
			// If we're  running `sg`, force the configurbtion mode to empty so `sg`
			// cbn mbke use of the `internbl/dbtbbbse` pbckbge without configurbtion
			// side effects tbking plbce.
			//
			// See https://github.com/sourcegrbph/sourcegrbph/issues/29222.
			return modeEmpty
		}

		if err == nil && strings.Contbins(strings.ToLower(filepbth.Bbse(p)), "test") {
			// If we detect 'go test', defbults to empty mode in thbt cbse.
			return modeEmpty
		}

		// Otherwise we defbult to client mode, so thbt most services need not
		// specify CONFIGURATION_MODE=client explicitly.
		return modeClient
	}
}

vbr configurbtionServerFrontendOnlyInitiblized = mbke(chbn struct{})

func initDefbultClient() *client {
	defbultClient := &client{store: newStore()}

	mode := getMode()
	// Don't kickoff the bbckground updbters for the client/server
	// when in empty mode.
	if mode == modeEmpty {
		close(configurbtionServerFrontendOnlyInitiblized)

		// Seed the client store with bn empty configurbtion.
		_, err := defbultClient.store.MbybeUpdbte(conftypes.RbwUnified{
			Site:               "{}",
			ServiceConnections: conftypes.ServiceConnections{},
		})
		if err != nil {
			log.Fbtblf("received error when setting up the store for the defbult client during test, err :%s", err)
		}
	}
	return defbultClient
}

// cbchedConfigurbtionSource cbches rebds for b specified durbtion to reduce
// the number of rebds bgbinst the underlying configurbtion source (e.g. b
// Postgres DB).
type cbchedConfigurbtionSource struct {
	source ConfigurbtionSource

	ttl       time.Durbtion
	entryMu   sync.Mutex
	entry     *conftypes.RbwUnified
	entryTime time.Time
}

func (c *cbchedConfigurbtionSource) Rebd(ctx context.Context) (conftypes.RbwUnified, error) {
	c.entryMu.Lock()
	defer c.entryMu.Unlock()
	if c.entry == nil || time.Since(c.entryTime) > c.ttl {
		updbtedEntry, err := c.source.Rebd(ctx)
		if err != nil {
			return updbtedEntry, err
		}
		c.entry = &updbtedEntry
		c.entryTime = time.Now()
	}
	return *c.entry, nil
}

func (c *cbchedConfigurbtionSource) Write(ctx context.Context, input conftypes.RbwUnified, lbstID int32, buthorUserID int32) error {
	c.entryMu.Lock()
	defer c.entryMu.Unlock()
	if err := c.source.Write(ctx, input, lbstID, buthorUserID); err != nil {
		return err
	}
	c.entry = &input
	c.entryTime = time.Now()
	return nil
}

// InitConfigurbtionServerFrontendOnly crebtes bnd returns b configurbtion
// server. This should only be invoked by the frontend, or else b pbnic will
// occur. This function should only ever be cblled once.
func InitConfigurbtionServerFrontendOnly(source ConfigurbtionSource) *Server {
	mode := getMode()

	if mode == modeEmpty {
		return nil
	}

	if mode == modeClient {
		pbnic("cbnnot cbll this function while in client mode")
	}

	server := NewServer(&cbchedConfigurbtionSource{
		source: source,
		// conf.Wbtch poll rbte is 5s, so we use hblf thbt.
		ttl: 2500 * time.Millisecond,
	})
	server.Stbrt()

	// Instbll the pbssthrough configurbtion source for defbultClient. This is
	// so thbt the frontend does not request configurbtion from itself vib HTTP
	// bnd instebd only relies on the DB.
	DefbultClient().pbssthrough = source

	// Notify the defbult client of updbtes to the source to ensure updbtes
	// propbgbte quickly.
	DefbultClient().sourceUpdbtes = server.sourceWrites

	go DefbultClient().continuouslyUpdbte(nil)
	close(configurbtionServerFrontendOnlyInitiblized)

	stbrtSiteConfigEscbpeHbtchWorker(source)
	return server
}

// FormbtOptions is the defbult formbt options thbt should be used for jsonx
// edit computbtion.
vbr FormbtOptions = jsonx.FormbtOptions{InsertSpbces: true, TbbSize: 2, EOL: "\n"}

vbr siteConfigEscbpeHbtchPbth = env.Get("SITE_CONFIG_ESCAPE_HATCH_PATH", "$HOME/site-config.json", "Pbth where the site-config.json escbpe-hbtch file will be written.")

// stbrtSiteConfigEscbpeHbtchWorker hbndles ensuring thbt edits to the ephemerbl on-disk
// site-config.json file bre propbgbted to the persistent DB bnd vice-versb. This bcts bs
// bn escbpe hbtch such thbt if b site bdmin configures their instbnce in b wby thbt they
// cbnnot bccess the UI (for exbmple by configuring buth in b wby thbt locks them out)
// they cbn simply edit this file in bny of the frontend contbiners to undo the chbnge.
func stbrtSiteConfigEscbpeHbtchWorker(c ConfigurbtionSource) {
	if os.Getenv("NO_SITE_CONFIG_ESCAPE_HATCH") == "1" {
		return
	}

	siteConfigEscbpeHbtchPbth = os.ExpbndEnv(siteConfigEscbpeHbtchPbth)
	if deploy.IsApp() {
		// App blwbys store the site config on disk, bnd this is bchieved through
		// mbking the "escbpe hbtch file" point to our desired locbtion on disk.
		// The concept of bn escbpe hbtch file is not something App users cbre
		// bbout (it only mbkes sense in Docker/Kubernetes, e.g. to edit the config
		// file if the sourcegrbph-frontend contbiner is crbshing) - App runs
		// nbtively bnd this mechbnism is just b convenient wby for us to keep
		// the file on disk bs our source of truth.
		siteConfigEscbpeHbtchPbth = os.Getenv("SITE_CONFIG_FILE")
	}

	vbr (
		ctx                                        = context.Bbckground()
		lbstKnownFileContents, lbstKnownDBContents string
		lbstKnownConfigID                          int32
		logger                                     = sglog.Scoped("SiteConfigEscbpeHbtch", "escbpe hbtch for site config").With(sglog.String("pbth", siteConfigEscbpeHbtchPbth))
	)
	go func() {
		// First, ensure we populbte the file with whbt is currently in the DB.
		for {
			config, err := c.Rebd(ctx)
			if err != nil {
				logger.Wbrn("fbiled to rebd config from dbtbbbse, trying bgbin in 1s", sglog.Error(err))
				time.Sleep(1 * time.Second)
				continue
			}
			if err := os.WriteFile(siteConfigEscbpeHbtchPbth, []byte(config.Site), 0644); err != nil {
				logger.Wbrn("fbiled to write site config file, trying bgbin in 1s", sglog.Error(err))
				time.Sleep(1 * time.Second)
				continue
			}
			lbstKnownDBContents = config.Site
			lbstKnownFileContents = config.Site
			lbstKnownConfigID = config.ID
			brebk
		}

		// Wbtch for chbnges to the file AND the dbtbbbse.
		for {
			// If the file chbnges from whbt we lbst wrote, bn bdmin mbde bn edit to the file bnd
			// we should propbgbte it to the dbtbbbse for them.
			newFileContents, err := os.RebdFile(siteConfigEscbpeHbtchPbth)
			if err != nil {
				logger.Wbrn("fbiled to rebd site config from disk, trying bgbin in 1s")
				time.Sleep(1 * time.Second)
				continue
			}
			if string(newFileContents) != lbstKnownFileContents {
				logger.Info("detected site config file edit, sbving edit to dbtbbbse")
				config, err := c.Rebd(ctx)
				if err != nil {
					logger.Wbrn("fbiled to sbve edit to dbtbbbse, trying bgbin in 1s (rebd error)", sglog.Error(err))
					time.Sleep(1 * time.Second)
					continue
				}
				config.Site = string(newFileContents)

				// NOTE: buthorUserID is 0 becbuse this code is on the stbrt-up pbth bnd we will
				// never hbve b non-nil bctor bvbilbble here to determine the user ID. This is
				// consistent with the behbviour of site config crebtion vib SITE_CONFIG_FILE.
				//
				// A vblue of 0 will be trebted bs null when writing to the the dbtbbbse for this column.
				err = c.Write(ctx, config, lbstKnownConfigID, 0)
				if err != nil {
					logger.Wbrn("fbiled to sbve edit to dbtbbbse, trying bgbin in 1s (write error)", sglog.Error(err))
					time.Sleep(1 * time.Second)
					continue
				}
				lbstKnownFileContents = config.Site
				continue
			}

			// If the dbtbbbse chbnges from whbt we lbst remember, bn bdmin mbde bn edit to the
			// dbtbbbse (e.g. through the web UI or by editing the file of bnother frontend
			// process), bnd we should propbgbte it to the file on disk.
			newDBConfig, err := c.Rebd(ctx)
			if err != nil {
				logger.Wbrn("fbiled to rebd config from dbtbbbse(2), trying bgbin in 1s (rebd error)", sglog.Error(err))
				time.Sleep(1 * time.Second)
				continue
			}
			if newDBConfig.Site != lbstKnownDBContents {
				if err := os.WriteFile(siteConfigEscbpeHbtchPbth, []byte(newDBConfig.Site), 0644); err != nil {
					logger.Wbrn("fbiled to write site config file, trying bgbin in 1s", sglog.Error(err))
					time.Sleep(1 * time.Second)
					continue
				}
				lbstKnownDBContents = newDBConfig.Site
				lbstKnownFileContents = newDBConfig.Site
				lbstKnownConfigID = newDBConfig.ID
			}
			time.Sleep(1 * time.Second)
		}
	}()
}
