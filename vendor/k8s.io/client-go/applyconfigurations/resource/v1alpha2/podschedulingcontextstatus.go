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

package v1alpha2

// PodSchedulingContextStatusApplyConfiguration represents an declarative configuration of the PodSchedulingContextStatus type for use
// with apply.
type PodSchedulingContextStatusApplyConfiguration struct {
	ResourceClaims []ResourceClaimSchedulingStatusApplyConfiguration `json:"resourceClaims,omitempty"`
}

// PodSchedulingContextStatusApplyConfiguration constructs an declarative configuration of the PodSchedulingContextStatus type for use with
// apply.
func PodSchedulingContextStatus() *PodSchedulingContextStatusApplyConfiguration {
	return &PodSchedulingContextStatusApplyConfiguration{}
}

// WithResourceClaims adds the given value to the ResourceClaims field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the ResourceClaims field.
func (b *PodSchedulingContextStatusApplyConfiguration) WithResourceClaims(values ...*ResourceClaimSchedulingStatusApplyConfiguration) *PodSchedulingContextStatusApplyConfiguration {
	for i := range values {
		if values[i] == nil {
			panic("nil value passed to WithResourceClaims")
		}
		b.ResourceClaims = append(b.ResourceClaims, *values[i])
	}
	return b
}
