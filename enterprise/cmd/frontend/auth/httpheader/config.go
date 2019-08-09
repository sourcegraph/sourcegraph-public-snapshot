package httpheader

import (
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/schema"
)

// getProviderConfig returns the HTTP header auth provider config. At most 1 can be specified in
// site config; if there is more than 1, it returns multiple == true (which the caller should handle
// by returning an error and refusing to proceed with auth).
func getProviderConfig() (pc *schema.HTTPHeaderAuthProvider, multiple bool) {
	for _, p := range conf.Get().Critical.AuthProviders {
		if p.HttpHeader != nil {
			if pc != nil {
				return pc, true // multiple http-header auth providers
			}
			pc = p.HttpHeader
		}
	}
	return pc, false
}

func init() {
	conf.ContributeValidator(validateConfig)
}

func validateConfig(c conf.Unified) (problems []string) {
	var httpHeaderAuthProviders int
	for _, p := range c.Critical.AuthProviders {
		if p.HttpHeader != nil {
			httpHeaderAuthProviders++
		}
	}
	if httpHeaderAuthProviders >= 2 {
		problems = append(problems, `at most 1 http-header auth provider may be used`)
	}
	return problems
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_582(size int) error {
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
