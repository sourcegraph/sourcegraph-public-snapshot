# Sourcegraph Helm Chart

## Requirements

* [Helm 3 CLI](https://helm.sh/docs/intro/install/)
* Kubernetes 1.19 or greater

## Quickstart

To use the Helm chart, add Sourcegraph helm repository:
 
```sh
helm repo add sourcegraph https://sourcegraph.github.io/deploy-sourcegraph-helm/
```

Install the Sourcegraph chart using default values:

```sh
helm install sourcegraph sourcegraph/sourcegraph
```

## Configuration guide

Helm customizations can be applied using an override file. Using an override file allows customizations to persist through upgrades without needing to manage merge conflicts.

To customize configuration settings with an override file, create an empty yaml file (e.g. `override.yaml`) and configure overrides.

> WARNING: __DO NOT__ copy the [default values file](https://github.com/sourcegraph/deploy-sourcegraph-helm/blob/main/charts/sourcegraph/values.yaml) as a boilerplate for your override file. You will be risk having outdated values during upgrade.

Example overrides can be found in the [examples](https://github.com/sourcegraph/deploy-sourcegraph-helm/tree/main/charts/sourcegraph/examples) folder. Please take a look at our examples before providing your own configuration and consider using them as boilerplates.

Provide the override file to helm:

```sh
# Installation
helm install --values ./override.yaml sourcegraph sourcegraph/sourcegraph

# Upgrade
helm upgrade --values ./override.yaml sourcegraph sourcegraph/sourcegraph
```

## Configuration options

The Sourcegraph chart is highly customizable to support a wide-range of environment. Please review the default values from [values.yaml](https://github.com/sourcegraph/deploy-sourcegraph-helm/blob/main/charts/sourcegraph/values.yaml) and all [supported options](https://github.com/sourcegraph/deploy-sourcegraph-helm/tree/main/charts/sourcegraph#configuration-options).

## Advanced configuration

The Helm chart is new and still under active development, and we may not cover all your use cases. 

Please reach out to your account team to discuess your specific need.

For advanced users who are looking for a temporary workaround, we __recommend__ applying [Kustomize](https://kustomize.io) on the rendered manifests from our chart. Plesae __do not__ maintain your own fork of our chart, this may impact our ability to support you if you run into issues.

You can learn more about how to integrate Kustomize with Helm from our [example](https://github.com/sourcegraph/deploy-sourcegraph-helm/tree/main/charts/sourcegraph/examples/kustomize-chart).
