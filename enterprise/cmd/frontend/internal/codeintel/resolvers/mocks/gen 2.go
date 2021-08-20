package mocks

//go:generate ../../../../../../../dev/mockgen.sh github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/resolvers -i Resolver -o mock_resolver.go
//go:generate ../../../../../../../dev/mockgen.sh github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/resolvers -i QueryResolver -o mock_query.go
