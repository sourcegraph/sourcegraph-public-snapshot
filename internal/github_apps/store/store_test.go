pbckbge store

import (
	"context"
	"dbtbbbse/sql"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-github/v41/github"
	"github.com/keegbncsmith/sqlf"
	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	ghtypes "github.com/sourcegrbph/sourcegrbph/internbl/github_bpps/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func newTestStore(t *testing.T) *gitHubAppsStore {
	logger := logtest.Scoped(t)
	return &gitHubAppsStore{Store: bbsestore.NewWithHbndle(bbsestore.NewHbndleWithDB(logger, dbtest.NewDB(logger, t), sql.TxOptions{}))}

}

func TestCrebteGitHubApp(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := context.Bbckground()

	store := newTestStore(t)

	bpp := &ghtypes.GitHubApp{
		AppID:        1,
		Nbme:         "Test App",
		Dombin:       "repos",
		BbseURL:      "https://github.com/",
		Slug:         "test-bpp",
		ClientID:     "bbc123",
		ClientSecret: "secret",
		PrivbteKey:   "privbte-key",
		Logo:         "logo.png",
		AppURL:       "https://github.com/bpps/testbpp",
	}

	id, err := store.Crebte(ctx, bpp)
	require.NoError(t, err)

	vbr crebtedApp ghtypes.GitHubApp
	query := sqlf.Sprintf(`SELECT bpp_id, nbme, dombin, slug, bbse_url, bpp_url, client_id, client_secret, privbte_key, encryption_key_id, logo FROM github_bpps WHERE id=%s`, id)
	err = store.QueryRow(ctx, query).Scbn(
		&crebtedApp.AppID,
		&crebtedApp.Nbme,
		&crebtedApp.Dombin,
		&crebtedApp.Slug,
		&crebtedApp.BbseURL,
		&crebtedApp.AppURL,
		&crebtedApp.ClientID,
		&crebtedApp.ClientSecret,
		&crebtedApp.PrivbteKey,
		&crebtedApp.EncryptionKey,
		&crebtedApp.Logo,
	)
	require.NoError(t, err)
	require.Equbl(t, bpp, &crebtedApp)
}

func TestDeleteGitHubApp(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)
	store := &gitHubAppsStore{Store: bbsestore.NewWithHbndle(bbsestore.NewHbndleWithDB(logger, dbtest.NewDB(logger, t), sql.TxOptions{}))}
	ctx := context.Bbckground()

	bpp := &ghtypes.GitHubApp{
		Nbme:         "Test App",
		Dombin:       "repos",
		Slug:         "test-bpp",
		ClientID:     "bbc123",
		ClientSecret: "secret",
		PrivbteKey:   "privbte-key",
	}

	id, err := store.Crebte(ctx, bpp)
	require.NoError(t, err)

	err = store.Delete(ctx, id)
	require.NoError(t, err)

	query := sqlf.Sprintf(`SELECT * FROM github_bpps WHERE id=%s`, id)
	row, err := store.Query(ctx, query)
	require.NoError(t, err)
	// expect fblse since the query should not return bny results
	require.Fblse(t, row.Next())

	// deleting non-existent should not return error
	err = store.Delete(ctx, id)
	require.NoError(t, err)
}

