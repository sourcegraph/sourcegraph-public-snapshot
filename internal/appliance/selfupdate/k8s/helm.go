package k8s

import (
	"fmt"
	"log"
	"os"

	"github.com/Masterminds/semver/v3"
	"github.com/sourcegraph/sourcegraph/appliance/selfupdate/k8s/watcher"
	"github.com/sourcegraph/sourcegraph/appliance/selfupdate/server"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"
)

var (
	myVersion           = ensureEnv("VERSION")
	applianceVersionVar = ensureEnv("APPLIANCE_VERSION")
	applianceId         = ensureEnv("APPLIANCE_ID")
	applianceChart      = ensureEnv("APPLIANCE_CHART")
	applianceRepo       = ensureEnv("APPLIANCE_REPO")
)

var (
	logPrefix = fmt.Sprintf("k8s/v%s: ", myVersion)
)

type K8sUpdater interface {
	server.Updater
}

func New(watcher watcher.Watcher) K8sUpdater {
	return &k8sUpdater{
		watcher: watcher,
	}
}

func ensureEnv(name string) string {
	val := os.Getenv(name)
	if val == "" {
		log.Fatalf("env var %s is not set", name)
	}
	return val
}

type k8sUpdater struct {
	logger  logger
	watcher watcher.Watcher
}

func init() {
	logger := logger{}
	logger.Println("--------------------------------------------")
	logger.Println("Initializing k8s updater.")
	logger.Println("Current version:", applianceVersionVar)
	logger.Println("Appliance chart:", applianceChart)
	logger.Println("Appliance repo:", applianceRepo)
	logger.Println("--------------------------------------------")
}

func (k *k8sUpdater) Start() (server.UpdaterResult, error) {
	k.logger.Println("Starting k8s updater.", "Current version:", applianceVersionVar)

	actionConfig := new(action.Configuration)
	settings := cli.New()

	if err := actionConfig.Init(
		settings.RESTClientGetter(),
		settings.Namespace(),
		"secret",
		k.logger.Printf,
	); err != nil {
		return server.UpdaterResultFailed,
			fmt.Errorf("failed to initialize helm action configuration: %w", err)
	}

	repoEntry := &repo.Entry{
		Name: applianceChart,
		URL:  applianceRepo,
	}

	newerVersion, latestRepoVersion, err := k.searchNewerVersion(repoEntry, settings)
	if err != nil {
		return server.UpdaterResultFailed,
			fmt.Errorf("failed to search newer version: %w", err)
	}

	if newerVersion == nil {
		k.logger.Println(
			"Chart is at current version already.",
			"Latest version on repo is", latestRepoVersion.String(),
		)
		return server.UpdaterResultUpToDate, nil
	}

	return server.UpdaterResultUpgraded,
		k.upgrade(repoEntry, newerVersion, actionConfig, settings)
}

func (k *k8sUpdater) searchNewerVersion(
	repoEntry *repo.Entry,
	settings *cli.EnvSettings,
) (*repo.ChartVersion, *semver.Version, error) {
	repository, err := repo.NewChartRepository(repoEntry, getter.All(settings))
	if err != nil {
		k.logger.Println("Failed to create repo", err.Error())
		return nil, nil, err
	}

	// Load index
	var indexFile string
	if indexFile, err = repository.DownloadIndexFile(); err != nil {
		return nil, nil, err
	}
	var index *repo.IndexFile
	if index, err = repo.LoadIndexFile(indexFile); err != nil {
		return nil, nil, err
	}

	currentVersion, err := semver.NewVersion(applianceVersionVar)
	if err != nil {
		k.logger.Println("Failed to read current version", err.Error())
		return nil, nil, err
	}

	var newerChart *repo.ChartVersion = nil
	var newerChartVersion semver.Version
	var latestFoundVersion *semver.Version = semver.New(0, 0, 0, "", "")

	newerChartVersion = *currentVersion

	for _, chart := range index.Entries[applianceChart] {
		chartVersion, err := semver.NewVersion(chart.Version)
		if err != nil {
			k.logger.Println("Failed to read chart version", err.Error())
			return nil, nil, err
		}
		if chartVersion.GreaterThan(&newerChartVersion) {
			newerChart = chart
			newerChartVersion = *chartVersion
		}
		if chartVersion.GreaterThan(latestFoundVersion) {
			latestFoundVersion = chartVersion
		}
	}

	return newerChart, latestFoundVersion, nil
}

func (k *k8sUpdater) upgrade(
	repoEntry *repo.Entry,
	version *repo.ChartVersion,
	config *action.Configuration,
	settings *cli.EnvSettings,
) error {
	chartDownloader := downloader.ChartDownloader{
		Out:              os.Stdout,
		Verify:           downloader.VerifyNever,
		Keyring:          "",
		Getters:          getter.All(&cli.EnvSettings{}),
		RepositoryConfig: settings.RepositoryConfig,
		RepositoryCache:  settings.RepositoryCache,
	}

	dir, err := os.MkdirTemp("", "helm-upgrade")
	if err != nil {
		k.logger.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(dir)

	chartUrl := repoEntry.URL + "/" + version.URLs[0]
	chartPath, _, err := chartDownloader.DownloadTo(chartUrl, version.Version, dir)
	if err != nil {
		k.logger.Fatalf("Failed to download chart: %v", err)
	}

	chart, err := loader.Load(chartPath)
	if err != nil {
		k.logger.Fatalf("Failed to load chart: %v", err)
	}

	upgrade := action.NewUpgrade(config)

	_, err = upgrade.Run(applianceId, chart, nil)
	if err != nil {
		k.logger.Fatalf("Failed to upgrade release: %v", err)
	}

	fmt.Println("Chart upgraded successfully")
	applianceVersionVar = version.Version
	k.watcher.UpdateVersion(version.Version)

	return nil
}

type logger struct{}

func (l logger) Printf(format string, v ...interface{}) {
	log.Printf(logPrefix+format+"\n", v...)
}

func (l logger) Println(v ...interface{}) {
	var data []interface{}
	data = append(data, logPrefix)
	data = append(data, v...)

	log.Println(data...)
}

func (l logger) Fatalf(format string, v ...interface{}) {
	log.Printf(logPrefix+format+"\n", v...)
	os.Exit(1)
}
