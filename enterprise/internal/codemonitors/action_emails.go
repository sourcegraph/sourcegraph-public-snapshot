package codemonitors

import (
	"database/sql"
	"time"

	"github.com/keegancsmith/sqlf"
)

type MonitorEmail struct {
	Id        int64
	Monitor   int64
	Enabled   bool
	Priority  string
	Header    string
	CreatedBy int32
	CreatedAt time.Time
	ChangedBy int32
	ChangedAt time.Time
}

const getAllEmailActionsForTriggerQueryIDInt64FmtStr = `
SELECT e.id, e.monitor, e.enabled, e.priority, e.header, e.created_by, e.created_at, e.changed_by, e.changed_at
FROM cm_emails e INNER JOIN cm_queries q ON e.monitor = q.monitor
WHERE q.id = %s
`

var EmailsColumns = []*sqlf.Query{
	sqlf.Sprintf("cm_emails.id"),
	sqlf.Sprintf("cm_emails.monitor"),
	sqlf.Sprintf("cm_emails.enabled"),
	sqlf.Sprintf("cm_emails.priority"),
	sqlf.Sprintf("cm_emails.header"),
	sqlf.Sprintf("cm_emails.created_by"),
	sqlf.Sprintf("cm_emails.created_at"),
	sqlf.Sprintf("cm_emails.changed_by"),
	sqlf.Sprintf("cm_emails.changed_at"),
}

func ScanEmails(rows *sql.Rows) (ms []*MonitorEmail, err error) {
	for rows.Next() {
		m := &MonitorEmail{}
		if err = rows.Scan(
			&m.Id,
			&m.Monitor,
			&m.Enabled,
			&m.Priority,
			&m.Header,
			&m.CreatedBy,
			&m.CreatedAt,
			&m.ChangedBy,
			&m.ChangedAt,
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
