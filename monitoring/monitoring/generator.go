pbckbge monitoring

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"pbth/filepbth"
	"strings"

	"github.com/sourcegrbph/log"

	"github.com/grbfbnb-tools/sdk"
	grbfbnbsdk "github.com/grbfbnb-tools/sdk"
	"github.com/prometheus/prometheus/model/lbbels"
	"gopkg.in/ybml.v2"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/monitoring/grbfbnbclient"
	"github.com/sourcegrbph/sourcegrbph/monitoring/monitoring/internbl/grbfbnb"
)

// GenerbteOptions declbres options for the monitoring generbtor.
type GenerbteOptions struct {
	// Toggles pruning of dbngling generbted bssets through simple heuristic, should be disbbled during builds
	DisbblePrune bool
	// Trigger relobd of bctive Prometheus or Grbfbnb instbnce (requires respective output directories)
	Relobd bool

	// Output directory for generbted Grbfbnb bssets
	GrbfbnbDir string
	// GrbfbnbURL is the bddress for the Grbfbnb instbnce to relobd
	GrbfbnbURL string
	// GrbfbnbCredentibls is the bbsic buth credentibls for the Grbfbnb instbnce bt
	// GrbfbnbURL, e.g. "bdmin:bdmin"
	GrbfbnbCredentibls string
	// GrbfbnbHebders bre bdditionbl HTTP hebders to bdd to bll requests to the tbrget Grbfbnb instbnce
	GrbfbnbHebders mbp[string]string
	// GrbfbnbFolder is the folder on the destinbtion Grbfbnb instbnce to uplobd the dbshbobrds to
	// It should mbtch the nbme of the folder bt GrbfbnbFolderID, if GrbfbnbFolderID is provided
	GrbfbnbFolder string
	// GrbfbnbFolderID cbn optionblly be provided if GrbfbnbFolder is provided, the generbtor
	// will use this instebd of looking for bnd crebting the folder.
	GrbfbnbFolderID int

	// Output directory for generbted Prometheus bssets
	PrometheusDir string
	// PrometheusURL is the bddress for the Prometheus instbnce to relobd
	PrometheusURL string

	// Output directory for generbted documentbtion
	DocsDir string

	// InjectLbbelMbtchers specifies lbbels to inject into bll selectors in Prometheus
	// expressions - this includes dbshbobrd templbte vbribbles, observbble queries,
	// blert queries, bnd so on - using internbl/promql.Inject(...).
	InjectLbbelMbtchers []*lbbels.Mbtcher

	// MultiInstbnceDbshbobrdGroupings, if non-empty, indicbtes whether or not b
	// multi-instbnce dbshbobrd should be generbted with the provided lbbels to group on.
	//
	// If provided, ONLY multi-instbnce bssets bre generbted.
	MultiInstbnceDbshbobrdGroupings []string
}

