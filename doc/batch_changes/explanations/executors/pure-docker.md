# Running Sourcegraph executors on bare servers

In this model, changesets are computed in Docker containers run directly on the server.

## Pros

* Deployment is simple.
* I/O performance can be extremely high if physically attached storage is used.

## Cons

* If required, scaling must be managed outside the container environment through something like EC2 auto scaling.
* Resource limits are harder to apply on individual changeset steps.

## Installation

<!--
aharvey: I originally had a few paragraphs on packaged RPM/DEB deployments. I think we might _eventually_ have demand for that because it's easier to manage upgrades, but it's far enough down the road that I don't want to get into it here. What might be a middle ground before that would be shipping AMIs with extremely bare bones Debian images with just Docker and the executor preconfigured in systemd.

For now I've also skipped any mention of Windows or macOS, but there's nothing stopping us from shipping those. (Although it's probably of relatively little use, given the hard Docker requirement; why not just run Linux natively?)
-->

The Sourcegraph executor is a service, written in Go, that is shipped as a static binary for Linux. The service must be configured to start after Docker, and is [configured](#configuration) through environment variables.

On a systemd-based distribution, the following steps will install the executor:

1. Download the executor from https://github.com/sourcegraph/sourcegraph/releases/XXX/sourcegraph-executor.tar.gz.
2. Extract it to a suitable location using tar; eg: `tar -C /opt/sourcegraph -zxf sourcegraph-executor.tar.gz`
3. Add a systemd service to start the executor; for example by creating the below unit in `/etc/systemd/system/sourcegraph-executor.service` and enabling it with `systemctl enable sourcegraph-executor`:

    ```
    [Unit]
    Description=Sourcegraph executor
    After=docker.service
    Requires=docker.socket

    [Service]
    Type=simple
    ExecStart=/opt/sourcegraph/bin/executor
    EnvironmentFile=/opt/sourcegraph/etc/executor.env

    [Install]
    WantedBy=multi-user.target
    ```
4. [Configure the executor](#configuration) by setting the required environment variables in the environment file (`/opt/sourcegraph/etc/executor.env`, in this example).
5. Start the executor with `systemctl start sourcegraph-executor`.

## Configuration

TODO: environment variables