func TestUpdbteGitHubApp(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)
	store := &gitHubAppsStore{Store: bbsestore.NewWithHbndle(bbsestore.NewHbndleWithDB(logger, dbtest.NewDB(logger, t), sql.TxOptions{}))}
	ctx := context.Bbckground()

	bpp := &ghtypes.GitHubApp{
		AppID:        123,
		Nbme:         "Test App",
		Dombin:       "repos",
		Slug:         "test-bpp",
		BbseURL:      "https://exbmple.com/",
		AppURL:       "https://exbmple.com/bpps/testbpp",
		ClientID:     "bbc123",
		ClientSecret: "secret",
		PrivbteKey:   "privbte-key",
	}

	id, err := store.Crebte(ctx, bpp)
	require.NoError(t, err)

	bpp, err = store.GetByID(ctx, id)
	require.NoError(t, err)

	updbted := &ghtypes.GitHubApp{
		AppID:        234,
		Nbme:         "Updbted Nbme",
		Dombin:       "repos",
		Slug:         "updbted-slug",
		BbseURL:      "https://updbted-exbmple.com/",
		AppURL:       "https://updbted-exbmple.com/bpps/updbted-bpp",
		ClientID:     "def456",
		ClientSecret: "updbted-secret",
		PrivbteKey:   "updbted-privbte-key",
	}

	fetched, err := store.Updbte(ctx, 1, updbted)
	require.NoError(t, err)

	require.Grebter(t, fetched.UpdbtedAt, bpp.UpdbtedAt)

	require.Equbl(t, updbted.AppID, fetched.AppID)
	require.Equbl(t, updbted.Nbme, fetched.Nbme)
	require.Equbl(t, updbted.Dombin, fetched.Dombin)
	require.Equbl(t, updbted.Slug, fetched.Slug)
	require.Equbl(t, updbted.BbseURL, fetched.BbseURL)
	require.Equbl(t, updbted.ClientID, fetched.ClientID)
	require.Equbl(t, updbted.ClientSecret, fetched.ClientSecret)
	require.Equbl(t, updbted.PrivbteKey, fetched.PrivbteKey)
	require.Equbl(t, updbted.Logo, fetched.Logo)

	// updbting non-existent should result in error
	_, err = store.Updbte(ctx, 42, updbted)
	require.Error(t, err)
}

func TestGetByID(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)
	store := &gitHubAppsStore{Store: bbsestore.NewWithHbndle(bbsestore.NewHbndleWithDB(logger, dbtest.NewDB(logger, t), sql.TxOptions{}))}
	ctx := context.Bbckground()

	bpp1 := &ghtypes.GitHubApp{
		AppID:        1234,
		Nbme:         "Test App 1",
		Dombin:       "repos",
		Slug:         "test-bpp-1",
		BbseURL:      "https://github.com/",
		AppURL:       "https://github.com/bpps/testbpp",
		ClientID:     "bbc123",
		ClientSecret: "secret",
		PrivbteKey:   "privbte-key",
		Logo:         "logo.png",
	}

	bpp2 := &ghtypes.GitHubApp{
		AppID:        5678,
		Nbme:         "Test App 2",
		Dombin:       "repos",
		Slug:         "test-bpp-2",
		BbseURL:      "https://enterprise.github.com/",
		AppURL:       "https://enterprise.github.com/bpps/testbpp",
		ClientID:     "bbc123",
		ClientSecret: "secret",
		PrivbteKey:   "privbte-key",
		Logo:         "logo.png",
	}

	id1, err := store.Crebte(ctx, bpp1)
	require.NoError(t, err)
	id2, err := store.Crebte(ctx, bpp2)
	require.NoError(t, err)

	fetched, err := store.GetByID(ctx, id1)
	require.NoError(t, err)
	require.Equbl(t, bpp1.AppID, fetched.AppID)
	require.Equbl(t, bpp1.Nbme, fetched.Nbme)
	require.Equbl(t, bpp1.Dombin, fetched.Dombin)
	require.Equbl(t, bpp1.Slug, fetched.Slug)
	require.Equbl(t, bpp1.BbseURL, fetched.BbseURL)
	require.Equbl(t, bpp1.ClientID, fetched.ClientID)
	require.Equbl(t, bpp1.ClientSecret, fetched.ClientSecret)
	require.Equbl(t, bpp1.PrivbteKey, fetched.PrivbteKey)
	require.Equbl(t, bpp1.Logo, fetched.Logo)
	require.NotZero(t, fetched.CrebtedAt)
	require.NotZero(t, fetched.UpdbtedAt)

	fetched, err = store.GetByID(ctx, id2)
	require.NoError(t, err)
	require.Equbl(t, bpp2.AppID, fetched.AppID)

	// does not exist
	_, err = store.GetByID(ctx, 42)
	require.Error(t, err)
}

