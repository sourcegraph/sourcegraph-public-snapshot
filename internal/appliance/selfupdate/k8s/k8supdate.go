package k8s

import (
	"fmt"
	"log"
	"os"

	"github.com/Masterminds/semver/v3"
	"github.com/sourcegraph/sourcegraph/appliance/selfupdate"
	"github.com/sourcegraph/sourcegraph/appliance/selfupdate/schema"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"
)

var (
	currentVersionVar = ensureEnv("VERSION")
	applianceId       = ensureEnv("APPLIANCE_ID")
	applianceChart    = ensureEnv("APPLIANCE_CHART")
	applianceRepo     = ensureEnv("APPLIANCE_REPO")
)

type K8sUpdater interface {
	selfupdate.ComponentUpdate
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

func (k *k8sUpdater) Update(comp *schema.ComponentUpdateInformation) (*semver.Version, error) {
	newVer, err := semver.NewVersion(comp.Version)
	if err != nil {
		log.Println("Failed to read component version", err.Error())
		return nil, err
	}
	if current, err := k.installedVersion(comp.Name); err != nil {
		log.Println("Failed to list releases", err.Error())
		return nil, err
	} else {
		switch current.Compare(newVer) {
		case 0:
			log.Println("Already installed", comp.Name, current.String())
			return current, nil
		case 1:
			log.Println("Current version is newer. Downgrading",
				"from", current.String(), "to", newVer.String())
			return current, nil
		case -1:
			log.Println("Current needs update", comp.Name,
				"from", current.String(), "to", newVer.String())
			return k.updateToVersion(comp, newVer)
		}
	}
	return nil, nil
}

func (k *k8sUpdater) installedVersion(pkgName string) (*semver.Version, error) {
	actionConfig := new(action.Configuration)
	settings := cli.New()

	if err := actionConfig.Init(settings.RESTClientGetter(), settings.Namespace(), "secret", log.Printf); err != nil {
		return nil, fmt.Errorf("failed to initialize helm action configuration: %w", err)
	}

	newerVersion, err := k.searchNewerVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to search newer version: %w", err)
	}

	if newerVersion == nil {
		log.Println("Chart is at current version already.")
		return nil, nil // no newer version
	}

	k.upgrade(newerVersion, actionConfig)
	return nil, nil
}

func (k *k8sUpdater) searchNewerVersion() (*repo.ChartVersion, error) {
	repoEntry := &repo.Entry{
		Name: applianceChart,
		URL:  applianceRepo,
	}

	settings := cli.New()
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
	version *repo.ChartVersion,
	config *action.Configuration,
) error {
	settings := cli.New()

	chartDownloader := downloader.ChartDownloader{
		Out:              os.Stdout,
		Verify:           downloader.VerifyNever,
		Keyring:          "",
		Getters:          getter.All(&cli.EnvSettings{}),
		RepositoryConfig: settings.RepositoryConfig,
		RepositoryCache:  settings.RepositoryCache,
	}

	chartPath, _, err := chartDownloader.DownloadTo(version.Name, version.Version, os.TempDir())
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

	return nil
}
