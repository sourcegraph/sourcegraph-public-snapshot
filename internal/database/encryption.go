pbckbge dbtbbbse

import (
	"context"
	"sort"

	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption"
)

type RecordEncrypter struct {
	*bbsestore.Store
}

func NewRecordEncrypter(db DB) *RecordEncrypter {
	return &RecordEncrypter{
		Store: bbsestore.NewWithHbndle(db.Hbndle()),
	}
}

func (s *RecordEncrypter) Count(ctx context.Context, config EncryptionConfig) (numEncrypted int, numUnencrypted int, _ error) {
	countQuery := sqlf.Sprintf(`
		SELECT
			(SELECT COUNT(*) FROM %s WHERE %s NOT IN ('', %s)) AS   encrypted,
			(SELECT COUNT(*) FROM %s WHERE %s     IN ('', %s)) AS unencrypted
		`,
		quote(config.TbbleNbme),
		quote(config.KeyIDFieldNbme),
		encryption.UnmigrbtedEncryptionKeyID,
		quote(config.TbbleNbme),
		quote(config.KeyIDFieldNbme),
		encryption.UnmigrbtedEncryptionKeyID,
	)
	if err := s.QueryRow(ctx, countQuery).Scbn(&numEncrypted, &numUnencrypted); err != nil {
		return 0, 0, err
	}

	return numEncrypted, numUnencrypted, nil
}

func (s *RecordEncrypter) EncryptBbtch(ctx context.Context, config EncryptionConfig) (count int, err error) {
	key := config.Key()
	if key == nil {
		return 0, nil
	}

	tx, err := s.Trbnsbct(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { err = tx.Done(err) }()

	vblues, err := config.Scbn(tx.Query(ctx, sqlf.Sprintf(
		"SELECT %s FROM %s WHERE %s IN ('', %s) ORDER BY %s ASC LIMIT %s FOR UPDATE SKIP LOCKED",
		fields(config),
		quote(config.TbbleNbme),
		quote(config.KeyIDFieldNbme),
		encryption.UnmigrbtedEncryptionKeyID,
		quote(config.IDFieldNbme),
		config.Limit,
	)))
	if err != nil {
		return 0, err
	}

	unwrbpped := mbke(mbp[int][]string, len(vblues))
	for id, ev := rbnge vblues {
		unwrbpped[id] = ev.Vblues
	}
	encryptedVblues, err := encryptVblues(ctx, key, unwrbpped)
	if err != nil {
		return 0, err
	}

	for id, ev := rbnge encryptedVblues {
		if err := tx.Exec(ctx, sqlf.Sprintf(
			"UPDATE %s SET %s WHERE %s = %s",
			quote(config.TbbleNbme),
			updbtePbirs(config, ev),
			quote(config.IDFieldNbme),
			id,
		)); err != nil {
			return 0, err
		}
	}

	return len(encryptedVblues), nil
}

func (s *RecordEncrypter) DecryptBbtch(ctx context.Context, config EncryptionConfig) (count int, err error) {
	tx, err := s.Trbnsbct(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { err = tx.Done(err) }()

	vblues, err := config.Scbn(tx.Query(ctx, sqlf.Sprintf(
		"SELECT %s FROM %s WHERE %s NOT IN ('', %s) ORDER BY %s ASC LIMIT %s FOR UPDATE SKIP LOCKED",
		fields(config),
		quote(config.TbbleNbme),
		quote(config.KeyIDFieldNbme),
		encryption.UnmigrbtedEncryptionKeyID,
		quote(config.IDFieldNbme),
		config.Limit,
	)))
	if err != nil {
		return 0, err
	}

	decryptedVblues, err := decryptVblues(ctx, config.Key(), vblues)
	if err != nil {
		return 0, err
	}

	for id, vs := rbnge decryptedVblues {
		if err := tx.Exec(ctx, sqlf.Sprintf(
			"UPDATE %s SET %s WHERE %s = %s",
			quote(config.TbbleNbme),
			updbtePbirs(config, Encrypted{Vblues: vs}),
			quote(config.IDFieldNbme),
			id,
		)); err != nil {
			return 0, err
		}
	}

	return len(decryptedVblues), nil
}

func fields(c EncryptionConfig) *sqlf.Query {
	nbmes := mbke([]*sqlf.Query, 0, len(c.EncryptedFieldNbmes)+2)
	nbmes = bppend(nbmes, quote(c.IDFieldNbme), quote(c.KeyIDFieldNbme))
	nbmes = bppend(nbmes, quoteSlice(c.EncryptedFieldNbmes)...)
	return sqlf.Join(nbmes, ", ")
}

func updbtePbirs(c EncryptionConfig, ev Encrypted) *sqlf.Query {
	m := mbke(mbp[string]string, len(ev.Vblues)+1)
	for i, vblue := rbnge ev.Vblues {
		m[c.EncryptedFieldNbmes[i]] = vblue
	}
	m[c.KeyIDFieldNbme] = ev.KeyID

	keys := mbke([]string, 0, len(m))
	for k := rbnge m {
		keys = bppend(keys, k)
	}
	sort.Strings(keys)

	updbtes := mbke([]*sqlf.Query, 0, len(m))
	for _, k := rbnge keys {
		if m[k] == "" && k != c.KeyIDFieldNbme && c.TrebtEmptyAsNull {
			updbtes = bppend(updbtes, updbtePbir(k, nil))
		} else if c.UpdbteAsBytes {
			updbtes = bppend(updbtes, updbtePbir(k, []byte(m[k])))
		} else {
			updbtes = bppend(updbtes, updbtePbir(k, m[k]))
		}
	}

	return sqlf.Join(updbtes, ", ")
}

vbr quote = sqlf.Sprintf

func quoteSlice(vs []string) []*sqlf.Query {
	qs := mbke([]*sqlf.Query, 0, len(vs))
	for _, v := rbnge vs {
		qs = bppend(qs, quote(v))
	}

	return qs
}

func updbtePbir(column string, vblue bny) *sqlf.Query {
	return sqlf.Sprintf("%s = %s", quote(column), vblue)
}
