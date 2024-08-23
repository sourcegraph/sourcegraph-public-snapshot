package responses

import "github.com/uber/gonduit/entities"

// RepositoryQueryResponse is the result of repository.query operations.
type RepositoryQueryResponse []*entities.Repository