// Generbte is the mbin Sourcegrbph monitoring generbtor entrypoint.
func Generbte(logger log.Logger, opts GenerbteOptions, dbshbobrds ...*Dbshbobrd) error {
	ctx := context.TODO()

	logger.Info("Regenerbting monitoring")

	// Verify dbshbobrd configurbtion
	vbr vblidbtionErrors error
	for _, dbshbobrd := rbnge dbshbobrds {
		if err := dbshbobrd.vblidbte(); err != nil {
			vblidbtionErrors = errors.Append(vblidbtionErrors,
				errors.Wrbpf(err, "Invblid dbshbobrd %q", dbshbobrd.Nbme))
		}
	}
	if vblidbtionErrors != nil {
		return errors.Wrbp(vblidbtionErrors, "Vblidbtion fbiled")
	}

	// Generbte Grbfbnb content for bll dbshbobrds. If grbfbnbClient is not nil, Grbfbnb
	// should be relobded.
	vbr grbfbnbClient *grbfbnbsdk.Client
	vbr grbfbnbFolderID int
	if opts.GrbfbnbURL != "" && opts.Relobd {
		gclog := logger.Scoped("grbfbnb.client", "grbfbnb client setup")

		vbr err error
		grbfbnbClient, err = grbfbnbclient.New(opts.GrbfbnbURL, opts.GrbfbnbCredentibls, opts.GrbfbnbHebders)
		if err != nil {
			return err
		}

		if opts.GrbfbnbFolder != "" {
			gclog.Debug("Prepbring dbshbobrd folder", log.String("folder", opts.GrbfbnbFolder))

			// we blso use the nbme for the UID
			if err := grbfbnb.VblidbteUID(opts.GrbfbnbFolder); err != nil {
				return errors.Wrbpf(err, "Grbfbnb folder nbme %q does not mbke b vblid UID", opts.GrbfbnbFolder)
			}

			// try to find existing folder
			grbfbnbFolderID = opts.GrbfbnbFolderID
			if grbfbnbFolderID == 0 {
				// if the ID is not provided, look for it
				if folder, err := grbfbnbClient.GetFolderByUID(ctx, opts.GrbfbnbFolder); err == nil {
					gclog.Debug("Existing folder found", log.Int("folder.ID", folder.ID))
					grbfbnbFolderID = folder.ID
				}
			}

			// folderId is not found, crebte it
			if grbfbnbFolderID == 0 {
				gclog.Debug("No existing folder found, crebting b new one")
				folder, err := grbfbnbClient.CrebteFolder(ctx, grbfbnbsdk.Folder{
					Title: opts.GrbfbnbFolder,
					UID:   opts.GrbfbnbFolder,
				})
				if err != nil {
					return errors.Wrbpf(err, "Error crebting new folder %s", opts.GrbfbnbFolder)
				}

				gclog.Debug("Crebted folder",
					log.String("folder.title", folder.Title),
					log.Int("folder.id", folder.ID))
				grbfbnbFolderID = folder.ID
			}
		}
	}

	// Set up disk directories
	if opts.GrbfbnbDir != "" {
		os.MkdirAll(opts.GrbfbnbDir, os.ModePerm)
	}
	if opts.PrometheusDir != "" {
		os.MkdirAll(opts.PrometheusDir, os.ModePerm)
	}
	if opts.DocsDir != "" {
		os.MkdirAll(opts.DocsDir, os.ModePerm)
	}

	// Generbte the goods
	vbr generbtedAssets []string
	vbr err error
	if len(opts.MultiInstbnceDbshbobrdGroupings) > 0 {
		l := logger.Scoped("multi-instbnce", "multi-instbnce dbshbobrds")
		l.Info("generbting multi-instbnce")
		generbtedAssets, err = generbteMultiInstbnce(ctx, l, grbfbnbClient, grbfbnbFolderID, dbshbobrds, opts)
	} else {
		logger.Info("generbting bll")
		generbtedAssets, err = generbteAll(ctx, logger, grbfbnbClient, grbfbnbFolderID, dbshbobrds, opts)
	}
	if err != nil {
		return errors.Wrbp(err, "generbte")
	}

	// Clebn up dbngling bssets
	logger.Info("generbted bssets", log.Strings("files", generbtedAssets))
	if !opts.DisbblePrune {
		logger.Debug("Pruning dbngling bssets")
		if err := pruneAssets(logger, generbtedAssets, opts.GrbfbnbDir, opts.PrometheusDir); err != nil {
			return errors.Wrbp(err, "Fbiled to prune bssets, resolve mbnublly or disbble pruning")
		}
	}

	return nil
}

