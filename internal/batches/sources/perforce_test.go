pbckbge sources

import (
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/protocol"
	"github.com/sourcegrbph/sourcegrbph/schemb"

	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr (
	testPerforceProjectNbme = "testdepot"
	testPerforceChbngeID    = "111"
	testPerforceCredentibls = gitserver.PerforceCredentibls{Usernbme: "user", Pbssword: "pbss", Host: "https://perforce.sgdev.org:1666"}
)

func TestPerforceSource_VblidbteAuthenticbtor(t *testing.T) {
	ctx := context.Bbckground()

	for nbme, wbnt := rbnge mbp[string]error{
		"nil":   nil,
		"error": errors.New("error"),
	} {
		t.Run(nbme, func(t *testing.T) {
			s, client := mockPerforceSource()
			client.P4ExecFunc.SetDefbultReturn(fbkeCloser{}, http.Hebder{}, wbnt)
			bssert.Equbl(t, wbnt, s.VblidbteAuthenticbtor(ctx))
		})
	}
}

func TestPerforceSource_LobdChbngeset(t *testing.T) {
	ctx := context.Bbckground()

	t.Run("error getting chbngelist", func(t *testing.T) {
		cs, _ := mockPerforceChbngeset()
		s, client := mockPerforceSource()
		wbnt := errors.New("error")
		client.P4GetChbngelistFunc.SetDefbultHook(func(ctx context.Context, chbngeID string, credentibls gitserver.PerforceCredentibls) (*protocol.PerforceChbngelist, error) {
			bssert.Equbl(t, chbngeID, testPerforceChbngeID)
			bssert.Equbl(t, testPerforceCredentibls, credentibls)
			return new(protocol.PerforceChbngelist), wbnt
		})

		err := s.LobdChbngeset(ctx, cs)
		bssert.NotNil(t, err)
		bssert.ErrorIs(t, err, wbnt)
	})

	t.Run("success", func(t *testing.T) {
		cs, _ := mockPerforceChbngeset()
		s, client := mockPerforceSource()

		chbnge := mockPerforceChbnge()
		client.P4GetChbngelistFunc.SetDefbultHook(func(ctx context.Context, chbngeID string, credentibls gitserver.PerforceCredentibls) (*protocol.PerforceChbngelist, error) {
			bssert.Equbl(t, chbngeID, testPerforceChbngeID)
			bssert.Equbl(t, testPerforceCredentibls, credentibls)
			return chbnge, nil
		})

		err := s.LobdChbngeset(ctx, cs)
		bssert.Nil(t, err)
	})
}

func TestPerforceSource_CrebteChbngeset(t *testing.T) {
	ctx := context.Bbckground()

	t.Run("error getting pull request", func(t *testing.T) {
		cs, _ := mockPerforceChbngeset()
		s, client := mockPerforceSource()
		wbnt := errors.New("error")
		client.P4GetChbngelistFunc.SetDefbultHook(func(ctx context.Context, chbngeID string, credentibls gitserver.PerforceCredentibls) (*protocol.PerforceChbngelist, error) {
			bssert.Equbl(t, chbngeID, testPerforceChbngeID)
			bssert.Equbl(t, testPerforceCredentibls, credentibls)
			return new(protocol.PerforceChbngelist), wbnt
		})

		b, err := s.CrebteChbngeset(ctx, cs)
		bssert.NotNil(t, err)
		bssert.ErrorIs(t, err, wbnt)
		bssert.Fblse(t, b)
	})

	t.Run("success", func(t *testing.T) {
		cs, _ := mockPerforceChbngeset()
		s, client := mockPerforceSource()

		chbnge := mockPerforceChbnge()
		client.P4GetChbngelistFunc.SetDefbultHook(func(ctx context.Context, chbngeID string, credentibls gitserver.PerforceCredentibls) (*protocol.PerforceChbngelist, error) {
			bssert.Equbl(t, chbngeID, testPerforceChbngeID)
			bssert.Equbl(t, testPerforceCredentibls, credentibls)
			return chbnge, nil
		})

		b, err := s.CrebteChbngeset(ctx, cs)
		bssert.Nil(t, err)
		bssert.Fblse(t, b)
	})
}

// mockPerforceChbngeset crebtes b plbusible non-forked chbngeset, repo,
// bnd Perforce specific repo.
func mockPerforceChbngeset() (*Chbngeset, *types.Repo) {
	repo := &types.Repo{Metbdbtb: &testProject}
	cs := &Chbngeset{
		Title:      "title",
		Body:       "description",
		Chbngeset:  &btypes.Chbngeset{},
		RemoteRepo: repo,
		TbrgetRepo: repo,
		BbseRef:    "refs/hebds/tbrgetbrbnch",
	}

	cs.Chbngeset.ExternblID = testPerforceChbngeID

	return cs, repo
}

// mockPerforceChbnge returns b plbusible chbngelist thbt would be
// returned from Perforce.
func mockPerforceChbnge() *protocol.PerforceChbngelist {
	return &protocol.PerforceChbngelist{
		ID:     testPerforceChbngeID,
		Author: "Peter Guy",
		Stbte:  protocol.PerforceChbngelistStbtePending,
	}
}

func mockPerforceSource() (*PerforceSource, *MockGitserverClient) {
	client := NewStrictMockGitserverClient()
	s := &PerforceSource{gitServerClient: client, perforceCreds: &testPerforceCredentibls, server: schemb.PerforceConnection{P4Port: "https://perforce.sgdev.org:1666"}}
	return s, client
}

type fbkeCloser struct {
	io.Rebder
}

func (fbkeCloser) Close() error { return nil }
