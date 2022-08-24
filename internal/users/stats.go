package users

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type UsersStatsDateTimeRange struct {
	Before *string
	After  *string
}

type UsersStatsNumberRange struct {
	Min  *float64
	Max  *float64
}

func (d *UsersStatsDateTimeRange) Parse() (before *time.Time, after *time.Time, err error) {
	if d.Before != nil {
		beforeTime, err := time.Parse(time.RFC3339, *d.Before)
		if err != nil {
			return nil, nil, err
		}
		before = &beforeTime
	}
	if d.After != nil {
		afterTime, err := time.Parse(time.RFC3339, *d.After)
		if err != nil {
			return nil, nil, err
		}
		after = &afterTime
	}
	return before, after, nil
}

type UsersStatsNullableDateRange struct {
	UsersStatsDateTimeRange
	IsNull *bool
}

type UsersStatsFilters struct {
	Query        *string
	SiteAdmin    *bool
	Username     *string
	Email        *string
	LastActiveAt *UsersStatsNullableDateRange
	DeletedAt    *UsersStatsNullableDateRange
	CreatedAt    *UsersStatsDateTimeRange
	EventsCount  *UsersStatsNumberRange
}

type UsersStats struct {
	DB      database.DB
	Filters UsersStatsFilters
}

func (s *UsersStats) makeQueryParameters() ([]*sqlf.Query, error) {
	conds := []*sqlf.Query{sqlf.Sprintf("TRUE")}
	if s.Filters.Query != nil && *s.Filters.Query != "" {
		query := "%" + *s.Filters.Query + "%"
		conds = append(conds, sqlf.Sprintf("(username ILIKE %s OR display_name ILIKE %s OR primary_email ILIKE %s)", query, query, query))
	}
	if s.Filters.SiteAdmin != nil {
		conds = append(conds, sqlf.Sprintf("site_admin = %s", *s.Filters.SiteAdmin))
	}
	if s.Filters.Username != nil {
		conds = append(conds, sqlf.Sprintf("username ILIKE %s", "%"+*s.Filters.Username+"%"))
	}
	if s.Filters.Email != nil {
		conds = append(conds, sqlf.Sprintf("primary_email ILIKE %s", "%"+*s.Filters.Email+"%"))
	}
	if s.Filters.DeletedAt != nil {
		if s.Filters.DeletedAt.IsNull != nil && *s.Filters.DeletedAt.IsNull {
			conds = append(conds, sqlf.Sprintf("deleted_at IS NULL"))
		} else {
			before, after, err := s.Filters.DeletedAt.Parse()
			if err != nil {
				return nil, err
			}
			if before != nil {
				conds = append(conds, sqlf.Sprintf("deleted_at <= %s", before))
			}
			if after != nil {
				conds = append(conds, sqlf.Sprintf("deleted_at >= %s", after))
			}
		}
	}

	if s.Filters.LastActiveAt != nil {
		if s.Filters.LastActiveAt.IsNull != nil && *s.Filters.LastActiveAt.IsNull {
			conds = append(conds, sqlf.Sprintf("last_active_at IS NULL"))
		} else {
			before, after, err := s.Filters.LastActiveAt.Parse()
			if err != nil {
				return nil, err
			}
			if before != nil {
				conds = append(conds, sqlf.Sprintf("last_active_at <= %s", before))
			}
			if after != nil {
				conds = append(conds, sqlf.Sprintf("last_active_at >= %s", after))
			}
		}
	}
	if s.Filters.CreatedAt != nil {
		before, after, err := s.Filters.CreatedAt.Parse()
		if err != nil {
			return nil, err
		}
		if before != nil {
			conds = append(conds, sqlf.Sprintf("created_at <= %s", before))
		}
		if after != nil {
			conds = append(conds, sqlf.Sprintf("created_at >= %s", after))
		}
	}

	if s.Filters.EventsCount != nil {
		if s.Filters.EventsCount.Max != nil {
			conds = append(conds, sqlf.Sprintf("events_count <= %s", *s.Filters.EventsCount.Max))
		}
		if s.Filters.EventsCount.Min != nil {
			conds = append(conds, sqlf.Sprintf("events_count >= %s", *s.Filters.EventsCount.Min))
		}
	}
	return conds, nil
}

