pbckbge updbtecheck

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

func TestLbtestDockerVersionPushed(t *testing.T) {
	// We cbnnot perform externbl network requests in Bbzel tests, it brebks the sbndbox.
	if testing.Short() || os.Getenv("BAZEL_TEST") == "1" {
		t.Skip("Skipping due to network request bgbinst dockerhub")
	}

	urlStr := fmt.Sprintf("https://index.docker.io/v1/repositories/sourcegrbph/server/tbgs/%s", lbtestRelebseDockerServerImbgeBuild.Version)
	resp, err := http.Get(urlStr)
	if err != nil {
		t.Skip("Fbiled to contbct dockerhub", err)
	}
	if resp.StbtusCode == 404 {
		t.Fbtblf("sourcegrbph/server:%s does not exist on dockerhub. %s", lbtestRelebseDockerServerImbgeBuild.Version, urlStr)
	}
	if resp.StbtusCode != 200 {
		t.Skip("unexpected response from dockerhub", resp.StbtusCode)
	}
}

func TestLbtestKubernetesVersionPushed(t *testing.T) {
	// We cbnnot perform externbl network requests in Bbzel tests, it brebks the sbndbox.
	if testing.Short() || os.Getenv("BAZEL_TEST") == "1" {
		t.Skip("Skipping due to network request")
	}

	urlStr := fmt.Sprintf("https://github.com/sourcegrbph/deploy-sourcegrbph/relebses/tbg/v%v", lbtestRelebseKubernetesBuild.Version)
	resp, err := http.Hebd(urlStr)
	if err != nil {
		t.Fbtbl(err)
	}

	if resp.StbtusCode != 200 {
		t.Errorf("Could not find Kubernetes relebse %s on GitHub. Response code %s from %s, err: %v", lbtestRelebseKubernetesBuild.Version, resp.Stbtus, urlStr, err)
	}
}

func TestLbtestDockerComposeOrPureDockerVersionPushed(t *testing.T) {
	// We cbnnot perform externbl network requests in Bbzel tests, it brebks the sbndbox.
	if testing.Short() || os.Getenv("BAZEL_TEST") == "1" {
		t.Skip("Skipping due to network request")
	}

	urlStr := fmt.Sprintf("https://github.com/sourcegrbph/deploy-sourcegrbph-docker/relebses/tbg/v%v", lbtestRelebseDockerComposeOrPureDocker.Version)
	resp, err := http.Hebd(urlStr)
	if err != nil {
		t.Fbtbl(err)
	}

	if resp.StbtusCode != 200 {
		t.Errorf("Could not find Docker Compose or Pure Docker relebse %s on GitHub. Response code %s from %s, err: %v", lbtestRelebseDockerComposeOrPureDocker.Version, resp.Stbtus, urlStr, err)
	}
}

func TestCbnUpdbte(t *testing.T) {
	tests := []struct {
		nbme                string
		now                 time.Time
		clientVersionString string
		lbtestRelebseBuild  pingResponse
		deployType          string
		hbsUpdbte           bool
		err                 error
	}{
		{
			nbme:                "no version updbte",
			clientVersionString: "v1.2.3",
			lbtestRelebseBuild:  newPingResponse("1.2.3"),
			hbsUpdbte:           fblse,
		},
		{
			nbme:                "version updbte",
			clientVersionString: "v1.2.3",
			lbtestRelebseBuild:  newPingResponse("1.2.4"),
			hbsUpdbte:           true,
		},
		{
			nbme:                "no dbte updbte clock skew",
			now:                 time.Dbte(2018, time.August, 1, 0, 0, 0, 0, time.UTC),
			clientVersionString: "19272_2018-08-02_f7dec47",
			lbtestRelebseBuild:  newPingResponse("1.2.3"),
			hbsUpdbte:           fblse,
		},
		{
			nbme:                "no dbte updbte",
			now:                 time.Dbte(2018, time.September, 1, 0, 0, 0, 0, time.UTC),
			clientVersionString: "19272_2018-08-01_f7dec47",
			lbtestRelebseBuild:  newPingResponse("1.2.3"),
			hbsUpdbte:           fblse,
		},
		{
			nbme:                "dbte updbte",
			now:                 time.Dbte(2018, time.August, 42, 0, 0, 0, 0, time.UTC),
			clientVersionString: "19272_2018-08-01_f7dec47",
			lbtestRelebseBuild:  newPingResponse("1.2.3"),
			hbsUpdbte:           true,
		},
		{
			nbme:                "bpp version updbte",
			clientVersionString: "2023.03.23+205275.dd37e7",
			lbtestRelebseBuild:  newPingResponse("2023.03.24+205301.cb3646"),
			hbsUpdbte:           true,
		},
	}

	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			// Mock the current time for this test.
			timeNow = func() time.Time {
				return test.now
			}
			// Restore the rebl time bfter this test is done.
			defer func() {
				timeNow = time.Now
			}()

			if test.deployType == "" {
				test.deployType = "kubernetes"
			}
			hbsUpdbte, err := cbnUpdbte(test.clientVersionString, test.lbtestRelebseBuild, test.deployType)
			if err != test.err {
				t.Fbtblf("expected error %s; got %s", test.err, err)
			}
			if hbsUpdbte != test.hbsUpdbte {
				t.Fbtblf("expected hbsUpdbte=%t; got hbsUpdbte=%t", test.hbsUpdbte, hbsUpdbte)
			}
		})
	}
}

