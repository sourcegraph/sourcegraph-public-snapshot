package kube

import (
	"context"

	"github.com/sourcegraph/src-cli/internal/validate"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Aks() Option {
	return func(config *Config) {
		config.aks = true
	}
}

func AksCsiDrivers(ctx context.Context, config *Config) ([]validate.Result, error) {
	var results []validate.Result

	storageClassResults, err := validateStorageClass(ctx, config)
	if err != nil {
		results = append(results, validate.Result{
			Status:  validate.Failure,
			Message: "AKS: could not validate if persistent volumes are enabled",
		})

		return results, err
	}

	results = append(results, storageClassResults...)

	return results, nil
}

func validateStorageClass(ctx context.Context, config *Config) ([]validate.Result, error) {
	var results []validate.Result

	storageClient := config.clientSet.StorageV1()
	storageClasses, err := storageClient.StorageClasses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, item := range storageClasses.Items {
		if item.Name == "sourcegraph" {
			if item.Provisioner != "disk.csi.azure.com" {
				results = append(results, validate.Result{
					Status:  validate.Failure,
					Message: "provisioner does not enable persistent volumes",
				})
			} else {
				results = append(results, validate.Result{
					Status:  validate.Success,
					Message: "persistent volumes enabled",
				})
			}

			if string(*item.ReclaimPolicy) != "Retain" {
				results = append(results, validate.Result{
					Status:  validate.Failure,
					Message: "storageclass has a reclaim policy other than 'Retain'",
				})
			} else {
				results = append(results, validate.Result{
					Status:  validate.Success,
					Message: "storageclass has correct reclaim policy (Retain)",
				})
			}

			if string(*item.VolumeBindingMode) != "WaitForFirstConsumer" {
				results = append(results, validate.Result{
					Status:  validate.Failure,
					Message: "storageclass has a binding mode other than 'WaitForFirstConsumer'",
				})
			} else {
				results = append(results, validate.Result{
					Status:  validate.Success,
					Message: "storageclass has correct volumeBindingMode (WaitForFirstConsumer)",
				})
			}
		}
	}

	if len(results) == 0 {
		results = append(results, validate.Result{
			Status:  validate.Warning,
			Message: "you have not yet deployed a sourcegraph instance to this cluster, or you've named your storageclass something other than 'sourcegraph'",
		})
	}

	return results, nil
}
