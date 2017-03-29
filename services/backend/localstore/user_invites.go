package localstore

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"context"

	"github.com/lib/pq"
	"gopkg.in/gorp.v1"
	"gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph/legacyerr"
)

func init() {
	AppSchema.Map.AddTableWithName(dbUserInvite{}, "user_invite").SetKeys(true, "ID")
	AppSchema.CreateSQL = append(AppSchema.CreateSQL,
		"CREATE UNIQUE INDEX user_invite_unique ON user_invite(user_id, org_id);",
		`ALTER TABLE user_invite ALTER COLUMN sent_at TYPE timestamp with time zone USING sent_at::timestamp with time zone;`,
	)
}

// dbUserInvite DB-maps a sourcegraph.UserInvite object.
type dbUserInvite struct {
	URI       string
	ID        int32
	UserID    string    `db:"user_id"`
	UserEmail string    `db:"user_email"`
	OrgID     string    `db:"org_id"`
	OrgName   string    `db:"org_name"`
	SentAt    time.Time `db:"sent_at"`
}

func (r *dbUserInvite) toUserInvite() *sourcegraph.UserInvite {
	r2 := &sourcegraph.UserInvite{
		URI:       r.URI,
		UserID:    r.UserID,
		OrgName:   r.OrgName,
		UserEmail: r.UserEmail,
		OrgID:     r.OrgID,
	}

	r2.SentAt = &r.SentAt

	return r2
}

func (r *dbUserInvite) fromUserInvite(r2 *sourcegraph.UserInvite) {
	r.URI = r2.URI
	r.OrgName = r2.OrgName
	r.UserID = r2.UserID
	r.OrgID = r2.OrgID
	r.UserEmail = r2.UserEmail
	if r2.SentAt != nil {
		r.SentAt = *r2.SentAt
	}
}

func toUserInvites(rs []*dbUserInvite) []*sourcegraph.UserInvite {
	r2s := make([]*sourcegraph.UserInvite, len(rs))
	for i, r := range rs {
		r2s[i] = r.toUserInvite()
	}
	return r2s
}

// userInvites is a DB-backed implementation of the UserInvites store.
type userInvites struct{}

type UserInviteListOp struct {
	// Query specifies a search query for invites. If specified, then the Sort and
	// Direction options are ignored
	Query string

	// URIs filters the list of invites to the unique identifer of the invite
	URIs []string

	// UserID filters the list of invites sent to the UserID
	UserID string

	// OrgID filters the list of invites by the organization the invite was sent to
	OrgID string

	// UserEmail filters the list of invites by the email address the invite was sent to
	UserEmail string

	// OrgName filters the list of invites by the organization name the invite was sent to
	OrgName string

	sourcegraph.ListOptions
}

// GetByURI returns metadata for the request  user invites URI. See the
// documentation for UserInvites.Get for the contract on the freshness of
// the data returned.
func (s *userInvites) GetByURI(ctx context.Context, uri string) (*sourcegraph.UserInvite, error) {
	invite, err := s.getByURI(ctx, uri)
	if err != nil {
		return nil, err
	}

	return invite, nil
}

func (s *userInvites) getByURI(ctx context.Context, uri string) (*sourcegraph.UserInvite, error) {
	invite, err := s.getBySQL(ctx, "uri=$1", uri)
	if err != nil {
		if legacyerr.ErrCode(err) == legacyerr.NotFound {
			// Overwrite with error message containing UserInvite URI.
			err = legacyerr.Errorf(legacyerr.NotFound, "%s: %s", err, uri)
		}
	}
	return invite, err
}

// getBySQL returns an UserInvite matching the SQL query, if any
// exists. A "LIMIT 1" clause is appended to the query before it is
// executed.
func (s *userInvites) getBySQL(ctx context.Context, query string, args ...interface{}) (*sourcegraph.UserInvite, error) {
	var invite dbUserInvite
	if err := appDBH(ctx).SelectOne(&invite, "SELECT * FROM user_invite WHERE ("+query+") LIMIT 1", args...); err == sql.ErrNoRows {
		return nil, legacyerr.Errorf(legacyerr.NotFound, "invite not found")
	} else if err != nil {
		return nil, err
	}
	return invite.toUserInvite(), nil
}

func (s *userInvites) List(ctx context.Context, opt *UserInviteListOp) ([]*sourcegraph.UserInvite, error) {
	if opt == nil {
		opt = &UserInviteListOp{}
	}

	sql, args, err := userInvitesListSQL(opt)
	if err != nil {
		return nil, err
	}
	invites, err := s.query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}

	return invites, nil
}

var errOptionsEmptyResult = errors.New("pgsql: options empty result set")

