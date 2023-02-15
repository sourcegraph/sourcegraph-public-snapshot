# `src validate kube`


## Flags

| Name | Description | Default Value |
|------|-------------|---------------|
| `-kubeconfig` | Absolute path to the kubeconfig file | `~/.kube/config` |
| `-namespace` | Specify the Kubernetes namespace to use | `default` |
| `-quiet` | Suppress output and return exit status only | `false` |


## Usage

```
Usage of 'src validate kube':
  -kubeconfig string
    	(optional) absolute path to the kubeconfig file (default "~/.kube/config")
  -namespace string
    	(optional) specify the kubernetes namespace to use
  -quiet
    	(optional) suppress output and return exit status only
'src validate kube' is a tool that validates a Kubernetes based Sourcegraph deployment

Examples:

	Run default deployment validation:
		$ src validate kube

	Specify Kubernetes namespace:
		$ src validate kube --namespace sourcegraph

	Specify the kubeconfig file location:
		$ src validate kube --kubeconfig ~/.kube/config

	Suppress output (useful for CI/CD pipelines)
		$ src validate kube --quiet
```