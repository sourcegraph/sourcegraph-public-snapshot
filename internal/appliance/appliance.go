package appliance

import (
	"context"

	"dario.cat/mergo"
	"golang.org/x/crypto/bcrypt"
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
	status                 Status
	sourcegraph            *config.Sourcegraph
	releaseRegistryClient  *releaseregistry.Client
	latestSupportedVersion string
	logger                 log.Logger

	// Embed the UnimplementedApplianceServiceServer structs to ensure forwards compatibility (if the service is
	// compiled against a newer version of the proto file, the server will still have default implementations of any new
	// RPCs).
	pb.UnimplementedApplianceServiceServer
}

// Status is a Stage that an Appliance can be in.
type Status string

const (
	StatusUnknown         Status = "unknown"
	StatusInstall         Status = "install"
	StatusInstalling      Status = "installing"
	StatusIdle            Status = "idle"
	StatusUpgrading       Status = "upgrading"
	StatusWaitingForAdmin Status = "wait-for-admin"
	StatusRefresh         Status = "refresh"

	// Secret and key names
	dataSecretName                   = "appliance-data"
	dataSecretJWTSigningKeyKey       = "jwt-signing-key"
	dataSecretEncryptedPasswordKey   = "encrypted-admin-password"
	initialPasswordSecretName        = "appliance-password"
	initialPasswordSecretPasswordKey = "password"
)

func (s Status) String() string {
	return string(s)
}

func NewAppliance(
	client client.Client,
	relregClient *releaseregistry.Client,
	latestSupportedVersion string,
	namespace string,
	logger log.Logger,
) (*Appliance, error) {
	app := &Appliance{
		client:                 client,
		releaseRegistryClient:  relregClient,
		latestSupportedVersion: latestSupportedVersion,
		namespace:              namespace,
		status:                 StatusInstall,
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

func (a *Appliance) GetCurrentStatus(ctx context.Context) Status {
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

			existingCfgMap.Labels = map[string]string{
				"deploy": "sourcegraph",
			}

			existingCfgMap.Annotations = map[string]string{
				// required annotation for our controller filter.
				config.AnnotationKeyManaged: "true",
				config.AnnotationConditions: "",
			}

			if configMap.ObjectMeta.Annotations != nil {
				existingCfgMap.ObjectMeta.Annotations = configMap.ObjectMeta.Annotations
			}

			existingCfgMap.Immutable = pointers.Ptr(false)
			existingCfgMap.Data = map[string]string{"spec": string(spec)}

			return a.client.Create(ctx, existingCfgMap)
		}

		return errors.Wrap(err, "getting configmap")
	}

	// The configmap already exists, update with any changed values
	if err := mergo.Merge(existingCfgMap, configMap, mergo.WithOverride); err != nil {
		return errors.Wrap(err, "merging configmaps")
	}

	return a.client.Update(ctx, existingCfgMap)
}
