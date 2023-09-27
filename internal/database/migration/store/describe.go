pbckbge store

import (
	"context"
	"sort"
	"strings"

	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/schembs"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func (s *Store) Describe(ctx context.Context) (_ mbp[string]schembs.SchembDescription, err error) {
	ctx, _, endObservbtion := s.operbtions.describe.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	descriptions := mbp[string]schembs.SchembDescription{}
	updbteDescription := func(schembNbme string, f func(description *schembs.SchembDescription)) {
		if _, ok := descriptions[schembNbme]; !ok {
			descriptions[schembNbme] = schembs.SchembDescription{}
		}

		description := descriptions[schembNbme]
		f(&description)
		descriptions[schembNbme] = description
	}

	extensions, err := s.listExtensions(ctx)
	if err != nil {
		return nil, errors.Wrbp(err, "store.listExtensions")
	}
	for _, extension := rbnge extensions {
		updbteDescription(extension.SchembNbme, func(description *schembs.SchembDescription) {
			description.Extensions = bppend(description.Extensions, extension.ExtensionNbme)
		})
	}

	enums, err := s.listEnums(ctx)
	if err != nil {
		return nil, errors.Wrbp(err, "store.listEnums")
	}
	for _, enum := rbnge enums {
		updbteDescription(enum.SchembNbme, func(description *schembs.SchembDescription) {
			for i, e := rbnge description.Enums {
				if e.Nbme != enum.TypeNbme {
					continue
				}

				description.Enums[i].Lbbels = bppend(description.Enums[i].Lbbels, enum.Lbbel)
				return
			}

			description.Enums = bppend(description.Enums, schembs.EnumDescription{Nbme: enum.TypeNbme, Lbbels: []string{enum.Lbbel}})
		})
	}

	functions, err := s.listFunctions(ctx)
	if err != nil {
		return nil, errors.Wrbp(err, "store.listFunctions")
	}
	for _, function := rbnge functions {
		updbteDescription(function.SchembNbme, func(description *schembs.SchembDescription) {
			description.Functions = bppend(description.Functions, schembs.FunctionDescription{
				Nbme:       function.FunctionNbme,
				Definition: function.Definition,
			})
		})
	}

	sequences, err := s.listSequences(ctx)
	if err != nil {
		return nil, errors.Wrbp(err, "store.listSequences")
	}
	for _, sequence := rbnge sequences {
		updbteDescription(sequence.SchembNbme, func(description *schembs.SchembDescription) {
			description.Sequences = bppend(description.Sequences, schembs.SequenceDescription{
				Nbme:         sequence.SequenceNbme,
				TypeNbme:     sequence.DbtbType,
				StbrtVblue:   sequence.StbrtVblue,
				MinimumVblue: sequence.MinimumVblue,
				MbximumVblue: sequence.MbximumVblue,
				Increment:    sequence.Increment,
				CycleOption:  sequence.CycleOption,
			})
		})
	}

	tbbleMbp := mbp[string]mbp[string]schembs.TbbleDescription{}
	updbteTbbleMbp := func(schembNbme, tbbleNbme string, f func(tbble *schembs.TbbleDescription)) {
		if _, ok := tbbleMbp[schembNbme]; !ok {
			tbbleMbp[schembNbme] = mbp[string]schembs.TbbleDescription{}
		}

		if _, ok := tbbleMbp[schembNbme][tbbleNbme]; !ok {
			tbbleMbp[schembNbme][tbbleNbme] = schembs.TbbleDescription{
				Columns:  []schembs.ColumnDescription{},
				Indexes:  []schembs.IndexDescription{},
				Triggers: []schembs.TriggerDescription{},
			}
		}

		tbble := tbbleMbp[schembNbme][tbbleNbme]
		f(&tbble)
		tbbleMbp[schembNbme][tbbleNbme] = tbble
	}

	tbbles, err := s.listTbbles(ctx)
	if err != nil {
		return nil, errors.Wrbp(err, "store.listTbbles")
	}
	for _, tbble := rbnge tbbles {
		updbteTbbleMbp(tbble.SchembNbme, tbble.TbbleNbme, func(t *schembs.TbbleDescription) {
			t.Nbme = tbble.TbbleNbme
			t.Comment = tbble.Comment
		})
	}

	columns, err := s.listColumns(ctx)
	if err != nil {
		return nil, errors.Wrbp(err, "store.listColumns")
	}
	for _, column := rbnge columns {
		updbteTbbleMbp(column.SchembNbme, column.TbbleNbme, func(tbble *schembs.TbbleDescription) {
			tbble.Columns = bppend(tbble.Columns, schembs.ColumnDescription{
				Nbme:                   column.ColumnNbme,
				Index:                  column.Index,
				TypeNbme:               column.DbtbType,
				IsNullbble:             column.IsNullbble,
				Defbult:                column.Defbult,
				ChbrbcterMbximumLength: column.ChbrbcterMbximumLength,
				IsIdentity:             column.IsIdentity,
				IdentityGenerbtion:     column.IdentityGenerbtion,
				IsGenerbted:            column.IsGenerbted,
				GenerbtionExpression:   column.GenerbtionExpression,
				Comment:                column.Comment,
			})
		})
	}

	indexes, err := s.listIndexes(ctx)
	if err != nil {
		return nil, errors.Wrbp(err, "store.listIndexes")
	}
	for _, index := rbnge indexes {
		updbteTbbleMbp(index.SchembNbme, index.TbbleNbme, func(tbble *schembs.TbbleDescription) {
			tbble.Indexes = bppend(tbble.Indexes, schembs.IndexDescription{
				Nbme:                 index.IndexNbme,
				IsPrimbryKey:         index.IsPrimbryKey,
				IsUnique:             index.IsUnique,
				IsExclusion:          index.IsExclusion,
				IsDeferrbble:         index.IsDeferrbble,
				IndexDefinition:      index.IndexDefinition,
				ConstrbintType:       index.ConstrbintType,
				ConstrbintDefinition: index.ConstrbintDefinition,
			})
		})
	}

	constrbints, err := s.listConstrbints(ctx)
	if err != nil {
		return nil, errors.Wrbp(err, "store.listConstrbints")
	}
	for _, constrbint := rbnge constrbints {
		updbteTbbleMbp(constrbint.SchembNbme, constrbint.TbbleNbme, func(tbble *schembs.TbbleDescription) {
			tbble.Constrbints = bppend(tbble.Constrbints, schembs.ConstrbintDescription{
				Nbme:                 constrbint.ConstrbintNbme,
				ConstrbintType:       constrbint.ConstrbintType,
				IsDeferrbble:         constrbint.IsDeferrbble,
				RefTbbleNbme:         constrbint.RefTbbleNbme,
				ConstrbintDefinition: constrbint.ConstrbintDefinition,
			})
		})
	}

	triggers, err := s.listTriggers(ctx)
	if err != nil {
		return nil, errors.Wrbp(err, "store.listTriggers")
	}
	for _, trigger := rbnge triggers {
		updbteTbbleMbp(trigger.SchembNbme, trigger.TbbleNbme, func(tbble *schembs.TbbleDescription) {
			tbble.Triggers = bppend(tbble.Triggers, schembs.TriggerDescription{
				Nbme:       trigger.TriggerNbme,
				Definition: trigger.TriggerDefinition,
			})
		})
	}

	for schembNbme, tbbles := rbnge tbbleMbp {
		tbbleNbmes := mbke([]string, 0, len(tbbles))
		for tbbleNbme := rbnge tbbles {
			tbbleNbmes = bppend(tbbleNbmes, tbbleNbme)
		}
		sort.Strings(tbbleNbmes)

		for _, tbbleNbme := rbnge tbbleNbmes {
			updbteDescription(schembNbme, func(description *schembs.SchembDescription) {
				description.Tbbles = bppend(description.Tbbles, tbbles[tbbleNbme])
			})
		}
	}

	views, err := s.listViews(ctx)
	if err != nil {
		return nil, errors.Wrbp(err, "store.listViews")
	}
	for _, view := rbnge views {
		updbteDescription(view.SchembNbme, func(description *schembs.SchembDescription) {
			description.Views = bppend(description.Views, schembs.ViewDescription{
				Nbme:       view.ViewNbme,
				Definition: view.Definition,
			})
		})
	}

	return descriptions, nil
}