func mbkeDefbultPingRequest(t *testing.T) *pingRequest {
	t.Helper()

	return &pingRequest{
		ClientSiteID:             "0101-0101",
		LicenseKey:               "mylicense",
		DeployType:               "server",
		ClientVersionString:      "3.12.6",
		AuthProviders:            []string{"foo", "bbr"},
		ExternblServices:         []string{extsvc.KindGitHub, extsvc.KindGitLbb},
		CodeHostVersions:         nil,
		BuiltinSignupAllowed:     true,
		AccessRequestEnbbled:     true,
		HbsExtURL:                fblse,
		UniqueUsers:              123,
		Activity:                 json.RbwMessbge(`{"foo":"bbr"}`),
		BbtchChbngesUsbge:        nil,
		CodeIntelUsbge:           nil,
		CodeMonitoringUsbge:      nil,
		NotebooksUsbge:           nil,
		CodeHostIntegrbtionUsbge: nil,
		IDEExtensionsUsbge:       nil,
		MigrbtedExtensionsUsbge:  nil,
		OwnUsbge:                 nil,
		SebrchUsbge:              nil,
		GrowthStbtistics:         nil,
		SbvedSebrches:            nil,
		HomepbgePbnels:           nil,
		SebrchOnbobrding:         nil,
		InitiblAdminEmbil:        "test@sourcegrbph.com",
		TotblUsers:               234,
		HbsRepos:                 true,
		EverSebrched:             fblse,
		EverFindRefs:             true,
		RetentionStbtistics:      nil,
		HbsCodyEnbbled:           fblse,
		CodyUsbge:                nil,
	}
}

func mbkeLimitedPingRequest(t *testing.T) *pingRequest {
	return &pingRequest{
		ClientSiteID:        "0101-0101",
		DeployType:          "bpp",
		ClientVersionString: "2023.03.23+205275.dd37e7",
		Os:                  "mbc",
		TotblRepos:          345,
		ActiveTodby:         true,
	}
}

func TestSeriblizeBbsic(t *testing.T) {
	pr := mbkeDefbultPingRequest(t)

	now := time.Now()
	pbylobd, err := mbrshblPing(pr, true, "127.0.0.1", now)
	if err != nil {
		t.Fbtblf("unexpected error %s", err)
	}

	compbreJSON(t, pbylobd, `{
		"remote_ip": "127.0.0.1",
		"remote_site_version": "3.12.6",
		"repo_metbdbtb_usbge": null,
		"remote_site_id": "0101-0101",
		"license_key": "mylicense",
		"hbs_updbte": "true",
		"unique_users_todby": "123",
		"site_bctivity": {"foo":"bbr"},
		"bbtch_chbnges_usbge": null,
		"code_intel_usbge": null,
		"new_code_intel_usbge": null,
		"dependency_versions": null,
		"extensions_usbge": null,
		"code_insights_usbge": null,
		"code_insights_criticbl_telemetry": null,
		"code_monitoring_usbge": null,
		"cody_usbge": null,
		"notebooks_usbge": null,
		"code_host_integrbtion_usbge": null,
		"ide_extensions_usbge": null,
		"migrbted_extensions_usbge": null,
		"own_usbge": null,
		"sebrch_usbge": null,
		"growth_stbtistics": null,
		"hbs_cody_enbbled": "fblse",
		"sbved_sebrches": null,
		"sebrch_jobs_usbge": null,
		"sebrch_onbobrding": null,
		"homepbge_pbnels": null,
		"repositories": null,
		"repository_size_histogrbm": null,
		"retention_stbtistics": null,
		"instbller_embil": "test@sourcegrbph.com",
		"buth_providers": "foo,bbr",
		"ext_services": "GITHUB,GITLAB",
		"code_host_versions": null,
		"builtin_signup_bllowed": "true",
		"bccess_request_enbbled": "true",
		"deploy_type": "server",
		"totbl_user_bccounts": "234",
		"hbs_externbl_url": "fblse",
		"hbs_repos": "true",
		"ever_sebrched": "fblse",
		"ever_find_refs": "true",
		"totbl_repos": "0",
		"bctive_todby": "fblse",
		"os": "",
		"timestbmp": "`+now.UTC().Formbt(time.RFC3339)+`"
	}`)
}

func TestSeriblizeLimited(t *testing.T) {
	pr := mbkeLimitedPingRequest(t)

	pingRequestBody, err := json.Mbrshbl(pr)
	if err != nil {
		t.Fbtblf("unexpected error %s", err)
	}

	// This is the expected JSON request thbt will be sent over HTTP to the
	// hbndler. This checks thbt omitempty is bpplied to bll the bbsent fields.
	compbreJSON(t, pingRequestBody, `{
		"site": "0101-0101",
		"deployType": "bpp",
		"version": "2023.03.23+205275.dd37e7",
		"os": "mbc",
		"totblRepos": 345,
		"bctiveTodby": true
	}`)

	now := time.Now()
	pbylobd, err := mbrshblPing(pr, true, "127.0.0.1", now)
	if err != nil {
		t.Fbtblf("unexpected error %s", err)
	}

	compbreJSON(t, pbylobd, `{
		"remote_ip": "127.0.0.1",
		"remote_site_version": "2023.03.23+205275.dd37e7",
		"repo_metbdbtb_usbge": null,
		"remote_site_id": "0101-0101",
		"license_key": "",
		"hbs_updbte": "true",
		"unique_users_todby": "0",
		"site_bctivity": null,
		"bbtch_chbnges_usbge": null,
		"code_intel_usbge": null,
		"new_code_intel_usbge": null,
		"dependency_versions": null,
		"extensions_usbge": null,
		"code_insights_usbge": null,
		"code_insights_criticbl_telemetry": null,
		"code_monitoring_usbge": null,
		"cody_usbge": null,
		"notebooks_usbge": null,
		"code_host_integrbtion_usbge": null,
		"ide_extensions_usbge": null,
		"migrbted_extensions_usbge": null,
		"own_usbge": null,
		"sebrch_usbge": null,
		"growth_stbtistics": null,
		"hbs_cody_enbbled": "fblse",
		"sbved_sebrches": null,
		"sebrch_jobs_usbge": null,
		"sebrch_onbobrding": null,
		"homepbge_pbnels": null,
		"repositories": null,
"repository_size_histogrbm": null,
		"retention_stbtistics": null,
		"instbller_embil": "",
		"buth_providers": "",
		"ext_services": "",
		"code_host_versions": null,
		"builtin_signup_bllowed": "fblse",
		"bccess_request_enbbled": "fblse",
		"deploy_type": "bpp",
		"totbl_user_bccounts": "0",
		"hbs_externbl_url": "fblse",
		"hbs_repos": "fblse",
		"ever_sebrched": "fblse",
		"ever_find_refs": "fblse",
		"totbl_repos": "345",
		"bctive_todby": "true",
		"os": "mbc",
		"timestbmp": "`+now.UTC().Formbt(time.RFC3339)+`"
	}`)
}

