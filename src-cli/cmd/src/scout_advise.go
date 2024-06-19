package main

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"

	"github.com/sourcegraph/src-cli/internal/scout/advise"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func init() {
	cmdUsage := `'src scout advise' is a tool that makes resource allocation recommendations. Based on current usage.
    Part of the EXPERIMENTAL "src scout" tool.
    
    Examples
        Make recommendations for all pods in a kubernetes deployment of Sourcegraph.
        $ src scout advise
        
        Make recommendations for specific pod:
        $ src scout advise --pod <podname>

        Add namespace if using namespace in a Kubernetes cluster
        $ src scout advise --namespace <namespace>

        Output advice to file
        $ src scout advise --o path/to/file

        Output with warnings
        $ src scout advise --warnings
    `

	flagSet := flag.NewFlagSet("advise", flag.ExitOnError)
	usage := func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of 'src scout %s':\n", flagSet.Name())
		flagSet.PrintDefaults()
		fmt.Println(cmdUsage)
	}

	var (
		kubeConfig *string
		namespace  = flagSet.String("namespace", "", "(optional) specify the kubernetes namespace to use")
		pod        = flagSet.String("pod", "", "(optional) specify a single pod")
		output     = flagSet.String("o", "", "(optional) output advice to file")
		warnings   = flagSet.Bool("warnings", false, "(optional) output advice with warnings")
	)

	if home := homedir.HomeDir(); home != "" {
		kubeConfig = flagSet.String(
			"kubeconfig",
			filepath.Join(home, ".kube", "config"),
			"(optional) absolute path to the kubeconfig file",
		)
	} else {
		kubeConfig = flagSet.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	handler := func(args []string) error {
		if err := flagSet.Parse(args); err != nil {
			return err
		}

		config, err := clientcmd.BuildConfigFromFlags("", *kubeConfig)
		if err != nil {
			return errors.Wrap(err, "failed to load .kube config: ")
		}

		clientSet, err := kubernetes.NewForConfig(config)
		if err != nil {
			return errors.Wrap(err, "failed to initiate kubernetes client: ")
		}

		metricsClient, err := metricsv.NewForConfig(config)
		if err != nil {
			return errors.Wrap(err, "failed to initiate metrics client")
		}

		var options []advise.Option

		if *namespace != "" {
			options = append(options, advise.WithNamespace(*namespace))
		}
		if *pod != "" {
			options = append(options, advise.WithPod(*pod))
		}
		if *output != "" {
			options = append(options, advise.WithOutput(*output))
		}
		if *warnings {
			options = append(options, advise.WithWarnings(true))
		}

		return advise.K8s(
			context.Background(),
			clientSet,
			metricsClient,
			config,
			options...,
		)
	}

	scoutCommands = append(scoutCommands, &command{
		flagSet:   flagSet,
		handler:   handler,
		usageFunc: usage,
	})
}