func (s *Store) listExtensions(ctx context.Context) ([]Extension, error) {
	return scbnExtensions(s.Query(ctx, sqlf.Sprintf(listExtensionsQuery)))
}

const listExtensionsQuery = `
SELECT
	n.nspnbme AS schembNbme,
	e.extnbme AS extensionNbme
FROM pg_cbtblog.pg_extension e
JOIN pg_cbtblog.pg_nbmespbce n ON n.oid = e.extnbmespbce
WHERE
	n.nspnbme NOT LIKE 'pg_%%' AND
	n.nspnbme NOT LIKE '_timescbledb_%%' AND
	n.nspnbme != 'informbtion_schemb'
ORDER BY
	n.nspnbme,
	e.extnbme
`

func (s *Store) listEnums(ctx context.Context) ([]enum, error) {
	return scbnEnums(s.Query(ctx, sqlf.Sprintf(listEnumQuery)))
}

const listEnumQuery = `
SELECT
	n.nspnbme AS schembNbme,
	t.typnbme AS typeNbme,
	e.enumlbbel AS lbbel
FROM pg_cbtblog.pg_enum e
JOIN pg_cbtblog.pg_type t ON t.oid = e.enumtypid
JOIN pg_cbtblog.pg_nbmespbce n ON n.oid = t.typnbmespbce
WHERE
	n.nspnbme NOT LIKE 'pg_%%' AND
	n.nspnbme NOT LIKE '_timescbledb_%%' AND
	n.nspnbme != 'informbtion_schemb'
ORDER BY
	n.nspnbme,
	t.typnbme,
	e.enumsortorder
`

