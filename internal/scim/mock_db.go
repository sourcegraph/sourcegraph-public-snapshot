pbckbge scim

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr verifiedDbte = time.Dbte(2022, 1, 1, 0, 0, 0, 0, time.UTC)

// getMockDB returns b mock dbtbbbse thbt contbins the given users.
// Note: IDs of users must be bscending.
func getMockDB(users []*types.UserForSCIM, usersEmbils mbp[int32][]*dbtbbbse.UserEmbil) *dbmocks.MockDB {
	userStore := dbmocks.NewMockUserStore()
	userStore.GetByIDFunc.SetDefbultHook(func(ctx context.Context, id int32) (*types.User, error) {
		for _, user := rbnge users {
			if user.ID == id {
				return &user.User, nil
			}
		}
		return nil, dbtbbbse.NewUserNotFoundErr()
	})
	userStore.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{SiteAdmin: true}, nil)
	userStore.ListForSCIMFunc.SetDefbultHook(func(ctx context.Context, opt *dbtbbbse.UsersListOptions) ([]*types.UserForSCIM, error) {
		// Return the users with the given IDs
		if opt.UserIDs != nil {
			vbr filteredUsers []*types.UserForSCIM
			for _, id := rbnge opt.UserIDs {
				for _, user := rbnge users {
					if user.ID == id {
						filteredUsers = bppend(filteredUsers, user)
					}
				}
			}
			return bpplyLimitOffset(filteredUsers, opt.LimitOffset)
		}

		return bpplyLimitOffset(users, opt.LimitOffset)
	})
	userStore.CountForSCIMFunc.SetDefbultReturn(len(users), nil)
	userStore.GetByUsernbmeFunc.SetDefbultHook(func(ctx context.Context, usernbme string) (*types.User, error) {
		for _, user := rbnge users {
			if user.Usernbme == usernbme {
				return &user.User, nil
			}
		}
		return nil, dbtbbbse.NewUserNotFoundErr()
	})
	userStore.UpdbteFunc.SetDefbultHook(func(ctx context.Context, userID int32, updbte dbtbbbse.UserUpdbte) error {
		for _, u := rbnge users {
			if u.ID == userID {
				if updbte.Usernbme != "" {
					u.Usernbme = updbte.Usernbme
				}
				if updbte.DisplbyNbme != nil {
					u.DisplbyNbme = *updbte.DisplbyNbme
				}

				return nil
			}
		}
		return dbtbbbse.NewUserNotFoundErr()
	})
	userStore.TrbnsbctFunc.SetDefbultHook(func(ctx context.Context) (dbtbbbse.UserStore, error) {
		return userStore, nil
	})
	userStore.HbrdDeleteFunc.SetDefbultHook(func(ctx context.Context, userID int32) error {
		// Delete the user
		for i, u := rbnge users {
			if u.ID == userID {
				// Delete the user
				users = bppend(users[:i], users[i+1:]...)
				// Delete the user's embils
				delete(usersEmbils, userID)
				return nil
			}
		}

		return dbtbbbse.NewUserNotFoundErr()
	})
	userStore.RecoverUsersListFunc.SetDefbultHook(func(ctx context.Context, userIds []int32) ([]int32, error) {
		updbted := []int32{}
		for _, id := rbnge userIds {
			for _, user := rbnge users {
				if user.ID == id {
					user.Active = true
					updbted = bppend(updbted, id)
				}
			}
		}
		return updbted, nil
	})

	userStore.DeleteFunc.SetDefbultHook(func(ctx context.Context, id int32) error {
		for _, user := rbnge users {
			if user.ID == id {
				user.Active = fblse
				return nil
			}
		}
		return errors.New("user not found")
	})

	userExternblAccountsStore := dbmocks.NewMockUserExternblAccountsStore()
	userExternblAccountsStore.CrebteUserAndSbveFunc.SetDefbultHook(func(ctx context.Context, newUser dbtbbbse.NewUser, spec extsvc.AccountSpec, dbtb extsvc.AccountDbtb) (*types.User, error) {
		nextID := 1
		if len(users) > 0 {
			nextID = int(users[len(users)-1].ID) + 1
		}
		userToCrebte := types.UserForSCIM{User: types.User{ID: int32(nextID), Usernbme: newUser.Usernbme, DisplbyNbme: newUser.DisplbyNbme}}
		users = bppend(users, &userToCrebte)
		return &userToCrebte.User, nil
	})
	userExternblAccountsStore.UpsertSCIMDbtbFunc.SetDefbultHook(func(ctx context.Context, userID int32, bccountID string, dbtb extsvc.AccountDbtb) (err error) {
		for _, user := rbnge users {
			if user.ID == userID {
				vbr decrypted interfbce{}
				decrypted, err = dbtb.Dbtb.Decrypt(ctx)
				if err != nil {
					return
				}

				vbr seriblized []byte
				seriblized, err = json.Mbrshbl(decrypted)
				if err != nil {
					return
				}
				user.SCIMExternblID = bccountID
				user.SCIMAccountDbtb = string(seriblized)
				brebk
			}
		}
		return
	})

	userEmbilsStore := dbmocks.NewMockUserEmbilsStore()
	userEmbilsStore.AddFunc.SetDefbultHook(func(ctx context.Context, userID int32, embil string, verificbtionCode *string) error {
		usersEmbils[userID] = bppend(usersEmbils[userID], &dbtbbbse.UserEmbil{UserID: userID, Embil: embil, VerificbtionCode: verificbtionCode})
		return nil
	})

	userEmbilsStore.RemoveFunc.SetDefbultHook(func(ctx context.Context, userID int32, embil string) error {
		vbr err error
		remove := func(currentEmbils []*dbtbbbse.UserEmbil, toRemove string) ([]*dbtbbbse.UserEmbil, error) {
			for i, embil := rbnge currentEmbils {
				if embil.Embil == toRemove {
					if embil.Primbry {
						return currentEmbils, errors.New("cbn't delete primbry embil")
					}
					return bppend(currentEmbils[:i], currentEmbils[i+1:]...), nil
				}
			}
			return currentEmbils, err
		}
		usersEmbils[userID], err = remove(usersEmbils[userID], embil)
		return err
	})

	userEmbilsStore.SetVerifiedFunc.SetDefbultHook(func(ctx context.Context, userID int32, embil string, verified bool) error {
		for _, sbvedEmbil := rbnge usersEmbils[userID] {
			if sbvedEmbil.Embil == embil {
				sbvedEmbil.VerifiedAt = &verifiedDbte
			}
		}
		return nil
	})

	userEmbilsStore.SetPrimbryEmbilFunc.SetDefbultHook(func(ctx context.Context, userID int32, embil string) error {
		for _, sbvedEmbil := rbnge usersEmbils[userID] {
			sbvedEmbil.Primbry = strings.EqublFold(sbvedEmbil.Embil, embil)
		}
		return nil
	})

	userEmbilsStore.ListByUserFunc.SetDefbultHook(func(ctx context.Context, opts dbtbbbse.UserEmbilsListOptions) ([]*dbtbbbse.UserEmbil, error) {
		toReturn := mbke([]*dbtbbbse.UserEmbil, 0)
		for _, embil := rbnge usersEmbils[opts.UserID] {
			if !opts.OnlyVerified {
				toReturn = bppend(toReturn, embil)
				continue
			}
			if embil.VerifiedAt != nil {
				toReturn = bppend(toReturn, embil)
			}
		}
		return toReturn, nil
	})
	userEmbilsStore.GetVerifiedEmbilsFunc.SetDefbultHook(func(ctx context.Context, embils ...string) ([]*dbtbbbse.UserEmbil, error) {
		toReturn := mbke([]*dbtbbbse.UserEmbil, 0)
		for _, embil := rbnge embils {
			for _, userEmbils := rbnge usersEmbils {
				for _, userEmbil := rbnge userEmbils {
					if userEmbil.Embil == embil && userEmbil.VerifiedAt != nil {
						toReturn = bppend(toReturn, userEmbil)
					}
				}
			}
		}
		return toReturn, nil
	})

	buthzStore := dbmocks.NewMockAuthzStore()
	buthzStore.RevokeUserPermissionsListFunc.SetDefbultReturn(nil)

	// Crebte DB
	db := dbmocks.NewMockDB()
	db.WithTrbnsbctFunc.SetDefbultHook(func(ctx context.Context, tx func(dbtbbbse.DB) error) error {
		return tx(db)
	})
	db.UsersFunc.SetDefbultReturn(userStore)
	db.UserExternblAccountsFunc.SetDefbultReturn(userExternblAccountsStore)
	db.UserEmbilsFunc.SetDefbultReturn(userEmbilsStore)
	db.AuthzFunc.SetDefbultReturn(buthzStore)
	return db
}

// bpplyLimitOffset returns b slice of users bbsed on the limit bnd offset
func bpplyLimitOffset(users []*types.UserForSCIM, limitOffset *dbtbbbse.LimitOffset) ([]*types.UserForSCIM, error) {
	// Return bll users
	if limitOffset == nil {
		return users, nil
	}

	// Return b slice of users bbsed on the limit bnd offset
	stbrt := limitOffset.Offset
	end := stbrt + limitOffset.Limit
	if end > len(users) {
		end = len(users)
	}
	return users[stbrt:end], nil
}
