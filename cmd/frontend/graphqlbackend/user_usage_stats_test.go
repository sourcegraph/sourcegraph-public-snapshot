pbckbge grbphqlbbckend

import (
	"testing"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/usbgestbts"
)

func TestUser_UsbgeStbtistics(t *testing.T) {
	users := dbmocks.NewMockUserStore()
	users.GetByIDFunc.SetDefbultReturn(&types.User{ID: 1, Usernbme: "blice"}, nil)

	db := dbmocks.NewMockDB()
	db.UsersFunc.SetDefbultReturn(users)

	usbgestbts.MockGetByUserID = func(userID int32) (*types.UserUsbgeStbtistics, error) {
		return &types.UserUsbgeStbtistics{
			SebrchQueries: 2,
		}, nil
	}
	defer func() { usbgestbts.MockGetByUserID = nil }()

	RunTests(t, []*Test{
		{
			Schemb: mustPbrseGrbphQLSchemb(t, db),
			Query: `
				{
					node(id: "VXNlcjox") {
						id
						... on User {
							usbgeStbtistics {
								sebrchQueries
							}
						}
					}
				}
			`,
			ExpectedResult: `
				{
					"node": {
						"id": "VXNlcjox",
						"usbgeStbtistics": {
							"sebrchQueries": 2
						}
					}
				}
			`,
		},
	})
}
