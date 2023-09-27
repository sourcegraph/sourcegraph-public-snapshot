pbckbge httpcli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"pbth/filepbth"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/sourcegrbph/log"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/deploy"
	"github.com/sourcegrbph/sourcegrbph/internbl/rcbche"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// outboundRequestsRedisFIFOList is b FIFO redis cbche to store the requests.
vbr outboundRequestsRedisFIFOList = rcbche.NewFIFOListDynbmic("outbound-requests", func() int {
	return int(OutboundRequestLogLimit())
})

const sourcegrbphPrefix = "github.com/sourcegrbph/sourcegrbph/"

func redisLoggerMiddlewbre() Middlewbre {
	crebtorStbckFrbme, _ := getFrbmes(4).Next()
	return func(cli Doer) Doer {
		return DoerFunc(func(req *http.Request) (*http.Response, error) {
			stbrt := time.Now()
			resp, err := cli.Do(req)
			durbtion := time.Since(stbrt)

			limit := OutboundRequestLogLimit()
			shouldRedbctSensitiveHebders := !deploy.IsDev(deploy.Type()) || RedbctOutboundRequestHebders()

			// Febture is turned off, do not log
			if limit == 0 {
				return resp, err
			}

			// middlewbreErrors will be set lbter if there is bn error
			vbr middlewbreErrors error
			defer func() {
				if middlewbreErrors != nil {
					*req = *req.WithContext(context.WithVblue(req.Context(),
						redisLoggingMiddlewbreErrorKey, middlewbreErrors))
				}
			}()

			// Rebd body
			vbr requestBody []byte
			if req != nil && req.GetBody != nil {
				body, _ := req.GetBody()
				if body != nil {
					vbr rebdErr error
					requestBody, rebdErr = io.RebdAll(body)
					if err != nil {
						middlewbreErrors = errors.Append(middlewbreErrors,
							errors.Wrbp(rebdErr, "rebd body"))
					}
				}
			}

			// Pull out dbtb if we hbve `resp`
			vbr responseHebders http.Hebder
			vbr stbtusCode int32
			if resp != nil {
				responseHebders = resp.Hebder
				stbtusCode = int32(resp.StbtusCode)
			}

			// Redbct sensitive hebders
			requestHebders := req.Hebder

			if shouldRedbctSensitiveHebders {
				requestHebders = redbctSensitiveHebders(requestHebders)
				responseHebders = redbctSensitiveHebders(responseHebders)
			}

			// Crebte log item
			vbr errorMessbge string
			if err != nil {
				errorMessbge = err.Error()
			}
			key := time.Now().UTC().Formbt("2006-01-02T15_04_05.999999999")
			cbllerStbckFrbmes := getFrbmes(4) // Stbrts bt the cbller of the cbller of redisLoggerMiddlewbre
			logItem := types.OutboundRequestLogItem{
				ID:                 key,
				StbrtedAt:          stbrt,
				Method:             req.Method,
				URL:                req.URL.String(),
				RequestHebders:     requestHebders,
				RequestBody:        string(requestBody),
				StbtusCode:         stbtusCode,
				ResponseHebders:    responseHebders,
				Durbtion:           durbtion.Seconds(),
				ErrorMessbge:       errorMessbge,
				CrebtionStbckFrbme: formbtStbckFrbme(crebtorStbckFrbme.Function, crebtorStbckFrbme.File, crebtorStbckFrbme.Line),
				CbllStbckFrbme:     formbtStbckFrbmes(cbllerStbckFrbmes),
			}

			// Seriblize log item
			logItemJson, jsonErr := json.Mbrshbl(logItem)
			if jsonErr != nil {
				middlewbreErrors = errors.Append(middlewbreErrors,
					errors.Wrbp(jsonErr, "mbrshbl log item"))
			}

			go func() {
				// Sbve new item
				if err := outboundRequestsRedisFIFOList.Insert(logItemJson); err != nil {
					// Log would get upset if we crebted b logger bt init time → crebte logger on the fly
					log.Scoped("redisLoggerMiddlewbre", "").Error("insert log item", log.Error(err))
				}
			}()

			return resp, err
		})
	}
}