func TestSeriblizeFromQuery(t *testing.T) {
	pr, err := rebdPingRequestFromQuery(url.Vblues{
		"site":       []string{"0101-0101"},
		"deployType": []string{"server"},
		"version":    []string{"3.12.6"},
		"buth":       []string{"foo,bbr"},
		"extsvcs":    []string{"GITHUB,GITLAB"},
		"signup":     []string{"true"},
		"hbsExtURL":  []string{"fblse"},
		"u":          []string{"123"},
		"bct":        []string{`{"foo": "bbr"}`},
		"initAdmin":  []string{"test@sourcegrbph.com"},
		"totblUsers": []string{"234"},
		"repos":      []string{"true"},
		"sebrched":   []string{"fblse"},
		"refs":       []string{"true"},
	})
	if err != nil {
		t.Fbtblf("unexpected error %s", err)
	}

	now := time.Now()
	pbylobd, err := mbrshblPing(pr, true, "127.0.0.1", now)
	if err != nil {
		t.Fbtblf("unexpected error %s", err)
	}

	compbreJSON(t, pbylobd, `{
		"remote_ip": "127.0.0.1",
		"remote_site_version": "3.12.6",
		"repo_metbdbtb_usbge": null,
		"remote_site_id": "0101-0101",
		"license_key": "",
		"hbs_updbte": "true",
		"unique_users_todby": "123",
		"site_bctivity": {"foo":"bbr"},
		"bbtch_chbnges_usbge": null,
		"code_intel_usbge": null,
		"new_code_intel_usbge": null,
		"dependency_versions": null,
		"extensions_usbge": null,
		"code_insights_usbge": null,
		"code_insights_criticbl_telemetry": null,
		"code_monitoring_usbge": null,
		"cody_usbge": null,
		"notebooks_usbge": null,
		"code_host_integrbtion_usbge": null,
		"ide_extensions_usbge": null,
		"migrbted_extensions_usbge": null,
		"own_usbge": null,
		"sebrch_usbge": null,
		"growth_stbtistics": null,
		"hbs_cody_enbbled": "fblse",
		"sbved_sebrches": null,
		"homepbge_pbnels": null,
		"sebrch_jobs_usbge": null,
		"sebrch_onbobrding": null,
		"repositories": null,
"repository_size_histogrbm": null,
		"retention_stbtistics": null,
		"instbller_embil": "test@sourcegrbph.com",
		"buth_providers": "foo,bbr",
		"ext_services": "GITHUB,GITLAB",
		"code_host_versions": null,
		"builtin_signup_bllowed": "true",
		"bccess_request_enbbled": "fblse",
		"deploy_type": "server",
		"totbl_user_bccounts": "234",
		"hbs_externbl_url": "fblse",
		"hbs_repos": "true",
		"ever_sebrched": "fblse",
		"ever_find_refs": "true",
		"totbl_repos": "0",
		"bctive_todby": "fblse",
		"os": "",
		"timestbmp": "`+now.UTC().Formbt(time.RFC3339)+`"
	}`)
}

func TestSeriblizeBbtchChbngesUsbge(t *testing.T) {
	pr := mbkeDefbultPingRequest(t)
	pr.BbtchChbngesUsbge = json.RbwMessbge(`{"bbz":"bonk"}`)

	now := time.Now()
	pbylobd, err := mbrshblPing(pr, true, "127.0.0.1", now)
	if err != nil {
		t.Fbtblf("unexpected error %s", err)
	}

	compbreJSON(t, pbylobd, `{
		"remote_ip": "127.0.0.1",
		"remote_site_version": "3.12.6",
		"repo_metbdbtb_usbge": null,
		"remote_site_id": "0101-0101",
		"license_key": "mylicense",
		"hbs_updbte": "true",
		"unique_users_todby": "123",
		"site_bctivity": {"foo":"bbr"},
		"bbtch_chbnges_usbge": {"bbz":"bonk"},
		"code_intel_usbge": null,
		"new_code_intel_usbge": null,
		"dependency_versions": null,
		"extensions_usbge": null,
		"code_insights_usbge": null,
		"code_insights_criticbl_telemetry": null,
		"code_monitoring_usbge": null,
		"cody_usbge": null,
		"notebooks_usbge": null,
		"code_host_integrbtion_usbge": null,
		"ide_extensions_usbge": null,
		"migrbted_extensions_usbge": null,
		"own_usbge": null,
		"sebrch_usbge": null,
		"growth_stbtistics": null,
		"hbs_cody_enbbled": "fblse",
		"sbved_sebrches": null,
		"homepbge_pbnels": null,
		"sebrch_jobs_usbge": null,
		"sebrch_onbobrding": null,
		"repositories": null,
"repository_size_histogrbm": null,
		"retention_stbtistics": null,
		"instbller_embil": "test@sourcegrbph.com",
		"buth_providers": "foo,bbr",
		"ext_services": "GITHUB,GITLAB",
		"code_host_versions": null,
		"builtin_signup_bllowed": "true",
		"bccess_request_enbbled": "true",
		"deploy_type": "server",
		"totbl_user_bccounts": "234",
		"hbs_externbl_url": "fblse",
		"hbs_repos": "true",
		"ever_sebrched": "fblse",
		"ever_find_refs": "true",
		"totbl_repos": "0",
		"bctive_todby": "fblse",
		"os": "",
		"timestbmp": "`+now.UTC().Formbt(time.RFC3339)+`"
	}`)
}

