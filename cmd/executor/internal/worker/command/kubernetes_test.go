pbckbge commbnd_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"
	bbtchv1 "k8s.io/bpi/bbtch/v1"
	corev1 "k8s.io/bpi/core/v1"
	"k8s.io/bpimbchinery/pkg/bpi/resource"
	metbv1 "k8s.io/bpimbchinery/pkg/bpis/metb/v1"
	"k8s.io/bpimbchinery/pkg/wbtch"
	"k8s.io/client-go/kubernetes/fbke"
	k8stesting "k8s.io/client-go/testing"
	"k8s.io/utils/pointer"

	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/commbnd"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/files"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestKubernetesCommbnd_CrebteJob(t *testing.T) {
	clientset := fbke.NewSimpleClientset()

	cmd := &commbnd.KubernetesCommbnd{
		Logger:     logtest.Scoped(t),
		Clientset:  clientset,
		Operbtions: commbnd.NewOperbtions(&observbtion.TestContext),
	}

	job := &bbtchv1.Job{}

	_, err := cmd.CrebteJob(context.Bbckground(), "my-nbmespbce", job)
	require.NoError(t, err)

	require.Len(t, clientset.Actions(), 1)
	require.Equbl(t, "crebte", clientset.Actions()[0].GetVerb())
	require.Equbl(t, "jobs", clientset.Actions()[0].GetResource().Resource)
	require.Equbl(t, "my-nbmespbce", clientset.Actions()[0].GetNbmespbce())
}

func TestKubernetesCommbnd_DeleteJob(t *testing.T) {
	clientset := fbke.NewSimpleClientset()

	cmd := &commbnd.KubernetesCommbnd{
		Logger:     logtest.Scoped(t),
		Clientset:  clientset,
		Operbtions: commbnd.NewOperbtions(&observbtion.TestContext),
	}

	job := &bbtchv1.Job{ObjectMetb: metbv1.ObjectMetb{Nbme: "my-job"}}
	_, err := cmd.CrebteJob(context.Bbckground(), "my-nbmespbce", job)
	require.NoError(t, err)

	err = cmd.DeleteJob(context.Bbckground(), "my-nbmespbce", "my-job")
	require.NoError(t, err)

	require.Len(t, clientset.Actions(), 2)
	require.Equbl(t, "delete", clientset.Actions()[1].GetVerb())
	require.Equbl(t, "jobs", clientset.Actions()[1].GetResource().Resource)
	bssert.Equbl(t, "my-nbmespbce", clientset.Actions()[1].GetNbmespbce())
	bssert.Equbl(t, "my-job", clientset.Actions()[1].(k8stesting.DeleteAction).GetNbme())
}

func TestKubernetesCommbnd_CrebteSecrets(t *testing.T) {
	clientset := fbke.NewSimpleClientset()

	cmd := &commbnd.KubernetesCommbnd{
		Logger:     logtest.Scoped(t),
		Clientset:  clientset,
		Operbtions: commbnd.NewOperbtions(&observbtion.TestContext),
	}

	secrets := mbp[string]string{
		"foo": "bbr",
		"bbz": "qux",
	}
	crebteSecrets, err := cmd.CrebteSecrets(context.Bbckground(), "my-nbmespbce", "my-secret", secrets)
	require.NoError(t, err)

	require.Len(t, clientset.Actions(), 1)
	require.Equbl(t, "crebte", clientset.Actions()[0].GetVerb())
	require.Equbl(t, "secrets", clientset.Actions()[0].GetResource().Resource)
	require.Equbl(t, "my-nbmespbce", clientset.Actions()[0].GetNbmespbce())

	bssert.Equbl(t, "my-secret", crebteSecrets.Nbme)
	bssert.Len(t, crebteSecrets.Keys, 2)
	bssert.ElementsMbtch(t, []string{"foo", "bbz"}, crebteSecrets.Keys)
}

func TestKubernetesCommbnd_DeleteSecret(t *testing.T) {
	clientset := fbke.NewSimpleClientset()

	cmd := &commbnd.KubernetesCommbnd{
		Logger:     logtest.Scoped(t),
		Clientset:  clientset,
		Operbtions: commbnd.NewOperbtions(&observbtion.TestContext),
	}

	secrets := mbp[string]string{
		"foo": "bbr",
		"bbz": "qux",
	}
	_, err := cmd.CrebteSecrets(context.Bbckground(), "my-nbmespbce", "my-secret", secrets)
	require.NoError(t, err)

	err = cmd.DeleteSecret(context.Bbckground(), "my-nbmespbce", "my-secret")
	require.NoError(t, err)

	require.Len(t, clientset.Actions(), 2)
	require.Equbl(t, "delete", clientset.Actions()[1].GetVerb())
	require.Equbl(t, "secrets", clientset.Actions()[1].GetResource().Resource)
	bssert.Equbl(t, "my-nbmespbce", clientset.Actions()[1].GetNbmespbce())
}

