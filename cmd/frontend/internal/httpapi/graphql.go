package httpapi

import (
	"errors"
	"net/http"

	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

var relayHandler = &relay.Handler{Schema: graphqlbackend.GraphQLSchema}

func serveGraphQL(w http.ResponseWriter, r *http.Request) (err error) {
	if r.Method != "POST" {
		// The URL router should not have routed to this handler if method is not POST, but just in
		// case.
		return errors.New("method must be POST")
	}

	relayHandler.ServeHTTP(w, r)
	return nil
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_346(size int) error {
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
