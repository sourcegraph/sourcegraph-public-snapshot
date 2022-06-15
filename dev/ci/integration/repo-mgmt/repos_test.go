package main

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/gqltestutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestIndexReposThenSearch(t *testing.T) {
	if len(*githubToken) == 0 {
		t.Skip("Environment variable GITHUB_TOKEN is not set")
	}

	// Set up external service
	esID, err := client.AddExternalService(gqltestutil.AddExternalServiceInput{
		Kind:        extsvc.KindGitHub,
		DisplayName: "repomgmt-github-repoPathPattern",
		Config: mustMarshalJSONString(struct {
			URL   string   `json:"url"`
			Token string   `json:"token"`
			Repos []string `json:"repos"`
		}{
			URL:   "https://github.com/",
			Token: *githubToken,
			Repos: hugeReposList,
			// RepositoryPathPattern: "github.com/{nameWithOwner}",
		}),
	})
	// The repo-updater might not be up yet, but it will eventually catch up for the external
	// service we just added, thus it is OK to ignore this transient error.
	if err != nil && !strings.Contains(err.Error(), "/sync-external-service") {
		t.Fatal(err)
	}
	defer func() {
		// err := client.DeleteExternalService(esID, false)
		// if err != nil {
		// 	t.Fatal(err)
		// }
		println(esID)
	}()

	err = waitForReposToBeIndexed(client, hugeReposList...)
	if err != nil {
		t.Fatal(err)
	}

	res, err := client.SearchFiles(`index:yes sql.Open("genji", ":memory:")`)
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for _, r := range res.Results {
		if r.Repository.Name == "github.com/genjidb/genji" {
			found = true
			break
		}
	}

	if !found {
		t.Log("expected to find 'github.com/genjidb/genji' in the results, but did not")
		t.Fail()
	}
}

func waitForReposToBeIndexed(c *gqltestutil.Client, repos ...string) error {
	timeout := 3000 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var missing []string
	for {
		select {
		case <-ctx.Done():
			return errors.Errorf("wait for repos to be indexed timed out in %s, still missing %v", timeout, missing)
		default:
		}

		var err error
		missing, err = queryReposToBeIndexed(c, repos...)
		if err != nil {
			return errors.Wrap(err, "wait for repos to be indexd")
		}
		if len(missing) == 0 {
			break
		}

		time.Sleep(100 * time.Millisecond)
	}
	return nil
}

func queryReposToBeIndexed(c *gqltestutil.Client, repos ...string) ([]string, error) {
	var resp struct {
		Data struct {
			Repositories struct {
				Nodes []struct {
					Name       string `json:"name"`
					MirrorInfo struct {
						Cloned bool `json:"cloned"`
					} `json:"mirrorInfo"`
				} `json:"nodes"`
			} `json:"repositories"`
		} `json:"data"`
	}
	query := `{
   repositories(first: 1000, notIndexed: false) {
    totalCount(precise: true)
    nodes {
      name
      mirrorInfo {
        cloned
      }
    }
  }   
}`
	err := c.GraphQL("", query, nil, &resp)
	if err != nil {
		return nil, errors.Wrap(err, "request GraphQL")
	}

	repoSet := make(map[string]struct{}, len(repos))
	for _, repo := range repos {
		repoSet[fmt.Sprintf("%s%s", "github.com/", repo)] = struct{}{}
	}
	for _, node := range resp.Data.Repositories.Nodes {
		if node.MirrorInfo.Cloned {
			delete(repoSet, node.Name)
		}
	}
	if len(repoSet) > 0 {
		missing := make([]string, 0, len(repoSet))
		for name := range repoSet {
			missing = append(missing, name)
		}
		return missing, nil
	}

	return nil, nil
}

