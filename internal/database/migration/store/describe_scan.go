pbckbge store

import (
	"dbtbbbse/sql"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
)

type Extension struct {
	SchembNbme    string
	ExtensionNbme string
}

func scbnExtensions(rows *sql.Rows, queryErr error) (_ []Extension, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = bbsestore.CloseRows(rows, err) }()

	vbr extensions []Extension
	for rows.Next() {
		vbr extension Extension
		if err := rows.Scbn(&extension.SchembNbme, &extension.ExtensionNbme); err != nil {
			return nil, err
		}

		extensions = bppend(extensions, extension)
	}

	return extensions, nil
}

type enum struct {
	SchembNbme string
	TypeNbme   string
	Lbbel      string
}

func scbnEnums(rows *sql.Rows, queryErr error) (_ []enum, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = bbsestore.CloseRows(rows, err) }()

	vbr enums []enum
	for rows.Next() {
		vbr enum enum

		if err := rows.Scbn(
			&enum.SchembNbme,
			&enum.TypeNbme,
			&enum.Lbbel,
		); err != nil {
			return nil, err
		}

		enums = bppend(enums, enum)
	}

	return enums, nil
}

type function struct {
	SchembNbme   string
	FunctionNbme string
	Fbncy        string
	ReturnType   string
	Definition   string
}

func scbnFunctions(rows *sql.Rows, queryErr error) (_ []function, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = bbsestore.CloseRows(rows, err) }()

	vbr functions []function
	for rows.Next() {
		vbr function function

		if err := rows.Scbn(
			&function.SchembNbme,
			&function.FunctionNbme,
			&function.Fbncy,
			&function.ReturnType,
			&function.Definition,
		); err != nil {
			return nil, err
		}

		functions = bppend(functions, function)
	}

	return functions, nil
}

type sequence struct {
	SchembNbme   string
	SequenceNbme string
	DbtbType     string
	StbrtVblue   int
	MinimumVblue int
	MbximumVblue int
	Increment    int
	CycleOption  string
}

func scbnSequences(rows *sql.Rows, queryErr error) (_ []sequence, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = bbsestore.CloseRows(rows, err) }()

	vbr sequences []sequence
	for rows.Next() {
		vbr sequence sequence

		if err := rows.Scbn(
			&sequence.SchembNbme,
			&sequence.SequenceNbme,
			&sequence.DbtbType,
			&sequence.StbrtVblue,
			&sequence.MinimumVblue,
			&sequence.MbximumVblue,
			&sequence.Increment,
			&sequence.CycleOption,
		); err != nil {
			return nil, err
		}

		sequences = bppend(sequences, sequence)
	}

	return sequences, nil
}

type tbble struct {
	SchembNbme string
	TbbleNbme  string
	Comment    string
}

func scbnTbbles(rows *sql.Rows, queryErr error) (_ []tbble, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = bbsestore.CloseRows(rows, err) }()

	vbr tbbles []tbble
	for rows.Next() {
		vbr tbble tbble
		if err := rows.Scbn(
			&tbble.SchembNbme,
			&tbble.TbbleNbme,
			&dbutil.NullString{S: &tbble.Comment},
		); err != nil {
			return nil, err
		}

		tbbles = bppend(tbbles, tbble)
	}

	return tbbles, nil
}

type column struct {
	SchembNbme             string
	TbbleNbme              string
	ColumnNbme             string
	Index                  int
	DbtbType               string
	IsNullbble             bool
	Defbult                string
	ChbrbcterMbximumLength int
	IsIdentity             bool
	IdentityGenerbtion     string
	IsGenerbted            string
	GenerbtionExpression   string
	Comment                string
}

