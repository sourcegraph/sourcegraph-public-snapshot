package codemonitors

import (
	"context"
	"database/sql"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

type Recipient struct {
	ID              int64
	Email           int64
	NamespaceUserID *int32
	NamespaceOrgID  *int32
}

const createRecipientFmtStr = `
INSERT INTO cm_recipients (email, namespace_user_id, namespace_org_id)
VALUES (%s,%s,%s)
`

func (s *codeMonitorStore) CreateRecipient(ctx context.Context, emailID int64, userID, orgID *int32) error {
	return s.Exec(ctx, sqlf.Sprintf(createRecipientFmtStr, emailID, userID, orgID))
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

type ListRecipientsOpts struct {
	EmailID *int64
	First   *int
	After   *int64
}

func (l ListRecipientsOpts) Conds() *sqlf.Query {
	conds := []*sqlf.Query{sqlf.Sprintf("TRUE")}
	if l.EmailID != nil {
		conds = append(conds, sqlf.Sprintf("email = %s", *l.EmailID))
	}
	if l.After != nil {
		conds = append(conds, sqlf.Sprintf("id > %s", *l.After))
	}
	return sqlf.Join(conds, "AND")
}

func (l ListRecipientsOpts) Limit() *sqlf.Query {
	if l.First == nil {
		return sqlf.Sprintf("ALL")
	}
	return sqlf.Sprintf("%s", *l.First)
}

const readRecipientQueryFmtStr = `
SELECT id, email, namespace_user_id, namespace_org_id
FROM cm_recipients
WHERE %s
ORDER BY id ASC
LIMIT %s;
`

func (s *codeMonitorStore) ListRecipients(ctx context.Context, args ListRecipientsOpts) ([]*Recipient, error) {
	q := sqlf.Sprintf(
		readRecipientQueryFmtStr,
		args.Conds(),
		args.Limit(),
	)
	rows, err := s.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRecipients(rows)
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

func scanRecipients(rows *sql.Rows) ([]*Recipient, error) {
	var rs []*Recipient
	for rows.Next() {
		r, err := scanRecipient(rows)
		if err != nil {
			return nil, err
		}
		rs = append(rs, r)
	}
	return rs, rows.Err()
}

func scanRecipient(scanner dbutil.Scanner) (*Recipient, error) {
	var r Recipient
	err := scanner.Scan(
		&r.ID,
		&r.Email,
		&r.NamespaceUserID,
		&r.NamespaceOrgID,
	)
	return &r, err
}
