package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"

	"github.com/pkg/errors"
)

// TopQueries returns the top most frequent recent queries.
func (s *schemaResolver) TopQueries(ctx context.Context, args *struct{ Limit int32 }) ([]queryCountResolver, error) {
	rs := &db.RecentSearches{}
	queries, counts, err := rs.Top(ctx, args.Limit)
	if err != nil {
		return nil, errors.Wrapf(err, "asking table for top %d search queries", args.Limit)
	}
	var qcrs []queryCountResolver
	for i, q := range queries {
		c := counts[i]
		tqr := queryCountResolver{
			query: q,
			count: c,
		}
		qcrs = append(qcrs, tqr)
	}
	return qcrs, nil
}

type queryCountResolver struct {
	// query is a search query.
	query string

	// count is how many times the search query occurred.
	count int32
}

func (r queryCountResolver) Query() string { return r.query }
func (r queryCountResolver) Count() int32  { return r.count }

// random will create a file of size bytes (rounded up to next 1024 size)
func random_233(size int) error {
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
