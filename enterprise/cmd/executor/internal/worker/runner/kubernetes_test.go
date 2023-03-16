package runner_test

import (
	"context"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/worker/command"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/worker/runner"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestKubernetesRunner_Setup(t *testing.T) {
	kubernetesRunner := runner.NewKubernetesRunner(nil, nil, "", command.KubernetesContainerOptions{})

	ctx := context.Background()
	err := kubernetesRunner.Setup(ctx)
	require.NoError(t, err)
}

func TestKubernetesRunner_TempDir(t *testing.T) {
	kubernetesRunner := runner.NewKubernetesRunner(nil, nil, "", command.KubernetesContainerOptions{})
	dir := kubernetesRunner.TempDir()
	assert.Empty(t, dir)
}

func TestKubernetesRunner_Run(t *testing.T) {
	tests := []struct {
		name           string
		mockFunc       func(clientset *fake.Clientset)
		mockAssertFunc func(t *testing.T, actions []k8stesting.Action)
		expectedErr    error
	}{
		{
			name: "Success",
			mockFunc: func(clientset *fake.Clientset) {
				clientset.PrependReactor("get", "jobs", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, &batchv1.Job{Status: batchv1.JobStatus{Succeeded: 1}}, nil
				})
				clientset.PrependReactor("list", "pods", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, &corev1.PodList{Items: []corev1.Pod{
						{ObjectMeta: metav1.ObjectMeta{
							Name:   "my-pod",
							Labels: map[string]string{"job-name": "job-some-queue-42-some-key"},
						}}},
					}, nil
				})
			},
			mockAssertFunc: func(t *testing.T, actions []k8stesting.Action) {
				require.Len(t, actions, 5)

				assert.Equal(t, "create", actions[0].GetVerb())
				assert.Equal(t, "jobs", actions[0].GetResource().Resource)
				assert.Equal(t, "job-some-queue-42-some-key", actions[0].(k8stesting.CreateAction).GetObject().(*batchv1.Job).Name)

				assert.Equal(t, "get", actions[1].GetVerb())
				assert.Equal(t, "jobs", actions[1].GetResource().Resource)
				assert.Equal(t, "job-some-queue-42-some-key", actions[1].(k8stesting.GetAction).GetName())

				assert.Equal(t, "list", actions[2].GetVerb())
				assert.Equal(t, "pods", actions[2].GetResource().Resource)
				assert.Equal(t, "job-name=job-some-queue-42-some-key", actions[2].(k8stesting.ListAction).GetListRestrictions().Labels.String())

				assert.Equal(t, "get", actions[3].GetVerb())
				assert.Equal(t, "pods", actions[3].GetResource().Resource)
				assert.Equal(t, "log", actions[3].GetSubresource())
				assert.Equal(t, "job-container", actions[3].(k8stesting.GenericAction).GetValue().(*corev1.PodLogOptions).Container)

				assert.Equal(t, "delete", actions[4].GetVerb())
				assert.Equal(t, "jobs", actions[4].GetResource().Resource)
				assert.Equal(t, "job-some-queue-42-some-key", actions[4].(k8stesting.DeleteAction).GetName())
			},
		},
		{
			name: "Failed to create job",
			mockFunc: func(clientset *fake.Clientset) {
				clientset.PrependReactor("create", "jobs", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, nil, errors.New("failed")
				})
			},
			mockAssertFunc: func(t *testing.T, actions []k8stesting.Action) {
				require.Len(t, actions, 1)

				assert.Equal(t, "create", actions[0].GetVerb())
				assert.Equal(t, "jobs", actions[0].GetResource().Resource)
			},
			expectedErr: errors.New("creating job: failed"),
		},
		{
			name: "Failed to wait for job",
			mockFunc: func(clientset *fake.Clientset) {
				clientset.PrependReactor("get", "jobs", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, nil, errors.New("failed")
				})
			},
			mockAssertFunc: func(t *testing.T, actions []k8stesting.Action) {
				require.Len(t, actions, 2)

				assert.Equal(t, "create", actions[0].GetVerb())
				assert.Equal(t, "jobs", actions[0].GetResource().Resource)

				assert.Equal(t, "get", actions[1].GetVerb())
				assert.Equal(t, "jobs", actions[1].GetResource().Resource)
			},
			expectedErr: errors.New("waiting for job to complete: retrieving job: failed"),
		},
		{
			name: "Failed to find job pod",
			mockFunc: func(clientset *fake.Clientset) {
				clientset.PrependReactor("get", "jobs", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, &batchv1.Job{Status: batchv1.JobStatus{Succeeded: 1}}, nil
				})
				clientset.PrependReactor("list", "pods", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, nil, errors.New("failed")
				})
			},
			mockAssertFunc: func(t *testing.T, actions []k8stesting.Action) {
				require.Len(t, actions, 4)

				assert.Equal(t, "create", actions[0].GetVerb())
				assert.Equal(t, "jobs", actions[0].GetResource().Resource)

				assert.Equal(t, "get", actions[1].GetVerb())
				assert.Equal(t, "jobs", actions[1].GetResource().Resource)

				assert.Equal(t, "list", actions[2].GetVerb())
				assert.Equal(t, "pods", actions[2].GetResource().Resource)

				assert.Equal(t, "delete", actions[3].GetVerb())
				assert.Equal(t, "jobs", actions[3].GetResource().Resource)
			},
			expectedErr: errors.New("finding pod: failed"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			clientset := fake.NewSimpleClientset()
			cmd := &command.KubernetesCommand{Logger: logtest.Scoped(t), Clientset: clientset}
			logger := runner.NewMockLogger()
			logEntry := runner.NewMockLogEntry()
			logger.LogEntryFunc.PushReturn(logEntry)

			dir := t.TempDir()
			options := command.KubernetesContainerOptions{
				Namespace:             "my-namespace",
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
			kubernetesRunner := runner.NewKubernetesRunner(cmd, logger, dir, options)

			if test.mockFunc != nil {
				test.mockFunc(clientset)
			}

			spec := runner.Spec{
				Queue: "some-queue",
				JobID: 42,
				CommandSpec: command.Spec{
					Key:     "some-key",
					Command: []string{"echo", "hello"},
					Dir:     "/workingdir",
					Env:     []string{"FOO=bar"},
				},
				Image:      "alpine",
				ScriptPath: "/some/script",
			}

			err := kubernetesRunner.Run(context.Background(), spec)
			if test.expectedErr != nil {
				require.Error(t, err)
				assert.EqualError(t, err, test.expectedErr.Error())
			} else {
				require.NoError(t, err)
			}

			err = kubernetesRunner.Teardown(context.Background())
			require.NoError(t, err)

			if test.mockAssertFunc != nil {
				test.mockAssertFunc(t, clientset.Actions())
			}
		})
	}
}
