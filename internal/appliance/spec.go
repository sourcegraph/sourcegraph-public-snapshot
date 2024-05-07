package appliance

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/sourcegraph/sourcegraph/internal/appliance/config"
)

type ManagementStateType string

const (
	// ManagementStateManaged denotes that Sourcegraph should be reconciled
	// by the operator.
	ManagementStateManaged ManagementStateType = "Managed"

	// ManagementStateUnmanaged denotes that Sourcegraph should not be reconciled
	// by the operator.
	ManagementStateUnmanaged ManagementStateType = "Unmanaged"
)

type DatabaseSpec struct {
	Host     string `json:"host,omitempty"`
	Port     string `json:"port,omitempty"`
	User     string `json:"user,omitempty"`
	Password string `json:"password,omitempty"`
	Database string `json:"database,omitempty"`
}

// BlobstoreSpec defines the desired state of Blobstore.
type BlobstoreSpec struct {
	config.StandardConfig

	// StorageSize defines the requested amount of storage for the PVC.
	// Default: 200Gi
	StorageSize string `json:"storageSize,omitempty"`

	// Env defines environment variables for Blobstore.
	Env map[string]string `json:"env,omitempty"`
}

// CodeInsightsDBSpec defines the desired state of Code Insights database.
type CodeInsightsDBSpec struct {
	// Disabled defines if Code Insights is enabled or not.
	// Default: false
	Disabled bool `json:"disabled,omitempty"`

	// ExistingSecret is the name of an existing secret to use for CodeInsights DB credentials.
	ExistingSecret string `json:"existingSecret,omitempty"`

	// Database allows for custom database connection details.
	Database *DatabaseSpec `json:"database,omitempty"`

	// StorageSize defines the requested amount of storage for the PVC.
	// Default: 200Gi
	StorageSize string `json:"storageSize,omitempty"`

	// Resources allows for custom resource limits and requests.
	Resources *corev1.ResourceList `json:"resources,omitempty"`

	// Env defines environment variables for Code Insights.
	Env map[string]string `json:"env,omitempty"`
}

// CodeIntelDBSpec defines the desired state of Code Intel database.
type CodeIntelDBSpec struct {
	// Disabled defines if Code Intel is enabled or not.
	// Default: false
	Disabled bool `json:"disabled,omitempty"`

	// ExistingSecret is the name of an existing secret to use for CodeIntel DB credentials.
	ExistingSecret string `json:"existingSecret,omitempty"`

	// Database allows for custom database connection details.
	Database *DatabaseSpec `json:"database,omitempty"`

	// StorageSize defines the requested amount of storage for the PVC.
	// Default: 200Gi
	StorageSize string `json:"storageSize,omitempty"`

	// Resources allows for custom resource limits and requests.
	Resources *corev1.ResourceList `json:"resources,omitempty"`

	// Env defines environment variables for Code Intel.
	Env map[string]string `json:"env,omitempty"`
}

type IngressSpec struct {
	Disabled         bool              `json:"enabled,omitempty"`
	Annotations      map[string]string `json:"annotations,omitempty"`
	Host             string            `json:"host,omitempty"`
	IngressClassName string            `json:"ingressClassName,omitempty"`
	TLSSecret        string            `json:"tlsSecret,omitempty"`
}

// FrontendSpec defines the desired state of Frontend.
type FrontendSpec struct {
	// Replicas defines the number of Frontend pod replicas.
	// Default: 2
	Replicas int32 `json:"replicas,omitempty"`

	// Ingress allows for changes to the custom Sourcegraph ingress.
	Ingress *IngressSpec `json:"ingress,omitempty"`

	// ExistingSecret is the name of an existing secret to use for Postgres credentials.
	ExistingSecret string `json:"existingSecret,omitempty"`

	// Resources allows for custom resource limits and requests.
	Resources *corev1.ResourceList `json:"resources,omitempty"`

	// Env defines environment variables for Frontend.
	Env map[string]string `json:"env,omitempty"`
}

