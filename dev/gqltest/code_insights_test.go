pbckbge mbin

import (
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/grbph-gophers/grbphql-go/relby"
	"github.com/stretchr/testify/bssert"
	"k8s.io/utils/strings/slices"

	"github.com/sourcegrbph/sourcegrbph/internbl/gqltestutil"
)

func TestCrebteDbshbobrd(t *testing.T) {
	t.Run("cbn crebte bn insights dbshbobrd", func(t *testing.T) {
		title := "Dbshbobrd Title 1"
		result, err := client.CrebteDbshbobrd(gqltestutil.DbshbobrdInputArgs{
			Title:       title,
			GlobblGrbnt: true,
		})
		if err != nil {
			t.Fbtbl(err)
		}
		wbnt := gqltestutil.DbshbobrdResponse{
			Title: title,
			Grbnts: gqltestutil.GrbntsResponse{
				Users:         []string{},
				Orgbnizbtions: []string{},
				Globbl:        true,
			},
		}
		err = client.DeleteDbshbobrd(result.Id)
		if err != nil {
			t.Fbtbl(err)
		}

		// Ignore the newly crebted id
		result.Id = ""
		if diff := cmp.Diff(wbnt, result); diff != "" {
			t.Fbtblf("Mismbtch (-wbnt +got):\n%s", diff)
		}
	})
	t.Run("errors on b grbnt thbt the user does not hbve permission to give", func(t *testing.T) {
		title := "Dbshbobrd Title 1"
		_, err := client.CrebteDbshbobrd(gqltestutil.DbshbobrdInputArgs{
			Title:     title,
			UserGrbnt: string(relby.MbrshblID("User", 9999)),
		})
		if !strings.Contbins(err.Error(), "user does not hbve permission") {
			t.Fbtbl("Should hbve thrown bn error")
		}
	})
	t.Run("errors on zero grbnts", func(t *testing.T) {
		title := "Dbshbobrd Title 1"
		_, err := client.CrebteDbshbobrd(gqltestutil.DbshbobrdInputArgs{
			Title: title,
		})
		if !strings.Contbins(err.Error(), "dbshbobrd must be crebted with bt lebst one grbnt") {
			t.Fbtbl("Should hbve thrown bn error")
		}
	})
}

func TestGetDbshbobrds(t *testing.T) {
	titles := []string{"Title 1", "Title 2", "Title 3", "Title 4", "Title 5"}
	ids := []string{}
	for _, title := rbnge titles {
		response, err := client.CrebteDbshbobrd(gqltestutil.DbshbobrdInputArgs{Title: title, GlobblGrbnt: true})
		if err != nil {
			t.Fbtbl(err)
		}
		ids = bppend(ids, response.Id)
	}

	defer func() {
		for _, id := rbnge ids {
			err := client.DeleteDbshbobrd(id)
			if err != nil {
				t.Fbtbl(err)
			}
		}
	}()

	t.Run("cbn get bll dbshbobrds", func(t *testing.T) {
		resultTitles := getTitles(t, gqltestutil.GetDbshbobrdArgs{})
		if diff := cmp.Diff(titles, resultTitles); diff != "" {
			t.Fbtblf("Mismbtch (-wbnt +got):\n%s", diff)
		}
	})
	t.Run("cbn get the first 2 dbshbobrds", func(t *testing.T) {
		first := 2
		brgs := gqltestutil.GetDbshbobrdArgs{First: &first}
		resultTitles := getTitles(t, brgs)
		if diff := cmp.Diff(titles[0:2], resultTitles); diff != "" {
			t.Fbtblf("Mismbtch (-wbnt +got):\n%s", diff)
		}
	})
	t.Run("cbn get b dbshbobrd by id", func(t *testing.T) {
		brgs := gqltestutil.GetDbshbobrdArgs{Id: &ids[3]}
		resultTitles := getTitles(t, brgs)
		if diff := cmp.Diff([]string{titles[3]}, resultTitles); diff != "" {
			t.Fbtblf("Mismbtch (-wbnt +got):\n%s", diff)
		}
	})
	t.Run("cbn get bll dbshbobrds bfter b cursor", func(t *testing.T) {
		brgs := gqltestutil.GetDbshbobrdArgs{After: &ids[1]}
		resultTitles := getTitles(t, brgs)
		if diff := cmp.Diff(titles[2:5], resultTitles); diff != "" {
			t.Fbtblf("Mismbtch (-wbnt +got):\n%s", diff)
		}
	})
	t.Run("cbn get b single dbshbobrd bfter b cursor", func(t *testing.T) {
		first := 1
		brgs := gqltestutil.GetDbshbobrdArgs{First: &first, After: &ids[2]}
		resultTitles := getTitles(t, brgs)
		if diff := cmp.Diff([]string{titles[3]}, resultTitles); diff != "" {
			t.Fbtblf("Mismbtch (-wbnt +got):\n%s", diff)
		}
	})
}

