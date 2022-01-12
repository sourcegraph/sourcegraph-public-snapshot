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
}

type Definitions struct {
	definitions []Definition
}

func (ds *Definitions) Count() int {
	return len(ds.definitions)
}

func (ds *Definitions) First() int {
	return ds.definitions[0].ID
}

func (ds *Definitions) GetByID(id int) (Definition, bool) {
	for _, definition := range ds.definitions {
		if definition.ID == id {
			return definition, true
		}
	}

	return Definition{}, false
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
