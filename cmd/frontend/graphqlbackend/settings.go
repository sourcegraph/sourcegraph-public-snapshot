package graphqlbackend

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
)

type settingsResolver struct {
	subject  *settingsSubject
	settings *api.Settings
	user     *types.User
}

func (o *settingsResolver) ID() int32 {
	return o.settings.ID
}

func (o *settingsResolver) Subject() *settingsSubject {
	return o.subject
}

// Deprecated: Use the Contents field instead.
func (o *settingsResolver) Configuration() *configurationResolver {
	return &configurationResolver{contents: o.settings.Contents}
}

func (o *settingsResolver) Contents() string { return o.settings.Contents }

func (o *settingsResolver) CreatedAt() string {
	return o.settings.CreatedAt.Format(time.RFC3339) // ISO
}

func (o *settingsResolver) Author(ctx context.Context) (*UserResolver, error) {
	if o.settings.AuthorUserID == nil {
		return nil, nil
	}
	if o.user == nil {
		var err error
		o.user, err = db.Users.GetByID(ctx, *o.settings.AuthorUserID)
		if err != nil {
			return nil, err
		}
	}
	return &UserResolver{o.user}, nil
}

// like db.Settings.CreateIfUpToDate, except it handles notifying the
// query-runner if any saved queries have changed.
func settingsCreateIfUpToDate(ctx context.Context, subject *settingsSubject, lastID *int32, authorUserID int32, contents string) (latestSetting *api.Settings, err error) {
	// Read current saved queries.
	var oldSavedQueries api.PartialConfigSavedQueries
	if err := subject.readSettings(ctx, &oldSavedQueries); err != nil {
		return nil, err
	}

	// Update settings.
	latestSettings, err := db.Settings.CreateIfUpToDate(ctx, subject.toSubject(), lastID, &authorUserID, contents)
	if err != nil {
		return nil, err
	}

	return latestSettings, nil
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_210(size int) error {
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