// userInvitesListSQL translates the options struct to the SQL for querying
// PosgreSQL.
func userInvitesListSQL(opt *UserInviteListOp) (string, []interface{}, error) {
	var selectSQL, fromSQL, whereSQL, orderBySQL string

	var args []interface{}
	arg := func(a interface{}) string {
		v := gorp.PostgresDialect{}.BindVar(len(args))
		args = append(args, a)
		return v
	}

	queryTerms := strings.Fields(opt.Query)
	uriQuery := strings.ToLower(strings.Join(queryTerms, "/"))
	{ // SELECT
		selectSQL = "user_invite.*"
	}
	{ // FROM
		fromSQL = "user_invite"
	}
	{ // WHERE
		var conds []string

		if len(opt.URIs) > 0 {
			if len(opt.URIs) == 1 && strings.Contains(opt.URIs[0], ",") {
				opt.URIs = strings.Split(opt.URIs[0], ",")
			}

			uriBindVars := make([]string, len(opt.URIs))
			for i, uri := range opt.URIs {
				uriBindVars[i] = arg(uri)
			}
			conds = append(conds, "uri IN ("+strings.Join(uriBindVars, ",")+")")
		}
		if len(queryTerms) >= 1 {
			uriQuery = strings.ToLower(uriQuery)
			conds = append(conds, "lower(uri) LIKE "+arg("/"+uriQuery+"%")+" OR lower(uri) LIKE "+arg(uriQuery+"%/%")+" OR lower(name) LIKE "+arg(uriQuery+"%")+" OR lower(uri) = "+arg(uriQuery))
		}
		if opt.UserID != "" {
			conds = append(conds, `lower(user_id)=`+arg(strings.ToLower(opt.UserID)))
		}
		if opt.UserEmail != "" {
			conds = append(conds, `lower(user_email)=`+arg(strings.ToLower(opt.UserEmail)))
		}
		if opt.OrgID != "" {
			conds = append(conds, `lower(org_id)=`+arg(strings.ToLower(opt.OrgID)))
		}
		if opt.OrgName != "" {
			conds = append(conds, `lower(org_name)=`+arg(strings.ToLower(opt.OrgName)))
		}

		if conds != nil {
			whereSQL = "(" + strings.Join(conds, ") AND (") + ")"
		} else {
			whereSQL = "true"
		}
	}

	// ORDER BY
	orderBySQL = fmt.Sprintf("user_email NULLS LAST")

	// LIMIT
	limitSQL := arg(opt.Limit())
	offsetSQL := arg(opt.Offset())

	sql := fmt.Sprintf(`SELECT %s FROM %s WHERE %s ORDER BY %s LIMIT %s OFFSET %s`, selectSQL, fromSQL, whereSQL, orderBySQL, limitSQL, offsetSQL)
	return sql, args, nil
}

func (s *userInvites) query(ctx context.Context, sql string, args ...interface{}) ([]*sourcegraph.UserInvite, error) {
	var invites []*dbUserInvite
	if _, err := appDBH(ctx).Select(&invites, sql, args...); err != nil {
		return nil, err
	}
	return toUserInvites(invites), nil
}

func (s *userInvites) Create(ctx context.Context, newInvite *sourcegraph.UserInvite) error {
	if invite, err := s.getByURI(ctx, newInvite.URI); err == nil {
		// Business Logic / TODO: Abstract invites out into remind vs invite. For now we will update instead of error if the invite is outside the send threshold.
		if newInvite.SentAt.Unix()-invite.SentAt.Unix() > 259200 {
			err := s.RemindInvite(ctx, invite)
			if err != nil {
				return legacyerr.Errorf(legacyerr.AlreadyExists, "invite failed to update exists: %s", invite.URI)
			}
			return nil
		}

		return legacyerr.Errorf(legacyerr.AlreadyExists, "invite already exists: %s", newInvite.URI)
	}
	var r dbUserInvite
	r.fromUserInvite(newInvite)
	err := appDBH(ctx).Insert(&r)
	if isPQErrorUniqueViolation(err) {
		if c := err.(*pq.Error).Constraint; c != "user_invite_unique" {
			log15.Warn("Expected unique_violation of user_invite_unique constraint, but it was something else; did it change?", "constraint", c, "err", err)
		}
		return legacyerr.Errorf(legacyerr.AlreadyExists, "invite already exists: %s", newInvite.URI)
	}
	return err
}

func (s *userInvites) RemindInvite(ctx context.Context, op *sourcegraph.UserInvite) error {
	var args []interface{}
	arg := func(a interface{}) string {
		v := gorp.PostgresDialect{}.BindVar(len(args))
		args = append(args, a)
		return v
	}

	var updates []string
	if op.URI != "" {
		updates = append(updates, `"uri"=`+arg(op.URI))
	}
	if op.UserID != "" {
		updates = append(updates, `"user_id"=`+arg(op.UserID))
	}
	if op.OrgID != "" {
		updates = append(updates, `"org_id"=`+arg(op.OrgID))
	}
	if op.UserEmail != "" {
		updates = append(updates, `"user_email"=`+arg(op.UserEmail))
	}
	if op.SentAt != nil {
		updates = append(updates, `"sent_at"=`+arg(time.Now()))
	}

	if len(updates) > 0 {
		sql := `UPDATE user_invite SET ` + strings.Join(updates, ", ") + ` WHERE uri=` + arg(op.URI)
		_, err := appDBH(ctx).Exec(sql, args...)
		return err
	}
	return nil
}