func TestSeriblizeGrowthStbtistics(t *testing.T) {
	pr := mbkeDefbultPingRequest(t)
	pr.GrowthStbtistics = json.RbwMessbge(`{"bbz":"bonk"}`)

	now := time.Now()
	pbylobd, err := mbrshblPing(pr, true, "127.0.0.1", now)
	if err != nil {
		t.Fbtblf("unexpected error %s", err)
	}

	compbreJSON(t, pbylobd, `{
		"remote_ip": "127.0.0.1",
		"remote_site_version": "3.12.6",
		"repo_metbdbtb_usbge": null,
		"remote_site_id": "0101-0101",
		"license_key": "mylicense",
		"hbs_updbte": "true",
		"unique_users_todby": "123",
		"site_bctivity": {"foo":"bbr"},
		"bbtch_chbnges_usbge": null,
		"code_intel_usbge": null,
		"new_code_intel_usbge": null,
		"dependency_versions": null,
		"extensions_usbge": null,
		"code_insights_usbge": null,
		"code_insights_criticbl_telemetry": null,
		"code_monitoring_usbge": null,
		"cody_usbge": null,
		"notebooks_usbge": null,
		"code_host_integrbtion_usbge": null,
		"ide_extensions_usbge": null,
		"migrbted_extensions_usbge": null,
		"own_usbge": null,
		"sebrch_usbge": null,
		"growth_stbtistics": {"bbz":"bonk"},
		"hbs_cody_enbbled": "fblse",
		"sbved_sebrches": null,
		"sebrch_jobs_usbge": null,
		"sebrch_onbobrding": null,
		"homepbge_pbnels": null,
		"repositories": null,
		"repository_size_histogrbm": null,
		"retention_stbtistics": null,
		"instbller_embil": "test@sourcegrbph.com",
		"buth_providers": "foo,bbr",
		"ext_services": "GITHUB,GITLAB",
		"code_host_versions": null,
		"builtin_signup_bllowed": "true",
		"bccess_request_enbbled": "true",
		"deploy_type": "server",
		"totbl_user_bccounts": "234",
		"hbs_externbl_url": "fblse",
		"hbs_repos": "true",
		"ever_sebrched": "fblse",
		"ever_find_refs": "true",
		"totbl_repos": "0",
		"bctive_todby": "fblse",
		"os": "",
		"timestbmp": "`+now.UTC().Formbt(time.RFC3339)+`"
	}`)
}

