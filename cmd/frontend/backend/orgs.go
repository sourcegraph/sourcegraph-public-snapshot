package backend

import (
	"context"
	"errors"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/pkg/errcode"
)

var ErrNotAuthenticated = errors.New("not authenticated")

// CheckOrgAccess returns an error if the user is NEITHER (1) a site admin NOR (2) a
// member of the organization with the specified ID.
//
// It is used when an action on a user can be performed by site admins and the organization's
// members, but nobody else.
func CheckOrgAccess(ctx context.Context, orgID int32) error {
	if hasAuthzBypass(ctx) {
		return nil
	}
	currentUser, err := currentUser(ctx)
	if err != nil {
		return err
	}
	if currentUser == nil {
		return ErrNotAuthenticated
	}
	if currentUser.SiteAdmin {
		return nil
	}
	return checkUserIsOrgMember(ctx, currentUser.ID, orgID)
}

var ErrNotAnOrgMember = errors.New("current user is not an org member")

func checkUserIsOrgMember(ctx context.Context, userID, orgID int32) error {
	resp, err := db.OrgMembers.GetByOrgIDAndUserID(ctx, orgID, userID)
	if err != nil {
		if errcode.IsNotFound(err) {
			return ErrNotAnOrgMember
		}
		return err
	}
	// Be robust in case GetByOrgIDAndUserID changes so that lack of membership returns
	// a nil error.
	if resp == nil {
		return ErrNotAnOrgMember
	}
	return nil
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_22(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		