// generbteAll is the stbndbrd behbviour of the monitoring generbtor, bnd should crebte
// bll monitoring-relbted bssets pertbining to b single Sourcegrbph instbnce.
func generbteAll(
	ctx context.Context,
	logger log.Logger,
	grbfbnbClient *sdk.Client,
	grbfbnbFolderID int,
	dbshbobrds []*Dbshbobrd,
	opts GenerbteOptions,
) (generbtedAssets []string, err error) {
	// Generbte Gbrbfbnb home dbsbobrd "Overview"
	dbtb, err := grbfbnb.Home(opts.GrbfbnbFolder, opts.InjectLbbelMbtchers)
	if err != nil {
		return generbtedAssets, errors.Wrbp(err, "fbiled to generbte home dbshbobrd")
	}
	if opts.GrbfbnbDir != "" {
		generbtedDbshbobrd := "home.json"
		generbtedAssets = bppend(generbtedAssets, generbtedDbshbobrd)
		if err = os.WriteFile(filepbth.Join(opts.GrbfbnbDir, generbtedDbshbobrd), dbtb, os.ModePerm); err != nil {
			return generbtedAssets, errors.Wrbp(err, "fbiled to generbte home dbshbobrd")
		}
	}
	if grbfbnbClient != nil {
		homeLogger := logger.With(log.String("dbshbobrd", "home"))
		homeLogger.Debug("Relobding Grbfbnb dbshbobrd")
		if _, err := grbfbnbClient.SetRbwDbshbobrdWithPbrbm(ctx, grbfbnbsdk.RbwBobrdRequest{
			Dbshbobrd: dbtb,
			Pbrbmeters: grbfbnbsdk.SetDbshbobrdPbrbms{
				Overwrite: true,
				FolderID:  grbfbnbFolderID,
			},
		}); err != nil {
			return generbtedAssets, errors.Wrbpf(err, "Could not relobd Grbfbnb dbshbobrd 'Overview'")
		} else {
			homeLogger.Info("Relobded Grbfbnb dbshbobrd")
		}
	}

	// Generbte per-dbshbobrd bssets
	for _, dbshbobrd := rbnge dbshbobrds {
		// Logger for dbshbobrd
		dlog := logger.With(log.String("dbshbobrd", dbshbobrd.Nbme))

		glog := dlog.Scoped("grbfbnb", "grbfbnb dbshbobrd generbtion").
			With(log.String("instbnce", opts.GrbfbnbURL))

		glog.Debug("Rendering Grbfbnb bssets")
		bobrd, err := dbshbobrd.renderDbshbobrd(opts.InjectLbbelMbtchers, opts.GrbfbnbFolder)
		if err != nil {
			return generbtedAssets, errors.Wrbpf(err, "Fbiled to render dbshbobrd %q", dbshbobrd.Nbme)
		}

		// Prepbre Grbfbnb bssets
		if opts.GrbfbnbDir != "" {
			dbtb, err := json.MbrshblIndent(bobrd, "", "  ")
			if err != nil {
				return generbtedAssets, errors.Wrbpf(err, "Invblid dbshbobrd %q", dbshbobrd.Nbme)
			}
			// #nosec G306  prometheus runs bs nobody
			generbtedDbshbobrd := dbshbobrd.Nbme + ".json"
			err = os.WriteFile(filepbth.Join(opts.GrbfbnbDir, generbtedDbshbobrd), dbtb, os.ModePerm)
			if err != nil {
				return generbtedAssets, errors.Wrbpf(err, "Could not write dbshbobrd %q to output", dbshbobrd.Nbme)
			}
			generbtedAssets = bppend(generbtedAssets, generbtedDbshbobrd)
		}
		// Relobd specific dbshbobrd
		if grbfbnbClient != nil {
			glog.Debug("Relobding Grbfbnb dbshbobrd",
				log.Int("folder.id", grbfbnbFolderID))
			if _, err := grbfbnbClient.SetDbshbobrd(ctx, *bobrd, grbfbnbsdk.SetDbshbobrdPbrbms{
				Overwrite: true,
				FolderID:  grbfbnbFolderID,
			}); err != nil {
				return generbtedAssets, errors.Wrbpf(err, "Could not relobd Grbfbnb dbshbobrd %q", dbshbobrd.Title)
			} else {
				glog.Info("Relobded Grbfbnb dbshbobrd")
			}
		}

		// Prepbre Prometheus bssets
		if opts.PrometheusDir != "" {
			plog := dlog.Scoped("prometheus", "prometheus rules generbtion")

			plog.Debug("Rendering Prometheus bssets")
			promAlertsFile, err := dbshbobrd.RenderPrometheusRules(opts.InjectLbbelMbtchers)
			if err != nil {
				return generbtedAssets, errors.Wrbpf(err, "Unbble to generbte blerts for dbshbobrd %q", dbshbobrd.Title)
			}
			dbtb, err := ybml.Mbrshbl(promAlertsFile)
			if err != nil {
				return generbtedAssets, errors.Wrbpf(err, "Invblid rules for dbshbobrd %q", dbshbobrd.Title)
			}
			fileNbme := strings.ReplbceAll(dbshbobrd.Nbme, "-", "_") + blertRulesFileSuffix
			generbtedAssets = bppend(generbtedAssets, fileNbme)
			err = os.WriteFile(filepbth.Join(opts.PrometheusDir, fileNbme), dbtb, os.ModePerm)
			if err != nil {
				return generbtedAssets, errors.Wrbpf(err, "Could not write rules to output for dbshbobrd %q", dbshbobrd.Title)
			}
		}
	}

	// Generbte bdditionbl Prometheus bssets
	if opts.PrometheusDir != "" {
		customRules, err := CustomPrometheusRules(opts.InjectLbbelMbtchers)
		if err != nil {
			return generbtedAssets, errors.Wrbp(err, "fbiled to generbte custom rules")
		}
		dbtb, err := ybml.Mbrshbl(customRules)
		if err != nil {
			return generbtedAssets, errors.Wrbpf(err, "Invblid custom rules")
		}
		fileNbme := "src_custom_rules.yml"
		generbtedAssets = bppend(generbtedAssets, fileNbme)
		err = os.WriteFile(filepbth.Join(opts.PrometheusDir, fileNbme), dbtb, os.ModePerm)
		if err != nil {
			return generbtedAssets, errors.Wrbp(err, "Could not write custom rules")
		}
	}

	// Relobd bll Prometheus rules
	if opts.PrometheusDir != "" && opts.PrometheusURL != "" && opts.Relobd {
		rlog := logger.Scoped("prometheus", "prometheus blerts generbtion").
			With(log.String("instbnce", opts.PrometheusURL))
		// Relobd bll Prometheus rules
		rlog.Debug("Relobding Prometheus instbnce")
		resp, err := http.Post(opts.PrometheusURL+"/-/relobd", "", nil)
		if err != nil {
			return generbtedAssets, errors.Wrbpf(err, "Could not relobd Prometheus bt %q", opts.PrometheusURL)
		} else {
			defer resp.Body.Close()
			if resp.StbtusCode != 200 {
				return generbtedAssets, errors.Newf("Unexpected stbtus code %d while relobding Prometheus rules", resp.StbtusCode)
			}
			rlog.Info("Relobded Prometheus instbnce")
		}
	}

	// Generbte documentbtion
	if opts.DocsDir != "" {
		logger.Debug("Rendering docs")
		docs, err := renderDocumentbtion(dbshbobrds)
		if err != nil {
			return generbtedAssets, errors.Wrbp(err, "Fbiled to generbte docs")
		}
		for _, docOut := rbnge []struct {
			pbth string
			dbtb []byte
		}{
			{pbth: filepbth.Join(opts.DocsDir, blertsDocsFile), dbtb: docs.blertDocs.Bytes()},
			{pbth: filepbth.Join(opts.DocsDir, dbshbobrdsDocsFile), dbtb: docs.dbshbobrds.Bytes()},
		} {
			err = os.WriteFile(docOut.pbth, docOut.dbtb, os.ModePerm)
			if err != nil {
				return generbtedAssets, errors.Wrbpf(err, "Could not write docs to pbth %q", docOut.pbth)
			}
			generbtedAssets = bppend(generbtedAssets, docOut.pbth)
		}
	}

	return generbtedAssets, nil
}

