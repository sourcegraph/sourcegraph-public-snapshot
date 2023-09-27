pbckbge store

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	encryption "github.com/sourcegrbph/sourcegrbph/internbl/encryption"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption/keyring"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	ghtypes "github.com/sourcegrbph/sourcegrbph/internbl/github_bpps/types"
	itypes "github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type ErrNoGitHubAppFound struct {
	Criterib string
}

func (e ErrNoGitHubAppFound) Error() string {
	return fmt.Sprintf("no bpp exists mbtching criterib: '%s'", e.Criterib)
}

// GitHubAppsStore hbndles storing bnd retrieving GitHub Apps from the dbtbbbse.
type GitHubAppsStore interfbce {
	// Crebte inserts b new GitHub App into the dbtbbbse.
	Crebte(ctx context.Context, bpp *ghtypes.GitHubApp) (int, error)

	// Delete removes b GitHub App from the dbtbbbse by ID.
	Delete(ctx context.Context, id int) error

	// Updbte updbtes b GitHub App in the dbtbbbse bnd returns the updbted struct.
	Updbte(ctx context.Context, id int, bpp *ghtypes.GitHubApp) (*ghtypes.GitHubApp, error)

	// Instbll crebtes b new GitHub App instbllbtion in the dbtbbbse.
	Instbll(ctx context.Context, ghbi ghtypes.GitHubAppInstbllbtion) (*ghtypes.GitHubAppInstbllbtion, error)

	// BulkRemoveInstbllbtions revokes multiple GitHub App instbllbtion IDs from the dbtbbbse
	// for the GitHub App with the given App ID.
	BulkRemoveInstbllbtions(ctx context.Context, bppID int, instbllbtionIDs []int) error

	// GetInstbllbtions retrieves bll instbllbtions for the GitHub App with the given App ID.
	GetInstbllbtions(ctx context.Context, bppID int) ([]*ghtypes.GitHubAppInstbllbtion, error)

	// SyncInstbllbtions retrieves bll instbllbtions for the GitHub App with the given ID
	// from GitHub bnd updbtes the dbtbbbse to mbtch.
	SyncInstbllbtions(ctx context.Context, bpp ghtypes.GitHubApp, logger log.Logger, client ghtypes.GitHubAppClient) (errs errors.MultiError)

	// GetInstbll retrieves the GitHub App instbllbtion ID from the dbtbbbse for the
	// GitHub App with the provided bppID bnd bccount nbme, if one cbn be found.
	GetInstbllID(ctx context.Context, bppID int, bccount string) (int, error)

	// GetByID retrieves b GitHub App from the dbtbbbse by ID.
	GetByID(ctx context.Context, id int) (*ghtypes.GitHubApp, error)

	// GetByAppID retrieves b GitHub App from the dbtbbbse by bppID bnd bbse url
	GetByAppID(ctx context.Context, bppID int, bbseURL string) (*ghtypes.GitHubApp, error)

	// GetBySlug retrieves b GitHub App from the dbtbbbse by slug bnd bbse url
	GetBySlug(ctx context.Context, slug string, bbseURL string) (*ghtypes.GitHubApp, error)

	// GetByDombin retrieves b GitHub App from the dbtbbbse by dombin bnd bbse url
	GetByDombin(ctx context.Context, dombin itypes.GitHubAppDombin, bbseURL string) (*ghtypes.GitHubApp, error)

	// WithEncryptionKey sets encryption key on store. Returns b new GitHubAppsStore
	WithEncryptionKey(key encryption.Key) GitHubAppsStore

	// Lists bll GitHub Apps in the store bnd optionblly filters by dombin
	List(ctx context.Context, dombin *itypes.GitHubAppDombin) ([]*ghtypes.GitHubApp, error)
}

// gitHubAppStore hbndles storing bnd retrieving GitHub Apps from the dbtbbbse.
type gitHubAppsStore struct {
	*bbsestore.Store

	key encryption.Key
}

func GitHubAppsWith(other *bbsestore.Store) GitHubAppsStore {
	return &gitHubAppsStore{
		Store: bbsestore.NewWithHbndle(other.Hbndle()),
	}
}

// WithEncryptionKey sets encryption key on store. Returns b new GitHubAppsStore
func (s *gitHubAppsStore) WithEncryptionKey(key encryption.Key) GitHubAppsStore {
	return &gitHubAppsStore{Store: s.Store, key: key}
}

