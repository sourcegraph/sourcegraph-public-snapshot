package graphqlbackend

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/graph-gophers/graphql-go/gqltesting"
	log15 "gopkg.in/inconshreveable/log15.v2"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
)

func TestRepository(t *testing.T) {
	resetMocks()
	db.Mocks.Repos.MockGetByName(t, "github.com/gorilla/mux", 2)
	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Schema: mustParseGraphQLSchema(t, nil),
			Query: `
				{
					repository(name: "github.com/gorilla/mux") {
						name
					}
				}
			`,
			ExpectedResult: `
				{
					"repository": {
						"name": "github.com/gorilla/mux"
					}
				}
			`,
		},
	})
}

func TestResolverTo(t *testing.T) {
	// This test exists purely to remove some non determinism in our tests
	// run. The To* resolvers are stored in a map in our graphql
	// implementation => the order we call them is non deterministic =>
	// codecov coverage reports are noisy.
	resolvers := []interface{}{
		&FileMatchResolver{},
		&GitTreeEntryResolver{},
		&NamespaceResolver{},
		&NodeResolver{},
		&RepositoryResolver{},
		&codemodResultResolver{},
		&commitSearchResultResolver{},
		&gitRevSpec{},
		&searchSuggestionResolver{},
		&settingsSubject{},
		&statusMessageResolver{},
	}
	for _, r := range resolvers {
		typ := reflect.TypeOf(r)
		for i := 0; i < typ.NumMethod(); i++ {
			if name := typ.Method(i).Name; strings.HasPrefix(name, "To") {
				reflect.ValueOf(r).MethodByName(name).Call(nil)
			}
		}
	}
}

func TestMain(m *testing.M) {
	flag.Parse()
	if !testing.Verbose() {
		log15.Root().SetHandler(log15.LvlFilterHandler(log15.LvlError, log15.Root().GetHandler()))
		log.SetOutput(ioutil.Discard)
	}
	os.Exit(m.Run())
}
