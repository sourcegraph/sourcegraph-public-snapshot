package reader

import (
	"context"
	"encoding/binary"
	"fmt"
	"strconv"

	badger "github.com/dgraph-io/badger/v2"
	pkgerrors "github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/serializer"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/types"
)

type badgerReader struct {
	db         *badger.DB
	serializer serializer.Serializer
}

var _ Reader = &badgerReader{}

func NewBadgerReader(dirname string, serializer serializer.Serializer) (_ Reader, err error) {
	db, err := badger.Open(badger.DefaultOptions(dirname))
	if err != nil {
		return nil, err
	}

	return &badgerReader{
		db:         db,
		serializer: serializer,
	}, nil
}

func (r *badgerReader) ReadMeta(ctx context.Context) (lsifVersion string, sourcegraphVersion string, numResultChunks int, _ error) {
	err := r.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lsifVersion"))
		if err != nil {
			return err
		}
		if err := item.Value(func(val []byte) error {
			lsifVersion = string(val)
			return nil
		}); err != nil {
			return err
		}

		item, err = txn.Get([]byte("internalVersion"))
		if err != nil {
			return err
		}
		if err := item.Value(func(val []byte) error {
			sourcegraphVersion = string(val)
			return nil
		}); err != nil {
			return err
		}

		item, err = txn.Get([]byte("numResultChunks"))
		if err != nil {
			return err
		}
		if err := item.Value(func(val []byte) error {
			numResultChunks, _ = strconv.Atoi(string(val))
			return nil
		}); err != nil {
			return err
		}

		return nil
	})

	return lsifVersion, sourcegraphVersion, numResultChunks, err
}

func (r *badgerReader) ReadDocument(ctx context.Context, path string) (types.DocumentData, bool, error) {
	var documentData types.DocumentData
	err := r.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(makeKey("document", path))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			if documentData, err = r.serializer.UnmarshalDocumentData(val); err != nil {
				return pkgerrors.Wrap(err, "serializer.UnmarshalDocumentData")
			}

			return nil
		})
	})
	if err != nil {
		if err == badger.ErrEmptyKey {
			return types.DocumentData{}, false, nil
		}
		return types.DocumentData{}, false, err
	}

	return documentData, true, nil

}

func (r *badgerReader) ReadResultChunk(ctx context.Context, id int) (types.ResultChunkData, bool, error) {
	var resultChunkData types.ResultChunkData
	err := r.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(makeKey("resultChunk", fmt.Sprintf("%d", id)))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			if resultChunkData, err = r.serializer.UnmarshalResultChunkData(val); err != nil {
				return pkgerrors.Wrap(err, "serializer.UnmarshalResultChunkData")
			}

			return nil
		})
	})
	if err != nil {
		if err == badger.ErrEmptyKey {
			return types.ResultChunkData{}, false, nil
		}
		return types.ResultChunkData{}, false, err
	}

	return resultChunkData, true, nil
}

func (r *badgerReader) ReadDefinitions(ctx context.Context, scheme, identifier string, skip, take int) ([]types.DefinitionReferenceRow, int, error) {
	return r.readDefinitionReferences(ctx, "definition", scheme, identifier, skip, take)
}

func (r *badgerReader) ReadReferences(ctx context.Context, scheme, identifier string, skip, take int) ([]types.DefinitionReferenceRow, int, error) {
	return r.readDefinitionReferences(ctx, "reference", scheme, identifier, skip, take)
}

func (r *badgerReader) readDefinitionReferences(ctx context.Context, prefix, scheme, identifier string, skip, take int) ([]types.DefinitionReferenceRow, int, error) {
	var total int
	var rows []types.DefinitionReferenceRow

	err := r.db.View(func(txn *badger.Txn) error {
		p := makeKey(prefix, scheme, identifier, "")

		it := txn.NewIterator(badger.IteratorOptions{
			PrefetchValues: true,
			PrefetchSize:   100,
		})
		defer it.Close()

		inf := skip == 0 && take == 0

		for it.Seek(p); it.ValidForPrefix(p); it.Next() {
			total++

			skip--
			if !inf && skip >= 0 {
				continue
			}

			if !inf && len(rows) >= take {
				continue
			}

			if err := it.Item().Value(func(val []byte) error {
				if len(val) < 17 {
					return fmt.Errorf("Short key")
				}

				startLine := int(binary.LittleEndian.Uint32(val[0:]))
				startCharacter := int(binary.LittleEndian.Uint32(val[4:]))
				endLine := int(binary.LittleEndian.Uint32(val[8:]))
				endCharacter := int(binary.LittleEndian.Uint32(val[12:]))

				rows = append(rows, types.DefinitionReferenceRow{
					Scheme:         scheme,
					Identifier:     identifier,
					URI:            string(val[16:]),
					StartLine:      startLine,
					StartCharacter: startCharacter,
					EndLine:        endLine,
					EndCharacter:   endCharacter,
				})
				return nil
			}); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, 0, err
	}

	return rows, total, nil
}

func (r *badgerReader) Close() error {
	return r.db.Close()
}

func makeKey(values ...string) []byte {
	s := len(values) - 1
	for _, v := range values {
		s += len(v)
	}

	idx := 0
	buf := make([]byte, s)
	for _, v := range values {
		copy(buf, v)
		buf[len(v)] = 0
		idx += len(v) + 1
	}

	return buf
}
