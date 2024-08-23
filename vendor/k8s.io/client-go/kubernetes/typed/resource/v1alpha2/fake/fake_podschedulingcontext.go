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

	v1alpha2 "k8s.io/api/resource/v1alpha2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	resourcev1alpha2 "k8s.io/client-go/applyconfigurations/resource/v1alpha2"
	testing "k8s.io/client-go/testing"
)

// FakePodSchedulingContexts implements PodSchedulingContextInterface
type FakePodSchedulingContexts struct {
	Fake *FakeResourceV1alpha2
	ns   string
}

var podschedulingcontextsResource = v1alpha2.SchemeGroupVersion.WithResource("podschedulingcontexts")

var podschedulingcontextsKind = v1alpha2.SchemeGroupVersion.WithKind("PodSchedulingContext")

// Get takes name of the podSchedulingContext, and returns the corresponding podSchedulingContext object, and an error if there is any.
func (c *FakePodSchedulingContexts) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha2.PodSchedulingContext, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(podschedulingcontextsResource, c.ns, name), &v1alpha2.PodSchedulingContext{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha2.PodSchedulingContext), err
}

// List takes label and field selectors, and returns the list of PodSchedulingContexts that match those selectors.
func (c *FakePodSchedulingContexts) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha2.PodSchedulingContextList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(podschedulingcontextsResource, podschedulingcontextsKind, c.ns, opts), &v1alpha2.PodSchedulingContextList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha2.PodSchedulingContextList{ListMeta: obj.(*v1alpha2.PodSchedulingContextList).ListMeta}
	for _, item := range obj.(*v1alpha2.PodSchedulingContextList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested podSchedulingContexts.
func (c *FakePodSchedulingContexts) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(podschedulingcontextsResource, c.ns, opts))

}

// Create takes the representation of a podSchedulingContext and creates it.  Returns the server's representation of the podSchedulingContext, and an error, if there is any.
func (c *FakePodSchedulingContexts) Create(ctx context.Context, podSchedulingContext *v1alpha2.PodSchedulingContext, opts v1.CreateOptions) (result *v1alpha2.PodSchedulingContext, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(podschedulingcontextsResource, c.ns, podSchedulingContext), &v1alpha2.PodSchedulingContext{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha2.PodSchedulingContext), err
}

// Update takes the representation of a podSchedulingContext and updates it. Returns the server's representation of the podSchedulingContext, and an error, if there is any.
func (c *FakePodSchedulingContexts) Update(ctx context.Context, podSchedulingContext *v1alpha2.PodSchedulingContext, opts v1.UpdateOptions) (result *v1alpha2.PodSchedulingContext, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(podschedulingcontextsResource, c.ns, podSchedulingContext), &v1alpha2.PodSchedulingContext{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha2.PodSchedulingContext), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakePodSchedulingContexts) UpdateStatus(ctx context.Context, podSchedulingContext *v1alpha2.PodSchedulingContext, opts v1.UpdateOptions) (*v1alpha2.PodSchedulingContext, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(podschedulingcontextsResource, "status", c.ns, podSchedulingContext), &v1alpha2.PodSchedulingContext{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha2.PodSchedulingContext), err
}

// Delete takes name of the podSchedulingContext and deletes it. Returns an error if one occurs.
func (c *FakePodSchedulingContexts) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(podschedulingcontextsResource, c.ns, name, opts), &v1alpha2.PodSchedulingContext{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakePodSchedulingContexts) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(podschedulingcontextsResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha2.PodSchedulingContextList{})
	return err
}

// Patch applies the patch and returns the patched podSchedulingContext.
func (c *FakePodSchedulingContexts) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha2.PodSchedulingContext, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(podschedulingcontextsResource, c.ns, name, pt, data, subresources...), &v1alpha2.PodSchedulingContext{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha2.PodSchedulingContext), err
}

// Apply takes the given apply declarative configuration, applies it and returns the applied podSchedulingContext.
func (c *FakePodSchedulingContexts) Apply(ctx context.Context, podSchedulingContext *resourcev1alpha2.PodSchedulingContextApplyConfiguration, opts v1.ApplyOptions) (result *v1alpha2.PodSchedulingContext, err error) {
	if podSchedulingContext == nil {
		return nil, fmt.Errorf("podSchedulingContext provided to Apply must not be nil")
	}
	data, err := json.Marshal(podSchedulingContext)
	if err != nil {
		return nil, err
	}
	name := podSchedulingContext.Name
	if name == nil {
		return nil, fmt.Errorf("podSchedulingContext.Name must be provided to Apply")
	}
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(podschedulingcontextsResource, c.ns, *name, types.ApplyPatchType, data), &v1alpha2.PodSchedulingContext{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha2.PodSchedulingContext), err
}

// ApplyStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating ApplyStatus().
func (c *FakePodSchedulingContexts) ApplyStatus(ctx context.Context, podSchedulingContext *resourcev1alpha2.PodSchedulingContextApplyConfiguration, opts v1.ApplyOptions) (result *v1alpha2.PodSchedulingContext, err error) {
	if podSchedulingContext == nil {
		return nil, fmt.Errorf("podSchedulingContext provided to Apply must not be nil")
	}
	data, err := json.Marshal(podSchedulingContext)
	if err != nil {
		return nil, err
	}
	name := podSchedulingContext.Name
	if name == nil {
		return nil, fmt.Errorf("podSchedulingContext.Name must be provided to Apply")
	}
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(podschedulingcontextsResource, c.ns, *name, types.ApplyPatchType, data, "status"), &v1alpha2.PodSchedulingContext{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha2.PodSchedulingContext), err
}
