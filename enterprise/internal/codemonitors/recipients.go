package codemonitors

import (
	"context"
	"database/sql"

	"github.com/cockroachdb/errors"
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

func (s *codeMonitorStore) CreateRecipients(ctx context.Context, recipients []graphql.ID, emailID int64) error {
	for _, r := range recipients {
		err := s.createRecipient(ctx, r, emailID)
		if err != nil {
			return err
		}
	}
	return nil
}

const deleteRecipientFmtStr = `
DELETE FROM cm_recipients
WHERE email = %s
`

func (s *codeMonitorStore) DeleteRecipients(ctx context.Context, emailID int64) error {
	q := sqlf.Sprintf(
		deleteRecipientFmtStr,
		emailID,
	)
	return s.Exec(ctx, q)
}

const readRecipientQueryFmtStr = `
SELECT id, email, namespace_user_id, namespace_org_id
FROM cm_recipients
WHERE email = %s
AND id > %s
ORDER BY id ASC
LIMIT %s;
`

func (s *codeMonitorStore) ListRecipientsForEmailAction(ctx context.Context, emailID int64, args *graphqlbackend.ListRecipientsArgs) ([]*Recipient, error) {
	after, err := unmarshalAfter(args.After)
	if err != nil {
		return nil, err
	}
	q := sqlf.Sprintf(
		readRecipientQueryFmtStr,
		emailID,
		after,
		args.First,
	)
	rows, err := s.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRecipients(rows)
}

const allRecipientsForEmailIDInt64FmtStr = `
SELECT id, email, namespace_user_id, namespace_org_id
FROM cm_recipients
WHERE email = %s
`

func (s *codeMonitorStore) ListAllRecipientsForEmailAction(ctx context.Context, emailID int64) ([]*Recipient, error) {
	rows, err := s.Query(ctx, sqlf.Sprintf(allRecipientsForEmailIDInt64FmtStr, emailID))
	if err != nil {
		return nil, errors.Errorf("store.AllRecipientsForEmailIDInt64: %w", err)
	}
	defer rows.Close()
	return scanRecipients(rows)
}

const createRecipientFmtStr = `
INSERT INTO cm_recipients (email, namespace_user_id, namespace_org_id)
VALUES (%s,%s,%s)
`

func (s *codeMonitorStore) createRecipient(ctx context.Context, recipient graphql.ID, emailID int64) error {
	var userID, orgID int32
	err := graphqlbackend.UnmarshalNamespaceID(recipient, &userID, &orgID)
	if err != nil {
		return err
	}
	return s.Exec(ctx, sqlf.Sprintf(createRecipientFmtStr, emailID, nilOrInt32(userID), nilOrInt32(orgID)))
}

const totalCountRecipientsFmtStr = `
SELECT COUNT(*)
FROM cm_recipients
WHERE email = %s
`

func (s *codeMonitorStore) CountRecipients(ctx context.Context, emailID int64) (int32, error) {
	var count int32
	err := s.QueryRow(ctx, sqlf.Sprintf(totalCountRecipientsFmtStr, emailID)).Scan(&count)
	return count, err
}

func nilOrInt32(n int32) *int32 {
	if n == 0 {
		return nil
	}
	return &n
}

func scanRecipients(rows *sql.Rows) ([]*Recipient, error) {
	var ms []*Recipient
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
	return ms, rows.Err()
}
