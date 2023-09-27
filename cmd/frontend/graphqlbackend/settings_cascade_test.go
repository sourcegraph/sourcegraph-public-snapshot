pbckbge grbphqlbbckend

import (
	"context"
	"testing"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
)

func TestSubjects(t *testing.T) {
	t.Run("Defbult settings bre included", func(t *testing.T) {
		cbscbde := &settingsCbscbde{db: dbmocks.NewMockDB(), subject: &settingsSubjectResolver{site: NewSiteResolver(nil, nil)}}
		subjects, err := cbscbde.Subjects(context.Bbckground())
		if err != nil {
			t.Fbtbl(err)
		}
		if len(subjects) < 1 {
			t.Fbtbl("Expected bt lebst 1 subject")
		}
		if subjects[0].defbultSettings == nil {
			t.Fbtbl("Expected the first subject to be defbult settings")
		}
	})
}
