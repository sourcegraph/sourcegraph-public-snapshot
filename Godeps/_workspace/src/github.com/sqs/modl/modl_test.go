package modl

import (
	"bytes"
	"database/sql"
	"fmt"
	"log"
	"os"
	"reflect"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

var _ = log.Fatal

type Invoice struct {
	ID       int64
	Created  int64 `db:"date_created"`
	Updated  int64
	Memo     string
	PersonID int64
	IsPaid   bool
}

type Person struct {
	ID      int64
	Created int64
	Updated int64
	FName   string
	LName   string
	Version int64
}

type InvoicePersonView struct {
	InvoiceID     int64
	PersonID      int64
	Memo          string
	FName         string
	LegacyVersion int64
}

type TableWithNull struct {
	ID      int64
	Str     sql.NullString
	Int64   sql.NullInt64
	Float64 sql.NullFloat64
	Bool    sql.NullBool
	Bytes   []byte
}

type WithIgnoredColumn struct {
	internal int64 `db:"-"`
	ID       int64
	Created  int64
}

type WithStringPk struct {
	ID   string
	Name string
}

type CustomStringType string

type WithEmbeddedStruct struct {
	Id int64
	Names
}

type Names struct {
	FirstName string
	LastName  string
}

func (p *Person) PreInsert(s SqlExecutor) error {
	p.Created = time.Now().UnixNano()
	p.Updated = p.Created
	if p.FName == "badname" {
		return fmt.Errorf("invalid name: %s", p.FName)
	}
	return nil
}

func (p *Person) PostInsert(s SqlExecutor) error {
	p.LName = "postinsert"
	return nil
}

func (p *Person) PreUpdate(s SqlExecutor) error {
	p.FName = "preupdate"
	return nil
}

func (p *Person) PostUpdate(s SqlExecutor) error {
	p.LName = "postupdate"
	return nil
}

func (p *Person) PreDelete(s SqlExecutor) error {
	p.FName = "predelete"
	return nil
}

func (p *Person) PostDelete(s SqlExecutor) error {
	p.LName = "postdelete"
	return nil
}

func (p *Person) PostGet(s SqlExecutor) error {
	p.LName = "postget"
	return nil
}

type PersistentUser struct {
	Key            int32 `db:"mykey"`
	ID             string
	PassedTraining bool
}

func TestCreateTablesIfNotExists(t *testing.T) {
	dbmap := initDbMap()
	defer dbmap.Cleanup()

	err := dbmap.CreateTablesIfNotExists()
	if err != nil {
		t.Error(err)
	}
}

func TestPersistentUser(t *testing.T) {
	dbmap := newDbMap()
	dbmap.Exec("drop table if exists persistentuser")
	if len(os.Getenv("MODL_TEST_TRACE")) > 0 {
		dbmap.TraceOn("test", log.New(os.Stdout, "modltest: ", log.Lmicroseconds))
	}
	dbmap.AddTable(PersistentUser{}).SetKeys(false, "mykey")
	err := dbmap.CreateTablesIfNotExists()
	if err != nil {
		panic(err)
	}
	defer dbmap.Cleanup()
	pu := &PersistentUser{43, "33r", false}
	err = dbmap.Insert(pu)
	if err != nil {
		panic(err)
	}

	// prove we can pass a pointer into Get
	pu2 := &PersistentUser{}
	err = dbmap.Get(pu2, pu.Key)
	if err != nil {
		panic(err)
	}
	if !reflect.DeepEqual(pu, pu2) {
		t.Errorf("%v!=%v", pu, pu2)
	}

	arr := []*PersistentUser{}
	err = dbmap.Select(&arr, "select * from persistentuser")
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(pu, arr[0]) {
		t.Errorf("%v!=%v", pu, arr[0])
	}

	// prove we can get the results back in a slice
	puArr := []PersistentUser{}
	err = dbmap.Select(&puArr, "select * from persistentuser")
	if err != nil {
		t.Error(err)
	}
	if len(puArr) != 1 {
		t.Errorf("Expected one persistentuser, found none")
	}
	if !reflect.DeepEqual(pu, &puArr[0]) {
		t.Errorf("%v!=%v", pu, puArr[0])
	}
}

func TestOverrideVersionCol(t *testing.T) {
	dbmap := initDbMap()
	dbmap.DropTables()

	t1 := dbmap.AddTable(InvoicePersonView{}).SetKeys(false, "invoiceid", "personid")
	err := dbmap.CreateTables()

	if err != nil {
		panic(err)
	}
	defer dbmap.Cleanup()
	c1 := t1.SetVersionCol("legacyversion")
	if c1.ColumnName != "legacyversion" {
		t.Errorf("Wrong col returned: %v", c1)
	}

	ipv := &InvoicePersonView{1, 2, "memo", "fname", 0}
	_update(dbmap, ipv)
	if ipv.LegacyVersion != 1 {
		t.Errorf("LegacyVersion not updated: %d", ipv.LegacyVersion)
	}
}

func TestDontPanicOnInsert(t *testing.T) {
	var err error
	dbmap := initDbMap()
	defer dbmap.Cleanup()

	err = dbmap.Insert(&TableWithNull{ID: 10})
	if err == nil {
		t.Errorf("Should have received an error for inserting without a known table.")
	}
}

func TestOptimisticLocking(t *testing.T) {
	var err error
	dbmap := initDbMap()
	defer dbmap.Cleanup()

	p1 := &Person{0, 0, 0, "Bob", "Smith", 0}
	dbmap.Insert(p1) // Version is now 1
	if p1.Version != 1 {
		t.Errorf("Insert didn't incr Version: %d != %d", 1, p1.Version)
		return
	}
	if p1.ID == 0 {
		t.Errorf("Insert didn't return a generated PK")
		return
	}

	p2 := &Person{}
	err = dbmap.Get(p2, p1.ID)
	if err != nil {
		panic(err)
	}
	p2.LName = "Edwards"
	_, err = dbmap.Update(p2) // Version is now 2

	if err != nil {
		panic(err)
	}

	if p2.Version != 2 {
		t.Errorf("Update didn't incr Version: %d != %d", 2, p2.Version)
	}

	p1.LName = "Howard"
	count, err := dbmap.Update(p1)
	if _, ok := err.(OptimisticLockError); !ok {
		t.Errorf("update - Expected OptimisticLockError, got: %v", err)
	}
	if count != -1 {
		t.Errorf("update - Expected -1 count, got: %d", count)
	}

	count, err = dbmap.Delete(p1)
	if _, ok := err.(OptimisticLockError); !ok {
		t.Errorf("delete - Expected OptimisticLockError, got: %v", err)
	}
	if count != -1 {
		t.Errorf("delete - Expected -1 count, got: %d", count)
	}
}

// what happens if a legacy table has a null value?
func TestDoubleAddTable(t *testing.T) {
	dbmap := newDbMap()
	t1 := dbmap.AddTable(TableWithNull{}).SetKeys(false, "ID")
	t2 := dbmap.AddTable(TableWithNull{})
	if t1 != t2 {
		t.Errorf("%v != %v", t1, t2)
	}
}

// test overriding the create sql
func TestColMapCreateSql(t *testing.T) {
	dbmap := newDbMap()
	t1 := dbmap.AddTable(TableWithNull{})
	b := t1.ColMap("Bytes")
	custom := "bytes text NOT NULL"
	b.SetSqlCreate(custom)
	var buf bytes.Buffer
	writeColumnSql(&buf, b)
	s := buf.String()
	if s != custom {
		t.Errorf("Expected custom sql `%s`, got %s", custom, s)
	}
	err := dbmap.CreateTables()
	defer dbmap.Cleanup()
	if err != nil {
		t.Error(err)
	}
}

// what happens if a legacy table has a null value?
func TestNullValues(t *testing.T) {
	dbmap := initDbMapNulls()
	defer dbmap.Cleanup()

	// insert a row directly
	_, err := dbmap.Exec(`insert into tablewithnull values (10, null, null, null, null, null)`)
	if err != nil {
		panic(err)
	}

	// try to load it
	expected := &TableWithNull{ID: 10}
	t1 := &TableWithNull{}
	MustGet(dbmap, t1, 10)
	if !reflect.DeepEqual(expected, t1) {
		t.Errorf("%v != %v", expected, t1)
	}

	// update it
	t1.Str = sql.NullString{"hi", true}
	expected.Str = t1.Str
	t1.Int64 = sql.NullInt64{999, true}
	expected.Int64 = t1.Int64
	t1.Float64 = sql.NullFloat64{53.33, true}
	expected.Float64 = t1.Float64
	t1.Bool = sql.NullBool{true, true}
	expected.Bool = t1.Bool
	t1.Bytes = []byte{1, 30, 31, 33}
	expected.Bytes = t1.Bytes
	_update(dbmap, t1)

	MustGet(dbmap, t1, 10)
	if t1.Str.String != "hi" {
		t.Errorf("%s != hi", t1.Str.String)
	}
	if !reflect.DeepEqual(expected, t1) {
		t.Errorf("%v != %v", expected, t1)
	}
}

func TestColumnProps(t *testing.T) {
	dbmap := newDbMap()
	//dbmap.TraceOn("", log.New(os.Stdout, "modltest: ", log.Lmicroseconds))
	t1 := dbmap.AddTable(Invoice{}).SetKeys(true, "ID")
	//t1.ColMap("Created").Rename("date_created")
	t1.ColMap("Updated").SetTransient(true)
	t1.ColMap("Memo").SetMaxSize(10)
	t1.ColMap("PersonID").SetUnique(true)

	err := dbmap.CreateTables()
	if err != nil {
		panic(err)
	}
	defer dbmap.Cleanup()

	// test transient
	inv := &Invoice{0, 0, 1, "my invoice", 0, true}
	_insert(dbmap, inv)
	inv2 := Invoice{}
	MustGet(dbmap, &inv2, inv.ID)
	if inv2.Updated != 0 {
		t.Errorf("Saved transient column 'Updated'")
	}

	// test max size
	inv2.Memo = "this memo is too long"
	err = dbmap.Insert(inv2)
	if err == nil {
		t.Errorf("max size exceeded, but Insert did not fail.")
	}

	// test unique - same person id
	inv = &Invoice{0, 0, 1, "my invoice2", 0, false}
	err = dbmap.Insert(inv)
	if err == nil {
		t.Errorf("same PersonID inserted, but Insert did not fail.")
	}
}

func TestAutoIncrOverride(t *testing.T) {
	dbmap := newDbMap()
	type T struct {
		A  string
		ID int
		B  string
	}
	//dbmap.TraceOn("", log.New(os.Stdout, "modltest: ", log.Lmicroseconds))
	dbmap.AddTable(T{}).SetKeys(true, "ID")

	err := dbmap.CreateTables()
	if err != nil {
		panic(err)
	}
	defer dbmap.Cleanup()

	// Insert with a zero pkey.
	t0 := &T{ID: 0}
	_insert(dbmap, t0)
	t1 := T{}
	MustGet(dbmap, &t1, t0.ID)
	if t1.ID != t0.ID {
		t.Errorf("got t1.ID == %d, want %d", t1.ID, t0.ID)
	}

	// Insert with an overridden nonzero pkey.
	t2 := &T{ID: 123}
	_insert(dbmap, t2)
	if t2.ID != 123 {
		t.Errorf("got t2.ID == %d, want %d", t2.ID, 123)
	}
	t3 := T{}
	MustGet(dbmap, &t3, 123)
	if t3.ID != 123 {
		t.Errorf("got t3.ID == %d, want %d", t3.ID, 123)
	}

	// Insert with a zero pkey.
	t4 := &T{ID: 0}
	_insert(dbmap, t4)
	t5 := T{}
	MustGet(dbmap, &t5, t4.ID)
	if t5.ID != t4.ID {
		t.Errorf("got t5.ID == %d, want %d", t5.ID, t4.ID)
	}
}

func TestRawSelect(t *testing.T) {
	dbmap := initDbMap()
	defer dbmap.Cleanup()

	p1 := &Person{0, 0, 0, "bob", "smith", 0}
	_insert(dbmap, p1)

	inv1 := &Invoice{0, 0, 0, "xmas order", p1.ID, true}
	_insert(dbmap, inv1)

	expected := &InvoicePersonView{inv1.ID, p1.ID, inv1.Memo, p1.FName, 0}

	query := "select i.id invoiceid, p.id personid, i.memo, p.fname " +
		"from invoice_test i, person_test p " +
		"where i.personid = p.id"
	list := []InvoicePersonView{}
	MustSelect(dbmap, &list, query)
	if len(list) != 1 {
		t.Errorf("len(list) != 1: %d", len(list))
	} else if !reflect.DeepEqual(expected, &list[0]) {
		t.Errorf("%v != %v", expected, list[0])
	}
}

func TestHooks(t *testing.T) {
	dbmap := initDbMap()
	defer dbmap.Cleanup()

	p1 := &Person{0, 0, 0, "bob", "smith", 0}
	_insert(dbmap, p1)
	if p1.Created == 0 || p1.Updated == 0 {
		t.Errorf("p1.PreInsert() didn't run: %v", p1)
	} else if p1.LName != "postinsert" {
		t.Errorf("p1.PostInsert() didn't run: %v", p1)
	}

	MustGet(dbmap, p1, p1.ID)
	if p1.LName != "postget" {
		t.Errorf("p1.PostGet() didn't run: %v", p1)
	}

	p1.LName = "smith"
	_update(dbmap, p1)
	if p1.FName != "preupdate" {
		t.Errorf("p1.PreUpdate() didn't run: %v", p1)
	} else if p1.LName != "postupdate" {
		t.Errorf("p1.PostUpdate() didn't run: %v", p1)
	}

	var persons []*Person
	bindVar := dbmap.Dialect.BindVar(0)
	MustSelect(dbmap, &persons, "select * from person_test where id = "+bindVar, p1.ID)
	if persons[0].LName != "postget" {
		t.Errorf("p1.PostGet() didn't run after select: %v", p1)
	}

	_del(dbmap, p1)
	if p1.FName != "predelete" {
		t.Errorf("p1.PreDelete() didn't run: %v", p1)
	} else if p1.LName != "postdelete" {
		t.Errorf("p1.PostDelete() didn't run: %v", p1)
	}

	// Test error case
	p2 := &Person{0, 0, 0, "badname", "", 0}
	err := dbmap.Insert(p2)
	if err == nil {
		t.Errorf("p2.PreInsert() didn't return an error")
	}
}

func TestTransaction(t *testing.T) {
	dbmap := initDbMap()
	defer dbmap.Cleanup()

	inv1 := &Invoice{0, 100, 200, "t1", 0, true}
	inv2 := &Invoice{0, 100, 200, "t2", 0, false}

	trans, err := dbmap.Begin()
	if err != nil {
		panic(err)
	}
	trans.Insert(inv1, inv2)
	err = trans.Commit()
	if err != nil {
		panic(err)
	}

	obj := &Invoice{}
	err = dbmap.Get(obj, inv1.ID)
	if err != nil {
		panic(err)
	}
	if !reflect.DeepEqual(inv1, obj) {
		t.Errorf("%v != %v", inv1, obj)
	}
	err = dbmap.Get(obj, inv2.ID)
	if err != nil {
		panic(err)
	}
	if !reflect.DeepEqual(inv2, obj) {
		t.Errorf("%v != %v", inv2, obj)
	}
}

func TestMultiple(t *testing.T) {
	dbmap := initDbMap()
	defer dbmap.Cleanup()

	inv1 := &Invoice{0, 100, 200, "a", 0, false}
	inv2 := &Invoice{0, 100, 200, "b", 0, true}
	_insert(dbmap, inv1, inv2)

	inv1.Memo = "c"
	inv2.Memo = "d"
	_update(dbmap, inv1, inv2)

	count := _del(dbmap, inv1, inv2)
	if count != 2 {
		t.Errorf("%d != 2", count)
	}
}

func TestCrud(t *testing.T) {
	dbmap := initDbMap()
	defer dbmap.Cleanup()

	inv := &Invoice{0, 100, 200, "first order", 0, true}

	// INSERT row
	_insert(dbmap, inv)
	if inv.ID == 0 {
		t.Errorf("inv.ID was not set on INSERT")
		return
	}

	// SELECT row
	inv2 := &Invoice{}
	MustGet(dbmap, inv2, inv.ID)
	if !reflect.DeepEqual(inv, inv2) {
		t.Errorf("%v != %v", inv, inv2)
	}

	// UPDATE row and SELECT
	inv.Memo = "second order"
	inv.Created = 999
	inv.Updated = 11111
	count := _update(dbmap, inv)
	if count != 1 {
		t.Errorf("update 1 != %d", count)
	}

	MustGet(dbmap, inv2, inv.ID)
	if !reflect.DeepEqual(inv, inv2) {
		t.Errorf("%v != %v", inv, inv2)
	}

	// DELETE row
	deleted := _del(dbmap, inv)
	if deleted != 1 {
		t.Errorf("Did not delete row with ID: %d", inv.ID)
		return
	}

	// VERIFY deleted
	err := dbmap.Get(inv2, inv.ID)
	if err != sql.ErrNoRows {
		t.Errorf("Found invoice with id: %d after Delete()", inv.ID)
	}
}

func TestWithIgnoredColumn(t *testing.T) {
	dbmap := initDbMap()
	defer dbmap.Cleanup()

	ic := &WithIgnoredColumn{-1, 0, 1}
	_insert(dbmap, ic)
	expected := &WithIgnoredColumn{0, 1, 1}

	ic2 := &WithIgnoredColumn{}
	MustGet(dbmap, ic2, ic.ID)

	if !reflect.DeepEqual(expected, ic2) {
		t.Errorf("%v != %v", expected, ic2)
	}

	if _del(dbmap, ic) != 1 {
		t.Errorf("Did not delete row with ID: %d", ic.ID)
		return
	}

	err := dbmap.Get(ic2, ic.ID)
	if err != sql.ErrNoRows {
		t.Errorf("Found id: %d after Delete() (%#v)", ic.ID, ic2)
	}
}

func TestVersionMultipleRows(t *testing.T) {
	dbmap := initDbMap()
	defer dbmap.Cleanup()

	persons := []*Person{
		&Person{0, 0, 0, "Bob", "Smith", 0},
		&Person{0, 0, 0, "Jane", "Smith", 0},
		&Person{0, 0, 0, "Mike", "Smith", 0},
	}

	_insert(dbmap, persons[0], persons[1], persons[2])

	for x, p := range persons {
		if p.Version != 1 {
			t.Errorf("person[%d].Version != 1: %d", x, p.Version)
		}
	}
}

func TestWithStringPk(t *testing.T) {
	dbmap := newDbMap()
	//dbmap.TraceOn("", log.New(os.Stdout, "modltest: ", log.Lmicroseconds))
	dbmap.AddTableWithName(WithStringPk{}, "string_pk_test").SetKeys(true, "ID")
	_, err := dbmap.Exec("create table string_pk_test (ID varchar(255), Name varchar(255));")
	if err != nil {
		t.Errorf("couldn't create string_pk_test: %v", err)
	}
	defer dbmap.Cleanup()

	row := &WithStringPk{"1", "foo"}
	if err := dbmap.Insert(row); err != nil {
		t.Error(err)
	}
}

func TestWithEmbeddedStruct(t *testing.T) {
	dbmap := newDbMap()
	//dbmap.TraceOn("", log.New(os.Stdout, "modltest: ", log.Lmicroseconds))
	dbmap.AddTableWithName(WithEmbeddedStruct{}, "embedded_struct_test").SetKeys(true, "ID")
	err := dbmap.CreateTables()
	if err != nil {
		t.Errorf("couldn't create embedded_struct_test: %v", err)
	}
	defer dbmap.DropTables()

	row := &WithEmbeddedStruct{Names: Names{"Alice", "Smith"}}
	err = dbmap.Insert(row)
	if err != nil {
		t.Errorf("Error inserting into table w/embedded struct: %v", err)
	}

	var es WithEmbeddedStruct
	err = dbmap.Get(&es, row.Id)
	if err != nil {
		t.Errorf("Error selecting from table w/embedded struct: %v", err)
	}
}

func TestWithEmbeddedStructAutoIncrColNotFirst(t *testing.T) {
	// Tests that the tablemap retains separate indices for SQL
	// columns (which are flattened with respect to struct embedding)
	// and Go fields (which are not). In this test case, the
	// auto-incremented column is the 3rd column in SQL but the 2nd Go
	// field.

	type Embedded struct{ A, B string }
	type withAutoIncrColNotFirst struct {
		Embedded
		ID int
	}

	dbmap := newDbMap()
	//dbmap.TraceOn("", log.New(os.Stdout, "modltest: ", log.Lmicroseconds))
	dbmap.AddTableWithName(withAutoIncrColNotFirst{}, "auto_incr_col_not_first_test").SetKeys(true, "ID")
	if err := dbmap.CreateTables(); err != nil {
		t.Errorf("couldn't create auto_incr_col_not_first_test: %v", err)
	}
	defer dbmap.Cleanup()

	row := withAutoIncrColNotFirst{Embedded: Embedded{A: "a"}, ID: 0}
	if err := dbmap.Insert(&row); err != nil {
		t.Fatal(err)
	}

	var got withAutoIncrColNotFirst
	if err := dbmap.Get(&got, row.ID); err != nil {
		t.Fatal(err)
	}
	if got != row {
		t.Errorf("Got %+v, want %+v", got, row)
	}
}

func TestWithEmbeddedAutoIncrCol(t *testing.T) {
	type EmbeddedID struct {
		A  string
		ID int
	}
	type embeddedAutoIncrCol struct{ EmbeddedID }

	dbmap := newDbMap()
	//dbmap.TraceOn("", log.New(os.Stdout, "modltest: ", log.Lmicroseconds))
	dbmap.AddTableWithName(embeddedAutoIncrCol{}, "embedded_auto_incr_col_test").SetKeys(true, "ID")
	if err := dbmap.CreateTables(); err != nil {
		t.Errorf("couldn't create embedded_auto_incr_col_test: %v", err)
	}
	defer dbmap.Cleanup()

	row := embeddedAutoIncrCol{EmbeddedID{A: "a", ID: 0}}
	if err := dbmap.Insert(&row); err != nil {
		t.Fatal(err)
	}

	var got embeddedAutoIncrCol
	if err := dbmap.Get(&got, row.ID); err != nil {
		t.Fatal(err)
	}
	if got != row {
		t.Errorf("Got %+v, want %+v", got, row)
	}
}

func BenchmarkNativeCrud(b *testing.B) {
	var err error

	b.StopTimer()
	dbmap := initDbMapBench()
	defer dbmap.Cleanup()
	b.StartTimer()

	insert := "insert into invoice_test (date_created, updated, memo, personid) values (?, ?, ?, ?)"
	sel := "select id, date_created, updated, memo, personid from invoice_test where id=?"
	update := "update invoice_test set date_created=?, updated=?, memo=?, personid=? where id=?"
	delete := "delete from invoice_test where id=?"

	suffix := dbmap.Dialect.AutoIncrInsertSuffix(&ColumnMap{ColumnName: "id"})

	insert = ReBind(insert, dbmap.Dialect) + suffix
	sel = ReBind(sel, dbmap.Dialect)
	update = ReBind(update, dbmap.Dialect)
	delete = ReBind(delete, dbmap.Dialect)

	inv := &Invoice{0, 100, 200, "my memo", 0, false}

	for i := 0; i < b.N; i++ {
		if len(suffix) == 0 {
			res, err := dbmap.Db.Exec(insert, inv.Created, inv.Updated, inv.Memo, inv.PersonID)
			if err != nil {
				panic(err)
			}

			newid, err := res.LastInsertId()
			if err != nil {
				panic(err)
			}
			inv.ID = newid
		} else {
			rows, err := dbmap.Db.Query(insert, inv.Created, inv.Updated, inv.Memo, inv.PersonID)
			if err != nil {
				panic(err)
			}

			if rows.Next() {
				err = rows.Scan(&inv.ID)
				if err != nil {
					panic(err)
				}
			}
			rows.Close()

		}

		row := dbmap.Db.QueryRow(sel, inv.ID)
		err = row.Scan(&inv.ID, &inv.Created, &inv.Updated, &inv.Memo, &inv.PersonID)
		if err != nil {
			panic(err)
		}

		inv.Created = 1000
		inv.Updated = 2000
		inv.Memo = "my memo 2"
		inv.PersonID = 3000

		_, err = dbmap.Db.Exec(update, inv.Created, inv.Updated, inv.Memo,
			inv.PersonID, inv.ID)
		if err != nil {
			panic(err)
		}

		_, err = dbmap.Db.Exec(delete, inv.ID)
		if err != nil {
			panic(err)
		}
	}

}

func BenchmarkModlCrud(b *testing.B) {
	b.StopTimer()
	dbmap := initDbMapBench()
	defer dbmap.Cleanup()
	//dbmap.TraceOn("", log.New(os.Stdout, "modltest: ", log.Lmicroseconds))
	b.StartTimer()

	inv := &Invoice{0, 100, 200, "my memo", 0, true}
	for i := 0; i < b.N; i++ {
		err := dbmap.Insert(inv)
		if err != nil {
			panic(err)
		}

		inv2 := Invoice{}
		err = dbmap.Get(&inv2, inv.ID)
		if err != nil {
			panic(err)
		}

		inv2.Created = 1000
		inv2.Updated = 2000
		inv2.Memo = "my memo 2"
		inv2.PersonID = 3000
		_, err = dbmap.Update(&inv2)
		if err != nil {
			panic(err)
		}

		_, err = dbmap.Delete(&inv2)
		if err != nil {
			panic(err)
		}

	}
}

func initDbMapBench() *DbMap {
	dbmap := newDbMap()
	dbmap.Db.Exec("drop table if exists invoice_test")
	dbmap.AddTableWithName(Invoice{}, "invoice_test").SetKeys(true, "id")
	err := dbmap.CreateTables()
	if err != nil {
		panic(err)
	}
	return dbmap
}

func (d *DbMap) Cleanup() {
	err := d.DropTables()
	if err != nil {
		panic(err)
	}
	err = d.Dbx.Close()
	if err != nil {
		panic(err)
	}
}

func initDbMap() *DbMap {
	dbmap := newDbMap()
	//dbmap.TraceOn("", log.New(os.Stdout, "modltest: ", log.Lmicroseconds))
	dbmap.AddTableWithName(Invoice{}, "invoice_test").SetKeys(true, "id")
	dbmap.AddTableWithName(Person{}, "person_test").SetKeys(true, "id")
	dbmap.AddTableWithName(WithIgnoredColumn{}, "ignored_column_test").SetKeys(true, "id")
	dbmap.AddTableWithName(WithTime{}, "time_test").SetKeys(true, "ID")
	err := dbmap.CreateTables()
	if err != nil {
		panic(err)
	}

	return dbmap
}

func TestTruncateTables(t *testing.T) {
	dbmap := initDbMap()
	defer dbmap.Cleanup()
	err := dbmap.CreateTablesIfNotExists()
	if err != nil {
		t.Error(err)
	}

	// Insert some data
	p1 := &Person{0, 0, 0, "Bob", "Smith", 0}
	dbmap.Insert(p1)
	inv := &Invoice{0, 0, 1, "my invoice", 0, true}
	dbmap.Insert(inv)

	err = dbmap.TruncateTables()
	if err != nil {
		t.Error(err)
	}

	// Make sure all rows are deleted
	people := []Person{}
	invoices := []Invoice{}
	dbmap.Select(&people, "SELECT * FROM person_test")
	if len(people) != 0 {
		t.Errorf("Expected 0 person rows, got %d", len(people))
	}
	dbmap.Select(&invoices, "SELECT * FROM invoice_test")
	if len(invoices) != 0 {
		t.Errorf("Expected 0 invoice rows, got %d", len(invoices))
	}
}

func TestTruncateTablesIdentityRestart(t *testing.T) {
	dbmap := initDbMap()
	defer dbmap.Cleanup()
	err := dbmap.CreateTablesIfNotExists()
	if err != nil {
		t.Error(err)
	}

	// Insert some data
	p1 := &Person{0, 0, 0, "Bob", "Smith", 0}
	dbmap.Insert(p1)
	inv := &Invoice{0, 0, 1, "my invoice", 0, true}
	dbmap.Insert(inv)

	err = dbmap.TruncateTablesIdentityRestart()
	if err != nil {
		t.Error(err)
	}

	// Make sure all rows are deleted
	people := []Person{}
	invoices := []Invoice{}
	dbmap.Select(&people, "SELECT * FROM person_test")
	if len(people) != 0 {
		t.Errorf("Expected 0 person rows, got %d", len(people))
	}
	dbmap.Select(&invoices, "SELECT * FROM invoice_test")
	if len(invoices) != 0 {
		t.Errorf("Expected 0 invoice rows, got %d", len(invoices))
	}

	p2 := &Person{0, 0, 0, "Other", "Person", 0}
	dbmap.Insert(p2)
	if p2.ID != int64(1) {
		t.Errorf("Expected new person ID to be equal to 1, was %d", p2.ID)
	}
}

func TestSelectBehavior(t *testing.T) {
	db := initDbMap()
	defer db.Cleanup()

	p := Person{}

	// check that SelectOne with no rows returns ErrNoRows
	err := db.SelectOne(&p, "select * from person_test")
	if err == nil || err != sql.ErrNoRows {
		t.Fatal(err)
	}

	// insert and ensure SelectOne works properly
	bob := Person{0, 0, 0, "Bob", "Smith", 0}
	db.Insert(&bob)

	err = db.SelectOne(&p, "select * from person_test")
	if err != nil {
		t.Fatal(err)
	}
	if p.FName != "Bob" {
		t.Errorf("Wrong FName: %s", p.FName)
	}
	// there's a post hook on this that sets it to postget, ensure it ran
	if p.LName != "postget" {
		t.Errorf("Wrong LName: %s", p.LName)
	}

	// insert again and ensure SelectOne *does not* error in rows > 1
	ben := Person{0, 0, 0, "Ben", "Smith", 0}
	db.Insert(&ben)

	err = db.SelectOne(&p, "select * from person_test ORDER BY fname ASC")
	if err != nil {
		t.Fatal(err)
	}
	if p.FName != "Ben" {
		t.Errorf("Wrong FName: %s", p.FName)
	}
}

func TestQuoteTableNames(t *testing.T) {
	dbmap := initDbMap()
	defer dbmap.Cleanup()

	quotedTableName := dbmap.Dialect.QuoteField("person_test")

	// Use a buffer to hold the log to check generated queries
	var logBuffer bytes.Buffer
	dbmap.TraceOn("", log.New(&logBuffer, "modltest:", log.Lmicroseconds))

	// Create some rows
	p1 := &Person{0, 0, 0, "bob", "smith", 0}
	errorTemplate := "Expected quoted table name %v in query but didn't find it"

	// Check if Insert quotes the table name
	id := dbmap.Insert(p1)
	if !bytes.Contains(logBuffer.Bytes(), []byte(quotedTableName)) {
		t.Log("log:", logBuffer.String())
		t.Errorf(errorTemplate, quotedTableName)
	}
	logBuffer.Reset()

	// Check if Get quotes the table name
	dbmap.Get(Person{}, id)
	if !bytes.Contains(logBuffer.Bytes(), []byte(quotedTableName)) {
		t.Errorf(errorTemplate, quotedTableName)
	}
	logBuffer.Reset()
}

type WithTime struct {
	ID   int64
	Time time.Time
}

func TestWithTime(t *testing.T) {
	dbmap := initDbMap()
	defer dbmap.Cleanup()

	// FIXME: there seems to be a bug with go-sql-driver and timezones?
	// MySQL doesn't have any timestamp support, but since it is not
	// sending any, the scan assumes UTC, so the scanner should
	// probably convert to UTC before storing.  Also, note that time.Time
	// support requires a special bit to be added to the DSN
	t1, err := time.Parse("2006-01-02 15:04:05 -0700 MST",
		"2013-08-09 21:30:43 +0000 UTC")
	if err != nil {
		t.Fatal(err)
	}

	w1 := WithTime{1, t1}
	dbmap.Insert(&w1)

	w2 := WithTime{}
	dbmap.Get(&w2, w1.ID)

	if w1.Time.UnixNano() != w2.Time.UnixNano() {
		t.Errorf("%v != %v", w1, w2)
	}
}

func initDbMapNulls() *DbMap {
	dbmap := newDbMap()
	//dbmap.TraceOn("", log.New(os.Stdout, "modltest: ", log.Lmicroseconds))
	dbmap.AddTable(TableWithNull{}).SetKeys(false, "id")
	err := dbmap.CreateTables()
	if err != nil {
		panic(err)
	}
	return dbmap
}

func newDbMap() *DbMap {
	dialect, driver := dialectAndDriver()
	return NewDbMap(connect(driver), dialect)
}

func connect(driver string) *sql.DB {
	dsn := os.Getenv("MODL_TEST_DSN")
	if dsn == "" {
		panic("MODL_TEST_DSN env variable is not set. Please see README.md")
	}

	db, err := sql.Open(driver, dsn)
	if err != nil {
		panic("Error connecting to db: " + err.Error())
	}
	err = db.Ping()
	if err != nil {
		panic("Error connecting to db: " + err.Error())
	}
	return db
}

func dialectAndDriver() (Dialect, string) {
	switch os.Getenv("MODL_TEST_DIALECT") {
	case "mysql":
		return MySQLDialect{"InnoDB", "UTF8"}, "mysql"
	case "postgres":
		return PostgresDialect{}, "postgres"
	case "sqlite":
		return SqliteDialect{}, "sqlite3"
	}
	panic("MODL_TEST_DIALECT env variable is not set or is invalid. Please see README.md")
}

func _insert(dbmap *DbMap, list ...interface{}) {
	err := dbmap.Insert(list...)
	if err != nil {
		panic(err)
	}
}

func _update(dbmap *DbMap, list ...interface{}) int64 {
	count, err := dbmap.Update(list...)
	if err != nil {
		panic(err)
	}
	return count
}

func _del(dbmap *DbMap, list ...interface{}) int64 {
	count, err := dbmap.Delete(list...)
	if err != nil {
		panic(err)
	}

	return count
}

func MustGet(dbmap *DbMap, i interface{}, keys ...interface{}) {
	err := dbmap.Get(i, keys...)
	if err != nil {
		panic(err)
	}
}

func MustSelect(dbmap *DbMap, dest interface{}, query string, args ...interface{}) {
	err := dbmap.Select(dest, query, args...)
	if err != nil {
		panic(err)
	}
}
