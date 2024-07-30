package appliance

import (
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Task is a task that some states may have to complete to exit.
type Task struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Started     bool      `json:"started"`
	Finished    bool      `json:"finished"`
	Weight      int       `json:"weight"`
	Progress    int       `json:"progress"`
	LastUpdate  time.Time `json:"lastUpdate"`
}

func installTasks() []Task {
	return []Task{
		{
			Title:       "Setup",
			Description: "Setting up Sourcegraph Search",
			Started:     false,
			Finished:    false,
			Weight:      25,
		},
	}
}

// IsObjectReady checks if a k8s object is ready, with the definition of ready depending on the object type.
// Supported resource types are StatefulSets, Deployments, and PersistentVolumeClaims.
func IsObjectReady(obj client.Object) (bool, error) {
	switch o := obj.(type) {
	case *appsv1.StatefulSet:
		return IsStatefulSetReady(o)
	case *appsv1.Deployment:
		return IsDeploymentReady(o)
	case *corev1.PersistentVolumeClaim:
		return IsPersistentVolumeClaimReady(o)
	default:
		return false, errors.Newf("unsupported resource type: %T", obj)
	}
}

// IsStatefulSetReady checks if a StatefulSet is ready
func IsStatefulSetReady(sts *appsv1.StatefulSet) (bool, error) {
	if sts.Status.ReadyReplicas != *sts.Spec.Replicas {
		return false, nil
	}
	// StatefulSet controller has not processed most recent changes
	if sts.Status.ObservedGeneration < sts.Generation {
		return false, nil
	}
	// StatefulSet controller has not updated all pods to the latest version
	if sts.Status.CurrentRevision != sts.Status.UpdateRevision {
		return false, nil
	}
	return true, nil
}

// IsDeploymentReady checks if a Deployment is ready
func IsDeploymentReady(deploy *appsv1.Deployment) (bool, error) {
	if deploy.Status.ReadyReplicas != *deploy.Spec.Replicas {
		return false, nil
	}
	for _, condition := range deploy.Status.Conditions {
		if condition.Type == appsv1.DeploymentAvailable && condition.Status == corev1.ConditionTrue {
			return true, nil
		}
	}
	return false, nil
}

// IsPersistentVolumeClaimReady checks if a PersistentVolumeClaim is ready
func IsPersistentVolumeClaimReady(pvc *corev1.PersistentVolumeClaim) (bool, error) {
	return pvc.Status.Phase == corev1.ClaimBound, nil
}

func calculateProgress(tasks []Task) ([]Task, int) {
	totalWeight := sumTaskWeights(tasks)
	progress := calculateTotalProgress(tasks)
	overallProgress := calculateOverallProgress(progress, totalWeight)

	return tasks, overallProgress
}

func sumTaskWeights(tasks []Task) int {
	total := 0
	for _, task := range tasks {
		total += task.Weight
	}
	return total
}

func calculateTotalProgress(tasks []Task) float32 {
	var progress float32
	for _, task := range tasks {
		progress += calculateTaskProgress(task)
	}
	return progress
}

func calculateTaskProgress(task Task) float32 {
	if task.Finished {
		return float32(task.Weight)
	}
	if task.Started {
		return float32(task.Weight) * clampProgress(task.Progress) / 100
	}
	return 0
}

func clampProgress(progress int) float32 {
	if progress > 100 {
		return 100
	}
	return float32(progress)
}

func calculateOverallProgress(progress float32, totalWeight int) int {
	return int(progress / float32(totalWeight) * 100)
}