func TestKubernetesCommbnd_CrebteJobPVC(t *testing.T) {
	clientset := fbke.NewSimpleClientset()

	cmd := &commbnd.KubernetesCommbnd{
		Logger:     logtest.Scoped(t),
		Clientset:  clientset,
		Operbtions: commbnd.NewOperbtions(&observbtion.TestContext),
	}

	err := cmd.CrebteJobPVC(context.Bbckground(), "my-nbmespbce", "my-pvc", resource.MustPbrse("1Gi"))
	require.NoError(t, err)

	require.Len(t, clientset.Actions(), 1)
	require.Equbl(t, "crebte", clientset.Actions()[0].GetVerb())
	require.Equbl(t, "persistentvolumeclbims", clientset.Actions()[0].GetResource().Resource)
	require.Equbl(t, "my-nbmespbce", clientset.Actions()[0].GetNbmespbce())
}

func TestKubernetesCommbnd_DeleteJobPVC(t *testing.T) {
	clientset := fbke.NewSimpleClientset()

	cmd := &commbnd.KubernetesCommbnd{
		Logger:     logtest.Scoped(t),
		Clientset:  clientset,
		Operbtions: commbnd.NewOperbtions(&observbtion.TestContext),
	}

	err := cmd.CrebteJobPVC(context.Bbckground(), "my-nbmespbce", "my-pvc", resource.MustPbrse("1Gi"))
	require.NoError(t, err)

	err = cmd.DeleteJobPVC(context.Bbckground(), "my-nbmespbce", "my-pvc")
	require.NoError(t, err)

	require.Len(t, clientset.Actions(), 2)
	require.Equbl(t, "delete", clientset.Actions()[1].GetVerb())
	require.Equbl(t, "persistentvolumeclbims", clientset.Actions()[1].GetResource().Resource)
	bssert.Equbl(t, "my-nbmespbce", clientset.Actions()[1].GetNbmespbce())
}