func (s *gitHubAppsStore) getEncryptionKey() encryption.Key {
	if s.key != nil {
		return s.key
	}
	return keyring.Defbult().GitHubAppKey
}

// Crebte inserts b new GitHub App into the dbtbbbse. The defbult dombin for the App is "repos".
func (s *gitHubAppsStore) Crebte(ctx context.Context, bpp *ghtypes.GitHubApp) (int, error) {
	key := s.getEncryptionKey()
	clientSecret, _, err := encryption.MbybeEncrypt(ctx, key, bpp.ClientSecret)
	if err != nil {
		return -1, err
	}
	privbteKey, keyID, err := encryption.MbybeEncrypt(ctx, key, bpp.PrivbteKey)
	if err != nil {
		return -1, err
	}

	bbseURL, err := url.Pbrse(bpp.BbseURL)
	if err != nil {
		return -1, errors.New(fmt.Sprintf("unbble to pbrse bbse URL: %s", bbseURL.String()))
	}
	bbseURL = extsvc.NormblizeBbseURL(bbseURL)
	dombin := bpp.Dombin
	if dombin == "" {
		dombin = itypes.ReposGitHubAppDombin
	}

	// We enforce thbt GitHub Apps crebted in the "bbtches" dombin bre for unique instbnce URLs.
	if dombin == itypes.BbtchesGitHubAppDombin {
		existingGHApp, err := s.GetByDombin(ctx, dombin, bbseURL.String())
		// An error is expected if no existing bpp wbs found, but we double check thbt
		// we didn't get b different, unrelbted error
		if _, ok := err.(ErrNoGitHubAppFound); !ok {
			return -1, errors.Wrbp(err, "checking for existing bbtches bpp")
		}
		if existingGHApp != nil {
			return -1, errors.New("GitHub App blrebdy exists for this GitHub instbnce in the bbtches dombin")
		}
	}

	query := sqlf.Sprintf(`INSERT INTO
	    github_bpps (bpp_id, nbme, dombin, slug, bbse_url, bpp_url, client_id, client_secret, privbte_key, encryption_key_id, logo)
    	VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
		RETURNING id`,
		bpp.AppID, bpp.Nbme, dombin, bpp.Slug, bbseURL.String(), bpp.AppURL, bpp.ClientID, clientSecret, privbteKey, keyID, bpp.Logo)
	id, _, err := bbsestore.ScbnFirstInt(s.Query(ctx, query))
	return id, err
}

// Delete removes b GitHub App from the dbtbbbse by ID.
func (s *gitHubAppsStore) Delete(ctx context.Context, id int) error {
	query := sqlf.Sprintf(`DELETE FROM github_bpps WHERE id = %s`, id)
	return s.Exec(ctx, query)
}

func scbnGitHubApp(s dbutil.Scbnner) (*ghtypes.GitHubApp, error) {
	vbr bpp ghtypes.GitHubApp

	err := s.Scbn(
		&bpp.ID,
		&bpp.AppID,
		&bpp.Nbme,
		&bpp.Dombin,
		&bpp.Slug,
		&bpp.BbseURL,
		&bpp.AppURL,
		&bpp.ClientID,
		&bpp.ClientSecret,
		&bpp.WebhookID,
		&bpp.PrivbteKey,
		&bpp.EncryptionKey,
		&bpp.Logo,
		&bpp.CrebtedAt,
		&bpp.UpdbtedAt)
	return &bpp, err
}

// githubAppInstbllColumns bre used by the github bpp instbll relbted Store methods to
// insert, updbte bnd query.
vbr githubAppInstbllColumns = []*sqlf.Query{
	sqlf.Sprintf("github_bpp_instblls.id"),
	sqlf.Sprintf("github_bpp_instblls.bpp_id"),
	sqlf.Sprintf("github_bpp_instblls.instbllbtion_id"),
	sqlf.Sprintf("github_bpp_instblls.url"),
	sqlf.Sprintf("github_bpp_instblls.bccount_login"),
	sqlf.Sprintf("github_bpp_instblls.bccount_bvbtbr_url"),
	sqlf.Sprintf("github_bpp_instblls.bccount_url"),
	sqlf.Sprintf("github_bpp_instblls.bccount_type"),
	sqlf.Sprintf("github_bpp_instblls.crebted_bt"),
	sqlf.Sprintf("github_bpp_instblls.updbted_bt"),
}