var hugeReposList = []string{
	"MaulingMonkey/untokio",
	"cescoferraro/sketch",
	"imallan/tuchong-daily-android",
	"streamx-co/streamx-co.github.io",
	"EpicStep/echo-pprof",
	"adegenhardt/fakemon-generator",
	"STA210-Sp19/supplemental-notes",
	"Adlink-ROS/rootOnStorage",
	"vshymanskyy/Preferences",
	"fm-polimi/zot",
	// "OverTheWireOrg/statuspage",
	// "rodluger/bfgtest",
	// "tomast1337/Rio-Vagas-Bot",
	// "Ajeethkumar-r/todo-app",
	// "biltzerpos/APTLY",
	// "maryamhonari/faceRecognitionDevelopment",
	// "cretz/bevy_skeleton_poc",
	// "dd32/meta.w.org",
	// "chzchzchz/midispa",
	// "deniszh/graphite-web",
	// "datocms/remix-example",
	// "pmclSF/DeepCompress",
	// "Pearuss/redux-thunk",
	// "mattec92/KTH",
	// "sainttelant/yolov5",
	// "lalitkapoor/webpack-es6-react-boilerplate",
	// "pixelgrade/customify",
	// "appintheair/WatchAnimationHelper",
	// "wentaocheng-cv/cpf_localization",
	// "xiaoruiguo/vue3-antd-admin",
	// "zb121/100-gdb-tips",
	// "akramer/aoc2021",
	// "eyalperry88/eyalperry88.github.io",
	// "lexaguilar/NodeServer",
	// "Danceiny/parser_engine",
	// "tctien342/Asus-Vivobook-S510UA-Hackintosh",
	// "guilhermedeoliveira/ns3-training-challenges",

	// that's the one we're using to write queries against
	"genjidb/genji",

	// "boostcampaitech2/final-project-level3-nlp-19",
	// "dfawley/grpc.io",
	// "DevClancy/aurora",
	// "bljessica/netease-cloud-music",
	// "wan2land/style-splitter",
	// "papachristoumarios/call-graphs",
	// "artex2000/bpe",
	// "ERPG/Foreigh-exchange-widget",
	// "ironSource/node-if-async",
	// "Justinliu123/miniProject",
	// "gkmngrgn/config",
	// "TrungJamin/spectrum",
	// "mubaidr/vue2-migration-helper",
	// "elk-audio/elkpy",
	// "trarbr/nerves_livebook",
	// "jiwei0921/DMRA",
	// "LordLiang/veloNet",
	// "kraxarn/spotify-qt-builds",
	// "Terkwood/forest",
	// "pBlasiak/interThermalPhaseChangeFoam_OF240",
	// "magnetikonline/ssh-diff",
	// "SolaceProducts/solace-spring-cloud",
	// "mhmtaltnts/tema",
	// "psalm/psalm.dev",
	// "benjaminwgordon/be-kind",
	// "seolhw/packagingRequest",
	// "ranran472970026/TextGrapher",
	// "AkiniKites/hzd-model-db-gen",
	// "karma-runner/karma-closure",
	// "akira-cn/vue3-sfc-loader",
	// "afeiship/react-ant-cascader",
	// "VesperPiggy/StudyProcessing",
	// "leffss/devops",
	// "henryholtgeerts/ghostdar",
	// "jackzampolin/bsk-idx",
	// "ThinkingPractice/InfiniTime",
	// "mitchsw/perf_data_converter",
	// "kmarques/esgi-api-tests",
	// "NLMichaud/WeeklyCDCPlot",
	// "podgorskiy/StyleGANCpp",
	// "ThinkingPractice/spec_mtk",
	// "stonneau/love",
	// "ra-kesh/armor-backend",
	// "papay0/draft-js-checkable-list-item",
	// "JuliaCrypto/Ripemd.jl",
	// "SaigyoujiNono/edu-online-front",
	// "evelynf/Ops",
	// "yuto-oshima/use-japan-businee-day",
	// "jaredmdobson/rack-livereload",
	// "kubesphere-sigs/kubesphere",
	// "gscgoon/OnLine_B",
	// "ALTURKA/gitlab-compose-kit-custom-config",
	// "LucidApp/bubbletea",
	// "MaartenGr/ReinforcementLearning",
	// "zjuluoyang/evapVOFHardt",
	// "ThinkingPractice/MTK-IOT",
	// "hin1115/Place365-Res2net-classifier",
	// "Stuff90/kalypsospa",
	// "AnnThanicha/FMD-spatial-transmission-kernel",
	// "ZhengtongYan/every-programmer-should-know",
	// "POSTECH-IMLAB/PIMNet_Internal_Environment_Recognition",
	// "PlatziMaster/react-shop",
	// "ghlonghan007/ghlonghan007.github.io",
	// "2048JiaLi/Chinese-Text-Mining-Model-LDA",
	// "randomprocess/SUIToolKit",
	// "Cloudox/SegueTest",
	// "ainblockchain/ain-js",
	// "ThatsMrTalbot/scaffold",
	// "oradwell/verbose-twit-banner",
	// "turtlecoin/.trtl",
	// "mohammedovich/PS4EMX",
	// "jedevc/apparea",
	// "ariefdarvin/MobileProgProject",
	// "s3cretclub/Proxy-Scraper-Finder",
	// "cfbolz/minitrace",
	// "bleidornm/PSADTK-Prerequisits",
	// "makerinchina-iot/mcuoneclipse",
	// "ThinkingPractice/MTK2503-DOCUMENTATION",
	// "MrEliptik/godot_video_to_animated_texture",
	// "CroatiaParanoia/ts-morph-playground",
	// "MaratSaidov/artificial-text-detection",
	// "dionyziz/advent-of-code",
	// "vgribok/Aspectacular",
	// "ShengChangWei/e-vue-esrimapjs",
	// "travellerwjoe/QuickSearch-Chrome-Extension",
	// "Stevemcold/Hospital-Simulator",
	// "Saul-RobotPenguin/User-Dunmy-Data",
	// "sergevdz/profastapi",
	// "ExcaliburPro/superl-url",
	// "Swagga5aur/interThermalPhaseFoam",
	// "KinVen-Lee/Javascript",
	// "AkiniKites/FullRareSetManager",
	// "m-esm/node-multi-branch",
	// "binary-coffee-dev/blog-database",
	// "JintaoWang1/newRepo",
	// "icela/FriceEngine-CSharp",
	// "telegeography/www.internetexchangemap.com",
	// "camilla-design/javascript-1-ca",
	// "ynfw183/dy123",
	// "ifyall/zephyr",
	// "stinsonga/geo-prime-workspace",
	// "Lewis-Marshall/WinUI3NavigationExample",
	// "droqen/magical-paradise-train",
	// "jmickey/logmon",
	// "chinmaynautiyal/curveFitting-Backpropagation",
	// "tbels/recipe-crawler",
	// "pinbo/SNP_Primer_Pipeline",
	// "tmtk75/aliyun-sdk-go",
	// "pinzolo/errz",
	// "dxdbl/mergeExcel",
	// "zhangry868/Coursera-DataScience-Series",
	// "cotequeiroz/afalg_engine",
	// "wechaty/ha",
	// "agilepro/posthoc",
	// "brlin-tw/bashfuscator-snap",
	// "ThinkingPractice/test",
	// "hanthomas/protractor-recorder",
	// "ECNURoboLab/ecnu_pick_place",
	// "be5invis/be5invis.github.io",
	// "abdox1234/coupon",
	// "andywu188/XQuickEnergy",
	// "fullstack-development/plutus",
	// "yxdz2020/freeJiedian",
	// "Tornaco/CheckableImageView",
	// "chinmaynautiyal/curveFitting-Backpropagation",
	// "tbels/recipe-crawler",
	// "pinbo/SNP_Primer_Pipeline",
	// "tmtk75/aliyun-sdk-go",
	// "pinzolo/errz",
	// "dxdbl/mergeExcel",
	// "zhangry868/Coursera-DataScience-Series",
	// "cotequeiroz/afalg_engine",
	// "wechaty/ha",
	// "agilepro/posthoc",
	// "brlin-tw/bashfuscator-snap",
	// "ThinkingPractice/test",
	// "hanthomas/protractor-recorder",
	// "ECNURoboLab/ecnu_pick_place",
	// "be5invis/be5invis.github.io",
	// "abdox1234/coupon",
	// "andywu188/XQuickEnergy",
	// "fullstack-development/plutus",
	// "yxdz2020/freeJiedian",
	// "Tornaco/CheckableImageView",
	// "zhstark/crawler_1point3",
	// "toposoid/toposoid",
	// "MatiasParedesF/miBancoFrontEnd",
	// "littledivy/indexeddb-sqlite",
	// "Jansora/OnlineCompiler",
	// "kashav/bugs",
	// "aliyun/AliyunPlayer-android-sample",
	// "jiang4yu/Windows-10-LTSC-MicrosoftStore",
	// "tonogram/nix",
	// "MEDSL/election-scrapers",
	// "Gabse/DMX2PMX",
	// "ThinkingPractice/sourcecode",
	// "OCB7D2D/ElectricityOverhaul",
	// "comeforu2012/Zbackup",
	// "yonghakim/m1_benchmark",
	// "midi-mixer/plugin-discord",
	// "redbeard28/ansible_role_iLo",
	// "ngx-semantic/ngx-semantic-docs",
	// "ello/kinesis-stream-reader",
	// "sigma/persp-mode.el",
	// "bogdansemkin/GWO-algorithm",
	// "theblockstalk/eosio-contracts",
	// "sirkkalap/bfg-example",
	// "ycwu1030/LamWZ",
	// "denyska/jfrog-client-go",
	// "InkiYinji/Python",
	// "qiyan98/Manipulator-dynamics",
	// "DavoLarris/springAPIEmployee",
	// "W-OVERFLOW/KotlinForgeMixinTemplate",
	// "arunprasadmudaliar/pura",
	// "vim-scripts/ats-lang-vim",
	// "CopernicusAustralasia/auscophub",
	// "erdmenchen/laravel-bitbucket-deploy",
	// "Ramkec15aur039/CoreUi-AdminTemplate-React-JavaScript",
}