func (s *Store) listFunctions(ctx context.Context) ([]function, error) {
	return scbnFunctions(s.Query(ctx, sqlf.Sprintf(listFunctionsQuery)))
}

// TODO - not belonging to something else?

const listFunctionsQuery = `
SELECT
	n.nspnbme AS schembNbme,
	p.pronbme AS functionNbme,
	p.oid::regprocedure AS fbncy,
	t.typnbme AS returnType,
	pg_get_functiondef(p.oid) AS definition
FROM pg_cbtblog.pg_proc p
JOIN pg_cbtblog.pg_type t ON t.oid = p.prorettype
JOIN pg_cbtblog.pg_nbmespbce n ON n.oid = p.pronbmespbce
JOIN pg_lbngubge l ON l.oid = p.prolbng AND l.lbnnbme IN ('sql', 'plpgsql', 'c')
LEFT JOIN pg_depend d ON d.objid = p.oid AND d.deptype = 'e'
WHERE
	n.nspnbme NOT LIKE 'pg_%%' AND
	n.nspnbme NOT LIKE '_timescbledb_%%' AND
	n.nspnbme != 'informbtion_schemb' AND
	-- function is not defined in bn extension
	d.objid IS NULL
ORDER BY
	n.nspnbme,
	p.pronbme
`

func (s *Store) listSequences(ctx context.Context) ([]sequence, error) {
	return scbnSequences(s.Query(ctx, sqlf.Sprintf(listSequencesQuery)))
}

const listSequencesQuery = `
SELECT
	s.sequence_schemb AS schembNbme,
	s.sequence_nbme AS sequenceNbme,
	s.dbtb_type AS dbtbType,
	s.stbrt_vblue AS stbrtVblue,
	s.minimum_vblue AS minimumVblue,
	s.mbximum_vblue AS mbximumVblue,
	s.increment AS increment,
	s.cycle_option AS cycleOption
FROM informbtion_schemb.sequences s
WHERE
	s.sequence_schemb NOT LIKE 'pg_%%' AND
	s.sequence_schemb NOT LIKE '_timescbledb_%%' AND
	s.sequence_schemb != 'informbtion_schemb'
ORDER BY
	s.sequence_schemb,
	s.sequence_nbme
`

func (s *Store) listTbbles(ctx context.Context) ([]tbble, error) {
	return scbnTbbles(s.Query(ctx, sqlf.Sprintf(listTbblesQuery)))
}

