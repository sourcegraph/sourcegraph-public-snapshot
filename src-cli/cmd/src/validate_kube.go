package main

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	"github.com/sourcegraph/src-cli/internal/validate/kube"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func init() {
	usage := `'src validate kube' is a tool that validates a Kubernetes based Sourcegraph deployment
	
Examples:

	Run default deployment validation:
		$ src validate kube
		
	Specify Kubernetes namespace:
		$ src validate kube --namespace sourcegraph
		
	Specify the kubeconfig file())) location:
		$ src validate kube --kubeconfig ~/.kube/config
	
	Suppress output (useful for CI/CD pipelines)
		$ src validate kube --quiet

    Validate EKS cluster:
        $ src validate kube --eks
        
    Validate GKE cluster:
        $ src validate kube --gke
        
    Validate AKS cluster:
        $ src validate kube --aks
`

	flagSet := flag.NewFlagSet("kube", flag.ExitOnError)
	usageFunc := func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of 'src validate %s':\n", flagSet.Name())
		flagSet.PrintDefaults()
		fmt.Println(usage)
	}

	var (
		kubeConfig *string
		namespace  = flagSet.String("namespace", "", "(optional) specify the kubernetes namespace to use")
		quiet      = flagSet.Bool("quiet", false, "(optional) suppress output and return exit status only")
		eks        = flagSet.Bool("eks", false, "(optional) validate EKS cluster")
		gke        = flagSet.Bool("gke", false, "(optional) validate GKE cluster")
		aks        = flagSet.Bool("aks", false, "(optional) validate AKS cluster")
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
		ctx := context.Background()
		if err := flagSet.Parse(args); err != nil {
			return err
		}

		// use the current context in kubeConfig
		config, err := clientcmd.BuildConfigFromFlags("", *kubeConfig)
		if err != nil {
			return errors.Wrap(err, "failed to load kubernetes config")
		}

		clientSet, err := kubernetes.NewForConfig(config)
		if err != nil {
			return errors.Wrap(err, "failed to create kubernetes client")
		}

		// parse through flag options
		var options []kube.Option

		if *namespace != "" {
			options = append(options, kube.WithNamespace(*namespace))
		}

		if *quiet {
			options = append(options, kube.Quiet())
		}

		if *eks {
			options = append(options, kube.GenerateAWSClients(ctx))
		}

		if *gke {
			options = append(options, kube.Gke())
		}

		if *aks {
			options = append(options, kube.Aks())
		}

		return kube.Validate(context.Background(), clientSet, config, options...)
	}

	validateCommands = append(validateCommands, &command{
		flagSet:   flagSet,
		handler:   handler,
		usageFunc: usageFunc,
	})
}
