package command_test

import (
	"context"
	"os"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/worker/command"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestKubernetesCommand_CreateJob(t *testing.T) {
	mockKubeInterface := command.NewMockInterface()
	mockBatchInterface := command.NewMockBatchV1Interface()
	mockJobInterface := command.NewMockJobInterface()

	cmd := &command.KubernetesCommand{
		Logger:    logtest.Scoped(t),
		Clientset: mockKubeInterface,
	}

	mockKubeInterface.BatchV1Func.PushReturn(mockBatchInterface)
	mockBatchInterface.JobsFunc.PushReturn(mockJobInterface)

	job := &batchv1.Job{}

	_, err := cmd.CreateJob(context.Background(), "my-namespace", job)
	require.NoError(t, err)

	require.Len(t, mockKubeInterface.BatchV1Func.History(), 1)
	require.Len(t, mockBatchInterface.JobsFunc.History(), 1)
	assert.Equal(t, "my-namespace", mockBatchInterface.JobsFunc.History()[0].Arg0)
	require.Len(t, mockJobInterface.CreateFunc.History(), 1)
	assert.NotNil(t, mockJobInterface.CreateFunc.History()[0].Arg0)
	assert.NotNil(t, mockJobInterface.CreateFunc.History()[0].Arg1)
}

func TestKubernetesCommand_DeleteJob(t *testing.T) {
	mockKubeInterface := command.NewMockInterface()
	mockBatchInterface := command.NewMockBatchV1Interface()
	mockJobInterface := command.NewMockJobInterface()

	cmd := &command.KubernetesCommand{
		Logger:    logtest.Scoped(t),
		Clientset: mockKubeInterface,
	}

	mockKubeInterface.BatchV1Func.PushReturn(mockBatchInterface)
	mockBatchInterface.JobsFunc.PushReturn(mockJobInterface)

	err := cmd.DeleteJob(context.Background(), "my-namespace", "my-job")
	require.NoError(t, err)

	require.Len(t, mockKubeInterface.BatchV1Func.History(), 1)
	require.Len(t, mockBatchInterface.JobsFunc.History(), 1)
	assert.Equal(t, "my-namespace", mockBatchInterface.JobsFunc.History()[0].Arg0)
	require.Len(t, mockJobInterface.DeleteFunc.History(), 1)
	assert.NotNil(t, mockJobInterface.DeleteFunc.History()[0].Arg0)
	assert.Equal(t, "my-job", mockJobInterface.DeleteFunc.History()[0].Arg1)
}

