package lsifstore

// func TestDatabaseSymbols(t *testing.T) {
// 	if testing.Short() {
// 		t.Skip()
// 	}
// 	dbtesting.SetupGlobalTestDB(t)
// 	populateTestStore(t)
// 	store := NewStore(dbconn.Global, &observation.TestContext)

// 	actualList, totalCount, err := store.Symbols(context.Background(), testBundleID, nil, 0, 1000)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	// TODO(sqs): test totalCount
// 	_ = totalCount

// 	// Filter down the actual list to a single symbol that we test against.
// 	const testMonikerIdentifier = "github.com/sourcegraph/lsif-go/protocol:ToolInfo"

// 	actual := findSymbolsMatching(actualList, func(symbol *Symbol) bool {
// 		for _, m := range symbol.Monikers {
// 			if m.Identifier == testMonikerIdentifier {
// 				return true
// 			}
// 		}
// 		return false
// 	})

// 	expected := []*Symbol{
// 		{
// 			DumpID: testBundleID,
// 			SymbolData: protocol.SymbolData{
// 				Text: "ToolInfo",
// 				Kind: 11,
// 			},
// 			Locations: []protocol.SymbolLocation{
// 				{
// 					URI: "protocol/protocol.go",
// 					Range: &protocol.RangeData{
// 						Start: protocol.Pos{Line: 66, Character: 5},
// 						End:   protocol.Pos{Line: 66, Character: 13},
// 					},
// 					FullRange: protocol.RangeData{
// 						Start: protocol.Pos{Line: 66, Character: 0},
// 						End:   protocol.Pos{Line: 73, Character: 1},
// 					},
// 				},
// 			},
// 			Monikers: []MonikerData{
// 				{
// 					Kind:       "export",
// 					Scheme:     "gomod",
// 					Identifier: testMonikerIdentifier,
// 				},
// 			},
// 		},
// 	}
// 	if diff := cmp.Diff(expected, actual); diff != "" {
// 		t.Errorf("unexpected symbols (-want +got):\n%s", diff)
// 	}
// }

// TODO(sqs): steps for regenerating lsif-go@ad3507cb.sql
//
// NOTE: THIS WILL DELETE ALL LSIF DATA FROM YOUR LOCAL DATABASE!!!
//
// # FIRST: make sure the github.com/sourcegraph/lsif-go repo exists on your sourcegraph instance
//
// pg_dump --data-only --format plain --inserts --table 'lsif_data_*' --file lsif_backup.sql
//
// for table in $(psql -XAt -c "select tablename from pg_tables where schemaname='public' and tablename like 'lsif_data_%';"); do psql -c "truncate table $table"; done
//
// pg_dump --data-only --format plain --inserts --table 'lsif_data_*' | grep -v '^--' | grep -v '^SET' | grep -v '^SELECT' | sed 's/VALUES ([[:digit:]]\+,/VALUES (447,/g' > tmp.sql
// mv tmp.sql enterprise/internal/codeintel/stores/lsifstore/testdata/lsif-go@ad3507cb.sql
//
// # RESTORE
// # rerun the `for table` thing above, then:
// psql -X < lsif_backup.sql

// func TestBuildSymbolTree(t *testing.T) {
// 	symbolData := func(id uint64, children ...uint64) SymbolData {
// 		return SymbolData{
// 			ID:         id,
// 			SymbolData: protocol.SymbolData{Text: fmt.Sprint(id)},
// 			Children:   children,
// 		}
// 	}
// 	tests := []struct {
// 		datas []SymbolData
// 		want  []Symbol
// 	}{
// 		{
// 			datas: []SymbolData{symbolData(1, 2), symbolData(2)},
// 			want: []Symbol{
// 				{
// 					SymbolData: protocol.SymbolData{Text: "1"},
// 					Children: []Symbol{
// 						{SymbolData: protocol.SymbolData{Text: "2"}},
// 					},
// 				},
// 			},
// 		},
// 		{
// 			datas: []SymbolData{
// 				symbolData(10, 20, 30),
// 				symbolData(20, 21, 22),
// 				symbolData(21, 23),
// 				symbolData(22),
// 				symbolData(23),
// 				symbolData(30, 31),
// 				symbolData(31),
// 			},
// 			want: []Symbol{
// 				{
// 					SymbolData: protocol.SymbolData{Text: "10"},
// 					Children: []Symbol{
// 						{
// 							SymbolData: protocol.SymbolData{Text: "20"},
// 							Children: []Symbol{
// 								{
// 									SymbolData: protocol.SymbolData{Text: "21"},
// 									Children: []Symbol{
// 										{SymbolData: protocol.SymbolData{Text: "23"}},
// 									},
// 								},
// 								{
// 									SymbolData: protocol.SymbolData{Text: "22"},
// 								},
// 							},
// 						},
// 						{
// 							SymbolData: protocol.SymbolData{Text: "30"},
// 							Children: []Symbol{
// 								{SymbolData: protocol.SymbolData{Text: "31"}},
// 							},
// 						},
// 					},
// 				},
// 			},
// 		},
// 	}