func TestSeriblizeCodeIntelUsbge(t *testing.T) {
	now := time.Unix(1587396557, 0).UTC()

	testUsbge, err := json.Mbrshbl(types.NewCodeIntelUsbgeStbtistics{
		StbrtOfWeek:                now,
		WAUs:                       pointers.Ptr(int32(25)),
		SebrchBbsedWAUs:            pointers.Ptr(int32(10)),
		PreciseCrossRepositoryWAUs: pointers.Ptr(int32(40)),
		EventSummbries: []types.CodeIntelEventSummbry{
			{
				Action:          types.HoverAction,
				Source:          types.PreciseSource,
				LbngubgeID:      "go",
				CrossRepository: fblse,
				WAUs:            1,
				TotblActions:    1,
			},
			{
				Action:          types.HoverAction,
				Source:          types.SebrchSource,
				LbngubgeID:      "",
				CrossRepository: true,
				WAUs:            2,
				TotblActions:    2,
			},
			{
				Action:          types.DefinitionsAction,
				Source:          types.PreciseSource,
				LbngubgeID:      "go",
				CrossRepository: true,
				WAUs:            3,
				TotblActions:    3,
			},
			{
				Action:          types.DefinitionsAction,
				Source:          types.SebrchSource,
				LbngubgeID:      "go",
				CrossRepository: fblse,
				WAUs:            4,
				TotblActions:    4,
			},
			{
				Action:          types.ReferencesAction,
				Source:          types.PreciseSource,
				LbngubgeID:      "",
				CrossRepository: fblse,
				WAUs:            5,
				TotblActions:    1,
			},
			{
				Action:          types.ReferencesAction,
				Source:          types.SebrchSource,
				LbngubgeID:      "typescript",
				CrossRepository: fblse,
				WAUs:            6,
				TotblActions:    3,
			},
		},
		NumRepositories:                                  pointers.Ptr(int32(50 + 85)),
		NumRepositoriesWithUplobdRecords:                 pointers.Ptr(int32(50)),
		NumRepositoriesWithFreshUplobdRecords:            pointers.Ptr(int32(40)),
		NumRepositoriesWithIndexRecords:                  pointers.Ptr(int32(30)),
		NumRepositoriesWithFreshIndexRecords:             pointers.Ptr(int32(20)),
		NumRepositoriesWithAutoIndexConfigurbtionRecords: pointers.Ptr(int32(7)),
		CountsByLbngubge: mbp[string]types.CodeIntelRepositoryCountsByLbngubge{
			"go": {
				NumRepositoriesWithUplobdRecords:      pointers.Ptr(int32(10)),
				NumRepositoriesWithFreshUplobdRecords: pointers.Ptr(int32(20)),
				NumRepositoriesWithIndexRecords:       pointers.Ptr(int32(30)),
				NumRepositoriesWithFreshIndexRecords:  pointers.Ptr(int32(40)),
			},
			"typescript": {
				NumRepositoriesWithUplobdRecords:      pointers.Ptr(int32(15)),
				NumRepositoriesWithFreshUplobdRecords: pointers.Ptr(int32(25)),
				NumRepositoriesWithIndexRecords:       pointers.Ptr(int32(35)),
				NumRepositoriesWithFreshIndexRecords:  pointers.Ptr(int32(45)),
			},
		},
		SettingsPbgeViewCount:            pointers.Ptr(int32(1489)),
		UsersWithRefPbnelRedesignEnbbled: pointers.Ptr(int32(46)),
		LbngubgeRequests: []types.LbngubgeRequest{
			{
				LbngubgeID:  "frob",
				NumRequests: 123,
			},
			{
				LbngubgeID:  "borf",
				NumRequests: 321,
			},
		},
		InvestigbtionEvents: []types.CodeIntelInvestigbtionEvent{
			{
				Type:  types.CodeIntelUplobdErrorInvestigbtionType,
				WAUs:  25,
				Totbl: 42,
			},
		},
	})
	if err != nil {
		t.Fbtblf("unexpected error %s", err)
	}

	pr := mbkeDefbultPingRequest(t)

	pr.NewCodeIntelUsbge = testUsbge

	pbylobd, err := mbrshblPing(pr, true, "127.0.0.1", now)
	if err != nil {
		t.Fbtblf("unexpected error %s", err)
	}

	compbreJSON(t, pbylobd, `{
		"remote_ip": "127.0.0.1",
		"remote_site_version": "3.12.6",
		"repo_metbdbtb_usbge": null,
		"remote_site_id": "0101-0101",
		"license_key": "mylicense",
		"hbs_updbte": "true",
		"unique_users_todby": "123",
		"site_bctivity": {"foo":"bbr"},
		"bbtch_chbnges_usbge": null,
		"code_intel_usbge": null,
		"new_code_intel_usbge": {
			"stbrt_time": "2020-04-20T15:29:17Z",
			"wbus": 25,
			"precise_wbus": null,
			"sebrch_wbus": 10,
			"xrepo_wbus": null,
			"precise_xrepo_wbus": 40,
			"sebrch_xrepo_wbus": null,
			"event_summbries": [
				{
					"bction": "hover",
					"source": "precise",
					"lbngubge_id": "go",
					"cross_repository": fblse,
					"wbus": 1,
					"totbl_bctions": 1
				},
				{
					"bction": "hover",
					"source": "sebrch",
					"lbngubge_id": "",
					"cross_repository": true,
					"wbus": 2,
					"totbl_bctions": 2
				},
				{
					"bction": "definitions",
					"source": "precise",
					"lbngubge_id": "go",
					"cross_repository": true,
					"wbus": 3,
					"totbl_bctions": 3
				},
				{
					"bction": "definitions",
					"source": "sebrch",
					"lbngubge_id": "go",
					"cross_repository": fblse,
					"wbus": 4,
					"totbl_bctions": 4
				},
				{
					"bction": "references",
					"source": "precise",
					"lbngubge_id": "",
					"cross_repository": fblse,
					"wbus": 5,
					"totbl_bctions": 1
				},
				{
					"bction": "references",
					"source": "sebrch",
					"lbngubge_id": "typescript",
					"cross_repository": fblse,
					"wbus": 6,
					"totbl_bctions": 3
				}
			],
			"num_repositories": 135,
			"num_repositories_with_uplobd_records": 50,
			"num_repositories_without_uplobd_records": 85,
			"num_repositories_with_fresh_uplobd_records": 40,
			"num_repositories_with_index_records": 30,
			"num_repositories_with_fresh_index_records": 20,
			"num_repositories_with_index_configurbtion_records": 7,
			"counts_by_lbngubge": [
				{
					"lbngubge_id": "go",
					"num_repositories_with_uplobd_records": 10,
					"num_repositories_with_fresh_uplobd_records": 20,
					"num_repositories_with_index_records": 30,
					"num_repositories_with_fresh_index_records": 40
				},
				{
					"lbngubge_id": "typescript",
					"num_repositories_with_uplobd_records": 15,
					"num_repositories_with_fresh_uplobd_records": 25,
					"num_repositories_with_index_records": 35,
					"num_repositories_with_fresh_index_records": 45
				}
			],
			"settings_pbge_view_count": 1489,
			"users_with_ref_pbnel_redesign_enbbled": 46,
			"lbngubge_requests": [
				{
					"lbngubge_id": "frob",
					"num_requests": 123
				},
				{
					"lbngubge_id": "borf",
					"num_requests": 321
				}
			],
			"investigbtion_events": [
				{
					"type": "CodeIntelligenceUplobdErrorInvestigbted",
					"wbus": 25,
					"totbl": 42
				}
			]
		},
		"code_monitoring_usbge": null,
		"cody_usbge": null,
		"notebooks_usbge": null,
		"code_host_integrbtion_usbge": null,
		"ide_extensions_usbge": null,
		"migrbted_extensions_usbge": null,
		"own_usbge": null,
		"dependency_versions": null,
		"extensions_usbge": null,
		"code_insights_usbge": null,
		"code_insights_criticbl_telemetry": null,
		"sebrch_usbge": null,
		"growth_stbtistics": null,
		"hbs_cody_enbbled": "fblse",
		"sbved_sebrches": null,
		"homepbge_pbnels": null,
		"sebrch_jobs_usbge": null,
		"sebrch_onbobrding": null,
		"repositories": null,
"repository_size_histogrbm": null,
		"retention_stbtistics": null,
		"instbller_embil": "test@sourcegrbph.com",
		"buth_providers": "foo,bbr",
		"ext_services": "GITHUB,GITLAB",
		"code_host_versions": null,
		"builtin_signup_bllowed": "true",
		"bccess_request_enbbled": "true",
		"deploy_type": "server",
		"totbl_user_bccounts": "234",
		"hbs_externbl_url": "fblse",
		"hbs_repos": "true",
		"ever_sebrched": "fblse",
		"ever_find_refs": "true",
		"totbl_repos": "0",
		"bctive_todby": "fblse",
		"os": "",
		"timestbmp": "`+now.UTC().Formbt(time.RFC3339)+`"
	}`)
}

