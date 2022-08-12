package users

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/csv"
	"strconv"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type UsersStatsFilters struct {
	Query            *string
	SiteAdmin        *bool
	Username         *string
	Email            *string
	LastActivePeriod *string
}

type UsersStats struct {
	Cache   bool // TODO: implement caching
	DB      database.DB
	Filters UsersStatsFilters
}

func (s *UsersStats) makeQueryParameters() ([]*sqlf.Query, []*sqlf.Query, error) {
	preConds := []*sqlf.Query{sqlf.Sprintf("TRUE")}
	postConds := []*sqlf.Query{sqlf.Sprintf("TRUE")}
	if s.Filters.Query != nil && *s.Filters.Query != "" {
		query := "%" + *s.Filters.Query + "%"
		preConds = append(preConds, sqlf.Sprintf("(username ILIKE %s OR display_name ILIKE %s)", query, query))
	}
	if s.Filters.SiteAdmin != nil {
		preConds = append(preConds, sqlf.Sprintf("site_admin = %s", *s.Filters.SiteAdmin))
	}
	if s.Filters.Username != nil {
		preConds = append(preConds, sqlf.Sprintf("username ILIKE %s", "%"+*s.Filters.Username+"%"))
	}
	if s.Filters.Email != nil {
		preConds = append(preConds, sqlf.Sprintf("email ILIKE %s", "%"+*s.Filters.Email+"%"))
	}
	if s.Filters.LastActivePeriod != nil && *s.Filters.LastActivePeriod != "ALL" {
		lastActiveStartTime, err := makeLastActiveStartTime(*s.Filters.LastActivePeriod)
		if err != nil {
			return nil, nil, err
		}
		postConds = append(preConds, sqlf.Sprintf("last_active_at >= %s", lastActiveStartTime))
	}
	return preConds, postConds, nil
}

func makeLastActiveStartTime(lastActivePeriod string) (time.Time, error) {
	now := time.Now()
	switch lastActivePeriod {
	case "TODAY":
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC), nil
	case "THIS_WEEK":
		return timeutil.StartOfWeek(now.UTC(), 0), nil
	case "THIS_MONTH":
		return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC), nil
	default:
		return now, errors.Newf("invalid lastActivePeriod: %s", lastActivePeriod)
	}
}

var (
	statsSubQuery = `
	WITH stats AS (
		SELECT
			users.id,
			users.username,
			display_name,
			(SELECT email FROM user_emails WHERE user_id = users.id AND is_primary = true LIMIT 1) AS email,
			users.created_at,
			(SELECT timestamp FROM event_logs WHERE user_id = users.id ORDER BY timestamp DESC LIMIT 1) AS last_active_at,
			users.deleted_at,
			users.site_admin,
			(SELECT COUNT(id) FROM event_logs WHERE user_id = users.id) AS events_count
		FROM
			users
			LEFT JOIN user_emails emails ON emails.user_id = users.id AND emails.is_primary = true
		WHERE %s
	)
	%s
	`
)

func (s *UsersStats) TotalCount(ctx context.Context) (float64, error) {
	var totalCount float64

	preConds, postConds, err := s.makeQueryParameters()
	if err != nil {
		return 0, err
	}

	query := sqlf.Sprintf(statsSubQuery, sqlf.Join(preConds, "AND"), sqlf.Sprintf(`SELECT COUNT(DISTINCT id) FROM stats WHERE %s`, sqlf.Join(postConds, "AND")))
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

	preConds, postConds, err := s.makeQueryParameters()
	if err != nil {
		return nil, err
	}

	query := sqlf.Sprintf(statsSubQuery, sqlf.Join(preConds, "AND"), sqlf.Sprintf(`
	SELECT id, username, email, created_at, last_active_at, deleted_at, site_admin, events_count FROM stats
	WHERE %s
	ORDER BY %s
	LIMIT %s
	`, sqlf.Join(postConds, "AND"), orderBy, limit))

	rows, err := s.DB.QueryContext(ctx, query.Query(sqlf.PostgresBindVar), query.Args()...)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	nodes := make([]*UserStatItem, 0)
	for rows.Next() {
		var node UserStatItem

		if err := rows.Scan(&node.Id, &node.Username, &node.Email, &node.CreatedAt, &node.LastActiveAt, &node.DeletedAt, &node.SiteAdmin, &node.EventsCount); err != nil {
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
		return "email", nil
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
	Email        *string
	CreatedAt    time.Time
	LastActiveAt *time.Time
	DeletedAt    *time.Time
	SiteAdmin    bool
	EventsCount  float64
}

// GetArchive generates and returns a usage statistics ZIP archive containing the CSV
// files defined in RFC 145, or an error in case of failure.
func (s *UsersStats) GetArchive(ctx context.Context) ([]byte, error) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	file, err := zw.Create("UsersStats.csv")
	if err != nil {
		return nil, err
	}

	writer := csv.NewWriter(file)

	record := []string{
		"user_id",
		"created_at",
		"events_count",
		"last_active_at",
		"deleted_at",
	}

	if err := writer.Write(record); err != nil {
		return nil, err
	}

	users, err := s.ListUsers(ctx, nil)
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		record[0] = strconv.FormatUint(uint64(user.Id), 10)
		record[1] = user.CreatedAt.Format(time.RFC3339)
		record[2] = strconv.FormatInt(int64(user.EventsCount), 10)
		if user.LastActiveAt == nil {
			record[3] = "NULL"
		} else {
			record[3] = user.LastActiveAt.Format(time.RFC3339)
		}
		if user.DeletedAt == nil {
			record[4] = "NULL"
		} else {
			record[4] = user.DeletedAt.Format(time.RFC3339)
		}
		if err := writer.Write(record); err != nil {
			return nil, err
		}
	}

	writer.Flush()

	if err := zw.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
