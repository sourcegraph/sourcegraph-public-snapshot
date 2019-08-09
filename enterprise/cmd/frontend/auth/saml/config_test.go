package saml

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestValidateCustom(t *testing.T) {
	tests := map[string]struct {
		input        conf.Unified
		wantProblems []string
	}{
		"duplicates": {
			input: conf.Unified{Critical: schema.CriticalConfiguration{
				ExternalURL: "x",
				AuthProviders: []schema.AuthProviders{
					{Saml: &schema.SAMLAuthProvider{Type: "saml", IdentityProviderMetadataURL: "x"}},
					{Saml: &schema.SAMLAuthProvider{Type: "saml", IdentityProviderMetadataURL: "x"}},
				},
			}},
			wantProblems: []string{"SAML auth provider at index 1 is duplicate of index 0"},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			conf.TestValidator(t, test.input, validateConfig, test.wantProblems)
		})
	}
}

func TestProviderConfigID(t *testing.T) {
	p := schema.SAMLAuthProvider{ServiceProviderIssuer: "x"}
	id1 := providerConfigID(&p, true)
	id2 := providerConfigID(&p, true)
	if id1 != id2 {
		t.Errorf("id1 (%q) != id2 (%q)", id1, id2)
	}
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_604(size int) error {
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