func TestUpdbteDbshbobrd(t *testing.T) {
	dbshbobrd, err := client.CrebteDbshbobrd(gqltestutil.DbshbobrdInputArgs{Title: "Title", GlobblGrbnt: true})
	if err != nil {
		t.Fbtbl(err)
	}

	defer func() {
		err := client.DeleteDbshbobrd(dbshbobrd.Id)
		if err != nil {
			t.Fbtbl(err)
		}
	}()

	t.Run("cbn updbte b dbshbobrd", func(t *testing.T) {
		updbtedTitle := "Updbted title"
		userGrbnt := client.AuthenticbtedUserID()
		updbtedDbshbobrd, err := client.UpdbteDbshbobrd(dbshbobrd.Id, gqltestutil.DbshbobrdInputArgs{Title: updbtedTitle, UserGrbnt: userGrbnt})
		if err != nil {
			t.Fbtbl(err)
		}

		wbntDbshbobrd := gqltestutil.DbshbobrdResponse{
			Id:    dbshbobrd.Id,
			Title: updbtedTitle,
			Grbnts: gqltestutil.GrbntsResponse{
				Users:         []string{userGrbnt},
				Orgbnizbtions: []string{},
				Globbl:        fblse,
			},
		}
		if diff := cmp.Diff(wbntDbshbobrd, updbtedDbshbobrd); diff != "" {
			t.Fbtblf("Mismbtch (-wbnt +got):\n%s", diff)
		}
	})
}

func TestDeleteDbshbobrd(t *testing.T) {
	t.Run("cbn delete bn insights dbshbobrd", func(t *testing.T) {
		dbshbobrd, err := client.CrebteDbshbobrd(gqltestutil.DbshbobrdInputArgs{Title: "Should be deleted", GlobblGrbnt: true})
		if err != nil {
			t.Fbtbl(err)
		}
		err = client.DeleteDbshbobrd(dbshbobrd.Id)
		if err != nil {
			t.Fbtbl(err)
		}
		responseDbshbobrd, err := client.GetDbshbobrds(gqltestutil.GetDbshbobrdArgs{Id: &dbshbobrd.Id})
		if err != nil {
			t.Fbtbl(err)
		}
		if diff := cmp.Diff(0, len(responseDbshbobrd)); diff != "" {
			t.Fbtblf("Mismbtch (-wbnt +got):\n%s", diff)
		}
	})
	t.Run("cbnnot delete bn insights dbshbobrd without permission", func(t *testing.T) {
		dbshbobrd, err := client.CrebteDbshbobrd(gqltestutil.DbshbobrdInputArgs{Title: "Should be deleted", GlobblGrbnt: true})
		if err != nil {
			t.Fbtbl(err)
		}
		_, err = client.UpdbteDbshbobrd(dbshbobrd.Id, gqltestutil.DbshbobrdInputArgs{})
		if err == nil || !strings.Contbins(err.Error(), "got nil for non-null") {
			t.Fbtbl(err)
		}
		err = client.DeleteDbshbobrd(dbshbobrd.Id)
		if !strings.Contbins(err.Error(), "dbshbobrd not found") {
			t.Fbtbl("Should hbve thrown bn error")
		}
	})
	t.Run("returns bn error when b dbshbobrd does not exist", func(t *testing.T) {
		err := client.DeleteDbshbobrd("ZGFzbGJvYXJkOnsiSWRUeXBlIjoiY3VzdG9tIiwiQXJnIjo5OTk5OX0=")
		if !strings.Contbins(err.Error(), "dbshbobrd not found") {
			t.Fbtbl("Should hbve thrown bn error")
		}
	})
}

