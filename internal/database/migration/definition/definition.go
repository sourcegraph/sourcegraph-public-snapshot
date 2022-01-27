package definition

import (
	"sort"

	"github.com/cockroachdb/errors"
	"github.com/keegancsmith/sqlf"
)

type Definition struct {
	ID           int
	UpFilename   string
	UpQuery      *sqlf.Query
	DownFilename string
	DownQuery    *sqlf.Query
	Parents      []int
}

type IndexMetadata struct {
	TableName string
	IndexName string
}

type Definitions struct {
	definitions    []Definition
	definitionsMap map[int]Definition
}

func newDefinitions(migrationDefinitions []Definition) *Definitions {
	definitionsMap := make(map[int]Definition, len(migrationDefinitions))
	for _, migrationDefinition := range migrationDefinitions {
		definitionsMap[migrationDefinition.ID] = migrationDefinition
	}

	return &Definitions{
		definitions:    migrationDefinitions,
		definitionsMap: definitionsMap,
	}
}

// All returns the set of all definitions ordered such that each migration occurs
// only after all of its parents. The returned slice is not a copy, so it is not
// meant to be mutated.
func (ds *Definitions) All() []Definition {
	return ds.definitions
}

func (ds *Definitions) Count() int {
	return len(ds.definitions)
}

func (ds *Definitions) First() int {
	return ds.definitions[0].ID
}

func (ds *Definitions) GetByID(id int) (Definition, bool) {
	definition, ok := ds.definitionsMap[id]
	return definition, ok
}

// Root returns the definition with no parents.
func (ds *Definitions) Root() Definition {
	return ds.definitions[0]
}

// Leaves returns the definitions with no children.
func (ds *Definitions) Leaves() []Definition {
	childrenMap := children(ds.definitions)

	leaves := make([]Definition, 0, 4)
	for _, definition := range ds.definitions {
		if len(childrenMap[definition.ID]) == 0 {
			leaves = append(leaves, definition)
		}
	}

	return leaves
}

// Filter returns a new definitions object that contains the intersection of the
// receiver's definitions and the given identifiers. This operation is designed to
// cut complete branches of migrations from the tree (for use in squash operations).
// Therefore, it is an error for any of the remaining migrations to reference a
// parent that was not included in the target set of migrations.
func (ds *Definitions) Filter(ids []int) (*Definitions, error) {
	idMap := map[int]struct{}{}
	for _, id := range ids {
		idMap[id] = struct{}{}
	}

	filtered := make([]Definition, 0, len(ds.definitions)-len(ids))
	for _, definition := range ds.definitions {
		if _, ok := idMap[definition.ID]; ok {
			filtered = append(filtered, definition)
		}
	}

	for _, definition := range filtered {
		for _, parent := range definition.Parents {
			if _, ok := idMap[parent]; !ok {
				return nil, fmt.Errorf("illegal filter: migration %d (included) references parent migration %d (excluded)", definition.ID, parent)
			}
		}
	}

	return newDefinitions(filtered), nil
}

func (ds *Definitions) UpTo(id, target int) ([]Definition, error) {
	if target == 0 {
		return ds.UpFrom(id, 0)
	}

	if _, ok := ds.GetByID(target); !ok {
		return nil, errors.Newf("unknown target %d", target)
	}
	if target < id {
		return nil, errors.Newf("migration %d is behind version %d", target, id)
	}
	if target == id {
		// n == 0 has special meaning; handle case immediately
		return nil, nil
	}

	return ds.UpFrom(id, target-id)
}

func (ds *Definitions) UpFrom(id, n int) ([]Definition, error) {
	slice := make([]Definition, 0, len(ds.definitions))
	for _, definition := range ds.definitions {
		if definition.ID <= id {
			continue
		}

		slice = append(slice, definition)
	}

	if n > 0 && len(slice) > n {
		slice = slice[:n]
	}

	if id != 0 && len(slice) != 0 && slice[0].ID != id+1 {
		return nil, errors.Newf("missing migrations [%d, %d]", id+1, slice[0].ID-1)
	}

	return slice, nil
}

func (ds *Definitions) DownTo(id, target int) ([]Definition, error) {
	if target == 0 {
		return nil, errors.Newf("illegal downgrade target %d", target)
	}

	if _, ok := ds.GetByID(target); !ok {
		return nil, errors.Newf("unknown target %d", target)
	}
	if id < target {
		return nil, errors.Newf("migration %d is ahead of version %d", target, id)
	}

	return ds.DownFrom(id, id-target)
}

func (ds *Definitions) DownFrom(id, n int) ([]Definition, error) {
	slice := make([]Definition, 0, len(ds.definitions))
	for _, definition := range ds.definitions {
		if definition.ID <= id {
			slice = append(slice, definition)
		}
	}

	sort.Slice(slice, func(i, j int) bool {
		return slice[j].ID < slice[i].ID
	})

	if len(slice) > n {
		slice = slice[:n]
	}

	if id != 0 && len(slice) != 0 && slice[0].ID != id {
		return nil, errors.Newf("missing migrations [%d, %d]", slice[0].ID+1, id)
	}

	return slice, nil
}
