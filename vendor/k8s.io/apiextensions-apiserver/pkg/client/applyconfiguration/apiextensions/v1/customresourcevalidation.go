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

package v1

// CustomResourceValidationApplyConfiguration represents an declarative configuration of the CustomResourceValidation type for use
// with apply.
type CustomResourceValidationApplyConfiguration struct {
	OpenAPIV3Schema *JSONSchemaPropsApplyConfiguration `json:"openAPIV3Schema,omitempty"`
}

// CustomResourceValidationApplyConfiguration constructs an declarative configuration of the CustomResourceValidation type for use with
// apply.
func CustomResourceValidation() *CustomResourceValidationApplyConfiguration {
	return &CustomResourceValidationApplyConfiguration{}
}

// WithOpenAPIV3Schema sets the OpenAPIV3Schema field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the OpenAPIV3Schema field is set to the value of the last call.
func (b *CustomResourceValidationApplyConfiguration) WithOpenAPIV3Schema(value *JSONSchemaPropsApplyConfiguration) *CustomResourceValidationApplyConfiguration {
	b.OpenAPIV3Schema = value
	return b
}
