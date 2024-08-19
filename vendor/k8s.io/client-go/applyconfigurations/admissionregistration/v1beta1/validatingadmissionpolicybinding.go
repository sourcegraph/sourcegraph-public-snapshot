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

// Code generated by applyconfiguration-gen. DO NOT EDIT.

package v1beta1

import (
	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	managedfields "k8s.io/apimachinery/pkg/util/managedfields"
	internal "k8s.io/client-go/applyconfigurations/internal"
	v1 "k8s.io/client-go/applyconfigurations/meta/v1"
)

// ValidatingAdmissionPolicyBindingApplyConfiguration represents an declarative configuration of the ValidatingAdmissionPolicyBinding type for use
// with apply.
type ValidatingAdmissionPolicyBindingApplyConfiguration struct {
	v1.TypeMetaApplyConfiguration    `json:",inline"`
	*v1.ObjectMetaApplyConfiguration `json:"metadata,omitempty"`
	Spec                             *ValidatingAdmissionPolicyBindingSpecApplyConfiguration `json:"spec,omitempty"`
}

// ValidatingAdmissionPolicyBinding constructs an declarative configuration of the ValidatingAdmissionPolicyBinding type for use with
// apply.
func ValidatingAdmissionPolicyBinding(name string) *ValidatingAdmissionPolicyBindingApplyConfiguration {
	b := &ValidatingAdmissionPolicyBindingApplyConfiguration{}
	b.WithName(name)
	b.WithKind("ValidatingAdmissionPolicyBinding")
	b.WithAPIVersion("admissionregistration.k8s.io/v1beta1")
	return b
}

// ExtractValidatingAdmissionPolicyBinding extracts the applied configuration owned by fieldManager from
// validatingAdmissionPolicyBinding. If no managedFields are found in validatingAdmissionPolicyBinding for fieldManager, a
// ValidatingAdmissionPolicyBindingApplyConfiguration is returned with only the Name, Namespace (if applicable),
// APIVersion and Kind populated. It is possible that no managed fields were found for because other
// field managers have taken ownership of all the fields previously owned by fieldManager, or because
// the fieldManager never owned fields any fields.
// validatingAdmissionPolicyBinding must be a unmodified ValidatingAdmissionPolicyBinding API object that was retrieved from the Kubernetes API.
// ExtractValidatingAdmissionPolicyBinding provides a way to perform a extract/modify-in-place/apply workflow.
// Note that an extracted apply configuration will contain fewer fields than what the fieldManager previously
// applied if another fieldManager has updated or force applied any of the previously applied fields.
// Experimental!
func ExtractValidatingAdmissionPolicyBinding(validatingAdmissionPolicyBinding *admissionregistrationv1beta1.ValidatingAdmissionPolicyBinding, fieldManager string) (*ValidatingAdmissionPolicyBindingApplyConfiguration, error) {
	return extractValidatingAdmissionPolicyBinding(validatingAdmissionPolicyBinding, fieldManager, "")
}

// ExtractValidatingAdmissionPolicyBindingStatus is the same as ExtractValidatingAdmissionPolicyBinding except
// that it extracts the status subresource applied configuration.
// Experimental!
func ExtractValidatingAdmissionPolicyBindingStatus(validatingAdmissionPolicyBinding *admissionregistrationv1beta1.ValidatingAdmissionPolicyBinding, fieldManager string) (*ValidatingAdmissionPolicyBindingApplyConfiguration, error) {
	return extractValidatingAdmissionPolicyBinding(validatingAdmissionPolicyBinding, fieldManager, "status")
}

func extractValidatingAdmissionPolicyBinding(validatingAdmissionPolicyBinding *admissionregistrationv1beta1.ValidatingAdmissionPolicyBinding, fieldManager string, subresource string) (*ValidatingAdmissionPolicyBindingApplyConfiguration, error) {
	b := &ValidatingAdmissionPolicyBindingApplyConfiguration{}
	err := managedfields.ExtractInto(validatingAdmissionPolicyBinding, internal.Parser().Type("io.k8s.api.admissionregistration.v1beta1.ValidatingAdmissionPolicyBinding"), fieldManager, b, subresource)
	if err != nil {
		return nil, err
	}
	b.WithName(validatingAdmissionPolicyBinding.Name)

	b.WithKind("ValidatingAdmissionPolicyBinding")
	b.WithAPIVersion("admissionregistration.k8s.io/v1beta1")
	return b, nil
}

// WithKind sets the Kind field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Kind field is set to the value of the last call.
func (b *ValidatingAdmissionPolicyBindingApplyConfiguration) WithKind(value string) *ValidatingAdmissionPolicyBindingApplyConfiguration {
	b.Kind = &value
	return b
}

// WithAPIVersion sets the APIVersion field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the APIVersion field is set to the value of the last call.
func (b *ValidatingAdmissionPolicyBindingApplyConfiguration) WithAPIVersion(value string) *ValidatingAdmissionPolicyBindingApplyConfiguration {
	b.APIVersion = &value
	return b
}

// WithName sets the Name field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Name field is set to the value of the last call.
func (b *ValidatingAdmissionPolicyBindingApplyConfiguration) WithName(value string) *ValidatingAdmissionPolicyBindingApplyConfiguration {
	b.ensureObjectMetaApplyConfigurationExists()
	b.Name = &value
	return b
}

// WithGenerateName sets the GenerateName field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the GenerateName field is set to the value of the last call.
func (b *ValidatingAdmissionPolicyBindingApplyConfiguration) WithGenerateName(value string) *ValidatingAdmissionPolicyBindingApplyConfiguration {
	b.ensureObjectMetaApplyConfigurationExists()
	b.GenerateName = &value
	return b
}

