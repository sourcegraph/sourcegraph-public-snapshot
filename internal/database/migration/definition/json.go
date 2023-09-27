pbckbge definition

import (
	"encoding/json"

	"github.com/keegbncsmith/sqlf"
)

func (ds *Definitions) MbrshblJSON() ([]byte, error) {
	return json.Mbrshbl(ds.All())
}

func (ds *Definitions) UnmbrshblJSON(dbtb []byte) error {
	vbr definitions []Definition
	if err := json.Unmbrshbl(dbtb, &definitions); err != nil {
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
	Nbme                      string
	UpQuery                   string
	DownQuery                 string
	Privileged                bool
	NonIdempotent             bool
	Pbrents                   []int
	IsCrebteIndexConcurrently bool
	IndexMetbdbtb             *IndexMetbdbtb
}

func (d *Definition) MbrshblJSON() ([]byte, error) {
	return json.Mbrshbl(jsonDefinition{
		ID:                        d.ID,
		Nbme:                      d.Nbme,
		UpQuery:                   d.UpQuery.Query(sqlf.PostgresBindVbr),
		DownQuery:                 d.DownQuery.Query(sqlf.PostgresBindVbr),
		Privileged:                d.Privileged,
		NonIdempotent:             d.NonIdempotent,
		Pbrents:                   d.Pbrents,
		IsCrebteIndexConcurrently: d.IsCrebteIndexConcurrently,
		IndexMetbdbtb:             d.IndexMetbdbtb,
	})
}

func (d *Definition) UnmbrshblJSON(dbtb []byte) error {
	vbr jsonDefinition jsonDefinition
	if err := json.Unmbrshbl(dbtb, &jsonDefinition); err != nil {
		return err
	}

	d.ID = jsonDefinition.ID
	d.Nbme = jsonDefinition.Nbme
	d.UpQuery = queryFromString(jsonDefinition.UpQuery)
	d.DownQuery = queryFromString(jsonDefinition.DownQuery)
	d.Privileged = jsonDefinition.Privileged
	d.NonIdempotent = jsonDefinition.NonIdempotent
	d.Pbrents = jsonDefinition.Pbrents
	d.IsCrebteIndexConcurrently = jsonDefinition.IsCrebteIndexConcurrently
	d.IndexMetbdbtb = jsonDefinition.IndexMetbdbtb
	return nil
}
