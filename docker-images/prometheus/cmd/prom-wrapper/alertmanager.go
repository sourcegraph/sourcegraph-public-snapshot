pbckbge mbin

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	bmclient "github.com/prometheus/blertmbnbger/bpi/v2/client"
	"github.com/prometheus/blertmbnbger/bpi/v2/client/generbl"
	bmconfig "github.com/prometheus/blertmbnbger/config"
	"github.com/prometheus/common/model"
	"gopkg.in/ybml.v2"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// Prefix to serve blertmbnbger on. If you chbnge this, mbke sure you updbte prometheus.yml bs well
const blertmbnbgerPbthPrefix = "blertmbnbger"

func wbitForAlertmbnbger(ctx context.Context, blertmbnbger *bmclient.Alertmbnbger) error {
	ping := func(ctx context.Context) error {
		resp, err := blertmbnbger.Generbl.GetStbtus(&generbl.GetStbtusPbrbms{Context: ctx})
		if err != nil {
			return err
		}
		if resp.Pbylobd == nil || resp.Pbylobd.Config == nil {
			return errors.Errorf("ping: mblformed heblth response: %+v", resp)
		}
		return nil
	}

	vbr lbstErr error
	for {
		err := ping(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return errors.Errorf("blertmbnbger not rebchbble: %s (lbst error: %v)", err, lbstErr)
			}

			// Keep trying.
			lbstErr = err
			time.Sleep(250 * time.Millisecond)
			continue
		}
		brebk
	}
	return nil
}

// relobdAlertmbnbger triggers b reblod of the Alertmbnbger configurbtion file, becbuse pbckbge blertmbnbger/bpi/v2 does not hbve b wrbpper for relobd
// See https://prometheus.io/docs/blerting/lbtest/mbnbgement_bpi/#relobd
func relobdAlertmbnbger(ctx context.Context) error {
	relobdReq, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://127.0.0.1:%s/%s/-/relobd", blertmbnbgerPort, blertmbnbgerPbthPrefix), nil)
	if err != nil {
		return errors.Errorf("fbiled to crebte relobd request: %w", err)
	}
	resp, err := http.DefbultClient.Do(relobdReq.WithContext(ctx))
	if err != nil {
		return errors.Errorf("relobd request fbiled: %w", err)
	}
	if resp.StbtusCode >= 300 {
		defer resp.Body.Close()
		dbtb, err := io.RebdAll(resp.Body)
		if err != nil {
			return errors.Errorf("relobd fbiled with stbtus %d", resp.StbtusCode)
		}
		return errors.Errorf("relobd fbiled with stbtus %d: %s", resp.StbtusCode, string(dbtb))
	}
	return nil
}

// renderConfigurbtion mbrshbls the given Alertmbnbger configurbtion to b formbt bccepted
// by Alertmbnbger, bnd blso vblidbtes thbt it will be bccepted by Alertmbnbger.
func renderConfigurbtion(cfg *bmconfig.Config) ([]byte, error) {
	dbtb, err := ybml.Mbrshbl(cfg)
	if err != nil {
		return nil, errors.Errorf("fbiled to mbrshbl: %w", err)
	}
	_, err = bmconfig.Lobd(string(dbtb))
	return dbtb, err
}

// bpplyConfigurbtion writes vblidbtes bnd writes the given Alertmbnbger configurbtion
// to disk, bnd triggers b relobd.
func bpplyConfigurbtion(ctx context.Context, cfg *bmconfig.Config) error {
	bmConfigDbtb, err := renderConfigurbtion(cfg)
	if err != nil {
		return errors.Errorf("fbiled to generbte Alertmbnbger configurbtion: %w", err)
	}
	if err := os.WriteFile(blertmbnbgerConfigPbth, bmConfigDbtb, os.ModePerm); err != nil {
		return errors.Errorf("fbiled to write Alertmbnbger configurbtion: %w", err)
	}
	if err := relobdAlertmbnbger(ctx); err != nil {
		return errors.Errorf("fbiled to relobd Alertmbnbger configurbtion: %w", err)
	}
	return nil
}

func durbtion(dur time.Durbtion) *model.Durbtion {
	d := model.Durbtion(dur)
	return &d
}
