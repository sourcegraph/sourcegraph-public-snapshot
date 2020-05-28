package writer

import (
	"context"
	"encoding/binary"
	"fmt"

	badger "github.com/dgraph-io/badger/v2"
	badgeroptions "github.com/dgraph-io/badger/v2/options"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/serializer"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/types"
)

type badgerWriter struct {
	db         *badger.DB
	serializer serializer.Serializer
}

var _ Writer = &badgerWriter{}

func NewBadgerWriter(dirname string, serializer serializer.Serializer) (Writer, error) {
	options := badger.DefaultOptions(dirname).
		WithCompression(badgeroptions.None).
		WithLogger(nil).
		WithSyncWrites(false).
		WithTableLoadingMode(badgeroptions.MemoryMap).
		WithValueThreshold(1 << 10) // 1kb

	db, err := badger.Open(options)
	if err != nil {
		return nil, err
	}

	return &badgerWriter{
		db:         db,
		serializer: serializer,
	}, nil
}

type bytePair struct {
	Key   []byte
	Value []byte
}

func (w *badgerWriter) WriteMeta(ctx context.Context, lsifVersion string, numResultChunks int) error {
	return w.write([]bytePair{
		bytePair{
			Key:   makeKey("0:metaData"),
			Value: marshalMetaData(lsifVersion, InternalVersion, numResultChunks),
		},
	})
}

func (w *badgerWriter) WriteDocuments(ctx context.Context, documents map[string]types.DocumentData) error {
	pairs := make([]bytePair, 0, len(documents))
	for path, document := range documents {
		ser, err := w.serializer.MarshalDocumentData(document)
		if err != nil {
			return errors.Wrap(err, "serializer.MarshalDocumentData")
		}

		pairs = append(pairs, bytePair{
			Key:   makeKey("1:document", path),
			Value: ser,
		})
	}

	return w.write(pairs)
}

func (w *badgerWriter) WriteResultChunks(ctx context.Context, resultChunks map[int]types.ResultChunkData) error {
	pairs := make([]bytePair, 0, len(resultChunks))
	for key, value := range resultChunks {
		ser, err := w.serializer.MarshalResultChunkData(value)
		if err != nil {
			return errors.Wrap(err, "serializer.MarshalResultChunkData")
		}

		pairs = append(pairs, bytePair{
			Key:   makeKey("2:resultChunk", fmt.Sprintf("%d", key)),
			Value: ser,
		})
	}

	return w.write(pairs)
}

func (w *badgerWriter) WriteDefinitions(ctx context.Context, definitions []types.DefinitionReferenceRow) error {
	return w.writeDefinitionReferences(ctx, "3:definition", definitions)
}

func (w *badgerWriter) WriteReferences(ctx context.Context, references []types.DefinitionReferenceRow) error {
	return w.writeDefinitionReferences(ctx, "4:reference", references)
}

func (w *badgerWriter) writeDefinitionReferences(ctx context.Context, prefix string, rows []types.DefinitionReferenceRow) error {
	pairs := make([]bytePair, 0, len(rows))
	for i, r := range rows {
		pairs = append(pairs, bytePair{
			Key:   makeKey(prefix, r.Scheme, r.Identifier, fmt.Sprintf("%d", i)),
			Value: marshalDefinitionReferenceRow(r),
		})
	}

	return w.write(pairs)
}

func (w *badgerWriter) Flush(ctx context.Context) error {
	return nil
}

func (w *badgerWriter) Close() error {
	return w.db.Close()
}

func (w *badgerWriter) write(pairs []bytePair) error {
	txn := w.db.NewTransaction(true)
	defer txn.Discard()

	for _, e := range pairs {
		err := txn.Set(e.Key, e.Value)
		if err == badger.ErrTxnTooBig {
			if commitErr := txn.Commit(); commitErr != nil {
				return commitErr
			}

			txn = w.db.NewTransaction(true)
			err = txn.Set(e.Key, e.Value)
		}
		if err != nil {
			return err
		}
	}

	return txn.Commit()
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
		idx += len(v) + 1
	}

	return buf
}

func marshalMetaData(lsifVersion, internalVersion string, numResultChunks int) []byte {
	buf := make([]byte, 4*3+len(lsifVersion)+len(internalVersion))
	binary.LittleEndian.PutUint32(buf[0:], uint32(len(lsifVersion)))
	binary.LittleEndian.PutUint32(buf[4:], uint32(len(internalVersion)))
	binary.LittleEndian.PutUint32(buf[8:], uint32(numResultChunks))
	copy(buf[12:], lsifVersion)
	copy(buf[12+len(lsifVersion):], internalVersion)
	return buf
}

func marshalDefinitionReferenceRow(row types.DefinitionReferenceRow) []byte {
	buf := make([]byte, 4*4+len(row.URI))
	binary.LittleEndian.PutUint32(buf[0:], uint32(row.StartLine))
	binary.LittleEndian.PutUint32(buf[4:], uint32(row.StartCharacter))
	binary.LittleEndian.PutUint32(buf[8:], uint32(row.EndLine))
	binary.LittleEndian.PutUint32(buf[12:], uint32(row.EndCharacter))
	copy(buf[16:], row.URI)
	return buf
}
