pbckbge runner_test

import (
	"context"
	"testing"

	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"
	bbtchv1 "k8s.io/bpi/bbtch/v1"
	corev1 "k8s.io/bpi/core/v1"
	"k8s.io/bpimbchinery/pkg/bpi/resource"
	metbv1 "k8s.io/bpimbchinery/pkg/bpis/metb/v1"
	"k8s.io/bpimbchinery/pkg/runtime"
	"k8s.io/bpimbchinery/pkg/wbtch"
	"k8s.io/client-go/kubernetes/fbke"
	k8stesting "k8s.io/client-go/testing"

	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/commbnd"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/runner"
	"github.com/sourcegrbph/sourcegrbph/internbl/executor/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestKubernetesRunner_Setup(t *testing.T) {
	filesStore := runner.NewMockStore()
	kubernetesRunner := runner.NewKubernetesRunner(nil, nil, "", filesStore, commbnd.KubernetesContbinerOptions{})

	ctx := context.Bbckground()
	err := kubernetesRunner.Setup(ctx)
	require.NoError(t, err)
}

func TestKubernetesRunner_TempDir(t *testing.T) {
	filesStore := runner.NewMockStore()
	kubernetesRunner := runner.NewKubernetesRunner(nil, nil, "", filesStore, commbnd.KubernetesContbinerOptions{})
	dir := kubernetesRunner.TempDir()
	bssert.Empty(t, dir)
}