var (
	statsCTEQuery = `
	WITH aggregated_stats AS (
		SELECT
			users.id AS id,
			users.username,
			users.display_name,
			emails.email primary_email,
			users.created_at,
			stats.user_last_active_at AS last_active_at,
			users.deleted_at,
			users.site_admin,
			COALESCE(stats.user_events_count, 0) AS events_count
		FROM users
			LEFT JOIN aggregated_user_statistics stats ON stats.user_id = users.id
			LEFT JOIN user_emails emails ON emails.user_id = users.id AND emails.is_primary = true
	)
	%s
	`
)

func (s *UsersStats) TotalCount(ctx context.Context) (float64, error) {
	var totalCount float64

	conds, err := s.makeQueryParameters()
	if err != nil {
		return 0, err
	}

	query := sqlf.Sprintf(statsCTEQuery, sqlf.Sprintf(`SELECT COUNT(id) FROM aggregated_stats WHERE %s`, sqlf.Join(conds, "AND")))
	if err := s.DB.QueryRowContext(ctx, query.Query(sqlf.PostgresBindVar), query.Args()...).Scan(&totalCount); err != nil {
		return 0, err
	}

	return totalCount, nil
}

type UsersStatsListUsersFilters struct {
	OrderBy    *string
	Descending *bool
	First      *int32
}

func (s *UsersStats) ListUsers(ctx context.Context, filters *UsersStatsListUsersFilters) ([]*UserStatItem, error) {
	// ORDER BY
	orderDirection := "ASC"
	if filters == nil {
		filters = &UsersStatsListUsersFilters{}
	}
	if filters.Descending != nil && *filters.Descending {
		orderDirection = "DESC"
	}
	orderBy := sqlf.Sprintf("id " + orderDirection)
	if filters.OrderBy != nil {
		newOrderColumn, err := toUsersField(*filters.OrderBy)
		orderBy = sqlf.Sprintf(newOrderColumn + " " + orderDirection)
		if err != nil {
			return nil, err
		}
	}

	// LIMIT
	limit := int32(100)
	if filters.First != nil {
		limit = *filters.First
	}

	conds, err := s.makeQueryParameters()
	if err != nil {
		return nil, err
	}

	query := sqlf.Sprintf(statsCTEQuery, sqlf.Sprintf(`
	SELECT id, username, display_name, primary_email, created_at, last_active_at, deleted_at, site_admin, events_count FROM aggregated_stats WHERE %s ORDER BY %s LIMIT %s`, sqlf.Join(conds, "AND"), orderBy, limit))

	rows, err := s.DB.QueryContext(ctx, query.Query(sqlf.PostgresBindVar), query.Args()...)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	nodes := make([]*UserStatItem, 0)
	for rows.Next() {
		var node UserStatItem

		if err := rows.Scan(&node.Id, &node.Username, &node.DisplayName, &node.PrimaryEmail, &node.CreatedAt, &node.LastActiveAt, &node.DeletedAt, &node.SiteAdmin, &node.EventsCount); err != nil {
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
		return "primary_email", nil
	case "CREATED_AT":
		return "created_at", nil
	case "LAST_ACTIVE_AT":
		return "last_active_at", nil
	case "DELETED_AT":
		return "deleted_at", nil
	case "EVENTS_COUNT":
		return "events_count", nil
	case "SITE_ADMIN":
		return "site_admin", nil
	default:
		return "", errors.New("invalid orderBy")
	}
}

type UserStatItem struct {
	Id           int32
	Username     string
	DisplayName  *string
	PrimaryEmail *string
	CreatedAt    time.Time
	LastActiveAt *time.Time
	DeletedAt    *time.Time
	SiteAdmin    bool
	EventsCount  float64
}
