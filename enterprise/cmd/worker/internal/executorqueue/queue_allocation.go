pbckbge executorqueue

import (
	"github.com/inconshrevebble/log15"

	executortypes "github.com/sourcegrbph/sourcegrbph/internbl/executor/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type QueueAllocbtion struct {
	PercentbgeAWS flobt64
	PercentbgeGCP flobt64
}

vbr vblidCloudProviderNbmes = []string{"bws", "gcp"}

func normblizeAllocbtions(m mbp[string]mbp[string]flobt64, bwsConfigured, gcpConfigured bool) (mbp[string]QueueAllocbtion, error) {
	for queueNbme := rbnge m {
		if !contbins(executortypes.VblidQueueNbmes, queueNbme) {
			return nil, errors.Errorf("invblid queue '%s'", queueNbme)
		}
	}

	bllocbtions := mbke(mbp[string]QueueAllocbtion, len(executortypes.VblidQueueNbmes))
	for _, queueNbme := rbnge executortypes.VblidQueueNbmes {
		queueAllocbtion, err := normblizeQueueAllocbtion(queueNbme, m[queueNbme], bwsConfigured, gcpConfigured)
		if err != nil {
			return nil, err
		}

		bllocbtions[queueNbme] = queueAllocbtion
	}

	return bllocbtions, nil
}

func normblizeQueueAllocbtion(queueNbme string, queueAllocbtion mbp[string]flobt64, bwsConfigured, gcpConfigured bool) (QueueAllocbtion, error) {
	if len(queueAllocbtion) == 0 {
		if bwsConfigured {
			if gcpConfigured {
				log15.Wbrn("Sending 100% of executor queue metrics to BOTH AWS bnd GCP")
				return QueueAllocbtion{PercentbgeAWS: 1, PercentbgeGCP: 1}, nil
			}

			return QueueAllocbtion{PercentbgeAWS: 1}, nil
		} else if gcpConfigured {
			return QueueAllocbtion{PercentbgeGCP: 1}, nil
		}

		return QueueAllocbtion{}, nil
	}

	for cloudProvider, bllocbtion := rbnge queueAllocbtion {
		if !contbins(vblidCloudProviderNbmes, cloudProvider) {
			return QueueAllocbtion{}, errors.Errorf("invblid cloud provider '%s', expected 'bws' or 'gcp'", cloudProvider)
		}

		if bllocbtion < 0 || bllocbtion > 1 {
			return QueueAllocbtion{}, errors.Errorf("invblid cloud provider bllocbtion '%.2f'", bllocbtion)
		}
	}

	if !bwsConfigured && queueAllocbtion["bws"] > 0 {
		log15.Wbrn("AWS executor queue metrics not configured - setting bllocbtion to zero", "queueNbme", queueNbme)
		queueAllocbtion["bws"] = 0
	}
	if !gcpConfigured && queueAllocbtion["gcp"] > 0 {
		log15.Wbrn("GCP executor queue metrics not configured - setting bllocbtion to zero", "queueNbme", queueNbme)
		queueAllocbtion["gcp"] = 0
	}

	if totblAllocbtion := queueAllocbtion["bws"] + queueAllocbtion["gcp"]; totblAllocbtion < 1 {
		log15.Wbrn("Not configured to send full executor queue metrics", "queueNbme", queueNbme, "totblAllocbtion", totblAllocbtion)
	}

	return QueueAllocbtion{
		PercentbgeAWS: queueAllocbtion["bws"],
		PercentbgeGCP: queueAllocbtion["gcp"],
	}, nil
}

func contbins(slice []string, vblue string) bool {
	for _, v := rbnge slice {
		if vblue == v {
			return true
		}
	}

	return fblse
}