// GitServerSpec defines the desired state of GitServer.
type GitServerSpec struct {
	config.StandardConfig

	// Replicas defines the number of Symbols pod replicas.
	// Default: 1
	Replicas int32 `json:"replicas,omitempty"`

	// StorageSize defines the requested amount of storage for the PVC.
	// Default: 200Gi
	StorageSize string `json:"storageSize,omitempty"`

	// SSHSecret is the name of existing secret that contains SSH credentials to clone repositories.
	// This secret generally contains keys such as `id_rsa` (private key) and `known_hosts`.
	SSHSecret string `json:"sshSecret,omitempty"`

	// Env defines environment variables for Git Server.
	Env map[string]string `json:"env,omitempty"`
}

// IndexedSearchSpec defines the desired state of Index Search.
type IndexedSearchSpec struct {
	// Replicas defines the number of Index Search pod replicas.
	// Default: 1
	Replicas int32 `json:"replicas,omitempty"`

	// StorageSize defines the requested amount of storage for the PVC.
	// Default: 200Gi
	StorageSize string `json:"storageSize,omitempty"`

	// Resources allows for custom resource limits and requests.
	Resources *corev1.ResourceList `json:"resources,omitempty"`

	// Env defines environment variables for Indexed Search.
	Env map[string]string `json:"env,omitempty"`
}

// IndexedSearchIndexerSpec defines the desired state of the Index Search Indexer.
type IndexedSearchIndexerSpec struct {
	// Resources allows for custom resource limits and requests.
	Resources *corev1.ResourceList `json:"resources,omitempty"`

	// Env defines environment variables for Indexed Search Indexer.
	Env map[string]string `json:"env,omitempty"`
}

// PGSQLSpec defines the desired state of the PostgreSQL server.
type PGSQLSpec struct {
	// Disabled defines if pgsql is enabled or not.
	// Default: false
	Disabled bool `json:"disabled,omitempty"`

	// ExistingSecret is the name of an existing secret to use for Postgres credentials.
	ExistingSecret string `json:"existingSecret,omitempty"`

	// Database allows for custom database connection details.
	Database *DatabaseSpec `json:"database,omitempty"`

	// StorageSize defines the requested amount of storage for the PVC.
	// Default: 200Gi
	StorageSize string `json:"storageSize,omitempty"`

	// Resources allows for custom resource limits and requests.
	Resources *corev1.ResourceList `json:"resources,omitempty"`

	// Env defines environment variables for PGSQL.
	Env map[string]string `json:"env,omitempty"`
}

type PostgresExporterSpec struct {
	// Resources allows for custom resource limits and requests.
	Resources *corev1.ResourceList `json:"resources,omitempty"`

	// Env defines environment variables for Postgres Exporter.
	Env map[string]string `json:"env,omitempty"`
}

type PreciseCodeIntelSpec struct {
	// Replicas defines the number of Precise Code Intel Worker pod replicas.
	// Default: 2
	Replicas int32 `json:"replicas,omitempty"`

	// Resources allows for custom resource limits and requests.
	Resources *corev1.ResourceList `json:"resources,omitempty"`

	// Env defines environment variables for Code Intel.
	Env map[string]string `json:"env,omitempty"`
}

// RedisCacheSpec defines the desired state of the Redis cache service.
type RedisCacheSpec struct {
	// Disabled defines if Redis Cache is enabled or not.
	// Default: false
	Disabled bool `json:"disabled,omitempty"`

	// ExistingSecret is the name of an existing secret to use.
	ExistingSecret string `json:"existingSecret,omitempty"`

	// StorageSize defines the requested amount of storage for the PVC.
	// Default: 100Gi
	StorageSize string `json:"storageSize,omitempty"`

	// Resources allows for custom resource limits and requests.
	Resources *corev1.ResourceList `json:"resources,omitempty"`

	// Env defines environment variables for Redis Cache.
	Env map[string]string `json:"env,omitempty"`
}

