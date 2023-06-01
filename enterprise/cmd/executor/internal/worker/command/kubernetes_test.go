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
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
	"k8s.io/utils/pointer"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/worker/command"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestKubernetesCommand_CreateJob(t *testing.T) {
	clientset := fake.NewSimpleClientset()

	cmd := &command.KubernetesCommand{
		Logger:     logtest.Scoped(t),
		Clientset:  clientset,
		Operations: command.NewOperations(&observation.TestContext),
	}

	job := &batchv1.Job{}

	_, err := cmd.CreateJob(context.Background(), "my-namespace", job)
	require.NoError(t, err)

	require.Len(t, clientset.Actions(), 1)
	require.Equal(t, "create", clientset.Actions()[0].GetVerb())
	require.Equal(t, "jobs", clientset.Actions()[0].GetResource().Resource)
	require.Equal(t, "my-namespace", clientset.Actions()[0].GetNamespace())
}

func TestKubernetesCommand_DeleteJob(t *testing.T) {
	clientset := fake.NewSimpleClientset()

	cmd := &command.KubernetesCommand{
		Logger:     logtest.Scoped(t),
		Clientset:  clientset,
		Operations: command.NewOperations(&observation.TestContext),
	}

	job := &batchv1.Job{ObjectMeta: metav1.ObjectMeta{Name: "my-job"}}
	_, err := cmd.CreateJob(context.Background(), "my-namespace", job)
	require.NoError(t, err)

	err = cmd.DeleteJob(context.Background(), "my-namespace", "my-job")
	require.NoError(t, err)

	require.Len(t, clientset.Actions(), 2)
	require.Equal(t, "delete", clientset.Actions()[1].GetVerb())
	require.Equal(t, "jobs", clientset.Actions()[1].GetResource().Resource)
	assert.Equal(t, "my-namespace", clientset.Actions()[1].GetNamespace())
	assert.Equal(t, "my-job", clientset.Actions()[1].(k8stesting.DeleteAction).GetName())
}

func TestKubernetesCommand_ReadLogs(t *testing.T) {
	tests := []struct {
		name           string
		pod            *corev1.Pod
		mockFunc       func(clientset *fake.Clientset, logEntry *command.MockLogEntry)
		mockAssertFunc func(t *testing.T, actions []k8stesting.Action, logEntry *command.MockLogEntry)
		expectedErr    error
	}{
		{
			name: "Logs read",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-pod",
				},
				Status: corev1.PodStatus{
					ContainerStatuses: []corev1.ContainerStatus{
						{
							Name: command.KubernetesJobContainerName,
							State: corev1.ContainerState{
								Terminated: &corev1.ContainerStateTerminated{
									ExitCode: 0,
								},
							},
						},
					},
				},
			},
			mockFunc: func(clientset *fake.Clientset, logEntry *command.MockLogEntry) {
				clientset.PrependReactor("list", "pods", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, &corev1.PodList{Items: []corev1.Pod{
						{ObjectMeta: metav1.ObjectMeta{
							Name:   "my-pod",
							Labels: map[string]string{"job-name": "job-some-queue-42-some-key"},
						}}},
					}, nil
				})
			},
			mockAssertFunc: func(t *testing.T, actions []k8stesting.Action, logEntry *command.MockLogEntry) {
				require.Len(t, actions, 1)
				assert.Equal(t, "get", actions[0].GetVerb())
				assert.Equal(t, "pods", actions[0].GetResource().Resource)
				assert.Equal(t, "log", actions[0].GetSubresource())
				assert.Equal(t, "sg-executor-job-container", actions[0].(k8stesting.GenericAction).GetValue().(*corev1.PodLogOptions).Container)

				require.Len(t, logEntry.WriteFunc.History(), 1)
				assert.Equal(t, "stdout: fake logs\n", string(logEntry.WriteFunc.History()[0].Arg0))

				require.Len(t, logEntry.FinalizeFunc.History(), 1)
				assert.Equal(t, 0, logEntry.FinalizeFunc.History()[0].Arg0)
			},
		},
		{
			name: "Out of memory",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-pod",
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodFailed,
				},
			},
			mockAssertFunc: func(t *testing.T, actions []k8stesting.Action, logEntry *command.MockLogEntry) {
				require.Len(t, actions, 0)

				require.Len(t, logEntry.WriteFunc.History(), 0)

				require.Len(t, logEntry.FinalizeFunc.History(), 1)
				assert.Equal(t, 1, logEntry.FinalizeFunc.History()[0].Arg0)
			},
		},
		{
			name: "Bad exit code",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-pod",
				},
				Status: corev1.PodStatus{
					ContainerStatuses: []corev1.ContainerStatus{
						{
							Name: command.KubernetesJobContainerName,
							State: corev1.ContainerState{
								Terminated: &corev1.ContainerStateTerminated{
									ExitCode: 128,
								},
							},
						},
					},
				},
			},
			mockFunc: func(clientset *fake.Clientset, logEntry *command.MockLogEntry) {
				clientset.PrependReactor("list", "pods", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, &corev1.PodList{Items: []corev1.Pod{
						{ObjectMeta: metav1.ObjectMeta{
							Name:   "my-pod",
							Labels: map[string]string{"job-name": "job-some-queue-42-some-key"},
						}}},
					}, nil
				})
			},
			mockAssertFunc: func(t *testing.T, actions []k8stesting.Action, logEntry *command.MockLogEntry) {
				require.Len(t, actions, 1)

				require.Len(t, logEntry.WriteFunc.History(), 1)

				require.Len(t, logEntry.FinalizeFunc.History(), 1)
				assert.Equal(t, 128, logEntry.FinalizeFunc.History()[0].Arg0)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			clientset := fake.NewSimpleClientset()
			logger := command.NewMockLogger()
			logEntry := command.NewMockLogEntry()
			logger.LogEntryFunc.PushReturn(logEntry)

			if test.mockFunc != nil {
				test.mockFunc(clientset, logEntry)
			}

			cmd := &command.KubernetesCommand{
				Logger:     logtest.Scoped(t),
				Clientset:  clientset,
				Operations: command.NewOperations(&observation.TestContext),
			}

			err := cmd.ReadLogs(
				context.Background(),
				"my-namespace",
				test.pod,
				command.KubernetesJobContainerName,
				logEntry,
			)
			if test.expectedErr != nil {
				require.Error(t, err)
				assert.EqualError(t, err, test.expectedErr.Error())
			} else {
				require.NoError(t, err)
			}

			if test.mockAssertFunc != nil {
				test.mockAssertFunc(t, clientset.Actions(), logEntry)
			}
		})
	}
}

