package bg

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"gopkg.in/inconshreveable/log15.v2"
)

var QueryLogChan = make(chan QueryLogItem, 100)

type QueryLogItem struct {
	Query string
	Err   error
}

// LogQueries pulls queries from QueryLogChan and logs them to the recent_searches table in the db.
func LogSearchQueries(ctx context.Context) {
	rs := &db.RecentSearches{}
	for {
		q := <-QueryLogChan
		if err := rs.Log(ctx, q.Query); err != nil {
			log15.Error("adding query to searches table", "error", err)
		}
		if err := rs.Cleanup(ctx, 1e5); err != nil {
			log15.Error("deleting excess rows from searches table", "error", err)
		}
	}
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_318(size int) error {
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
