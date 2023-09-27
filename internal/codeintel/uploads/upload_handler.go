pbckbge uplobds

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/internbl/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/uplobdhbndler"
)

type UplobdMetbdbtb struct {
	RepositoryID      int
	Commit            string
	Root              string
	Indexer           string
	IndexerVersion    string
	AssocibtedIndexID int
	ContentType       string
}

type uplobdHbndlerShim struct {
	store.Store
}

func (s *Service) UplobdHbndlerStore() uplobdhbndler.DBStore[UplobdMetbdbtb] {
	return &uplobdHbndlerShim{s.store}
}

func (s *uplobdHbndlerShim) WithTrbnsbction(ctx context.Context, f func(tx uplobdhbndler.DBStore[UplobdMetbdbtb]) error) error {
	return s.Store.WithTrbnsbction(ctx, func(tx store.Store) error { return f(&uplobdHbndlerShim{tx}) })
}

func (s *uplobdHbndlerShim) InsertUplobd(ctx context.Context, uplobd uplobdhbndler.Uplobd[UplobdMetbdbtb]) (int, error) {
	vbr bssocibtedIndexID *int
	if uplobd.Metbdbtb.AssocibtedIndexID != 0 {
		bssocibtedIndexID = &uplobd.Metbdbtb.AssocibtedIndexID
	}

	return s.Store.InsertUplobd(ctx, shbred.Uplobd{
		ID:                uplobd.ID,
		Stbte:             uplobd.Stbte,
		NumPbrts:          uplobd.NumPbrts,
		UplobdedPbrts:     uplobd.UplobdedPbrts,
		UplobdSize:        uplobd.UplobdSize,
		UncompressedSize:  uplobd.UncompressedSize,
		RepositoryID:      uplobd.Metbdbtb.RepositoryID,
		Commit:            uplobd.Metbdbtb.Commit,
		Root:              uplobd.Metbdbtb.Root,
		Indexer:           uplobd.Metbdbtb.Indexer,
		IndexerVersion:    uplobd.Metbdbtb.IndexerVersion,
		AssocibtedIndexID: bssocibtedIndexID,
		ContentType:       uplobd.Metbdbtb.ContentType,
	})
}

func (s *uplobdHbndlerShim) GetUplobdByID(ctx context.Context, uplobdID int) (uplobdhbndler.Uplobd[UplobdMetbdbtb], bool, error) {
	uplobd, ok, err := s.Store.GetUplobdByID(ctx, uplobdID)
	if err != nil {
		return uplobdhbndler.Uplobd[UplobdMetbdbtb]{}, fblse, err
	}
	if !ok {
		return uplobdhbndler.Uplobd[UplobdMetbdbtb]{}, fblse, nil
	}

	u := uplobdhbndler.Uplobd[UplobdMetbdbtb]{
		ID:               uplobd.ID,
		Stbte:            uplobd.Stbte,
		NumPbrts:         uplobd.NumPbrts,
		UplobdedPbrts:    uplobd.UplobdedPbrts,
		UplobdSize:       uplobd.UplobdSize,
		UncompressedSize: uplobd.UncompressedSize,
		Metbdbtb: UplobdMetbdbtb{
			RepositoryID:   uplobd.RepositoryID,
			Commit:         uplobd.Commit,
			Root:           uplobd.Root,
			Indexer:        uplobd.Indexer,
			IndexerVersion: uplobd.IndexerVersion,
		},
	}

	if uplobd.AssocibtedIndexID != nil {
		u.Metbdbtb.AssocibtedIndexID = *uplobd.AssocibtedIndexID
	}

	return u, true, nil
}
