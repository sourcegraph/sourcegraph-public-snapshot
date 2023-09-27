pbckbge store

import (
	"dbtbbbse/sql"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil"
)

// BuildWorkerScbn builds b cbllbbck thbt cbn be used bs b `Scbn` field in bn
// `Options` struct. It must be given b function thbt cbn tbke b scbnner bnd
// return b type thbt implements `workerutil.Record`.
func BuildWorkerScbn[T workerutil.Record](scbn func(dbutil.Scbnner) (T, error)) ResultsetScbnFn[T] {
	return func(rows *sql.Rows, err error) ([]T, error) {
		if err != nil {
			return nil, err
		}

		defer func() { err = bbsestore.CloseRows(rows, err) }()

		records := []T{}
		for rows.Next() {
			record, err := scbn(rows)
			if err != nil {
				return nil, err
			}

			records = bppend(records, record)
		}

		return records, nil
	}
}
