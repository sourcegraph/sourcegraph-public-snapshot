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

// Code generated by lister-gen. DO NOT EDIT.

package v1alpha2

import (
	v1alpha2 "k8s.io/api/resource/v1alpha2"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// ResourceClaimTemplateLister helps list ResourceClaimTemplates.
// All objects returned here must be treated as read-only.
type ResourceClaimTemplateLister interface {
	// List lists all ResourceClaimTemplates in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha2.ResourceClaimTemplate, err error)
	// ResourceClaimTemplates returns an object that can list and get ResourceClaimTemplates.
	ResourceClaimTemplates(namespace string) ResourceClaimTemplateNamespaceLister
	ResourceClaimTemplateListerExpansion
}

// resourceClaimTemplateLister implements the ResourceClaimTemplateLister interface.
type resourceClaimTemplateLister struct {
	indexer cache.Indexer
}

// NewResourceClaimTemplateLister returns a new ResourceClaimTemplateLister.
func NewResourceClaimTemplateLister(indexer cache.Indexer) ResourceClaimTemplateLister {
	return &resourceClaimTemplateLister{indexer: indexer}
}

// List lists all ResourceClaimTemplates in the indexer.
func (s *resourceClaimTemplateLister) List(selector labels.Selector) (ret []*v1alpha2.ResourceClaimTemplate, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha2.ResourceClaimTemplate))
	})
	return ret, err
}

// ResourceClaimTemplates returns an object that can list and get ResourceClaimTemplates.
func (s *resourceClaimTemplateLister) ResourceClaimTemplates(namespace string) ResourceClaimTemplateNamespaceLister {
	return resourceClaimTemplateNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// ResourceClaimTemplateNamespaceLister helps list and get ResourceClaimTemplates.
// All objects returned here must be treated as read-only.
type ResourceClaimTemplateNamespaceLister interface {
	// List lists all ResourceClaimTemplates in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha2.ResourceClaimTemplate, err error)
	// Get retrieves the ResourceClaimTemplate from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha2.ResourceClaimTemplate, error)
	ResourceClaimTemplateNamespaceListerExpansion
}

// resourceClaimTemplateNamespaceLister implements the ResourceClaimTemplateNamespaceLister
// interface.
type resourceClaimTemplateNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all ResourceClaimTemplates in the indexer for a given namespace.
func (s resourceClaimTemplateNamespaceLister) List(selector labels.Selector) (ret []*v1alpha2.ResourceClaimTemplate, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha2.ResourceClaimTemplate))
	})
	return ret, err
}

// Get retrieves the ResourceClaimTemplate from the indexer for a given namespace and name.
func (s resourceClaimTemplateNamespaceLister) Get(name string) (*v1alpha2.ResourceClaimTemplate, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha2.Resource("resourceclaimtemplate"), name)
	}
	return obj.(*v1alpha2.ResourceClaimTemplate), nil
}
