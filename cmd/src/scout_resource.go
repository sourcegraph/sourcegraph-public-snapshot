package main

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	"github.com/sourcegraph/src-cli/internal/scout/resource"
)

func init() {
	usage := `'src scout resource' is a tool that provides an overview of resource usage
    across an instance of Sourcegraph. Part of the EXPERIMENTAL "src scout" tool.
    
    Examples
        List pods and resource allocations in a Kubernetes deployment:
        $ src scout resource

        Add namespace if using namespace in a Kubernetes cluster
        $ src scout resource --namespace sg
    `

	flagSet := flag.NewFlagSet("resource", flag.ExitOnError)
	usageFunc := func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of 'src scout %s':\n", flagSet.Name())
		flagSet.PrintDefaults()
		fmt.Println(usage)
	}

	var (
		kubeConfig *string
		namespace  = flagSet.String("namespace", "", "(optional) specify the kubernetes namespace to use")
		// TODO: option for getting resource allocation of the Node
		// nodes      = flagSet.Bool("node", false, "(optional) view resources for node(s)")
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
			return errors.New(fmt.Sprintf("%v: failed to load kubernetes config", err))
		}

		clientSet, err := kubernetes.NewForConfig(config)
		if err != nil {
			return errors.New(fmt.Sprintf("%v: failed to load kubernetes config", err))
		}

		var options []resource.Option

		if *namespace != "" {
			options = append(options, resource.WithNamespace(*namespace))
		}

		return resource.K8s(context.Background(), clientSet, config, options...)
	}

	scoutCommands = append(scoutCommands, &command{
		flagSet:   flagSet,
		handler:   handler,
		usageFunc: usageFunc,
	})
}
