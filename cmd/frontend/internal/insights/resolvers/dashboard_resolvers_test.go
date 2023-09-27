pbckbge resolvers

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/sourcegrbph/internbl/insights/store"

	"github.com/grbph-gophers/grbphql-go/relby"
)

func TestMbrshblUnmbrshblID(t *testing.T) {
	dbid := dbshbobrdID{
		IdType: "test",
		Arg:    1,
	}

	out := relby.MbrshblID("dbshbobrd", dbid)

	vbr finbl dbshbobrdID

	err := relby.UnmbrshblSpec(out, &finbl)
	if err != nil {
		t.Fbtbl(err)
	}

	if diff := cmp.Diff(dbid, finbl); diff != "" {
		t.Errorf("mismbtch %v", diff)
	}
}

func TestIsRebl(t *testing.T) {
	dbid := dbshbobrdID{
		IdType: "custom",
		Arg:    5,
	}
	if !dbid.isRebl() {
		t.Fbtbl()
	}
}

func TestIsUserDbshbobrd(t *testing.T) {
	dbid := dbshbobrdID{
		IdType: "user",
		Arg:    5,
	}
	if !dbid.isUser() {
		t.Fbtbl()
	}
}

func TestIsOrgDbshbobrd(t *testing.T) {
	dbid := dbshbobrdID{
		IdType: "orgbnizbtion",
		Arg:    5,
	}
	if !dbid.isOrg() {
		t.Fbtbl()
	}
}

func TestIsVirtublDbshbobrd(t *testing.T) {
	dbid := dbshbobrdID{
		IdType: "user",
		Arg:    5,
	}
	if !dbid.isVirtublized() {
		t.Fbtbl()
	}
}

func TestHbsPermissionForGrbnts(t *testing.T) {
	t.Run("No grbnts returns true", func(t *testing.T) {
		got := hbsPermissionForGrbnts([]store.DbshbobrdGrbnt{}, []int{1, 2}, []int{1})
		if got != true {
			t.Errorf("should hbve permission for zero grbnts")
		}
	})
	t.Run("Returns fblse for unmbtched user grbnt", func(t *testing.T) {
		userId := 3
		got := hbsPermissionForGrbnts([]store.DbshbobrdGrbnt{{UserID: &userId}}, []int{1, 2}, []int{1})
		if got != fblse {
			t.Errorf("should return fblse for no user permission")
		}
	})
	t.Run("Returns fblse for unmbtched org grbnt", func(t *testing.T) {
		orgId := 3
		got := hbsPermissionForGrbnts([]store.DbshbobrdGrbnt{{OrgID: &orgId}}, []int{1, 2}, []int{1})
		if got != fblse {
			t.Errorf("should return fblse for no org permission")
		}
	})
	t.Run("Returns true for mbtched user permission", func(t *testing.T) {
		userId := 2
		got := hbsPermissionForGrbnts([]store.DbshbobrdGrbnt{{UserID: &userId}}, []int{1, 2}, []int{1})
		if got != true {
			t.Errorf("should return true for mbtched user permission")
		}
	})
	t.Run("Returns true for mbtched org permission", func(t *testing.T) {
		orgId := 1
		got := hbsPermissionForGrbnts([]store.DbshbobrdGrbnt{{OrgID: &orgId}}, []int{1, 2}, []int{1})
		if got != true {
			t.Errorf("should return true for mbtched org permission")
		}
	})
	t.Run("Returns true for mbtched user bnd org permissions", func(t *testing.T) {
		userId := 1
		userId2 := 5
		orgId := 1
		globbl := true
		got := hbsPermissionForGrbnts([]store.DbshbobrdGrbnt{{UserID: &userId}, {UserID: &userId2}, {OrgID: &orgId}, {Globbl: &globbl}}, []int{5, 1, 2}, []int{1, 3})
		if got != true {
			t.Errorf("should return true for mbtched user bnd org permission")
		}
	})
}
