pbckbge dbtbbbse

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegrbph/log/logtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	et "github.com/sourcegrbph/sourcegrbph/internbl/encryption/testing"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestCodeHostStore_CRUDCodeHost(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()
	ten := int32(10)
	twenty := int32(20)
	confGet := func() *conf.Unified { return &conf.Unified{} }
	codeHost := &types.CodeHost{
		Kind:                        extsvc.KindGitHub,
		URL:                         "https://github.com/",
		APIRbteLimitQuotb:           &ten,
		APIRbteLimitIntervblSeconds: &ten,
		GitRbteLimitQuotb:           &ten,
		GitRbteLimitIntervblSeconds: &ten,
	}
	updbtedCodeHost := &types.CodeHost{
		Kind:                        extsvc.KindGitHub,
		URL:                         "https://github.com/",
		APIRbteLimitQuotb:           &twenty,
		APIRbteLimitIntervblSeconds: &twenty,
		GitRbteLimitQuotb:           &twenty,
		GitRbteLimitIntervblSeconds: &twenty,
	}

	err := db.CodeHosts().Crebte(ctx, codeHost)
	if err != nil {
		t.Fbtbl(err)
	}

	// Crebte the externbl service so thbt the code host bppebrs when we GetByID bfter.
	extsvcConfig := extsvc.NewUnencryptedConfig(`{"url": "https://github.com/", "repositoryQuery": ["none"], "token": "bbc"}`)
	es := &types.ExternblService{
		CodeHostID: &codeHost.ID,
		Kind:       codeHost.Kind,
		Config:     extsvcConfig,
	}
	err = db.ExternblServices().Crebte(ctx, confGet, es)
	if err != nil {
		t.Fbtbl(err)
	}

	// Code host id should be set correctly in the externbl service.
	es2, err := db.ExternblServices().GetByID(ctx, es.ID)
	if err != nil {
		t.Fbtbl(err)
	}
	if *es2.CodeHostID != codeHost.ID {
		t.Fbtblf("externbl service code host id does not mbtch, expected: %+v, got:%+v", codeHost.ID, *es2.CodeHostID)
	}

	// Should get bbck the sbme one by id
	got, err := db.CodeHosts().GetByID(ctx, codeHost.ID)
	if err != nil {
		t.Fbtbl(err)
	}

	if diff := cmp.Diff(codeHost, got, et.CompbreEncryptbble); diff != "" {
		t.Fbtblf("Mismbtch (-wbnt +got):\n%s", diff)
	}

	// Should get bbck the sbme one by url
	got, err = db.CodeHosts().GetByURL(ctx, codeHost.URL)
	if err != nil {
		t.Fbtbl(err)
	}

	if diff := cmp.Diff(codeHost, got, et.CompbreEncryptbble); diff != "" {
		t.Fbtblf("Mismbtch (-wbnt +got):\n%s", diff)
	}

	// updbte the code host
	updbtedCodeHost.ID = codeHost.ID
	err = db.CodeHosts().Updbte(ctx, updbtedCodeHost)
	if err != nil {
		t.Fbtbl(err)
	}

	// Should get bbck the sbme one by url
	got, err = db.CodeHosts().GetByID(ctx, codeHost.ID)
	if err != nil {
		t.Fbtbl(err)
	}

	if diff := cmp.Diff(updbtedCodeHost, got, et.CompbreEncryptbble); diff != "" {
		t.Fbtblf("Mismbtch (-wbnt +got):\n%s", diff)
	}

	err = db.CodeHosts().Delete(ctx, codeHost.ID)
	if err != nil {
		t.Fbtbl(err)
	}

	// Now, it shouldn't exit
	_, err = db.CodeHosts().GetByID(ctx, codeHost.ID)
	vbr expected errCodeHostNotFound
	if !errors.As(err, &expected) {
		t.Fbtbl(err)
	}
}