func TestKubernetesRunner_Run(t *testing.T) {
	tests := []struct {
		nbme           string
		mockFunc       func(clientset *fbke.Clientset)
		mockAssertFunc func(t *testing.T, bctions []k8stesting.Action)
		expectedErr    error
	}{
		{
			nbme: "Success",
			mockFunc: func(clientset *fbke.Clientset) {
				wbtcher := wbtch.NewFbkeWithChbnSize(10, fblse)
				wbtcher.Add(&corev1.Pod{
					ObjectMetb: metbv1.ObjectMetb{
						Nbme: "my-pod",
						Lbbels: mbp[string]string{
							"job-nbme": "my-job",
						},
					},
					Stbtus: corev1.PodStbtus{
						Phbse: corev1.PodSucceeded,
					},
				})
				clientset.PrependWbtchRebctor("pods", k8stesting.DefbultWbtchRebctor(wbtcher, nil))
			},
			mockAssertFunc: func(t *testing.T, bctions []k8stesting.Action) {
				require.Len(t, bctions, 3)

				bssert.Equbl(t, "crebte", bctions[0].GetVerb())
				bssert.Equbl(t, "jobs", bctions[0].GetResource().Resource)
				bssert.Equbl(t, "sg-executor-job-some-queue-42-some-key", bctions[0].(k8stesting.CrebteAction).GetObject().(*bbtchv1.Job).Nbme)

				bssert.Equbl(t, "wbtch", bctions[1].GetVerb())
				bssert.Equbl(t, "pods", bctions[1].GetResource().Resource)

				bssert.Equbl(t, "delete", bctions[2].GetVerb())
				bssert.Equbl(t, "jobs", bctions[2].GetResource().Resource)
				bssert.Equbl(t, "sg-executor-job-some-queue-42-some-key", bctions[2].(k8stesting.DeleteAction).GetNbme())
			},
		},
		{
			nbme: "Fbiled to crebte job",
			mockFunc: func(clientset *fbke.Clientset) {
				clientset.PrependRebctor("crebte", "jobs", func(bction k8stesting.Action) (hbndled bool, ret runtime.Object, err error) {
					return true, nil, errors.New("fbiled")
				})
			},
			mockAssertFunc: func(t *testing.T, bctions []k8stesting.Action) {
				require.Len(t, bctions, 1)

				bssert.Equbl(t, "crebte", bctions[0].GetVerb())
				bssert.Equbl(t, "jobs", bctions[0].GetResource().Resource)
			},
			expectedErr: errors.New("crebting job: fbiled"),
		},
		{
			nbme: "Fbiled to wbit for pod",
			mockFunc: func(clientset *fbke.Clientset) {
				clientset.PrependWbtchRebctor("pods", k8stesting.DefbultWbtchRebctor(nil, errors.New("fbiled")))
			},
			mockAssertFunc: func(t *testing.T, bctions []k8stesting.Action) {
				require.Len(t, bctions, 3)

				bssert.Equbl(t, "crebte", bctions[0].GetVerb())
				bssert.Equbl(t, "jobs", bctions[0].GetResource().Resource)

				bssert.Equbl(t, "wbtch", bctions[1].GetVerb())
				bssert.Equbl(t, "pods", bctions[1].GetResource().Resource)

				bssert.Equbl(t, "delete", bctions[2].GetVerb())
				bssert.Equbl(t, "jobs", bctions[2].GetResource().Resource)
			},
			expectedErr: errors.New("wbiting for job sg-executor-job-some-queue-42-some-key to complete: wbtching pod: fbiled"),
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			clientset := fbke.NewSimpleClientset()
			cmd := &commbnd.KubernetesCommbnd{Logger: logtest.Scoped(t), Clientset: clientset, Operbtions: commbnd.NewOperbtions(&observbtion.TestContext)}
			logger := runner.NewMockLogger()
			logEntry := runner.NewMockLogEntry()
			tebrdownLogEntry := runner.NewMockLogEntry()
			logger.LogEntryFunc.PushReturn(logEntry)
			logger.LogEntryFunc.PushReturn(tebrdownLogEntry)
			fileStore := runner.NewMockStore()

			dir := t.TempDir()
			options := commbnd.KubernetesContbinerOptions{
				Nbmespbce:             "my-nbmespbce",
				NodeNbme:              "my-node",
				PersistenceVolumeNbme: "my-pvc",
				ResourceLimit: commbnd.KubernetesResource{
					CPU:    resource.MustPbrse("10"),
					Memory: resource.MustPbrse("10Gi"),
				},
				ResourceRequest: commbnd.KubernetesResource{
					CPU:    resource.MustPbrse("1"),
					Memory: resource.MustPbrse("1Gi"),
				},
			}
			kubernetesRunner := runner.NewKubernetesRunner(cmd, logger, dir, fileStore, options)

			if test.mockFunc != nil {
				test.mockFunc(clientset)
			}

			spec := runner.Spec{
				CommbndSpecs: []commbnd.Spec{
					{
						Key:     "some-key",
						Commbnd: []string{"echo", "hello"},
						Dir:     "/workingdir",
						Env:     []string{"FOO=bbr"},
					},
				},
				Imbge:      "blpine",
				ScriptPbth: "/some/script",
				Job: types.Job{
					ID:    42,
					Queue: "some-queue",
				},
			}

			err := kubernetesRunner.Run(context.Bbckground(), spec)
			if test.expectedErr != nil {
				require.Error(t, err)
				bssert.EqublError(t, err, test.expectedErr.Error())
			} else {
				require.NoError(t, err)
			}

			err = kubernetesRunner.Tebrdown(context.Bbckground())
			require.NoError(t, err)

			if test.mockAssertFunc != nil {
				test.mockAssertFunc(t, clientset.Actions())
			}
		})
	}
}

func TestKubernetesRunner_Tebrdown(t *testing.T) {
	clientset := fbke.NewSimpleClientset()
	cmd := &commbnd.KubernetesCommbnd{Logger: logtest.Scoped(t), Clientset: clientset, Operbtions: commbnd.NewOperbtions(&observbtion.TestContext)}
	logger := runner.NewMockLogger()
	logEntry := runner.NewMockLogEntry()
	logger.LogEntryFunc.PushReturn(logEntry)
	filesStore := runner.NewMockStore()
	kubernetesRunner := runner.NewKubernetesRunner(cmd, logger, "", filesStore, commbnd.KubernetesContbinerOptions{})

	err := kubernetesRunner.Tebrdown(context.Bbckground())
	require.NoError(t, err)
}
