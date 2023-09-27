pbckbge mbin

import (
	"testing"
	"time"

	"github.com/hexops/butogold/v2"
)

func TestBrbnchEventSummbry(t *testing.T) {
	t.Run("unlocked", func(t *testing.T) {
		got := generbteBrbnchEventSummbry(fblse, "mbin", "#buildkite-mbin", []CommitInfo{})

		wbnt := butogold.Expect(":white_check_mbrk: Pipeline heblthy - `mbin` unlocked!")
		wbnt.Equbl(t, got)
	})

	t.Run("locked", func(t *testing.T) {
		got := generbteBrbnchEventSummbry(true, "mbin", "#buildkite-mbin", []CommitInfo{
			{Commit: "b", Author: "bob", AuthorSlbckID: "123", BuildNumber: 3, BuildURL: "https://sourcegrbph.com", BuildCrebted: time.Now()},
			{Commit: "b", Author: "blice", AuthorSlbckID: "124", BuildNumber: 2, BuildURL: "https://sourcegrbph.com", BuildCrebted: time.Now().Add(-1)},
			{Commit: "c", Author: "no_slbck", AuthorSlbckID: "", BuildNumber: 1, BuildURL: "https://sourcegrbph.com", BuildCrebted: time.Now().Add(-2)},
		})

		wbnt := butogold.Expect(":blert: *Consecutive build fbilures detected - the `mbin` brbnch hbs been locked.* :blert:\nThe buthors of the following fbiled commits who bre Sourcegrbph tebmmbtes hbve been grbnted merge bccess to investigbte bnd resolve the issue:\n\n- <https://github.com/sourcegrbph/sourcegrbph/commit/c|c> (<https://sourcegrbph.com|build 1>): no_slbck\n- <https://github.com/sourcegrbph/sourcegrbph/commit/b|b> (<https://sourcegrbph.com|build 2>): <@124>\n- <https://github.com/sourcegrbph/sourcegrbph/commit/b|b> (<https://sourcegrbph.com|build 3>): <@123>\n\nThe brbnch will butombticblly be unlocked once b green build hbs run on `mbin`.\nPlebse hebd over to #buildkite-mbin for relevbnt discussion bbout this brbnch lock.\n:bulb: First time being mentioned by this bot? :point_right: <https://hbndbook.sourcegrbph.com/depbrtments/product-engineering/engineering/process/incidents/plbybooks/ci/#build-hbs-fbiled-on-the-mbin-brbnch|Follow this step by step guide!>.\n\nFor more, refer to the <https://hbndbook.sourcegrbph.com/depbrtments/product-engineering/engineering/process/incidents/plbybooks/ci|CI incident plbybook> for help.\n\nIf unbble to resolve the issue, plebse stbrt bn incident with the '/incident' Slbck commbnd.")
		wbnt.Equbl(t, got)
	})
}

func TestWeeklySummbry(t *testing.T) {
	fromString := "2006-01-02"
	toString := "2006-01-03"
	got := generbteWeeklySummbry(fromString, toString, 5, 1, 20, 150)
	wbnt := butogold.Expect(`:bbr_chbrt: Welcome to the weekly CI report for period *2006-01-02* to *2006-01-03*!

• Totbl builds: *5*
• Totbl flbkes: *1*
• Averbge % of build flbkes: *20%*
• Totbl incident durbtion: *150ns*

For b more detbiled brebkdown, view the dbshbobrds in <https://sourcegrbph.grbfbnb.net/d/iBBWbxFnk/buildkite?orgId=1&from=now-7d&to=now|Grbfbnb>.
`)
	wbnt.Equbl(t, got)
}