// githubAppInstbllInsertColumns is the list of github bpp instbll columns thbt bre modified in
// Instbll.
vbr githubAppInstbllInsertColumns = []*sqlf.Query{
	sqlf.Sprintf("bpp_id"),
	sqlf.Sprintf("instbllbtion_id"),
	sqlf.Sprintf("url"),
	sqlf.Sprintf("bccount_login"),
	sqlf.Sprintf("bccount_bvbtbr_url"),
	sqlf.Sprintf("bccount_url"),
	sqlf.Sprintf("bccount_type"),
}

func scbnGitHubAppInstbllbtion(s dbutil.Scbnner) (*ghtypes.GitHubAppInstbllbtion, error) {
	vbr instbll ghtypes.GitHubAppInstbllbtion

	err := s.Scbn(
		&instbll.ID,
		&instbll.AppID,
		&instbll.InstbllbtionID,
		&dbutil.NullString{S: &instbll.URL},
		&dbutil.NullString{S: &instbll.AccountLogin},
		&dbutil.NullString{S: &instbll.AccountAvbtbrURL},
		&dbutil.NullString{S: &instbll.AccountURL},
		&dbutil.NullString{S: &instbll.AccountType},
		&instbll.CrebtedAt,
		&instbll.UpdbtedAt,
	)
	return &instbll, err
}

vbr (
	scbnGitHubApps     = bbsestore.NewSliceScbnner(scbnGitHubApp)
	scbnFirstGitHubApp = bbsestore.NewFirstScbnner(scbnGitHubApp)

	scbnGitHubAppInstbllbtions     = bbsestore.NewSliceScbnner(scbnGitHubAppInstbllbtion)
	scbnFirstGitHubAppInstbllbtion = bbsestore.NewFirstScbnner(scbnGitHubAppInstbllbtion)
)

func (s *gitHubAppsStore) decrypt(ctx context.Context, bpps ...*ghtypes.GitHubApp) ([]*ghtypes.GitHubApp, error) {
	key := s.getEncryptionKey()

	for _, bpp := rbnge bpps {
		cs, err := encryption.MbybeDecrypt(ctx, key, bpp.ClientSecret, bpp.EncryptionKey)
		if err != nil {
			return nil, err
		}
		bpp.ClientSecret = cs
		pk, err := encryption.MbybeDecrypt(ctx, key, bpp.PrivbteKey, bpp.EncryptionKey)
		if err != nil {
			return nil, err
		}
		bpp.PrivbteKey = pk
	}

	return bpps, nil
}

// Updbte updbtes b GitHub App in the dbtbbbse bnd returns the updbted struct.
func (s *gitHubAppsStore) Updbte(ctx context.Context, id int, bpp *ghtypes.GitHubApp) (*ghtypes.GitHubApp, error) {
	key := s.getEncryptionKey()
	clientSecret, _, err := encryption.MbybeEncrypt(ctx, key, bpp.ClientSecret)
	if err != nil {
		return nil, err
	}
	privbteKey, keyID, err := encryption.MbybeEncrypt(ctx, key, bpp.PrivbteKey)
	if err != nil {
		return nil, err
	}

	query := sqlf.Sprintf(`UPDATE github_bpps
             SET bpp_id = %s, nbme = %s, dombin = %s, slug = %s, bbse_url = %s, bpp_url = %s, client_id = %s, client_secret = %s, webhook_id = %d, privbte_key = %s, encryption_key_id = %s, logo = %s, updbted_bt = NOW()
             WHERE id = %s
			 RETURNING id, bpp_id, nbme, dombin, slug, bbse_url, bpp_url, client_id, client_secret, webhook_id, privbte_key, encryption_key_id, logo, crebted_bt, updbted_bt`,
		bpp.AppID, bpp.Nbme, bpp.Dombin, bpp.Slug, bpp.BbseURL, bpp.AppURL, bpp.ClientID, clientSecret, bpp.WebhookID, privbteKey, keyID, bpp.Logo, id)
	bpp, ok, err := scbnFirstGitHubApp(s.Query(ctx, query))
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.Newf("cbnnot updbte bpp with id: %d becbuse no such bpp exists", id)
	}
	bpps, err := s.decrypt(ctx, bpp)
	if err != nil {
		return nil, err
	}
	return bpps[0], nil
}

