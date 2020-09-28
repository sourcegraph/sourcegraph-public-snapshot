package db

import (
	"context"
	"database/sql"
	"strings"

	"github.com/keegancsmith/sqlf"
	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/internal/db/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/secret"
)

// SecretColumn contains the name of the secret column and whether it is NULLABLE.
type SecretColumn struct {
	Name     string
	Nullable bool
}

// TableRotateEncryption looks for rows in given table that have secret columns not encrypted with
// the current primary key, then decrypts (if it was encrypted by a deprecated key) and (re-)encrypts
// those column values with the current primary key.
//
// Example:
//	TableRotateEncryption(ctx, db, "user_external_accounts",
//		SecretColumn{Name: "auth_data", Nullable: true},
//		SecretColumn{Name: "account_data", Nullable: true},
//	)
func TableRotateEncryption(ctx context.Context, db *sql.DB, table string, cols ...SecretColumn) error {
	// No need to do expensive queries if encryption is not even configured.
	if !secret.ConfiguredToEncrypt() {
		return nil
	}

	colNames := make([]string, 0, len(cols))
	conds := make([]*sqlf.Query, 0, len(cols))
	for _, col := range cols {
		colNames = append(colNames, col.Name)

		qs := []*sqlf.Query{
			sqlf.Sprintf(col.Name+" IS NOT LIKE %s", secret.PrimaryKeyHash()+secret.Separator+"%"),
		}
		if col.Nullable {
			qs = append(qs, sqlf.Sprintf(col.Name+" IS NOT NULL"))
		}
		conds = append(conds, sqlf.Join(qs, "AND"))
	}

	// TODO: Paginated query
	return dbutil.Transaction(ctx, db, func(tx *sql.Tx) error {
		// NOTE: We want to lock these rows, otherwise their values might have changed
		// during the transaction and we ended up stored encrypted version of stale values.

		// Example:
		//	SELECT id, auth_data, account_data
		//	FROM user_external_accounts
		//	WHERE
		//		(auth_data IS NOT LIKE 'ceda1a$%' AND auth_data IS NOT NULL)
		//	OR	(account_data IS NOT LIKE 'ceda1a$%' AND account_data IS NOT NULL)
		//	FOR UPDATE
		q := sqlf.Sprintf(`
-- source: internal/db/encrypt_table.go:TableRotateEncryption.select
SELECT id, `+strings.Join(colNames, ", ")+`
FROM `+table+`
WHERE %s
FOR UPDATE`,
			sqlf.Join(conds, "OR"),
		)

		var updatedVals [][]interface{}

		rows, err := tx.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
		if err != nil {
			return errors.Wrap(err, "select")
		}
		for rows.Next() {
			var id int64
			dests := make([]interface{}, 1, len(cols)+1)
			dests[0] = &id
			for i := 1; i < len(dests); i++ {
				var s string
				// We only select rows with no NULL values, therefore we can safely assume
				// that we won't encounter any NULL values here.
				dests[i] = secret.StringValue{S: &s}
			}
			if err = rows.Scan(dests...); err != nil {
				return errors.Wrap(err, "scan")
			}

			updatedVals = append(updatedVals, dests)
		}
		if err = rows.Err(); err != nil {
			return errors.Wrap(err, "rows.Err")
		}

		if len(updatedVals) == 0 {
			return nil
		}

		setClauses := make([]string, 0, len(cols))
		for _, col := range cols {
			setClauses = append(setClauses, col.Name+" = update."+col.Name)
		}

		// Example: (%s, %s, %s), the extra %s is for the id column.
		tupleFmt := "(%s" + strings.Repeat(", %s", len(cols)) + ")"
		tuples := make([]*sqlf.Query, 0, len(updatedVals))
		for _, vals := range updatedVals {
			tuples = append(tuples, sqlf.Sprintf(tupleFmt, vals...))
		}

		// Example:
		//	UPDATE user_external_accounts
		//	SET
		//		auth_data = update.auth_data,
		//		account_data = update.account_data
		//	FROM (
		//		VALUES
		//		(1, 'ceda1a$<ciphertext>', 'ceda1a$<ciphertext>'),
		//		(2, 'ceda1a$<ciphertext>', 'ceda1a$<ciphertext>')
		//	) AS update (id, auth_data, account_data)
		//	WHERE user_external_accounts.id = update.id
		q = sqlf.Sprintf(`
-- source: internal/db/encrypt_table.go:TableRotateEncryption.update
UPDATE `+table+`
SET `+strings.Join(setClauses, ", ")+`
FROM (VALUES %s) AS update (id, `+strings.Join(colNames, ", ")+`)
WHERE `+table+`.id = update.id`,
			sqlf.Join(tuples, ", "),
		)

		_, err = tx.ExecContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
		if err != nil {
			return errors.Wrap(err, "update")
		}
		return nil
	})
}
