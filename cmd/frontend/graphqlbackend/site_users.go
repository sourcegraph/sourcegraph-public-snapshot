package graphqlbackend

import (
	"context"
	"fmt"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (s *siteResolver) Users(ctx context.Context, args *struct {
	Query            *string
	SiteAdmin        *bool
	Username         *string
	Email            *string
	LastActivePeriod *string
}) (*SiteUsersResolver, error) {
	// 🚨 SECURITY: Only site admins can see users.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, s.db); err != nil {
		return nil, err
	}

	conds := []*sqlf.Query{sqlf.Sprintf("TRUE")}
	tables := []*sqlf.Query{sqlf.Sprintf(`users`)}

	if args.Query != nil && *args.Query != "" {
		query := "%" + *args.Query + "%"
		conds = append(conds, sqlf.Sprintf("(username ILIKE %s OR display_name ILIKE %s)", query, query))
	}
	if args.SiteAdmin != nil {
		conds = append(conds, sqlf.Sprintf("site_admin = %s", *args.SiteAdmin))
	}
	if args.Username != nil {
		conds = append(conds, sqlf.Sprintf("username ILIKE %s", "%"+*args.Username+"%"))
	}
	if args.Email != nil {
		tables = append(tables, sqlf.Sprintf("LEFT JOIN user_emails emails ON emails.user_id = users.id AND emails.is_primary = true"))
		conds = append(conds, sqlf.Sprintf("email ILIKE %s", "%"+*args.Email+"%"))
	}
	if args.LastActivePeriod != nil && *args.LastActivePeriod != "ALL" {
		lastActiveStartTime, err := makeLastActiveStartTime(*args.LastActivePeriod)
		if err != nil {
			return nil, err
		}
		tables = append(tables, sqlf.Sprintf("LEFT JOIN event_logs events ON events.user_id = users.id"))
		conds = append(conds, sqlf.Sprintf("events.timestamp >= %s", lastActiveStartTime))
	}
	totalsQuery := sqlf.Sprintf(`SELECT COUNT(DISTINCT users.id) FROM %s WHERE %s`, sqlf.Join(tables, " "), sqlf.Join(conds, "AND"))
	return &SiteUsersResolver{s.db, conds, totalsQuery}, nil
}

func makeLastActiveStartTime(lastActivePeriod string) (time.Time, error) {
	now := time.Now()
	switch lastActivePeriod {
	case "TODAY":
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC), nil
	case "THIS_WEEK":
		return timeutil.StartOfWeek(timeNow().UTC(), 0), nil
	case "THIS_MONTH":
		return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC), nil
	default:
		return now, errors.Newf("invalid lastActivePeriod: %s", lastActivePeriod)
	}
}

type SiteUsersResolver struct {
	db          database.DB
	conds       []*sqlf.Query
	totalsQuery *sqlf.Query
}

func (s *SiteUsersResolver) TotalCount(ctx context.Context) (float64, error) {
	var totalCount float64

	if err := s.db.QueryRowContext(ctx, s.totalsQuery.Query(sqlf.PostgresBindVar), s.totalsQuery.Args()...).Scan(&totalCount); err != nil {
		return 0, err
	}

	return totalCount, nil
}

func (s *SiteUsersResolver) Nodes(ctx context.Context, args *struct {
	OrderBy    *string
	Descending *bool
	First      *int32
}) ([]*SiteUserResolver, error) {
	// ORDER BY
	orderDirection := "ASC"
	if args.Descending != nil && *args.Descending {
		orderDirection = "DESC"
	}
	orderBy := sqlf.Sprintf("id " + orderDirection)
	if args.OrderBy != nil {
		newOrderBy, err := toUsersField(*args.OrderBy)
		orderBy = sqlf.Sprintf(newOrderBy + " " + orderDirection)
		if err != nil {
			return nil, err
		}
	}

	// LIMIT
	limit := int32(100)
	if args.First != nil {
		limit = *args.First
	}

	query := sqlf.Sprintf(`
	SELECT
		users.id,
		users.username,
		emails.email,
		users.created_at,
		MAX(events.timestamp) AS last_active_at,
		users.deleted_at,
		users.site_admin,
		COUNT(events.*) AS events_count
	FROM
		users
		LEFT JOIN event_logs events ON events.user_id = users.id
		LEFT JOIN user_emails emails ON emails.user_id = users.id AND emails.is_primary = true
	WHERE %s
	GROUP BY
		users.id,
		users.username,
		emails.email,
		users.created_at,
		users.deleted_at,
		users.site_admin
	ORDER BY %s
	LIMIT %s`, sqlf.Join(s.conds, "AND"), orderBy, limit)
	fmt.Println(query.Query(sqlf.PostgresBindVar))
	fmt.Println(query.Args())

	rows, err := s.db.QueryContext(ctx, query.Query(sqlf.PostgresBindVar), query.Args()...)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	nodes := make([]*SiteUserResolver, 0)
	for rows.Next() {
		var node SiteUserResolver

		if err := rows.Scan(&node.id, &node.username, &node.email, &node.createdAt, &node.lastActiveAt, &node.deletedAt, &node.siteAdmin, &node.eventsCount); err != nil {
			return nil, err
		}

		nodes = append(nodes, &node)
	}

	return nodes, nil
}

func toUsersField(orderBy string) (string, error) {
	switch orderBy {
	case "USERNAME":
		return "username", nil
	case "EMAIL":
		return "emails.email", nil
	case "CREATED_AT":
		return "users.created_at", nil
	case "LAST_ACTIVE_AT":
		return "last_active_at", nil
	case "DELETED_AT":
		return "users.deleted_at", nil
	case "EVENTS_COUNT":
		return "events_count", nil
	case "SITE_ADMIN":
		return "site_admin", nil
	default:
		return "", errors.New("invalid orderBy")
	}
}

type SiteUserResolver struct {
	id           int32
	username     string
	email        *string
	createdAt    time.Time
	lastActiveAt *time.Time
	deletedAt    *time.Time
	siteAdmin    bool
	eventsCount  float64
}

func (s *SiteUserResolver) ID(ctx context.Context) graphql.ID { return MarshalUserID(s.id) }

func (s *SiteUserResolver) Username(ctx context.Context) string { return s.username }

func (s *SiteUserResolver) Email(ctx context.Context) *string { return s.email }

func (s *SiteUserResolver) CreatedAt(ctx context.Context) string {
	return s.createdAt.Format(time.RFC3339)
}

func (s *SiteUserResolver) LastActiveAt(ctx context.Context) *string {
	if s.lastActiveAt != nil {
		result := s.lastActiveAt.Format(time.RFC3339)
		return &result
	}
	return nil
}

func (s *SiteUserResolver) DeletedAt(ctx context.Context) *string {
	if s.deletedAt != nil {
		result := s.deletedAt.Format(time.RFC3339)
		return &result
	}
	return nil
}

func (s *SiteUserResolver) SiteAdmin(ctx context.Context) bool { return s.siteAdmin }

func (s *SiteUserResolver) EventsCount(ctx context.Context) float64 { return s.eventsCount }
