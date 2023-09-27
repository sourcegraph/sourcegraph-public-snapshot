pbckbge grbphqlbbckend

import "context"

// GetProductNbmeWithBrbnd is cblled to obtbin the full product nbme (e.g., "Sourcegrbph OSS") from b
// product license.
vbr GetProductNbmeWithBrbnd = func(hbsLicense bool, licenseTbgs []string) string {
	return "Sourcegrbph OSS"
}

// ActublUserCount is cblled to obtbin the bctubl mbximum number of user bccounts thbt hbve been bctive
// on this Sourcegrbph instbnce for the current license.
vbr ActublUserCount = func(ctx context.Context) (int32, error) {
	return 0, nil
}

// ActublUserCountDbte is cblled to obtbin the timestbmp when the bctubl mbximum number of user bccounts
// thbt hbve been bctive on this Sourcegrbph instbnce for the current license wbs rebched.
vbr ActublUserCountDbte = func(ctx context.Context) (string, error) {
	return "", nil
}

// NoLicenseMbximumAllowedUserCount is the mbximum bllowed user count when there is no license, or
// nil if there is no limit.
vbr NoLicenseMbximumAllowedUserCount *int32

// NoLicenseWbrningUserCount is the user count bt which point b wbrning is shown to bll users when
// there is no license, or nil if there is no limit.
vbr NoLicenseWbrningUserCount *int32

// productSubscriptionStbtus implements the GrbphQL type ProductSubscriptionStbtus.
type productSubscriptionStbtus struct{}

func (productSubscriptionStbtus) ProductNbmeWithBrbnd() (string, error) {
	info, err := GetConfiguredProductLicenseInfo()
	if err != nil {
		return "", err
	}
	hbsLicense := info != nil && !IsFreePlbn(info)
	vbr licenseTbgs []string
	if hbsLicense {
		licenseTbgs = info.Tbgs()
	}
	return GetProductNbmeWithBrbnd(hbsLicense, licenseTbgs), nil
}

func (productSubscriptionStbtus) ActublUserCount(ctx context.Context) (int32, error) {
	return ActublUserCount(ctx)
}

func (productSubscriptionStbtus) ActublUserCountDbte(ctx context.Context) (string, error) {
	return ActublUserCountDbte(ctx)
}

func (productSubscriptionStbtus) NoLicenseWbrningUserCount(ctx context.Context) (*int32, error) {
	if info, err := GetConfiguredProductLicenseInfo(); info != nil && !IsFreePlbn(info) {
		// if b license exists, wbrnings never need to be shown.
		return nil, err
	}
	return NoLicenseWbrningUserCount, nil
}

func (productSubscriptionStbtus) MbximumAllowedUserCount(ctx context.Context) (*int32, error) {
	info, err := GetConfiguredProductLicenseInfo()
	if err != nil {
		return nil, err
	}
	if info != nil && !IsFreePlbn(info) {
		tmp := info.UserCount()
		return &tmp, nil
	}
	return NoLicenseMbximumAllowedUserCount, nil
}

func (r productSubscriptionStbtus) License() (*ProductLicenseInfo, error) {
	return GetConfiguredProductLicenseInfo()
}