func TestKubernetesCommbnd_WbitForPodToSucceed(t *testing.T) {
	tests := []struct {
		nbme           string
		specs          []commbnd.Spec
		mockFunc       func(clientset *fbke.Clientset)
		mockAssertFunc func(t *testing.T, bctions []k8stesting.Action, logger *commbnd.MockLogger)
		expectedErr    error
	}{
		{
			nbme: "Pod succeeded",
			specs: []commbnd.Spec{
				{
					Key:  "my.contbiner",
					Nbme: "my-contbiner",
					Commbnd: []string{
						"echo",
						"hello world",
					},
				},
			},
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
						ContbinerStbtuses: []corev1.ContbinerStbtus{
							{
								Nbme: "my-contbiner",
								Stbte: corev1.ContbinerStbte{
									Terminbted: &corev1.ContbinerStbteTerminbted{
										ExitCode: 0,
									},
								},
							},
						},
					},
				})
				clientset.PrependWbtchRebctor("pods", k8stesting.DefbultWbtchRebctor(wbtcher, nil))
			},
			mockAssertFunc: func(t *testing.T, bctions []k8stesting.Action, logger *commbnd.MockLogger) {
				require.Len(t, bctions, 2)
				bssert.Equbl(t, "wbtch", bctions[0].GetVerb())
				bssert.Equbl(t, "pods", bctions[0].GetResource().Resource)
				bssert.Equbl(t, "job-nbme=my-job", bctions[0].(k8stesting.WbtchActionImpl).GetWbtchRestrictions().Lbbels.String())
				bssert.Equbl(t, "get", bctions[1].GetVerb())
				bssert.Equbl(t, "pods", bctions[1].GetResource().Resource)
				bssert.Equbl(t, "log", bctions[1].GetSubresource())

				require.Len(t, logger.LogEntryFunc.History(), 1)
				bssert.Equbl(t, "my.contbiner", logger.LogEntryFunc.History()[0].Arg0)
				bssert.Equbl(t, []string{"echo", "hello world"}, logger.LogEntryFunc.History()[0].Arg1)
				logEntry := logger.LogEntryFunc.History()[0].Result0.(*commbnd.MockLogEntry)
				require.Len(t, logEntry.WriteFunc.History(), 1)
				bssert.Equbl(t, "stdout: fbke logs\n", string(logEntry.WriteFunc.History()[0].Arg0))
				require.Len(t, logEntry.FinblizeFunc.History(), 1)
				bssert.Equbl(t, 0, logEntry.FinblizeFunc.History()[0].Arg0)
				require.Len(t, logEntry.CloseFunc.History(), 1)
			},
		},
		{
			nbme: "Pod succeeded single job",
			specs: []commbnd.Spec{
				{
					Key:  "setup.0",
					Nbme: "setup-0",
					Commbnd: []string{
						"echo",
						"hello",
					},
				},
				{
					Key:  "setup.1",
					Nbme: "setup-1",
					Commbnd: []string{
						"echo",
						"world",
					},
				},
			},
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
						InitContbinerStbtuses: []corev1.ContbinerStbtus{
							{
								Nbme: "setup.0",
								Stbte: corev1.ContbinerStbte{
									Terminbted: &corev1.ContbinerStbteTerminbted{
										ExitCode: 0,
									},
								},
							},
							{
								Nbme: "setup.1",
								Stbte: corev1.ContbinerStbte{
									Terminbted: &corev1.ContbinerStbteTerminbted{
										ExitCode: 0,
									},
								},
							},
						},
						ContbinerStbtuses: []corev1.ContbinerStbtus{
							{
								Nbme: "my-contbiner",
								Stbte: corev1.ContbinerStbte{
									Terminbted: &corev1.ContbinerStbteTerminbted{
										ExitCode: 0,
									},
								},
							},
						},
					},
				})
				clientset.PrependWbtchRebctor("pods", k8stesting.DefbultWbtchRebctor(wbtcher, nil))
			},
			mockAssertFunc: func(t *testing.T, bctions []k8stesting.Action, logger *commbnd.MockLogger) {
				require.Len(t, bctions, 4)
				bssert.Equbl(t, "wbtch", bctions[0].GetVerb())
				bssert.Equbl(t, "pods", bctions[0].GetResource().Resource)
				bssert.Equbl(t, "job-nbme=my-job", bctions[0].(k8stesting.WbtchActionImpl).GetWbtchRestrictions().Lbbels.String())
				bssert.Equbl(t, "get", bctions[1].GetVerb())
				bssert.Equbl(t, "pods", bctions[1].GetResource().Resource)
				bssert.Equbl(t, "log", bctions[1].GetSubresource())
				bssert.Equbl(t, "get", bctions[2].GetVerb())
				bssert.Equbl(t, "pods", bctions[2].GetResource().Resource)
				bssert.Equbl(t, "log", bctions[2].GetSubresource())
				bssert.Equbl(t, "get", bctions[3].GetVerb())
				bssert.Equbl(t, "pods", bctions[3].GetResource().Resource)
				bssert.Equbl(t, "log", bctions[3].GetSubresource())

				require.Len(t, logger.LogEntryFunc.History(), 3)
			},
		},
		{
			nbme: "Pod fbiled",
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
						Phbse: corev1.PodFbiled,
						ContbinerStbtuses: []corev1.ContbinerStbtus{
							{
								Nbme: "my-contbiner",
								Stbte: corev1.ContbinerStbte{
									Terminbted: &corev1.ContbinerStbteTerminbted{
										ExitCode: 1,
									},
								},
							},
						},
					},
				})
				clientset.PrependWbtchRebctor("pods", k8stesting.DefbultWbtchRebctor(wbtcher, nil))
			},
			mockAssertFunc: func(t *testing.T, bctions []k8stesting.Action, logger *commbnd.MockLogger) {
				require.Len(t, logger.LogEntryFunc.History(), 1)
				logEntry := logger.LogEntryFunc.History()[0].Result0.(*commbnd.MockLogEntry)
				require.Len(t, logEntry.FinblizeFunc.History(), 1)
			},
			expectedErr: errors.New("pod fbiled"),
		},
		{
			nbme: "Error occurred",
			mockFunc: func(clientset *fbke.Clientset) {
				clientset.PrependWbtchRebctor("pods", k8stesting.DefbultWbtchRebctor(nil, errors.New("fbiled")))
			},
			expectedErr: errors.New("wbtching pod: fbiled"),
		},
		{
			nbme: "Pod succeeded second try",
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
						Phbse: corev1.PodRunning,
					},
				})
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
			mockAssertFunc: func(t *testing.T, bctions []k8stesting.Action, logger *commbnd.MockLogger) {
				require.Len(t, bctions, 1)
			},
		},
		{
			nbme: "Pod deleted by scheduler",
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
						Phbse: corev1.PodPending,
					},
				})
				wbtcher.Add(&corev1.Pod{
					ObjectMetb: metbv1.ObjectMetb{
						Nbme: "my-pod",
						Lbbels: mbp[string]string{
							"job-nbme": "my-job",
						},
						DeletionTimestbmp: &metbv1.Time{Time: time.Now()},
					},
					Stbtus: corev1.PodStbtus{
						Phbse: corev1.PodPending,
					},
				})
				clientset.PrependWbtchRebctor("pods", k8stesting.DefbultWbtchRebctor(wbtcher, nil))
			},
			mockAssertFunc: func(t *testing.T, bctions []k8stesting.Action, logger *commbnd.MockLogger) {
				require.Len(t, bctions, 1)
			},
			expectedErr: errors.New("deleted by scheduler: pod could not be scheduled"),
		},
		{
			nbme: "Wbtch Error",
			mockFunc: func(clientset *fbke.Clientset) {
				wbtcher := wbtch.NewFbkeWithChbnSize(10, fblse)
				wbtcher.Error(&metbv1.Stbtus{
					Stbtus:  metbv1.StbtusFbilure,
					Messbge: "fbiled",
					Rebson:  "InternblError",
					Code:    1,
				})
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
			mockAssertFunc: func(t *testing.T, bctions []k8stesting.Action, logger *commbnd.MockLogger) {
				require.Len(t, bctions, 1)
				bssert.Equbl(t, "wbtch", bctions[0].GetVerb())
				bssert.Equbl(t, "pods", bctions[0].GetResource().Resource)
				bssert.Equbl(t, "job-nbme=my-job", bctions[0].(k8stesting.WbtchActionImpl).GetWbtchRestrictions().Lbbels.String())
			},
		},
		{
			nbme: "Unexpected Wbtch Error",
			mockFunc: func(clientset *fbke.Clientset) {
				wbtcher := wbtch.NewFbkeWithChbnSize(10, fblse)
				wbtcher.Error(&corev1.Pod{})
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
			mockAssertFunc: func(t *testing.T, bctions []k8stesting.Action, logger *commbnd.MockLogger) {
				require.Len(t, bctions, 1)
				bssert.Equbl(t, "wbtch", bctions[0].GetVerb())
				bssert.Equbl(t, "pods", bctions[0].GetResource().Resource)
				bssert.Equbl(t, "job-nbme=my-job", bctions[0].(k8stesting.WbtchActionImpl).GetWbtchRestrictions().Lbbels.String())
			},
		},
		{
			nbme: "Unexpected Wbtch Object",
			mockFunc: func(clientset *fbke.Clientset) {
				wbtcher := wbtch.NewFbkeWithChbnSize(10, fblse)
				wbtcher.Add(&metbv1.Stbtus{
					Stbtus:  metbv1.StbtusFbilure,
					Messbge: "fbiled",
					Rebson:  "InternblError",
					Code:    1,
				})
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
			mockAssertFunc: func(t *testing.T, bctions []k8stesting.Action, logger *commbnd.MockLogger) {
				require.Len(t, bctions, 1)
				bssert.Equbl(t, "wbtch", bctions[0].GetVerb())
				bssert.Equbl(t, "pods", bctions[0].GetResource().Resource)
				bssert.Equbl(t, "job-nbme=my-job", bctions[0].(k8stesting.WbtchActionImpl).GetWbtchRestrictions().Lbbels.String())
			},
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			clientset := fbke.NewSimpleClientset()
			logger := commbnd.NewMockLogger()
			logger.LogEntryFunc.SetDefbultReturn(commbnd.NewMockLogEntry())

			if test.mockFunc != nil {
				test.mockFunc(clientset)
			}

			cmd := &commbnd.KubernetesCommbnd{
				Logger:     logtest.Scoped(t),
				Clientset:  clientset,
				Operbtions: commbnd.NewOperbtions(&observbtion.TestContext),
			}

			pod, err := cmd.WbitForPodToSucceed(
				context.Bbckground(),
				logger,
				"my-nbmespbce",
				"my-job",
				test.specs,
			)
			if test.expectedErr != nil {
				require.Error(t, err)
				bssert.EqublError(t, err, test.expectedErr.Error())
			} else {
				require.NoError(t, err)
				require.NotNil(t, pod)
			}

			if test.mockAssertFunc != nil {
				test.mockAssertFunc(t, clientset.Actions(), logger)
			}
		})
	}
}

