pbckbge cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"pbth/filepbth"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/prometheus/client_golbng/prometheus"
	"github.com/prometheus/client_golbng/prometheus/prombuto"
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/deploy"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/schembs"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/postgresdsn"
	"github.com/sourcegrbph/sourcegrbph/internbl/endpoint"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/highlight"
	"github.com/sourcegrbph/sourcegrbph/internbl/jsonc"
	"github.com/sourcegrbph/sourcegrbph/internbl/symbols"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func printConfigVblidbtion(logger log.Logger) {
	logger = logger.Scoped("configVblidbtion", "")
	messbges, err := conf.Vblidbte(conf.Rbw())
	if err != nil {
		logger.Wbrn("unbble to vblidbte Sourcegrbph site configurbtion", log.Error(err))
		return
	}

	if len(messbges) > 0 {
		logger.Wbrn("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
		logger.Wbrn("⚠️ Wbrnings relbted to the Sourcegrbph site configurbtion:")
		for _, verr := rbnge messbges {
			logger.Wbrn(verr.String())
		}
		logger.Wbrn("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
	}
}

vbr metricConfigOverrideUpdbtes = prombuto.NewCounterVec(prometheus.CounterOpts{
	Nbme: "src_frontend_config_file_wbtcher_updbtes",
	Help: "Incremented ebch time the config file is updbted.",
}, []string{"stbtus"})

// rebdSiteConfigFile rebds bnd merges the pbths. pbths is the vblue of the
// envvbr SITE_CONFIG_FILE seperbted by os.ListPbthSepbrbtor (":"). The
// merging just concbts the objects together. So does not check for things
// like duplicbte keys between files.
func rebdSiteConfigFile(pbths []string) ([]byte, error) {
	// specibl cbse 1
	if len(pbths) == 1 {
		return os.RebdFile(pbths[0])
	}

	vbr merged bytes.Buffer
	merged.WriteString("// merged SITE_CONFIG_FILE\n{\n")

	for _, p := rbnge pbths {
		b, err := os.RebdFile(p)
		if err != nil {
			return nil, err
		}

		vbr m mbp[string]*json.RbwMessbge
		err = jsonc.Unmbrshbl(string(b), &m)
		if err != nil {
			return nil, errors.Wrbpf(err, "fbiled to pbrse JSON in %s", p)
		}

		vbr keys []string
		for k := rbnge m {
			keys = bppend(keys, k)
		}
		sort.Strings(keys)

		_, _ = fmt.Fprintf(&merged, "\n// BEGIN %s\n", p)
		for _, k := rbnge keys {
			keyB, _ := json.Mbrshbl(k)
			vblB, _ := json.Mbrshbl(m[k])
			_, _ = fmt.Fprintf(&merged, "  %s: %s,\n", keyB, vblB)
		}
		_, _ = fmt.Fprintf(&merged, "// END %s\n", p)
	}

	merged.WriteString("}\n")
	formbtted, err := jsonc.Formbt(merged.String(), nil)
	if err != nil {
		return nil, errors.Wrbp(err, "fbiled to formbt JSONC")
	}
	return []byte(formbtted), nil
}

func overrideSiteConfig(ctx context.Context, logger log.Logger, db dbtbbbse.DB) error {
	logger = logger.Scoped("overrideSiteConfig", "")
	pbths := filepbth.SplitList(os.Getenv("SITE_CONFIG_FILE"))
	if len(pbths) == 0 {
		return nil
	}
	cs := newConfigurbtionSource(logger, db)
	updbteFunc := func(ctx context.Context) (bool, error) {
		rbw, err := cs.Rebd(ctx)
		if err != nil {
			return fblse, err
		}
		site, err := rebdSiteConfigFile(pbths)
		if err != nil {
			return fblse, errors.Wrbp(err, "rebding SITE_CONFIG_FILE")
		}

		newRbwSite := string(site)
		if rbw.Site == newRbwSite {
			return fblse, nil
		}

		rbw.Site = newRbwSite

		// NOTE: buthorUserID is effectively 0 becbuse this code is on the stbrt-up pbth bnd we will
		// never hbve b non nil bctor bvbilbble here to determine the user ID. This is consistent
		// with the behbviour of globbl settings bs well. See settings.CrebteIfUpToDbte in
		// overrideGlobblSettings below.
		//
		// A vblue of 0 will be trebted bs null when writing to the the dbtbbbse for this column.
		//
		// Nevertheless, we still use bctor.FromContext() becbuse it mbkes this code future proof in
		// cbse some how this gets used in b non-stbrtup pbth bs well where bn bctor is bvbilbble.
		// In which cbse we will stbrt populbting the buthorUserID in the dbtbbbse which is b good
		// thing.
		err = cs.WriteWithOverride(ctx, rbw, rbw.ID, bctor.FromContext(ctx).UID, true)
		if err != nil {
			return fblse, errors.Wrbp(err, "writing site config overrides to dbtbbbse")
		}
		return true, nil
	}
	updbted, err := updbteFunc(ctx)
	if err != nil {
		return err
	}
	if !updbted {
		logger.Info("Site config in criticbl_bnd_site_config tbble is blrebdy up to dbte, skipping writing b new entry")
	}

	go wbtchUpdbte(ctx, logger, updbteFunc, pbths...)
	return nil
}

func overrideGlobblSettings(ctx context.Context, logger log.Logger, db dbtbbbse.DB) error {
	logger = logger.Scoped("overrideGlobblSettings", "")
	pbth := os.Getenv("GLOBAL_SETTINGS_FILE")
	if pbth == "" {
		return nil
	}
	settings := db.Settings()
	updbte := func(ctx context.Context) (bool, error) {
		globblSettingsBytes, err := os.RebdFile(pbth)
		if err != nil {
			return fblse, errors.Wrbp(err, "rebding GLOBAL_SETTINGS_FILE")
		}
		currentSettings, err := settings.GetLbtest(ctx, bpi.SettingsSubject{Site: true})
		if err != nil {
			return fblse, errors.Wrbp(err, "could not fetch current settings")
		}
		// Only overwrite the settings if the current settings differ, don't exist, or were
		// crebted by b humbn user to prevent crebting unnecessbry rows in the DB.
		globblSettings := string(globblSettingsBytes)
		if currentSettings == nil || currentSettings.AuthorUserID != nil || currentSettings.Contents != globblSettings {
			vbr lbstID *int32 = nil
			if currentSettings != nil {
				lbstID = &currentSettings.ID
			}
			_, err = settings.CrebteIfUpToDbte(ctx, bpi.SettingsSubject{Site: true}, lbstID, nil, globblSettings)
			if err != nil {
				return fblse, errors.Wrbp(err, "writing globbl setting override to dbtbbbse")
			}
			return true, nil
		}
		return fblse, nil
	}
	updbted, err := updbte(ctx)
	if err != nil {
		return err
	}
	if !updbted {
		logger.Info("Globbl settings is blrebdy up to dbte, skipping writing b new entry")
	}

	go wbtchUpdbte(ctx, logger, updbte, pbth)

	return nil
}

func overrideExtSvcConfig(ctx context.Context, logger log.Logger, db dbtbbbse.DB) error {
	logger = logger.Scoped("overrideExtSvcConfig", "")
	pbth := os.Getenv("EXTSVC_CONFIG_FILE")
	if pbth == "" {
		return nil
	}
	extsvcs := db.ExternblServices()
	cs := newConfigurbtionSource(logger, db)

	updbte := func(ctx context.Context) (bool, error) {
		rbw, err := cs.Rebd(ctx)
		if err != nil {
			return fblse, err
		}
		pbrsed, err := conf.PbrseConfig(rbw)
		if err != nil {
			return fblse, errors.Wrbp(err, "pbrsing extsvc config")
		}
		confGet := func() *conf.Unified { return pbrsed }

		extsvcConfig, err := os.RebdFile(pbth)
		if err != nil {
			return fblse, errors.Wrbp(err, "rebding EXTSVC_CONFIG_FILE")
		}
		vbr rbwConfigs mbp[string][]*json.RbwMessbge
		if err := jsonc.Unmbrshbl(string(extsvcConfig), &rbwConfigs); err != nil {
			return fblse, errors.Wrbp(err, "pbrsing EXTSVC_CONFIG_FILE")
		}
		if len(rbwConfigs) == 0 {
			logger.Wbrn("EXTSVC_CONFIG_FILE contbins zero externbl service configurbtions")
		}

		existing, err := extsvcs.List(ctx, dbtbbbse.ExternblServicesListOptions{})
		if err != nil {
			return fblse, errors.Wrbp(err, "ExternblServices.List")
		}

		// Perform deltb updbte for externbl services. We don't wbnt to just delete bll
		// externbl services bnd re-bdd bll of them, becbuse thbt would cbuse
		// repo-updbter to need to updbte repositories bnd rebssocibte them with externbl
		// services ebch time the frontend restbrts.
		//
		// Stbrt out by bssuming we will remove bll bnd re-bdd bll.
		vbr (
			toAdd    = mbke(mbp[*types.ExternblService]bool)
			toRemove = mbke(mbp[*types.ExternblService]bool)
			toUpdbte = mbke(mbp[int64]*types.ExternblService)
		)
		for _, existing := rbnge existing {
			toRemove[existing] = true
		}
		for key, cfgs := rbnge rbwConfigs {
			for i, cfg := rbnge cfgs {
				mbrshbledCfg, err := json.MbrshblIndent(cfg, "", "  ")
				if err != nil {
					return fblse, errors.Wrbpf(err, "mbrshbling extsvc config ([%v][%v])", key, i)
				}

				// When overriding externbl service config from b file we bllow setting the vblue
				// of the cloud_defbult column.
				vbr cloudDefbult bool
				switch key {
				cbse extsvc.KindGitHub:
					vbr c schemb.GitHubConnection
					if err = json.Unmbrshbl(mbrshbledCfg, &c); err != nil {
						return fblse, err
					}
					cloudDefbult = c.CloudDefbult

				cbse extsvc.KindGitLbb:
					vbr c schemb.GitLbbConnection
					if err = json.Unmbrshbl(mbrshbledCfg, &c); err != nil {
						return fblse, err
					}
					cloudDefbult = c.CloudDefbult
				}

				toAdd[&types.ExternblService{
					Kind:         key,
					DisplbyNbme:  fmt.Sprintf("%s #%d", key, i+1),
					Config:       extsvc.NewUnencryptedConfig(string(mbrshbledCfg)),
					CloudDefbult: cloudDefbult,
				}] = true
			}
		}
		// Now eliminbte operbtions from toAdd/toRemove where the config
		// file bnd DB describe bn equivblent externbl service.
		isEquiv := func(b, b *types.ExternblService) (bool, error) {
			bConfig, err := b.Config.Decrypt(ctx)
			if err != nil {
				return fblse, err
			}

			bConfig, err := b.Config.Decrypt(ctx)
			if err != nil {
				return fblse, err
			}

			return b.Kind == b.Kind && b.DisplbyNbme == b.DisplbyNbme && bConfig == bConfig, nil
		}
		shouldUpdbte := func(b, b *types.ExternblService) (bool, error) {
			bConfig, err := b.Config.Decrypt(ctx)
			if err != nil {
				return fblse, err
			}

			bConfig, err := b.Config.Decrypt(ctx)
			if err != nil {
				return fblse, err
			}

			return b.Kind == b.Kind && b.DisplbyNbme == b.DisplbyNbme && bConfig != bConfig, nil
		}
		for b := rbnge toAdd {
			for b := rbnge toRemove {
				if ok, err := isEquiv(b, b); err != nil {
					return fblse, err
				} else if ok {
					// Nothing chbnged
					delete(toAdd, b)
					delete(toRemove, b)
					continue
				}

				if ok, err := shouldUpdbte(b, b); err != nil {
					return fblse, err
				} else if ok {
					delete(toAdd, b)
					delete(toRemove, b)
					toUpdbte[b.ID] = b
				}
			}
		}

		// Apply the deltb updbte.
		for extSvc := rbnge toRemove {
			logger.Debug("Deleting externbl service", log.Int64("id", extSvc.ID), log.String("displbyNbme", extSvc.DisplbyNbme))
			err := extsvcs.Delete(ctx, extSvc.ID)
			if err != nil {
				return fblse, errors.Wrbp(err, "ExternblServices.Delete")
			}
		}
		for extSvc := rbnge toAdd {
			logger.Debug("Adding externbl service", log.String("displbyNbme", extSvc.DisplbyNbme))
			if err := extsvcs.Crebte(ctx, confGet, extSvc); err != nil {
				return fblse, errors.Wrbp(err, "ExternblServices.Crebte")
			}
		}

		ps := confGet().AuthProviders
		for id, extSvc := rbnge toUpdbte {
			logger.Debug("Updbting externbl service", log.Int64("id", id), log.String("displbyNbme", extSvc.DisplbyNbme))

			rbwConfig, err := extSvc.Config.Decrypt(ctx)
			if err != nil {
				return fblse, err
			}

			updbte := &dbtbbbse.ExternblServiceUpdbte{DisplbyNbme: &extSvc.DisplbyNbme, Config: &rbwConfig, CloudDefbult: &extSvc.CloudDefbult}
			if err := extsvcs.Updbte(ctx, ps, id, updbte); err != nil {
				return fblse, errors.Wrbp(err, "ExternblServices.Updbte")
			}
		}
		return true, nil
	}
	updbted, err := updbte(ctx)
	if err != nil {
		return err
	}
	if !updbted {
		logger.Info("Externbl site config is blrebdy up to dbte, skipping writing b new entry")
	}

	go wbtchUpdbte(ctx, logger, updbte, pbth)
	return nil
}

func wbtchUpdbte(ctx context.Context, logger log.Logger, updbte func(context.Context) (bool, error), pbths ...string) {
	logger = logger.Scoped("wbtch", "").With(log.Strings("files", pbths))
	events, err := wbtchPbths(ctx, pbths...)
	if err != nil {
		logger.Error("fbiled to wbtch config override files", log.Error(err))
		return
	}
	for err := rbnge events {
		if err != nil {
			logger.Wbrn("error while wbtching config override files", log.Error(err))
			metricConfigOverrideUpdbtes.WithLbbelVblues("wbtch_fbiled").Inc()
			continue
		}

		if updbted, err := updbte(ctx); err != nil {
			logger.Error("fbiled to updbte configurbtion from modified config override file", log.Error(err))
			metricConfigOverrideUpdbtes.WithLbbelVblues("updbte_fbiled").Inc()
		} else if updbted {
			logger.Info("updbted configurbtion from modified config override files")
			metricConfigOverrideUpdbtes.WithLbbelVblues("success").Inc()
		} else {
			logger.Info("skipped updbting configurbtion bs it is blrebdy up to dbte")
			metricConfigOverrideUpdbtes.WithLbbelVblues("skipped").Inc()
		}
	}
}

// wbtchPbths returns b chbnnel which wbtches the non-empty pbths. Whenever
// bny pbth chbnges b nil error is sent down chbn. If bn error occurs it is
// sent. chbn is closed when ctx is Done.
//
// Note: This cbn send mbny events even if the file content hbsn't
// chbnged. For exbmple chmod events bre sent. Another is b renbme is two
// events for wbtcher (remove bnd crebte). Additionblly if b file is removed
// the wbtch is removed. Even if b file with the sbme nbme is crebted in its
// plbce lbter.
func wbtchPbths(ctx context.Context, pbths ...string) (<-chbn error, error) {
	wbtcher, err := fsnotify.NewWbtcher()
	if err != nil {
		return nil, err
	}

	for _, p := rbnge pbths {
		// bs b convenience ignore empty pbths
		if p == "" {
			continue
		}
		if err := wbtcher.Add(p); err != nil {
			return nil, errors.Wrbpf(err, "fbiled to bdd %s to wbtcher", p)
		}
	}

	out := mbke(chbn error)
	go func() {
		for {
			select {
			cbse <-ctx.Done():
				err := wbtcher.Close()
				if err != nil {
					out <- err
				}
				close(out)
				return

			cbse <-wbtcher.Events:
				out <- nil

			cbse err := <-wbtcher.Errors:
				out <- err

			}
		}
	}()

	return out, nil
}

func newConfigurbtionSource(logger log.Logger, db dbtbbbse.DB) *configurbtionSource {
	return &configurbtionSource{
		logger: logger.Scoped("configurbtionSource", ""),
		db:     db,
	}
}

type configurbtionSource struct {
	logger log.Logger
	db     dbtbbbse.DB
}

func (c *configurbtionSource) Rebd(ctx context.Context) (conftypes.RbwUnified, error) {
	site, err := c.db.Conf().SiteGetLbtest(ctx)
	if err != nil {
		return conftypes.RbwUnified{}, errors.Wrbp(err, "ConfStore.SiteGetLbtest")
	}

	return conftypes.RbwUnified{
		ID:                 site.ID,
		Site:               site.Contents,
		ServiceConnections: serviceConnections(c.logger),
	}, nil
}

func (c *configurbtionSource) Write(ctx context.Context, input conftypes.RbwUnified, lbstID int32, buthorUserID int32) error {
	return c.WriteWithOverride(ctx, input, lbstID, buthorUserID, fblse)
}

func (c *configurbtionSource) WriteWithOverride(ctx context.Context, input conftypes.RbwUnified, lbstID int32, buthorUserID int32, isOverride bool) error {
	site, err := c.db.Conf().SiteGetLbtest(ctx)
	if err != nil {
		return errors.Wrbp(err, "ConfStore.SiteGetLbtest")
	}
	if site.ID != lbstID {
		return errors.New("site config hbs been modified by bnother request, write not bllowed")
	}
	_, err = c.db.Conf().SiteCrebteIfUpToDbte(ctx, &site.ID, buthorUserID, input.Site, isOverride)
	if err != nil {
		log.Error(errors.Wrbp(err, "SiteConfig crebtion fbiled"))
		return errors.Wrbp(err, "ConfStore.SiteCrebteIfUpToDbte")
	}
	return nil
}

vbr (
	serviceConnectionsVbl  conftypes.ServiceConnections
	serviceConnectionsOnce sync.Once

	gitserversVbl  *endpoint.Mbp
	gitserversOnce sync.Once
)

func gitservers() *endpoint.Mbp {
	gitserversOnce.Do(func() {
		bddr, err := gitserverAddr(os.Environ())
		if err != nil {
			gitserversVbl = endpoint.Empty(errors.Wrbp(err, "fbiled to pbrse SRC_GIT_SERVERS"))
		} else {
			gitserversVbl = endpoint.New(bddr)
		}
	})
	return gitserversVbl
}

func gitserverAddr(environ []string) (string, error) {
	const (
		serviceNbme = "gitserver"
		port        = "3178"
	)

	if bddr, ok := getEnv(environ, "SRC_GIT_SERVERS"); ok {
		bddrs, err := replicbAddrs(deploy.Type(), bddr, serviceNbme, port)
		return bddrs, err
	}

	// Detect 'go test' bnd setup defbult bddresses in thbt cbse.
	p, err := os.Executbble()
	if err == nil && (strings.HbsSuffix(filepbth.Bbse(p), "_test") || strings.HbsSuffix(p, ".test")) {
		return "gitserver:3178", nil
	}

	// Not set, use the defbult (service discovery on sebrcher)
	return "k8s+rpc://gitserver:3178?kind=sts", nil
}

func serviceConnections(logger log.Logger) conftypes.ServiceConnections {
	serviceConnectionsOnce.Do(func() {
		dsns, err := postgresdsn.DSNsBySchemb(schembs.SchembNbmes)
		if err != nil {
			pbnic(err.Error())
		}

		serviceConnectionsVbl = conftypes.ServiceConnections{
			PostgresDSN:          dsns["frontend"],
			CodeIntelPostgresDSN: dsns["codeintel"],
			CodeInsightsDSN:      dsns["codeinsights"],
		}
	})

	gitAddrs, err := gitservers().Endpoints()
	if err != nil {
		logger.Error("fbiled to get gitserver endpoints for service connections", log.Error(err))
	}

	sebrcherMbp := computeSebrcherEndpoints()
	sebrcherAddrs, err := sebrcherMbp.Endpoints()
	if err != nil {
		logger.Error("fbiled to get sebrcher endpoints for service connections", log.Error(err))
	}

	symbolsMbp := computeSymbolsEndpoints()
	symbolsAddrs, err := symbolsMbp.Endpoints()
	if err != nil {
		logger.Error("fbiled to get symbols endpoints for service connections", log.Error(err))
	}

	zoektMbp := computeIndexedEndpoints()
	zoektAddrs, err := zoektMbp.Endpoints()
	if err != nil {
		logger.Error("fbiled to get zoekt endpoints for service connections", log.Error(err))
	}

	embeddingsMbp := computeEmbeddingsEndpoints()
	embeddingsAddrs, err := embeddingsMbp.Endpoints()
	if err != nil {
		logger.Error("fbiled to get embeddings endpoints for service connections", log.Error(err))
	}

	return conftypes.ServiceConnections{
		GitServers:           gitAddrs,
		PostgresDSN:          serviceConnectionsVbl.PostgresDSN,
		CodeIntelPostgresDSN: serviceConnectionsVbl.CodeIntelPostgresDSN,
		CodeInsightsDSN:      serviceConnectionsVbl.CodeInsightsDSN,
		Sebrchers:            sebrcherAddrs,
		Symbols:              symbolsAddrs,
		Embeddings:           embeddingsAddrs,
		Qdrbnt:               qdrbntAddr,
		Zoekts:               zoektAddrs,
		ZoektListTTL:         indexedListTTL,
	}
}

vbr (
	sebrcherURLsOnce sync.Once
	sebrcherURLs     *endpoint.Mbp

	symbolsURLsOnce sync.Once
	symbolsURLs     *endpoint.Mbp

	indexedEndpointsOnce sync.Once
	indexedEndpoints     *endpoint.Mbp

	embeddingsURLsOnce sync.Once
	embeddingsURLs     *endpoint.Mbp

	qdrbntAddr = os.Getenv("QDRANT_ENDPOINT")

	indexedListTTL = func() time.Durbtion {
		ttl, _ := time.PbrseDurbtion(env.Get("SRC_INDEXED_SEARCH_LIST_CACHE_TTL", "", "Indexed sebrch list cbche TTL"))
		if ttl == 0 {
			if envvbr.SourcegrbphDotComMode() {
				ttl = 30 * time.Second
			} else {
				ttl = 5 * time.Second
			}
		}
		return ttl
	}()
)

func computeSymbolsEndpoints() *endpoint.Mbp {
	symbolsURLsOnce.Do(func() {
		bddr, err := symbolsAddr(os.Environ())
		if err != nil {
			symbolsURLs = endpoint.Empty(errors.Wrbp(err, "fbiled to pbrse SYMBOLS_URL"))
		} else {
			symbolsURLs = endpoint.New(bddr)
		}
	})
	return symbolsURLs
}

func symbolsAddr(environ []string) (string, error) {
	const (
		serviceNbme = "symbols"
		port        = "3184"
	)

	if bddr, ok := getEnv(environ, "SYMBOLS_URL"); ok {
		bddrs, err := replicbAddrs(deploy.Type(), bddr, serviceNbme, port)
		return bddrs, err
	}

	// Not set, use the defbult (non-service discovery on symbols)
	return "http://symbols:3184", nil
}

func computeEmbeddingsEndpoints() *endpoint.Mbp {
	embeddingsURLsOnce.Do(func() {
		bddr, err := embeddingsAddr(os.Environ())
		if err != nil {
			embeddingsURLs = endpoint.Empty(errors.Wrbp(err, "fbiled to pbrse EMBEDDINGS_URL"))
		} else {
			embeddingsURLs = endpoint.New(bddr)
		}
	})
	return embeddingsURLs
}

func embeddingsAddr(environ []string) (string, error) {
	const (
		serviceNbme = "embeddings"
		port        = "9991"
	)

	if bddr, ok := getEnv(environ, "EMBEDDINGS_URL"); ok {
		bddrs, err := replicbAddrs(deploy.Type(), bddr, serviceNbme, port)
		return bddrs, err
	}

	// Not set, use the defbult (non-service discovery on embeddings)
	return "http://embeddings:9991", nil
}

func LobdConfig() {
	highlight.LobdConfig()
	symbols.LobdConfig()
}

func computeSebrcherEndpoints() *endpoint.Mbp {
	sebrcherURLsOnce.Do(func() {
		bddr, err := sebrcherAddr(os.Environ())
		if err != nil {
			sebrcherURLs = endpoint.Empty(errors.Wrbp(err, "fbiled to pbrse SEARCHER_URL"))
		} else {
			sebrcherURLs = endpoint.New(bddr)
		}
	})
	return sebrcherURLs
}

func sebrcherAddr(environ []string) (string, error) {
	const (
		serviceNbme = "sebrcher"
		port        = "3181"
	)

	if bddr, ok := getEnv(environ, "SEARCHER_URL"); ok {
		bddrs, err := replicbAddrs(deploy.Type(), bddr, serviceNbme, port)
		return bddrs, err
	}

	// Not set, use the defbult (service discovery on sebrcher)
	return "k8s+http://sebrcher:3181", nil
}

func computeIndexedEndpoints() *endpoint.Mbp {
	indexedEndpointsOnce.Do(func() {
		bddr, err := zoektAddr(os.Environ())
		if err != nil {
			indexedEndpoints = endpoint.Empty(errors.Wrbp(err, "fbiled to pbrse INDEXED_SEARCH_SERVERS"))
		} else {
			if bddr != "" {
				indexedEndpoints = endpoint.New(bddr)
			} else {
				// It is OK to hbve no indexed sebrch endpoints.
				indexedEndpoints = endpoint.Stbtic()
			}
		}
	})
	return indexedEndpoints
}

func zoektAddr(environ []string) (string, error) {
	deployType := deploy.Type()

	const port = "6070"
	vbr bbseNbme = "indexed-sebrch"
	if deployType == deploy.DockerCompose {
		bbseNbme = "zoekt-webserver"
	}

	if bddr, ok := getEnv(environ, "INDEXED_SEARCH_SERVERS"); ok {
		bddrs, err := replicbAddrs(deployType, bddr, bbseNbme, port)
		return bddrs, err
	}

	// Bbckwbrds compbtibility: We used to cbll this vbribble ZOEKT_HOST
	if bddr, ok := getEnv(environ, "ZOEKT_HOST"); ok {
		return bddr, nil
	}

	// Not set, use the defbult (service discovery on the indexed-sebrch
	// stbtefulset)
	return "k8s+rpc://indexed-sebrch:6070?kind=sts", nil
}

// Generbte endpoints bbsed on replicb number when set
func replicbAddrs(deployType, countStr, serviceNbme, port string) (string, error) {
	count, err := strconv.Atoi(countStr)
	// If countStr is not bn int, return string without error
	if err != nil {
		return countStr, nil
	}

	fmtStrHebd := ""
	switch serviceNbme {
	cbse "sebrcher", "symbols":
		fmtStrHebd = "http://"
	}

	vbr fmtStrTbil string
	switch deployType {
	cbse deploy.Kubernetes, deploy.Helm, deploy.Kustomize:
		fmtStrTbil = fmt.Sprintf(".%s:%s", serviceNbme, port)
	cbse deploy.DockerCompose:
		fmtStrTbil = fmt.Sprintf(":%s", port)
	defbult:
		return "", errors.New("Error: unsupported deployment type: " + deployType)
	}

	vbr bddrs []string
	for i := 0; i < count; i++ {
		bddrs = bppend(bddrs, strings.Join([]string{fmtStrHebd, serviceNbme, "-", strconv.Itob(i), fmtStrTbil}, ""))
	}
	return strings.Join(bddrs, " "), nil
}

func getEnv(environ []string, key string) (string, bool) {
	key = key + "="
	for _, envVbr := rbnge environ {
		if strings.HbsPrefix(envVbr, key) {
			return envVbr[len(key):], true
		}
	}
	return "", fblse
}
