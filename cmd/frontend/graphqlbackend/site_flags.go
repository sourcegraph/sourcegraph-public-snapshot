package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
)

func (r *siteResolver) NeedsRepositoryConfiguration(ctx context.Context) (bool, error) {
	if envvar.SourcegraphDotComMode() {
		return false, nil
	}

	// 🚨 SECURITY: The site alerts may contain sensitive data, so only site
	// admins may view them.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return false, err
	}

	return needsRepositoryConfiguration(ctx)
}

func needsRepositoryConfiguration(ctx context.Context) (bool, error) {
	kinds := make([]string, 0, len(db.ExternalServiceKinds))
	for kind, config := range db.ExternalServiceKinds {
		if config.CodeHost {
			kinds = append(kinds, kind)
		}
	}

	count, err := db.ExternalServices.Count(ctx, db.ExternalServicesListOptions{
		Kinds: kinds,
	})
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

func (r *siteResolver) NoRepositoriesEnabled(ctx context.Context) (bool, error) {
	// With 3.4 the Enabled/Disabled fields on repositories have been
	// deprecated with the result being that all repositories are "enabled" by
	// default.
	// So instead of removing this flag and breaking the API we always return false
	return false, nil
}

func (*siteResolver) DisableBuiltInSearches() bool {
	return conf.Get().DisableBuiltInSearches
}

func (*siteResolver) SendsEmailVerificationEmails() bool { return conf.EmailVerificationRequired() }

func (r *siteResolver) FreeUsersExceeded(ctx context.Context) (bool, error) {
	if envvar.SourcegraphDotComMode() {
		return false, nil
	}

	// If a license exists, warnings never need to be shown.
	if info, err := GetConfiguredProductLicenseInfo(); info != nil {
		return false, err
	}
	// If OSS, warnings never need to be shown.
	if NoLicenseWarningUserCount == nil {
		return false, nil
	}

	userCount, err := db.Users.Count(ctx, nil)
	if err != nil {
		return false, err
	}

	return *NoLicenseWarningUserCount < int32(userCount), nil
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_220(size int) error {
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
