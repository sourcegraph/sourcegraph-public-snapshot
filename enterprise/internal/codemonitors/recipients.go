package codemonitors

import (
	"context"
	"database/sql"
	"strings"

	"github.com/graph-gophers/graphql-go"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

type Recipient struct {
	ID              int64
	Email           int64
	NamespaceUserID *int32
	NamespaceOrgID  *int32
}

var recipientsColumns = []*sqlf.Query{
	sqlf.Sprintf("cm_recipients.id"),
	sqlf.Sprintf("cm_recipients.email"),
	sqlf.Sprintf("cm_recipients.namespace_user_id"),
	sqlf.Sprintf("cm_recipients.namespace_org_id"),
}

func (s *Store) CreateRecipients(ctx context.Context, recipients []graphql.ID, monitorID int64) (err error) {
	var q *sqlf.Query
	q, err = createRecipientsQuery(ctx, recipients, monitorID)
	if err != nil {
		return err
	}
	err = s.Exec(ctx, q)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) DeleteRecipients(ctx context.Context, emailID int64) (err error) {
	var q *sqlf.Query
	q, err = deleteRecipientsQuery(ctx, emailID)
	if err != nil {
		return err
	}
	err = s.Exec(ctx, q)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) RecipientsForEmailIDInt64(ctx context.Context, emailID int64, args *graphqlbackend.ListRecipientsArgs) ([]*Recipient, error) {
	q, err := s.ReadRecipientQuery(ctx, emailID, args)
	if err != nil {
		return nil, err
	}
	rows, err := s.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ms, err := ScanRecipients(rows)
	if err != nil {
		return nil, err
	}
	return ms, nil
}

const totalCountRecipientsFmtStr = `
SELECT COUNT(*)
FROM cm_recipients
WHERE email = %s
`

func (s *Store) TotalCountRecipients(ctx context.Context, emailID int64) (count int32, err error) {
	err = s.QueryRow(ctx, sqlf.Sprintf(totalCountRecipientsFmtStr, emailID)).Scan(&count)
	return count, err
}

func ScanRecipients(rows *sql.Rows) (ms []*Recipient, err error) {
	for rows.Next() {
		m := &Recipient{}
		if err := rows.Scan(
			&m.ID,
			&m.Email,
			&m.NamespaceUserID,
			&m.NamespaceOrgID,
		); err != nil {
			return nil, err
		}
		ms = append(ms, m)
	}
	err = rows.Close()
	if err != nil {
		return nil, err
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return ms, nil
}

// CreateRecipientsQuery returns a query that inserts several recipients at once.
func createRecipientsQuery(ctx context.Context, namespaces []graphql.ID, emailID int64) (*sqlf.Query, error) {
	const header = `
INSERT INTO cm_recipients (email, namespace_user_id, namespace_org_id)
VALUES`
	const values = `
(%s,%s,%s),`
	var (
		userID        int32
		orgID         int32
		combinedQuery string
		args          []interface{}
	)
	combinedQuery = header
	for range namespaces {
		combinedQuery += values
	}
	combinedQuery = strings.TrimSuffix(combinedQuery, ",") + ";"
	for _, ns := range namespaces {
		err := graphqlbackend.UnmarshalNamespaceID(ns, &userID, &orgID)
		if err != nil {
			return nil, err
		}
		args = append(args, emailID, nilOrInt32(userID), nilOrInt32(orgID))
	}
	return sqlf.Sprintf(
		combinedQuery,
		args...,
	), nil
}

const deleteRecipientFmtStr = `DELETE FROM cm_recipients WHERE email = %s`

func deleteRecipientsQuery(ctx context.Context, emailId int64) (*sqlf.Query, error) {
	return sqlf.Sprintf(
		deleteRecipientFmtStr,
		emailId,
	), nil
}

const readRecipientQueryFmtStr = `
SELECT id, email, namespace_user_id, namespace_org_id
FROM cm_recipients
WHERE email = %s
AND id > %s
ORDER BY id ASC
LIMIT %s;
`

func (s *Store) ReadRecipientQuery(ctx context.Context, emailId int64, args *graphqlbackend.ListRecipientsArgs) (*sqlf.Query, error) {
	after, err := unmarshalAfter(args.After)
	if err != nil {
		return nil, err
	}
	return sqlf.Sprintf(
		readRecipientQueryFmtStr,
		emailId,
		after,
		args.First,
	), nil
}

func nilOrInt32(n int32) *int32 {
	if n == 0 {
		return nil
	}
	return &n
}
