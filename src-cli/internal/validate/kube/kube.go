package kube

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/client-go/util/homedir"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/sourcegraph/src-cli/internal/validate"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var (
	sourcegraphFrontend    = regexp.MustCompile(`^sourcegraph-frontend-.*`)
	sourcegraphRepoUpdater = regexp.MustCompile(`^repo-updater-.*`)
	sourcegraphWorker      = regexp.MustCompile(`^worker-.*`)
)

type Option = func(config *Config)

type Config struct {
	namespace  string
	output     io.Writer
	exitStatus bool
	clientSet  *kubernetes.Clientset
	restConfig *rest.Config
	eks        bool
	gke        bool
	aks        bool
	eksClient  *eks.Client
	ec2Client  *ec2.Client
	iamClient  *iam.Client
}

func WithNamespace(namespace string) Option {
	return func(config *Config) {
		config.namespace = namespace
	}
}

func Quiet() Option {
	return func(config *Config) {
		config.output = io.Discard
		config.exitStatus = true
	}
}

type validation struct {
	Validate   func(ctx context.Context, config *Config) ([]validate.Result, error)
	WaitMsg    string
	SuccessMsg string
	ErrMsg     string
}

// Validate will call a series of validation functions in a table driven tests style.
func Validate(ctx context.Context, clientSet *kubernetes.Clientset, restConfig *rest.Config, opts ...Option) error {
	cfg := &Config{
		namespace:  "default",
		output:     os.Stdout,
		exitStatus: false,
		clientSet:  clientSet,
		restConfig: restConfig,
		eks:        false,
		gke:        false,
		aks:        false,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	log.SetOutput(cfg.output)

	validations := []validation{
		{Pods, "validating pods", "pods validated", "validating pods failed"},
		{Services, "validating services", "services validated", "validating services failed"},
		{PVCs, "validating pvcs", "pvcs validated", "validating pvcs failed"},
		{Connections, "validating connections", "connections validated", "validating connections failed"},
	}

	if cfg.eks {
		if err := CurrentContextSetTo("eks"); err != nil {
			return errors.Newf("%s %s", validate.FailureEmoji, err)
		}

		GenerateAWSClients(ctx)

		validations = append(validations, validation{
			Validate:   EksEbsCsiDrivers,
			WaitMsg:    "EKS: validating ebs-csi drivers",
			SuccessMsg: "EKS: ebs-csi drivers validated",
			ErrMsg:     "EKS: validating ebs-csi drivers failed",
		})

		validations = append(validations, validation{
			Validate:   EksVpc,
			WaitMsg:    "EKS: validating vpc",
			SuccessMsg: "EKS: vpc validated",
			ErrMsg:     "EKS: validating vpc failed",
		})
	}

	if cfg.gke {
		if err := CurrentContextSetTo("gke"); err != nil {
			return errors.Newf("%s %s", validate.FailureEmoji, err)
		}

		Gke()

		validations = append(validations, validation{
			Validate:   GkeGcePersistentDiskCSIDrivers,
			WaitMsg:    "GKE: validating persistent volumes",
			SuccessMsg: "GKE: persistent volumes validated",
			ErrMsg:     "GKE: validating peristent volumes failed",
		})
	}

	if cfg.aks {
		if err := CurrentContextSetTo("aks"); err != nil {
			return errors.Newf("%s %s", validate.FailureEmoji, err)
		}

		Aks()

		validations = append(validations, validation{
			Validate:   AksCsiDrivers,
			WaitMsg:    "AKS: validating persistent volumes",
			SuccessMsg: "AKS: persistent volumes validated",
			ErrMsg:     "AKS: validating persistent volumes failed",
		})
	}

	var totalFailCount int

	for _, v := range validations {
		log.Printf("%s %s...", validate.HourglassEmoji, v.WaitMsg)
		results, err := v.Validate(ctx, cfg)
		if err != nil {
			return errors.Wrapf(err, v.ErrMsg)
		}

		var failCount int
		var warnCount int
		var succCount int

		for _, r := range results {
			switch r.Status {
			case validate.Failure:
				log.Printf("  %s failure: %s", validate.FailureEmoji, r.Message)
				failCount++
			case validate.Warning:
				log.Printf("  %s warning: %s", validate.WarningSign, r.Message)
				warnCount++
			case validate.Success:
				succCount++
			}
		}

		if failCount > 0 || warnCount > 0 {
			log.Printf("\n%s %s", validate.FlashingLightEmoji, v.ErrMsg)
		}

		if failCount > 0 {
			log.Printf("  %s %d total failure(s)", validate.EmojiFingerPointRight, failCount)

			totalFailCount = totalFailCount + failCount
		}

		if warnCount > 0 {
			log.Printf("  %s %d total warning(s)", validate.EmojiFingerPointRight, warnCount)
		}

		if failCount == 0 && warnCount == 0 {
			log.Printf("%s %s!", validate.SuccessEmoji, v.SuccessMsg)
		}
	}

	if totalFailCount > 0 {
		return errors.Newf("validation failed: %d failures", totalFailCount)
	}

	return nil
}

// Pods will validate all pods in a given namespace.
func Pods(ctx context.Context, config *Config) ([]validate.Result, error) {
	pods, err := config.clientSet.CoreV1().Pods(config.namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var results []validate.Result

	if len(pods.Items) == 0 {
		results = append(results, validate.Result{
			Status: validate.Warning,
			Message: fmt.Sprintf(
				"No pods exist on namespace '%s'. check namespace/cluster",
				config.namespace,
			),
		})

		return results, nil
	}

	for _, pod := range pods.Items {
		r := validatePod(&pod)
		results = append(results, r...)
	}

	return results, nil
}

func validatePod(pod *corev1.Pod) []validate.Result {
	var results []validate.Result

	if pod.Name == "" {
		results = append(results, validate.Result{Status: validate.Failure, Message: "pod.Name is empty"})
	}

	if pod.Namespace == "" {
		results = append(results, validate.Result{Status: validate.Failure, Message: "pod.Namespace is empty"})
	}

	if len(pod.Spec.Containers) == 0 {
		results = append(results, validate.Result{Status: validate.Failure, Message: "spec.Containers is empty"})
	}

	switch pod.Status.Phase {
	case corev1.PodPending:
		results = append(results, validate.Result{
			Status:  validate.Failure,
			Message: fmt.Sprintf("pod '%s' has a status 'pending'", pod.Name),
		})
	case corev1.PodFailed:
		results = append(results, validate.Result{
			Status:  validate.Failure,
			Message: fmt.Sprintf("pod '%s' has a status 'failed'", pod.Name),
		})
	}

	for _, container := range pod.Spec.Containers {
		if container.Name == "" {
			results = append(results, validate.Result{
				Status:  validate.Failure,
				Message: fmt.Sprintf("container.Name is empty, pod '%s'", pod.Name),
			})
		}

		if container.Image == "" {
			results = append(results, validate.Result{
				Status:  validate.Failure,
				Message: fmt.Sprintf("container.Image is empty, pod '%s'", pod.Name),
			})
		}
	}

	for _, c := range pod.Status.ContainerStatuses {
		if !c.Ready {
			results = append(results, validate.Result{
				Status:  validate.Failure,
				Message: fmt.Sprintf("container '%s' is not ready, pod '%s'", c.ContainerID, pod.Name),
			})
		}

		if c.RestartCount > 50 {
			results = append(results, validate.Result{
				Status:  validate.Warning,
				Message: fmt.Sprintf("container '%s' has high restart count: %d restarts", c.ContainerID, c.RestartCount),
			})
		}
	}

	return results
}

// Services will validate all  services in a given namespace.
func Services(ctx context.Context, config *Config) ([]validate.Result, error) {
	services, err := config.clientSet.CoreV1().Services(config.namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var results []validate.Result

	if len(services.Items) <= 1 {
		results = append(results, validate.Result{
			Status: validate.Warning,
			Message: fmt.Sprintf(
				"unexpected number of services on namespace '%s'; check namespace/cluster",
				config.namespace,
			),
		})

		return results, nil
	}

	for _, service := range services.Items {
		r := validateService(&service)
		results = append(results, r...)
	}

	return results, nil
}

func validateService(service *corev1.Service) []validate.Result {
	var results []validate.Result

	if service.Name == "" {
		results = append(results, validate.Result{Status: validate.Failure, Message: "service.Name is empty"})
	}

	if service.Namespace == "" {
		results = append(results, validate.Result{Status: validate.Failure, Message: "service.Namespace is empty"})
	}

	if len(service.Spec.Ports) == 0 {
		results = append(results, validate.Result{Status: validate.Failure, Message: "service.Ports is empty"})
	}

	return results
}

// PVCs will validate all persistent volume claims on a given namespace
func PVCs(ctx context.Context, config *Config) ([]validate.Result, error) {
	pvcs, err := config.clientSet.CoreV1().PersistentVolumeClaims(config.namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var results []validate.Result

	if len(pvcs.Items) == 0 {
		results = append(results, validate.Result{
			Status: validate.Warning,
			Message: fmt.Sprintf(
				"no PVCs exist in namespace '%s'; check namespace/cluster",
				config.namespace,
			),
		})

		return results, nil
	}

	for _, pvc := range pvcs.Items {
		r := validatePVC(&pvc)
		results = append(results, r...)
	}

	return results, nil
}

func validatePVC(pvc *corev1.PersistentVolumeClaim) []validate.Result {
	var results []validate.Result

	if pvc.Name == "" {
		results = append(results, validate.Result{Status: validate.Failure, Message: "pvc.Name is empty"})
	}

	if pvc.Status.Phase != "Bound" {
		results = append(results, validate.Result{Status: validate.Failure, Message: "pvc.Status is not bound"})
	}

	return results
}

type connection struct {
	src  corev1.Pod
	dest []dest
}

type dest struct {
	addr string
	port string
}

// Connections will validate that Sourcegraph services can reach each other over the network.
func Connections(ctx context.Context, config *Config) ([]validate.Result, error) {
	var results []validate.Result
	var connections []connection

	pods, err := config.clientSet.CoreV1().Pods(config.namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	if len(pods.Items) == 0 {
		results = append(results, validate.Result{
			Status: validate.Warning,
			Message: fmt.Sprintf(
				"cannot check connections: zero pods exist in namespace '%s'",
				config.namespace,
			),
		})

		return results, nil
	}

	// iterate through pods looking for specific pod name prefixes, then construct
	// a relationship map between pods that should have connectivity with each other
	for _, pod := range pods.Items {
		switch name := pod.Name; {
		case sourcegraphFrontend.MatchString(name): // pod is one of the sourcegraph front-end pods
			connections = append(connections, connection{
				src: pod,
				dest: []dest{
					{
						addr: "pgsql",
						port: "5432",
					},
					{
						addr: "indexed-search",
						port: "6070",
					},
					{
						addr: "repo-updater",
						port: "3182",
					},
					{
						addr: "syntect-server",
						port: "9238",
					},
				},
			})
		case sourcegraphWorker.MatchString(name): // pod is a worker pod
			connections = append(connections, connection{
				src: pod,
				dest: []dest{
					{
						addr: "pgsql",
						port: "5432",
					},
				},
			})
		case sourcegraphRepoUpdater.MatchString(name):
			connections = append(connections, connection{
				src: pod,
				dest: []dest{
					{
						addr: "pgsql",
						port: "5432",
					},
				},
			})
		}
	}

	// use network relationships constructed above to test network connection for each relationship
	for _, c := range connections {
		for _, d := range c.dest {
			req := config.clientSet.CoreV1().RESTClient().Post().
				Resource("pods").
				Name(c.src.Name).
				Namespace(c.src.Namespace).
				SubResource("exec")

			req.VersionedParams(&corev1.PodExecOptions{
				Command: []string{"/usr/bin/nc", "-z", d.addr, d.port},
				Stdin:   false,
				Stdout:  true,
				Stderr:  true,
				TTY:     false,
			}, scheme.ParameterCodec)

			exec, err := remotecommand.NewSPDYExecutor(config.restConfig, "POST", req.URL())
			if err != nil {
				return nil, err
			}

			var stdout, stderr bytes.Buffer

			err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
				Stdout: &stdout,
				Stderr: &stderr,
			})
			if err != nil {
				return nil, err
			}

			if stderr.String() != "" {
				results = append(results, validate.Result{Status: validate.Failure, Message: stderr.String()})
			}
		}
	}

	return results, nil
}

func CurrentContextSetTo(clusterService string) error {
	currentContext, err := GetCurrentContext()
	if err != nil {
		return err
	}

	if clusterService == "gke" {
		got := strings.Split(currentContext, "_")[0]
		want := clusterService
		if got != want {
			return errors.New("no gke cluster configured")
		}
	} else if clusterService == "eks" {
		got := strings.Split(currentContext, ":")
		want := []string{"arn", "aws", clusterService}

		if got[0] != "arn" {
			return errors.New("no eks cluster configured")
		}

		if len(got) >= 3 {
			got = got[:3]
			if !reflect.DeepEqual(got, want) {
				return errors.New("no eks cluster configured")
			}
		}
	} else if clusterService == "aks" {
		colons := strings.Split(currentContext, ":")      // aws string
		underscores := strings.Split(currentContext, "_") // gke string

		// if current context has 'arn' or 'gke' in string, return error
		if colons[0] == "arn" || underscores[0] == "gke" {
			return errors.New("no aks cluster configured")
		}
	}

	return nil
}

func GetCurrentContext() (string, error) {
	home := homedir.HomeDir()
	pathToKubeConfig := filepath.Join(home, ".kube", "config")

	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: pathToKubeConfig},
		&clientcmd.ConfigOverrides{
			CurrentContext: "",
		}).RawConfig()

	if err != nil {
		return "", err
	}

	return config.CurrentContext, nil
}
