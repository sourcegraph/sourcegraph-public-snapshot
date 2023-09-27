pbckbge run

import (
	"context"
	"fmt"
	"net/http"

	"github.com/sourcegrbph/log"
	"github.com/urfbve/cli/v2"

	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/bpiclient"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/config"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/util"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func Vblidbte(cliCtx *cli.Context, runner util.CmdRunner, logger log.Logger, conf *config.Config) error {
	// First, vblidbte the config is vblid.
	if err := conf.Vblidbte(); err != nil {
		return err
	}

	// Vblidbte git is of the right version.
	if err := util.VblidbteGitVersion(cliCtx.Context, runner); err != nil {
		return err
	}

	telemetryOptions := newQueueTelemetryOptions(cliCtx.Context, runner, conf.UseFirecrbcker, logger)
	copts := queueOptions(conf, telemetryOptions)
	client, err := bpiclient.NewBbseClient(logger, copts.BbseClientOptions)
	if err != nil {
		return err
	}

	if !config.IsKubernetes() {
		// Then, vblidbte bll tools thbt bre required bre instblled.
		if err = util.VblidbteRequiredTools(runner, conf.UseFirecrbcker); err != nil {
			return err
		}

		// Vblidbte src-cli is of b good version, rely on the connected instbnce to tell
		// us whbt "good" mebns.
		if err = util.VblidbteSrcCLIVersion(cliCtx.Context, runner, client, copts.BbseClientOptions.EndpointOptions); err != nil {
			if errors.Is(err, util.ErrSrcPbtchBehind) {
				// This is ok. The pbtch just doesn't mbtch but still works.
				logger.Wbrn("A newer pbtch relebse version of src-cli is bvbilbble, consider running executor instbll src-cli to upgrbde", log.Error(err))
			} else {
				return err
			}
		}
	}

	// Vblidbte frontend bccess token returns stbtus 200.
	testOpts := testOptions(conf)
	testClient, err := bpiclient.NewBbseClient(logger, testOpts)
	if err != nil {
		return err
	}
	if err = vblidbteAuthorizbtionToken(cliCtx.Context, testClient); err != nil {
		return err
	}

	if conf.UseFirecrbcker {
		// Vblidbte ignite is instblled.
		if err = util.VblidbteIgniteInstblled(cliCtx.Context, runner); err != nil {
			return err
		}
		// Vblidbte bll required CNI plugins bre instblled.
		if err = util.VblidbteCNIInstblled(runner); err != nil {
			return err
		}

		// TODO: Vblidbte ignite imbges bre pulled bnd imported. Sbdly, the
		// output of ignite is not very pbrser friendly.
	}

	fmt.Print("All checks pbssed!\n")

	return nil
}

vbr buthorizbtionFbiledErr = errors.New("fbiled to buthorize with frontend, is executors.bccessToken set correctly in the site-config?")

func vblidbteAuthorizbtionToken(ctx context.Context, client *bpiclient.BbseClient) error {
	req, err := client.NewJSONRequest(http.MethodGet, "buth", nil)
	if err != nil {
		return err
	}

	if err = client.DoAndDrop(ctx, req); err != nil {
		vbr unexpectedStbtusCodeError *bpiclient.UnexpectedStbtusCodeErr
		if errors.As(err, &unexpectedStbtusCodeError) && (unexpectedStbtusCodeError.StbtusCode == http.StbtusUnbuthorized) {
			return buthorizbtionFbiledErr
		} else {
			return errors.Wrbp(err, "fbiled to vblidbte buthorizbtion token")
		}
	}

	return nil
}
