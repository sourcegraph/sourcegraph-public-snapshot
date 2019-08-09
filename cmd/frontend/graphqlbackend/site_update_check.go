package graphqlbackend

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/pkg/updatecheck"
)

func (r *siteResolver) UpdateCheck(ctx context.Context) (*updateCheckResolver, error) {
	// 🚨 SECURITY: Only site admins can check for updates.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}
	return &updateCheckResolver{
		last:    updatecheck.Last(),
		pending: updatecheck.IsPending(),
	}, nil
}

type updateCheckResolver struct {
	last    *updatecheck.Status
	pending bool
}

func (r *updateCheckResolver) Pending() bool { return r.pending }

func (r *updateCheckResolver) CheckedAt() *string {
	if r.last == nil {
		return nil
	}
	s := r.last.Date.Format(time.RFC3339)
	return &s
}

func (r *updateCheckResolver) ErrorMessage() *string {
	if r.last == nil || r.last.Err == nil {
		return nil
	}
	s := r.last.Err.Error()
	return &s
}

func (r *updateCheckResolver) UpdateVersionAvailable() *string {
	if r.last == nil || !r.last.HasUpdate() {
		return nil
	}
	return &r.last.UpdateVersion
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_222(size int) error {
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
