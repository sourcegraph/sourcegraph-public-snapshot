package basestore

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	orderedmap "github.com/wk8/go-ordered-map/v2"

	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

type reducee struct {
	ID     int
	Values []int
}

type reduceeRow struct {
	ID, Value int
}

type testReducer struct{}

func (d testReducer) Create() reducee {
	return reducee{Values: make([]int, 0)}
}

func (d testReducer) Reduce(collection reducee, value reduceeRow) reducee {
	collection.ID = value.ID
	collection.Values = append(collection.Values, value.Value)
	return collection
}

func Test_KeyedCollectionScannerOrdered(t *testing.T) {
	data := []reduceeRow{
		{
			ID:    0,
			Value: 0,
		},
		{
			ID:    0,
			Value: 1,
		},
		{
			ID:    0,
			Value: 2,
		},
		{
			ID:    2,
			Value: 0,
		},
		{
			ID:    1,
			Value: 1,
		},
		{
			ID:    0,
			Value: 3,
		},
		{
			ID:    -1,
			Value: -1,
		},
		{
			ID:    1,
			Value: 0,
		},
	}
	offset := -1

	rows := NewMockRows()
	rows.NextFunc.SetDefaultHook(func() bool {
		offset++
		return offset < len(data)
	})
	rows.ScanFunc.SetDefaultHook(func(i ...interface{}) error {
		*(i[0].(*int)) = data[offset].ID
		*(i[1].(*int)) = data[offset].Value
		return nil
	})

	m := &OrderedMap[int, reducee]{m: orderedmap.New[int, reducee]()}
	NewKeyedCollectionScanner[int, reduceeRow, reducee](m, func(s dbutil.Scanner) (int, reduceeRow, error) {
		var red reduceeRow
		err := s.Scan(&red.ID, &red.Value)
		return red.ID, red, err
	}, testReducer{})(rows, nil)

	if m.Len() != 4 {
		t.Errorf("unexpected map size: want=%d got=%d\n%v", 4, m.Len(), m.Values())
	}

	if diff := cmp.Diff([]reducee{
		{
			ID:     0,
			Values: []int{0, 1, 2, 3},
		},
		{
			ID:     2,
			Values: []int{0},
		},
		{
			ID:     1,
			Values: []int{1, 0},
		},
		{
			ID:     -1,
			Values: []int{-1},
		},
	}, m.Values()); diff != "" {
		t.Errorf("unexpected collection output (-want,+got):\n%s", diff)
	}
}