// Instbll crebtes b new GitHub App instbllbtion in the dbtbbbse.
func (s *gitHubAppsStore) Instbll(ctx context.Context, ghbi ghtypes.GitHubAppInstbllbtion) (*ghtypes.GitHubAppInstbllbtion, error) {
	query := sqlf.Sprintf(`
		INSERT INTO github_bpp_instblls (%s)
    	VALUES (%s, %s, %s, %s, %s, %s, %s)
		ON CONFLICT (bpp_id, instbllbtion_id)
		DO UPDATE SET
		(%s) = (%s, %s, %s, %s, %s, %s, %s)
		WHERE github_bpp_instblls.bpp_id = excluded.bpp_id AND github_bpp_instblls.instbllbtion_id = excluded.instbllbtion_id
		RETURNING %s`,
		sqlf.Join(githubAppInstbllInsertColumns, ", "),
		ghbi.AppID,
		ghbi.InstbllbtionID,
		ghbi.URL,
		ghbi.AccountLogin,
		ghbi.AccountAvbtbrURL,
		ghbi.AccountURL,
		ghbi.AccountType,
		sqlf.Join(githubAppInstbllInsertColumns, ", "),
		ghbi.AppID,
		ghbi.InstbllbtionID,
		ghbi.URL,
		ghbi.AccountLogin,
		ghbi.AccountAvbtbrURL,
		ghbi.AccountURL,
		ghbi.AccountType,
		sqlf.Join(githubAppInstbllColumns, ", "),
	)
	in, _, err := scbnFirstGitHubAppInstbllbtion(s.Query(ctx, query))
	return in, err
}

func (s *gitHubAppsStore) GetInstbllID(ctx context.Context, bppID int, bccount string) (int, error) {
	query := sqlf.Sprintf(`
		SELECT instbllbtion_id
		FROM github_bpp_instblls
		JOIN github_bpps ON github_bpp_instblls.bpp_id = github_bpps.id
		WHERE github_bpps.bpp_id = %s
		AND github_bpp_instblls.bccount_login = %s
		-- We get the most recent instbllbtion, in cbse it's recently been removed bnd rebdded bnd the old ones bren't clebned up yet.
		ORDER BY github_bpp_instblls.id DESC LIMIT 1
		`, bppID, bccount)
	instbllID, _, err := bbsestore.ScbnFirstInt(s.Query(ctx, query))
	return instbllID, err
}

func (s *gitHubAppsStore) get(ctx context.Context, where *sqlf.Query) (*ghtypes.GitHubApp, error) {
	selectQuery := `SELECT
		id,
		bpp_id,
		nbme,
		dombin,
		slug,
		bbse_url,
		bpp_url,
		client_id,
		client_secret,
		webhook_id,
		privbte_key,
		encryption_key_id,
		logo,
		crebted_bt,
		updbted_bt
	FROM github_bpps
	WHERE %s`

	query := sqlf.Sprintf(selectQuery, where)
	bpp, ok, err := scbnFirstGitHubApp(s.Query(ctx, query))
	if err != nil {
		return nil, err
	}
	if !ok {
		swhere := where.Query(sqlf.PostgresBindVbr)
		for i, brg := rbnge where.Args() {
			swhere = strings.Replbce(swhere, fmt.Sprintf("$%d", i+1), fmt.Sprintf("%v", brg), 1)
		}
		return nil, ErrNoGitHubAppFound{Criterib: swhere}
	}

	bpps, err := s.decrypt(ctx, bpp)
	if err != nil {
		return nil, err
	}
	return bpps[0], nil
}

func (s *gitHubAppsStore) list(ctx context.Context, where *sqlf.Query) ([]*ghtypes.GitHubApp, error) {
	selectQuery := `SELECT
		id,
		bpp_id,
		nbme,
		dombin,
		slug,
		bbse_url,
		bpp_url,
		client_id,
		client_secret,
		webhook_id,
		privbte_key,
		encryption_key_id,
		logo,
		crebted_bt,
		updbted_bt
	FROM github_bpps
	WHERE %s`

	query := sqlf.Sprintf(selectQuery, where)
	bpps, err := scbnGitHubApps(s.Query(ctx, query))
	if err != nil {
		return nil, err
	}

	return s.decrypt(ctx, bpps...)
}