// WithNamespace sets the Namespace field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Namespace field is set to the value of the last call.
func (b *ValidatingAdmissionPolicyBindingApplyConfiguration) WithNamespace(value string) *ValidatingAdmissionPolicyBindingApplyConfiguration {
	b.ensureObjectMetaApplyConfigurationExists()
	b.Namespace = &value
	return b
}

// WithUID sets the UID field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the UID field is set to the value of the last call.
func (b *ValidatingAdmissionPolicyBindingApplyConfiguration) WithUID(value types.UID) *ValidatingAdmissionPolicyBindingApplyConfiguration {
	b.ensureObjectMetaApplyConfigurationExists()
	b.UID = &value
	return b
}

// WithResourceVersion sets the ResourceVersion field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the ResourceVersion field is set to the value of the last call.
func (b *ValidatingAdmissionPolicyBindingApplyConfiguration) WithResourceVersion(value string) *ValidatingAdmissionPolicyBindingApplyConfiguration {
	b.ensureObjectMetaApplyConfigurationExists()
	b.ResourceVersion = &value
	return b
}

// WithGeneration sets the Generation field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Generation field is set to the value of the last call.
func (b *ValidatingAdmissionPolicyBindingApplyConfiguration) WithGeneration(value int64) *ValidatingAdmissionPolicyBindingApplyConfiguration {
	b.ensureObjectMetaApplyConfigurationExists()
	b.Generation = &value
	return b
}

// WithCreationTimestamp sets the CreationTimestamp field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the CreationTimestamp field is set to the value of the last call.
func (b *ValidatingAdmissionPolicyBindingApplyConfiguration) WithCreationTimestamp(value metav1.Time) *ValidatingAdmissionPolicyBindingApplyConfiguration {
	b.ensureObjectMetaApplyConfigurationExists()
	b.CreationTimestamp = &value
	return b
}

// WithDeletionTimestamp sets the DeletionTimestamp field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the DeletionTimestamp field is set to the value of the last call.
func (b *ValidatingAdmissionPolicyBindingApplyConfiguration) WithDeletionTimestamp(value metav1.Time) *ValidatingAdmissionPolicyBindingApplyConfiguration {
	b.ensureObjectMetaApplyConfigurationExists()
	b.DeletionTimestamp = &value
	return b
}

// WithDeletionGracePeriodSeconds sets the DeletionGracePeriodSeconds field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the DeletionGracePeriodSeconds field is set to the value of the last call.
func (b *ValidatingAdmissionPolicyBindingApplyConfiguration) WithDeletionGracePeriodSeconds(value int64) *ValidatingAdmissionPolicyBindingApplyConfiguration {
	b.ensureObjectMetaApplyConfigurationExists()
	b.DeletionGracePeriodSeconds = &value
	return b
}

// WithLabels puts the entries into the Labels field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, the entries provided by each call will be put on the Labels field,
// overwriting an existing map entries in Labels field with the same key.
func (b *ValidatingAdmissionPolicyBindingApplyConfiguration) WithLabels(entries map[string]string) *ValidatingAdmissionPolicyBindingApplyConfiguration {
	b.ensureObjectMetaApplyConfigurationExists()
	if b.Labels == nil && len(entries) > 0 {
		b.Labels = make(map[string]string, len(entries))
	}
	for k, v := range entries {
		b.Labels[k] = v
	}
	return b
}

// WithAnnotations puts the entries into the Annotations field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, the entries provided by each call will be put on the Annotations field,
// overwriting an existing map entries in Annotations field with the same key.
func (b *ValidatingAdmissionPolicyBindingApplyConfiguration) WithAnnotations(entries map[string]string) *ValidatingAdmissionPolicyBindingApplyConfiguration {
	b.ensureObjectMetaApplyConfigurationExists()
	if b.Annotations == nil && len(entries) > 0 {
		b.Annotations = make(map[string]string, len(entries))
	}
	for k, v := range entries {
		b.Annotations[k] = v
	}
	return b
}

// WithOwnerReferences adds the given value to the OwnerReferences field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the OwnerReferences field.
func (b *ValidatingAdmissionPolicyBindingApplyConfiguration) WithOwnerReferences(values ...*v1.OwnerReferenceApplyConfiguration) *ValidatingAdmissionPolicyBindingApplyConfiguration {
	b.ensureObjectMetaApplyConfigurationExists()
	for i := range values {
		if values[i] == nil {
			panic("nil value passed to WithOwnerReferences")
		}
		b.OwnerReferences = append(b.OwnerReferences, *values[i])
	}
	return b
}

// WithFinalizers adds the given value to the Finalizers field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the Finalizers field.
func (b *ValidatingAdmissionPolicyBindingApplyConfiguration) WithFinalizers(values ...string) *ValidatingAdmissionPolicyBindingApplyConfiguration {
	b.ensureObjectMetaApplyConfigurationExists()
	for i := range values {
		b.Finalizers = append(b.Finalizers, values[i])
	}
	return b
}

func (b *ValidatingAdmissionPolicyBindingApplyConfiguration) ensureObjectMetaApplyConfigurationExists() {
	if b.ObjectMetaApplyConfiguration == nil {
		b.ObjectMetaApplyConfiguration = &v1.ObjectMetaApplyConfiguration{}
	}
}

// WithSpec sets the Spec field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Spec field is set to the value of the last call.
func (b *ValidatingAdmissionPolicyBindingApplyConfiguration) WithSpec(value *ValidatingAdmissionPolicyBindingSpecApplyConfiguration) *ValidatingAdmissionPolicyBindingApplyConfiguration {
	b.Spec = value
	return b
}
