package definition

import (
	"encoding/json"

	"github.com/keegancsmith/sqlf"
)

func (ds *Definitions) MarshalJSON() ([]byte, error) {
	return json.Marshal(ds.All())
}

func (ds *Definitions) UnmarshalJSON(data []byte) error {
	var definitions []Definition
	if err := json.Unmarshal(data, &definitions); err != nil {
		return err
	}

	newDefinitions, err := NewDefinitions(definitions)
	if err != nil {
		return err
	}

	*ds = *newDefinitions
	return nil
}

type jsonDefinition struct {
	ID                        int
	Name                      string
	UpQuery                   string
	DownQuery                 string
	Privileged                bool
	NonIdempotent             bool
	Parents                   []int
	IsCreateIndexConcurrently bool
	IndexMetadata             *IndexMetadata
}

func (d *Definition) MarshalJSON() ([]byte, error) {
	return json.Marshal(jsonDefinition{
		ID:                        d.ID,
		Name:                      d.Name,
		UpQuery:                   d.UpQuery.Query(sqlf.PostgresBindVar),
		DownQuery:                 d.DownQuery.Query(sqlf.PostgresBindVar),
		Privileged:                d.Privileged,
		NonIdempotent:             d.NonIdempotent,
		Parents:                   d.Parents,
		IsCreateIndexConcurrently: d.IsCreateIndexConcurrently,
		IndexMetadata:             d.IndexMetadata,
	})
}

func (d *Definition) UnmarshalJSON(data []byte) error {
	var jsonDefinition jsonDefinition
	if err := json.Unmarshal(data, &jsonDefinition); err != nil {
		return err
	}

	d.ID = jsonDefinition.ID
	d.Name = jsonDefinition.Name
	d.UpQuery = queryFromString(jsonDefinition.UpQuery)
	d.DownQuery = queryFromString(jsonDefinition.DownQuery)
	d.Privileged = jsonDefinition.Privileged
	d.NonIdempotent = jsonDefinition.NonIdempotent
	d.Parents = jsonDefinition.Parents
	d.IsCreateIndexConcurrently = jsonDefinition.IsCreateIndexConcurrently
	d.IndexMetadata = jsonDefinition.IndexMetadata
	return nil
}