func TestKubernetesCommand_WaitForPodToStart(t *testing.T) {
	tests := []struct {
		name            string
		mockFunc        func(podInterface *command.MockPodInterface)
		mockAssertFunc  func(t *testing.T, podInterface *command.MockPodInterface)
		expectedPodName string
		expectedErr     error
	}{
		{
			name: "Pod running",
			mockFunc: func(podInterface *command.MockPodInterface) {
				podInterface.ListFunc.PushReturn(&corev1.PodList{Items: []corev1.Pod{{
					ObjectMeta: metav1.ObjectMeta{Name: "my-pod"},
					Status:     corev1.PodStatus{Phase: corev1.PodRunning},
				}}}, nil)
			},
			mockAssertFunc: func(t *testing.T, podInterface *command.MockPodInterface) {
				require.Len(t, podInterface.ListFunc.History(), 1)
				assert.Equal(t, metav1.ListOptions{LabelSelector: "job-name=my-pod"}, podInterface.ListFunc.History()[0].Arg1)
				require.Len(t, podInterface.GetFunc.History(), 0)
			},
			expectedPodName: "my-pod",
		},
		{
			name: "Pod succeeded",
			mockFunc: func(podInterface *command.MockPodInterface) {
				podInterface.ListFunc.PushReturn(&corev1.PodList{Items: []corev1.Pod{{
					ObjectMeta: metav1.ObjectMeta{Name: "my-pod"},
					Status:     corev1.PodStatus{Phase: corev1.PodSucceeded},
				}}}, nil)
			},
			mockAssertFunc: func(t *testing.T, podInterface *command.MockPodInterface) {
				require.Len(t, podInterface.ListFunc.History(), 1)
			},
			expectedPodName: "my-pod",
		},
		{
			name: "Pod Failed",
			mockFunc: func(podInterface *command.MockPodInterface) {
				podInterface.ListFunc.PushReturn(&corev1.PodList{Items: []corev1.Pod{{
					ObjectMeta: metav1.ObjectMeta{Name: "my-pod"},
					Status:     corev1.PodStatus{Phase: corev1.PodFailed},
				}}}, nil)
			},
			mockAssertFunc: func(t *testing.T, podInterface *command.MockPodInterface) {
				require.Len(t, podInterface.ListFunc.History(), 1)
			},
			expectedPodName: "my-pod",
		},
		{
			name: "Pod container running",
			mockFunc: func(podInterface *command.MockPodInterface) {
				podInterface.ListFunc.PushReturn(&corev1.PodList{Items: []corev1.Pod{{
					ObjectMeta: metav1.ObjectMeta{Name: "my-pod"},
					Status: corev1.PodStatus{
						Phase: corev1.PodPending,
						ContainerStatuses: []corev1.ContainerStatus{{
							State: corev1.ContainerState{Running: &corev1.ContainerStateRunning{}},
						}},
					},
				}}}, nil)
			},
			mockAssertFunc: func(t *testing.T, podInterface *command.MockPodInterface) {
				require.Len(t, podInterface.ListFunc.History(), 1)
			},
			expectedPodName: "my-pod",
		},
		{
			name: "Pod container waiting then running",
			mockFunc: func(podInterface *command.MockPodInterface) {
				podInterface.ListFunc.PushReturn(&corev1.PodList{Items: []corev1.Pod{{
					ObjectMeta: metav1.ObjectMeta{Name: "my-pod"},
					Status: corev1.PodStatus{
						Phase: corev1.PodPending,
						ContainerStatuses: []corev1.ContainerStatus{{
							State: corev1.ContainerState{Waiting: &corev1.ContainerStateWaiting{}},
						}},
					},
				}}}, nil)
				podInterface.GetFunc.PushReturn(&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{Name: "my-pod"},
					Status: corev1.PodStatus{
						Phase: corev1.PodPending,
						ContainerStatuses: []corev1.ContainerStatus{{
							State: corev1.ContainerState{Running: &corev1.ContainerStateRunning{}},
						}},
					},
				}, nil)
			},
			mockAssertFunc: func(t *testing.T, podInterface *command.MockPodInterface) {
				require.Len(t, podInterface.ListFunc.History(), 1)
				require.Len(t, podInterface.GetFunc.History(), 1)
				assert.Equal(t, "my-pod", podInterface.GetFunc.History()[0].Arg1)
			},
			expectedPodName: "my-pod",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockKubeInterface := command.NewMockInterface()
			mockCoreV1Interface := command.NewMockCoreV1Interface()
			mockPodInterface := command.NewMockPodInterface()

			mockKubeInterface.CoreV1Func.SetDefaultReturn(mockCoreV1Interface)
			mockCoreV1Interface.PodsFunc.SetDefaultReturn(mockPodInterface)

			if test.mockFunc != nil {
				test.mockFunc(mockPodInterface)
			}

			cmd := &command.KubernetesCommand{
				Logger:    logtest.Scoped(t),
				Clientset: mockKubeInterface,
			}

			name, err := cmd.WaitForPodToStart(context.Background(), "my-namespace", "my-pod")
			if test.expectedErr != nil {
				require.Error(t, err)
				assert.EqualError(t, err, test.expectedErr.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expectedPodName, name)
			}

			if test.mockAssertFunc != nil {
				test.mockAssertFunc(t, mockPodInterface)
			}
		})
	}
}

