package modl

// Changes Copyright 2013 Jason Moiron.  Original Gorp code
// Copyright 2012 James Cooper. All rights reserved.
//
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.
//
// Source code and project home:
// https://github.com/jmoiron/modl

import (
	"database/sql"
	"fmt"
	"reflect"
)

// NoKeysErr is a special error type returned when modl's CRUD helpers are
// used on tables which have not been set up with a primary key.
type NoKeysErr struct {
	Table *TableMap
}

// Error returns the string representation of a NoKeysError.
func (n NoKeysErr) Error() string {
	return fmt.Sprintf("Could not find keys for table %v", n.Table)
}

const versFieldConst = "[modl_ver_field]"

// OptimisticLockError is returned by Update() or Delete() if the
// struct being modified has a Version field and the value is not equal to
// the current value in the database
type OptimisticLockError struct {
	// Table name where the lock error occurred
	TableName string

	// Primary key values of the row being updated/deleted
	Keys []interface{}

	// true if a row was found with those keys, indicating the
	// LocalVersion is stale.  false if no value was found with those
	// keys, suggesting the row has been deleted since loaded, or
	// was never inserted to begin with
	RowExists bool

	// Version value on the struct passed to Update/Delete. This value is
	// out of sync with the database.
	LocalVersion int64
}

// Error returns a description of the cause of the lock error
func (e OptimisticLockError) Error() string {
	if e.RowExists {
		return fmt.Sprintf("OptimisticLockError table=%s keys=%v out of date version=%d", e.TableName, e.Keys, e.LocalVersion)
	}

	return fmt.Sprintf("OptimisticLockError no row found for table=%s keys=%v", e.TableName, e.Keys)
}

// A bindPlan saves a query type (insert, get, updated, delete) so it doesn't
// have to be re-created every time it's executed.
type bindPlan struct {
	query       string
	argFields   []string
	keyFields   []string
	versField   string
	autoIncrIdx []int // Go struct field index
}

func (plan bindPlan) createBindInstance(elem reflect.Value) bindInstance {
	bi := bindInstance{query: plan.query, autoIncrIdx: plan.autoIncrIdx, versField: plan.versField}
	if plan.versField != "" {
		bi.existingVersion = elem.FieldByName(plan.versField).Int()
	}

	for i := 0; i < len(plan.argFields); i++ {
		k := plan.argFields[i]
		if k == versFieldConst {
			newVer := bi.existingVersion + 1
			bi.args = append(bi.args, newVer)
			if bi.existingVersion == 0 {
				elem.FieldByName(plan.versField).SetInt(int64(newVer))
			}
		} else {
			val := elem.FieldByName(k).Interface()
			bi.args = append(bi.args, val)
		}
	}

	for i := 0; i < len(plan.keyFields); i++ {
		k := plan.keyFields[i]
		val := elem.FieldByName(k).Interface()
		bi.keys = append(bi.keys, val)
	}

	return bi
}

type bindInstance struct {
	query           string
	args            []interface{}
	keys            []interface{}
	existingVersion int64
	versField       string
	autoIncrIdx     []int // the autoincr. column's Go struct field index
}

// SqlExecutor exposes modl operations that can be run from Pre/Post
// hooks.  This hides whether the current operation that triggered the
// hook is in a transaction.
//
// See the DbMap function docs for each of the functions below for more
// information.
type SqlExecutor interface {
	Get(dest interface{}, keys ...interface{}) error
	Insert(list ...interface{}) error
	Update(list ...interface{}) (int64, error)
	Delete(list ...interface{}) (int64, error)
	Exec(query string, args ...interface{}) (sql.Result, error)
	Select(dest interface{}, query string, args ...interface{}) error
	SelectOne(dest interface{}, query string, args ...interface{}) error
	handle() handle
}

// Compile-time check that DbMap and Transaction implement the SqlExecutor
// interface.
var (
        _ SqlExecutor = &DbMap{}
        _ SqlExecutor = &Transaction{}
)

///////////////

func hookedget(m *DbMap, e SqlExecutor, dest interface{}, query string, args ...interface{}) error {
	err := e.handle().Get(dest, query, args...)
	if err != nil {
		return err
	}

	table := m.TableFor(dest)

	if table != nil && table.CanPostGet {
		err = dest.(PostGetter).PostGet(e)
		if err != nil {
			return err
		}
	}
	return nil
}

func hookedselect(m *DbMap, e SqlExecutor, dest interface{}, query string, args ...interface{}) error {

	err := e.handle().Select(dest, query, args...)
	if err != nil {
		return err
	}

	// select can use arbitrary structs for join queries, so we needn't find a table
	table := m.TableFor(dest)

	if table != nil && table.CanPostGet {
		var x interface{}
		v := reflect.ValueOf(dest)
		if v.Kind() == reflect.Ptr {
			v = reflect.Indirect(v)
		}
		l := v.Len()
		for i := 0; i < l; i++ {
			x = v.Index(i).Interface()
			err = x.(PostGetter).PostGet(e)
			if err != nil {
				return err
			}

		}
	}
	return nil
}

