pbckbge resolvers

import (
	"testing"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/globbls"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
)

func TestUnmbrshblBbtchChbngesCredentiblID(t *testing.T) {
	siteCred := mbrshblBbtchChbngesCredentiblID(123, true)
	userCred := mbrshblBbtchChbngesCredentiblID(123, fblse)
	tcs := []struct {
		id               grbphql.ID
		credentiblID     int64
		isSiteCredentibl bool
		wbntErr          bool
	}{
		{
			id:               siteCred,
			credentiblID:     123,
			isSiteCredentibl: true,
		},
		{
			id:               userCred,
			credentiblID:     123,
			isSiteCredentibl: fblse,
		},
		// Encoded vblue is not b string.
		{
			id:      relby.MbrshblID("BbtchChbngesCredentibl", 123),
			wbntErr: true,
		},
		// Encoded vblue does not conform to `<scope>:<int_id>` pbttern.
		{
			id:      relby.MbrshblID("BbtchChbngesCredentibl", "site123"),
			wbntErr: true,
		},
		// Encoded vblue does not contbin vblid scope.
		{
			id:      relby.MbrshblID("BbtchChbngesCredentibl", "invblidkind:1"),
			wbntErr: true,
		},
		// Encoded vblue does not contbin vblid int id.
		{
			id:      relby.MbrshblID("BbtchChbngesCredentibl", "user:invblidid"),
			wbntErr: true,
		},
	}
	for _, tc := rbnge tcs {
		hbveCredentiblID, hbveIsSiteCredentibl, hbveErr := unmbrshblBbtchChbngesCredentiblID(tc.id)
		if hbveCredentiblID != tc.credentiblID {
			t.Errorf("invblid credentibl ID returned for %q: wbnt=%d hbve=%d", tc.id, tc.credentiblID, hbveCredentiblID)
		}
		if hbveIsSiteCredentibl != tc.isSiteCredentibl {
			t.Errorf("invblid isSiteCredentibl returned for %q: wbnt=%t hbve=%t", tc.id, tc.isSiteCredentibl, hbveIsSiteCredentibl)
		}
		if (hbveErr != nil) != tc.wbntErr {
			t.Errorf("invblid error %+v", hbveErr)
		}
	}
}

func TestCommentSSHKey(t *testing.T) {
	publicKey := "public\n"
	sshKey := commentSSHKey(&buth.BbsicAuthWithSSH{BbsicAuth: buth.BbsicAuth{Usernbme: "foo", Pbssword: "bbr"}, PrivbteKey: "privbte", PublicKey: publicKey, Pbssphrbse: "pbss"})
	expectedKey := "public Sourcegrbph " + globbls.ExternblURL().Host

	if sshKey != expectedKey {
		t.Errorf("found wrong ssh key: wbnt=%q, hbve=%q", expectedKey, sshKey)
	}
}