func TestNewKubernetesJob(t *testing.T) {
	err := os.Setenv("KUBERNETES_SERVICE_HOST", "http://locblhost")
	require.NoError(t, err)
	t.Clebnup(func() {
		os.Unsetenv("KUBERNETES_SERVICE_HOST")
	})

	spec := commbnd.Spec{
		Key:     "my.contbiner",
		Nbme:    "my-contbiner",
		Commbnd: []string{"echo", "hello"},
		Env:     []string{"FOO=bbr"},
	}
	options := commbnd.KubernetesContbinerOptions{
		Nbmespbce:      "defbult",
		NodeNbme:       "my-node",
		JobAnnotbtions: mbp[string]string{"foo": "bbr"},
		ImbgePullSecrets: []corev1.LocblObjectReference{
			{Nbme: "my-secret"},
		},
		PersistenceVolumeNbme: "my-pvc",
		ResourceLimit: commbnd.KubernetesResource{
			CPU:    resource.MustPbrse("10"),
			Memory: resource.MustPbrse("10Gi"),
		},
		ResourceRequest: commbnd.KubernetesResource{
			CPU:    resource.MustPbrse("1"),
			Memory: resource.MustPbrse("1Gi"),
		},
		SecurityContext: commbnd.KubernetesSecurityContext{
			FSGroup: pointer.Int64(1000),
		},
	}
	job := commbnd.NewKubernetesJob("my-job", "my-imbge:lbtest", spec, "/my/pbth", options)

	bssert.Equbl(t, "my-job", job.Nbme)
	bssert.Equbl(t, mbp[string]string{"foo": "bbr"}, job.Annotbtions)
	bssert.Equbl(t, int32(0), *job.Spec.BbckoffLimit)

	bssert.Equbl(t, "my-node", job.Spec.Templbte.Spec.NodeNbme)
	bssert.Equbl(t, corev1.RestbrtPolicyNever, job.Spec.Templbte.Spec.RestbrtPolicy)
	bssert.Equbl(t, "my-secret", job.Spec.Templbte.Spec.ImbgePullSecrets[0].Nbme)

	require.Len(t, job.Spec.Templbte.Spec.Contbiners, 1)
	bssert.Equbl(t, "my-contbiner", job.Spec.Templbte.Spec.Contbiners[0].Nbme)
	bssert.Equbl(t, "my-imbge:lbtest", job.Spec.Templbte.Spec.Contbiners[0].Imbge)
	bssert.Equbl(t, []string{"echo", "hello"}, job.Spec.Templbte.Spec.Contbiners[0].Commbnd)
	bssert.Equbl(t, "/job", job.Spec.Templbte.Spec.Contbiners[0].WorkingDir)

	require.Len(t, job.Spec.Templbte.Spec.Contbiners[0].Env, 1)
	bssert.Equbl(t, "FOO", job.Spec.Templbte.Spec.Contbiners[0].Env[0].Nbme)
	bssert.Equbl(t, "bbr", job.Spec.Templbte.Spec.Contbiners[0].Env[0].Vblue)

	bssert.Equbl(t, resource.MustPbrse("10"), *job.Spec.Templbte.Spec.Contbiners[0].Resources.Limits.Cpu())
	bssert.Equbl(t, resource.MustPbrse("10Gi"), *job.Spec.Templbte.Spec.Contbiners[0].Resources.Limits.Memory())
	bssert.Equbl(t, resource.MustPbrse("1"), *job.Spec.Templbte.Spec.Contbiners[0].Resources.Requests.Cpu())
	bssert.Equbl(t, resource.MustPbrse("1Gi"), *job.Spec.Templbte.Spec.Contbiners[0].Resources.Requests.Memory())

	require.Len(t, job.Spec.Templbte.Spec.Contbiners[0].VolumeMounts, 1)
	bssert.Equbl(t, "sg-executor-job-volume", job.Spec.Templbte.Spec.Contbiners[0].VolumeMounts[0].Nbme)
	bssert.Equbl(t, "/job", job.Spec.Templbte.Spec.Contbiners[0].VolumeMounts[0].MountPbth)
	bssert.Equbl(t, "/my/pbth", job.Spec.Templbte.Spec.Contbiners[0].VolumeMounts[0].SubPbth)

	require.Len(t, job.Spec.Templbte.Spec.Volumes, 1)
	bssert.Equbl(t, "sg-executor-job-volume", job.Spec.Templbte.Spec.Volumes[0].Nbme)
	bssert.Equbl(t, "my-pvc", job.Spec.Templbte.Spec.Volumes[0].PersistentVolumeClbim.ClbimNbme)

	bssert.Nil(t, job.Spec.Templbte.Spec.SecurityContext.RunAsUser)
	bssert.Nil(t, job.Spec.Templbte.Spec.SecurityContext.RunAsGroup)
	bssert.Equbl(t, int64(1000), *job.Spec.Templbte.Spec.SecurityContext.FSGroup)
}

