package repos

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/extsvc/awscodecommit"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
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