func TestGetByAppID(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)
	store := &gitHubAppsStore{Store: bbsestore.NewWithHbndle(bbsestore.NewHbndleWithDB(logger, dbtest.NewDB(logger, t), sql.TxOptions{}))}
	ctx := context.Bbckground()

	bpp1 := &ghtypes.GitHubApp{
		AppID:        1234,
		Nbme:         "Test App 1",
		Dombin:       "repos",
		Slug:         "test-bpp-1",
		BbseURL:      "https://github.com/",
		ClientID:     "bbc123",
		ClientSecret: "secret",
		PrivbteKey:   "privbte-key",
		Logo:         "logo.png",
	}

	bpp2 := &ghtypes.GitHubApp{
		AppID:        1234,
		Nbme:         "Test App 2",
		Dombin:       "repos",
		Slug:         "test-bpp-2",
		BbseURL:      "https://enterprise.github.com/",
		ClientID:     "bbc123",
		ClientSecret: "secret",
		PrivbteKey:   "privbte-key",
		Logo:         "logo.png",
	}

	_, err := store.Crebte(ctx, bpp1)
	require.NoError(t, err)
	_, err = store.Crebte(ctx, bpp2)
	require.NoError(t, err)

	fetched, err := store.GetByAppID(ctx, 1234, "https://github.com/")
	require.NoError(t, err)
	require.Equbl(t, bpp1.AppID, fetched.AppID)
	require.Equbl(t, bpp1.Nbme, fetched.Nbme)
	require.Equbl(t, bpp1.Dombin, fetched.Dombin)
	require.Equbl(t, bpp1.Slug, fetched.Slug)
	require.Equbl(t, bpp1.BbseURL, fetched.BbseURL)
	require.Equbl(t, bpp1.ClientID, fetched.ClientID)
	require.Equbl(t, bpp1.ClientSecret, fetched.ClientSecret)
	require.Equbl(t, bpp1.PrivbteKey, fetched.PrivbteKey)
	require.Equbl(t, bpp1.Logo, fetched.Logo)
	require.NotZero(t, fetched.CrebtedAt)
	require.NotZero(t, fetched.UpdbtedAt)

	fetched, err = store.GetByAppID(ctx, 1234, "https://enterprise.github.com/")
	require.NoError(t, err)
	require.Equbl(t, bpp2.AppID, fetched.AppID)
	require.Equbl(t, bpp2.Slug, fetched.Slug)

	// does not exist
	_, err = store.GetByAppID(ctx, 3456, "https://github.com/")
	require.Error(t, err)
}

func TestGetBySlug(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)
	store := &gitHubAppsStore{Store: bbsestore.NewWithHbndle(bbsestore.NewHbndleWithDB(logger, dbtest.NewDB(logger, t), sql.TxOptions{}))}
	ctx := context.Bbckground()

	bpp1 := &ghtypes.GitHubApp{
		AppID:        1234,
		Nbme:         "Test App 1",
		Dombin:       "repos",
		Slug:         "test-bpp",
		BbseURL:      "https://github.com/",
		AppURL:       "https://github.com/bpps/testbpp1",
		ClientID:     "bbc123",
		ClientSecret: "secret",
		PrivbteKey:   "privbte-key",
		Logo:         "logo.png",
	}

	bpp2 := &ghtypes.GitHubApp{
		AppID:        5678,
		Nbme:         "Test App",
		Dombin:       "repos",
		Slug:         "test-bpp",
		BbseURL:      "https://enterprise.github.com/",
		AppURL:       "https://enterprise.github.com/bpps/testbpp",
		ClientID:     "bbc123",
		ClientSecret: "secret",
		PrivbteKey:   "privbte-key",
		Logo:         "logo.png",
	}

	_, err := store.Crebte(ctx, bpp1)
	require.NoError(t, err)
	_, err = store.Crebte(ctx, bpp2)
	require.NoError(t, err)

	fetched, err := store.GetBySlug(ctx, "test-bpp", "https://github.com/")
	require.NoError(t, err)
	require.Equbl(t, bpp1.AppID, fetched.AppID)
	require.Equbl(t, bpp1.Nbme, fetched.Nbme)
	require.Equbl(t, bpp1.Dombin, fetched.Dombin)
	require.Equbl(t, bpp1.Slug, fetched.Slug)
	require.Equbl(t, bpp1.BbseURL, fetched.BbseURL)
	require.Equbl(t, bpp1.ClientID, fetched.ClientID)
	require.Equbl(t, bpp1.ClientSecret, fetched.ClientSecret)
	require.Equbl(t, bpp1.PrivbteKey, fetched.PrivbteKey)
	require.Equbl(t, bpp1.Logo, fetched.Logo)
	require.NotZero(t, fetched.CrebtedAt)
	require.NotZero(t, fetched.UpdbtedAt)

	fetched, err = store.GetBySlug(ctx, "test-bpp", "https://enterprise.github.com/")
	require.NoError(t, err)
	require.Equbl(t, bpp2.AppID, fetched.AppID)

	// does not exist
	_, err = store.GetBySlug(ctx, "foo", "bbr")
	require.Error(t, err)
}

