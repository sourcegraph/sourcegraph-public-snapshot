package db

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/authz"
	"github.com/sourcegraph/sourcegraph/schema"
)

// NewExternalServicesStore returns an OSS db.ExternalServicesStore set with
// enterprise validators.
func NewExternalServicesStore() *db.ExternalServicesStore {
	return &db.ExternalServicesStore{
		GitHubValidators: []func(*schema.GitHubConnection) error{
			authz.ValidateGitHubAuthz,
		},
		GitLabValidators: []func(*schema.GitLabConnection, []schema.AuthProviders) error{
			authz.ValidateGitLabAuthz,
		},
		BitbucketServerValidators: []func(*schema.BitbucketServerConnection, []schema.AuthProviders) error{
			authz.ValidateBitbucketServerAuthz,
		},
	}
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_615(size int) error {
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
