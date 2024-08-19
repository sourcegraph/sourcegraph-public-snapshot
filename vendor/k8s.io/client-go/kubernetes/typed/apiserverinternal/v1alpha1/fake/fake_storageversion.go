/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"
	json "encoding/json"
	"fmt"

	v1alpha1 "k8s.io/api/apiserverinternal/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	apiserverinternalv1alpha1 "k8s.io/client-go/applyconfigurations/apiserverinternal/v1alpha1"
	testing "k8s.io/client-go/testing"
)

// FakeStorageVersions implements StorageVersionInterface
type FakeStorageVersions struct {
	Fake *FakeInternalV1alpha1
}

var storageversionsResource = v1alpha1.SchemeGroupVersion.WithResource("storageversions")

var storageversionsKind = v1alpha1.SchemeGroupVersion.WithKind("StorageVersion")

// Get takes name of the storageVersion, and returns the corresponding storageVersion object, and an error if there is any.
func (c *FakeStorageVersions) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.StorageVersion, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootGetAction(storageversionsResource, name), &v1alpha1.StorageVersion{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.StorageVersion), err
}

// List takes label and field selectors, and returns the list of StorageVersions that match those selectors.
func (c *FakeStorageVersions) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.StorageVersionList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootListAction(storageversionsResource, storageversionsKind, opts), &v1alpha1.StorageVersionList{})
	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.StorageVersionList{ListMeta: obj.(*v1alpha1.StorageVersionList).ListMeta}
	for _, item := range obj.(*v1alpha1.StorageVersionList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested storageVersions.
func (c *FakeStorageVersions) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchAction(storageversionsResource, opts))
}

// Create takes the representation of a storageVersion and creates it.  Returns the server's representation of the storageVersion, and an error, if there is any.
func (c *FakeStorageVersions) Create(ctx context.Context, storageVersion *v1alpha1.StorageVersion, opts v1.CreateOptions) (result *v1alpha1.StorageVersion, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateAction(storageversionsResource, storageVersion), &v1alpha1.StorageVersion{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.StorageVersion), err
}

// Update takes the representation of a storageVersion and updates it. Returns the server's representation of the storageVersion, and an error, if there is any.
func (c *FakeStorageVersions) Update(ctx context.Context, storageVersion *v1alpha1.StorageVersion, opts v1.UpdateOptions) (result *v1alpha1.StorageVersion, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateAction(storageversionsResource, storageVersion), &v1alpha1.StorageVersion{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.StorageVersion), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeStorageVersions) UpdateStatus(ctx context.Context, storageVersion *v1alpha1.StorageVersion, opts v1.UpdateOptions) (*v1alpha1.StorageVersion, error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateSubresourceAction(storageversionsResource, "status", storageVersion), &v1alpha1.StorageVersion{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.StorageVersion), err
}

// Delete takes name of the storageVersion and deletes it. Returns an error if one occurs.
func (c *FakeStorageVersions) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteActionWithOptions(storageversionsResource, name, opts), &v1alpha1.StorageVersion{})
	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeStorageVersions) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewRootDeleteCollectionAction(storageversionsResource, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.StorageVersionList{})
	return err
}

// Patch applies the patch and returns the patched storageVersion.
func (c *FakeStorageVersions) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.StorageVersion, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceAction(storageversionsResource, name, pt, data, subresources...), &v1alpha1.StorageVersion{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.StorageVersion), err
}

// Apply takes the given apply declarative configuration, applies it and returns the applied storageVersion.
func (c *FakeStorageVersions) Apply(ctx context.Context, storageVersion *apiserverinternalv1alpha1.StorageVersionApplyConfiguration, opts v1.ApplyOptions) (result *v1alpha1.StorageVersion, err error) {
	if storageVersion == nil {
		return nil, fmt.Errorf("storageVersion provided to Apply must not be nil")
	}
	data, err := json.Marshal(storageVersion)
	if err != nil {
		return nil, err
	}
	name := storageVersion.Name
	if name == nil {
		return nil, fmt.Errorf("storageVersion.Name must be provided to Apply")
	}
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceAction(storageversionsResource, *name, types.ApplyPatchType, data), &v1alpha1.StorageVersion{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.StorageVersion), err
}

// ApplyStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating ApplyStatus().
func (c *FakeStorageVersions) ApplyStatus(ctx context.Context, storageVersion *apiserverinternalv1alpha1.StorageVersionApplyConfiguration, opts v1.ApplyOptions) (result *v1alpha1.StorageVersion, err error) {
	if storageVersion == nil {
		return nil, fmt.Errorf("storageVersion provided to Apply must not be nil")
	}
	data, err := json.Marshal(storageVersion)
	if err != nil {
		return nil, err
	}
	name := storageVersion.Name
	if name == nil {
		return nil, fmt.Errorf("storageVersion.Name must be provided to Apply")
	}
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceAction(storageversionsResource, *name, types.ApplyPatchType, data, "status"), &v1alpha1.StorageVersion{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.StorageVersion), err
}