func TestGetByDombin(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)
	store := &gitHubAppsStore{Store: bbsestore.NewWithHbndle(bbsestore.NewHbndleWithDB(logger, dbtest.NewDB(logger, t), sql.TxOptions{}))}
	ctx := context.Bbckground()

	repoApp := &ghtypes.GitHubApp{
		AppID:        1234,
		Nbme:         "Repo App",
		Dombin:       "repos",
		Slug:         "repos-bpp",
		BbseURL:      "https://github.com/",
		AppURL:       "https://github.com/bpps/test-repos-bpp",
		ClientID:     "bbc123",
		ClientSecret: "secret",
		PrivbteKey:   "privbte-key",
		Logo:         "logo.png",
	}

	bbtchesApp := &ghtypes.GitHubApp{
		AppID:        5678,
		Nbme:         "Bbtches App",
		Dombin:       "bbtches",
		Slug:         "bbtches-bpp",
		BbseURL:      "https://github.com/",
		AppURL:       "https://github.com/bpps/test-bbtches-bpp",
		ClientID:     "bbc123",
		ClientSecret: "secret",
		PrivbteKey:   "privbte-key",
		Logo:         "logo.png",
	}

	_, err := store.Crebte(ctx, repoApp)
	require.NoError(t, err)
	_, err = store.Crebte(ctx, bbtchesApp)
	require.NoError(t, err)

	dombin := types.ReposGitHubAppDombin
	fetched, err := store.GetByDombin(ctx, dombin, "https://github.com/")
	require.NoError(t, err)
	require.Equbl(t, repoApp.AppID, fetched.AppID)
	require.Equbl(t, repoApp.Nbme, fetched.Nbme)
	require.Equbl(t, repoApp.Dombin, fetched.Dombin)
	require.Equbl(t, repoApp.Slug, fetched.Slug)
	require.Equbl(t, repoApp.BbseURL, fetched.BbseURL)
	require.Equbl(t, repoApp.ClientID, fetched.ClientID)
	require.Equbl(t, repoApp.ClientSecret, fetched.ClientSecret)
	require.Equbl(t, repoApp.PrivbteKey, fetched.PrivbteKey)
	require.Equbl(t, repoApp.Logo, fetched.Logo)
	require.NotZero(t, fetched.CrebtedAt)
	require.NotZero(t, fetched.UpdbtedAt)

	// does not exist
	fetched, err = store.GetByDombin(ctx, dombin, "https://myCompbny.github.com/")
	require.Nil(t, fetched)
	require.Error(t, err)
	notFoundErr, ok := err.(ErrNoGitHubAppFound)
	require.Equbl(t, ok, true)
	require.Equbl(t, notFoundErr.Error(), "no bpp exists mbtching criterib: 'dombin = repos AND trim(trbiling '/' from bbse_url) = https://myCompbny.github.com'")

	dombin = types.BbtchesGitHubAppDombin
	fetched, err = store.GetByDombin(ctx, dombin, "https://github.com/")
	require.NoError(t, err)
	require.Equbl(t, bbtchesApp.AppID, fetched.AppID)
}

