package graphqlbackend

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
)

func TestSubjects(t *testing.T) {
	t.Run("Default settings are included", func(t *testing.T) {
		subject := &settingsSubjectResolver{site: NewSiteResolver(nil, nil)}
		subject.mockCheckedAccessForTest()
		cascade := &settingsCascade{db: dbmocks.NewMockDB(), subject: subject}
		subjects, err := cascade.Subjects(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if len(subjects) < 1 {
			t.Fatal("Expected at least 1 subject")
		}
		if subjects[0].defaultSettings == nil {
			t.Fatal("Expected the first subject to be default settings")
		}
	})
}
