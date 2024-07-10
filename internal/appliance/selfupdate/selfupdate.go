package selfupdate

import (
	"context"
	"strings"
	"time"

	"github.com/life4/genesis/slices"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/appliance"
	"github.com/sourcegraph/sourcegraph/internal/appliance/config"
	"github.com/sourcegraph/sourcegraph/internal/releaseregistry"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type ReleaseRegistryClient interface {
	ListVersions(ctx context.Context, product string) ([]releaseregistry.ReleaseInfo, error)
}

type SelfUpdate struct {
	Interval       time.Duration
	Logger         log.Logger
	K8sClient      client.Client
	RelregClient   ReleaseRegistryClient
	DeploymentName string
	Namespace      string
}

func (u *SelfUpdate) Loop(ctx context.Context) error {
	u.Logger.Info("starting self-update loop")

	ticker := time.NewTicker(u.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := u.once(ctx); err != nil {
				u.Logger.Error("error self-updating", log.Error(err))
				return err
			}
		case <-ctx.Done():
			u.Logger.Error("self-update context done, exiting", log.Error(ctx.Err()))
			return ctx.Err()
		}
	}
}

func (u *SelfUpdate) once(ctx context.Context) error {
	u.Logger.Info("starting self-update")

	var dep appsv1.Deployment
	depName := types.NamespacedName{Name: u.DeploymentName, Namespace: u.Namespace}
	if err := u.K8sClient.Get(ctx, depName, &dep); err != nil {
		return errors.Wrap(err, "getting deployment")
	}

	currentSGVersion, err := u.getCurrentlyDeployedSGVersion(ctx)
	if err != nil {
		// Wait for SG to be deployed before alloweing self-update
		if kerrors.IsNotFound(err) {
			u.Logger.Info("Sourcegraph ConfigMap not found, exiting appliance self-update")
			return nil
		}
		return errors.Wrap(err, "determining current Sourcegraph version")
	}

	newTag, err := u.getLatestTag(ctx, currentSGVersion)
	if err != nil {
		return errors.Wrap(err, "getting latest tag")
	}

	dep.Spec.Template.Spec.Containers[0].Image = replaceTag(dep.Spec.Template.Spec.Containers[0].Image, newTag)
	if err := u.K8sClient.Update(ctx, &dep); err != nil {
		return errors.Wrap(err, "updating deployment")
	}

	return nil
}

func (u *SelfUpdate) getCurrentlyDeployedSGVersion(ctx context.Context) (string, error) {
	var cfgMap corev1.ConfigMap
	cfgMapName := types.NamespacedName{Name: config.ConfigmapName, Namespace: u.Namespace}
	if err := u.K8sClient.Get(ctx, cfgMapName, &cfgMap); err != nil {
		return "", err
	}
	return cfgMap.GetAnnotations()[config.AnnotationKeyCurrentVersion], nil
}

// Get latest appliance version that is no more than 2 minor versions ahead of
// the currently-deployed Sourcegraph version.
func (u *SelfUpdate) getLatestTag(ctx context.Context, currentSGVersion string) (string, error) {
	versions, err := u.RelregClient.ListVersions(ctx, "sourcegraph")
	if err != nil {
		return "", err
	}
	versionStrs := slices.MapFilter(versions, func(version releaseregistry.ReleaseInfo) (string, bool) {
		return version.Version, version.Public
	})
	if len(versionStrs) == 0 {
		return "", errors.New("no versions found")
	}
	latestVersion, err := appliance.HighestVersionNoMoreThanNMinorFromBase(versionStrs, currentSGVersion, 2)
	if err != nil {
		return "", err
	}
	u.Logger.Info("found latest version", log.String("version", latestVersion))
	return latestVersion, nil
}

// I thought about using regular expressions for this but I swear that's not
// better.
func replaceTag(image, newTag string) string {
	imgParts := strings.Split(image, ":")
	return strings.Join(imgParts[:len(imgParts)-1], ":") + ":" + newTag
}
