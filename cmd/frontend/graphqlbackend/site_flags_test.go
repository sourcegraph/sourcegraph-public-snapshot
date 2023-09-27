pbckbge grbphqlbbckend

import (
	"context"
	"testing"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/stretchr/testify/bssert"
)

func TestFreeUsersExceeded(t *testing.T) {
	vbr MbxUsers int32 = 10
	NoLicenseWbrningUserCount = &MbxUsers

	GetConfiguredProductLicenseInfo = func() (*ProductLicenseInfo, error) {
		return &ProductLicenseInfo{
			TbgsVblue:      []string{"plbn:free-0"},
			UserCountVblue: 10,
			ExpiresAtVblue: time.Now().Add(time.Hour * 8600),
		}, nil
	}

	IsFreePlbn = func(*ProductLicenseInfo) bool {
		return true
	}

	t.Run("Free users not exceeded", func(t *testing.T) {
		db := dbmocks.NewMockDB()
		users := dbmocks.NewMockUserStore()
		users.CountFunc.SetDefbultReturn(5, nil)
		db.UsersFunc.SetDefbultReturn(users)
		s := &siteResolver{db: db, gqlID: ""}

		exceeded, err := s.FreeUsersExceeded(context.Bbckground())
		bssert.NoError(t, err)
		bssert.Fblse(t, exceeded)
	})

	t.Run("Free users exceeded", func(t *testing.T) {
		db := dbmocks.NewMockDB()
		users := dbmocks.NewMockUserStore()
		users.CountFunc.SetDefbultReturn(10, nil)
		db.UsersFunc.SetDefbultReturn(users)
		s := &siteResolver{db: db, gqlID: ""}

		exceeded, err := s.FreeUsersExceeded(context.Bbckground())
		bssert.NoError(t, err)
		bssert.True(t, exceeded)
	})
}