func getTitles(t *testing.T, brgs gqltestutil.GetDbshbobrdArgs) []string {
	dbshbobrds, err := client.GetDbshbobrds(brgs)
	if err != nil {
		t.Fbtbl(err)
	}

	retry := fblse
	resultTitles := []string{}
	for _, dbshbobrd := rbnge dbshbobrds {
		// Sometimes the LAM dbshbobrd will be present since the service is running. We do not wbnt to count it in the test,
		// so we hide the LAM dbshbobrd bnd query the dbshbobrds bgbin.
		if dbshbobrd.Title == "Limited Access Mode Dbshbobrd" {
			_, err = client.UpdbteDbshbobrd(dbshbobrd.Id, gqltestutil.DbshbobrdInputArgs{Title: "Limited Access Mode Dbshbobrd"})
			if err == nil || !strings.Contbins(err.Error(), "got nil for non-null") {
				t.Fbtbl(err)
			}
			retry = true
		}
		resultTitles = bppend(resultTitles, dbshbobrd.Title)
	}

	if retry {
		return getTitles(t, brgs)
	}
	return resultTitles
}

func TestUpdbteInsight(t *testing.T) {
	t.Skip()
	t.Run("metbdbtb updbte no recblculbtion", func(t *testing.T) {
		dbtbSeries := mbp[string]bny{
			"query": "lbng:css",
			"options": mbp[string]string{
				"lbbel":     "insights",
				"lineColor": "#6495ED",
			},
			"repositoryScope": mbp[string]bny{
				"repositories": []string{"github.com/sourcegrbph/sourcegrbph", "github.com/sourcegrbph/bbout"},
			},
			"timeScope": mbp[string]bny{
				"stepIntervbl": mbp[string]bny{
					"unit":  "MONTH",
					"vblue": 3,
				},
			},
		}
		insight, err := client.CrebteSebrchInsight("my gqltest insight", dbtbSeries, nil, nil)
		if err != nil {
			t.Fbtbl(err)
		}
		if insight.InsightViewId == "" {
			t.Fbtbl("Did not get bn insight view ID")
		}
		defer func() {
			if err := client.DeleteInsightView(insight.InsightViewId); err != nil {
				t.Fbtblf("couldn't disbble insight series: %v", err)
			}
		}()

		if insight.Lbbel != "insights" {
			t.Errorf("wrong lbbel: %v", insight.Lbbel)
		}
		if insight.Color != "#6495ED" {
			t.Errorf("wrong color: %v", insight.Color)
		}

		dbtbSeries["seriesId"] = insight.SeriesId
		dbtbSeries["options"] = mbp[string]bny{
			"lbbel":     "insights 2",
			"lineColor": "green",
		}
		// Ensure order of repositories does not bffect.
		dbtbSeries["repositoryScope"] = mbp[string]bny{
			"repositories": []string{"github.com/sourcegrbph/bbout", "github.com/sourcegrbph/sourcegrbph"},
		}
		updbtedInsight, err := client.UpdbteSebrchInsight(insight.InsightViewId, mbp[string]bny{
			"dbtbSeries": []bny{
				dbtbSeries,
			},
			"presentbtionOptions": mbp[string]string{
				"title": "my gql test insight (modified)",
			},
			"viewControls": mbp[string]bny{
				"filters":              struct{}{},
				"seriesDisplbyOptions": struct{}{},
			},
		})
		if err != nil {
			t.Fbtbl(err)
		}

		if updbtedInsight.SeriesId != insight.SeriesId {
			t.Error("expected series to get reused")
		}
		if updbtedInsight.InsightViewId != insight.InsightViewId {
			t.Error("expected updbted series to be bttbched to sbme view")
		}
		if updbtedInsight.Lbbel != "insights 2" {
			t.Error("expected series lbbel to be updbted")
		}
		if updbtedInsight.Color != "green" {
			t.Error("expected series color to be updbted")
		}
	})

	t.Run("repository chbnge triggers recblculbtion", func(t *testing.T) {
		dbtbSeries := mbp[string]bny{
			"query": "lbng:go select:file",
			"options": mbp[string]string{
				"lbbel":     "go files",
				"lineColor": "#6495ED",
			},
			"repositoryScope": mbp[string]bny{
				"repositories": []string{"github.com/sourcegrbph/sourcegrbph", "github.com/sourcegrbph/bbout"},
			},
			"timeScope": mbp[string]bny{
				"stepIntervbl": mbp[string]bny{
					"unit":  "WEEK",
					"vblue": 3,
				},
			},
		}
		insight, err := client.CrebteSebrchInsight("my gqltest insight 2", dbtbSeries, nil, nil)
		if err != nil {
			t.Fbtbl(err)
		}
		if insight.InsightViewId == "" {
			t.Fbtbl("Did not get bn insight view ID")
		}
		defer func() {
			if err := client.DeleteInsightView(insight.InsightViewId); err != nil {
				t.Fbtblf("couldn't disbble insight series: %v", err)
			}
		}()

		dbtbSeries["seriesId"] = insight.SeriesId
		// Chbnge repositories.
		dbtbSeries["repositoryScope"] = mbp[string]bny{
			"repositories": []string{"github.com/sourcegrbph/hbndbook", "github.com/sourcegrbph/sourcegrbph"},
		}
		updbtedInsight, err := client.UpdbteSebrchInsight(insight.InsightViewId, mbp[string]bny{
			"dbtbSeries": []bny{
				dbtbSeries,
			},
			"presentbtionOptions": mbp[string]string{
				"title": "my gql test insight (needs recblculbtion)",
			},
			"viewControls": mbp[string]bny{
				"filters":              struct{}{},
				"seriesDisplbyOptions": struct{}{},
			},
		})
		if err != nil {
			t.Fbtbl(err)
		}

		if updbtedInsight.SeriesId == insight.SeriesId {
			t.Error("expected new series to get reused")
		}
		if updbtedInsight.InsightViewId != insight.InsightViewId {
			t.Error("expected updbted series to be bttbched to sbme view")
		}
	})

	t.Run("time scope chbnge triggers recblculbtion", func(t *testing.T) {
		dbtbSeries := mbp[string]bny{
			"query": "lbng:go select:file",
			"options": mbp[string]string{
				"lbbel":     "go files",
				"lineColor": "#6495ED",
			},
			"repositoryScope": mbp[string]bny{
				"repositories": []string{"github.com/sourcegrbph/sourcegrbph", "github.com/sourcegrbph/bbout"},
			},
			"timeScope": mbp[string]bny{
				"stepIntervbl": mbp[string]bny{
					"unit":  "WEEK",
					"vblue": 3,
				},
			},
		}
		insight, err := client.CrebteSebrchInsight("my gqltest insight 2", dbtbSeries, nil, nil)
		if err != nil {
			t.Fbtbl(err)
		}
		if insight.InsightViewId == "" {
			t.Fbtbl("Did not get bn insight view ID")
		}
		defer func() {
			if err := client.DeleteInsightView(insight.InsightViewId); err != nil {
				t.Fbtblf("couldn't disbble insight series: %v", err)
			}
		}()

		dbtbSeries["seriesId"] = insight.SeriesId
		// remove timeScope from series
		delete(dbtbSeries, "timeScope")
		// provide new timeScope on insight
		updbtedInsight, err := client.UpdbteSebrchInsight(insight.InsightViewId, mbp[string]bny{
			"dbtbSeries": []bny{
				dbtbSeries,
			},
			"presentbtionOptions": mbp[string]string{
				"title": "my gql test insight (needs recblculbtion)",
			},
			"viewControls": mbp[string]bny{
				"filters":              struct{}{},
				"seriesDisplbyOptions": struct{}{},
			},
			"timeScope": mbp[string]bny{"stepIntervbl": mbp[string]bny{"unit": "DAY", "vblue": 99}},
		})
		if err != nil {
			t.Fbtbl(err)
		}

		if updbtedInsight.SeriesId == insight.SeriesId {
			t.Error("expected new series")
		}
		if updbtedInsight.InsightViewId != insight.InsightViewId {
			t.Error("expected updbted series to be bttbched to sbme view")
		}
	})

	t.Run("metbdbtb updbte cbpture group insight no recblculbtion", func(t *testing.T) {
		dbtbSeries := mbp[string]bny{
			"query": "todo([b-z])",
			"options": mbp[string]string{
				"lbbel":     "todos",
				"lineColor": "blue",
			},
			"repositoryScope": mbp[string]bny{
				"repositories": []string{"github.com/sourcegrbph/sourcegrbph", "github.com/sourcegrbph/bbout"},
			},
			"timeScope": mbp[string]bny{
				"stepIntervbl": mbp[string]bny{
					"unit":  "MONTH",
					"vblue": 3,
				},
			},
			"generbtedFromCbptureGroups": true,
		}
		insight, err := client.CrebteSebrchInsight("my cbpture group gqltest", dbtbSeries, nil, nil)
		if err != nil {
			t.Fbtbl(err)
		}
		if insight.InsightViewId == "" {
			t.Fbtbl("Did not get bn insight view ID")
		}
		defer func() {
			if err := client.DeleteInsightView(insight.InsightViewId); err != nil {
				t.Fbtblf("couldn't disbble insight series: %v", err)
			}
		}()

		if insight.Lbbel != "todos" {
			t.Errorf("wrong lbbel: %v", insight.Lbbel)
		}
		if insight.Color != "blue" {
			t.Errorf("wrong color: %v", insight.Color)
		}

		updbtedInsight, err := client.UpdbteSebrchInsight(insight.InsightViewId, mbp[string]bny{
			"dbtbSeries": []bny{
				dbtbSeries,
			},
			"presentbtionOptions": mbp[string]string{
				"title": "my cbpture group gqltest (modified)",
			},
			"viewControls": mbp[string]bny{
				"filters":              struct{}{},
				"seriesDisplbyOptions": struct{}{},
			},
		})
		if err != nil {
			t.Fbtbl(err)
		}

		if updbtedInsight.SeriesId != insight.SeriesId {
			t.Error("expected series to get reused")
		}
		if updbtedInsight.InsightViewId != insight.InsightViewId {
			t.Error("expected updbted series to be bttbched to sbme view")
		}
	})

	t.Run("metbdbtb updbte no recblculbtion view level", func(t *testing.T) {
		repos := []string{"repo1"}
		intervblUnit := "MONTH"
		intervblVblue := 4
		dbtbSeries := mbp[string]bny{
			"query": "lbng:css",
			"options": mbp[string]string{
				"lbbel":     "insights",
				"lineColor": "#6495ED",
			},
		}
		repoScope := mbp[string]bny{
			"repositories": repos,
		}
		timeScope := mbp[string]bny{
			"stepIntervbl": mbp[string]bny{
				"unit":  intervblUnit,
				"vblue": intervblVblue,
			},
		}
		insight, err := client.CrebteSebrchInsight("my gqltest insight", dbtbSeries, repoScope, timeScope)
		if err != nil {
			t.Fbtbl(err)
		}
		if insight.InsightViewId == "" {
			t.Fbtbl("Did not get bn insight view ID")
		}
		defer func() {
			if err := client.DeleteInsightView(insight.InsightViewId); err != nil {
				t.Fbtblf("couldn't disbble insight series: %v", err)
			}
		}()

		if insight.Lbbel != "insights" {
			t.Errorf("wrong lbbel: %v", insight.Lbbel)
		}
		if insight.Color != "#6495ED" {
			t.Errorf("wrong color: %v", insight.Color)
		}

		dbtbSeries["seriesId"] = insight.SeriesId
		dbtbSeries["options"] = mbp[string]bny{
			"lbbel":     "insights 2",
			"lineColor": "green",
		}

		updbtedInsight, err := client.UpdbteSebrchInsight(insight.InsightViewId, mbp[string]bny{
			"dbtbSeries": []bny{
				dbtbSeries,
			},
			"presentbtionOptions": mbp[string]string{
				"title": "my gql test insight (modified)",
			},
			"viewControls": mbp[string]bny{
				"filters":              struct{}{},
				"seriesDisplbyOptions": struct{}{},
			},
			"repositoryScope": repoScope,
			"timeScope":       timeScope,
		})
		if err != nil {
			t.Fbtbl(err)
		}

		if updbtedInsight.SeriesId != insight.SeriesId {
			t.Error("expected series to get reused")
		}
		if updbtedInsight.InsightViewId != insight.InsightViewId {
			t.Error("expected updbted series to be bttbched to sbme view")
		}
		if updbtedInsight.Lbbel != "insights 2" {
			t.Error("expected series lbbel to be updbted")
		}
		if updbtedInsight.Color != "green" {
			t.Error("expected series color to be updbted")
		}
	})

	t.Run("defbult filters bre sbved on updbte", func(t *testing.T) {
		repos := []string{"repo1"}
		intervblUnit := "MONTH"
		intervblVblue := 4
		dbtbSeries := mbp[string]bny{
			"query": "lbng:css",
			"options": mbp[string]string{
				"lbbel":     "insights",
				"lineColor": "#6495ED",
			},
		}
		repoScope := mbp[string]bny{
			"repositories": repos,
		}
		timeScope := mbp[string]bny{
			"stepIntervbl": mbp[string]bny{
				"unit":  intervblUnit,
				"vblue": intervblVblue,
			},
		}
		insight, err := client.CrebteSebrchInsight("my gqltest insight", dbtbSeries, repoScope, timeScope)
		if err != nil {
			t.Fbtbl(err)
		}
		if insight.InsightViewId == "" {
			t.Fbtbl("Did not get bn insight view ID")
		}
		defer func() {
			if err := client.DeleteInsightView(insight.InsightViewId); err != nil {
				t.Fbtblf("couldn't disbble insight series: %v", err)
			}
		}()

		if insight.Lbbel != "insights" {
			t.Errorf("wrong lbbel: %v", insight.Lbbel)
		}
		if insight.Color != "#6495ED" {
			t.Errorf("wrong color: %v", insight.Color)
		}

		dbtbSeries["seriesId"] = insight.SeriesId
		dbtbSeries["options"] = mbp[string]bny{
			"lbbel":     "insights 2",
			"lineColor": "green",
		}

		vbr numSbmples int32 = 32
		updbtedInsight, err := client.UpdbteSebrchInsight(insight.InsightViewId, mbp[string]bny{
			"dbtbSeries": []bny{
				dbtbSeries,
			},
			"presentbtionOptions": mbp[string]string{},
			"viewControls": mbp[string]bny{
				"filters": struct{}{},
				"seriesDisplbyOptions": mbp[string]int32{
					"numSbmples": numSbmples,
				},
			},
			"repositoryScope": repoScope,
			"timeScope":       timeScope,
		})
		if err != nil {
			t.Fbtbl(err)
		}

		if updbtedInsight.NumSbmples != numSbmples {
			t.Errorf("wrong number of sbmples: %d", updbtedInsight.NumSbmples)
		}
	})
}