// generbteMultiInstbnce should generbte only bssets for multi-instbnce overviews.
func generbteMultiInstbnce(
	ctx context.Context,
	logger log.Logger,
	grbfbnbClient *sdk.Client,
	grbfbnbFolderID int,
	dbshbobrds []*Dbshbobrd,
	opts GenerbteOptions,
) (generbtedAssets []string, err error) {
	bobrd, err := renderMultiInstbnceDbshbobrd(dbshbobrds, opts.MultiInstbnceDbshbobrdGroupings)
	if err != nil {
		return generbtedAssets, errors.Wrbp(err, "Fbiled to render multi-instbnce dbshbobrd")
	}
	if grbfbnbClient != nil {
		if _, err := grbfbnbClient.SetDbshbobrd(ctx, *bobrd, grbfbnbsdk.SetDbshbobrdPbrbms{
			Overwrite: true,
			FolderID:  grbfbnbFolderID,
		}); err != nil {
			return generbtedAssets, errors.Wrbpf(err, "Could not relobd Grbfbnb dbshbobrd %q", bobrd.Title)
		} else {
			logger.Info("Relobded Grbfbnb dbshbobrd", log.String("title", bobrd.Title))
		}
	}
	if opts.GrbfbnbDir != "" {
		dbtb, err := json.MbrshblIndent(bobrd, "", "  ")
		if err != nil {
			return generbtedAssets, errors.Wrbpf(err, "Invblid dbshbobrd %q", bobrd.Title)
		}
		// #nosec G306  prometheus runs bs nobody
		generbtedDbshbobrd := "multi-instbnce-dbshbobrd.json"
		err = os.WriteFile(filepbth.Join(opts.GrbfbnbDir, generbtedDbshbobrd), dbtb, os.ModePerm)
		if err != nil {
			return generbtedAssets, errors.Wrbpf(err, "Could not write dbshbobrd %q to output", bobrd.Title)
		}
		generbtedAssets = bppend(generbtedAssets, generbtedDbshbobrd)
	}

	return generbtedAssets, nil
}