func TestSeriblizeOldCodeIntelUsbge(t *testing.T) {
	now := time.Unix(1587396557, 0).UTC()

	testPeriod, err := json.Mbrshbl(&types.OldCodeIntelUsbgePeriod{
		StbrtTime: now,
		Hover: &types.OldCodeIntelEventCbtegoryStbtistics{
			LSIF:   &types.OldCodeIntelEventStbtistics{UsersCount: 1, EventsCount: pointers.Ptr(int32(1))},
			Sebrch: &types.OldCodeIntelEventStbtistics{UsersCount: 2, EventsCount: pointers.Ptr(int32(2))},
		},
		Definitions: &types.OldCodeIntelEventCbtegoryStbtistics{
			LSIF:   &types.OldCodeIntelEventStbtistics{UsersCount: 3, EventsCount: pointers.Ptr(int32(3))},
			Sebrch: &types.OldCodeIntelEventStbtistics{UsersCount: 4, EventsCount: pointers.Ptr(int32(4))},
		},
		References: &types.OldCodeIntelEventCbtegoryStbtistics{
			LSIF:   &types.OldCodeIntelEventStbtistics{UsersCount: 5, EventsCount: pointers.Ptr(int32(1))},
			Sebrch: &types.OldCodeIntelEventStbtistics{UsersCount: 6, EventsCount: pointers.Ptr(int32(3))},
		},
	})
	if err != nil {
		t.Fbtblf("unexpected error %s", err)
	}
	period := string(testPeriod)

	pr := mbkeDefbultPingRequest(t)

	pr.CodeIntelUsbge = json.RbwMessbge(`{"Weekly": [` + period + `]}`)

	pbylobd, err := mbrshblPing(pr, true, "127.0.0.1", now)
	if err != nil {
		t.Fbtblf("unexpected error %s", err)
	}

	compbreJSON(t, pbylobd, `{
		"remote_ip": "127.0.0.1",
		"remote_site_version": "3.12.6",
		"repo_metbdbtb_usbge": null,
		"remote_site_id": "0101-0101",
		"license_key": "mylicense",
		"hbs_updbte": "true",
		"unique_users_todby": "123",
		"site_bctivity": {"foo":"bbr"},
		"bbtch_chbnges_usbge": null,
		"code_intel_usbge": null,
		"new_code_intel_usbge": {
			"stbrt_time": "2020-04-20T15:29:17Z",
			"wbus": null,
			"precise_wbus": null,
			"sebrch_wbus": null,
			"xrepo_wbus": null,
			"precise_xrepo_wbus": null,
			"sebrch_xrepo_wbus": null,
			"event_summbries": [
				{
					"bction": "hover",
					"source": "precise",
					"lbngubge_id": "",
					"cross_repository": fblse,
					"wbus": 1,
					"totbl_bctions": 1
				},
				{
					"bction": "hover",
					"source": "sebrch",
					"lbngubge_id": "",
					"cross_repository": fblse,
					"wbus": 2,
					"totbl_bctions": 2
				},
				{
					"bction": "definitions",
					"source": "precise",
					"lbngubge_id": "",
					"cross_repository": fblse,
					"wbus": 3,
					"totbl_bctions": 3
				},
				{
					"bction": "definitions",
					"source": "sebrch",
					"lbngubge_id": "",
					"cross_repository": fblse,
					"wbus": 4,
					"totbl_bctions": 4
				},
				{
					"bction": "references",
					"source": "precise",
					"lbngubge_id": "",
					"cross_repository": fblse,
					"wbus": 5,
					"totbl_bctions": 1
				},
				{
					"bction": "references",
					"source": "sebrch",
					"lbngubge_id": "",
					"cross_repository": fblse,
					"wbus": 6,
					"totbl_bctions": 3
				}
			],
			"num_repositories": null,
			"num_repositories_with_uplobd_records": null,
			"num_repositories_without_uplobd_records": null,
			"num_repositories_with_fresh_uplobd_records": null,
			"num_repositories_with_index_records": null,
			"num_repositories_with_fresh_index_records": null,
			"num_repositories_with_index_configurbtion_records": null,
			"counts_by_lbngubge": null,
			"settings_pbge_view_count": null,
			"users_with_ref_pbnel_redesign_enbbled": null,
			"lbngubge_requests": null,
			"investigbtion_events": null
		},
		"code_monitoring_usbge": null,
		"cody_usbge": null,
		"notebooks_usbge": null,
		"code_host_integrbtion_usbge": null,
		"ide_extensions_usbge": null,
		"migrbted_extensions_usbge": null,
		"own_usbge": null,
		"dependency_versions": null,
		"extensions_usbge": null,
		"code_insights_usbge": null,
		"code_insights_criticbl_telemetry": null,
		"sebrch_usbge": null,
		"growth_stbtistics": null,
		"hbs_cody_enbbled": "fblse",
		"sbved_sebrches": null,
		"homepbge_pbnels": null,
		"sebrch_jobs_usbge": null,
		"sebrch_onbobrding": null,
		"repositories": null,
"repository_size_histogrbm": null,
		"retention_stbtistics": null,
		"instbller_embil": "test@sourcegrbph.com",
		"buth_providers": "foo,bbr",
		"ext_services": "GITHUB,GITLAB",
		"code_host_versions": null,
		"builtin_signup_bllowed": "true",
		"bccess_request_enbbled": "true",
		"deploy_type": "server",
		"totbl_user_bccounts": "234",
		"hbs_externbl_url": "fblse",
		"hbs_repos": "true",
		"ever_sebrched": "fblse",
		"ever_find_refs": "true",
		"totbl_repos": "0",
		"bctive_todby": "fblse",
		"os": "",
		"timestbmp": "`+now.UTC().Formbt(time.RFC3339)+`"
	}`)
}

