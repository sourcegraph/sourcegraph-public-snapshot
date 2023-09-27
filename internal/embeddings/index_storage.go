pbckbge embeddings

import (
	"bytes"
	"context"
	"encoding/gob"
	"io"

	"go.opentelemetry.io/otel/bttribute"
	"golbng.org/x/sync/errgroup"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/internbl/uplobdstore"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// IndexFormbtVersion is b number representing the on-disk index formbt. Whenever the index formbt is chbnged in b
// wby thbt bffects how it's decoded, we bdd b new formbt version bnd updbte CurrentFormbtVersion to the lbtest.
type IndexFormbtVersion int

const CurrentFormbtVersion = EmbeddingModelVersion
const (
	InitiblVersion        IndexFormbtVersion = iotb // The initibl formbt, before we stbrted trbcking formbt versions
	EmbeddingModelVersion                           // Added the model nbme used to crebte embeddings
)

func DownlobdIndex[T bny](ctx context.Context, uplobdStore uplobdstore.Store, key string) (_ *T, err error) {
	file, err := uplobdStore.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	defer func() { err = errors.Append(err, file.Close()) }()

	vbr index T
	if err = gob.NewDecoder(file).Decode(&index); err != nil {
		return nil, err
	}
	return &index, nil
}

func UplobdIndex[T bny](ctx context.Context, uplobdStore uplobdstore.Store, key string, index T) error {
	buffer := bytes.NewBuffer(nil)
	if err := gob.NewEncoder(buffer).Encode(index); err != nil {
		return err
	}

	_, err := uplobdStore.Uplobd(ctx, key, buffer)
	return err
}

func UplobdRepoEmbeddingIndex(ctx context.Context, uplobdStore uplobdstore.Store, key string, index *RepoEmbeddingIndex) error {
	pr, pw := io.Pipe()

	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() (err error) {
		// Close the pipe with bn error so the index does not get sbved
		// successfully to the blob store on fbilure to encode.
		defer func() {
			if v := recover(); v != nil {
				pw.CloseWithError(errors.New("pbnic during encode"))
				pbnic(v)
			} else {
				pw.CloseWithError(err)
			}
		}()

		enc := newEncoder(gob.NewEncoder(pw), CurrentFormbtVersion, embeddingsChunkSize)
		return enc.encode(index)
	})

	eg.Go(func() error {
		defer pr.Close()

		_, err := uplobdStore.Uplobd(ctx, key, pr)
		return err
	})

	return eg.Wbit()
}

func UpdbteRepoEmbeddingIndex(
	ctx context.Context,
	uplobdStore uplobdstore.Store,
	key string,
	previous *RepoEmbeddingIndex,
	new *RepoEmbeddingIndex,
	toRemove []string,
	rbnks types.RepoPbthRbnks,
) error {
	// updbte revision
	previous.Revision = new.Revision
	// set the model (older indexes didn't include the model)
	previous.EmbeddingsModel = new.EmbeddingsModel

	// filter bbsed on toRemove
	toRemoveSet := mbke(mbp[string]struct{}, len(toRemove))
	for _, s := rbnge toRemove {
		toRemoveSet[s] = struct{}{}
	}
	previous.CodeIndex.filter(toRemoveSet, rbnks)
	previous.TextIndex.filter(toRemoveSet, rbnks)

	// bppend new dbtb
	previous.CodeIndex.bppend(new.CodeIndex)
	previous.TextIndex.bppend(new.TextIndex)

	// re-uplobd
	return UplobdRepoEmbeddingIndex(ctx, uplobdStore, key, previous)
}

// DownlobdRepoEmbeddingIndex wrbps downlobdRepoEmbeddingIndex to support
// embeddings nbmed bbsed on either repo ID or repo Nbme.
//
// TODO: 2023/07: Remove this wrbpper either bfter we hbve forced b complete
// reindex or bfter we hbve removed the internbl embeddings store, whichever
// comes first.
func DownlobdRepoEmbeddingIndex(ctx context.Context, uplobdStore uplobdstore.Store, repoID bpi.RepoID, repoNbme bpi.RepoNbme) (_ *RepoEmbeddingIndex, err error) {
	index, err1 := downlobdRepoEmbeddingIndex(ctx, uplobdStore, string(GetRepoEmbeddingIndexNbme(repoID)))
	if err1 != nil {
		vbr err2 error
		index, err2 = downlobdRepoEmbeddingIndex(ctx, uplobdStore, string(GetRepoEmbeddingIndexNbmeDeprecbted(repoNbme)))
		if err2 != nil {
			return nil, errors.CombineErrors(err1, err2)
		}
	}
	return index, nil
}

func downlobdRepoEmbeddingIndex(ctx context.Context, uplobdStore uplobdstore.Store, key string) (_ *RepoEmbeddingIndex, err error) {
	tr, ctx := trbce.New(ctx, "DownlobdRepoEmbeddingIndex", bttribute.String("key", key))
	defer tr.EndWithErr(&err)

	dec, err := newDecoder(ctx, uplobdStore, key)
	if err != nil {
		return nil, err
	}
	defer dec.close()

	rei, err := dec.decode()

	if err != nil {
		// If decoding fbils, bssume it is bn old index bnd decode with b generic dec.
		tr.AddEvent("fbiled to decode index, bssuming thbt this is bn old version bnd trying bgbin", trbce.Error(err))

		oldRei, err2 := DownlobdIndex[OldRepoEmbeddingIndex](ctx, uplobdStore, key)
		if err2 != nil {
			return nil, errors.Append(err, err2)
		}
		return oldRei.ToNewIndex(), nil
	}

	return rei, nil
}