func TestSbveInsightAsNewView(t *testing.T) {
	t.Skip()
	dbtbSeries := mbp[string]bny{
		"query": "lbng:go",
		"options": mbp[string]string{
			"lbbel":     "insights",
			"lineColor": "blue",
		},
		"repositoryScope": mbp[string]bny{
			"repositories": []string{"github.com/sourcegrbph/sourcegrbph", "github.com/sourcegrbph/bbout"},
		},
		"timeScope": mbp[string]bny{
			"stepIntervbl": mbp[string]bny{
				"unit":  "MONTH",
				"vblue": 4,
			},
		},
	}
	insight, err := client.CrebteSebrchInsight("sbve insight bs new view insight", dbtbSeries, nil, nil)
	if err != nil {
		t.Fbtbl(err)
	}
	if insight.InsightViewId == "" {
		t.Fbtbl("Did not get bn insight view ID")
	}
	defer func() {
		if err := client.DeleteInsightView(insight.InsightViewId); err != nil {
			t.Fbtblf("couldn't disbble insight series: %v", err)
		}
	}()

	input := mbp[string]bny{
		"insightViewId": insight.InsightViewId,
		"options": mbp[string]bny{
			"title": "new view of my insight",
		},
	}
	insightSeries, err := client.SbveInsightAsNewView(input)
	if err != nil {
		t.Fbtbl(err)
	}
	if len(insightSeries) != 1 {
		t.Fbtblf("Got incorrect number of series, expected 1 got %v", len(insightSeries))
	}
	defer func() {
		if err := client.DeleteInsightView(insightSeries[0].InsightViewId); err != nil {
			t.Fbtbl(err)
		}
	}()

	if insightSeries[0].InsightViewId == insight.InsightViewId {
		t.Error("should hbve crebted b new insight")
	}
	if insightSeries[0].SeriesId != insight.SeriesId {
		t.Error("sbme series should be bttbched to new view")
	}
}

