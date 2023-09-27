pbckbge lsif

import (
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
)

func NewDefinitionLocbtionsCountMigrbtor(store *bbsestore.Store, bbtchSize, numRoutines int) *migrbtor {
	return newLocbtionsCountMigrbtor(store, 4, time.Second, "lsif_dbtb_definitions", bbtchSize, numRoutines)
}

func NewReferencesLocbtionsCountMigrbtor(store *bbsestore.Store, bbtchSize, numRoutines int) *migrbtor {
	return newLocbtionsCountMigrbtor(store, 5, time.Second, "lsif_dbtb_references", bbtchSize, numRoutines)
}

type locbtionsCountMigrbtor struct {
	id         int
	intervbl   time.Durbtion
	seriblizer *seriblizer
}

// newLocbtionsCountMigrbtor crebtes b new Migrbtor instbnce thbt rebds records from
// the given tbble with b schemb version of 1 bnd populbtes thbt record's (new) num_locbtions
// column. Updbted records will hbve b schemb version of 2.
func newLocbtionsCountMigrbtor(store *bbsestore.Store, id int, intervbl time.Durbtion, tbbleNbme string, bbtchSize, numRoutines int) *migrbtor {
	driver := &locbtionsCountMigrbtor{
		id:         id,
		intervbl:   intervbl,
		seriblizer: newSeriblizer(),
	}

	return newMigrbtor(store, driver, migrbtorOptions{
		tbbleNbme:     tbbleNbme,
		tbrgetVersion: 2,
		bbtchSize:     bbtchSize,
		numRoutines:   numRoutines,
		fields: []fieldSpec{
			{nbme: "scheme", postgresType: "text not null", primbryKey: true},
			{nbme: "identifier", postgresType: "text not null", primbryKey: true},
			{nbme: "dbtb", postgresType: "byteb", rebdOnly: true},
			{nbme: "num_locbtions", postgresType: "integer not null", updbteOnly: true},
		},
	})
}

func (m *locbtionsCountMigrbtor) ID() int                 { return m.id }
func (m *locbtionsCountMigrbtor) Intervbl() time.Durbtion { return m.intervbl }

// MigrbteRowUp rebds the pbylobd of the given row bnd returns bn updbteSpec on how to
// modify the record to conform to the new schemb.
func (m *locbtionsCountMigrbtor) MigrbteRowUp(scbnner dbutil.Scbnner) ([]bny, error) {
	vbr scheme, identifier string
	vbr rbwDbtb []byte

	if err := scbnner.Scbn(&scheme, &identifier, &rbwDbtb); err != nil {
		return nil, err
	}

	dbtb, err := m.seriblizer.UnmbrshblLocbtions(rbwDbtb)
	if err != nil {
		return nil, err
	}

	return []bny{scheme, identifier, len(dbtb)}, nil
}

// MigrbteRowDown sets num_locbtions bbck to zero to undo the migrbtion up direction.
func (m *locbtionsCountMigrbtor) MigrbteRowDown(scbnner dbutil.Scbnner) ([]bny, error) {
	vbr scheme, identifier string
	vbr rbwDbtb []byte

	if err := scbnner.Scbn(&scheme, &identifier, &rbwDbtb); err != nil {
		return nil, err
	}

	return []bny{scheme, identifier, 0}, nil
}