// RedisStoreSpec defines the desired state of the Redis Store service.
type RedisStoreSpec struct {
	// Disabled defines if Redis Store is enabled or not.
	// Default: false
	Disabled bool `json:"disabled,omitempty"`

	// ExistingSecret is the name of an existing secret to use.
	ExistingSecret string `json:"existingSecret,omitempty"`

	// StorageSize defines the requested amount of storage for the PVC.
	// Default: 100Gi
	StorageSize string `json:"storageSize,omitempty"`

	// Resources allows for custom resource limits and requests.
	Resources *corev1.ResourceList `json:"resources,omitempty"`

	// Env defines environment variables for Redis Store.
	Env map[string]string `json:"env,omitempty"`
}

// RepoUpdaterSpec defines the desired state of the Repo Updater service.
type RepoUpdaterSpec struct {
	config.StandardConfig

	// Env defines environment variables for Redis Updater.
	Env map[string]string `json:"env,omitempty"`
}

// SearcherSpec defines the desired state of the Searcher service.
type SearcherSpec struct {
	// Disabled defines if Code Intel is enabled or not.
	// Default: false
	Disabled bool `json:"disabled,omitempty"`

	// Replicas defines the number of Searcher pod replicas.
	// Default: 1
	Replicas int32 `json:"replicas,omitempty"`

	// StorageSize defines the requested amount of storage for the PVC.
	// Default: 26Gi
	StorageSize string `json:"storageSize,omitempty"`

	// Resources allows for custom resource limits and requests.
	Resources *corev1.ResourceList `json:"resources,omitempty"`

	// Env defines environment variables for Searcher.
	Env map[string]string `json:"env,omitempty"`
}

// SymbolsSpec defines the desired state of the Symbols service.
type SymbolsSpec struct {
	config.StandardConfig

	// Replicas defines the number of Symbols pod replicas.
	// Default: 1
	Replicas int32 `json:"replicas,omitempty"`

	// StorageSize defines the requested amount of storage for the PVC.
	// Default: 12Gi
	StorageSize string `json:"storageSize,omitempty"`

	// Env defines environment variables for Symbols.
	Env map[string]string `json:"env,omitempty"`
}

// SyntectServerSpec defines the desired state of the Syntect server service.
type SyntectServerSpec struct {
	// Replicas defines the number of Syntect Server pod replicas.
	// Default: 1
	Replicas int32 `json:"replicas,omitempty"`

	// Resources allows for custom resource limits and requests.
	Resources *corev1.ResourceList `json:"resources,omitempty"`

	// Env defines environment variables for Syntect Server.
	Env map[string]string `json:"env,omitempty"`
}

type WorkerSpec struct {
	// Replicas defines the number of Worker pod replicas.
	// Default: 1
	Replicas int32 `json:"replicas,omitempty"`

	// Resources allows for custom resource limits and requests.
	Resources *corev1.ResourceList `json:"resources,omitempty"`

	// Env defines environment variables for Worker.
	Env map[string]string `json:"env,omitempty"`
}

type StorageClassSpec struct {
	// Name is the name of the storageClass.
	// Default: sourcegraph
	Name string `json:"name,omitempty"`

	// Create will enable/disable the creation of storageClass.
	// Enable if you have your own existing storage class.
	// Default: false
	Create bool `json:"create,omitempty"`

	// Provisioner is the storageClass provisioner.
	// Default: kubernetes.io/no-provisioner
	Provisioner string `json:"provisioner,omitempty"`

	// Type is the `type` key in storageClass `parameters`.
	Type string `json:"type,omitempty"`

	// Parameters defines any extra parameters of StorageClass.
	Parameters map[string]string `json:"parameters,omitempty"`
}