// 	for i, test := range tests {
// 		t.Run(strconv.Itoa(i), func(t *testing.T) {
// 			got := buildSymbolTree(test.datas, 0)
// 			if diff := cmp.Diff(test.want, got); diff != "" {
// 				t.Errorf("unexpected tree (-want +got):\n%s", diff)
// 			}
// 		})
// 	}
// }

// func TestFindPathToSymbolInTree(t *testing.T) {
// 	matchFn := func(symbol *Symbol) bool { return symbol.Text == "*" }
// 	tests := []struct {
// 		root   Symbol
// 		want   []int
// 		wantOK bool
// 	}{
// 		{
// 			root:   Symbol{Children: []Symbol{{}, {}}},
// 			wantOK: false,
// 		},
// 		{
// 			root:   Symbol{SymbolData: protocol.SymbolData{Text: "*"}},
// 			want:   nil,
// 			wantOK: true,
// 		},
// 		{
// 			root: Symbol{
// 				Children: []Symbol{
// 					{},
// 					{
// 						Children: []Symbol{
// 							{}, {},
// 							{Children: []Symbol{{}, {}, {}}},
// 							{
// 								Children: []Symbol{
// 									{}, {},
// 									{SymbolData: protocol.SymbolData{Text: "*"}},
// 									{},
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 			want:   []int{1, 3, 2},
// 			wantOK: true,
// 		},
// 	}

// 	for i, test := range tests {
// 		t.Run(strconv.Itoa(i), func(t *testing.T) {
// 			got, ok := findPathToSymbolInTree(&test.root, matchFn)
// 			if ok != test.wantOK {
// 				t.Errorf("got ok %v, want %v", ok, test.wantOK)
// 			}
// 			if diff := cmp.Diff(test.want, got); diff != "" {
// 				t.Errorf("unexpected tree (-want +got):\n%s", diff)
// 			}
// 		})
// 	}
// }

// func TestTrimSymbolTree(t *testing.T) {
// 	tree := []Symbol{
// 		{
// 			SymbolData: protocol.SymbolData{Text: "0"},
// 			Children: []Symbol{
// 				{SymbolData: protocol.SymbolData{Text: "0a"}},
// 				{
// 					SymbolData: protocol.SymbolData{Text: "0b"},
// 					Children: []Symbol{
// 						{SymbolData: protocol.SymbolData{Text: "0b0"}},
// 						{SymbolData: protocol.SymbolData{Text: "0b1"}},
// 					},
// 				},
// 			},
// 		},
// 	}

// 	trimSymbolTree(&tree, func(symbol *Symbol) bool {
// 		return symbol.Text == "0" || symbol.Text == "0b" || symbol.Text == "0b1"
// 	})

// 	want := []Symbol{
// 		{
// 			SymbolData: protocol.SymbolData{Text: "0"},
// 			Children: []Symbol{
// 				{
// 					SymbolData: protocol.SymbolData{Text: "0b"},
// 					Children: []Symbol{
// 						{SymbolData: protocol.SymbolData{Text: "0b1"}},
// 					},
// 				},
// 			},
// 		},
// 	}

// 	if diff := cmp.Diff(want, tree); diff != "" {
// 		t.Errorf("unexpected tree (-want +got):\n%s", diff)
// 	}
// }

// func findSymbolsMatching(roots []Symbol, match func(symbol *Symbol) bool) (matches []*Symbol) {
// 	for i := range roots {
// 		WalkSymbolTree(&roots[i], func(symbol *Symbol) {
// 			if match(symbol) {
// 				matches = append(matches, symbol)
// 			}
// 		})
// 	}
// 	return matches
// }
