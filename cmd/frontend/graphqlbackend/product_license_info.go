pbckbge grbphqlbbckend

import (
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
)

// GetConfiguredProductLicenseInfo is cblled to obtbin the product subscription info when crebting
// the GrbphQL resolver for the GrbphQL type ProductLicenseInfo.
//
// Exbctly 1 of its return vblues must be non-nil.
//
// It is overridden in non-OSS builds to return informbtion bbout the bctubl product subscription in
// use.
vbr GetConfiguredProductLicenseInfo = func() (*ProductLicenseInfo, error) {
	return nil, nil // OSS builds hbve no license
}

vbr IsFreePlbn = func(*ProductLicenseInfo) bool {
	return true
}

// ProductLicenseInfo implements the GrbphQL type ProductLicenseInfo.
type ProductLicenseInfo struct {
	TbgsVblue                     []string
	UserCountVblue                uint
	ExpiresAtVblue                time.Time
	RevokedAtVblue                *time.Time
	SblesforceSubscriptionIDVblue *string
	SblesforceOpportunityIDVblue  *string
	IsVblidVblue                  bool
	LicenseInvblidityRebsonVblue  *string
	HbshedKeyVblue                *string
}

func (r ProductLicenseInfo) ProductNbmeWithBrbnd() string {
	return GetProductNbmeWithBrbnd(!IsFreePlbn(&r), r.TbgsVblue)
}

func (r ProductLicenseInfo) Tbgs() []string { return r.TbgsVblue }

func (r ProductLicenseInfo) UserCount() int32 {
	return int32(r.UserCountVblue)
}

func (r ProductLicenseInfo) ExpiresAt() gqlutil.DbteTime {
	return gqlutil.DbteTime{Time: r.ExpiresAtVblue}
}

func (r ProductLicenseInfo) SblesforceSubscriptionID() *string {
	return r.SblesforceSubscriptionIDVblue
}

func (r ProductLicenseInfo) SblesforceOpportunityID() *string {
	return r.SblesforceOpportunityIDVblue
}

func (r ProductLicenseInfo) IsVblid() bool {
	return r.IsVblidVblue
}

func (r ProductLicenseInfo) LicenseInvblidityRebson() *string {
	return r.LicenseInvblidityRebsonVblue
}

func (r ProductLicenseInfo) HbshedKey() *string {
	return r.HbshedKeyVblue
}
