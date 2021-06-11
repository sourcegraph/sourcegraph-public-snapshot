# Configuring Sourcegraph executors to compute Batch Changes

> Note: This is a draft documentation page describing the potential end state of a feature. It should serve only for discussion purposes.

By default, Batch Changes uses a command line interface in your local environment to compute diffs and create changesets. This can be impractical for creating batch changes affecting hundreds or thousands of repositories, with large numbers of workspaces, or if the batch change steps require CPU, memory, or disk resources that are unavailable locally.

<!-- aharvey: This is definitely not a sentence we'd generally write in a real documentation page of this nature, but I get the framing. -->
If you are on a developer tool team rolling out batch changes to your company, the requirement to create batch changes locally might make adoption more difficult for some of your users.

<!-- aharvey: Trying to make the point here that this is also (eventually) useful for Code Intel. -->
**The Sourcegraph executor is an application that allows users to offload expensive tasks, such as computing Batch Changes and generating LSIF data for [precise code intelligence](../../code_intelligence/explanations/precise_code_intelligence.md).**

- If you use Sourcegraph enterprise on-premise, you can install Sourcegraph executors on infrastructure you own or manage.
- Sourcegraph.com offers a batch change execution environment based on Sourcegraph executor. You don't have to do anything to use it.

The Sourcegraph executor is written in [Go](https://golang.org), and is available to [Sourcegraph Enterprise](../../admin/subscriptions.md) customers.

## Getting started

1. [Ensure you meet the prerequisites.](#prerequisites)
1. [Choose your deployment model.](#deployment-models)

Then follow the instructions for your chosen deployment model:

* [Docker containers running on a server](pure-docker.md)
* [Jobs running in a Kubernetes cluster](kubernetes.md)

You can also read about [how Sourcegraph executors work](#how-executors-work).

## Prerequisites

- Sourcegraph executors can be installed on any operating system that is supported by Go, but we recommend and test primarily against Linux.
- As Sourcegraph executor tasks are run in containers, access to either Docker or Kubernetes is required.
- Sourcegraph executor usage requires requires Sourcegraph version 3.XX.
- As Sourcegraph executors run arbitrary user-submitted code, take care to place the executors within an appropriately secured part of your infrastructure. Network access to Sourcegraph is required, and your users will likely need read only access to internal resources required to run typical Batch Changes, such as internal package repositories and proxies. In general, you should trust Sourcegraph executors as much as you trust your Sourcegraph users to spawn and run services within your infrastructure.

## Deployment models

Two deployment models are supported:

1. [**Docker containers running on a server.**](pure-docker.md) In this model, changesets are computed in Docker containers run directly on the server. This model fits well in environments that do not use Kubernetes, and that expect relatively steady usage of Batch Changes over time.
1. 

The linked pages provide instructions on installing and configuring Sourcegraph executors in their respective environments.

## Installing the executor

The executor can be installed in one of the following ways:

- [In a container](#in-a-container)
- [From an RPM/DEB repository](#from-an-rpm-deb-repository)
- [By downloading a binary manually](#by-downloading-a-binary-manually)

### In a container

We support running the container in [Docker](#docker) or [Kubernetes](#kubernetes) via a Helm chart.

#### Docker

You can start the container like any other Docker image:

```sh
docker run -d --name sourcegraph-executor --restart always \
  sourcegraph/executor:latest
```

If you intend to run executor tasks on the same Docker instance that is running the executor, then you'll also need to map in the Docker socket:

```sh
docker run -d --name sourcegraph-executor --restart always \
  -v /var/run/docker.sock:/var/run/docker.sock \
  sourcegraph/executor:latest
```

#### Kubernetes

_TODO: I think we're going to end up with something like https://docs.gitlab.com/executor/install/kubernetes.html here: we ship a Helm chart. I'd love Distribution to weigh in on this, though._

### From an RPM/DEB repository

<!-- aharvey: I've done this for a living in the past, and I don't consider this part of the MVP because it's painful, but it _is_ going to appeal strongly to a certain vintage of admin. (That is, people my age and older.) -->

Sourcegraph provides packages for supported versions of [Debian, Ubuntu](#debian), and [RHEL](#rhel). You can [install the binary manually](#by-downloading-a-binary-manually) on other Linux distributions, and on other operating systems.

#### Debian

1. Add the Sourcegraph repository:

    ```
    curl https://packages.sourcegraph.com/sourcegraph.gpg | apt-key add -
    add-apt-repository "deb https://packages.sourcegraph.com/debian $(lsb_release -c) non-free"
    ```

2. Install the `sourcegraph-executor` package:

    ```
    apt update
    apt install sourcegraph-executor
    ```

#### RHEL

<!-- TODO: as above, but yum flavoured -->

### By downloading a binary manually

You can download the current version of the executor binary for these platforms:

- [Windows x86-64](TODO)
- [macOS x86-64](TODO)
- [macOS ARM](TODO)
- [Linux x86-64](TODO)

Once downloaded, you should configure your operating system to execute this binary as a service, noting that you will need to configure environment variables [in the next step](#configuring-the-executor).

## Configuring the executor

The Sourcegraph executor is configured through environment variables. The first three variables below are required; others are optional.

| Variable | Default value | Description |
|----------|---------------|-------------|
| `EXECUTOR_QUEUE_NAME` | **No default; must be provided** | The name of the queue to listen to; `batches` for Batch Changes. |
| `EXECUTOR_FRONTEND_URL` | **No default; must be provided** | The Sourcegraph URL; eg `https://sourcegraph.com`. |
| `EXECUTOR_REGISTRATION_TOKEN` | **No default; must be provided** | The executor registration token, as downloaded from the [Sourcegraph site settings](TODO). |
| `EXECUTOR_BACKEND` | `docker` | The backend that should be used to execute tasks; either `docker` or `kubernetes`. |
| `EXECUTOR_MAX_NUM_JOBS` | `1` | The maximum number of jobs (or tasks) that will be run concurrently by this executor. |
<!-- aharvey: There are lots of othe  r options at https://sourcegraph.com/github.com/sourcegraph/sourcegraph@3b1fbde4e2207de103a6736706bbfd0adaa579b6/-/blob/enterprise/cmd/executor/config.go#L36-52; most aren't super relevant for user facing documentation. What we _will_ need is whatever options we need to wire up executing jobs on k8s. -->

<!-- aharvey: TODO: I might want to replace this with "connecting". -->
## Registering the executor

The Sourcegraph executor should register itself automatically when started with a valid `EXECUTOR_FRONTEND_URL` and `EXECUTOR_REGISTRATION_TOKEN`, as described in [configuring the executor](#configuring-the-executor). Once registered, it will appear in the executors page within the Sourcegraph site admin, and the following output will appear in the container log:

```
✱ Sourcegraph executor connected to https://sourcegraph.com; listening for jobs
```

If the executor fails to register, check the output of the executor container: it should include an error that points towards the problem. If not, please contact Sourcegraph support!

# Using executors

If one or several executors are registered with the Sourcegraph instance, they will be used to create changesets by default. You can choose to use your local environment to compute the changeset by adding the `--local` flag to the CLI command.

When running a batch change, the executor will attempt to pull missing container images. The environment running the executor must have access to the docker registry hosting the requested images.

## Who has access to executors

All batch changes users on your Sourcegraph instance have access to executors.

## Debugging

TODO: mockups or decription

## Scheduling

Jobs are executed on a FIFO-basis. Users can interrupt a job from the interface.

Each executor will run one job at a time: you should scale the number of executors you have running to the maximum concurrency you wish to support. <!-- TODO: get some input on autoscaling for k8s -->

# Administering and monitoring executors

TODO

# How executors work

TODO

# Executor vs CLI workflow comparison

With the CLI workflow, the changeset creation step has to be ran locally, which can take a long time for large, complex batch changes. With the executor workflow, the changeset creation step is offloaded to the executor, which means that all the steps can happen in the same user interface, and processing is offloaded to the executor.

<img src="https://mermaid.ink/svg/eyJjb2RlIjoiZ3JhcGggTFJcbiAgICBBKEluc3RhbGwgc3JjIENMSSkgLS0-IEIoV3JpdGUgc3BlYylcbiAgICAgLS0-IEMoUnVuIENMSSB0byA8YnIvPiBjcmVhdGUgY2hhbmdlc2V0cylcbiAgICAgLS0-IEQoUHJldmlldylcbiAgICAgLS0-IEUoQXBwbHkpXG4gICAgIC0tPiBGKFRyYWNrIHByb2dyZXNzKVxuICAgICBHKCAgKSAtLT4gSChXcml0ZSBzcGVjKSBcbiAgICAgLS0-IEkoRXhlY3V0b3IgY3JlYXRlcyA8YnIvPiBjaGFuZ2VzZXRzKVxuICAgICAtLT4gSihQcmV2aWV3KVxuICAgICAtLT4gSyhBcHBseSlcbiAgICAgLS0-IEwoVHJhY2sgcHJvZ3Jlc3MpXG4gICAgIE0obG9jYWxseSlcbiAgICAgTihvbiBTb3VyY2VncmFwaClcblxuc3R5bGUgTSBzdHJva2U6I0ZGNTU0Mywgc3Ryb2tlLXdpZHRoOjNweFxuc3R5bGUgQSBzdHJva2U6I0ZGNTU0Mywgc3Ryb2tlLXdpZHRoOjNweFxuc3R5bGUgQiBzdHJva2U6I0ZGNTU0Mywgc3Ryb2tlLXdpZHRoOjNweFxuc3R5bGUgQyBzdHJva2U6I0ZGNTU0Mywgc3Ryb2tlLXdpZHRoOjNweFxuXG5zdHlsZSBOIHN0cm9rZTojQTExMkZGLCBzdHJva2Utd2lkdGg6M3B4XG5zdHlsZSBEIHN0cm9rZTojQTExMkZGLCBzdHJva2Utd2lkdGg6M3B4XG5zdHlsZSBFIHN0cm9rZTojQTExMkZGLCBzdHJva2Utd2lkdGg6M3B4XG5zdHlsZSBGIHN0cm9rZTojQTExMkZGLCBzdHJva2Utd2lkdGg6M3B4XG5zdHlsZSBHIHN0cm9rZTojRkZGRkZGLCBmaWxsOiNGRkZGRkZcbnN0eWxlIEggc3Ryb2tlOiNBMTEyRkYsIHN0cm9rZS13aWR0aDozcHhcbnN0eWxlIEkgc3Ryb2tlOiNBMTEyRkYsIHN0cm9rZS13aWR0aDozcHhcbnN0eWxlIEogc3Ryb2tlOiNBMTEyRkYsIHN0cm9rZS13aWR0aDozcHhcbnN0eWxlIEsgc3Ryb2tlOiNBMTEyRkYsIHN0cm9rZS13aWR0aDozcHhcbnN0eWxlIEwgc3Ryb2tlOiNBMTEyRkYsIHN0cm9rZS13aWR0aDozcHhcblxubGlua1N0eWxlIDUgc3Ryb2tlLXdpZHRoOjBweFxuIiwibWVybWFpZCI6eyJ0aGVtZSI6ImRlZmF1bHQifSwidXBkYXRlRWRpdG9yIjpmYWxzZX0">

# Limitations

The current version of Sourcegraph executors has known limitations

- access control: if a executor is enabled, all Batch Changes users on the instance can submit jobs to it. In future versions, we may allow site admin to authorize only a group of users.

# FAQ

#### Can large batch changes execution be distributed on multiple executors?
Answer

#### I have several machines configured as executors, and they don't have the same specs (eg. memory). Can I submit some bacth changes specifically to a given machine?
Answer

#### What happens if execution fails?
Answer