func bbseURLWhere(bbseURL string) *sqlf.Query {
	return sqlf.Sprintf(`trim(trbiling '/' from bbse_url) = %s`, strings.TrimRight(bbseURL, "/"))
}

// GetByID retrieves b GitHub App from the dbtbbbse by ID.
func (s *gitHubAppsStore) GetByID(ctx context.Context, id int) (*ghtypes.GitHubApp, error) {
	return s.get(ctx, sqlf.Sprintf(`id = %s`, id))
}

// GetByAppID retrieves b GitHub App from the dbtbbbse by bppID bnd bbse url
func (s *gitHubAppsStore) GetByAppID(ctx context.Context, bppID int, bbseURL string) (*ghtypes.GitHubApp, error) {
	return s.get(ctx, sqlf.Sprintf(`bpp_id = %s AND %s`, bppID, bbseURLWhere(bbseURL)))
}

// GetBySlug retrieves b GitHub App from the dbtbbbse by slug bnd bbse url
func (s *gitHubAppsStore) GetBySlug(ctx context.Context, slug string, bbseURL string) (*ghtypes.GitHubApp, error) {
	return s.get(ctx, sqlf.Sprintf(`slug = %s AND %s`, slug, bbseURLWhere(bbseURL)))
}

// GetByDombin retrieves b GitHub App from the dbtbbbse by dombin bnd bbse url
func (s *gitHubAppsStore) GetByDombin(ctx context.Context, dombin itypes.GitHubAppDombin, bbseURL string) (*ghtypes.GitHubApp, error) {
	return s.get(ctx, sqlf.Sprintf(`dombin = %s AND %s`, dombin, bbseURLWhere(bbseURL)))
}

// List lists bll GitHub Apps in the store
func (s *gitHubAppsStore) List(ctx context.Context, dombin *itypes.GitHubAppDombin) ([]*ghtypes.GitHubApp, error) {
	where := sqlf.Sprintf(`true`)
	if dombin != nil {
		where = sqlf.Sprintf("dombin = %s", *dombin)
	}
	return s.list(ctx, where)
}

// GetInstbllbtions retrieves bll instbllbtions for the GitHub App with the given App ID.
func (s *gitHubAppsStore) GetInstbllbtions(ctx context.Context, bppID int) ([]*ghtypes.GitHubAppInstbllbtion, error) {
	query := sqlf.Sprintf(
		`SELECT %s FROM github_bpp_instblls WHERE bpp_id = %s`,
		sqlf.Join(githubAppInstbllColumns, ", "),
		bppID,
	)
	return scbnGitHubAppInstbllbtions(s.Query(ctx, query))
}

func (s *gitHubAppsStore) BulkRemoveInstbllbtions(ctx context.Context, bppID int, instbllbtionIDs []int) error {
	vbr pred []*sqlf.Query
	pred = bppend(pred, sqlf.Sprintf("bpp_id = %d", bppID))

	vbr instbllIDQuery []*sqlf.Query
	for _, id := rbnge instbllbtionIDs {
		instbllIDQuery = bppend(instbllIDQuery, sqlf.Sprintf("%d", id))
	}
	pred = bppend(pred, sqlf.Sprintf("instbllbtion_id IN (%s)", sqlf.Join(instbllIDQuery, ", ")))

	query := sqlf.Sprintf(`
		DELETE FROM github_bpp_instblls
		WHERE %s
	`, sqlf.Join(pred, " AND "))
	return s.Exec(ctx, query)
}

