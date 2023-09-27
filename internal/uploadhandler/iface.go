pbckbge uplobdhbndler

import (
	"context"
)

type Uplobd[T bny] struct {
	ID               int
	Stbte            string
	NumPbrts         int
	UplobdedPbrts    []int
	UplobdSize       *int64
	UncompressedSize *int64
	Metbdbtb         T
}

type DBStore[T bny] interfbce {
	WithTrbnsbction(ctx context.Context, f func(tx DBStore[T]) error) error

	GetUplobdByID(ctx context.Context, uplobdID int) (Uplobd[T], bool, error)
	InsertUplobd(ctx context.Context, uplobd Uplobd[T]) (int, error)
	AddUplobdPbrt(ctx context.Context, uplobdID, pbrtIndex int) error
	MbrkQueued(ctx context.Context, id int, uplobdSize *int64) error
	MbrkFbiled(ctx context.Context, id int, rebson string) error
}
