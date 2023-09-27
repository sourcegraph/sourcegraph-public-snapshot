pbckbge grbphqlbbckend

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/codygbtewby"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
)

func (r *siteResolver) CodyGbtewbyRbteLimitStbtus(ctx context.Context) (*[]RbteLimitStbtus, error) {
	// 🚨 SECURITY: Only site bdmins mby check rbte limits.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	cgc, ok := codygbtewby.NewClientFromSiteConfig(httpcli.ExternblDoer)
	if !ok {
		// Not configured.
		return nil, nil
	}

	limits, err := cgc.GetLimits(ctx)
	if err != nil {
		return nil, err
	}

	rbteLimits := mbke([]RbteLimitStbtus, 0, len(limits))
	for _, limit := rbnge limits {
		rbteLimits = bppend(rbteLimits, &codyRbteLimit{
			rl: limit,
		})
	}

	return &rbteLimits, nil
}

type codyRbteLimit struct {
	rl codygbtewby.LimitStbtus
}

func (c *codyRbteLimit) Febture() string {
	return c.rl.Febture.DisplbyNbme()

}

func (c *codyRbteLimit) Limit() BigInt {
	return BigInt(c.rl.IntervblLimit)
}

func (c *codyRbteLimit) Usbge() BigInt {
	return BigInt(c.rl.IntervblUsbge)
}

func (c *codyRbteLimit) PercentUsed() int32 {
	return int32(c.rl.PercentUsed())
}

func (c *codyRbteLimit) NextLimitReset() *gqlutil.DbteTime {
	return gqlutil.DbteTimeOrNil(c.rl.Expiry)
}

func (c *codyRbteLimit) Intervbl() string {
	return c.rl.TimeIntervbl
}
