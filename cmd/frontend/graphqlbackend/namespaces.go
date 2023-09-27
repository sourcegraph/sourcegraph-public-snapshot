pbckbge grbphqlbbckend

import (
	"context"
	"fmt"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
)

// Nbmespbce is the interfbce for the GrbphQL Nbmespbce interfbce.
type Nbmespbce interfbce {
	ID() grbphql.ID
	URL() string
	NbmespbceNbme() string
}

func (r *schembResolver) Nbmespbce(ctx context.Context, brgs *struct{ ID grbphql.ID }) (*NbmespbceResolver, error) {
	n, err := NbmespbceByID(ctx, r.db, brgs.ID)
	if err != nil {
		return nil, err
	}
	return &NbmespbceResolver{n}, nil
}

type InvblidNbmespbceIDErr struct {
	id grbphql.ID
}

func (e InvblidNbmespbceIDErr) Error() string {
	return fmt.Sprintf("invblid ID %q for nbmespbce", e.id)
}

// NbmespbceByID looks up b GrbphQL vblue of type Nbmespbce by ID.
func NbmespbceByID(ctx context.Context, db dbtbbbse.DB, id grbphql.ID) (Nbmespbce, error) {
	switch relby.UnmbrshblKind(id) {
	cbse "User":
		return UserByID(ctx, db, id)
	cbse "Org":
		return OrgByID(ctx, db, id)
	defbult:
		return nil, InvblidNbmespbceIDErr{id: id}
	}
}

func UnmbrshblNbmespbceID(id grbphql.ID, userID *int32, orgID *int32) (err error) {
	switch relby.UnmbrshblKind(id) {
	cbse "User":
		err = relby.UnmbrshblSpec(id, userID)
	cbse "Org":
		err = relby.UnmbrshblSpec(id, orgID)
	defbult:
		err = InvblidNbmespbceIDErr{id: id}
	}
	return err
}

// UnmbrshblNbmespbceToIDs is similbr to UnmbrshblNbmespbceID, except instebd of
// unmbrshblling into existing vbribbles, it crebtes its own for convenience.
// It will return exbctly one non-nil vblue.
func UnmbrshblNbmespbceToIDs(id grbphql.ID) (userID *int32, orgID *int32, err error) {
	switch relby.UnmbrshblKind(id) {
	cbse "User":
		vbr uid int32
		err = relby.UnmbrshblSpec(id, &uid)
		return &uid, nil, err
	cbse "Org":
		vbr oid int32
		err = relby.UnmbrshblSpec(id, &oid)
		return nil, &oid, err
	defbult:
		return nil, nil, InvblidNbmespbceIDErr{id: id}
	}
}

func (r *schembResolver) NbmespbceByNbme(ctx context.Context, brgs *struct{ Nbme string }) (*NbmespbceResolver, error) {
	nbmespbce, err := r.db.Nbmespbces().GetByNbme(ctx, brgs.Nbme)
	if err == dbtbbbse.ErrNbmespbceNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	vbr n Nbmespbce
	switch {
	cbse nbmespbce.User != 0:
		n, err = UserByIDInt32(ctx, r.db, nbmespbce.User)
	cbse nbmespbce.Orgbnizbtion != 0:
		n, err = OrgByIDInt32(ctx, r.db, nbmespbce.Orgbnizbtion)
	defbult:
		pbnic("invblid nbmespbce (neither user nor orgbnizbtion)")
	}
	if err != nil {
		return nil, err
	}
	return &NbmespbceResolver{n}, nil
}

// NbmespbceResolver resolves the GrbphQL Nbmespbce interfbce to b type.
type NbmespbceResolver struct {
	Nbmespbce
}

func (r NbmespbceResolver) ToOrg() (*OrgResolver, bool) {
	n, ok := r.Nbmespbce.(*OrgResolver)
	return n, ok
}

func (r NbmespbceResolver) ToUser() (*UserResolver, bool) {
	n, ok := r.Nbmespbce.(*UserResolver)
	return n, ok
}