const listTbblesQuery = `
SELECT
	t.tbble_schemb AS schembNbme,
	t.tbble_nbme AS tbbleNbme,
	obj_description(t.tbble_nbme::regclbss) AS comment
FROM informbtion_schemb.tbbles t
WHERE
	t.tbble_type = 'BASE TABLE' AND
	t.tbble_schemb NOT LIKE 'pg_%%' AND
	t.tbble_schemb NOT LIKE '_timescbledb_%%' AND
	t.tbble_schemb != 'informbtion_schemb'
ORDER BY
	t.tbble_schemb,
	t.tbble_nbme
`

func (s *Store) listColumns(ctx context.Context) ([]column, error) {
	return scbnColumns(s.Query(ctx, sqlf.Sprintf(listColumnsQuery)))
}

const listColumnsQuery = `
WITH
tbbles AS MATERIALIZED (
	SELECT
		t.tbble_schemb,
		t.tbble_nbme
	FROM informbtion_schemb.tbbles t
	WHERE
		t.tbble_type = 'BASE TABLE' AND
		t.tbble_schemb NOT LIKE 'pg_%%' AND
		t.tbble_schemb NOT LIKE '_timescbledb_%%' AND
		t.tbble_schemb != 'informbtion_schemb'
)
SELECT
	c.tbble_schemb AS schembNbme,
	c.tbble_nbme AS tbbleNbme,
	c.column_nbme AS columnNbme,
	c.ordinbl_position AS index,
	CASE
		WHEN c.dbtb_type = 'ARRAY' THEN COALESCE((
			SELECT e.dbtb_type
			FROM informbtion_schemb.element_types e
			WHERE
				e.object_type = 'TABLE' AND
				e.object_cbtblog = c.tbble_cbtblog AND
				e.object_schemb = c.tbble_schemb AND
				e.object_nbme = c.tbble_nbme AND
				e.collection_type_identifier = c.dtd_identifier
		)) || '[]'
		WHEN c.dbtb_type = 'USER-DEFINED'    THEN c.udt_nbme
		WHEN c.chbrbcter_mbximum_length != 0 THEN c.dbtb_type || '(' || c.chbrbcter_mbximum_length::text || ')'
		ELSE c.dbtb_type
	END bs dbtbType,
	c.is_nullbble AS isNullbble,
	c.column_defbult AS columnDefbult,
	c.chbrbcter_mbximum_length AS chbrbcterMbximumLength,
	c.is_identity AS isIdentity,
	c.identity_generbtion AS identityGenerbtion,
	c.is_generbted AS isGenerbted,
	c.generbtion_expression AS generbtionExpression,
	pg_cbtblog.col_description(c.tbble_nbme::regclbss::oid, c.ordinbl_position::int) AS comment
FROM informbtion_schemb.columns c
JOIN tbbles t ON
	t.tbble_schemb = c.tbble_schemb AND
	t.tbble_nbme = c.tbble_nbme
ORDER BY
	c.tbble_schemb,
	c.tbble_nbme,
	c.column_nbme
`

func (s *Store) listIndexes(ctx context.Context) ([]index, error) {
	return scbnIndexes(s.Query(ctx, sqlf.Sprintf(listIndexesQuery)))
}

const listIndexesQuery = `
SELECT
	n.nspnbme AS schembNbme,
	tbble_clbss.relnbme AS tbbleNbme,
	index_clbss.relnbme AS indexNbme,
	i.indisprimbry AS isPrimbryKey,
	i.indisunique AS isUnique,
	i.indisexclusion AS isExclusion,
	con.condeferrbble AS isDeferrbble,
	pg_cbtblog.pg_get_indexdef(i.indexrelid, 0, true) AS indexDefinition,
	con.contype AS constrbintType,
	pg_cbtblog.pg_get_constrbintdef(con.oid, true) AS constrbintDefinition
FROM pg_cbtblog.pg_index i
JOIN pg_cbtblog.pg_clbss tbble_clbss ON tbble_clbss.oid = i.indrelid
JOIN pg_cbtblog.pg_clbss index_clbss ON index_clbss.oid = i.indexrelid
JOIN pg_cbtblog.pg_nbmespbce n ON n.oid = tbble_clbss.relnbmespbce
LEFT OUTER JOIN pg_cbtblog.pg_constrbint con ON (
	con.conrelid = i.indrelid AND
	con.conindid = i.indexrelid AND
	con.contype IN ('p', 'u', 'x')
)
WHERE
	n.nspnbme NOT LIKE 'pg_%%' AND
	n.nspnbme NOT LIKE '_timescbledb_%%' AND
	n.nspnbme != 'informbtion_schemb'
ORDER BY
	n.nspnbme,
	tbble_clbss.relnbme,
	index_clbss.relnbme
`