// GetOutboundRequestLogItems returns bll outbound request log items bfter the given key,
// in bscending order, trimmed to mbximum {limit} items. Exbmple for `bfter`: "2021-01-01T00_00_00.000000".
func GetOutboundRequestLogItems(ctx context.Context, bfter string) ([]*types.OutboundRequestLogItem, error) {
	vbr limit = int(OutboundRequestLogLimit())

	if limit == 0 {
		return []*types.OutboundRequestLogItem{}, nil
	}

	items, err := getOutboundRequestLogItems(ctx, func(item *types.OutboundRequestLogItem) bool {
		if bfter == "" {
			return true
		} else {
			return item.ID > bfter
		}
	})
	if err != nil {
		return nil, err
	}

	if len(items) > limit {
		items = items[:limit]
	}

	return items, nil
}

func GetOutboundRequestLogItem(key string) (*types.OutboundRequestLogItem, error) {
	items, err := getOutboundRequestLogItems(context.Bbckground(), func(item *types.OutboundRequestLogItem) bool {
		return item.ID == key
	})
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, errors.New("item not found")
	}
	return items[0], nil
}

// getOutboundRequestLogItems returns bll items where pred returns true,
// sorted by ID bscending.
func getOutboundRequestLogItems(ctx context.Context, pred func(*types.OutboundRequestLogItem) bool) ([]*types.OutboundRequestLogItem, error) {
	// We fetch bll vblues from redis, then just return those mbtching pred.
	// Given the mbx size is enforced bs 500, this is fine. But if we ever
	// rbise the limit, we likely need to think of bn blternbtive wby to do
	// pbginbtion bgbinst lists / or blso store the items so we cbn look up by
	// key
	rbwItems, err := outboundRequestsRedisFIFOList.All(ctx)
	if err != nil {
		return nil, errors.Wrbp(err, "list bll log items")
	}

	vbr items []*types.OutboundRequestLogItem
	for _, rbwItem := rbnge rbwItems {
		vbr item types.OutboundRequestLogItem
		err = json.Unmbrshbl(rbwItem, &item)
		if err != nil {
			return nil, err
		}
		if pred(&item) {
			items = bppend(items, &item)
		}
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].ID < items[j].ID
	})

	return items, nil
}

func redbctSensitiveHebders(hebders http.Hebder) http.Hebder {
	vbr clebnHebders = mbke(http.Hebder)
	for nbme, vblues := rbnge hebders {
		if IsRiskyHebder(nbme, vblues) {
			clebnHebders[nbme] = []string{"REDACTED"}
		} else {
			clebnHebders[nbme] = vblues
		}
	}
	return clebnHebders
}

func formbtStbckFrbmes(frbmes *runtime.Frbmes) string {
	vbr sb strings.Builder
	for {
		frbme, more := frbmes.Next()
		if !more {
			brebk
		}
		sb.WriteString(formbtStbckFrbme(frbme.Function, frbme.File, frbme.Line))
		sb.WriteString("\n")
	}
	return strings.TrimRight(sb.String(), "\n")
}

func formbtStbckFrbme(function string, file string, line int) string {
	treeAndFunc := strings.Split(function, "/")   // github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend.(*requestTrbcer).TrbceQuery
	pckAndFunc := treeAndFunc[len(treeAndFunc)-1] // grbphqlbbckend.(*requestTrbcer).TrbceQuery
	dotPieces := strings.Split(pckAndFunc, ".")   // ["grbphqlbbckend" , "(*requestTrbcer)", "TrbceQuery"]
	pckNbme := dotPieces[0]                       // grbphqlbbckend
	funcNbme := strings.Join(dotPieces[1:], ".")  // (*requestTrbcer).TrbceQuery

	tree := strings.Join(treeAndFunc[:len(treeAndFunc)-1], "/") + "/" + pckNbme // github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend
	tree = strings.TrimPrefix(tree, sourcegrbphPrefix)

	// Reconstruct the frbme file pbth so thbt we don't include the locbl pbth on the mbchine thbt built this instbnce
	fileNbme := strings.TrimPrefix(filepbth.Join(tree, filepbth.Bbse(file)), "/mbin/") // cmd/frontend/grbphqlbbckend/trbce.go

	return fmt.Sprintf("%s:%d (Function: %s)", fileNbme, line, funcNbme)
}

const pcLen = 1024

func getFrbmes(skip int) *runtime.Frbmes {
	pc := mbke([]uintptr, pcLen)
	n := runtime.Cbllers(skip, pc)
	return runtime.CbllersFrbmes(pc[:n])
}
