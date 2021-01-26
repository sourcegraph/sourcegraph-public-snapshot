package lsifstore

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestBuildSymbolTree(t *testing.T) {
	symbolData := func(id uint64, children ...uint64) SymbolData {
		return SymbolData{
			ID:       id,
			Text:     fmt.Sprint(id),
			Children: children,
		}
	}
	tests := []struct {
		datas []SymbolData
		want  []Symbol
	}{
		{
			datas: []SymbolData{symbolData(1, 2), symbolData(2)},
			want: []Symbol{
				{
					Text: "1",
					Children: []Symbol{
						{Text: "2"},
					},
				},
			},
		},
		{
			datas: []SymbolData{
				symbolData(10, 20, 30),
				symbolData(20, 21, 22),
				symbolData(21, 23),
				symbolData(22),
				symbolData(23),
				symbolData(30, 31),
				symbolData(31),
			},
			want: []Symbol{
				{
					Text: "10",
					Children: []Symbol{
						{
							Text: "20",
							Children: []Symbol{
								{
									Text: "21",
									Children: []Symbol{
										{Text: "23"},
									},
								},
								{
									Text: "22",
								},
							},
						},
						{
							Text: "30",
							Children: []Symbol{
								{Text: "31"},
							},
						},
					},
				},
			},
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got := buildSymbolTree(test.datas, 0)
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("unexpected tree (-want +got):\n%s", diff)
			}
		})
	}
}

func TestFindPathToSymbolInTree(t *testing.T) {
	matchFn := func(symbol *Symbol) bool { return symbol.Text == "*" }
	tests := []struct {
		root   Symbol
		want   []int
		wantOK bool
	}{
		{
			root:   Symbol{Children: []Symbol{{}, {}}},
			wantOK: false,
		},
		{
			root:   Symbol{Text: "*"},
			want:   nil,
			wantOK: true,
		},
		{
			root: Symbol{
				Children: []Symbol{
					{},
					{
						Children: []Symbol{
							{}, {},
							{Children: []Symbol{{}, {}, {}}},
							{
								Children: []Symbol{
									{}, {},
									{Text: "*"},
									{},
								},
							},
						},
					},
				},
			},
			want:   []int{1, 3, 2},
			wantOK: true,
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got, ok := findPathToSymbolInTree(&test.root, matchFn)
			if ok != test.wantOK {
				t.Errorf("got ok %v, want %v", ok, test.wantOK)
			}
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("unexpected tree (-want +got):\n%s", diff)
			}
		})
	}
}

func TestTrimSymbolTree(t *testing.T) {
	tree := []Symbol{
		{
			Text: "0",
			Children: []Symbol{
				{Text: "0a"},
				{
					Text: "0b",
					Children: []Symbol{
						{Text: "0b0"},
						{Text: "0b1"},
					},
				},
			},
		},
	}

	trimSymbolTree(&tree, func(symbol *Symbol) bool {
		return symbol.Text == "0" || symbol.Text == "0b" || symbol.Text == "0b1"
	})

	want := []Symbol{
		{
			Text: "0",
			Children: []Symbol{
				{
					Text: "0b",
					Children: []Symbol{
						{Text: "0b1"},
					},
				},
			},
		},
	}

	if diff := cmp.Diff(want, tree); diff != "" {
		t.Errorf("unexpected tree (-want +got):\n%s", diff)
	}
}

func findSymbolsMatching(roots []Symbol, match func(symbol *Symbol) bool) (matches []*Symbol) {
	for i := range roots {
		WalkSymbolTree(&roots[i], func(symbol *Symbol) {
			if match(symbol) {
				matches = append(matches, symbol)
			}
		})
	}
	return matches
}
