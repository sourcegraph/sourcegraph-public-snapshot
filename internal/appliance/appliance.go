package appliance

import (
	"context"
	"crypto/rand"

	"golang.org/x/crypto/bcrypt"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	jwtSecret           []byte
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
	StatusUnknown    Status = "unknown"
	StatusSetup      Status = "setup"
	StatusInstalling Status = "installing"

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
		status:                 StatusSetup,
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
	if _, ok := secret.Data[dataSecretJWTSigningKeyKey]; !ok {
		jwtSigningKey, err := genRandomBytes(32)
		if err != nil {
			return err
		}
		secret.Data[dataSecretJWTSigningKeyKey] = jwtSigningKey
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
	a.jwtSecret = secret.Data[dataSecretJWTSigningKeyKey]
	a.adminPasswordBcrypt = secret.Data[dataSecretEncryptedPasswordKey]
}

func genRandomBytes(length int) ([]byte, error) {
	randomBytes := make([]byte, length)
	bytesRead, err := rand.Read(randomBytes)
	if err != nil {
		return nil, errors.Wrap(err, "reading random bytes")
	}
	if bytesRead != length {
		return nil, errors.Newf("expected to read %d random bytes, got %d", length, bytesRead)
	}
	return randomBytes, nil
}

func (a *Appliance) GetCurrentVersion(ctx context.Context) string {
	return a.sourcegraph.Status.CurrentVersion
}

func (a *Appliance) GetCurrentStatus(ctx context.Context) Status {
	return a.status
}

func (a *Appliance) CreateConfigMap(ctx context.Context, name string) (*corev1.ConfigMap, error) {
	spec, err := yaml.Marshal(a.sourcegraph)
	if err != nil {
		return nil, err
	}

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: a.namespace,
			Labels: map[string]string{
				"deploy": "sourcegraph",
			},
			Annotations: map[string]string{
				// required annotation for our controller filter.
				config.AnnotationKeyManaged: "true",
			},
		},
		Immutable: pointers.Ptr(false),
		Data: map[string]string{
			"spec": string(spec),
		},
	}

	if err := a.client.Create(ctx, configMap); err != nil {
		return nil, err
	}

	return configMap, nil
}

func (a *Appliance) GetConfigMap(ctx context.Context, name string) (*corev1.ConfigMap, error) {
	var applianceSpec corev1.ConfigMap
	err := a.client.Get(ctx, types.NamespacedName{Name: name, Namespace: a.namespace}, &applianceSpec)
	if apierrors.IsNotFound(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &applianceSpec, nil
}

func (a *Appliance) shouldSetupRun(ctx context.Context) (bool, error) {
	cfgMap, err := a.GetConfigMap(ctx, "sourcegraph-appliance")
	switch {
	case err != nil:
		return false, err
	case a.status == StatusInstalling:
		// configMap does not exist but is being created
		return false, nil
	case cfgMap == nil:
		// configMap does not exist
		return true, nil
	case cfgMap.Annotations[config.AnnotationKeyManaged] == "false":
		// appliance is not managed
		return false, nil
	default:
		return true, nil
	}
}
