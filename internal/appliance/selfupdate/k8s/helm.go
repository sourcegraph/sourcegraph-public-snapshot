package k8s

import (
	"fmt"
	"log"
	"os"

	"github.com/Masterminds/semver/v3"
	"github.com/sourcegraph/sourcegraph/appliance/selfupdate/server"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"
)

var (
	currentVersionVar = ensureEnv("VERSION")
	applianceChart    = ensureEnv("APPLIANCE_CHART")
	applianceRepo     = ensureEnv("APPLIANCE_REPO")
)

type K8sUpdater interface {
	server.Updater
}

func New() K8sUpdater {
	return &k8sUpdater{}
}

func ensureEnv(name string) string {
	val := os.Getenv(name)
	if val == "" {
		panic(fmt.Sprintf("env var %s is not set", name))
	}
	return val
}

type k8sUpdater struct{}

func (k *k8sUpdater) Start() (server.UpdaterResult, error) {
	log.Println("Starting k8s updater.", "Current version:", currentVersionVar)

	actionConfig := new(action.Configuration)
	settings := cli.New()

	if err := actionConfig.Init(
		settings.RESTClientGetter(),
		settings.Namespace(),
		"secret",
		log.Printf,
	); err != nil {
		return server.UpdaterResultFailed,
			fmt.Errorf("failed to initialize helm action configuration: %w", err)
	}

	repoEntry := &repo.Entry{
		Name: applianceChart,
		URL:  applianceRepo,
	}

	newerVersion, err := k.searchNewerVersion(repoEntry, settings)
	if err != nil {
		return server.UpdaterResultFailed,
			fmt.Errorf("failed to search newer version: %w", err)
	}

	if newerVersion == nil {
		log.Println("Chart is at current version already.")
		return server.UpdaterResultUpToDate, nil
	}

	return server.UpdaterResultUpgraded,
		k.upgrade(repoEntry, newerVersion, actionConfig, settings)
}

func (k *k8sUpdater) searchNewerVersion(
	repoEntry *repo.Entry,
	settings *cli.EnvSettings,
) (*repo.ChartVersion, error) {
	repository, err := repo.NewChartRepository(repoEntry, getter.All(settings))
	if err != nil {
		log.Println("Failed to create repo", err.Error())
		return nil, err
	}

	// Load index
	var indexFile string
	if indexFile, err = repository.DownloadIndexFile(); err != nil {
		return nil, err
	}
	var index *repo.IndexFile
	if index, err = repo.LoadIndexFile(indexFile); err != nil {
		return nil, err
	}

	currentVersion, err := semver.NewVersion(currentVersionVar)
	if err != nil {
		log.Println("Failed to read current version", err.Error())
		return nil, err
	}

	var newerChart *repo.ChartVersion = nil
	var newerChartVersion semver.Version

	newerChartVersion = *currentVersion

	for _, chart := range index.Entries[applianceChart] {
		chartVersion, err := semver.NewVersion(chart.Version)
		if err != nil {
			log.Println("Failed to read chart version", err.Error())
			return nil, err
		}
		if chartVersion.GreaterThan(&newerChartVersion) {
			newerChart = chart
			newerChartVersion = *chartVersion
		}
	}

	return newerChart, nil
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
		log.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(dir)

	chartUrl := repoEntry.URL + "/" + version.URLs[0]
	chartPath, _, err := chartDownloader.DownloadTo(chartUrl, version.Version, dir)
	if err != nil {
		log.Fatalf("Failed to download chart: %v", err)
	}

	chart, err := loader.Load(chartPath)
	if err != nil {
		log.Fatalf("Failed to load chart: %v", err)
	}

	upgrade := action.NewUpgrade(config)

	_, err = upgrade.Run(chart.Name(), chart, nil)
	if err != nil {
		log.Fatalf("Failed to upgrade release: %v", err)
	}

	fmt.Println("Chart upgraded successfully")
	currentVersionVar = version.Version

	return nil
}