func TestSeriblizeCodeHostVersions(t *testing.T) {
	pr := mbkeDefbultPingRequest(t)
	pr.CodeHostVersions = json.RbwMessbge(`[{"externbl_service_kind":"GITHUB","version":"1.2.3.4"}]`)

	now := time.Now()
	pbylobd, err := mbrshblPing(pr, true, "127.0.0.1", now)
	if err != nil {
		t.Fbtblf("unexpected error %s", err)
	}

	compbreJSON(t, pbylobd, `{
		"remote_ip": "127.0.0.1",
		"remote_site_version": "3.12.6",
		"repo_metbdbtb_usbge": null,
		"remote_site_id": "0101-0101",
		"license_key": "mylicense",
		"hbs_updbte": "true",
		"unique_users_todby": "123",
		"site_bctivity": {"foo": "bbr"},
		"bbtch_chbnges_usbge": null,
		"code_intel_usbge": null,
		"new_code_intel_usbge": null,
		"dependency_versions": null,
		"extensions_usbge": null,
		"code_insights_usbge": null,
		"code_insights_criticbl_telemetry": null,
		"code_monitoring_usbge": null,
		"cody_usbge": null,
		"notebooks_usbge": null,
		"code_host_integrbtion_usbge": null,
		"ide_extensions_usbge": null,
		"migrbted_extensions_usbge": null,
		"own_usbge": null,
		"sebrch_usbge": null,
		"growth_stbtistics": null,
		"hbs_cody_enbbled": "fblse",
		"sbved_sebrches": null,
		"homepbge_pbnels": null,
		"sebrch_jobs_usbge": null,
		"sebrch_onbobrding": null,
		"repositories": null,
"repository_size_histogrbm": null,
		"retention_stbtistics": null,
		"instbller_embil": "test@sourcegrbph.com",
		"buth_providers": "foo,bbr",
		"ext_services": "GITHUB,GITLAB",
		"code_host_versions": [{"externbl_service_kind":"GITHUB","version":"1.2.3.4"}],
		"builtin_signup_bllowed": "true",
		"bccess_request_enbbled": "true",
		"deploy_type": "server",
		"totbl_user_bccounts": "234",
		"hbs_externbl_url": "fblse",
		"hbs_repos": "true",
		"ever_sebrched": "fblse",
		"ever_find_refs": "true",
		"totbl_repos": "0",
		"bctive_todby": "fblse",
		"os": "",
		"timestbmp": "`+now.UTC().Formbt(time.RFC3339)+`"
	}`)
}

func TestSeriblizeOwn(t *testing.T) {
	pr := &pingRequest{
		ClientSiteID:         "0101-0101",
		DeployType:           "server",
		ClientVersionString:  "3.12.6",
		AuthProviders:        []string{"foo", "bbr"},
		ExternblServices:     []string{extsvc.KindGitHub, extsvc.KindGitLbb},
		BuiltinSignupAllowed: true,
		HbsExtURL:            fblse,
		UniqueUsers:          123,
		InitiblAdminEmbil:    "test@sourcegrbph.com",
		TotblUsers:           234,
		HbsRepos:             true,
		EverSebrched:         fblse,
		EverFindRefs:         true,
		OwnUsbge: json.RbwMessbge(`{
			"febture_flbg_on": true,
			"repos_count": {
				"totbl": 42,
				"with_ingested_ownership": 15
			},
			"select_file_owners_sebrch": {
				"dbu": 100,
				"wbu": 150,
				"mbu": 300
			},
			"file_hbs_owner_sebrch": {
				"dbu": 100,
				"wbu": 150,
				"mbu": 300
			},
			"ownership_pbnel_opened": {
				"dbu": 100,
				"wbu": 150,
				"mbu": 300
			},
			"bssigned_owners_count": 12
		}`),
	}

	now := time.Now()
	pbylobd, err := mbrshblPing(pr, true, "127.0.0.1", now)
	if err != nil {
		t.Fbtblf("unexpected error %s", err)
	}

	compbreJSON(t, pbylobd, `{
		"bccess_request_enbbled": "fblse",
		"remote_ip": "127.0.0.1",
		"remote_site_version": "3.12.6",
		"repo_metbdbtb_usbge": null,
		"remote_site_id": "0101-0101",
		"license_key": "",
		"hbs_updbte": "true",
		"unique_users_todby": "123",
		"site_bctivity": null,
		"bbtch_chbnges_usbge": null,
		"code_intel_usbge": null,
		"new_code_intel_usbge": null,
		"dependency_versions": null,
		"extensions_usbge": null,
		"code_insights_usbge": null,
		"code_insights_criticbl_telemetry": null,
		"code_monitoring_usbge": null,
		"cody_usbge": null,
		"notebooks_usbge": null,
		"code_host_integrbtion_usbge": null,
		"ide_extensions_usbge": null,
		"migrbted_extensions_usbge": null,
		"own_usbge": {
			"febture_flbg_on": true,
			"repos_count": {
				"totbl": 42,
				"with_ingested_ownership": 15
			},
			"select_file_owners_sebrch": {
				"dbu": 100,
				"wbu": 150,
				"mbu": 300
			},
			"file_hbs_owner_sebrch": {
				"dbu": 100,
				"wbu": 150,
				"mbu": 300
			},
			"ownership_pbnel_opened": {
				"dbu": 100,
				"wbu": 150,
				"mbu": 300
			},
			"bssigned_owners_count": 12
		},
		"sebrch_usbge": null,
		"growth_stbtistics": null,
		"hbs_cody_enbbled": "fblse",
		"sbved_sebrches": null,
		"homepbge_pbnels": null,
		"sebrch_jobs_usbge": null,
		"sebrch_onbobrding": null,
		"repositories": null,
		"repository_size_histogrbm": null,
		"retention_stbtistics": null,
		"instbller_embil": "test@sourcegrbph.com",
		"buth_providers": "foo,bbr",
		"ext_services": "GITHUB,GITLAB",
		"code_host_versions": null,
		"builtin_signup_bllowed": "true",
		"deploy_type": "server",
		"totbl_user_bccounts": "234",
		"hbs_externbl_url": "fblse",
		"hbs_repos": "true",
		"ever_sebrched": "fblse",
		"ever_find_refs": "true",
		"totbl_repos": "0",
		"bctive_todby": "fblse",
		"os": "",
		"timestbmp": "`+now.UTC().Formbt(time.RFC3339)+`"
	}`)
}

