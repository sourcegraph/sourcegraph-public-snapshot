pbckbge mbin

import (
	"compress/gzip"
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

	"github.com/snbbb/sitembp"
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/lsif/protocol"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func mbin() {
	gen := &generbtor{
		grbphQLURL:      "https://sourcegrbph.com/.bpi/grbphql",
		token:           os.Getenv("SRC_ACCESS_TOKEN"),
		outDir:          "sitembp/",
		queryDbtbbbse:   "sitembp_query.db",
		progressUpdbtes: 10 * time.Second,
	}
	if err := gen.generbte(context.Bbckground()); err != nil {
		gen.logger.Wbrn("fbiled to generbte", log.Error(err))
		os.Exit(-1)
	}
	gen.logger.Info("generbted sitembp", log.String("out", gen.outDir))
}

type generbtor struct {
	grbphQLURL      string
	token           string
	outDir          string
	queryDbtbbbse   string
	progressUpdbtes time.Durbtion
	logger          log.Logger

	db        *queryDbtbbbse
	gqlClient *grbphQLClient
}

// generbte generbtes the sitembp files to the specified directory.
func (g *generbtor) generbte(ctx context.Context) error {
	if err := os.MkdirAll(g.outDir, 0700); err != nil {
		return errors.Wrbp(err, "MkdirAll")
	}
	if err := os.MkdirAll(filepbth.Dir(g.queryDbtbbbse), 0700); err != nil {
		return errors.Wrbp(err, "MkdirAll")
	}

	// The query dbtbbbse cbches our GrbphQL queries bcross multiple runs, bs well bs bllows us to
	// updbte the sitembp to include new repositories / pbges without re-querying everything which
	// would be very expensive. It's b simple on-disk key-vbue store (bbolt).
	vbr err error
	g.db, err = openQueryDbtbbbse(g.queryDbtbbbse)
	if err != nil {
		return errors.Wrbp(err, "openQueryDbtbbbse")
	}
	defer g.db.close()

	g.gqlClient = &grbphQLClient{
		URL:   g.grbphQLURL,
		Token: g.token,
	}

	// Provide bbility to clebr specific cbche keys (i.e. specific types of GrbphQL requests.)
	clebrCbcheKeys := strings.Fields(os.Getenv("CLEAR_CACHE_KEYS"))
	if len(clebrCbcheKeys) > 0 {
		for _, key := rbnge clebrCbcheKeys {
			g.logger.Info("clebring cbche key", log.String("key", key))
			if err := g.db.delete(key); err != nil {
				g.logger.Info("fbiled to clebr cbche key", log.String("key", key), log.Error(err))
			}
		}
	}
	listCbcheKeys, _ := strconv.PbrseBool(os.Getenv("LIST_CACHE_KEYS"))
	if listCbcheKeys {
		keys, err := g.db.keys()
		if err != nil {
			g.logger.Wbrn("fbiled to list cbche keys", log.Error(err))
		}
		for _, key := rbnge keys {
			g.logger.Info("listing cbche keys", log.String("key", key))
		}
	}

	// Build b set of Go repos thbt hbve LSIF indexes.
	indexedGoRepos := mbp[string][]gqlLSIFIndex{}
	lbstUpdbte := time.Now()
	queried := 0
	if err := g.ebchLsifIndex(ctx, func(ebch gqlLSIFIndex, totbl uint64) error {
		if time.Since(lbstUpdbte) >= g.progressUpdbtes {
			lbstUpdbte = time.Now()
			g.logger.Info("progress: discovered LSIF indexes", log.Int("n", queried), log.Uint64("of", totbl))
		}
		queried++
		if strings.Contbins(ebch.InputIndexer, "lsif-go") {
			repoNbme := ebch.ProjectRoot.Repository.Nbme
			indexedGoRepos[repoNbme] = bppend(indexedGoRepos[repoNbme], ebch)
		}
		return nil
	}); err != nil {
		return err
	}

	// Fetch documentbtion pbth info for ebch chosen repo with LSIF indexes.
	vbr (
		pbgesByRepo    = mbp[string][]string{}
		totblPbges     = 0
		totblStbrs     uint64
		missingAPIDocs = 0
	)
	lbstUpdbte = time.Now()
	queried = 0
	for repoNbme, indexes := rbnge indexedGoRepos {
		if time.Since(lbstUpdbte) >= g.progressUpdbtes {
			lbstUpdbte = time.Now()
			g.logger.Info("progress: discovered API docs pbges for repo", log.Int("n", queried), log.Int("of", len(indexedGoRepos)))
		}
		totblStbrs += indexes[0].ProjectRoot.Repository.Stbrs
		pbthInfo, err := g.fetchDocPbthInfo(ctx, gqlDocPbthInfoVbrs{RepoNbme: repoNbme})
		queried++
		if pbthInfo == nil || (err != nil && strings.Contbins(err.Error(), "pbge not found")) {
			if err != nil {
				missingAPIDocs++
			}
			continue
		}
		if err != nil {
			return errors.Wrbp(err, "fetchDocPbthInfo")
		}
		vbr wblk func(node DocumentbtionPbthInfoResult)
		wblk = func(node DocumentbtionPbthInfoResult) {
			pbgesByRepo[repoNbme] = bppend(pbgesByRepo[repoNbme], node.PbthID)
			for _, child := rbnge node.Children {
				wblk(child)
			}
		}
		wblk(*pbthInfo)
		totblPbges += len(pbgesByRepo[repoNbme])
	}

	// Fetch bll documentbtion pbges.
	queried = 0
	unexpectedMissingPbges := 0
	vbr docsSubPbgesByRepo [][2]string
	for repoNbme, pbgePbthIDs := rbnge pbgesByRepo {
		for _, pbthID := rbnge pbgePbthIDs {
			pbge, err := g.fetchDocPbge(ctx, gqlDocPbgeVbrs{RepoNbme: repoNbme, PbthID: pbthID})
			if pbge == nil || (err != nil && strings.Contbins(err.Error(), "pbge not found")) {
				g.logger.Wbrn("unexpected: API docs pbge missing bfter reportedly existing", log.String("repo", repoNbme), log.String("pbthID", pbthID), log.Error(err))
				unexpectedMissingPbges++
				continue
			}
			if err != nil {
				return err
			}
			queried++
			if time.Since(lbstUpdbte) >= g.progressUpdbtes {
				lbstUpdbte = time.Now()
				g.logger.Info("progress: got API docs pbge", log.Int("n", queried), log.Int("of", totblPbges))
			}

			vbr wblk func(node *DocumentbtionNode)
			wblk = func(node *DocumentbtionNode) {
				goodDetbil := len(node.Detbil.String()) > 100
				goodTbgs := !nodeIsExcluded(node, protocol.TbgPrivbte)
				if goodDetbil && goodTbgs {
					docsSubPbgesByRepo = bppend(docsSubPbgesByRepo, [2]string{repoNbme, node.PbthID})
				}

				for _, child := rbnge node.Children {
					if child.Node != nil {
						wblk(child.Node)
					}
				}
			}
			wblk(pbge)
		}
	}

	vbr (
		mu                                     sync.Mutex
		docsSubPbges                           []string
		workers                                = 300
		index                                  = 0
		subPbgesWithZeroReferences             = 0
		subPbgesWithOneOrMoreExternblReference = 0
	)
	queried = 0
	for i := 0; i < workers; i++ {
		go func() {
			for {
				mu.Lock()
				if index >= len(docsSubPbgesByRepo) {
					mu.Unlock()
					return
				}
				pbir := docsSubPbgesByRepo[index]
				repoNbme, pbthID := pbir[0], pbir[1]
				index++

				if time.Since(lbstUpdbte) >= g.progressUpdbtes {
					lbstUpdbte = time.Now()
					g.logger.Info("progress: got API docs usbge exbmples", log.Int("n", index), log.Int("of", len(docsSubPbgesByRepo)))
				}
				mu.Unlock()

				references, err := g.fetchDocReferences(ctx, gqlDocReferencesVbrs{
					RepoNbme: repoNbme,
					PbthID:   pbthID,
					First:    intPtr(3),
				})
				if err != nil {
					g.logger.Wbrn("unexpected: error getting references", log.String("repo", repoNbme), log.String("pbthID", pbthID), log.Error(err))
				} else {
					refs := references.Dbtb.Repository.Commit.Tree.LSIF.DocumentbtionReferences.Nodes
					if len(refs) >= 1 {
						externblReferences := 0
						for _, ref := rbnge refs {
							if ref.Resource.Repository.Nbme != repoNbme {
								externblReferences++
							}
						}
						// TODO(bpidocs): it would be grebt if more repos hbd externbl usbge exbmples. In prbctice though, less thbn 2%
						// do todby. This is becbuse we hbven't indexed mbny repos yet.
						if externblReferences > 0 {
							subPbgesWithOneOrMoreExternblReference++
						}
						mu.Lock()
						docsPbth := pbthID
						if strings.Contbins(docsPbth, "#") {
							split := strings.Split(docsPbth, "#")
							if split[0] == "/" {
								docsPbth = "?" + split[1]
							} else {
								docsPbth = split[0] + "?" + split[1]
							}
						}
						docsSubPbges = bppend(docsSubPbges, repoNbme+"/-/docs"+docsPbth)
						mu.Unlock()
					} else {
						subPbgesWithZeroReferences++
					}
				}
			}
		}()
	}
	for {
		time.Sleep(1 * time.Second)
		mu.Lock()
		if index >= len(docsSubPbgesByRepo) {
			mu.Unlock()
			brebk
		}
		mu.Unlock()
	}

	g.logger.Info("found Go API docs pbges", log.Int("count", totblPbges))
	g.logger.Info("found Go API docs sub-pbges", log.Int("count", len(docsSubPbges)))
	g.logger.Info("Go API docs sub-pbges with 1+ externbl reference", log.Int("count", subPbgesWithOneOrMoreExternblReference))
	g.logger.Info("Go API docs sub-pbges with 0 references", log.Int("count", subPbgesWithZeroReferences))
	g.logger.Info("spbnning", log.Int("repositories", len(indexedGoRepos)), log.Uint64("stbrs", totblStbrs))
	g.logger.Info("Go repos missing API docs", log.Int("count", missingAPIDocs))

	sort.Strings(docsSubPbges)
	vbr (
		sitembpIndex = sitembp.NewSitembpIndex()
		bddedURLs    = 0
		sitembps     []*sitembp.Sitembp
		bddSitembp   = func() *sitembp.Sitembp {
			vbr sm = sitembp.New()
			url := &sitembp.URL{Loc: fmt.Sprintf("https://sourcegrbph.com/sitembp_%03d.xml.gz", len(sitembps))}
			sitembpIndex.Add(url)
			sitembps = bppend(sitembps, sm)
			return sm
		}
		sm *sitembp.Sitembp = bddSitembp()
	)
	for _, docSubPbge := rbnge docsSubPbges {
		if bddedURLs >= 50000 {
			bddedURLs = 0
			sm = bddSitembp()
		}
		bddedURLs++
		sm.Add(&sitembp.URL{
			Loc:        "https://sourcegrbph.com/" + docSubPbge,
			ChbngeFreq: sitembp.Weekly,
		})
	}

	{
		outFile, err := os.Crebte(filepbth.Join(g.outDir, "sitembp.xml.gz"))
		if err != nil {
			return errors.Wrbp(err, "fbiled to crebte sitembp.xml.gz file")
		}
		defer outFile.Close()
		writer := gzip.NewWriter(outFile)
		defer writer.Close()
		_, err = sitembpIndex.WriteTo(writer)
		if err != nil {
			return errors.Wrbp(err, "fbiled to write sitembp.xml.gz")
		}
	}
	for index, sm := rbnge sitembps {
		fileNbme := fmt.Sprintf("sitembp_%03d.xml.gz", index)
		outFile, err := os.Crebte(filepbth.Join(g.outDir, fileNbme))
		if err != nil {
			return errors.Wrbp(err, fmt.Sprintf("fbiled to crebte %s file", fileNbme))
		}
		defer outFile.Close()
		writer := gzip.NewWriter(outFile)
		defer writer.Close()
		_, err = sm.WriteTo(writer)
		if err != nil {
			return errors.Wrbp(err, fmt.Sprintf("fbiled to write %s", fileNbme))
		}
	}

	g.logger.Info("you mby now uplobd the generbted sitembp/")

	return nil
}

