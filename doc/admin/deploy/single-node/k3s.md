---
title: Install Sourcegraph locally with K3s
---

# Install Sourcegraph with K3s

This guide will take you through how to set up a Sourcegraph instance locally with [K3s](https://k3s.io/), a tool that lets you run a single-node Kubernetes cluster on your local machine, where we will deploy our Sourcegraph instance to using Sourcegraph Helm Charts.

## Prerequisites

Following are the prerequisites for running Sourcegraph with [K3s](https://k3s.io/) on your Linux machine.

- Ubuntu 18.04 or above
- Minimum of **8 CPU** and **32GB memory** available

The scripts below will install the following on your machine:

- [K3s](https://k3s.io/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [Helm](https://helm.sh/docs/intro/install/)
- [sourcegraph/deploy repository](https://github.com/sourcegraph/deploy)

## Deploy

Run the following scripts:

##### Start up script

This will always start Sourcegraph at the latest version.

```bash
curl -sfL https://raw.githubusercontent.com/sourcegraph/deploy/main/install/scripts/k3s/local.sh | bash
```

To start Sourcegraph at a specific version, add version number at the end of the curl command after `-`:

```bash
curl -sfL https://raw.githubusercontent.com/sourcegraph/deploy/main/install/scripts/k3s/local.sh | bash - v4.1.3
```


## Upgrade

Please refer to the [upgrade docs for all Sourcegraph Helm instances](../kubernetes/operations.md#upgrade).

## Downgrade

See instructions for upgrades.

## Uninstall

See the [official K3s docs](https://docs.k3s.io/installation/uninstall) for detailed instructions on uninstalling K3s.