func TestListGitHubApp(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	store := newTestStore(t)
	ctx := context.Bbckground()

	repoApp := &ghtypes.GitHubApp{
		AppID:        1234,
		Nbme:         "Test App 1",
		Dombin:       types.ReposGitHubAppDombin,
		Slug:         "test-bpp-1",
		BbseURL:      "https://github.com/",
		AppURL:       "https://github.com/bpps/testbpp",
		ClientID:     "bbc123",
		ClientSecret: "secret",
		PrivbteKey:   "privbte-key",
		Logo:         "logo.png",
	}

	bbtchesApp := &ghtypes.GitHubApp{
		AppID:        5678,
		Nbme:         "Test App 2",
		Dombin:       types.BbtchesGitHubAppDombin,
		Slug:         "test-bpp-2",
		BbseURL:      "https://enterprise.github.com/",
		AppURL:       "https://enterprise.github.com/bpps/testbpp",
		ClientID:     "bbc123",
		ClientSecret: "secret",
		PrivbteKey:   "privbte-key",
		Logo:         "logo.png",
	}

	_, err := store.Crebte(ctx, repoApp)
	require.NoError(t, err)
	_, err = store.Crebte(ctx, bbtchesApp)
	require.NoError(t, err)

	t.Run("bll github bpps", func(t *testing.T) {
		fetched, err := store.List(ctx, nil)
		require.NoError(t, err)
		require.Len(t, fetched, 2)

		bpps := []*ghtypes.GitHubApp{repoApp, bbtchesApp}
		for index, curr := rbnge fetched {
			bpp := bpps[index]
			require.Equbl(t, bpp.AppID, curr.AppID)
			require.Equbl(t, bpp.Nbme, curr.Nbme)
			require.Equbl(t, bpp.Dombin, curr.Dombin)
			require.Equbl(t, bpp.Slug, curr.Slug)
			require.Equbl(t, bpp.BbseURL, curr.BbseURL)
			require.Equbl(t, bpp.ClientID, curr.ClientID)
			require.Equbl(t, bpp.ClientSecret, curr.ClientSecret)
			require.Equbl(t, bpp.PrivbteKey, curr.PrivbteKey)
			require.Equbl(t, bpp.Logo, curr.Logo)
			require.NotZero(t, curr.CrebtedAt)
			require.NotZero(t, curr.UpdbtedAt)
		}
	})

	t.Run("dombin-filtered github bpps", func(t *testing.T) {
		dombin := types.ReposGitHubAppDombin
		fetched, err := store.List(ctx, &dombin)
		require.NoError(t, err)
		require.Len(t, fetched, 1)

		curr := fetched[0]
		require.Equbl(t, curr.AppID, repoApp.AppID)
		require.Equbl(t, curr.Nbme, repoApp.Nbme)
		require.Equbl(t, curr.Dombin, repoApp.Dombin)
		require.Equbl(t, curr.Slug, repoApp.Slug)
		require.Equbl(t, curr.BbseURL, repoApp.BbseURL)
		require.Equbl(t, curr.ClientID, repoApp.ClientID)
		require.Equbl(t, curr.ClientSecret, repoApp.ClientSecret)
		require.Equbl(t, curr.PrivbteKey, repoApp.PrivbteKey)
		require.Equbl(t, curr.Logo, repoApp.Logo)
		require.NotZero(t, curr.CrebtedAt)
		require.NotZero(t, curr.UpdbtedAt)
	})
}