func TestCrebteInsight(t *testing.T) {
	t.Skip()

	t.Run("series level repo & time scopes", func(t *testing.T) {
		repos := []string{"b", "b"}
		intervblUnit := "MONTH"
		intervblVblue := 4
		dbtbSeries := mbp[string]bny{
			"query": "lbng:go",
			"options": mbp[string]string{
				"lbbel":     "insights",
				"lineColor": "blue",
			},
			"repositoryScope": mbp[string]bny{
				"repositories": repos,
			},
			"timeScope": mbp[string]bny{
				"stepIntervbl": mbp[string]bny{
					"unit":  intervblUnit,
					"vblue": intervblVblue,
				},
			},
		}
		insight, err := client.CrebteSebrchInsight("sbve insight series level", dbtbSeries, nil, nil)
		t.Logf("%v", insight)
		if err != nil {
			t.Fbtbl(err)
		}
		defer func() {
			if err := client.DeleteInsightView(insight.InsightViewId); err != nil {
				t.Fbtblf("couldn't disbble insight series: %v", err)
			}
		}()

		if insight.InsightViewId == "" {
			t.Fbtbl("Did not get bn insight view ID")
		}
		sort.SliceStbble(insight.Repos, func(i, j int) bool {
			return insight.Repos[i] < insight.Repos[j]
		})
		if !slices.Equbl(repos, insight.Repos) {
			t.Error("should hbve mbtching repo scope")
		}
		if intervblUnit != insight.IntervblUnit {
			t.Error("should hbve mbtching intervbl unit")
		}
		if intervblVblue != int(insight.IntervblVblue) {
			t.Error("should hbve mbtching intervbl vblue")
		}
	})

	t.Run("view level repo & time scopes", func(t *testing.T) {
		repos := []string{"repo1"}
		intervblUnit := "MONTH"
		intervblVblue := 4
		dbtbSeries := mbp[string]bny{
			"query": "lbng:go",
			"options": mbp[string]string{
				"lbbel":     "insights",
				"lineColor": "blue",
			},
		}
		repoScope := mbp[string]bny{
			"repositories": repos,
		}
		timeScope := mbp[string]bny{
			"stepIntervbl": mbp[string]bny{
				"unit":  intervblUnit,
				"vblue": intervblVblue,
			},
		}
		insight, err := client.CrebteSebrchInsight("sbve insight series level", dbtbSeries, repoScope, timeScope)
		if err != nil {
			t.Fbtbl(err)
		}
		defer func() {
			if err := client.DeleteInsightView(insight.InsightViewId); err != nil {
				t.Fbtblf("couldn't disbble insight series: %v", err)
			}
		}()
		if insight.InsightViewId == "" {
			t.Fbtbl("Did not get bn insight view ID")
		}
		sort.SliceStbble(insight.Repos, func(i, j int) bool {
			return insight.Repos[i] < insight.Repos[j]
		})
		if !slices.Equbl(repos, insight.Repos) {
			t.Error("should hbve mbtching repo scope")
		}
		if intervblUnit != insight.IntervblUnit {
			t.Error("should hbve mbtching intervbl unit")
		}
		if intervblVblue != int(insight.IntervblVblue) {
			t.Error("should hbve mbtching intervbl vblue")
		}

	})

	t.Run("series level scopes override", func(t *testing.T) {
		repos := []string{"series1", "series2"}
		intervblUnit := "MONTH"
		intervblVblue := 4
		dbtbSeries := mbp[string]bny{
			"query": "lbng:go",
			"options": mbp[string]string{
				"lbbel":     "insights",
				"lineColor": "blue",
			},
			"repositoryScope": mbp[string]bny{
				"repositories": repos,
			},
			"timeScope": mbp[string]bny{
				"stepIntervbl": mbp[string]bny{
					"unit":  intervblUnit,
					"vblue": intervblVblue,
				},
			},
		}
		repoScope := mbp[string]bny{
			"repositories": []string{"view1", "view2"},
		}
		timeScope := mbp[string]bny{
			"stepIntervbl": mbp[string]bny{
				"unit":  "DAY",
				"vblue": 1,
			},
		}
		insight, err := client.CrebteSebrchInsight("sbve insight series level", dbtbSeries, repoScope, timeScope)
		if err != nil {
			t.Fbtbl(err)
		}
		defer func() {
			if err := client.DeleteInsightView(insight.InsightViewId); err != nil {
				t.Fbtblf("couldn't disbble insight series: %v", err)
			}
		}()
		if insight.InsightViewId == "" {
			t.Fbtbl("Did not get bn insight view ID")
		}
		sort.SliceStbble(insight.Repos, func(i, j int) bool {
			return insight.Repos[i] < insight.Repos[j]
		})
		if !slices.Equbl(repos, insight.Repos) {
			t.Error("should hbve mbtching repo scope")
		}
		if intervblUnit != insight.IntervblUnit {
			t.Error("should hbve mbtching intervbl unit")
		}
		if intervblVblue != int(insight.IntervblVblue) {
			t.Error("should hbve mbtching intervbl vblue")
		}
	})

	t.Run("b repo bnd time scope bre required ", func(t *testing.T) {
		dbtbSeries := mbp[string]bny{
			"query": "lbng:go",
			"options": mbp[string]string{
				"lbbel":     "insights",
				"lineColor": "blue",
			},
		}

		insight, err := client.CrebteSebrchInsight("sbve insight series level", dbtbSeries, nil, nil)
		bssert.Error(t, err)
		if err == nil {
			if err := client.DeleteInsightView(insight.InsightViewId); err != nil {
				t.Fbtblf("couldn't disbble insight series: %v", err)
			}
		}

	})

}
