---
title: Install Sourcegraph locally with K3s
---

# Install Sourcegraph with K3s

This guide will take you through how to set up a Sourcegraph instance locally with [K3s](https://k3s.io/), a tool that lets you run a single-node Kubernetes cluster on your local machine, where we will deploy our Sourcegraph instance to using Sourcegraph Helm Charts.

## Prerequisites

Following are the prerequisites for running Sourcegraph with [K3s](https://k3s.io/) on your local machine.

- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [K3s](https://k3s.io/)
- [Helm](https://helm.sh/docs/intro/install/)
- Enable Kubernetes in Docker Desktop. See the [official docs](https://docs.docker.com/desktop/kubernetes/#enable-kubernetes) for detailed instruction.

> NOTE: Running Sourcegraph on k3s requires a minimum of **8 CPU** and **32GB memory** assigned to your Kubernetes instance.

## Deploy

## Upgrade

Please refer to the [upgrade docs for all Sourcegraph Helm instances](http://localhost:5080/admin/deploy/kubernetes/operations#upgrade).

## Downgrade

Same instruction as upgrades.

## Uninstall

