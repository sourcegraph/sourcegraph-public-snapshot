package kube

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/sourcegraph/src-cli/internal/validate"
)

func TestValidatePod(t *testing.T) {
	cases := []struct {
		name   string
		pod    func(pod *corev1.Pod)
		result []validate.Result
	}{
		{
			name: "valid pod",
		},
		{
			name: "invalid pod: pod name is not set",
			pod: func(pod *corev1.Pod) {
				pod.Name = ""
			},
			result: []validate.Result{
				{
					Status:  validate.Failure,
					Message: "pod.Name is empty",
				},
			},
		},
		{
			name: "invalid pod: pod namespace is empty",
			pod: func(pod *corev1.Pod) {
				pod.Namespace = ""
			},
			result: []validate.Result{
				{
					Status:  validate.Failure,
					Message: "pod.Namespace is empty",
				},
			},
		},
		{
			name: "invalid pod: spec containers is empty",
			pod: func(pod *corev1.Pod) {
				pod.Spec.Containers = nil
			},
			result: []validate.Result{
				{
					Status:  validate.Failure,
					Message: "spec.Containers is empty",
				},
			},
		},
		{
			name: "invalid pod: pod status is pending",
			pod: func(pod *corev1.Pod) {
				pod.Status.Phase = corev1.PodPending
			},
			result: []validate.Result{
				{
					Status:  validate.Failure,
					Message: "pod 'sourcegraph-frontend-' has a status 'pending'",
				},
			},
		},
		{
			name: "invalid pod: pod status failed",
			pod: func(pod *corev1.Pod) {
				pod.Status.Phase = corev1.PodFailed
			},
			result: []validate.Result{
				{
					Status:  validate.Failure,
					Message: "pod 'sourcegraph-frontend-' has a status 'failed'",
				},
			},
		},
		{
			name: "invalid pod: container name is empty",
			pod: func(pod *corev1.Pod) {
				pod.Spec.Containers[0].Name = ""
			},
			result: []validate.Result{
				{
					Status:  validate.Failure,
					Message: "container.Name is empty, pod 'sourcegraph-frontend-'",
				},
			},
		},
		{
			name: "invalid pod: container image is empty",
			pod: func(pod *corev1.Pod) {
				pod.Spec.Containers[0].Image = ""
			},
			result: []validate.Result{
				{
					Status:  validate.Failure,
					Message: "container.Image is empty, pod 'sourcegraph-frontend-'",
				},
			},
		},
		{
			name: "invalid pod: image is not set",
			pod: func(pod *corev1.Pod) {
				pod.Spec.Containers[0].Image = ""
			},
			result: []validate.Result{
				{
					Status:  validate.Failure,
					Message: "container.Image is empty, pod 'sourcegraph-frontend-'",
				},
			},
		},
		{
			name: "invalid pod: container status not ready",
			pod: func(pod *corev1.Pod) {
				pod.Status.ContainerStatuses[0].Ready = false
			},
			result: []validate.Result{
				{
					Status:  validate.Failure,
					Message: "container 'sourcegraph-test-id' is not ready, pod 'sourcegraph-frontend-'",
				},
			},
		},
		{
			name: "invalid pod: container restart count is high",
			pod: func(pod *corev1.Pod) {
				pod.Status.ContainerStatuses[0].RestartCount = 100
			},
			result: []validate.Result{
				{
					Status:  validate.Warning,
					Message: "container 'sourcegraph-test-id' has high restart count: 100 restarts",
				},
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			pod := testPod()
			if tc.pod != nil {
				tc.pod(pod)
			}
			result := validatePod(pod)

			// test should error
			if len(tc.result) > 0 {
				if result == nil {
					t.Fatal("validate should return result")
					return
				}
				if result[0].Status != tc.result[0].Status {
					t.Errorf("result status\nwant: %v\n got: %v", tc.result[0].Status, result[0].Status)
				}
				if result[0].Message != tc.result[0].Message {
					t.Errorf("result msg\nwant: %s\n got: %s", tc.result[0].Message, result[0].Message)
				}
				return
			}

			// test should not error
			if result != nil {
				t.Fatalf("ValidatePod error: %v", result)
			}
		})
	}
}