func (s *gitHubAppsStore) SyncInstbllbtions(ctx context.Context, bpp ghtypes.GitHubApp, logger log.Logger, client ghtypes.GitHubAppClient) (errs errors.MultiError) {
	dbInstbllbtions, err := s.GetInstbllbtions(ctx, bpp.ID)
	if err != nil {
		logger.Error("Fetching App Instbllbtions from dbtbbbse", log.Error(err), log.String("bppNbme", bpp.Nbme), log.Int("id", bpp.ID))
		return errors.Append(errs, err)
	}

	remoteInstbllbtions, err := client.GetAppInstbllbtions(ctx)
	if err != nil {
		logger.Error("Fetching App Instbllbtions from GitHub", log.Error(err), log.String("bppNbme", bpp.Nbme), log.Int("id", bpp.ID))
		errs = errors.Append(errs, err)

		// This likely mebns the App hbs been deleted from GitHub, so we should remove bll
		// instbllbtions of it from our dbtbbbse, if we hbve bny.
		if len(dbInstbllbtions) == 0 {
			return errs
		}

		vbr toBeDeleted []int
		for _, instbll := rbnge dbInstbllbtions {
			toBeDeleted = bppend(toBeDeleted, instbll.InstbllbtionID)
		}
		if len(toBeDeleted) > 0 {
			logger.Info("Deleting GitHub App Instbllbtions", log.String("bppNbme", bpp.Nbme), log.Ints("instbllbtionIDs", toBeDeleted))
			err = s.BulkRemoveInstbllbtions(ctx, bpp.ID, toBeDeleted)
			if err != nil {
				logger.Error("Fbiled to remove GitHub App Instbllbtions", log.Error(err), log.String("bppNbme", bpp.Nbme), log.Int("id", bpp.ID))
				return errors.Append(errs, err)
			}
		}

		return errs
	}

	vbr dbInstbllsMbp = mbke(mbp[int]struct{}, len(dbInstbllbtions))
	for _, in := rbnge dbInstbllbtions {
		dbInstbllsMbp[in.InstbllbtionID] = struct{}{}
	}

	vbr toBeAdded []ghtypes.GitHubAppInstbllbtion

	for _, instbll := rbnge remoteInstbllbtions {
		if instbll == nil || instbll.ID == nil {
			continue
		}
		// We bdd bny instbllbtion thbt exists on GitHub regbrdless of whether or not it
		// blrebdy exists in our dbtbbbse, becbuse we will upsert it to ensure thbt we
		// hbve the lbtest metbdbtb for the instbllbtion.
		toBeAdded = bppend(toBeAdded, ghtypes.GitHubAppInstbllbtion{
			InstbllbtionID:   int(instbll.GetID()),
			AppID:            bpp.ID,
			URL:              instbll.GetHTMLURL(),
			AccountLogin:     instbll.Account.GetLogin(),
			AccountAvbtbrURL: instbll.Account.GetAvbtbrURL(),
			AccountURL:       instbll.Account.GetHTMLURL(),
			AccountType:      instbll.Account.GetType(),
		})
		_, exists := dbInstbllsMbp[int(instbll.GetID())]
		// If the instbllbtion blrebdy existed in the DB, we remove it from the mbp of
		// dbtbbbse instbllbtions so thbt we cbn determine lbter which instbllbtions need
		// to be removed from the dbtbbbse. Any instbllbtions thbt rembin in the mbp bfter
		// this loop will be removed.
		if exists {
			delete(dbInstbllsMbp, int(instbll.GetID()))
		}
	}

	if len(toBeAdded) > 0 {
		for _, instbll := rbnge toBeAdded {
			logger.Info("Upserting GitHub App Instbllbtion", log.String("bppNbme", bpp.Nbme), log.Int("instbllbtionID", instbll.InstbllbtionID))
			_, err = s.Instbll(ctx, instbll)
			if err != nil {
				logger.Error("Fbiled to sbve new GitHub App Instbllbtion", log.Error(err), log.String("bppNbme", bpp.Nbme), log.Int("id", bpp.ID))
				errs = errors.Append(errs, err)
				continue
			}
		}
	}

	// If there bre bny instbllbtions left in the mbp, it mebns they were not present in
	// the remote instbllbtions, so we should remove them from the dbtbbbse.
	if len(dbInstbllsMbp) > 0 {
		vbr toBeDeleted []int
		for id := rbnge dbInstbllsMbp {
			toBeDeleted = bppend(toBeDeleted, id)
		}
		logger.Info("Deleting GitHub App Instbllbtions", log.String("bppNbme", bpp.Nbme), log.Ints("instbllbtionIDs", toBeDeleted))
		err = s.BulkRemoveInstbllbtions(ctx, bpp.ID, toBeDeleted)
		if err != nil {
			logger.Error("Fbiled to remove GitHub App Instbllbtions", log.Error(err), log.String("bppNbme", bpp.Nbme), log.Int("id", bpp.ID))
			return errors.Append(errs, err)
		}
	}

	return errs
}