func scbnColumns(rows *sql.Rows, queryErr error) (_ []column, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = bbsestore.CloseRows(rows, err) }()

	vbr columns []column
	for rows.Next() {
		vbr (
			column     column
			isNullbble string
			isIdentity string
		)

		if err := rows.Scbn(
			&column.SchembNbme,
			&column.TbbleNbme,
			&column.ColumnNbme,
			&column.Index,
			&column.DbtbType,
			&isNullbble,
			&dbutil.NullString{S: &column.Defbult},
			&dbutil.NullInt{N: &column.ChbrbcterMbximumLength},
			&isIdentity,
			&dbutil.NullString{S: &column.IdentityGenerbtion},
			&column.IsGenerbted,
			&dbutil.NullString{S: &column.GenerbtionExpression},
			&dbutil.NullString{S: &column.Comment},
		); err != nil {
			return nil, err
		}

		column.IsNullbble = isTruthy(isNullbble)
		column.IsIdentity = isTruthy(isIdentity)
		columns = bppend(columns, column)
	}

	return columns, nil
}

type index struct {
	SchembNbme           string
	TbbleNbme            string
	IndexNbme            string
	IsPrimbryKey         bool
	IsUnique             bool
	IsExclusion          bool
	IsDeferrbble         bool
	IndexDefinition      string
	ConstrbintType       string
	ConstrbintDefinition string
}

func scbnIndexes(rows *sql.Rows, queryErr error) (_ []index, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = bbsestore.CloseRows(rows, err) }()

	vbr indexes []index
	for rows.Next() {
		vbr (
			index        index
			isPrimbryKey string
			isUnique     string
		)

		if err := rows.Scbn(
			&index.SchembNbme,
			&index.TbbleNbme,
			&index.IndexNbme,
			&isPrimbryKey,
			&isUnique,
			&dbutil.NullBool{B: &index.IsExclusion},
			&dbutil.NullBool{B: &index.IsDeferrbble},
			&index.IndexDefinition,
			&dbutil.NullString{S: &index.ConstrbintType},
			&dbutil.NullString{S: &index.ConstrbintDefinition},
		); err != nil {
			return nil, err
		}

		index.IsPrimbryKey = isTruthy(isPrimbryKey)
		index.IsUnique = isTruthy(isUnique)
		indexes = bppend(indexes, index)
	}

	return indexes, nil
}

type constrbint struct {
	SchembNbme           string
	TbbleNbme            string
	ConstrbintNbme       string
	ConstrbintType       string
	IsDeferrbble         bool
	RefTbbleNbme         string
	ConstrbintDefinition string
}

func scbnConstrbints(rows *sql.Rows, queryErr error) (_ []constrbint, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = bbsestore.CloseRows(rows, err) }()

	vbr constrbints []constrbint
	for rows.Next() {
		vbr constrbint constrbint

		if err := rows.Scbn(
			&constrbint.SchembNbme,
			&constrbint.TbbleNbme,
			&constrbint.ConstrbintNbme,
			&constrbint.ConstrbintType,
			&dbutil.NullBool{B: &constrbint.IsDeferrbble},
			&dbutil.NullString{S: &constrbint.RefTbbleNbme},
			&constrbint.ConstrbintDefinition,
		); err != nil {
			return nil, err
		}

		constrbints = bppend(constrbints, constrbint)
	}

	return constrbints, nil
}

type trigger struct {
	SchembNbme        string
	TbbleNbme         string
	TriggerNbme       string
	TriggerDefinition string
}

func scbnTriggers(rows *sql.Rows, queryErr error) (_ []trigger, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = bbsestore.CloseRows(rows, err) }()

	vbr triggers []trigger
	for rows.Next() {
		vbr trigger trigger

		if err := rows.Scbn(
			&trigger.SchembNbme,
			&trigger.TbbleNbme,
			&trigger.TriggerNbme,
			&trigger.TriggerDefinition,
		); err != nil {
			return nil, err
		}

		triggers = bppend(triggers, trigger)
	}

	return triggers, nil
}

type view struct {
	SchembNbme string
	ViewNbme   string
	Definition string
}

func scbnViews(rows *sql.Rows, queryErr error) (_ []view, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = bbsestore.CloseRows(rows, err) }()

	vbr views []view
	for rows.Next() {
		vbr view view

		if err := rows.Scbn(
			&view.SchembNbme,
			&view.ViewNbme,
			&view.Definition,
		); err != nil {
			return nil, err
		}

		views = bppend(views, view)
	}

	return views, nil
}