func TestValidateService(t *testing.T) {
	cases := []struct {
		name    string
		service func(service *corev1.Service)
		result  []validate.Result
	}{
		{
			name: "valid service",
		},
		{
			name: "invalid service: service name is not set",
			service: func(service *corev1.Service) {
				service.Name = ""
			},
			result: []validate.Result{
				{
					Status:  validate.Failure,
					Message: "service.Name is empty",
				},
			},
		},
		{
			name: "invalid service: service namespace is not set",
			service: func(service *corev1.Service) {
				service.Namespace = ""
			},
			result: []validate.Result{
				{
					Status:  validate.Failure,
					Message: "service.Namespace is empty",
				},
			},
		},
		{
			name: "invalid service: service ports is empty",
			service: func(service *corev1.Service) {
				service.Spec.Ports = nil
			},
			result: []validate.Result{
				{
					Status:  validate.Failure,
					Message: "service.Ports is empty",
				},
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			service := testService()
			if tc.service != nil {
				tc.service(service)
			}
			result := validateService(service)

			// test should error
			if len(tc.result) > 0 {
				if result == nil {
					t.Fatal("validate should return result")
					return
				}
				if result[0].Status != tc.result[0].Status {
					t.Errorf("result status\nwant: %v\n got: %v", tc.result[0].Status, result[0].Status)
				}
				if result[0].Message != tc.result[0].Message {
					t.Errorf("result msg\nwant: %s\n got: %s", tc.result[0].Message, result[0].Message)
				}
				return
			}

			// test should not error
			if result != nil {
				t.Fatalf("ValidateService error: %v", result)
			}
		})
	}
}

func TestValidatePVC(t *testing.T) {
	cases := []struct {
		name   string
		pvc    func(pvc *corev1.PersistentVolumeClaim)
		result []validate.Result
	}{
		{
			name: "valid pvc",
		},
		{
			name: "invalid pvc: status not bound",
			pvc: func(pvc *corev1.PersistentVolumeClaim) {
				pvc.Status.Phase = "Waiting"
			},
			result: []validate.Result{
				{
					Status:  validate.Failure,
					Message: "pvc.Status is not bound",
				},
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			pvc := testPVC()
			if tc.pvc != nil {
				tc.pvc(pvc)
			}
			result := validatePVC(pvc)

			// test should error
			if len(tc.result) > 0 {
				if result == nil {
					t.Fatal("validate should return result")
					return
				}
				if result[0].Status != tc.result[0].Status {
					t.Errorf("result status\nwant: %v\n got: %v", tc.result[0].Status, result[0].Status)
				}
				if result[0].Message != tc.result[0].Message {
					t.Errorf("result msg\nwant: %s\n got: %s", tc.result[0].Message, result[0].Message)
				}
				return
			}

			// test should not error
			if result != nil {
				t.Fatalf("ValidateService error: %v", result)
			}
		})
	}
}

// helper test function to return a valid pod
func testPod() *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "sourcegraph-frontend-",
			Labels: map[string]string{
				"deploy": "sourcegraph",
			},
			Annotations: map[string]string{
				"kubectl.kubernetes.io/default-container": "frontend",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "sourcegraph-frontend",
					Image: "sourcegraph/foo:test",
					Ports: []corev1.ContainerPort{
						{
							ContainerPort: 8800,
							Protocol:      corev1.ProtocolTCP,
						},
						{
							ContainerPort: 3090,
							Protocol:      corev1.ProtocolTCP,
						},
						{
							ContainerPort: 6060,
							Protocol:      corev1.ProtocolTCP,
						},
					},
					Args: []string{"serve"},
				},
			},
		},
		Status: corev1.PodStatus{
			ContainerStatuses: []corev1.ContainerStatus{
				{
					ContainerID:  "sourcegraph-test-id",
					Ready:        true,
					RestartCount: 0,
				},
			},
		},
	}
}

// helper test function to return a valid service
func testService() *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "symbols",
			Labels: map[string]string{
				"deploy": "sourcegraph",
			},
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{
					Name:     "http",
					Port:     3184,
					Protocol: corev1.ProtocolTCP,
				},
			},
		},
		Status: corev1.ServiceStatus{},
	}
}

// helper test function to return a valid PVC
func testPVC() *corev1.PersistentVolumeClaim {
	return &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "pgsql",
			Labels: map[string]string{
				"deploy": "sourcegraph",
			},
		},
		Status: corev1.PersistentVolumeClaimStatus{
			Phase: "Bound",
		},
	}
}