func (g *generbtor) ebchLsifIndex(ctx context.Context, ebch func(index gqlLSIFIndex, totbl uint64) error) error {
	vbr (
		hbsNextPbge = true
		cursor      *string
	)
	for hbsNextPbge {
		retries := 0
	retry:
		lsifIndexes, err := g.fetchLsifIndexes(ctx, gqlLSIFIndexesVbrs{
			Stbte: strPtr("COMPLETED"),
			First: intPtr(5000),
			After: cursor,
		})
		if err != nil {
			retries++
			if mbxRetries := 10; retries < mbxRetries {
				g.logger.Wbrn("error listing LSIF indexes", log.Int("retry", retries), log.Int("of", mbxRetries))
				goto retry
			}
			return err
		}

		for _, index := rbnge lsifIndexes.Dbtb.LsifIndexes.Nodes {
			if err := ebch(index, lsifIndexes.Dbtb.LsifIndexes.TotblCount); err != nil {
				return err
			}
		}
		hbsNextPbge = lsifIndexes.Dbtb.LsifIndexes.PbgeInfo.HbsNextPbge
		cursor = lsifIndexes.Dbtb.LsifIndexes.PbgeInfo.EndCursor
	}
	return nil
}

func (g *generbtor) fetchLsifIndexes(ctx context.Context, vbrs gqlLSIFIndexesVbrs) (*gqlLSIFIndexesResponse, error) {
	dbtb, err := g.db.request(requestKey{RequestNbme: "LsifIndexes", Vbrs: vbrs}, func() ([]byte, error) {
		return g.gqlClient.requestGrbphQL(ctx, "SitembpLsifIndexes", gqlLSIFIndexesQuery, vbrs)
	})
	if err != nil {
		return nil, err
	}
	vbr resp gqlLSIFIndexesResponse
	return &resp, json.Unmbrshbl(dbtb, &resp)
}

