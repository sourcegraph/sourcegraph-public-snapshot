package intel

import (
	"context"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Location struct {
	Repo     string
	Commit   string
	Path     string
	Position Position
}

type Position struct {
	Line      int
	Character int
}

func positionFromGraphQL[T interface {
	GetLine() int
	GetCharacter() int
}](gql T) Position {
	return Position{
		Line:      gql.GetLine(),
		Character: gql.GetCharacter(),
	}
}

type Range struct {
	Start Position
	End   Position
}

type Reference struct {
	Repo   string
	Commit string
	Path   string
	Range  Range
}

func (s *IntelService) GetReferences(ctx context.Context, location Location) ([]Reference, error) {
	resp, err := getReferences(ctx, s.client, location.Repo, location.Commit, location.Path, location.Position.Line, location.Position.Character)
	if err != nil {
		return nil, errors.Wrap(err, "getting references")
	}

	refs := make([]Reference, 0, len(resp.Repository.Commit.Blob.Lsif.References.Nodes))
	// FIXME: not sure what happens if precise code intel isn't available, in
	// which case the Lsif element should be null. Zero valued? Should probably
	// figure that out.
	for _, ref := range resp.Repository.Commit.Blob.Lsif.References.Nodes {
		refs = append(refs, Reference{
			Repo:   ref.Resource.Repository.Name,
			Commit: ref.Resource.Commit.Oid,
			Path:   ref.Resource.Path,
			Range: Range{
				Start: positionFromGraphQL(&ref.Range.Start),
				End:   positionFromGraphQL(&ref.Range.End),
			},
		})
	}

	return refs, nil
}