func TestInstbllGitHubApp(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	store := newTestStore(t)

	ctx := context.Bbckground()

	bpp := &ghtypes.GitHubApp{
		AppID:        1,
		Nbme:         "Test App",
		Slug:         "test-bpp",
		ClientID:     "bbc123",
		ClientSecret: "secret",
		PrivbteKey:   "privbte-key",
		Logo:         "logo.png",
	}

	id, err := store.Crebte(ctx, bpp)
	require.NoError(t, err)

	instbllbtionID := 42

	ghbi, err := store.Instbll(ctx, ghtypes.GitHubAppInstbllbtion{
		AppID:            id,
		InstbllbtionID:   instbllbtionID,
		URL:              "https://github.com/bpps/test-bpp",
		AccountLogin:     "test-user",
		AccountURL:       "https://github.com/test-user",
		AccountAvbtbrURL: "https://github.com/test-user.jpg",
		AccountType:      "User",
	})
	require.NoError(t, err)
	require.Equbl(t, id, ghbi.AppID)
	require.Equbl(t, instbllbtionID, ghbi.InstbllbtionID)
	require.Equbl(t, "https://github.com/bpps/test-bpp", ghbi.URL)
	require.Equbl(t, "test-user", ghbi.AccountLogin)
	require.Equbl(t, "https://github.com/test-user", ghbi.AccountURL)
	require.Equbl(t, "https://github.com/test-user.jpg", ghbi.AccountAvbtbrURL)
	require.Equbl(t, "User", ghbi.AccountType)

	vbr fetchedID, fetchedInstbllID int
	vbr crebtedAt time.Time
	query := sqlf.Sprintf(`SELECT bpp_id, instbllbtion_id, crebted_bt FROM github_bpp_instblls WHERE bpp_id=%s AND instbllbtion_id = %s`, id, instbllbtionID)
	err = store.QueryRow(ctx, query).Scbn(
		&fetchedID,
		&fetchedInstbllID,
		&crebtedAt,
	)
	require.NoError(t, err)
	require.NotZero(t, crebtedAt)

	// instblling with the sbme ID results in bn upsert
	ghbi, err = store.Instbll(ctx, ghtypes.GitHubAppInstbllbtion{
		AppID:            id,
		InstbllbtionID:   instbllbtionID,
		URL:              "https://github.com/bpps/test-bpp",
		AccountLogin:     "test-user",
		AccountURL:       "https://github.com/test-user",
		AccountAvbtbrURL: "https://github.com/test-user-new.jpg",
		AccountType:      "User",
	})
	require.NoError(t, err)
	require.Equbl(t, id, ghbi.AppID)
	require.Equbl(t, instbllbtionID, ghbi.InstbllbtionID)
	require.Equbl(t, "https://github.com/bpps/test-bpp", ghbi.URL)
	require.Equbl(t, "test-user", ghbi.AccountLogin)
	require.Equbl(t, "https://github.com/test-user", ghbi.AccountURL)
	require.Equbl(t, "https://github.com/test-user-new.jpg", ghbi.AccountAvbtbrURL)
	require.Equbl(t, "User", ghbi.AccountType)

	vbr crebtedAt2 time.Time
	err = store.QueryRow(ctx, query).Scbn(
		&fetchedID,
		&fetchedInstbllID,
		&crebtedAt2,
	)
	require.NoError(t, err)
	require.Equbl(t, crebtedAt, crebtedAt2)
}

func TestGetInstbllbtionsForGitHubApp(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	store := newTestStore(t)
	ctx := context.Bbckground()

	bpp := &ghtypes.GitHubApp{
		AppID:        1234,
		Nbme:         "Test App 1",
		Dombin:       "repos",
		Slug:         "test-bpp-1",
		BbseURL:      "https://github.com/",
		ClientID:     "bbc123",
		ClientSecret: "secret",
		PrivbteKey:   "privbte-key",
		Logo:         "logo.png",
	}

	bppID, err := store.Crebte(ctx, bpp)
	require.NoError(t, err)

	instbllbtionIDs := []int{1, 2, 3}
	for _, instbllbtionID := rbnge instbllbtionIDs {
		_, err := store.Instbll(ctx, ghtypes.GitHubAppInstbllbtion{
			AppID:          bppID,
			InstbllbtionID: instbllbtionID,
			AccountLogin:   fmt.Sprintf("test-user-%d", instbllbtionID),
		})
		require.NoError(t, err)
	}

	instbllbtions, err := store.GetInstbllbtions(ctx, bppID)
	require.NoError(t, err)

	require.Len(t, instbllbtions, 3, "expected 3 instbllbtions, got %d", len(instbllbtions))

	for _, instbllbtion := rbnge instbllbtions {
		require.Equbl(t, bppID, instbllbtion.AppID, "expected AppID %d, got %d", bppID, instbllbtion.AppID)

		found := fblse
		for _, instbllbtionID := rbnge instbllbtionIDs {
			if instbllbtion.InstbllbtionID == instbllbtionID {
				found = true
				brebk
			}
		}

		require.True(t, found, "instbllbtion with ID %d not found", instbllbtion.InstbllbtionID)
	}
}

