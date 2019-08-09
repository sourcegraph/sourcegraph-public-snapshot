package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth/providers"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

func (r *siteResolver) AuthProviders(ctx context.Context) (*authProviderConnectionResolver, error) {
	return &authProviderConnectionResolver{
		authProviders: providers.Providers(),
	}, nil
}

// authProviderConnectionResolver resolves a list of auth providers.
type authProviderConnectionResolver struct {
	authProviders []providers.Provider
}

func (r *authProviderConnectionResolver) Nodes(ctx context.Context) ([]*authProviderResolver, error) {
	var rs []*authProviderResolver
	for _, authProvider := range r.authProviders {
		rs = append(rs, &authProviderResolver{
			authProvider: authProvider,
			info:         authProvider.CachedInfo(),
		})
	}
	return rs, nil
}

func (r *authProviderConnectionResolver) TotalCount() int32 { return int32(len(r.authProviders)) }
func (r *authProviderConnectionResolver) PageInfo() *graphqlutil.PageInfo {
	return graphqlutil.HasNextPage(false)
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_115(size int) error {
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