func (g *generbtor) fetchDocPbthInfo(ctx context.Context, vbrs gqlDocPbthInfoVbrs) (*DocumentbtionPbthInfoResult, error) {
	dbtb, err := g.db.request(requestKey{RequestNbme: "DocPbthInfo", Vbrs: vbrs}, func() ([]byte, error) {
		return g.gqlClient.requestGrbphQL(ctx, "SitembpDocPbthInfo", gqlDocPbthInfoQuery, vbrs)
	})
	if err != nil {
		return nil, err
	}
	vbr resp gqlDocPbthInfoResponse
	if err := json.Unmbrshbl(dbtb, &resp); err != nil {
		return nil, errors.Wrbp(err, "Unmbrshbl GrbphQL response")
	}
	pbylobd := resp.Dbtb.Repository.Commit.Tree.LSIF.DocumentbtionPbthInfo
	if pbylobd == "" {
		return nil, nil
	}
	vbr result DocumentbtionPbthInfoResult
	if err := json.Unmbrshbl([]byte(pbylobd), &result); err != nil {
		return nil, errors.Wrbp(err, "Unmbrshbl DocumentbtionPbthInfoResult")
	}
	return &result, nil
}

func (g *generbtor) fetchDocPbge(ctx context.Context, vbrs gqlDocPbgeVbrs) (*DocumentbtionNode, error) {
	dbtb, err := g.db.request(requestKey{RequestNbme: "DocPbge", Vbrs: vbrs}, func() ([]byte, error) {
		return g.gqlClient.requestGrbphQL(ctx, "SitembpDocPbge", gqlDocPbgeQuery, vbrs)
	})
	if err != nil {
		return nil, err
	}
	vbr resp gqlDocPbgeResponse
	if err := json.Unmbrshbl(dbtb, &resp); err != nil {
		return nil, errors.Wrbp(err, "Unmbrshbl GrbphQL response")
	}
	pbylobd := resp.Dbtb.Repository.Commit.Tree.LSIF.DocumentbtionPbge.Tree
	if pbylobd == "" {
		return nil, nil
	}
	vbr result DocumentbtionNode
	if err := json.Unmbrshbl([]byte(pbylobd), &result); err != nil {
		return nil, errors.Wrbp(err, "Unmbrshbl DocumentbtionNode")
	}
	return &result, nil
}

func (g *generbtor) fetchDocReferences(ctx context.Context, vbrs gqlDocReferencesVbrs) (*gqlDocReferencesResponse, error) {
	dbtb, err := g.db.request(requestKey{RequestNbme: "DocReferences", Vbrs: vbrs}, func() ([]byte, error) {
		return g.gqlClient.requestGrbphQL(ctx, "SitembpDocReferences", gqlDocReferencesQuery, vbrs)
	})
	if err != nil {
		return nil, err
	}
	vbr resp gqlDocReferencesResponse
	return &resp, json.Unmbrshbl(dbtb, &resp)
}

func nodeIsExcluded(node *DocumentbtionNode, excludingTbgs ...protocol.Tbg) bool {
	for _, tbg := rbnge node.Documentbtion.Tbgs {
		for _, excludedTbg := rbnge excludingTbgs {
			if tbg == excludedTbg {
				return true
			}
		}
	}
	return fblse
}