func get(m *DbMap, e SqlExecutor, dest interface{}, keys ...interface{}) error {

	table := m.TableFor(dest)

	if table == nil {
		return fmt.Errorf("could not find table for %v", dest)
	}
	if len(table.Keys) < 1 {
		return &NoKeysErr{table}
	}

	plan := table.bindGet()
	err := e.handle().Get(dest, plan.query, keys...)

	if err != nil {
		return err
	}

	if table.CanPostGet {
		err = dest.(PostGetter).PostGet(e)
		if err != nil {
			return err
		}
	}

	return nil
}

func deletes(m *DbMap, e SqlExecutor, list ...interface{}) (int64, error) {
	var err error
	var table *TableMap
	var elem reflect.Value
	var count int64

	for _, ptr := range list {
		table, elem, err = tableForPointer(m, ptr, true)
		if err != nil {
			return -1, err
		}

		if table.CanPreDelete {
			err = ptr.(PreDeleter).PreDelete(e)
			if err != nil {
				return -1, err
			}
		}

		bi := table.bindDelete(elem)

		res, err := e.Exec(bi.query, bi.args...)
		if err != nil {
			return -1, err
		}

		rows, err := res.RowsAffected()
		if err != nil {
			return -1, err
		}

		if rows == 0 && bi.existingVersion > 0 {
			return lockError(m, e, table.TableName, bi.existingVersion, elem, bi.keys...)
		}

		count += rows

		if table.CanPostDelete {
			err = ptr.(PostDeleter).PostDelete(e)
			if err != nil {
				return -1, err
			}
		}
	}

	return count, nil
}

func update(m *DbMap, e SqlExecutor, list ...interface{}) (int64, error) {
	var err error
	var table *TableMap
	var elem reflect.Value
	var count int64

	for _, ptr := range list {
		table, elem, err = tableForPointer(m, ptr, true)
		if err != nil {
			return -1, err
		}

		if table.CanPreUpdate {
			err = ptr.(PreUpdater).PreUpdate(e)
			if err != nil {
				return -1, err
			}
		}

		bi := table.bindUpdate(elem)
		if err != nil {
			return -1, err
		}

		res, err := e.Exec(bi.query, bi.args...)
		if err != nil {
			return -1, err
		}

		rows, err := res.RowsAffected()
		if err != nil {
			return -1, err
		}

		if rows == 0 && bi.existingVersion > 0 {
			return lockError(m, e, table.TableName,
				bi.existingVersion, elem, bi.keys...)
		}

		if bi.versField != "" {
			elem.FieldByName(bi.versField).SetInt(bi.existingVersion + 1)
		}

		count += rows

		if table.CanPostUpdate {
			err = ptr.(PostUpdater).PostUpdate(e)

			if err != nil {
				return -1, err
			}
		}
	}
	return count, nil
}

func insert(m *DbMap, e SqlExecutor, list ...interface{}) error {
	var err error
	var table *TableMap
	var elem reflect.Value

	for _, ptr := range list {
		table, elem, err = tableForPointer(m, ptr, false)
		if err != nil {
			return err
		}

		if table.CanPreInsert {
			err = ptr.(PreInserter).PreInsert(e)
			if err != nil {
				return err
			}
		}

		bi := table.bindInsert(elem)

		if bi.autoIncrIdx != nil {
			id, err := m.Dialect.InsertAutoIncr(e, bi.query, bi.args...)
			if err != nil {
				return err
			}
			f := elem.FieldByIndex(bi.autoIncrIdx)
			k := f.Kind()
			if (k == reflect.Int) || (k == reflect.Int16) || (k == reflect.Int32) || (k == reflect.Int64) {
				f.SetInt(id)
			} else {
				return fmt.Errorf("modl: Cannot set autoincrement value on non-Int field. SQL=%s  autoIncrIdx=%d", bi.query, bi.autoIncrIdx)
			}
		} else {
			_, err := e.Exec(bi.query, bi.args...)
			if err != nil {
				return err
			}
		}

		if table.CanPostInsert {
			err = ptr.(PostInserter).PostInsert(e)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func lockError(m *DbMap, e SqlExecutor, tableName string, existingVer int64, elem reflect.Value, keys ...interface{}) (int64, error) {

	dest := reflect.New(elem.Type()).Interface()
	err := get(m, e, dest, keys...)
	if err != nil {
		return -1, err
	}

	ole := OptimisticLockError{tableName, keys, true, existingVer}
	if dest == nil {
		ole.RowExists = false
	}
	return -1, ole
}
