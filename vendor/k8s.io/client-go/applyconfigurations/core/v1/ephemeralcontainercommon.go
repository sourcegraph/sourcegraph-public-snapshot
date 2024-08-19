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

import (
	corev1 "k8s.io/api/core/v1"
)

// EphemeralContainerCommonApplyConfiguration represents an declarative configuration of the EphemeralContainerCommon type for use
// with apply.
type EphemeralContainerCommonApplyConfiguration struct {
	Name                     *string                                   `json:"name,omitempty"`
	Image                    *string                                   `json:"image,omitempty"`
	Command                  []string                                  `json:"command,omitempty"`
	Args                     []string                                  `json:"args,omitempty"`
	WorkingDir               *string                                   `json:"workingDir,omitempty"`
	Ports                    []ContainerPortApplyConfiguration         `json:"ports,omitempty"`
	EnvFrom                  []EnvFromSourceApplyConfiguration         `json:"envFrom,omitempty"`
	Env                      []EnvVarApplyConfiguration                `json:"env,omitempty"`
	Resources                *ResourceRequirementsApplyConfiguration   `json:"resources,omitempty"`
	ResizePolicy             []ContainerResizePolicyApplyConfiguration `json:"resizePolicy,omitempty"`
	RestartPolicy            *corev1.ContainerRestartPolicy            `json:"restartPolicy,omitempty"`
	VolumeMounts             []VolumeMountApplyConfiguration           `json:"volumeMounts,omitempty"`
	VolumeDevices            []VolumeDeviceApplyConfiguration          `json:"volumeDevices,omitempty"`
	LivenessProbe            *ProbeApplyConfiguration                  `json:"livenessProbe,omitempty"`
	ReadinessProbe           *ProbeApplyConfiguration                  `json:"readinessProbe,omitempty"`
	StartupProbe             *ProbeApplyConfiguration                  `json:"startupProbe,omitempty"`
	Lifecycle                *LifecycleApplyConfiguration              `json:"lifecycle,omitempty"`
	TerminationMessagePath   *string                                   `json:"terminationMessagePath,omitempty"`
	TerminationMessagePolicy *corev1.TerminationMessagePolicy          `json:"terminationMessagePolicy,omitempty"`
	ImagePullPolicy          *corev1.PullPolicy                        `json:"imagePullPolicy,omitempty"`
	SecurityContext          *SecurityContextApplyConfiguration        `json:"securityContext,omitempty"`
	Stdin                    *bool                                     `json:"stdin,omitempty"`
	StdinOnce                *bool                                     `json:"stdinOnce,omitempty"`
	TTY                      *bool                                     `json:"tty,omitempty"`
}

// EphemeralContainerCommonApplyConfiguration constructs an declarative configuration of the EphemeralContainerCommon type for use with
// apply.
func EphemeralContainerCommon() *EphemeralContainerCommonApplyConfiguration {
	return &EphemeralContainerCommonApplyConfiguration{}
}

// WithName sets the Name field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Name field is set to the value of the last call.
func (b *EphemeralContainerCommonApplyConfiguration) WithName(value string) *EphemeralContainerCommonApplyConfiguration {
	b.Name = &value
	return b
}

// WithImage sets the Image field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Image field is set to the value of the last call.
func (b *EphemeralContainerCommonApplyConfiguration) WithImage(value string) *EphemeralContainerCommonApplyConfiguration {
	b.Image = &value
	return b
}

// WithCommand adds the given value to the Command field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the Command field.
func (b *EphemeralContainerCommonApplyConfiguration) WithCommand(values ...string) *EphemeralContainerCommonApplyConfiguration {
	for i := range values {
		b.Command = append(b.Command, values[i])
	}
	return b
}

// WithArgs adds the given value to the Args field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the Args field.
func (b *EphemeralContainerCommonApplyConfiguration) WithArgs(values ...string) *EphemeralContainerCommonApplyConfiguration {
	for i := range values {
		b.Args = append(b.Args, values[i])
	}
	return b
}

// WithWorkingDir sets the WorkingDir field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the WorkingDir field is set to the value of the last call.
func (b *EphemeralContainerCommonApplyConfiguration) WithWorkingDir(value string) *EphemeralContainerCommonApplyConfiguration {
	b.WorkingDir = &value
	return b
}

// WithPorts adds the given value to the Ports field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the Ports field.
func (b *EphemeralContainerCommonApplyConfiguration) WithPorts(values ...*ContainerPortApplyConfiguration) *EphemeralContainerCommonApplyConfiguration {
	for i := range values {
		if values[i] == nil {
			panic("nil value passed to WithPorts")
		}
		b.Ports = append(b.Ports, *values[i])
	}
	return b
}

// WithEnvFrom adds the given value to the EnvFrom field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the EnvFrom field.
func (b *EphemeralContainerCommonApplyConfiguration) WithEnvFrom(values ...*EnvFromSourceApplyConfiguration) *EphemeralContainerCommonApplyConfiguration {
	for i := range values {
		if values[i] == nil {
			panic("nil value passed to WithEnvFrom")
		}
		b.EnvFrom = append(b.EnvFrom, *values[i])
	}
	return b
}

// WithEnv adds the given value to the Env field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the Env field.
func (b *EphemeralContainerCommonApplyConfiguration) WithEnv(values ...*EnvVarApplyConfiguration) *EphemeralContainerCommonApplyConfiguration {
	for i := range values {
		if values[i] == nil {
			panic("nil value passed to WithEnv")
		}
		b.Env = append(b.Env, *values[i])
	}
	return b
}

