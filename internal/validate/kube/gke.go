package kube

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/sourcegraph/src-cli/internal/validate"
)

type ClusterInfo struct {
	ServiceType string
	ProjectId   string
	Region      string
	ClusterName string
}

func Gke() Option {
	return func(config *Config) {
		config.gke = true
	}
}

func GkeGcePersistentDiskCSIDrivers(ctx context.Context, config *Config) ([]validate.Result, error) {
	var results []validate.Result

	checkStorageClassesResults, err := validateStorageClasses(ctx, config)
	if err != nil {
		results = append(results, validate.Result{
			Status:  validate.Failure,
			Message: "GKE: could not check StorageClasses",
		})
		return results, nil
	}

	results = append(results, checkStorageClassesResults...)
	return results, nil
}

// validateStorageClasses checks for GKE specific storageClasses:
//
// After the compute engine persistent disk CSI driver is enabled,
// gke automatically installs the standard-rwo and the premium-rwo
// storage classes. This function checks that those storage
// classes exist on the cluster.
//
// Ref: shorturl.at/dnKV0
func validateStorageClasses(ctx context.Context, config *Config) ([]validate.Result, error) {
	var results []validate.Result

	storageClient := config.clientSet.StorageV1()
	storageClasses, err := storageClient.StorageClasses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	classes := 0
	for _, item := range storageClasses.Items {
		if item.Name == "premium-rwo" || item.Name == "standard-rwo" {
			classes += 1
		}
	}

	if classes == 2 {
		results = append(results, validate.Result{
			Status:  validate.Success,
			Message: "persistent volumes enabled: validated",
		})

		return results, nil
	}

	results = append(results, validate.Result{
		Status:  validate.Failure,
		Message: "validate persistent volumes enabled: failed",
	})

	return results, nil
}
