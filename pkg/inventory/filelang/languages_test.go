package filelang

import (
	"reflect"
	"testing"
)

func TestLanguages_Generated(t *testing.T) {
	langs := Langs.ByFilename("file.c")
	if want := []string{"C"}; !reflect.DeepEqual(langNames(langs), want) {
		t.Errorf("got %v, want %v", langNames(langs), want)
	}
}

func TestLanguages_ByFilename(t *testing.T) {
	tests := []struct {
		name          string
		langs         Languages
		wantLangNames []string
	}{
		// No matches
		{name: "file.a"},
		{name: ""},
		{name: "."},
		{name: "\x00"},

		// Extension matches
		{
			name:          "file.a",
			langs:         Languages{{Name: "A", Extensions: []string{".a"}}},
			wantLangNames: []string{"A"},
		},
		{
			name:          ".a",
			langs:         Languages{{Name: "A", Extensions: []string{".a"}}},
			wantLangNames: []string{"A"},
		},
		{
			name:          "file.A",
			langs:         Languages{{Name: "A", Extensions: []string{".a"}}},
			wantLangNames: []string{"A"},
		},

		// Filename matches
		{
			name:          "F",
			langs:         Languages{{Name: "A", Filenames: []string{"F"}}},
			wantLangNames: []string{"A"},
		},
		{
			name:          "F",
			langs:         Languages{{Name: "A", Filenames: []string{"f"}}},
			wantLangNames: nil,
		},

		// Sort by primary match then language name
		{
			name: "file.b",
			langs: Languages{
				{Name: "A", Extensions: []string{".a", ".b"}},
				{Name: "B", Extensions: []string{".b"}},
			},
			wantLangNames: []string{"B", "A"},
		},
	}
	for _, test := range tests {
		langNames := langNames(test.langs.ByFilename(test.name))
		if !reflect.DeepEqual(langNames, test.wantLangNames) {
			t.Errorf("filename %q: got languages %v, want %v", test.name, langNames, test.wantLangNames)
			continue
		}
	}
}

func langNames(langs []*Language) []string {
	if langs == nil {
		return nil
	}
	names := make([]string, len(langs))
	for i, l := range langs {
		names[i] = l.Name
	}
	return names
}