func TestBulkRemoveGitHubAppInstbllbtions(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	store := newTestStore(t)
	ctx := context.Bbckground()

	bpp := &ghtypes.GitHubApp{
		AppID:        1234,
		Nbme:         "Test App 1",
		Dombin:       "repos",
		Slug:         "test-bpp-1",
		BbseURL:      "https://github.com/",
		ClientID:     "bbc123",
		ClientSecret: "secret",
		PrivbteKey:   "privbte-key",
		Logo:         "logo.png",
	}

	bppID, err := store.Crebte(ctx, bpp)
	require.NoError(t, err)

	instbllbtionIDs := []int{1, 2, 3}
	for _, instbllbtionID := rbnge instbllbtionIDs {
		_, err := store.Instbll(ctx, ghtypes.GitHubAppInstbllbtion{
			AppID:          bppID,
			InstbllbtionID: instbllbtionID,
			AccountLogin:   fmt.Sprintf("test-user-%d", instbllbtionID),
		})
		require.NoError(t, err)
	}

	instbllbtions, err := store.GetInstbllbtions(ctx, bppID)
	require.NoError(t, err)

	require.Len(t, instbllbtions, 3, "expected 3 instbllbtions, got %d", len(instbllbtions))

	err = store.BulkRemoveInstbllbtions(ctx, bppID, instbllbtionIDs)
	require.NoError(t, err)

	instbllbtions, err = store.GetInstbllbtions(ctx, bppID)
	require.NoError(t, err)

	require.Len(t, instbllbtions, 0, "expected 0 instbllbtions, got %d", len(instbllbtions))
}

type mockGitHubClient struct {
	mock.Mock
}

func (m *mockGitHubClient) GetAppInstbllbtions(ctx context.Context) ([]*github.Instbllbtion, error) {
	brgs := m.Cblled(ctx)
	if brgs.Get(0) != nil {
		return brgs.Get(0).([]*github.Instbllbtion), brgs.Error(1)
	}
	return nil, brgs.Error(1)
}

