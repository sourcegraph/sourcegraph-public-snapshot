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

package v2beta1

import (
	resource "k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/client-go/applyconfigurations/meta/v1"
)

// PodsMetricStatusApplyConfiguration represents an declarative configuration of the PodsMetricStatus type for use
// with apply.
type PodsMetricStatusApplyConfiguration struct {
	MetricName          *string                             `json:"metricName,omitempty"`
	CurrentAverageValue *resource.Quantity                  `json:"currentAverageValue,omitempty"`
	Selector            *v1.LabelSelectorApplyConfiguration `json:"selector,omitempty"`
}

// PodsMetricStatusApplyConfiguration constructs an declarative configuration of the PodsMetricStatus type for use with
// apply.
func PodsMetricStatus() *PodsMetricStatusApplyConfiguration {
	return &PodsMetricStatusApplyConfiguration{}
}

// WithMetricName sets the MetricName field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the MetricName field is set to the value of the last call.
func (b *PodsMetricStatusApplyConfiguration) WithMetricName(value string) *PodsMetricStatusApplyConfiguration {
	b.MetricName = &value
	return b
}

// WithCurrentAverageValue sets the CurrentAverageValue field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the CurrentAverageValue field is set to the value of the last call.
func (b *PodsMetricStatusApplyConfiguration) WithCurrentAverageValue(value resource.Quantity) *PodsMetricStatusApplyConfiguration {
	b.CurrentAverageValue = &value
	return b
}

// WithSelector sets the Selector field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Selector field is set to the value of the last call.
func (b *PodsMetricStatusApplyConfiguration) WithSelector(value *v1.LabelSelectorApplyConfiguration) *PodsMetricStatusApplyConfiguration {
	b.Selector = value
	return b
}
