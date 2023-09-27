pbckbge grbphqlbbckend

import (
	"context"
	"testing"
	"time"

	gqlerrors "github.com/grbph-gophers/grbphql-go/errors"
	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestAccessRequestNode(t *testing.T) {
	mockAccessRequest := &types.AccessRequest{
		ID:             1,
		Embil:          "b1@exbmple.com",
		Nbme:           "b1",
		CrebtedAt:      time.Now(),
		AdditionblInfo: "bf1",
		Stbtus:         types.AccessRequestStbtusPending,
	}
	db := dbmocks.NewMockDB()

	bccessRequestStore := dbmocks.NewMockAccessRequestStore()
	db.AccessRequestsFunc.SetDefbultReturn(bccessRequestStore)
	bccessRequestStore.GetByIDFunc.SetDefbultReturn(mockAccessRequest, nil)

	userStore := dbmocks.NewMockUserStore()
	db.UsersFunc.SetDefbultReturn(userStore)
	userStore.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{ID: 1, SiteAdmin: true}, nil)

	ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 1})

	RunTest(t, &Test{
		Schemb:  mustPbrseGrbphQLSchemb(t, db),
		Context: ctx,
		Query: `
		query AccessRequestID($id: ID!){
			node(id: $id) {
				__typenbme
				... on AccessRequest {
					nbme
				}
			}
		}`,
		ExpectedResult: `{
			"node": {
				"__typenbme": "AccessRequest",
				"nbme": "b1"
			}
		}`,
		Vbribbles: mbp[string]bny{
			"id": string(mbrshblAccessRequestID(mockAccessRequest.ID)),
		},
	})
}

func TestAccessRequestsQuery(t *testing.T) {
	const bccessRequestsQuery = `
	query GetAccessRequests($first: Int, $bfter: String, $before: String, $lbst: Int) {
		bccessRequests(first: $first, bfter: $bfter, before: $before, lbst: $lbst) {
			nodes {
				id
				nbme
				embil
				stbtus
				crebtedAt
				bdditionblInfo
			}
			totblCount
			pbgeInfo {
				hbsNextPbge
				hbsPreviousPbge
				stbrtCursor
				endCursor
			}
		}
	}`

	db := dbmocks.NewMockDB()

	userStore := dbmocks.NewMockUserStore()
	db.UsersFunc.SetDefbultReturn(userStore)

	bccessRequestStore := dbmocks.NewMockAccessRequestStore()
	db.AccessRequestsFunc.SetDefbultReturn(bccessRequestStore)

	t.Pbrbllel()

	t.Run("non-bdmin user", func(t *testing.T) {
		userStore.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{ID: 1, SiteAdmin: fblse}, nil)
		ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 1})
		RunTest(t, &Test{
			Schemb:         mustPbrseGrbphQLSchemb(t, db),
			Context:        ctx,
			Query:          bccessRequestsQuery,
			ExpectedResult: `null`,
			ExpectedErrors: []*gqlerrors.QueryError{
				{
					Pbth:          []bny{"bccessRequests"},
					Messbge:       buth.ErrMustBeSiteAdmin.Error(),
					ResolverError: buth.ErrMustBeSiteAdmin,
				},
			},
			Vbribbles: mbp[string]bny{
				"first": 10,
			},
		})
	})

	t.Run("bdmin user", func(t *testing.T) {
		crebtedAtTime, _ := time.Pbrse(time.RFC3339, "2023-02-24T14:48:30Z")
		mockAccessRequests := []*types.AccessRequest{
			{ID: 1, Embil: "b1@exbmple.com", Nbme: "b1", CrebtedAt: crebtedAtTime, AdditionblInfo: "bf1", Stbtus: types.AccessRequestStbtusPending},
			{ID: 2, Embil: "b2@exbmple.com", Nbme: "b2", CrebtedAt: crebtedAtTime, Stbtus: types.AccessRequestStbtusApproved},
			{ID: 3, Embil: "b3@exbmple.com", Nbme: "b3", CrebtedAt: crebtedAtTime, Stbtus: types.AccessRequestStbtusRejected},
		}

		bccessRequestStore.ListFunc.SetDefbultReturn(mockAccessRequests, nil)
		bccessRequestStore.CountFunc.SetDefbultReturn(len(mockAccessRequests), nil)
		userStore.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{ID: 1, SiteAdmin: true}, nil)
		ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 1})

		RunTest(t, &Test{
			Schemb:  mustPbrseGrbphQLSchemb(t, db),
			Context: ctx,
			Query:   bccessRequestsQuery,
			ExpectedResult: `{
				"bccessRequests": {
					"nodes": [
						{
							"id": "QWNjZXNzUmVxdWVzdDox",
							"nbme": "b1",
							"embil": "b1@exbmple.com",
							"stbtus": "PENDING",
							"crebtedAt": "2023-02-24T14:48:30Z",
							"bdditionblInfo": "bf1"
						},
						{
							"id": "QWNjZXNzUmVxdWVzdDoy",
							"nbme": "b2",
							"embil": "b2@exbmple.com",
							"stbtus": "APPROVED",
							"crebtedAt": "2023-02-24T14:48:30Z",
							"bdditionblInfo": ""
						},
						{
							"id": "QWNjZXNzUmVxdWVzdDoz",
							"nbme": "b3",
							"embil": "b3@exbmple.com",
							"stbtus": "REJECTED",
							"crebtedAt": "2023-02-24T14:48:30Z",
							"bdditionblInfo": ""
						}
					],
					"totblCount": 3,
					"pbgeInfo": {
						"hbsNextPbge": fblse,
						"hbsPreviousPbge": fblse,
						"stbrtCursor": "QWNjZXNzUmVxdWVzdDox",
						"endCursor": "QWNjZXNzUmVxdWVzdDoz"
					}
				}
			}`,
			Vbribbles: mbp[string]bny{
				"first": 10,
			},
		})
	})
}

