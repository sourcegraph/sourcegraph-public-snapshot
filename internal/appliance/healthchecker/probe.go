package healthchecker

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type PodProbe struct {
	K8sClient client.Client
}

func (p *PodProbe) CheckPods(ctx context.Context, labelSelector, namespace string) error {
	var pods corev1.PodList
	selector, err := labels.Parse(labelSelector)
	if err != nil {
		return errors.Wrap(err, "parsing label selector")
	}
	if err := p.K8sClient.List(ctx, &pods, &client.ListOptions{LabelSelector: selector, Namespace: namespace}); err != nil {
		return errors.Wrap(err, "listing pods")
	}
	for _, pod := range pods.Items {
		for _, condition := range pod.Status.Conditions {
			if condition.Type == corev1.PodReady {
				if condition.Status == corev1.ConditionTrue {
					// Return no error if even a single pod is ready
					return nil
				}
			}
		}
	}

	return errors.New("no pods are ready")
}
