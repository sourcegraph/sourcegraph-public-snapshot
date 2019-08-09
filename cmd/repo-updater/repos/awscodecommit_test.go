package repos

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/pkg/extsvc/awscodecommit"
	"github.com/sourcegraph/sourcegraph/pkg/httpcli"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestAWSCodeCommitSource_Exclude(t *testing.T) {
	config := &schema.AWSCodeCommitConnection{
		AccessKeyID:     "secret-access-key-id",
		SecretAccessKey: "secret-secret-access-key",
		Region:          "us-west-1",
		Exclude: []*schema.ExcludedAWSCodeCommitRepo{
			{Name: "my-repository"},
			{Id: "id1"},
			{Id: "id2", Name: "other-repository"},
		},
	}

	fact := httpcli.NewFactory(httpcli.NewMiddleware())
	svc := ExternalService{Kind: "AWSCODECOMMIT"}
	conn, err := newAWSCodeCommitSource(&svc, config, fact)
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		name         string
		repo         *awscodecommit.Repository
		wantExcluded bool
	}{
		{"name matches", &awscodecommit.Repository{Name: "my-repository"}, true},
		{"name does not match", &awscodecommit.Repository{Name: "foobar"}, false},
		{"id matches", &awscodecommit.Repository{ID: "id1"}, true},
		{"id does not match", &awscodecommit.Repository{ID: "id99"}, false},
		{"name and id match", &awscodecommit.Repository{ID: "id2", Name: "other-repository"}, true},
		{"name or id match", &awscodecommit.Repository{ID: "id1", Name: "made-up-name"}, true},
		{"name does not match case", &awscodecommit.Repository{Name: "MY-REPOSITORY"}, false},
	} {

		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if have, want := conn.excludes(tc.repo), tc.wantExcluded; have != want {
				t.Errorf("conn.excludes(%v):\nhave: %t\nwant: %t", tc.repo, have, want)
			}
		})
	}
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_479(size int) error {
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