func TestKubernetesCommand_FindPod(t *testing.T) {
	tests := []struct {
		name           string
		mockFunc       func(podInterface *command.MockPodInterface)
		mockAssertFunc func(t *testing.T, podInterface *command.MockPodInterface)
		expectedErr    error
	}{
		{
			name: "Pod found",
			mockFunc: func(podInterface *command.MockPodInterface) {
				podInterface.ListFunc.PushReturn(&corev1.PodList{Items: []corev1.Pod{{
					ObjectMeta: metav1.ObjectMeta{Name: "my-pod"},
				}}}, nil)
			},
			mockAssertFunc: func(t *testing.T, podInterface *command.MockPodInterface) {
				require.Len(t, podInterface.ListFunc.History(), 1)
				assert.Equal(t, metav1.ListOptions{
					LabelSelector: "job-name=my-pod",
				}, podInterface.ListFunc.History()[0].Arg1)
			},
		},
		{
			name: "Pod not found",
			mockFunc: func(podInterface *command.MockPodInterface) {
				podInterface.ListFunc.PushReturn(&corev1.PodList{Items: nil}, nil)
			},
			expectedErr: errors.New("no pods found for job my-pod"),
		},
		{
			name: "Error occurred",
			mockFunc: func(podInterface *command.MockPodInterface) {
				podInterface.ListFunc.PushReturn(nil, errors.New("failed"))
			},
			expectedErr: errors.New("failed"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockKubeInterface := command.NewMockInterface()
			mockCoreV1Interface := command.NewMockCoreV1Interface()
			mockPodInterface := command.NewMockPodInterface()

			mockKubeInterface.CoreV1Func.PushReturn(mockCoreV1Interface)
			mockCoreV1Interface.PodsFunc.PushReturn(mockPodInterface)

			if test.mockFunc != nil {
				test.mockFunc(mockPodInterface)
			}

			cmd := &command.KubernetesCommand{
				Logger:    logtest.Scoped(t),
				Clientset: mockKubeInterface,
			}

			_, err := cmd.FindPod(context.Background(), "my-namespace", "my-pod")
			if test.expectedErr != nil {
				require.Error(t, err)
				assert.EqualError(t, err, test.expectedErr.Error())
			} else {
				require.NoError(t, err)
			}

			if test.mockAssertFunc != nil {
				test.mockAssertFunc(t, mockPodInterface)
			}
		})
	}
}

func TestKubernetesCommand_WaitForJobToComplete(t *testing.T) {
	tests := []struct {
		name           string
		mockFunc       func(jobInterface *command.MockJobInterface)
		mockAssertFunc func(t *testing.T, jobInterface *command.MockJobInterface)
		expectedErr    error
	}{
		{
			name: "Job succeeded",
			mockFunc: func(jobInterface *command.MockJobInterface) {
				jobInterface.GetFunc.PushReturn(&batchv1.Job{Status: batchv1.JobStatus{Active: 0, Succeeded: 1}}, nil)
			},
			mockAssertFunc: func(t *testing.T, jobInterface *command.MockJobInterface) {
				require.Len(t, jobInterface.GetFunc.History(), 1)
				assert.Equal(t, "my-job", jobInterface.GetFunc.History()[0].Arg1)
			},
		},
		{
			name: "Job failed",
			mockFunc: func(jobInterface *command.MockJobInterface) {
				jobInterface.GetFunc.PushReturn(&batchv1.Job{Status: batchv1.JobStatus{Failed: 1}}, nil)
			},
			expectedErr: errors.New("job my-job failed"),
		},
		{
			name: "Error occurred",
			mockFunc: func(jobInterface *command.MockJobInterface) {
				jobInterface.GetFunc.PushReturn(nil, errors.New("failed"))
			},
			expectedErr: errors.New("retrieving job: failed"),
		},
		{
			name: "Job succeeded second try",
			mockFunc: func(jobInterface *command.MockJobInterface) {
				jobInterface.GetFunc.PushReturn(&batchv1.Job{Status: batchv1.JobStatus{Active: 0, Succeeded: 0}}, nil)
				jobInterface.GetFunc.PushReturn(&batchv1.Job{Status: batchv1.JobStatus{Active: 0, Succeeded: 1}}, nil)
			},
			mockAssertFunc: func(t *testing.T, jobInterface *command.MockJobInterface) {
				require.Len(t, jobInterface.GetFunc.History(), 2)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockKubeInterface := command.NewMockInterface()
			mockBatchInterface := command.NewMockBatchV1Interface()
			mockJobInterface := command.NewMockJobInterface()

			mockKubeInterface.BatchV1Func.SetDefaultReturn(mockBatchInterface)
			mockBatchInterface.JobsFunc.SetDefaultReturn(mockJobInterface)

			if test.mockFunc != nil {
				test.mockFunc(mockJobInterface)
			}

			cmd := &command.KubernetesCommand{
				Logger:    logtest.Scoped(t),
				Clientset: mockKubeInterface,
			}

			err := cmd.WaitForJobToComplete(context.Background(), "my-namespace", "my-job")
			if test.expectedErr != nil {
				require.Error(t, err)
				assert.EqualError(t, err, test.expectedErr.Error())
			} else {
				require.NoError(t, err)
			}

			if test.mockAssertFunc != nil {
				test.mockAssertFunc(t, mockJobInterface)
			}
		})
	}
}