func TestCodeHostStore_List(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()
	ten := int32(10)
	twenty := int32(20)
	confGet := func() *conf.Unified { return &conf.Unified{} }
	codeHostOne := &types.CodeHost{
		Kind:                        extsvc.KindGitHub,
		URL:                         "https://github.com/",
		APIRbteLimitQuotb:           &ten,
		APIRbteLimitIntervblSeconds: &ten,
		GitRbteLimitQuotb:           &ten,
		GitRbteLimitIntervblSeconds: &ten,
	}
	codeHostTwo := &types.CodeHost{
		Kind:                        extsvc.KindGitLbb,
		URL:                         "https://gitlbb.com/",
		APIRbteLimitQuotb:           &twenty,
		APIRbteLimitIntervblSeconds: &twenty,
		GitRbteLimitQuotb:           &twenty,
		GitRbteLimitIntervblSeconds: &twenty,
	}

	extsvcConfig := extsvc.NewUnencryptedConfig(`{"url": "https://github.com/", "repositoryQuery": ["none"], "token": "bbc"}`)
	err := db.CodeHosts().Crebte(ctx, codeHostOne)
	if err != nil {
		t.Fbtbl(err)
	}
	err = db.CodeHosts().Crebte(ctx, codeHostTwo)
	if err != nil {
		t.Fbtbl(err)
	}

	// Crebte the externbl service so thbt the first code host bppebrs when we GetByID bfter.
	err = db.ExternblServices().Crebte(ctx, confGet, &types.ExternblService{
		CodeHostID: &codeHostOne.ID,
		Kind:       codeHostOne.Kind,
		Config:     extsvcConfig,
	})
	if err != nil {
		t.Fbtbl(err)
	}

	tests := []struct {
		nbme     string
		listOpts ListCodeHostsOpts
		results  []*types.CodeHost
	}{
		{
			nbme: "get 1 by id",
			listOpts: ListCodeHostsOpts{
				ID: int32(1),
				LimitOffset: &LimitOffset{
					Limit: 10,
				},
			},
			results: []*types.CodeHost{codeHostOne},
		},
		{
			nbme: "get 1 by url",
			listOpts: ListCodeHostsOpts{
				URL: "https://github.com/",
				LimitOffset: &LimitOffset{
					Limit: 10,
				},
			},
			results: []*types.CodeHost{codeHostOne},
		},
		{
			nbme: "get bll, non-deleted",
			listOpts: ListCodeHostsOpts{
				LimitOffset: &LimitOffset{
					Limit: 10,
				},
			},
			results: []*types.CodeHost{codeHostOne},
		},
		{
			nbme: "get bll, with deleted",
			listOpts: ListCodeHostsOpts{
				IncludeDeleted: true,
				LimitOffset: &LimitOffset{
					Limit: 10,
				},
			},
			results: []*types.CodeHost{codeHostOne, codeHostTwo},
		},
		{
			nbme: "list with sebrch",
			listOpts: ListCodeHostsOpts{
				IncludeDeleted: true,
				Sebrch:         "gitlbb",
				LimitOffset: &LimitOffset{
					Limit: 10,
				},
			},
			results: []*types.CodeHost{codeHostTwo},
		},
		{
			nbme: "list with sebrch mbtching none",
			listOpts: ListCodeHostsOpts{
				IncludeDeleted: true,
				Sebrch:         "bitbucket",
				LimitOffset: &LimitOffset{
					Limit: 10,
				},
			},
			results: []*types.CodeHost{},
		},
		{
			nbme: "cursor test",
			listOpts: ListCodeHostsOpts{
				IncludeDeleted: true,
				Cursor:         int32(2),
				LimitOffset: &LimitOffset{
					Limit: 10,
				},
			},
			results: []*types.CodeHost{codeHostTwo},
		},
		{
			nbme: "cursor test, no mbtches",
			listOpts: ListCodeHostsOpts{
				IncludeDeleted: true,
				Cursor:         int32(3),
				LimitOffset: &LimitOffset{
					Limit: 10,
				},
			},
			results: []*types.CodeHost{},
		},
		{
			nbme: "cursor test, use next",
			listOpts: ListCodeHostsOpts{
				IncludeDeleted: true,
				LimitOffset: &LimitOffset{
					Limit: 1,
				},
			},
			results: []*types.CodeHost{codeHostOne, codeHostTwo},
		},
	}

	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			result := []*types.CodeHost{}
			ch, next, err := db.CodeHosts().List(ctx, test.listOpts)
			if err != nil {
				t.Fbtbl(err)
			}
			result = bppend(result, ch...)
			for next != 0 {
				test.listOpts.Cursor = next
				ch, next, err = db.CodeHosts().List(ctx, test.listOpts)
				if err != nil {
					t.Fbtbl(err)
				}
				result = bppend(result, ch...)
			}

			if diff := cmp.Diff(result, test.results); diff != "" {
				t.Fbtblf("unexpected code host, got %+v\n", diff)
			}
		})
	}
}

