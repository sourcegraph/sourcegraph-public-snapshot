package writer

import (
	"context"
	"encoding/binary"
	"fmt"

	badger "github.com/dgraph-io/badger/v2"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/serializer"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/types"
)

type badgerWriter struct {
	db         *badger.DB
	wb         *badger.WriteBatch
	serializer serializer.Serializer
}

var _ Writer = &badgerWriter{}

func NewBadgerWriter(dirname string, serializer serializer.Serializer) (Writer, error) {
	db, err := badger.Open(badger.DefaultOptions(dirname).WithSyncWrites(false))
	if err != nil {
		return nil, err
	}

	return &badgerWriter{
		db:         db,
		wb:         db.NewWriteBatch(),
		serializer: serializer,
	}, nil
}

func (w *badgerWriter) WriteMeta(ctx context.Context, lsifVersion string, numResultChunks int) error {
	// TODO - make a type for this instead
	if err := w.wb.Set(makeKey("lsifVersion"), []byte(lsifVersion)); err != nil {
		return err
	}
	if err := w.wb.Set(makeKey("internalVersion"), []byte(InternalVersion)); err != nil {
		return err
	}
	if err := w.wb.Set(makeKey("numResultChunks"), []byte(fmt.Sprintf("%d", numResultChunks))); err != nil {
		return err
	}
	return nil
}

func (w *badgerWriter) WriteDocuments(ctx context.Context, documents map[string]types.DocumentData) error {
	for key, value := range documents {
		ser, err := w.serializer.MarshalDocumentData(value)
		if err != nil {
			return errors.Wrap(err, "serializer.MarshalDocumentData")
		}

		if err := w.wb.Set(makeKey("document", key), ser); err != nil {
			return err
		}
	}

	return nil
}

func (w *badgerWriter) WriteResultChunks(ctx context.Context, resultChunks map[int]types.ResultChunkData) error {
	for key, value := range resultChunks {
		ser, err := w.serializer.MarshalResultChunkData(value)
		if err != nil {
			return errors.Wrap(err, "serializer.MarshalResultChunkData")
		}

		if err := w.wb.Set(makeKey("resultChunk", fmt.Sprintf("%d", key)), ser); err != nil {
			return err
		}
	}

	return nil
}

func (w *badgerWriter) WriteDefinitions(ctx context.Context, definitions []types.DefinitionReferenceRow) error {
	return w.writeDefinitionReferences(ctx, "definition", definitions)
}

func (w *badgerWriter) WriteReferences(ctx context.Context, references []types.DefinitionReferenceRow) error {
	return w.writeDefinitionReferences(ctx, "reference", references)
}

func (w *badgerWriter) writeDefinitionReferences(ctx context.Context, prefix string, rows []types.DefinitionReferenceRow) error {
	for i, r := range rows {
		ser := make([]byte, 4*4+len(r.URI))
		binary.LittleEndian.PutUint32(ser[0:], uint32(r.StartLine))
		binary.LittleEndian.PutUint32(ser[4:], uint32(r.StartCharacter))
		binary.LittleEndian.PutUint32(ser[8:], uint32(r.EndLine))
		binary.LittleEndian.PutUint32(ser[12:], uint32(r.EndCharacter))
		copy(ser[16:], r.URI)

		// TODO - make type
		// ser := makeKey(r.URI, fmt.Sprintf("%d", r.StartLine), fmt.Sprintf("%d", r.StartCharacter), fmt.Sprintf("%d", r.EndLine), fmt.Sprintf("%d", r.EndCharacter))

		if err := w.wb.Set(makeKey(prefix, r.Scheme, r.Identifier, fmt.Sprintf("%d", i)), ser); err != nil {
			return err
		}
	}

	return nil
}

func (w *badgerWriter) Flush(ctx context.Context) error {
	return w.wb.Flush()
}

func (w *badgerWriter) Close() error {
	w.wb.Cancel()
	return w.db.Close()
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
