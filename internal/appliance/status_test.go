package appliance

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestIsResourceReady(t *testing.T) {
	tests := []struct {
		name     string
		obj      client.Object
		expected bool
		wantErr  bool
	}{
		{
			name:     "Ready StatefulSet",
			obj:      createReadyStatefulSet(),
			expected: true,
			wantErr:  false,
		},
		{
			name:     "Not Ready StatefulSet",
			obj:      createNotReadyStatefulSet(),
			expected: false,
			wantErr:  false,
		},
		{
			name:     "Ready Deployment",
			obj:      createReadyDeployment(),
			expected: true,
			wantErr:  false,
		},
		{
			name:     "Not Ready Deployment",
			obj:      createNotReadyDeployment(),
			expected: false,
			wantErr:  false,
		},
		{
			name:     "Ready PersistentVolumeClaim",
			obj:      createReadyPVC(),
			expected: true,
			wantErr:  false,
		},
		{
			name:     "Not Ready PersistentVolumeClaim",
			obj:      createNotReadyPVC(),
			expected: false,
			wantErr:  false,
		},
		{
			name:     "Unsupported Resource Type",
			obj:      &corev1.Pod{},
			expected: false,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := IsObjectReady(tt.obj)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsObjectReady() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.expected {
				t.Errorf("IsObjectReady() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func createReadyStatefulSet() *appsv1.StatefulSet {
	replicas := int32(3)
	return &appsv1.StatefulSet{
		Spec: appsv1.StatefulSetSpec{
			Replicas: &replicas,
		},
		Status: appsv1.StatefulSetStatus{
			ReadyReplicas:      3,
			ObservedGeneration: 1,
			CurrentRevision:    "rev1",
			UpdateRevision:     "rev1",
		},
	}
}

func createNotReadyStatefulSet() *appsv1.StatefulSet {
	replicas := int32(3)
	return &appsv1.StatefulSet{
		Spec: appsv1.StatefulSetSpec{
			Replicas: &replicas,
		},
		Status: appsv1.StatefulSetStatus{
			ReadyReplicas:      2,
			ObservedGeneration: 1,
			CurrentRevision:    "rev1",
			UpdateRevision:     "rev1",
		},
	}
}

func createReadyDeployment() *appsv1.Deployment {
	replicas := int32(3)
	return &appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
		},
		Status: appsv1.DeploymentStatus{
			ReadyReplicas: 3,
			Conditions: []appsv1.DeploymentCondition{
				{
					Type:   appsv1.DeploymentAvailable,
					Status: corev1.ConditionTrue,
				},
			},
		},
	}
}

func createNotReadyDeployment() *appsv1.Deployment {
	replicas := int32(3)
	return &appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
		},
		Status: appsv1.DeploymentStatus{
			ReadyReplicas: 2,
			Conditions: []appsv1.DeploymentCondition{
				{
					Type:   appsv1.DeploymentAvailable,
					Status: corev1.ConditionTrue,
				},
			},
		},
	}
}

func createReadyPVC() *corev1.PersistentVolumeClaim {
	return &corev1.PersistentVolumeClaim{
		Status: corev1.PersistentVolumeClaimStatus{
			Phase: corev1.ClaimBound,
		},
	}
}

func createNotReadyPVC() *corev1.PersistentVolumeClaim {
	return &corev1.PersistentVolumeClaim{
		Status: corev1.PersistentVolumeClaimStatus{
			Phase: corev1.ClaimPending,
		},
	}
}