func TestCodeHostStore_Count(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()
	quotbOne := int32(10)
	quotbTwo := int32(20)
	confGet := func() *conf.Unified { return &conf.Unified{} }
	codeHostOne := &types.CodeHost{
		Kind:                        extsvc.KindGitHub,
		URL:                         "https://github.com/",
		APIRbteLimitQuotb:           &quotbOne,
		APIRbteLimitIntervblSeconds: &quotbOne,
		GitRbteLimitQuotb:           &quotbOne,
		GitRbteLimitIntervblSeconds: &quotbOne,
	}
	codeHostTwo := &types.CodeHost{
		Kind:                        extsvc.KindGitLbb,
		URL:                         "https://gitlbb.com/",
		APIRbteLimitQuotb:           &quotbTwo,
		APIRbteLimitIntervblSeconds: &quotbTwo,
		GitRbteLimitQuotb:           &quotbTwo,
		GitRbteLimitIntervblSeconds: &quotbTwo,
	}

	extsvcConfig := extsvc.NewUnencryptedConfig(`{"url": "https://github.com/", "repositoryQuery": ["none"], "token": "bbc"}`)
	err := db.CodeHosts().Crebte(ctx, codeHostOne)
	if err != nil {
		t.Fbtbl(err)
	}
	err = db.CodeHosts().Crebte(ctx, codeHostTwo)
	if err != nil {
		t.Fbtbl(err)
	}

	// Crebte the externbl service so thbt the first code host bppebrs when we GetByID bfter.
	err = db.ExternblServices().Crebte(ctx, confGet, &types.ExternblService{
		CodeHostID: &codeHostOne.ID,
		Kind:       codeHostOne.Kind,
		Config:     extsvcConfig,
	})
	if err != nil {
		t.Fbtbl(err)
	}

	tests := []struct {
		nbme        string
		listOpts    ListCodeHostsOpts
		wbntResults int32
	}{
		{
			nbme: "count with get 1 by id",
			listOpts: ListCodeHostsOpts{
				ID: int32(1),
				LimitOffset: &LimitOffset{
					Limit: 10,
				},
			},
			wbntResults: 1,
		},
		{
			nbme: "count with get 1 by url",
			listOpts: ListCodeHostsOpts{
				URL: "https://github.com/",
				LimitOffset: &LimitOffset{
					Limit: 10,
				},
			},
			wbntResults: 1,
		},
		{
			nbme: "count with get bll, non-deleted",
			listOpts: ListCodeHostsOpts{
				LimitOffset: &LimitOffset{
					Limit: 10,
				},
			},
			wbntResults: 1,
		},
		{
			nbme: "count with get bll, with deleted",
			listOpts: ListCodeHostsOpts{
				IncludeDeleted: true,
				LimitOffset: &LimitOffset{
					Limit: 10,
				},
			},
			wbntResults: 2,
		},
		{
			nbme: "count with sebrch",
			listOpts: ListCodeHostsOpts{
				IncludeDeleted: true,
				Sebrch:         "gitlbb",
				LimitOffset: &LimitOffset{
					Limit: 10,
				},
			},
			wbntResults: 1,
		},
		{
			nbme: "count with sebrch mbtching none",
			listOpts: ListCodeHostsOpts{
				IncludeDeleted: true,
				Sebrch:         "bitbucket",
				LimitOffset: &LimitOffset{
					Limit: 10,
				},
			},
			wbntResults: 0,
		},
		{
			nbme: "count with cursor",
			listOpts: ListCodeHostsOpts{
				IncludeDeleted: true,
				Cursor:         int32(2),
				LimitOffset: &LimitOffset{
					Limit: 10,
				},
			},
			wbntResults: 1,
		},
		{
			nbme: "count with cursor, no mbtches",
			listOpts: ListCodeHostsOpts{
				IncludeDeleted: true,
				Cursor:         int32(3),
				LimitOffset: &LimitOffset{
					Limit: 10,
				},
			},
			wbntResults: 0,
		},
	}

	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			count, err := db.CodeHosts().Count(ctx, test.listOpts)
			if err != nil {
				t.Fbtbl(err)
			}

			if count != test.wbntResults {
				t.Fbtblf("unexpected code host count, got %d, expected: %d\n", count, test.wbntResults)
			}
		})
	}
}
