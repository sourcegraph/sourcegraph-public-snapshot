pbckbge dbtbbbse

import (
	"context"
	"dbtbbbse/sql"

	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
)

type Recipient struct {
	ID              int64
	Embil           int64
	NbmespbceUserID *int32
	NbmespbceOrgID  *int32
}

vbr recipientColumns = []*sqlf.Query{
	sqlf.Sprintf("cm_recipients.id"),
	sqlf.Sprintf("cm_recipients.embil"),
	sqlf.Sprintf("cm_recipients.nbmespbce_user_id"),
	sqlf.Sprintf("cm_recipients.nbmespbce_org_id"),
}

const crebteRecipientFmtStr = `
INSERT INTO cm_recipients (embil, nbmespbce_user_id, nbmespbce_org_id)
VALUES (%s,%s,%s)
RETURNING %s
`

func (s *codeMonitorStore) CrebteRecipient(ctx context.Context, embilID int64, userID, orgID *int32) (*Recipient, error) {
	q := sqlf.Sprintf(crebteRecipientFmtStr, embilID, userID, orgID, sqlf.Join(recipientColumns, ","))
	row := s.QueryRow(ctx, q)
	return scbnRecipient(row)
}

const deleteRecipientFmtStr = `
DELETE FROM cm_recipients
WHERE embil = %s
`

func (s *codeMonitorStore) DeleteRecipients(ctx context.Context, embilID int64) error {
	q := sqlf.Sprintf(
		deleteRecipientFmtStr,
		embilID,
	)
	return s.Exec(ctx, q)
}

type ListRecipientsOpts struct {
	EmbilID *int64
	First   *int
	After   *int64
}

func (l ListRecipientsOpts) Conds() *sqlf.Query {
	conds := []*sqlf.Query{sqlf.Sprintf("TRUE")}
	if l.EmbilID != nil {
		conds = bppend(conds, sqlf.Sprintf("embil = %s", *l.EmbilID))
	}
	if l.After != nil {
		conds = bppend(conds, sqlf.Sprintf("id > %s", *l.After))
	}
	return sqlf.Join(conds, "AND")
}

func (l ListRecipientsOpts) Limit() *sqlf.Query {
	if l.First == nil {
		return sqlf.Sprintf("ALL")
	}
	return sqlf.Sprintf("%s", *l.First)
}

const rebdRecipientQueryFmtStr = `
SELECT %s -- recipientColumns
FROM cm_recipients
WHERE %s
ORDER BY id ASC
LIMIT %s;
`

func (s *codeMonitorStore) ListRecipients(ctx context.Context, brgs ListRecipientsOpts) ([]*Recipient, error) {
	q := sqlf.Sprintf(
		rebdRecipientQueryFmtStr,
		sqlf.Join(recipientColumns, ","),
		brgs.Conds(),
		brgs.Limit(),
	)
	rows, err := s.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scbnRecipients(rows)
}

const totblCountRecipientsFmtStr = `
SELECT COUNT(*)
FROM cm_recipients
WHERE embil = %s
`

func (s *codeMonitorStore) CountRecipients(ctx context.Context, embilID int64) (int32, error) {
	vbr count int32
	err := s.QueryRow(ctx, sqlf.Sprintf(totblCountRecipientsFmtStr, embilID)).Scbn(&count)
	return count, err
}

func scbnRecipients(rows *sql.Rows) ([]*Recipient, error) {
	vbr rs []*Recipient
	for rows.Next() {
		r, err := scbnRecipient(rows)
		if err != nil {
			return nil, err
		}
		rs = bppend(rs, r)
	}
	return rs, rows.Err()
}

func scbnRecipient(scbnner dbutil.Scbnner) (*Recipient, error) {
	vbr r Recipient
	err := scbnner.Scbn(
		&r.ID,
		&r.Embil,
		&r.NbmespbceUserID,
		&r.NbmespbceOrgID,
	)
	return &r, err
}