func TestKubernetesCommand_WaitForPodToSucceed(t *testing.T) {
	tests := []struct {
		name           string
		mockFunc       func(clientset *fake.Clientset)
		mockAssertFunc func(t *testing.T, actions []k8stesting.Action)
		expectedErr    error
	}{
		{
			name: "Pod succeeded",
			mockFunc: func(clientset *fake.Clientset) {
				watcher := watch.NewFakeWithChanSize(10, false)
				watcher.Add(&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: "my-pod",
						Labels: map[string]string{
							"job-name": "my-job",
						},
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodSucceeded,
					},
				})
				clientset.PrependWatchReactor("pods", k8stesting.DefaultWatchReactor(watcher, nil))
			},
			mockAssertFunc: func(t *testing.T, actions []k8stesting.Action) {
				require.Len(t, actions, 1)
				assert.Equal(t, "watch", actions[0].GetVerb())
				assert.Equal(t, "pods", actions[0].GetResource().Resource)
				assert.Equal(t, "job-name=my-job", actions[0].(k8stesting.WatchActionImpl).GetWatchRestrictions().Labels.String())
			},
		},
		{
			name: "Pod failed",
			mockFunc: func(clientset *fake.Clientset) {
				watcher := watch.NewFakeWithChanSize(10, false)
				watcher.Add(&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: "my-pod",
						Labels: map[string]string{
							"job-name": "my-job",
						},
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodFailed,
					},
				})
				clientset.PrependWatchReactor("pods", k8stesting.DefaultWatchReactor(watcher, nil))
			},
			expectedErr: errors.New("pod failed"),
		},
		{
			name: "Error occurred",
			mockFunc: func(clientset *fake.Clientset) {
				clientset.PrependWatchReactor("pods", k8stesting.DefaultWatchReactor(nil, errors.New("failed")))
			},
			expectedErr: errors.New("watching pod: failed"),
		},
		{
			name: "Pod succeeded second try",
			mockFunc: func(clientset *fake.Clientset) {
				watcher := watch.NewFakeWithChanSize(10, false)
				watcher.Add(&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: "my-pod",
						Labels: map[string]string{
							"job-name": "my-job",
						},
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodRunning,
					},
				})
				watcher.Add(&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: "my-pod",
						Labels: map[string]string{
							"job-name": "my-job",
						},
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodSucceeded,
					},
				})
				clientset.PrependWatchReactor("pods", k8stesting.DefaultWatchReactor(watcher, nil))
			},
			mockAssertFunc: func(t *testing.T, actions []k8stesting.Action) {
				require.Len(t, actions, 1)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			clientset := fake.NewSimpleClientset()

			if test.mockFunc != nil {
				test.mockFunc(clientset)
			}

			cmd := &command.KubernetesCommand{
				Logger:     logtest.Scoped(t),
				Clientset:  clientset,
				Operations: command.NewOperations(&observation.TestContext),
			}

			pod, err := cmd.WaitForPodToSucceed(
				context.Background(),
				"my-namespace",
				"my-job",
			)
			if test.expectedErr != nil {
				require.Error(t, err)
				assert.EqualError(t, err, test.expectedErr.Error())
			} else {
				require.NoError(t, err)
				require.NotNil(t, pod)
			}

			if test.mockAssertFunc != nil {
				test.mockAssertFunc(t, clientset.Actions())
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
		SecurityContext: command.KubernetesSecurityContext{
			FSGroup: pointer.Int64(1000),
		},
	}
	job := command.NewKubernetesJob("my-job", "my-image:latest", spec, "/my/path", options)

	assert.Equal(t, "my-job", job.Name)
	assert.Equal(t, int32(0), *job.Spec.BackoffLimit)

	assert.Equal(t, "my-node", job.Spec.Template.Spec.NodeName)
	assert.Equal(t, corev1.RestartPolicyNever, job.Spec.Template.Spec.RestartPolicy)

	require.Len(t, job.Spec.Template.Spec.Containers, 1)
	assert.Equal(t, "sg-executor-job-container", job.Spec.Template.Spec.Containers[0].Name)
	assert.Equal(t, "my-image:latest", job.Spec.Template.Spec.Containers[0].Image)
	assert.Equal(t, []string{"echo", "hello"}, job.Spec.Template.Spec.Containers[0].Command)
	assert.Equal(t, "/job", job.Spec.Template.Spec.Containers[0].WorkingDir)

	require.Len(t, job.Spec.Template.Spec.Containers[0].Env, 1)
	assert.Equal(t, "FOO", job.Spec.Template.Spec.Containers[0].Env[0].Name)
	assert.Equal(t, "bar", job.Spec.Template.Spec.Containers[0].Env[0].Value)

	assert.Equal(t, resource.MustParse("10"), *job.Spec.Template.Spec.Containers[0].Resources.Limits.Cpu())
	assert.Equal(t, resource.MustParse("10Gi"), *job.Spec.Template.Spec.Containers[0].Resources.Limits.Memory())
	assert.Equal(t, resource.MustParse("1"), *job.Spec.Template.Spec.Containers[0].Resources.Requests.Cpu())
	assert.Equal(t, resource.MustParse("1Gi"), *job.Spec.Template.Spec.Containers[0].Resources.Requests.Memory())

	require.Len(t, job.Spec.Template.Spec.Containers[0].VolumeMounts, 1)
	assert.Equal(t, "sg-executor-job-volume", job.Spec.Template.Spec.Containers[0].VolumeMounts[0].Name)
	assert.Equal(t, "/job", job.Spec.Template.Spec.Containers[0].VolumeMounts[0].MountPath)
	assert.Equal(t, "/my/path", job.Spec.Template.Spec.Containers[0].VolumeMounts[0].SubPath)

	require.Len(t, job.Spec.Template.Spec.Volumes, 1)
	assert.Equal(t, "sg-executor-job-volume", job.Spec.Template.Spec.Volumes[0].Name)
	assert.Equal(t, "my-pvc", job.Spec.Template.Spec.Volumes[0].PersistentVolumeClaim.ClaimName)

	assert.Nil(t, job.Spec.Template.Spec.SecurityContext.RunAsUser)
	assert.Nil(t, job.Spec.Template.Spec.SecurityContext.RunAsGroup)
	assert.Equal(t, int64(1000), *job.Spec.Template.Spec.SecurityContext.FSGroup)
}
