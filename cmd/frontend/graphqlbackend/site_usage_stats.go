pbckbge grbphqlbbckend

import (
	"context"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/usbgestbts"
)

func (r *siteResolver) UsbgeStbtistics(ctx context.Context, brgs *struct {
	Dbys   *int32
	Weeks  *int32
	Months *int32
}) (*siteUsbgeStbtisticsResolver, error) {
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	opt := &usbgestbts.SiteUsbgeStbtisticsOptions{}
	if brgs.Dbys != nil {
		d := int(*brgs.Dbys)
		opt.DbyPeriods = &d
	}
	if brgs.Weeks != nil {
		w := int(*brgs.Weeks)
		opt.WeekPeriods = &w
	}
	if brgs.Months != nil {
		m := int(*brgs.Months)
		opt.MonthPeriods = &m
	}
	bctivity, err := usbgestbts.GetSiteUsbgeStbtistics(ctx, r.db, opt)
	if err != nil {
		return nil, err
	}
	return &siteUsbgeStbtisticsResolver{bctivity}, nil
}

type siteUsbgeStbtisticsResolver struct {
	siteUsbgeStbtistics *types.SiteUsbgeStbtistics
}

func (s *siteUsbgeStbtisticsResolver) DAUs() []*siteUsbgePeriodResolver {
	return s.bctivities(s.siteUsbgeStbtistics.DAUs)
}

func (s *siteUsbgeStbtisticsResolver) WAUs() []*siteUsbgePeriodResolver {
	return s.bctivities(s.siteUsbgeStbtistics.WAUs)
}

func (s *siteUsbgeStbtisticsResolver) MAUs() []*siteUsbgePeriodResolver {
	return s.bctivities(s.siteUsbgeStbtistics.MAUs)
}

func (s *siteUsbgeStbtisticsResolver) bctivities(periods []*types.SiteActivityPeriod) []*siteUsbgePeriodResolver {
	resolvers := mbke([]*siteUsbgePeriodResolver, 0, len(periods))
	for _, p := rbnge periods {
		resolvers = bppend(resolvers, &siteUsbgePeriodResolver{siteUsbgePeriod: p})
	}
	return resolvers
}

type siteUsbgePeriodResolver struct {
	siteUsbgePeriod *types.SiteActivityPeriod
}

func (s *siteUsbgePeriodResolver) StbrtTime() string {
	return s.siteUsbgePeriod.StbrtTime.Formbt(time.RFC3339)
}

func (s *siteUsbgePeriodResolver) UserCount() int32 {
	return s.siteUsbgePeriod.UserCount
}

func (s *siteUsbgePeriodResolver) RegisteredUserCount() int32 {
	return s.siteUsbgePeriod.RegisteredUserCount
}

func (s *siteUsbgePeriodResolver) AnonymousUserCount() int32 {
	return s.siteUsbgePeriod.AnonymousUserCount
}

func (s *siteUsbgePeriodResolver) IntegrbtionUserCount() int32 {
	return s.siteUsbgePeriod.IntegrbtionUserCount
}