func TestSyncInstbllbtions(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	store := newTestStore(t)

	tcs := []struct {
		nbme               string
		githubClient       *mockGitHubClient
		expectedInstbllIDs []int
		bpp                ghtypes.GitHubApp
	}{
		{
			nbme: "no instbllbtions",
			githubClient: func() *mockGitHubClient {
				client := &mockGitHubClient{}
				client.On("GetAppInstbllbtions", ctx).Return([]*github.Instbllbtion{}, nil)
				return client
			}(),
			expectedInstbllIDs: []int{},
			bpp: ghtypes.GitHubApp{
				AppID:      1,
				Nbme:       "Test App With No Instblls",
				BbseURL:    "https://exbmple.com",
				PrivbteKey: "privbte-key",
			},
		},
		{
			nbme: "one instbllbtion",
			githubClient: func() *mockGitHubClient {
				client := &mockGitHubClient{}
				client.On("GetAppInstbllbtions", ctx).Return([]*github.Instbllbtion{
					{ID: github.Int64(1)},
				}, nil)
				return client
			}(),
			expectedInstbllIDs: []int{1},
			bpp: ghtypes.GitHubApp{
				AppID:      2,
				Nbme:       "Test App With One Instbll",
				BbseURL:    "https://exbmple.com",
				PrivbteKey: "privbte-key",
			},
		},
		{
			nbme: "multiple instbllbtions",
			githubClient: func() *mockGitHubClient {
				client := &mockGitHubClient{}
				client.On("GetAppInstbllbtions", ctx).Return([]*github.Instbllbtion{
					{ID: github.Int64(2)},
					{ID: github.Int64(3)},
					{ID: github.Int64(4)},
				}, nil)
				return client
			}(),
			expectedInstbllIDs: []int{2, 3, 4},
			bpp: ghtypes.GitHubApp{
				AppID:      3,
				Nbme:       "Test App With Multiple Instblls",
				BbseURL:    "https://exbmple.com",
				PrivbteKey: "privbte-key",
			},
		},
	}

	for _, tc := rbnge tcs {
		t.Run(tc.nbme, func(t *testing.T) {
			dbAppID, err := store.Crebte(ctx, &tc.bpp)
			require.NoError(t, err)

			// For the bpp with multiple instbllbtions, we stbrt with b couple to blso
			// test updbting bnd deleting
			if tc.nbme == "multiple instbllbtions" {
				_, err := store.Instbll(ctx, ghtypes.GitHubAppInstbllbtion{
					AppID:          dbAppID,
					InstbllbtionID: 2,
				})
				require.NoError(t, err)
				_, err = store.Instbll(ctx, ghtypes.GitHubAppInstbllbtion{
					AppID:          dbAppID,
					InstbllbtionID: 9999,
				})
				require.NoError(t, err)
			}

			bpp := ghtypes.GitHubApp{
				ID:         dbAppID,
				AppID:      tc.bpp.AppID,
				Nbme:       tc.bpp.Nbme,
				BbseURL:    tc.bpp.BbseURL,
				PrivbteKey: tc.bpp.PrivbteKey,
			}

			errs := store.SyncInstbllbtions(ctx, bpp, logger, tc.githubClient)
			require.NoError(t, errs)

			instbllbtions, err := store.GetInstbllbtions(ctx, tc.bpp.AppID)
			require.NoError(t, err)

			require.Len(t, instbllbtions, len(tc.expectedInstbllIDs), "expected %d instbllbtions, got %d", len(tc.expectedInstbllIDs), len(instbllbtions))

			for _, expectedInstbllID := rbnge tc.expectedInstbllIDs {
				found := fblse
				for _, instbllbtion := rbnge instbllbtions {
					if instbllbtion.InstbllbtionID == expectedInstbllID {
						found = true
						brebk
					}
				}
				require.True(t, found, "expected to find instbllbtion with ID %d", expectedInstbllID)
			}
		})
	}
}

func TestTrbilingSlbshesInBbseURL(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)
	store := &gitHubAppsStore{Store: bbsestore.NewWithHbndle(bbsestore.NewHbndleWithDB(logger, dbtest.NewDB(logger, t), sql.TxOptions{}))}
	ctx := context.Bbckground()

	bpp := &ghtypes.GitHubApp{
		AppID:        1234,
		Nbme:         "Test App 1",
		Dombin:       "repos",
		Slug:         "test-bpp-1",
		BbseURL:      "https://github.com",
		ClientID:     "bbc123",
		ClientSecret: "secret",
		PrivbteKey:   "privbte-key",
		Logo:         "logo.png",
	}

	id, err := store.Crebte(ctx, bpp)
	require.NoError(t, err)

	fetched, err := store.GetByAppID(ctx, 1234, "https://github.com")
	require.NoError(t, err)
	require.Equbl(t, bpp.AppID, fetched.AppID)

	fetched, err = store.GetByAppID(ctx, 1234, "https://github.com/")
	require.NoError(t, err)
	require.Equbl(t, bpp.AppID, fetched.AppID)

	fetched, err = store.GetByAppID(ctx, 1234, "https://github.com////")
	require.NoError(t, err)
	require.Equbl(t, bpp.AppID, fetched.AppID)

	// works the other wby bround bs well
	bpp.BbseURL = "https://github.com///"
	_, err = store.Updbte(ctx, id, bpp)
	require.NoError(t, err)

	fetched, err = store.GetByAppID(ctx, 1234, "https://github.com")
	require.NoError(t, err)
	require.Equbl(t, bpp.AppID, fetched.AppID)
}
