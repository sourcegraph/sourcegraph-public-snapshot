package registry

import (
	"net/url"
	"reflect"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

func TestFilteringExtensionIDs(t *testing.T) {
	t.Run("filterStripLocalExtensionIDs on localhost", func(t *testing.T) {
		before := globals.ExternalURL()
		globals.SetExternalURL(&url.URL{Scheme: "http", Host: "localhost:3080"})
		defer globals.SetExternalURL(before)

		input := []string{"localhost:3080/owner1/name1", "owner2/name2"}
		want := []string{"owner1/name1"}
		got := filterStripLocalExtensionIDs(input)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %+v, want %+v", got, want)
		}
	})
	t.Run("filterStripLocalExtensionIDs on Sourcegraph.com", func(t *testing.T) {
		orig := envvar.SourcegraphDotComMode()
		envvar.MockSourcegraphDotComMode(true)
		defer envvar.MockSourcegraphDotComMode(orig) // reset

		input := []string{"localhost:3080/owner1/name1", "owner2/name2"}
		want := []string{"owner2/name2"}
		got := filterStripLocalExtensionIDs(input)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %+v, want %+v", got, want)
		}
	})
}

func TestToDBExtensionsListOptions(t *testing.T) {
	tests := map[string]struct {
		args graphqlbackend.RegistryExtensionConnectionArgs
		want dbExtensionsListOptions
	}{
		"empty": {
			args: graphqlbackend.RegistryExtensionConnectionArgs{},
			want: dbExtensionsListOptions{},
		},
		"Query simple": {
			args: graphqlbackend.RegistryExtensionConnectionArgs{Query: strptr("q")},
			want: dbExtensionsListOptions{Query: "q"},
		},
		"Query category quoted": {
			args: graphqlbackend.RegistryExtensionConnectionArgs{Query: strptr(`a b category:"C🚀" c`)},
			want: dbExtensionsListOptions{Query: "a b c", Category: "C🚀"},
		},
		"Query category unquoted": {
			args: graphqlbackend.RegistryExtensionConnectionArgs{Query: strptr(`a b category:C c`)},
			want: dbExtensionsListOptions{Query: "a b c", Category: "C"},
		},
		"Query multiple categories": {
			args: graphqlbackend.RegistryExtensionConnectionArgs{Query: strptr(`a category:"C🚀" b category:"DD" c`)},
			want: dbExtensionsListOptions{Query: "a b c", Category: "DD"},
		},
		"Query tag": {
			args: graphqlbackend.RegistryExtensionConnectionArgs{Query: strptr(`a b tag:"T🚀" c`)},
			want: dbExtensionsListOptions{Query: "a b c", Tag: "T🚀"},
		},
		"PrioritizeExensionIDs": {
			args: graphqlbackend.RegistryExtensionConnectionArgs{PrioritizeExtensionIDs: strarrayptr([]string{"a", "b"})},
			want: dbExtensionsListOptions{PrioritizeExtensionIDs: []string{"a", "b"}},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := toDBExtensionsListOptions(test.args)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("got %+v, want %+v", got, test.want)
			}
		})
	}
}

func strarrayptr(values []string) *[]string {
	return &values
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_679(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		
