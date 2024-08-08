package codyaccess

import "github.com/jackc/pgx/v5/pgxpool"

// Store is the storage layer for Cody Access services.
type Store struct {
	db *pgxpool.Pool
}

func NewStore(db *pgxpool.Pool) *Store {
	return &Store{
		db: db,
	}
}

func (s *Store) CodyGateway() *CodyGatewayStore {
	return NewCodyGatewayStore(s.db)
}
