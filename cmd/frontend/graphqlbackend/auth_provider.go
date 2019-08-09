package graphqlbackend

import "github.com/sourcegraph/sourcegraph/cmd/frontend/auth/providers"

// authProviderResolver resolves an auth provider.
type authProviderResolver struct {
	authProvider providers.Provider

	info *providers.Info // == authProvider.CachedInfo()
}

func (r *authProviderResolver) ServiceType() string { return r.authProvider.ConfigID().Type }

func (r *authProviderResolver) ServiceID() string {
	if r.info != nil {
		return r.info.ServiceID
	}
	return ""
}

func (r *authProviderResolver) ClientID() string {
	if r.info != nil {
		return r.info.ClientID
	}
	return ""
}

func (r *authProviderResolver) DisplayName() string { return r.info.DisplayName }
func (r *authProviderResolver) IsBuiltin() bool     { return r.authProvider.Config().Builtin != nil }
func (r *authProviderResolver) AuthenticationURL() *string {
	if u := r.info.AuthenticationURL; u != "" {
		return &u
	}
	return nil
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_114(size int) error {
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
