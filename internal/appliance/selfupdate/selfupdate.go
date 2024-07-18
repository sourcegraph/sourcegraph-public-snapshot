package selfupdate

import (
	"context"
	"strings"
	"time"

	"github.com/life4/genesis/slices"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/appliance"
	"github.com/sourcegraph/sourcegraph/internal/releaseregistry"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type SelfUpdate struct {
	Interval       time.Duration
	Logger         log.Logger
	K8sClient      client.Client
	RelregClient   releaseregistry.ReleaseRegistryClient
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
			if err := u.Once(ctx); err != nil {
				u.Logger.Error("error self-updating", log.Error(err))
				return err
			}
		case <-ctx.Done():
			u.Logger.Error("self-update context done, exiting", log.Error(ctx.Err()))
			return ctx.Err()
		}
	}
}

func (u *SelfUpdate) Once(ctx context.Context) error {
	u.Logger.Info("starting self-update")

	var dep appsv1.Deployment
	depName := types.NamespacedName{Name: u.DeploymentName, Namespace: u.Namespace}
	if err := u.K8sClient.Get(ctx, depName, &dep); err != nil {
		return errors.Wrap(err, "getting deployment")
	}

	newTag, err := u.getLatestTag(ctx)
	if err != nil {
		return errors.Wrap(err, "getting latest tag")
	}

	dep.Spec.Template.Spec.Containers[0].Image = replaceTag(dep.Spec.Template.Spec.Containers[0].Image, newTag)
	if err := u.K8sClient.Update(ctx, &dep); err != nil {
		return errors.Wrap(err, "updating deployment")
	}

	return nil
}

func (u *SelfUpdate) getLatestTag(ctx context.Context) (string, error) {
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
	semvers, err := appliance.ParseVersions(versionStrs)
	if err != nil {
		return "", errors.Wrap(err, "parsing versions from release registry")
	}
	latestVersion := semvers[len(semvers)-1].String()

	u.Logger.Info("found latest version", log.String("version", latestVersion))
	return latestVersion, nil
}

// I thought about using regular expressions for this but I swear that's not
// better.
func replaceTag(image, newTag string) string {
	imgParts := strings.Split(image, ":")
	return strings.Join(imgParts[:len(imgParts)-1], ":") + ":" + newTag
}
