package appliance

import (
	"context"

	"dario.cat/mergo"
	"golang.org/x/crypto/bcrypt"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/appliance/config"
	pb "github.com/sourcegraph/sourcegraph/internal/appliance/v1"
	"github.com/sourcegraph/sourcegraph/internal/releaseregistry"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

type Appliance struct {
	adminPasswordBcrypt []byte

	client                 client.Client
	namespace              string
	status                 config.Status
	sourcegraph            *config.Sourcegraph
	releaseRegistryClient  *releaseregistry.Client
	pinnedReleasesFile     string
	latestSupportedVersion string
	noResourceRestrictions bool
	logger                 log.Logger

	// Embed the UnimplementedApplianceServiceServer structs to ensure forwards compatibility (if the service is
	// compiled against a newer version of the proto file, the server will still have default implementations of any new
	// RPCs).
	pb.UnimplementedApplianceServiceServer
}

const (
	// Secret and key names
	dataSecretName                   = "appliance-data"
	dataSecretEncryptedPasswordKey   = "encrypted-admin-password"
	initialPasswordSecretName        = "appliance-password"
	initialPasswordSecretPasswordKey = "password"
)

func NewAppliance(
	client client.Client,
	relregClient *releaseregistry.Client,
	pinnedReleasesFile string,
	latestSupportedVersion string,
	namespace string,
	noResourceRestrictions bool,
	logger log.Logger,
) (*Appliance, error) {
	app := &Appliance{
		client:                 client,
		releaseRegistryClient:  relregClient,
		pinnedReleasesFile:     pinnedReleasesFile,
		latestSupportedVersion: latestSupportedVersion,
		namespace:              namespace,
		status:                 config.StatusInstall,
		noResourceRestrictions: noResourceRestrictions,
		sourcegraph:            &config.Sourcegraph{},
		logger:                 logger,
	}
	if err := app.reconcileBackingSecret(context.Background()); err != nil {
		return nil, err
	}
	return app, nil
}

func (a *Appliance) reconcileBackingSecret(ctx context.Context) error {
	backingSecretName := types.NamespacedName{Name: dataSecretName, Namespace: a.namespace}
	backingSecret := &corev1.Secret{}
	if err := a.client.Get(ctx, backingSecretName, backingSecret); err != nil {
		// Create the secret if not found, generating and deriving fields.
		if apierrors.IsNotFound(err) {
			backingSecret.SetName(dataSecretName)
			backingSecret.SetNamespace(a.namespace)
			if err := a.ensureBackingSecretKeysExist(ctx, backingSecret); err != nil {
				return err
			}
			return a.client.Create(ctx, backingSecret)
		}

		// Any other kind of error, return it
		return errors.Wrap(err, "getting backing secret")
	}

	// The backing secret already exists
	if err := a.ensureBackingSecretKeysExist(ctx, backingSecret); err != nil {
		return err
	}
	return a.client.Update(ctx, backingSecret)
}

func (a *Appliance) ensureBackingSecretKeysExist(ctx context.Context, secret *corev1.Secret) error {
	if secret.Data == nil {
		secret.Data = map[string][]byte{}
	}

	if _, ok := secret.Data[dataSecretEncryptedPasswordKey]; !ok {
		// Get admin-supplied password from separate secret, then delete it
		initialPasswordSecretNsName := types.NamespacedName{Name: initialPasswordSecretName, Namespace: a.namespace}
		var initialPasswordSecret corev1.Secret
		if err := a.client.Get(ctx, initialPasswordSecretNsName, &initialPasswordSecret); err != nil {
			a.logger.Info("no initial password secret exists. Please refer to https://github.com/sourcegraph/sourcegraph/blob/main/cmd/appliance/README.md#development")
			// We don't return an error here because we don't want to crash on
			// startup. We want to show the user an error message guiding them
			// to configure a password.
		}
		if adminPassword, ok := initialPasswordSecret.Data[initialPasswordSecretPasswordKey]; ok {
			adminPasswordBcrypt, err := bcrypt.GenerateFromPassword(adminPassword, 14)
			if err != nil {
				return errors.Wrap(err, "bcrypt-hashing password")
			}
			secret.Data[dataSecretEncryptedPasswordKey] = adminPasswordBcrypt

			if err := a.client.Delete(ctx, &initialPasswordSecret); err != nil {
				return errors.Wrap(err, "deleting initial password secret")
			}
		} else {
			a.logger.Info("The k8s secret appliance-password exists, but it does not contain a key password. Please refer to https://github.com/sourcegraph/sourcegraph/blob/main/cmd/appliance/README.md#development")
			// We don't return an error here because we don't want to crash on
			// startup. We want to show the user an error message guiding them
			// to configure a password.
		}
	}

	a.loadValuesFromSecret(secret)
	return nil
}

func (a *Appliance) loadValuesFromSecret(secret *corev1.Secret) {
	a.adminPasswordBcrypt = secret.Data[dataSecretEncryptedPasswordKey]
}

func (a *Appliance) GetCurrentVersion(ctx context.Context) string {
	return a.sourcegraph.Status.CurrentVersion
}

func (a *Appliance) GetCurrentStatus(ctx context.Context) config.Status {
	return a.status
}

func (a *Appliance) reconcileConfigMap(ctx context.Context, configMap *corev1.ConfigMap) error {
	existingCfgMapName := types.NamespacedName{Name: config.ConfigmapName, Namespace: a.namespace}
	existingCfgMap := &corev1.ConfigMap{}
	if err := a.client.Get(ctx, existingCfgMapName, existingCfgMap); err != nil {
		// Create the ConfigMap if not found
		if apierrors.IsNotFound(err) {
			spec, err := yaml.Marshal(a.sourcegraph)
			if err != nil {
				return errors.Wrap(err, "failed to marshal configmap yaml")
			}

			cfgMap := &corev1.ConfigMap{}
			cfgMap.Name = config.ConfigmapName
			cfgMap.Namespace = a.namespace

			cfgMap.Labels = map[string]string{
				"deploy": "sourcegraph",
			}

			cfgMap.Annotations = map[string]string{
				// required annotation for our controller filter.
				config.AnnotationKeyManaged: "true",
				config.AnnotationKeyStatus:  string(config.StatusUnknown),
				config.AnnotationConditions: "",
			}

			if configMap.ObjectMeta.Annotations != nil {
				cfgMap.ObjectMeta.Annotations = configMap.ObjectMeta.Annotations
			}

			cfgMap.Immutable = pointers.Ptr(false)
			cfgMap.Data = map[string]string{"spec": string(spec)}

			return a.client.Create(ctx, cfgMap)
		}

		return errors.Wrap(err, "getting configmap")
	}

	// The configmap already exists, update with any changed values
	if err := mergo.Merge(existingCfgMap, configMap, mergo.WithOverride); err != nil {
		return errors.Wrap(err, "merging configmaps")
	}

	return a.client.Update(ctx, existingCfgMap)
}

// isSourcegraphFrontendReady is a "health check" that is used to be able to know when our backing sourcegraph
// deployment is ready. This is a "quick and dirty" function and should be replaced with a more comprehensive
// health check in the very near future.
func (a *Appliance) isSourcegraphFrontendReady(ctx context.Context) (bool, error) {
	frontendDeploymentName := types.NamespacedName{Name: "sourcegraph-frontend", Namespace: a.namespace}
	frontendDeployment := &appsv1.Deployment{}
	if err := a.client.Get(ctx, frontendDeploymentName, frontendDeployment); err != nil {
		// If the frontend deployment is not found, we can assume it's not ready
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, errors.Wrap(err, "fetching frontend deployment")
	}

	return IsObjectReady(frontendDeployment)
}

func (a *Appliance) getStatus(ctx context.Context) (config.Status, error) {
	configMapName := types.NamespacedName{Name: config.ConfigmapName, Namespace: a.namespace}
	configMap := &corev1.ConfigMap{}
	if err := a.client.Get(ctx, configMapName, configMap); err != nil {
		if apierrors.IsNotFound(err) {
			return config.StatusUnknown, nil
		}
		return config.StatusUnknown, err
	}

	return config.Status(configMap.ObjectMeta.Annotations[config.AnnotationKeyStatus]), nil
}

func (a *Appliance) setStatus(ctx context.Context, status config.Status) error {
	configMapName := types.NamespacedName{Name: config.ConfigmapName, Namespace: a.namespace}
	configMap := &corev1.ConfigMap{}
	if err := a.client.Get(ctx, configMapName, configMap); err != nil {
		return err
	}

	configMap.Annotations[config.AnnotationKeyStatus] = string(status)
	err := a.client.Update(ctx, configMap)
	if err != nil {
		return errors.Wrap(err, "failed set status")
	}

	return nil
}