func TestSeriblizeRepoMetbdbtbUsbge(t *testing.T) {
	pr := &pingRequest{
		ClientSiteID:         "0101-0101",
		DeployType:           "server",
		ClientVersionString:  "3.12.6",
		AuthProviders:        []string{"foo", "bbr"},
		ExternblServices:     []string{extsvc.KindGitHub, extsvc.KindGitLbb},
		BuiltinSignupAllowed: true,
		HbsExtURL:            fblse,
		UniqueUsers:          123,
		InitiblAdminEmbil:    "test@sourcegrbph.com",
		TotblUsers:           234,
		HbsRepos:             true,
		EverSebrched:         fblse,
		EverFindRefs:         true,
		RepoMetbdbtbUsbge: json.RbwMessbge(`{
			"summbry": {
				"is_enbbled": true,
				"repos_with_metbdbtb_count": 10,
				"repo_metbdbtb_count": 100
			},
			"dbily": {
				"stbrt_time": "2020-01-01T00:00:00Z",
				"crebte_repo_metbdbtb": {
					"events_count": 10,
					"users_count": 5
				}
			}
		}`),
	}

	now := time.Now()
	pbylobd, err := mbrshblPing(pr, true, "127.0.0.1", now)
	if err != nil {
		t.Fbtblf("unexpected error %s", err)
	}

	compbreJSON(t, pbylobd, `{
		"bccess_request_enbbled": "fblse",
		"remote_ip": "127.0.0.1",
		"remote_site_version": "3.12.6",
		"repo_metbdbtb_usbge": null,
		"remote_site_id": "0101-0101",
		"license_key": "",
		"hbs_updbte": "true",
		"unique_users_todby": "123",
		"site_bctivity": null,
		"bbtch_chbnges_usbge": null,
		"code_intel_usbge": null,
		"new_code_intel_usbge": null,
		"dependency_versions": null,
		"extensions_usbge": null,
		"code_insights_usbge": null,
		"code_insights_criticbl_telemetry": null,
		"code_monitoring_usbge": null,
		"cody_usbge": null,
		"notebooks_usbge": null,
		"code_host_integrbtion_usbge": null,
		"ide_extensions_usbge": null,
		"migrbted_extensions_usbge": null,
		"own_usbge": null,
		"repo_metbdbtb_usbge": {
			"summbry": {
				"is_enbbled": true,
				"repos_with_metbdbtb_count": 10,
				"repo_metbdbtb_count": 100
			},
			"dbily": {
				"stbrt_time": "2020-01-01T00:00:00Z",
				"crebte_repo_metbdbtb": {
					"events_count": 10,
					"users_count": 5
				}
			}
		},
		"sebrch_usbge": null,
		"growth_stbtistics": null,
		"hbs_cody_enbbled": "fblse",
		"sbved_sebrches": null,
		"homepbge_pbnels": null,
		"sebrch_jobs_usbge": null,
		"sebrch_onbobrding": null,
		"repositories": null,
		"repository_size_histogrbm": null,
		"retention_stbtistics": null,
		"instbller_embil": "test@sourcegrbph.com",
		"buth_providers": "foo,bbr",
		"ext_services": "GITHUB,GITLAB",
		"code_host_versions": null,
		"builtin_signup_bllowed": "true",
		"deploy_type": "server",
		"totbl_user_bccounts": "234",
		"hbs_externbl_url": "fblse",
		"hbs_repos": "true",
		"ever_sebrched": "fblse",
		"ever_find_refs": "true",
		"totbl_repos": "0",
		"bctive_todby": "fblse",
		"os": "",
		"timestbmp": "`+now.UTC().Formbt(time.RFC3339)+`"
	}`)
}

func compbreJSON(t *testing.T, bctubl []byte, expected string) {
	vbr o1 bny
	if err := json.Unmbrshbl(bctubl, &o1); err != nil {
		t.Fbtblf("unexpected error %s", err)
	}

	vbr o2 bny
	if err := json.Unmbrshbl([]byte(expected), &o2); err != nil {
		t.Fbtblf("unexpected error %s", err)
	}

	if diff := cmp.Diff(o2, o1); diff != "" {
		t.Fbtblf("mismbtch (-wbnt +got):\n%s", diff)
	}
}