func TestNewKubernetesJob(t *testing.T) {
	err := os.Setenv("KUBERNETES_SERVICE_HOST", "http://localhost")
	require.NoError(t, err)
	t.Cleanup(func() {
		os.Unsetenv("KUBERNETES_SERVICE_HOST")
	})

	spec := command.Spec{
		Command: []string{"echo", "hello"},
		Env:     []string{"FOO=bar"},
	}
	options := command.KubernetesContainerOptions{
		Namespace:             "default",
		NodeName:              "my-node",
		PersistenceVolumeName: "my-pvc",
		ResourceLimit: command.KubernetesResource{
			CPU:    resource.MustParse("10"),
			Memory: resource.MustParse("10Gi"),
		},
		ResourceRequest: command.KubernetesResource{
			CPU:    resource.MustParse("1"),
			Memory: resource.MustParse("1Gi"),
		},
	}
	job := command.NewKubernetesJob("my-job", "my-image:latest", spec, "/my/path", options)

	assert.Equal(t, "my-job", job.Name)

	assert.Equal(t, "my-node", job.Spec.Template.Spec.NodeName)
	assert.Equal(t, corev1.RestartPolicyNever, job.Spec.Template.Spec.RestartPolicy)

	require.Len(t, job.Spec.Template.Spec.Containers, 1)
	assert.Equal(t, "job-container", job.Spec.Template.Spec.Containers[0].Name)
	assert.Equal(t, "my-image:latest", job.Spec.Template.Spec.Containers[0].Image)
	assert.Equal(t, []string{"echo", "hello"}, job.Spec.Template.Spec.Containers[0].Command)
	assert.Equal(t, "/data", job.Spec.Template.Spec.Containers[0].WorkingDir)

	require.Len(t, job.Spec.Template.Spec.Containers[0].Env, 1)
	assert.Equal(t, "FOO", job.Spec.Template.Spec.Containers[0].Env[0].Name)
	assert.Equal(t, "bar", job.Spec.Template.Spec.Containers[0].Env[0].Value)

	assert.Equal(t, resource.MustParse("10"), *job.Spec.Template.Spec.Containers[0].Resources.Limits.Cpu())
	assert.Equal(t, resource.MustParse("10Gi"), *job.Spec.Template.Spec.Containers[0].Resources.Limits.Memory())
	assert.Equal(t, resource.MustParse("1"), *job.Spec.Template.Spec.Containers[0].Resources.Requests.Cpu())
	assert.Equal(t, resource.MustParse("1Gi"), *job.Spec.Template.Spec.Containers[0].Resources.Requests.Memory())

	require.Len(t, job.Spec.Template.Spec.Containers[0].VolumeMounts, 1)
	assert.Equal(t, "job-volume", job.Spec.Template.Spec.Containers[0].VolumeMounts[0].Name)
	assert.Equal(t, "/data", job.Spec.Template.Spec.Containers[0].VolumeMounts[0].MountPath)
	assert.Equal(t, "/my/path", job.Spec.Template.Spec.Containers[0].VolumeMounts[0].SubPath)

	require.Len(t, job.Spec.Template.Spec.Volumes, 1)
	assert.Equal(t, "job-volume", job.Spec.Template.Spec.Volumes[0].Name)
	assert.Equal(t, "my-pvc", job.Spec.Template.Spec.Volumes[0].PersistentVolumeClaim.ClaimName)
}
