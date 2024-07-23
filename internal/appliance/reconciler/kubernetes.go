package reconciler

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	rbacv1 "k8s.io/api/rbac/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/sourcegraph/sourcegraph/internal/appliance/config"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Create, update, or delete a Kubernetes object. If the service config doesn't
// conform to the Disableable interface, users will have to orchestrate their
// own deletion logic. They should be able to use the lower-level
// createOrUpdateObject() directly.
func reconcileObject[T client.Object](
	ctx context.Context, r *Reconciler,
	cfg config.Disableable,
	obj, objKind T,
	sg *config.Sourcegraph, owner client.Object,
) error {
	if cfg.IsDisabled() {
		return ensureObjectDeleted(ctx, r, owner, obj)
	}

	updateIfChanged := struct {
		Cfg     config.Disableable
		Version string
	}{
		Cfg:     cfg,
		Version: sg.Spec.RequestedVersion,
	}

	return createOrUpdateObject(ctx, r, updateIfChanged, owner, obj, objKind)
}

// Upsert a Kubernetes object.
//
// obj is the object you want to reconcile, updating an existing cluster object
// if it has changed, or creating it if none existed before.
//
// objKind should be the same type as obj, usually an instantiated
// struct-pointer to a particular Kubernetes object type, e.g.
// `&appsv1.Deployment{}`. It is used to hold data about any existing object of
// the same name, to compare it to obj, and possibly be replaced by obj.
//
// updateIfChanged is the object whose hash we store in an annotation to
// determine whether an existing in-cluster object is out of date and needs to
// be replaced.
//
// Takes the reconciler as a parameter rather than being a method on it due to
// limitations of Go generics.
func createOrUpdateObject[R client.Object](
	ctx context.Context, r *Reconciler, updateIfChanged any,
	owner client.Object, obj, objKind R,
) error {
	gvk, err := apiutil.GVKForObject(obj, r.Scheme)
	if err != nil {
		return errors.Wrap(err, "getting GVK for object")
	}
	logger := log.FromContext(ctx).WithValues("kind", gvk.String(), "namespace", obj.GetNamespace(), "name", obj.GetName())
	namespacedName := types.NamespacedName{Namespace: obj.GetNamespace(), Name: obj.GetName()}

	cfgHash, err := configHash(updateIfChanged)
	if err != nil {
		return err
	}
	annotations := obj.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}
	annotations[config.AnnotationKeyConfigHash] = cfgHash
	obj.SetAnnotations(annotations)

	// Namespaced objects can't own non-namespaced objects. Trying to
	// SetControllerReference on cluster-scoped resources gives the following
	// error: "cluster-scoped resource must not have a namespace-scoped owner".
	// non-namespaced resources will therefore not be garbage-collected when the
	// ConfigMap is deleted.
	if isNamespaced(obj) {
		if err := ctrl.SetControllerReference(owner, obj, r.Scheme); err != nil {
			return errors.Newf("setting controller reference: %w", err)
		}
	}

	existingRes := objKind
	if err := r.Client.Get(ctx, namespacedName, existingRes); err != nil {
		if kerrors.IsNotFound(err) {
			logger.Info("didn't find existing object, creating it")
			if err := r.Client.Create(ctx, obj); err != nil {
				logger.Error(err, "error creating object")
				return err
			}
			return nil
		}

		logger.Error(err, "unexpected error getting object")
		return err
	}

	if !isControlledBy(owner, existingRes) && isNamespaced(obj) && !config.ShouldAdopt(obj) {
		logger.Info("refusing to update non-owned resource")
		return nil
	}

	if cfgHash != existingRes.GetAnnotations()[config.AnnotationKeyConfigHash] {
		logger.Info("Found existing object with spec that does not match desired state. Clobbering it.")
		if err := r.Client.Update(ctx, obj); err != nil {
			logger.Error(err, "error updating object")
			return err
		}
		return nil
	}

	logger.Info("Found existing object with spec that matches the desired state. Will do nothing.")
	return nil
}

func isNamespaced(obj client.Object) bool {
	if _, ok := obj.(*rbacv1.ClusterRole); ok {
		return false
	}
	if _, ok := obj.(*rbacv1.ClusterRoleBinding); ok {
		return false
	}
	return true
}

func ensureObjectDeleted[T client.Object](ctx context.Context, r *Reconciler, owner client.Object, obj T) error {
	// We need to try to get the object first, in order to check its owner
	// references later.
	objKey := types.NamespacedName{Name: obj.GetName(), Namespace: obj.GetNamespace()}
	if err := r.Client.Get(ctx, objKey, obj); err != nil {
		if kerrors.IsNotFound(err) {
			// Object doesn't exist, we don't need to delete it
			return nil
		}
	}
	gvk, err := apiutil.GVKForObject(obj, r.Scheme)
	if err != nil {
		return errors.Wrap(err, "getting GVK for object")
	}

	logger := log.FromContext(ctx).WithValues("kind", gvk.String(), "namespace", obj.GetNamespace(), "name", obj.GetName())

	if !isControlledBy(owner, obj) && isNamespaced(obj) {
		logger.Info("refusing to delete non-owned resource")
		return nil
	}

	logger.Info("deleting resource")
	if err := r.Client.Delete(ctx, obj); err != nil {
		if kerrors.IsNotFound(err) {
			// If by chance it got deleted concurrently, no harm done.
			return nil
		}

		logger.Error(err, "unexpected error deleting resource")
		return err
	}
	return nil
}

func isControlledBy(owner, obj client.Object) bool {
	for _, ownerRef := range obj.GetOwnerReferences() {
		if owner.GetUID() == ownerRef.UID {
			return true
		}
	}
	return false
}

func configHash(configElement any) (string, error) {
	cfgBytes, err := json.Marshal(configElement)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(cfgBytes)
	return hex.EncodeToString(hash[:]), nil
}
