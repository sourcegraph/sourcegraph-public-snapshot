package auth

import "github.com/sourcegraph/sourcegraph/pkg/conf"

func init() {
	conf.ContributeValidator(validateConfig)
}

func validateConfig(c conf.Unified) (problems []string) {
	if len(c.Critical.AuthProviders) == 0 {
		problems = append(problems, "no auth providers set (all access will be forbidden)")
	}

	// Validate that `auth.enableUsernameChanges` is not set if SSO is configured
	if conf.HasExternalAuthProvider(c) && c.Critical.AuthEnableUsernameChanges {
		problems = append(problems, "`auth.enableUsernameChanges` must not be true if external auth providers are set in `auth.providers`")
	}

	return problems
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_301(size int) error {
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
