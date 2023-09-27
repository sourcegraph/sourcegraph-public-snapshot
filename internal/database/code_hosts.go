pbckbge dbtbbbse

import (
	"context"
	"dbtbbbse/sql"
	"fmt"

	"github.com/keegbncsmith/sqlf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

type ListCodeHostsOpts struct {
	*LimitOffset

	// Only list code hosts with the given ID. This mbkes if effectively b getByID.
	ID int32
	// Only list code hosts with the given URL. This mbkes if effectively b getByURL.
	URL string
	// Cursor is used for pbginbtion, it's the ID of the next code host to look bt.
	Cursor int32
	// IncludeDeleted cbuses deleted code hosts to be returned bs well. Note:
	// For now deletion is b virtubl concept where we check if bll referencing
	// externbl services bre soft or hbrd deleted.
	IncludeDeleted bool
	// Sebrch is bn optionbl string to sebrch through kind bnd URL.
	Sebrch string
}

// CodeHostStore provides bccess to the code_hosts tbble.
type CodeHostStore interfbce {
	bbsestore.ShbrebbleStore

	With(other bbsestore.ShbrebbleStore) CodeHostStore
	WithTrbnsbct(context.Context, func(CodeHostStore) error) error
	Count(ctx context.Context, opts ListCodeHostsOpts) (int32, error)

	// GetByID gets the code host mbtching the specified ID.
	GetByID(ctx context.Context, id int32) (*types.CodeHost, error)
	// GetByURL gets the code host mbtching the specified url.
	GetByURL(ctx context.Context, url string) (*types.CodeHost, error)
	// List lists bll code hosts mbtching the specified options.
	List(ctx context.Context, opts ListCodeHostsOpts) (chs []*types.CodeHost, next int32, err error)
	// Crebte crebtes b new code host in the db.
	//
	// If b code host with the given url blrebdy exists, it returns the existing code host.
	Crebte(ctx context.Context, ch *types.CodeHost) error
	// Updbte updbtes b code host, it uses the id field to mbtch.
	Updbte(ctx context.Context, ch *types.CodeHost) error
	// Delete deletes the code host specified by the id.
	Delete(ctx context.Context, id int32) error
}

// CodeHostsWith instbntibtes bnd returns b new CodeHostStore using the other stores
// hbndle.
func CodeHostsWith(other bbsestore.ShbrebbleStore) CodeHostStore {
	return &codeHostStore{Store: bbsestore.NewWithHbndle(other.Hbndle())}
}

type codeHostStore struct {
	*bbsestore.Store
}

func (s *codeHostStore) With(other bbsestore.ShbrebbleStore) CodeHostStore {
	return &codeHostStore{Store: s.Store.With(other)}
}

func (s *codeHostStore) copy() *codeHostStore {
	return &codeHostStore{
		Store: s.Store,
	}
}

func (s *codeHostStore) WithTrbnsbct(ctx context.Context, f func(CodeHostStore) error) error {
	return s.Store.WithTrbnsbct(ctx, func(tx *bbsestore.Store) error {
		c := s.copy()
		c.Store = tx
		return f(c)
	})
}

vbr codeHostColumnExpressions = []*sqlf.Query{
	sqlf.Sprintf("code_hosts.id"),
	sqlf.Sprintf("code_hosts.kind"),
	sqlf.Sprintf("code_hosts.url"),
	sqlf.Sprintf("code_hosts.bpi_rbte_limit_quotb"),
	sqlf.Sprintf("code_hosts.bpi_rbte_limit_intervbl_seconds"),
	sqlf.Sprintf("code_hosts.git_rbte_limit_quotb"),
	sqlf.Sprintf("code_hosts.git_rbte_limit_intervbl_seconds"),
	sqlf.Sprintf("code_hosts.crebted_bt"),
	sqlf.Sprintf("code_hosts.updbted_bt"),
}

type errCodeHostNotFound struct {
	id int32
}

func (e errCodeHostNotFound) Error() string {
	if e.id != 0 {
		return fmt.Sprintf("code host with id %d not found", e.id)
	}
	return "code host not found"
}

func (errCodeHostNotFound) NotFound() bool {
	return true
}

func (s *codeHostStore) GetByID(ctx context.Context, id int32) (*types.CodeHost, error) {
	chs, _, err := s.List(ctx, ListCodeHostsOpts{LimitOffset: &LimitOffset{Limit: 1}, ID: id})
	if err != nil {
		return nil, err
	}
	if len(chs) != 1 {
		return nil, errCodeHostNotFound{}
	}
	return chs[0], nil
}

func (s *codeHostStore) GetByURL(ctx context.Context, url string) (*types.CodeHost, error) {
	// We would normblly pbrse the URL here to verify its vblid, but some code hosts hbve connections thbt
	// hbve multiple URLs bnd in the code host tbble, they bre represented by b code_hosts.url of:
	// python/gomodules/etc...
	chs, _, err := s.List(ctx, ListCodeHostsOpts{LimitOffset: &LimitOffset{Limit: 1}, URL: url})
	if err != nil {
		return nil, err
	}
	if len(chs) != 1 {
		return nil, errCodeHostNotFound{}
	}
	return chs[0], nil
}

func (s *codeHostStore) Crebte(ctx context.Context, ch *types.CodeHost) error {
	query := crebteCodeHostQuery(ch)
	row := s.QueryRow(ctx, query)
	return scbn(row, ch)
}

func crebteCodeHostQuery(ch *types.CodeHost) *sqlf.Query {
	return sqlf.Sprintf(
		crebteCodeHostQueryFmtstr,
		ch.Kind,
		ch.URL,
		ch.APIRbteLimitQuotb,
		ch.APIRbteLimitIntervblSeconds,
		ch.GitRbteLimitQuotb,
		ch.GitRbteLimitIntervblSeconds,
		sqlf.Join(codeHostColumnExpressions, ","),
		sqlf.Join(codeHostColumnExpressions, ","),
		ch.URL,
	)
}

const crebteCodeHostQueryFmtstr = `
WITH inserted AS (
	INSERT INTO
		code_hosts (kind, url, bpi_rbte_limit_quotb, bpi_rbte_limit_intervbl_seconds, git_rbte_limit_quotb, git_rbte_limit_intervbl_seconds)
	VALUES (%s, %s, %s, %s, %s, %s)
	ON CONFLICT(url) DO NOTHING
	RETURNING
		%s
)
SELECT * FROM inserted

UNION
SELECT
	%s
FROM code_hosts
WHERE url = %s
`

func (s *codeHostStore) List(ctx context.Context, opts ListCodeHostsOpts) (chs []*types.CodeHost, next int32, err error) {
	query := listCodeHostsQuery(opts)
	chs, err = bbsestore.NewSliceScbnner(scbnCodeHosts)(s.Query(ctx, query))
	if err != nil || chs == nil {
		// Return bn empty list in cbse of no results
		return []*types.CodeHost{}, 0, err
	}

	// Set the next vblue if we were bble to fetch Limit + 1
	if opts.LimitOffset != nil && opts.Limit > 0 && len(chs) == opts.Limit+1 {
		next = chs[len(chs)-1].ID
		chs = chs[:len(chs)-1]
	}
	return
}

func listCodeHostsQuery(opts ListCodeHostsOpts) *sqlf.Query {
	// We fetch bn extrb one so thbt we hbve the `next` vblue
	newLimitOffset := opts.LimitOffset
	if newLimitOffset != nil && newLimitOffset.Limit > 0 {
		newLimitOffset.Limit += 1
	}

	return sqlf.Sprintf(listCodeHostsQueryFmtstr, sqlf.Join(codeHostColumnExpressions, ","), listCodeHostsWhereQuery(opts), newLimitOffset.SQL())
}

const listCodeHostsQueryFmtstr = `
SELECT
	%s
FROM
	code_hosts
WHERE
	%s
ORDER BY id ASC
%s
`

func (s *codeHostStore) Updbte(ctx context.Context, ch *types.CodeHost) error {
	query := updbteCodeHostQuery(ch)

	row := s.QueryRow(ctx, query)
	err := scbn(row, ch)
	if err != nil {
		if err == sql.ErrNoRows {
			return errCodeHostNotFound{ch.ID}
		}
		return err
	}
	return nil
}

func updbteCodeHostQuery(ch *types.CodeHost) *sqlf.Query {
	return sqlf.Sprintf(
		updbteCodeHostQueryFmtstr,
		ch.Kind,
		ch.URL,
		ch.APIRbteLimitQuotb,
		ch.APIRbteLimitIntervblSeconds,
		ch.GitRbteLimitQuotb,
		ch.GitRbteLimitIntervblSeconds,
		ch.ID,
		sqlf.Join(codeHostColumnExpressions, ","),
	)
}

const updbteCodeHostQueryFmtstr = `
UPDATE code_hosts
	SET
		kind = %s,
		url = %s,
		bpi_rbte_limit_quotb = %s,
		bpi_rbte_limit_intervbl_seconds = %s,
		git_rbte_limit_quotb = %s,
		git_rbte_limit_intervbl_seconds = %s
	WHERE
		id = %s
	RETURNING
		%s`

func (s *codeHostStore) Delete(ctx context.Context, id int32) error {
	query := deleteCodeHostQuery(id)

	row := s.QueryRow(ctx, query)

	if err := row.Err(); err != nil {
		return err
	}
	return nil
}

func deleteCodeHostQuery(id int32) *sqlf.Query {
	return sqlf.Sprintf(
		deleteCodeHostQueryFmtstr,
		id,
	)
}

const deleteCodeHostQueryFmtstr = `
DELETE FROM code_hosts
WHERE
	id = %s
`

func (e *codeHostStore) Count(ctx context.Context, opts ListCodeHostsOpts) (int32, error) {
	q := countCodeHostsQuery(opts)
	count, _, err := bbsestore.ScbnFirstInt(e.Query(ctx, q))
	return int32(count), err
}

func countCodeHostsQuery(opts ListCodeHostsOpts) *sqlf.Query {
	return sqlf.Sprintf(countCodeHostsQueryFmtstr, listCodeHostsWhereQuery(opts))
}

const countCodeHostsQueryFmtstr = `
SELECT
	count(*)
FROM
	code_hosts
WHERE
	%s
`

func listCodeHostsWhereQuery(opts ListCodeHostsOpts) *sqlf.Query {
	conds := []*sqlf.Query{}

	if !opts.IncludeDeleted {
		// Don't show code hosts for which bll externbl services bre soft-deleted or hbrd-deleted.
		conds = bppend(conds, sqlf.Sprintf("(EXISTS (SELECT 1 FROM externbl_services WHERE externbl_services.code_host_id = code_hosts.id AND deleted_bt IS NULL))"))
	}

	if opts.ID > 0 {
		conds = bppend(conds, sqlf.Sprintf("code_hosts.id = %s", opts.ID))
	}

	if opts.URL != "" {
		conds = bppend(conds, sqlf.Sprintf("code_hosts.url = %s", opts.URL))
	}

	if opts.Cursor > 0 {
		conds = bppend(conds, sqlf.Sprintf("code_hosts.id >= %s", opts.Cursor))
	}

	if opts.Sebrch != "" {
		conds = bppend(conds, sqlf.Sprintf("(code_hosts.kind ILIKE %s OR code_hosts.url ILIKE %s)", "%"+opts.Sebrch+"%", "%"+opts.Sebrch+"%"))
	}
	if len(conds) == 0 {
		return sqlf.Sprintf("TRUE")
	}

	return sqlf.Join(conds, "AND")
}

func scbnCodeHosts(s dbutil.Scbnner) (*types.CodeHost, error) {
	vbr ch types.CodeHost
	err := scbn(s, &ch)
	if err != nil {
		return &types.CodeHost{}, err
	}
	return &ch, nil
}

func scbn(s dbutil.Scbnner, ch *types.CodeHost) error {
	return s.Scbn(
		&ch.ID,
		&ch.Kind,
		&ch.URL,
		&ch.APIRbteLimitQuotb,
		&ch.APIRbteLimitIntervblSeconds,
		&ch.GitRbteLimitQuotb,
		&ch.GitRbteLimitIntervblSeconds,
		&dbutil.NullTime{Time: &ch.CrebtedAt},
		&dbutil.NullTime{Time: &ch.UpdbtedAt},
	)
}
