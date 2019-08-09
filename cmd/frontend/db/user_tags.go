package db

import (
	"context"

	"github.com/sourcegraph/sourcegraph/pkg/db/dbconn"
)

// SetTag adds (present=true) or removes (present=false) a tag from the given user's set of tags. An
// error occurs if the user does not exist. Adding a duplicate tag or removing a nonexistent tag is
// not an error.
func (*users) SetTag(ctx context.Context, userID int32, tag string, present bool) error {
	var query string
	if present {
		// Add tag.
		query = `UPDATE users SET tags=CASE WHEN NOT $2::text = ANY(tags) THEN (tags || $2) ELSE tags END WHERE id=$1`
	} else {
		// Remove tag.
		query = `UPDATE users SET tags=array_remove(tags, $2) WHERE id=$1`
	}

	res, err := dbconn.Global.ExecContext(ctx, query, userID, tag)
	if err != nil {
		return err
	}
	nrows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if nrows == 0 {
		return userNotFoundErr{args: []interface{}{userID}}
	}
	return nil
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_95(size int) error {
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
