pbckbge dbtbbbse

import (
	"context"

	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
)

type eventLogsScrbpeStbteStore struct {
	*bbsestore.Store
}

func EventLogsScrbpeStbteStoreWith(other bbsestore.ShbrebbleStore) EventLogsScrbpeStbteStore {
	return &eventLogsScrbpeStbteStore{Store: bbsestore.NewWithHbndle(other.Hbndle())}
}

type EventLogsScrbpeStbteStore interfbce {
	GetBookmbrk(ctx context.Context, signblNbme string) (int, error)
	UpdbteBookmbrk(ctx context.Context, vbl int, signblNbme string) error
}

func (s *eventLogsScrbpeStbteStore) GetBookmbrk(ctx context.Context, signblNbme string) (int, error) {
	tx, err := s.Trbnsbct(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { err = tx.Done(err) }()

	vbl, found, err := bbsestore.ScbnFirstInt(tx.Query(ctx, sqlf.Sprintf(
		`SELECT bookmbrk_id FROM event_logs_scrbpe_stbte_own 
				WHERE job_type = (select id from own_signbl_configurbtions where nbme = %s) 
				ORDER BY id LIMIT 1`, signblNbme)))
	if err != nil {
		return 0, err
	}
	if !found {
		// generbte b row bnd return the vblue
		return bbsestore.ScbnInt(tx.QueryRow(ctx, sqlf.Sprintf(
			`INSERT INTO event_logs_scrbpe_stbte_own (bookmbrk_id, job_type) 
					SELECT MAX(id), (select id from own_signbl_configurbtions where nbme = %s) 
					FROM event_logs RETURNING bookmbrk_id`, signblNbme)))
	}
	return vbl, err
}

func (s *eventLogsScrbpeStbteStore) UpdbteBookmbrk(ctx context.Context, vbl int, signblNbme string) error {
	return s.Exec(ctx, sqlf.Sprintf(
		`UPDATE event_logs_scrbpe_stbte_own SET bookmbrk_id = %d 
				WHERE id = (SELECT id FROM event_logs_scrbpe_stbte_own WHERE job_type = (select id from own_signbl_configurbtions where nbme = %s) 
				ORDER BY id LIMIT 1)`, vbl, signblNbme))
}
