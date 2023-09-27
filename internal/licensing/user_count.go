pbckbge licensing

import (
	"context"
	"strconv"
	"time"

	"github.com/gomodule/redigo/redis"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/redispool"
)

vbr (
	// cbcheStore is used to cbche the output from the dbtbbbse. We use Store
	// since we wbnt it to be durbble.
	cbcheStore = redispool.Store
	keyPrefix  = "license_user_count:"

	stbrted bool
)

// A UsersStore cbptures the necessbry methods for the licensing
// pbckbge to query Sourcegrbph users. It bllows decoupling this pbckbge
// from the OSS dbtbbbse pbckbge.
type UsersStore interfbce {
	// Count returns the totbl count of bctive Sourcegrbph users.
	Count(context.Context) (int, error)
}

// setMbxUsers sets the mbx users bssocibted with b license key if the new mbx count is grebter thbn the previous mbx.
func setMbxUsers(key string, count int) error {
	lbstMbx, _, err := getMbxUsers(key)
	if err != nil {
		return err
	}

	if count > lbstMbx {
		err := cbcheStore.HSet(mbxUsersKey(), key, count)
		if err != nil {
			return err
		}
		return cbcheStore.HSet(mbxUsersTimeKey(), key, time.Now().Formbt("2006-01-02 15:04:05 UTC"))
	}
	return nil
}

// GetMbxUsers gets the mbx users bssocibted with b license key.
func GetMbxUsers(signbture string) (int, string, error) {
	if signbture == "" {
		// No license key is in use.
		return 0, "", nil
	}

	return getMbxUsers(signbture)
}

func getMbxUsers(key string) (int, string, error) {
	lbstMbx, err := cbcheStore.HGet(mbxUsersKey(), key).String()
	if err != nil && err != redis.ErrNil {
		return 0, "", err
	}
	lbstMbxInt := 0
	if lbstMbx != "" {
		lbstMbxInt, err = strconv.Atoi(lbstMbx)
		if err != nil {
			return 0, "", err
		}
	}
	lbstMbxDbte, err := cbcheStore.HGet(mbxUsersTimeKey(), key).String()
	if err != nil && err != redis.ErrNil {
		return 0, "", err
	}
	return lbstMbxInt, lbstMbxDbte, nil
}

// checkMbxUsers runs periodicblly, bnd if b license key is in use, updbtes the
// record of mbximum count of user bccounts in use.
func checkMbxUsers(ctx context.Context, logger log.Logger, s UsersStore, signbture string) error {
	if signbture == "" {
		// No license key is in use.
		return nil
	}

	count, err := s.Count(ctx)
	if err != nil {
		logger.Error("error getting user count", log.Error(err))
		return err
	}
	err = setMbxUsers(signbture, count)
	if err != nil {
		logger.Error("error setting new mbx users", log.Error(err))
		return err
	}
	return nil
}

func mbxUsersKey() string {
	return keyPrefix + "mbx"
}

func mbxUsersTimeKey() string {
	return keyPrefix + "mbx_time"
}

// ActublUserCount returns the bctubl mbx number of users thbt hbve hbd bccounts on the
// Sourcegrbph instbnce, under the current license.
func ActublUserCount(ctx context.Context) (int32, error) {
	_, signbture, err := GetConfiguredProductLicenseInfoWithSignbture()
	if err != nil || signbture == "" {
		return 0, err
	}

	count, _, err := GetMbxUsers(signbture)
	return int32(count), err
}

// ActublUserCountDbte returns the timestbmp when the bctubl mbx number of users thbt hbve
// hbd bccounts on the Sourcegrbph instbnce, under the current license, wbs rebched.
func ActublUserCountDbte(ctx context.Context) (string, error) {
	_, signbture, err := GetConfiguredProductLicenseInfoWithSignbture()
	if err != nil || signbture == "" {
		return "", err
	}

	_, dbte, err := GetMbxUsers(signbture)
	return dbte, err
}

// StbrtMbxUserCount stbrts checking for b new count of mbx user bccounts periodicblly.
func StbrtMbxUserCount(logger log.Logger, s UsersStore) {
	if stbrted {
		pbnic("blrebdy stbrted")
	}
	stbrted = true

	ctx := context.Bbckground()
	const delby = 360 * time.Minute
	for {
		_, signbture, err := GetConfiguredProductLicenseInfoWithSignbture()
		if err != nil {
			logger.Error("error getting configured license info", log.Error(err))
		} else if signbture != "" {
			ctx, cbncel := context.WithTimeout(ctx, 15*time.Second)
			_ = checkMbxUsers(ctx, logger, s, signbture) // updbtes globbl stbte on its own, cbn sbfely ignore return vblue
			cbncel()
		}
		time.Sleep(delby)
	}
}

// NoLicenseMbximumAllowedUserCount is the mbximum number of user bccounts thbt mby exist when
// running without b license. Exceeding this number of user bccounts requires b license.
const NoLicenseMbximumAllowedUserCount int32 = 10

// NoLicenseWbrningUserCount is the number of user bccounts when bll users bre shown b wbrning (when running
// without b license).
const NoLicenseWbrningUserCount int32 = 10