// WithResources sets the Resources field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Resources field is set to the value of the last call.
func (b *EphemeralContainerCommonApplyConfiguration) WithResources(value *ResourceRequirementsApplyConfiguration) *EphemeralContainerCommonApplyConfiguration {
	b.Resources = value
	return b
}

// WithResizePolicy adds the given value to the ResizePolicy field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the ResizePolicy field.
func (b *EphemeralContainerCommonApplyConfiguration) WithResizePolicy(values ...*ContainerResizePolicyApplyConfiguration) *EphemeralContainerCommonApplyConfiguration {
	for i := range values {
		if values[i] == nil {
			panic("nil value passed to WithResizePolicy")
		}
		b.ResizePolicy = append(b.ResizePolicy, *values[i])
	}
	return b
}

// WithRestartPolicy sets the RestartPolicy field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the RestartPolicy field is set to the value of the last call.
func (b *EphemeralContainerCommonApplyConfiguration) WithRestartPolicy(value corev1.ContainerRestartPolicy) *EphemeralContainerCommonApplyConfiguration {
	b.RestartPolicy = &value
	return b
}

// WithVolumeMounts adds the given value to the VolumeMounts field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the VolumeMounts field.
func (b *EphemeralContainerCommonApplyConfiguration) WithVolumeMounts(values ...*VolumeMountApplyConfiguration) *EphemeralContainerCommonApplyConfiguration {
	for i := range values {
		if values[i] == nil {
			panic("nil value passed to WithVolumeMounts")
		}
		b.VolumeMounts = append(b.VolumeMounts, *values[i])
	}
	return b
}

// WithVolumeDevices adds the given value to the VolumeDevices field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the VolumeDevices field.
func (b *EphemeralContainerCommonApplyConfiguration) WithVolumeDevices(values ...*VolumeDeviceApplyConfiguration) *EphemeralContainerCommonApplyConfiguration {
	for i := range values {
		if values[i] == nil {
			panic("nil value passed to WithVolumeDevices")
		}
		b.VolumeDevices = append(b.VolumeDevices, *values[i])
	}
	return b
}

// WithLivenessProbe sets the LivenessProbe field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the LivenessProbe field is set to the value of the last call.
func (b *EphemeralContainerCommonApplyConfiguration) WithLivenessProbe(value *ProbeApplyConfiguration) *EphemeralContainerCommonApplyConfiguration {
	b.LivenessProbe = value
	return b
}

// WithReadinessProbe sets the ReadinessProbe field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the ReadinessProbe field is set to the value of the last call.
func (b *EphemeralContainerCommonApplyConfiguration) WithReadinessProbe(value *ProbeApplyConfiguration) *EphemeralContainerCommonApplyConfiguration {
	b.ReadinessProbe = value
	return b
}

// WithStartupProbe sets the StartupProbe field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the StartupProbe field is set to the value of the last call.
func (b *EphemeralContainerCommonApplyConfiguration) WithStartupProbe(value *ProbeApplyConfiguration) *EphemeralContainerCommonApplyConfiguration {
	b.StartupProbe = value
	return b
}

// WithLifecycle sets the Lifecycle field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Lifecycle field is set to the value of the last call.
func (b *EphemeralContainerCommonApplyConfiguration) WithLifecycle(value *LifecycleApplyConfiguration) *EphemeralContainerCommonApplyConfiguration {
	b.Lifecycle = value
	return b
}

// WithTerminationMessagePath sets the TerminationMessagePath field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the TerminationMessagePath field is set to the value of the last call.
func (b *EphemeralContainerCommonApplyConfiguration) WithTerminationMessagePath(value string) *EphemeralContainerCommonApplyConfiguration {
	b.TerminationMessagePath = &value
	return b
}

// WithTerminationMessagePolicy sets the TerminationMessagePolicy field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the TerminationMessagePolicy field is set to the value of the last call.
func (b *EphemeralContainerCommonApplyConfiguration) WithTerminationMessagePolicy(value corev1.TerminationMessagePolicy) *EphemeralContainerCommonApplyConfiguration {
	b.TerminationMessagePolicy = &value
	return b
}

// WithImagePullPolicy sets the ImagePullPolicy field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the ImagePullPolicy field is set to the value of the last call.
func (b *EphemeralContainerCommonApplyConfiguration) WithImagePullPolicy(value corev1.PullPolicy) *EphemeralContainerCommonApplyConfiguration {
	b.ImagePullPolicy = &value
	return b
}

// WithSecurityContext sets the SecurityContext field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the SecurityContext field is set to the value of the last call.
func (b *EphemeralContainerCommonApplyConfiguration) WithSecurityContext(value *SecurityContextApplyConfiguration) *EphemeralContainerCommonApplyConfiguration {
	b.SecurityContext = value
	return b
}

// WithStdin sets the Stdin field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Stdin field is set to the value of the last call.
func (b *EphemeralContainerCommonApplyConfiguration) WithStdin(value bool) *EphemeralContainerCommonApplyConfiguration {
	b.Stdin = &value
	return b
}

// WithStdinOnce sets the StdinOnce field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the StdinOnce field is set to the value of the last call.
func (b *EphemeralContainerCommonApplyConfiguration) WithStdinOnce(value bool) *EphemeralContainerCommonApplyConfiguration {
	b.StdinOnce = &value
	return b
}

// WithTTY sets the TTY field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the TTY field is set to the value of the last call.
func (b *EphemeralContainerCommonApplyConfiguration) WithTTY(value bool) *EphemeralContainerCommonApplyConfiguration {
	b.TTY = &value
	return b
}
