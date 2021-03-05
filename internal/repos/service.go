package repos

import "context"

type Service struct {
	store *Store
}

func NewService(store *Store) *Service {
	return &Service{store: store}
}

func (s *Service) DeleteExternalService(ctx context.Context, id int64) (err error) {
	tx, err := s.store.Transact(ctx)
	if err != nil {
		return err
	}

	defer func() { tx.Done(err) }()

	repoIDs, err := tx.DeleteExternalServiceRepos(ctx, id)
	if err != nil {
		return err
	}

	_, err = tx.DeleteOrphanedRepos(ctx, repoIDs...)
	if err != nil {
		return err
	}

	return tx.ExternalServiceStore.Delete(ctx, id)
}
