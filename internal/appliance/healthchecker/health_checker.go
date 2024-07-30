package healthchecker

import (
	"context"
	"time"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Probe interface {
	CheckPods(ctx context.Context, labelSelector, namespace string) error
}

type HealthChecker struct {
	Probe     Probe
	K8sClient client.Client
	Logger    log.Logger

	ServiceName client.ObjectKey
	Interval    time.Duration
	Graceperiod time.Duration
}

// ManageIngressFacingService waits for the `begin` channel to close, then periodically monitors the frontend
// service (the ingress-facing service). When there is at least one ready
// frontend pod, it ensures that the service points at the frontend pods. When
// there are no ready pods, it ensures that the service points to the appliance,
// so that the admin can log in and view maintenance status.
func (h *HealthChecker) ManageIngressFacingService(ctx context.Context, begin <-chan struct{}, labelSelector, namespace string) error {
	h.Logger.Info("waiting for signal to begin managing ingress-facing service for the appliance")
	select {
	case <-begin:
		// block

	case <-ctx.Done():
		h.Logger.Error("context done, exiting", log.Error(ctx.Err()))
		return ctx.Err()
	}

	h.Logger.Info("will periodically check health of frontend and re-point ingress appropriately")

	ticker := time.NewTicker(h.Interval)
	defer ticker.Stop()

	// Do one iteration without having to wait for the first tick
	if err := h.maybeFlipServiceOnce(ctx, labelSelector, namespace); err != nil {
		return err
	}
	for {
		select {
		case <-ticker.C:
			if err := h.maybeFlipServiceOnce(ctx, labelSelector, namespace); err != nil {
				return err
			}

		case <-ctx.Done():
			h.Logger.Error("context done, exiting", log.Error(ctx.Err()))
			return ctx.Err()
		}
	}
}

func (h *HealthChecker) maybeFlipServiceOnce(ctx context.Context, labelSelector, namespace string) error {
	h.Logger.Info("checking deployment health")
	if err := h.Probe.CheckPods(ctx, labelSelector, namespace); err != nil {
		h.Logger.Error("found unhealthy state, waiting for the grace period", log.Error(err), log.String("gracePeriod", h.Graceperiod.String()))
		time.Sleep(h.Graceperiod)
		if err := h.Probe.CheckPods(ctx, labelSelector, namespace); err != nil {
			h.Logger.Error("found unhealthy state, setting service selector to appliance", log.Error(err))
			return h.setServiceSelector(ctx, "sourcegraph-appliance-frontend")
		}
	}

	h.Logger.Info("deployment healthy")
	return h.setServiceSelector(ctx, "sourcegraph-frontend")
}

func (h *HealthChecker) setServiceSelector(ctx context.Context, to string) error {
	h.Logger.Info("setting service selector", log.String("to", to))

	var svc corev1.Service
	if err := h.K8sClient.Get(ctx, h.ServiceName, &svc); err != nil {
		h.Logger.Error("getting service", log.Error(err))
		return errors.Wrap(err, "getting service")
	}

	// no-op if the selector is unchanged
	svc.Spec.Selector["app"] = to
	return h.K8sClient.Update(ctx, &svc)
}