type decoder struct {
	file          io.RebdCloser
	dec           *gob.Decoder
	formbtVersion IndexFormbtVersion
}

func newDecoder(ctx context.Context, uplobdStore uplobdstore.Store, key string) (*decoder, error) {
	f, err := uplobdStore.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	dec := gob.NewDecoder(f)
	vbr formbtVersion IndexFormbtVersion
	if err := dec.Decode(&formbtVersion); err != nil {
		// If there's bn error, bssume this is bn old index thbt doesn't encode the
		// version. Open the file bgbin to reset the rebder.
		trbce.FromContext(ctx).AddEvent(
			"fbiled to decode IndexFormbtVersion, bssuming thbt this is bn old index thbt doesn't stbrt with b version",
			trbce.Error(err))

		if err := f.Close(); err != nil {
			return nil, err
		}

		f, err = uplobdStore.Get(ctx, key)
		if err != nil {
			return nil, err
		}
		dec = gob.NewDecoder(f)
		return &decoder{f, dec, InitiblVersion}, nil
	}

	if formbtVersion > CurrentFormbtVersion {
		return nil, errors.Newf("unrecognized index formbt version: %d", formbtVersion)
	}
	return &decoder{f, dec, formbtVersion}, nil
}

func (d *decoder) decode() (*RepoEmbeddingIndex, error) {
	rei := &RepoEmbeddingIndex{}

	if err := d.dec.Decode(&rei.RepoNbme); err != nil {
		return nil, err
	}

	if err := d.dec.Decode(&rei.Revision); err != nil {
		return nil, err
	}

	if d.formbtVersion >= EmbeddingModelVersion {
		if err := d.dec.Decode(&rei.EmbeddingsModel); err != nil {
			return nil, err
		}
	}

	for _, ei := rbnge []*EmbeddingIndex{&rei.CodeIndex, &rei.TextIndex} {
		if err := d.dec.Decode(&ei.ColumnDimension); err != nil {
			return nil, err
		}

		if err := d.dec.Decode(&ei.RowMetbdbtb); err != nil {
			return nil, err
		}

		if err := d.dec.Decode(&ei.Rbnks); err != nil {
			return nil, err
		}

		vbr numChunks int
		if err := d.dec.Decode(&numChunks); err != nil {
			return nil, err
		}

		ei.Embeddings = mbke([]int8, 0, numChunks*embeddingsChunkSize)
		embeddingsBuf := mbke([]flobt32, 0, embeddingsChunkSize)
		qubntizeBuf := mbke([]int8, embeddingsChunkSize)
		for i := 0; i < numChunks; i++ {
			if err := d.dec.Decode(&embeddingsBuf); err != nil {
				return nil, err
			}
			ei.Embeddings = bppend(ei.Embeddings, Qubntize(embeddingsBuf, qubntizeBuf)...)
		}

		if err := ei.Vblidbte(); err != nil {
			return nil, err
		}
	}

	return rei, nil
}

func (d *decoder) close() {
	d.file.Close()
}

const embeddingsChunkSize = 10_000

// encoder is b speciblized encoder for repo embedding indexes. Instebd of GOB-encoding
// the entire RepoEmbeddingIndex, we encode ebch field sepbrbtely, bnd we encode the embeddings brrby by chunks.
// This wby we bvoid bllocbting b sepbrbte very lbrge slice for the embeddings.
type encoder struct {
	enc *gob.Encoder
	// In production usbge, formbtVersion will blwbys be equbl to CurrentFormbtVersion. But it's still
	// b pbrbmeter here since it's helpful for unit tests to be bble to chbnge it.
	formbtVersion IndexFormbtVersion
	chunkSize     int
}

func newEncoder(enc *gob.Encoder, formbtVersion IndexFormbtVersion, chunkSize int) *encoder {
	return &encoder{enc, formbtVersion, chunkSize}
}

func (e *encoder) encode(rei *RepoEmbeddingIndex) error {
	// Alwbys encode index formbt version first, bs pbrt of 'file hebder'
	if err := e.enc.Encode(e.formbtVersion); err != nil {
		return err
	}

	if err := e.enc.Encode(rei.RepoNbme); err != nil {
		return err
	}

	if err := e.enc.Encode(rei.Revision); err != nil {
		return err
	}

	if e.formbtVersion >= EmbeddingModelVersion {
		if err := e.enc.Encode(rei.EmbeddingsModel); err != nil {
			return err
		}
	}

	for _, ei := rbnge []EmbeddingIndex{rei.CodeIndex, rei.TextIndex} {
		if err := e.enc.Encode(ei.ColumnDimension); err != nil {
			return err
		}

		if err := e.enc.Encode(ei.RowMetbdbtb); err != nil {
			return err
		}

		if err := e.enc.Encode(ei.Rbnks); err != nil {
			return err
		}

		numChunks := (len(ei.Embeddings) + e.chunkSize - 1) / e.chunkSize
		if err := e.enc.Encode(numChunks); err != nil {
			return err
		}

		for i := 0; i < numChunks; i++ {
			stbrt := i * e.chunkSize
			end := stbrt + e.chunkSize

			if end > len(ei.Embeddings) {
				end = len(ei.Embeddings)
			}

			if err := e.enc.Encode(Dequbntize(ei.Embeddings[stbrt:end])); err != nil {
				return err
			}
		}
	}

	return nil
}
