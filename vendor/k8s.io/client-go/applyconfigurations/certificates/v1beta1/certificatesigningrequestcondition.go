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
	v1beta1 "k8s.io/api/certificates/v1beta1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CertificateSigningRequestConditionApplyConfiguration represents an declarative configuration of the CertificateSigningRequestCondition type for use
// with apply.
type CertificateSigningRequestConditionApplyConfiguration struct {
	Type               *v1beta1.RequestConditionType `json:"type,omitempty"`
	Status             *v1.ConditionStatus           `json:"status,omitempty"`
	Reason             *string                       `json:"reason,omitempty"`
	Message            *string                       `json:"message,omitempty"`
	LastUpdateTime     *metav1.Time                  `json:"lastUpdateTime,omitempty"`
	LastTransitionTime *metav1.Time                  `json:"lastTransitionTime,omitempty"`
}

// CertificateSigningRequestConditionApplyConfiguration constructs an declarative configuration of the CertificateSigningRequestCondition type for use with
// apply.
func CertificateSigningRequestCondition() *CertificateSigningRequestConditionApplyConfiguration {
	return &CertificateSigningRequestConditionApplyConfiguration{}
}

// WithType sets the Type field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Type field is set to the value of the last call.
func (b *CertificateSigningRequestConditionApplyConfiguration) WithType(value v1beta1.RequestConditionType) *CertificateSigningRequestConditionApplyConfiguration {
	b.Type = &value
	return b
}

// WithStatus sets the Status field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Status field is set to the value of the last call.
func (b *CertificateSigningRequestConditionApplyConfiguration) WithStatus(value v1.ConditionStatus) *CertificateSigningRequestConditionApplyConfiguration {
	b.Status = &value
	return b
}

// WithReason sets the Reason field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Reason field is set to the value of the last call.
func (b *CertificateSigningRequestConditionApplyConfiguration) WithReason(value string) *CertificateSigningRequestConditionApplyConfiguration {
	b.Reason = &value
	return b
}

// WithMessage sets the Message field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Message field is set to the value of the last call.
func (b *CertificateSigningRequestConditionApplyConfiguration) WithMessage(value string) *CertificateSigningRequestConditionApplyConfiguration {
	b.Message = &value
	return b
}

// WithLastUpdateTime sets the LastUpdateTime field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the LastUpdateTime field is set to the value of the last call.
func (b *CertificateSigningRequestConditionApplyConfiguration) WithLastUpdateTime(value metav1.Time) *CertificateSigningRequestConditionApplyConfiguration {
	b.LastUpdateTime = &value
	return b
}

// WithLastTransitionTime sets the LastTransitionTime field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the LastTransitionTime field is set to the value of the last call.
func (b *CertificateSigningRequestConditionApplyConfiguration) WithLastTransitionTime(value metav1.Time) *CertificateSigningRequestConditionApplyConfiguration {
	b.LastTransitionTime = &value
	return b
}
