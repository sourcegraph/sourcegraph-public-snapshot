package database

import (
	"context"
	"sort"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/encryption"
)

type RecordEncrypter struct {
	*basestore.Store
}

func NewRecordEncrypter(db DB) *RecordEncrypter {
	return &RecordEncrypter{
		Store: basestore.NewWithHandle(db.Handle()),
	}
}

func (s *RecordEncrypter) Count(ctx context.Context, config EncryptionConfig) (numEncrypted int, numUnencrypted int, _ error) {
	countQuery := sqlf.Sprintf(`
		SELECT
			(SELECT COUNT(*) FROM %s WHERE %s NOT IN ('', %s)) AS   encrypted,
			(SELECT COUNT(*) FROM %s WHERE %s     IN ('', %s)) AS unencrypted
		`,
		quote(config.TableName),
		quote(config.KeyIDFieldName),
		encryption.UnmigratedEncryptionKeyID,
		quote(config.TableName),
		quote(config.KeyIDFieldName),
		encryption.UnmigratedEncryptionKeyID,
	)
	if err := s.QueryRow(ctx, countQuery).Scan(&numEncrypted, &numUnencrypted); err != nil {
		return 0, 0, err
	}

	return numEncrypted, numUnencrypted, nil
}

func (s *RecordEncrypter) EncryptBatch(ctx context.Context, config EncryptionConfig) (count int, err error) {
	key := config.Key()
	if key == nil {
		return 0, nil
	}

	tx, err := s.Transact(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { err = tx.Done(err) }()

	values, err := config.Scan(tx.Query(ctx, sqlf.Sprintf(
		"SELECT %s FROM %s WHERE %s IN ('', %s) ORDER BY %s ASC LIMIT %s FOR UPDATE SKIP LOCKED",
		fields(config),
		quote(config.TableName),
		quote(config.KeyIDFieldName),
		encryption.UnmigratedEncryptionKeyID,
		quote(config.IDFieldName),
		config.Limit,
	)))
	if err != nil {
		return 0, err
	}

	unwrapped := make(map[int][]string, len(values))
	for id, ev := range values {
		unwrapped[id] = ev.Values
	}
	encryptedValues, err := encryptValues(ctx, key, unwrapped)
	if err != nil {
		return 0, err
	}

	for id, ev := range encryptedValues {
		if err := tx.Exec(ctx, sqlf.Sprintf(
			"UPDATE %s SET %s WHERE %s = %s",
			quote(config.TableName),
			updatePairs(config, ev),
			quote(config.IDFieldName),
			id,
		)); err != nil {
			return 0, err
		}
	}

	return len(encryptedValues), nil
}

func (s *RecordEncrypter) DecryptBatch(ctx context.Context, config EncryptionConfig) (count int, err error) {
	tx, err := s.Transact(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { err = tx.Done(err) }()

	values, err := config.Scan(tx.Query(ctx, sqlf.Sprintf(
		"SELECT %s FROM %s WHERE %s NOT IN ('', %s) ORDER BY %s ASC LIMIT %s FOR UPDATE SKIP LOCKED",
		fields(config),
		quote(config.TableName),
		quote(config.KeyIDFieldName),
		encryption.UnmigratedEncryptionKeyID,
		quote(config.IDFieldName),
		config.Limit,
	)))
	if err != nil {
		return 0, err
	}

	decryptedValues, err := decryptValues(ctx, config.Key(), values)
	if err != nil {
		return 0, err
	}

	for id, vs := range decryptedValues {
		if err := tx.Exec(ctx, sqlf.Sprintf(
			"UPDATE %s SET %s WHERE %s = %s",
			quote(config.TableName),
			updatePairs(config, Encrypted{Values: vs}),
			quote(config.IDFieldName),
			id,
		)); err != nil {
			return 0, err
		}
	}

	return len(decryptedValues), nil
}

func fields(c EncryptionConfig) *sqlf.Query {
	names := make([]*sqlf.Query, 0, len(c.EncryptedFieldNames)+2)
	names = append(names, quote(c.IDFieldName), quote(c.KeyIDFieldName))
	names = append(names, quoteSlice(c.EncryptedFieldNames)...)
	return sqlf.Join(names, ", ")
}

func updatePairs(c EncryptionConfig, ev Encrypted) *sqlf.Query {
	m := make(map[string]string, len(ev.Values)+1)
	for i, value := range ev.Values {
		m[c.EncryptedFieldNames[i]] = value
	}
	m[c.KeyIDFieldName] = ev.KeyID

	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	updates := make([]*sqlf.Query, 0, len(m))
	for _, k := range keys {
		if m[k] == "" && k != c.KeyIDFieldName && c.TreatEmptyAsNull {
			updates = append(updates, updatePair(k, nil))
		} else if c.UpdateAsBytes {
			updates = append(updates, updatePair(k, []byte(m[k])))
		} else {
			updates = append(updates, updatePair(k, m[k]))
		}
	}

	return sqlf.Join(updates, ", ")
}

var quote = sqlf.Sprintf

func quoteSlice(vs []string) []*sqlf.Query {
	qs := make([]*sqlf.Query, 0, len(vs))
	for _, v := range vs {
		qs = append(qs, quote(v))
	}

	return qs
}

func updatePair(column string, value any) *sqlf.Query {
	return sqlf.Sprintf("%s = %s", quote(column), value)
}
