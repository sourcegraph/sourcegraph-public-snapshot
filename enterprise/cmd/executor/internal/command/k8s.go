package command

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strings"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/lytics/base62"
	"github.com/pkg/errors"
	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	batchclient "k8s.io/client-go/kubernetes/typed/batch/v1"
	coreclient "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type kubernetesRunner struct {
	jobs batchclient.JobInterface
	pods coreclient.PodInterface
	pv   coreclient.PersistentVolumeInterface
	pvc  coreclient.PersistentVolumeClaimInterface

	logger  *Logger
	options Options
}

var _ Runner = &kubernetesRunner{}

func newKubernetesRunner(configPath, namespace string, logger *Logger, options Options) (*kubernetesRunner, error) {
	config, err := clientcmd.BuildConfigFromFlags("", configPath)
	if err != nil {
		return nil, errors.Wrap(err, "building config")
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "creating client set")
	}

	core := clientset.CoreV1()

	return &kubernetesRunner{
		jobs:    clientset.BatchV1().Jobs(namespace),
		pods:    core.Pods(namespace),
		pv:      core.PersistentVolumes(),
		pvc:     core.PersistentVolumeClaims(namespace),
		logger:  logger,
		options: options,
	}, nil
}

func (r *kubernetesRunner) Setup(ctx context.Context, imageNames, scriptPaths []string) error {
	// Set up a persistent volume claim.
	pv := &apiv1.PersistentVolume{}
}

func (r *kubernetesRunner) Teardown(ctx context.Context) error {
	// Release the persistent volume claim.
	return nil
}

func (r *kubernetesRunner) Run(ctx context.Context, command CommandSpec) (err error) {
	ctx, endObservation := command.Operation.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	log15.Info(fmt.Sprintf("Preparing k8s job for command: %s", strings.Join(command.Command, " ")))

	// Create a unique name so we can track the job and its pods.
	hashed := sha256.Sum256([]byte(fmt.Sprintf("%v", time.Now())))
	jobName := fmt.Sprintf("job-%s", base62.StdEncoding.EncodeToString(hashed[:]))

	job := batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: jobName,
		},
		Spec: batchv1.JobSpec{
			// We never want failed pods to retry; just fail and we can handle
			// it from there.
			BackoffLimit: int32Ptr(1),
			Template: apiv1.PodTemplateSpec{
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:    command.Key,
							Image:   command.Image,
							Command: command.Command,
						},
					},
					RestartPolicy: apiv1.RestartPolicyNever,
				},
			},
		},
	}

	// TODO: copy the rest of this across from scratch/k8s/job/main.go
}

func int32Ptr(i int32) *int32 { return &i }