// SourcegraphSpec defines the desired state of Sourcegraph
type SourcegraphSpec struct {
	// RequestedVersion is the user-requested version of Sourcegraph to deploy.
	RequestedVersion string `json:"requestedVersion"`

	// ImageRepository overrides the default image repository.
	ImageRepository string `json:"imageRepository"`

	// ManagementState defines if Sourcegraph should be managed by the operator or not.
	// Default is managed.
	ManagementState ManagementStateType `json:"managementState,omitempty"`

	// MaintenancePassword will set the password for the administrator maintenance UI.
	// If no password is set, a random password will be generated and storage in a secret.
	MaintenancePassword string `json:"maintenancePassword,omitempty"`

	// Blobstore defines the desired state of the Blobstore service.
	Blobstore BlobstoreSpec `json:"blobstore,omitempty"`

	// CodeInsights defines the desired state of the Code Insights service.
	CodeInsights CodeInsightsDBSpec `json:"codeInsights,omitempty"`

	// CodeIntel defines the desired state of the Code Intel service.
	CodeIntel CodeIntelDBSpec `json:"codeIntel,omitempty"`

	// Frontend defines the desired state of the Sourcegraph Frontend.
	Frontend FrontendSpec `json:"frontend,omitempty"`

	// GitServer defines the desired state of the GitServer service.
	GitServer GitServerSpec `json:"gitServer,omitempty"`

	// IndexedSearch defines the desired state of the Indexed Search service.
	IndexedSearch IndexedSearchSpec `json:"indexedSearch,omitempty"`

	// IndexedSearchIndexer defines the desired state of the Indexed Search Indexer service.
	IndexedSearchIndexer IndexedSearchIndexerSpec `json:"indexedSearchIndexer,omitempty"`

	// PGSQL defines the desired state of the PostgreSQL database.
	PGSQL PGSQLSpec `json:"pgsql,omitempty"`

	// PostgresExporter defines the desired state of the Postgres exporter service.
	PostgresExporter PostgresExporterSpec `json:"postgresExporter,omitempty"`

	// PreciseCodeIntel defines the desired state of the Precise Code Intel service.
	PreciseCodeIntel PreciseCodeIntelSpec `json:"preciseCodeIntel,omitempty"`

	// RedisCache defines the desired state of the Redis cache service.
	RedisCache RedisCacheSpec `json:"redisCache,omitempty"`

	// RedisStore defines the desired state of the Redis store service.
	RedisStore RedisStoreSpec `json:"redisStore,omitempty"`

	// RepoUpdater defines the desired state of the Repo updater service.
	RepoUpdater RepoUpdaterSpec `json:"repoUpdater,omitempty"`

	// Searcher defines the desired state of the Searcher service.
	Searcher SearcherSpec `json:"searcher,omitempty"`

	// Symbols defines the desired state of the Symbols service.
	Symbols SymbolsSpec `json:"symbols,omitempty"`

	// SyntectServer defines the desired state of the Syntect Server service.
	SyntectServer SyntectServerSpec `json:"syntectServer,omitempty"`

	// Worker defines the desired state of the Worker service.
	Worker WorkerSpec `json:"worker,omitempty"`

	// StorageClass defines the desired state a custom storage class.
	// If none is specified, default cluster storage class will be used.
	StorageClass StorageClassSpec `json:"storageClass,omitempty"`
}

// SourcegraphStatus defines the observed state of Sourcegraph
type SourcegraphStatus struct {
	// CurrentVersion is the version of Sourcegraph currently running.
	CurrentVersion string `json:"currentVersion"`

	// Represents the latest available observations of Sourcegraph's current state.
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// Sourcegraph is the Schema for the Sourcegraph API
type Sourcegraph struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SourcegraphSpec   `json:"spec,omitempty"`
	Status SourcegraphStatus `json:"status,omitempty"`
}

// SourcegraphList contains a list of Sourcegraph
type SourcegraphList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Sourcegraph `json:"items"`
}
