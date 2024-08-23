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

package v1

import (
	"context"
	json "encoding/json"
	"fmt"
	"time"

	v1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	discoveryv1 "k8s.io/client-go/applyconfigurations/discovery/v1"
	scheme "k8s.io/client-go/kubernetes/scheme"
	rest "k8s.io/client-go/rest"
)

// EndpointSlicesGetter has a method to return a EndpointSliceInterface.
// A group's client should implement this interface.
type EndpointSlicesGetter interface {
	EndpointSlices(namespace string) EndpointSliceInterface
}

// EndpointSliceInterface has methods to work with EndpointSlice resources.
type EndpointSliceInterface interface {
	Create(ctx context.Context, endpointSlice *v1.EndpointSlice, opts metav1.CreateOptions) (*v1.EndpointSlice, error)
	Update(ctx context.Context, endpointSlice *v1.EndpointSlice, opts metav1.UpdateOptions) (*v1.EndpointSlice, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1.EndpointSlice, error)
	List(ctx context.Context, opts metav1.ListOptions) (*v1.EndpointSliceList, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.EndpointSlice, err error)
	Apply(ctx context.Context, endpointSlice *discoveryv1.EndpointSliceApplyConfiguration, opts metav1.ApplyOptions) (result *v1.EndpointSlice, err error)
	EndpointSliceExpansion
}

// endpointSlices implements EndpointSliceInterface
type endpointSlices struct {
	client rest.Interface
	ns     string
}

// newEndpointSlices returns a EndpointSlices
func newEndpointSlices(c *DiscoveryV1Client, namespace string) *endpointSlices {
	return &endpointSlices{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the endpointSlice, and returns the corresponding endpointSlice object, and an error if there is any.
func (c *endpointSlices) Get(ctx context.Context, name string, options metav1.GetOptions) (result *v1.EndpointSlice, err error) {
	result = &v1.EndpointSlice{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("endpointslices").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of EndpointSlices that match those selectors.
func (c *endpointSlices) List(ctx context.Context, opts metav1.ListOptions) (result *v1.EndpointSliceList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1.EndpointSliceList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("endpointslices").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested endpointSlices.
func (c *endpointSlices) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("endpointslices").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a endpointSlice and creates it.  Returns the server's representation of the endpointSlice, and an error, if there is any.
func (c *endpointSlices) Create(ctx context.Context, endpointSlice *v1.EndpointSlice, opts metav1.CreateOptions) (result *v1.EndpointSlice, err error) {
	result = &v1.EndpointSlice{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("endpointslices").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(endpointSlice).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a endpointSlice and updates it. Returns the server's representation of the endpointSlice, and an error, if there is any.
func (c *endpointSlices) Update(ctx context.Context, endpointSlice *v1.EndpointSlice, opts metav1.UpdateOptions) (result *v1.EndpointSlice, err error) {
	result = &v1.EndpointSlice{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("endpointslices").
		Name(endpointSlice.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(endpointSlice).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the endpointSlice and deletes it. Returns an error if one occurs.
func (c *endpointSlices) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("endpointslices").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *endpointSlices) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("endpointslices").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched endpointSlice.
func (c *endpointSlices) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.EndpointSlice, err error) {
	result = &v1.EndpointSlice{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("endpointslices").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}

// Apply takes the given apply declarative configuration, applies it and returns the applied endpointSlice.
func (c *endpointSlices) Apply(ctx context.Context, endpointSlice *discoveryv1.EndpointSliceApplyConfiguration, opts metav1.ApplyOptions) (result *v1.EndpointSlice, err error) {
	if endpointSlice == nil {
		return nil, fmt.Errorf("endpointSlice provided to Apply must not be nil")
	}
	patchOpts := opts.ToPatchOptions()
	data, err := json.Marshal(endpointSlice)
	if err != nil {
		return nil, err
	}
	name := endpointSlice.Name
	if name == nil {
		return nil, fmt.Errorf("endpointSlice.Name must be provided to Apply")
	}
	result = &v1.EndpointSlice{}
	err = c.client.Patch(types.ApplyPatchType).
		Namespace(c.ns).
		Resource("endpointslices").
		Name(*name).
		VersionedParams(&patchOpts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