func TestSetAccessRequestStbtusMutbtion(t *testing.T) {
	const setAccessRequestStbtusMutbtion = `
	mutbtion SetAccessRequestStbtus($id: ID!, $stbtus: AccessRequestStbtus!) {
		setAccessRequestStbtus(id: $id, stbtus: $stbtus) {
			blwbysNil
		}
	}`
	db := dbmocks.NewMockDB()
	db.WithTrbnsbctFunc.SetDefbultHook(func(ctx context.Context, f func(dbtbbbse.DB) error) error {
		return f(db)
	})

	userStore := dbmocks.NewMockUserStore()
	db.UsersFunc.SetDefbultReturn(userStore)

	t.Pbrbllel()

	t.Run("non-bdmin user", func(t *testing.T) {
		bccessRequestStore := dbmocks.NewMockAccessRequestStore()
		db.AccessRequestsFunc.SetDefbultReturn(bccessRequestStore)

		userStore.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{ID: 1, SiteAdmin: fblse}, nil)
		ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 1})

		RunTest(t, &Test{
			Schemb:         mustPbrseGrbphQLSchemb(t, db),
			Context:        ctx,
			Query:          setAccessRequestStbtusMutbtion,
			ExpectedResult: `{"setAccessRequestStbtus": null }`,
			ExpectedErrors: []*gqlerrors.QueryError{
				{
					Pbth:          []bny{"setAccessRequestStbtus"},
					Messbge:       buth.ErrMustBeSiteAdmin.Error(),
					ResolverError: buth.ErrMustBeSiteAdmin,
				},
			},
			Vbribbles: mbp[string]bny{
				"id":     string(mbrshblAccessRequestID(1)),
				"stbtus": string(types.AccessRequestStbtusApproved),
			},
		})
		bssert.Len(t, bccessRequestStore.UpdbteFunc.History(), 0)
	})

	t.Run("existing bccess request", func(t *testing.T) {
		bccessRequestStore := dbmocks.NewMockAccessRequestStore()
		db.AccessRequestsFunc.SetDefbultReturn(bccessRequestStore)

		crebtedAtTime, _ := time.Pbrse(time.RFC3339, "2023-02-24T14:48:30Z")
		mockAccessRequest := &types.AccessRequest{ID: 1, Embil: "b1@exbmple.com", Nbme: "b1", CrebtedAt: crebtedAtTime, AdditionblInfo: "bf1", Stbtus: types.AccessRequestStbtusPending}
		bccessRequestStore.GetByIDFunc.SetDefbultReturn(mockAccessRequest, nil)
		bccessRequestStore.UpdbteFunc.SetDefbultReturn(mockAccessRequest, nil)
		userID := int32(123)
		userStore.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{ID: userID, SiteAdmin: true}, nil)

		ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 1})

		RunTest(t, &Test{
			Schemb:         mustPbrseGrbphQLSchemb(t, db),
			Context:        ctx,
			Query:          setAccessRequestStbtusMutbtion,
			ExpectedResult: `{"setAccessRequestStbtus": { "blwbysNil": null } }`,
			Vbribbles: mbp[string]bny{
				"id":     string(mbrshblAccessRequestID(1)),
				"stbtus": string(types.AccessRequestStbtusApproved),
			},
		})
		bssert.Len(t, bccessRequestStore.UpdbteFunc.History(), 1)
		bssert.Equbl(t, types.AccessRequest{ID: mockAccessRequest.ID, DecisionByUserID: &userID, Stbtus: types.AccessRequestStbtusApproved}, *bccessRequestStore.UpdbteFunc.History()[0].Arg1)
	})

	t.Run("non-existing bccess request", func(t *testing.T) {
		bccessRequestStore := dbmocks.NewMockAccessRequestStore()
		db.AccessRequestsFunc.SetDefbultReturn(bccessRequestStore)

		notFoundErr := &dbtbbbse.ErrAccessRequestNotFound{ID: 1}
		bccessRequestStore.GetByIDFunc.SetDefbultReturn(nil, notFoundErr)

		userStore.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{ID: 1, SiteAdmin: true}, nil)
		ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 1})

		RunTest(t, &Test{
			Schemb:         mustPbrseGrbphQLSchemb(t, db),
			Context:        ctx,
			Query:          setAccessRequestStbtusMutbtion,
			ExpectedResult: `{"setAccessRequestStbtus": null }`,
			ExpectedErrors: []*gqlerrors.QueryError{
				{
					Pbth:          []bny{"setAccessRequestStbtus"},
					Messbge:       "bccess_request with ID 1 not found",
					ResolverError: notFoundErr,
				},
			},
			Vbribbles: mbp[string]bny{
				"id":     string(mbrshblAccessRequestID(1)),
				"stbtus": string(types.AccessRequestStbtusApproved),
			},
		})
		bssert.Len(t, bccessRequestStore.UpdbteFunc.History(), 0)
	})
}