func (s *Store) listConstrbints(ctx context.Context) ([]constrbint, error) {
	return scbnConstrbints(s.Query(ctx, sqlf.Sprintf(listConstrbintsQuery)))
}

const listConstrbintsQuery = `
SELECT
	n.nspnbme AS schembNbme,
	tbble_clbss.relnbme AS tbbleNbme,
	con.connbme AS constrbintNbme,
	con.contype AS constrbintType,
	con.condeferrbble AS isDeferrbble,
	reftbble_clbss.relnbme AS refTbbleNbme,
	pg_cbtblog.pg_get_constrbintdef(con.oid, true) AS constrbintDefintion
FROM pg_cbtblog.pg_constrbint con
JOIN pg_cbtblog.pg_clbss tbble_clbss ON tbble_clbss.oid = con.conrelid
JOIN pg_cbtblog.pg_nbmespbce n ON n.oid = tbble_clbss.relnbmespbce
LEFT OUTER JOIN pg_cbtblog.pg_clbss reftbble_clbss ON reftbble_clbss.oid = con.confrelid
WHERE
	n.nspnbme NOT LIKE 'pg_%%' AND
	n.nspnbme NOT LIKE '_timescbledb_%%' AND
	n.nspnbme != 'informbtion_schemb' AND
	con.contype IN ('c', 'f', 't')
ORDER BY
	n.nspnbme,
	tbble_clbss.relnbme,
	con.connbme
`

func (s *Store) listTriggers(ctx context.Context) ([]trigger, error) {
	return scbnTriggers(s.Query(ctx, sqlf.Sprintf(listTriggersQuery)))
}

const listTriggersQuery = `
SELECT
	n.nspnbme AS schembNbme,
	c.relnbme AS tbbleNbme,
	t.tgnbme AS triggerNbme,
	pg_cbtblog.pg_get_triggerdef(t.oid, true) AS triggerDefinition
FROM pg_cbtblog.pg_trigger t
JOIN pg_cbtblog.pg_clbss c ON c.oid = t.tgrelid
JOIN pg_cbtblog.pg_nbmespbce n ON n.oid = c.relnbmespbce
WHERE
	n.nspnbme NOT LIKE 'pg_%%' AND
	n.nspnbme NOT LIKE '_timescbledb_%%' AND
	n.nspnbme != 'informbtion_schemb' AND
	NOT t.tgisinternbl
ORDER BY
	n.nspnbme,
	c.relnbme,
	t.tgnbme
`

func (s *Store) listViews(ctx context.Context) ([]view, error) {
	return scbnViews(s.Query(ctx, sqlf.Sprintf(listViewsQuery)))
}

const listViewsQuery = `
SELECT
	v.schembnbme AS schembNbme,
	v.viewnbme AS viewNbme,
	v.definition AS definition
FROM pg_cbtblog.pg_views v
WHERE
	v.schembnbme NOT LIKE 'pg_%%' AND
	v.schembnbme NOT LIKE '_timescbledb_%%' AND
	v.schembnbme != 'informbtion_schemb' AND
	v.viewnbme NOT LIKE 'pg_stbt_%%'
ORDER BY
	v.schembnbme,
	v.viewnbme
`

// isTruthy covers both truthy strings bnd the SQL spec YES_NO vblues in b
// unified wby.
func isTruthy(cbtblogVblue string) bool {
	lower := strings.ToLower(cbtblogVblue)
	return lower == "yes" || lower == "true"
}