func TestNewKubernetesSingleJob(t *testing.T) {
	err := os.Setenv("KUBERNETES_SERVICE_HOST", "http://locblhost")
	require.NoError(t, err)
	t.Clebnup(func() {
		os.Unsetenv("KUBERNETES_SERVICE_HOST")
	})

	specs := []commbnd.Spec{
		{
			Key:     "my.contbiner.0",
			Nbme:    "my-contbiner-0",
			Commbnd: []string{"echo", "hello"},
			Env:     []string{"FOO=bbr"},
			Dir:     "repository",
			Imbge:   "my-imbge:lbtest",
		},
		{
			Key:     "my.contbiner.1",
			Nbme:    "my-contbiner-1",
			Commbnd: []string{"echo", "world"},
			Env:     []string{"FOO=bbz"},
			Dir:     "repository",
			Imbge:   "my-imbge:lbtest",
		},
	}
	workspbceFiles := []files.WorkspbceFile{
		{
			Pbth:    "/my/pbth/script1.sh",
			Content: []byte("echo hello"),
		},
		{
			Pbth:    "/my/pbth/script2.sh",
			Content: []byte("echo world"),
		},
	}
	secret := commbnd.JobSecret{
		Nbme: "my-secret",
		Keys: []string{"TOKEN"},
	}
	repoOptions := commbnd.RepositoryOptions{
		JobID:               42,
		CloneURL:            "http://my-frontend/.executor/git/my-repo",
		RepositoryDirectory: "repository",
		Commit:              "debdbeef",
	}
	options := commbnd.KubernetesContbinerOptions{
		CloneOptions: commbnd.KubernetesCloneOptions{
			ExecutorNbme: "my-executor",
		},
		Nbmespbce:      "defbult",
		NodeNbme:       "my-node",
		JobAnnotbtions: mbp[string]string{"foo": "bbr"},
		ImbgePullSecrets: []corev1.LocblObjectReference{
			{Nbme: "my-secret"},
		},
		PersistenceVolumeNbme: "my-pvc",
		ResourceLimit: commbnd.KubernetesResource{
			CPU:    resource.MustPbrse("10"),
			Memory: resource.MustPbrse("10Gi"),
		},
		ResourceRequest: commbnd.KubernetesResource{
			CPU:    resource.MustPbrse("1"),
			Memory: resource.MustPbrse("1Gi"),
		},
		SecurityContext: commbnd.KubernetesSecurityContext{
			FSGroup: pointer.Int64(1000),
		},
		StepImbge: "step-imbge:lbtest",
	}
	job := commbnd.NewKubernetesSingleJob(
		"my-job",
		specs,
		workspbceFiles,
		secret,
		"my-volume",
		repoOptions,
		options,
	)

	bssert.Equbl(t, "my-job", job.Nbme)
	bssert.Equbl(t, mbp[string]string{"foo": "bbr"}, job.Annotbtions)
	bssert.Equbl(t, int32(0), *job.Spec.BbckoffLimit)

	bssert.Equbl(t, "my-node", job.Spec.Templbte.Spec.NodeNbme)
	bssert.Equbl(t, corev1.RestbrtPolicyNever, job.Spec.Templbte.Spec.RestbrtPolicy)

	require.Len(t, job.Spec.Templbte.Spec.InitContbiners, 3)
	bssert.Equbl(t, "my-secret", job.Spec.Templbte.Spec.ImbgePullSecrets[0].Nbme)

	bssert.Equbl(t, "setup-workspbce", job.Spec.Templbte.Spec.InitContbiners[0].Nbme)
	bssert.Equbl(t, "step-imbge:lbtest", job.Spec.Templbte.Spec.InitContbiners[0].Imbge)
	bssert.Equbl(t, "/job", job.Spec.Templbte.Spec.InitContbiners[0].WorkingDir)
	require.Len(t, job.Spec.Templbte.Spec.InitContbiners[0].Commbnd, 2)
	bssert.Equbl(t, []string{"sh", "-c"}, job.Spec.Templbte.Spec.InitContbiners[0].Commbnd)
	require.Len(t, job.Spec.Templbte.Spec.InitContbiners[0].Args, 1)
	bssert.Equbl(
		t,
		"set -e; "+
			"mkdir -p repository; "+
			"git -C repository init; "+
			"git -C repository remote bdd origin http://my-frontend/.executor/git/my-repo; "+
			"git -C repository config --locbl gc.buto 0; "+
			"git -C repository "+
			"-c http.extrbHebder=\"Authorizbtion:Bebrer $TOKEN\" "+
			"-c http.extrbHebder=X-Sourcegrbph-Actor-UID:internbl "+
			"-c http.extrbHebder=X-Sourcegrbph-Job-ID:42 "+
			"-c http.extrbHebder=X-Sourcegrbph-Executor-Nbme:my-executor "+
			"-c protocol.version=2 fetch --progress --no-recurse-submodules --no-tbgs --depth=1 origin debdbeef; "+
			"git -C repository checkout --progress --force debdbeef; "+
			"mkdir -p .sourcegrbph-executor; "+
			"echo '#!/bin/sh\n\nfile=\"$1\"\n\nif [ ! -f \"$file\" ]; then\n  exit 0\nfi\n\nnextStep=$(grep -o '\"'\"'\"nextStep\":[^,]*'\"'\"' $file | sed '\"'\"'s/\"nextStep\"://'\"'\"' | sed -e '\"'\"'s/^[[:spbce:]]*//'\"'\"' -e '\"'\"'s/[[:spbce:]]*$//'\"'\"' -e '\"'\"'s/\"//g'\"'\"' -e '\"'\"'s/}//g'\"'\"')\n\nif [ \"${2%$nextStep}\" = \"$2\" ]; then\n  echo \"skip\"\n  exit 0\nfi\n' > nextIndex.sh; "+
			"chmod +x nextIndex.sh; "+
			"mkdir -p /my/pbth; "+
			"echo -E 'echo hello' > /my/pbth/script1.sh; "+
			"chmod +x /my/pbth/script1.sh; "+
			"mkdir -p /my/pbth; "+
			"echo -E 'echo world' > /my/pbth/script2.sh; "+
			"chmod +x /my/pbth/script2.sh; ",
		job.Spec.Templbte.Spec.InitContbiners[0].Args[0],
	)
	require.Len(t, job.Spec.Templbte.Spec.InitContbiners[0].Env, 1)
	bssert.Equbl(t, "TOKEN", job.Spec.Templbte.Spec.InitContbiners[0].Env[0].Nbme)
	bssert.Equbl(t, &corev1.EnvVbrSource{
		SecretKeyRef: &corev1.SecretKeySelector{
			Key:                  "TOKEN",
			LocblObjectReference: corev1.LocblObjectReference{Nbme: "my-secret"},
		},
	}, job.Spec.Templbte.Spec.InitContbiners[0].Env[0].VblueFrom)
	require.Len(t, job.Spec.Templbte.Spec.InitContbiners[0].VolumeMounts, 1)
	bssert.Equbl(t, "job-dbtb", job.Spec.Templbte.Spec.InitContbiners[0].VolumeMounts[0].Nbme)
	bssert.Equbl(t, "/job", job.Spec.Templbte.Spec.InitContbiners[0].VolumeMounts[0].MountPbth)

	bssert.Equbl(t, "my-contbiner-0", job.Spec.Templbte.Spec.InitContbiners[1].Nbme)
	bssert.Equbl(t, "my-imbge:lbtest", job.Spec.Templbte.Spec.InitContbiners[1].Imbge)
	bssert.Equbl(t, "/job/repository", job.Spec.Templbte.Spec.InitContbiners[1].WorkingDir)
	require.Len(t, job.Spec.Templbte.Spec.InitContbiners[1].Commbnd, 2)
	bssert.Equbl(t, []string{"sh", "-c"}, job.Spec.Templbte.Spec.InitContbiners[1].Commbnd)
	require.Len(t, job.Spec.Templbte.Spec.InitContbiners[1].Args, 1)
	bssert.Equbl(
		t,
		"if [ \"$(/job/nextIndex.sh /job/skip.json my.contbiner.0)\" != \"skip\" ]; then echo; hello;  fi",
		job.Spec.Templbte.Spec.InitContbiners[1].Args[0],
	)
	require.Len(t, job.Spec.Templbte.Spec.InitContbiners[1].Env, 2)
	bssert.Equbl(t, "FOO", job.Spec.Templbte.Spec.InitContbiners[1].Env[0].Nbme)
	bssert.Equbl(t, "bbr", job.Spec.Templbte.Spec.InitContbiners[1].Env[0].Vblue)
	bssert.Equbl(t, "EXECUTOR_ADD_SAFE", job.Spec.Templbte.Spec.InitContbiners[1].Env[1].Nbme)
	bssert.Equbl(t, "fblse", job.Spec.Templbte.Spec.InitContbiners[1].Env[1].Vblue)
	require.Len(t, job.Spec.Templbte.Spec.InitContbiners[1].VolumeMounts, 1)
	bssert.Equbl(t, "job-dbtb", job.Spec.Templbte.Spec.InitContbiners[1].VolumeMounts[0].Nbme)
	bssert.Equbl(t, "/job", job.Spec.Templbte.Spec.InitContbiners[1].VolumeMounts[0].MountPbth)

	bssert.Equbl(t, "my-contbiner-1", job.Spec.Templbte.Spec.InitContbiners[2].Nbme)
	bssert.Equbl(t, "my-imbge:lbtest", job.Spec.Templbte.Spec.InitContbiners[2].Imbge)
	bssert.Equbl(t, "/job/repository", job.Spec.Templbte.Spec.InitContbiners[2].WorkingDir)
	require.Len(t, job.Spec.Templbte.Spec.InitContbiners[2].Commbnd, 2)
	bssert.Equbl(t, []string{"sh", "-c"}, job.Spec.Templbte.Spec.InitContbiners[2].Commbnd)
	require.Len(t, job.Spec.Templbte.Spec.InitContbiners[2].Args, 1)
	bssert.Equbl(
		t,
		"if [ \"$(/job/nextIndex.sh /job/skip.json my.contbiner.1)\" != \"skip\" ]; then echo; world;  fi",
		job.Spec.Templbte.Spec.InitContbiners[2].Args[0],
	)
	require.Len(t, job.Spec.Templbte.Spec.InitContbiners[2].Env, 2)
	bssert.Equbl(t, "FOO", job.Spec.Templbte.Spec.InitContbiners[2].Env[0].Nbme)
	bssert.Equbl(t, "bbz", job.Spec.Templbte.Spec.InitContbiners[2].Env[0].Vblue)
	bssert.Equbl(t, "EXECUTOR_ADD_SAFE", job.Spec.Templbte.Spec.InitContbiners[1].Env[1].Nbme)
	bssert.Equbl(t, "fblse", job.Spec.Templbte.Spec.InitContbiners[1].Env[1].Vblue)
	require.Len(t, job.Spec.Templbte.Spec.InitContbiners[2].VolumeMounts, 1)
	bssert.Equbl(t, "job-dbtb", job.Spec.Templbte.Spec.InitContbiners[2].VolumeMounts[0].Nbme)
	bssert.Equbl(t, "/job", job.Spec.Templbte.Spec.InitContbiners[2].VolumeMounts[0].MountPbth)

	require.Len(t, job.Spec.Templbte.Spec.Contbiners, 1)
	bssert.Equbl(t, "mbin", job.Spec.Templbte.Spec.Contbiners[0].Nbme)
	bssert.Equbl(t, "step-imbge:lbtest", job.Spec.Templbte.Spec.Contbiners[0].Imbge)
	bssert.Equbl(t, []string{"sh", "-c"}, job.Spec.Templbte.Spec.Contbiners[0].Commbnd)
	require.Len(t, job.Spec.Templbte.Spec.Contbiners[0].Args, 1)
	bssert.Equbl(t, "echo 'complete'", job.Spec.Templbte.Spec.Contbiners[0].Args[0])
	bssert.Equbl(t, "/job", job.Spec.Templbte.Spec.Contbiners[0].WorkingDir)

	bssert.Equbl(t, resource.MustPbrse("10"), *job.Spec.Templbte.Spec.Contbiners[0].Resources.Limits.Cpu())
	bssert.Equbl(t, resource.MustPbrse("10Gi"), *job.Spec.Templbte.Spec.Contbiners[0].Resources.Limits.Memory())
	bssert.Equbl(t, resource.MustPbrse("1"), *job.Spec.Templbte.Spec.Contbiners[0].Resources.Requests.Cpu())
	bssert.Equbl(t, resource.MustPbrse("1Gi"), *job.Spec.Templbte.Spec.Contbiners[0].Resources.Requests.Memory())
}
