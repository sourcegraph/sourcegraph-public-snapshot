package reconciler

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/sourcegraph/sourcegraph/internal/appliance/config"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Utilities to cause rolling deployments when secrets change live here.
// Indirectly tested through service-definition-specific golden tests.

const (
	pgsqlSecretName          = "pgsql-auth"
	codeInsightsDBSecretName = "codeinsights-db-auth"
	codeIntelDBSecretName    = "codeintel-db-auth"
	redisCacheSecretName     = "redis-cache"
	redisStoreSecretName     = "redis-store"
)

type DBConnSpecs struct {
	PG           *config.DatabaseConnectionSpec `json:"pg,omitempty"`
	CodeIntel    *config.DatabaseConnectionSpec `json:"codeintel,omitempty"`
	CodeInsights *config.DatabaseConnectionSpec `json:"codeinsights,omitempty"`
}

type RedisConnSpecs struct {
	Cache string `json:"cache,omitempty"`
	Store string `json:"store,omitempty"`
}

func (r *Reconciler) getDBSecrets(ctx context.Context, sg *config.Sourcegraph) (DBConnSpecs, error) {
	dbConnSpec, err := r.getDBSecret(ctx, sg, pgsqlSecretName)
	if err != nil {
		return DBConnSpecs{}, err
	}
	codeIntelConnSpec, err := r.getDBSecret(ctx, sg, codeIntelDBSecretName)
	if err != nil {
		return DBConnSpecs{}, err
	}
	codeInsightsConnSpec, err := r.getDBSecret(ctx, sg, codeInsightsDBSecretName)
	if err != nil {
		return DBConnSpecs{}, err
	}
	return DBConnSpecs{
		PG:           dbConnSpec,
		CodeIntel:    codeIntelConnSpec,
		CodeInsights: codeInsightsConnSpec,
	}, nil
}

func (r *Reconciler) getRedisSecrets(ctx context.Context, sg *config.Sourcegraph) (RedisConnSpecs, error) {
	redisCacheEndpoint, err := r.getRedisSecret(ctx, sg, redisCacheSecretName)
	if err != nil {
		return RedisConnSpecs{}, err
	}
	redisStoreEndpoint, err := r.getRedisSecret(ctx, sg, redisStoreSecretName)
	if err != nil {
		return RedisConnSpecs{}, err
	}
	return RedisConnSpecs{
		Cache: redisCacheEndpoint,
		Store: redisStoreEndpoint,
	}, nil
}

func (r *Reconciler) getDBSecret(ctx context.Context, sg *config.Sourcegraph, secretName string) (*config.DatabaseConnectionSpec, error) {
	dbSecret, err := r.getSecret(ctx, sg, secretName)
	if err != nil {
		return nil, err
	}

	return &config.DatabaseConnectionSpec{
		Host:     string(dbSecret.Data["host"]),
		Port:     string(dbSecret.Data["port"]),
		User:     string(dbSecret.Data["user"]),
		Password: string(dbSecret.Data["password"]),
		Database: string(dbSecret.Data["database"]),
	}, nil
}

func (r *Reconciler) getRedisSecret(ctx context.Context, sg *config.Sourcegraph, secretName string) (string, error) {
	redisSecret, err := r.getSecret(ctx, sg, secretName)
	if err != nil {
		return "", err
	}

	return string(redisSecret.Data["endpoint"]), nil
}

func (r *Reconciler) getSecret(ctx context.Context, sg *config.Sourcegraph, secretName string) (*corev1.Secret, error) {
	var secret corev1.Secret
	secretNsName := types.NamespacedName{Name: secretName, Namespace: sg.Namespace}
	if err := r.Client.Get(ctx, secretNsName, &secret); err != nil {
		if !kerrors.IsNotFound(err) {
			return nil, errors.Wrapf(err, "getting secret %s", secretName)
		}

		// If we cannot find the secret, return nil but also no error. We can
		// still serialize an ifChanged object in reconcileFrontendDeployment().
		// We should do this rather than fail the reconcile loop here, because
		// Kubernetes does not have inter-service dependencies, so it is
		// idiomatic to finish the loop even if the desired global final state
		// has not been reached. The next reconciliation after the secret exists
		// will yield a different result, which will cause deployed pods to roll
		// (since the spec.template.metadata.annotations changes).
		//
		// We return a zero-valued secret to avoid nil pointer explosions. All
		// data fields will be empty. Currently, all callers only use this
		// function to hash the data to see if its changed, so this seems ok to
		// do.
		log.FromContext(ctx).Info("could not find secret", "secretName", secretName, "err", err)
		return &corev1.Secret{}, nil
	}

	return &secret, nil
}
