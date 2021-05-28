# How to restart a docker-compose managed instance

This document will take you through how to restart a managed instance deployed on the cloud.

Typically, the error you receive looks like:

`Server restart is required for the configuration to take effect.`
## Prerequisites
This document assumes that you have permissions to access customer instances deployed on GCP. This document also assumes that you have basic knowledge of Unix/Linux, docker commands and the GCP UI
## Steps to resolve
1. Navigate to the instance that you'd like to access in the `sourcegraph-managed-customer` project(the instance is either the `default-red-instance` or `default-black instance`
2. SSH into the instance within GCP via the option provided.
3. `cd` to the `deployment` folder and run:
4. `docker compose restart sourcegraph-frontend-0 sourcegraph-frontend-internal`

## Further resources
[Sourcegraph Managed Operations](https://https://about.sourcegraph.com/handbook/engineering/distribution/managed/operations)
