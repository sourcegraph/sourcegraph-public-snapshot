# Enable code intelligence auto-indexing

<aside class="experimental">
<p>
<span class="badge badge-experimental">Experimental</span> This feature is experimental and might change or be removed in the future. We've released it as an experimental feature to provide a preview of functionality we're working on.
</p>

<p><b>We're very much looking for input and feedback on this feature.</b> You can either <a href="https://about.sourcegraph.com/contact">contact us directly</a>, <a href="https://github.com/sourcegraph/sourcegraph">file an issue</a>, or <a href="https://twitter.com/sourcegraph">tweet at us</a>.</p>
</aside>

Let's walk through setting up a single executor VM on GCP and indexing a repository.

1. Install Terraform `0.13.7` (must be exact):

```
brew install tfenv
tfenv install 0.13.7
tfenv use 0.13.7
```

2. Install [`gcloud`](https://cloud.google.com/sdk/docs/install)
3. Run `gcloud auth application-default login`
4. Open your Sourcegraph instance in your browser, click your profile in the top right, click **Site admin**, expand **Configuration**, click **Site configuration**, and set:
  - `"externalURL": "<URL>"` to a URL that is accessible from the GCP VM that will be created later (e.g. a public URL such as `https://sourcegraph.acme.com`)
  - `"executors.accessToken": "<new long secret>"` to a new long secret (e.g. `cat /dev/random | base64 | head -c 20`)
  - `"codeIntelAutoIndexing.enabled": true`
5. Download the example [`main.tf`](https://github.com/sourcegraph/terraform-google-executors/blob/a0110747f70067b9b0d1c382334de02b86821ae3/examples/single-executor/main.tf) and change these:
  - `project`: your GCP project name and change `region` and `zone` if needed
  - `executor_sourcegraph_external_url`: this must match `externalURL` you set in your site config
  - `executor_sourcegraph_executor_proxy_password`: this must match `executors.accessToken` you set in your site config
6. Run `terraform init` to download the Sourcegraph executor modules
7. Run `terraform apply` and enter "yes" to create the executor VM
8. Go back to the site admin page, expand **Maintenance**, click **Executors**, and check to see if your executor shows up in the list with a green dot 🟢. If it's not there:
  - Make sure `terraform apply` exited with code 0 and did not print any errors
  - Make sure a GCP VM was created:

```
$ gcloud compute instances list
NAME                                          ZONE           MACHINE_TYPE   PREEMPTIBLE  INTERNAL_IP  EXTERNAL_IP    STATUS
sourcegraph-executor-h0rv                     us-central1-c  n1-standard-4               10.0.1.16    ...            RUNNING
sourcegraph-executors-docker-registry-mirror  us-central1-c  n1-standard-2               10.0.1.2     ...            RUNNING
```

  - Make sure the `executor` service is running:

```
you@sourcegraph-executor-h0rv:~$ systemctl status executor
🟢 executor.service - User code executor
     Loaded: loaded (/etc/systemd/system/executor.service; enabled; vendor preset: enabled)
     Active: active (running) since Thu 2021-11-18 02:28:48 UTC; 19s ago
```

  - Make sure there are no errors in the `executor` service logs:

```
you@sourcegraph-executor-h0rv:~$ journalctl -u executor | less
Nov 18 02:31:01 sourcegraph-executor-h0rv executor[2465]: t=2021-11-18T02:31:01+0000 lvl=dbug msg="TRACE internal" host=... path=/.executors/queue/codeintel/dequeue code=204 duration=92.131237ms
Nov 18 02:31:01 sourcegraph-executor-h0rv executor[2465]: t=2021-11-18T02:31:01+0000 lvl=dbug msg="TRACE internal" host=... path=/.executors/queue/codeintel/canceled code=200 duration=90.630467ms
Nov 18 02:31:02 sourcegraph-executor-h0rv executor[2465]: t=2021-11-18T02:31:02+0000 lvl=dbug msg="TRACE internal" host=... path=/.executors/queue/codeintel/dequeue code=204 duration=91.269106ms
Nov 18 02:31:02 sourcegraph-executor-h0rv executor[2465]: t=2021-11-18T02:31:02+0000 lvl=dbug msg="TRACE internal" host=... path=/.executors/queue/codeintel/canceled code=200 duration=161.469685ms
```

  - Make sure the `EXECUTOR_FRONTEND_URL` and `EXECUTOR_FRONTEND_PASSWORD` in `/etc/systemd/system/executor.env` are correct
  - Make sure the VM can hit your `externalURL`:

```
you@sourcegraph-executor-h0rv:~$ curl <your externalURL here>
<a href="/sign-in?returnTo=%2F">Found</a>
```

9. Go back to the site admin page, expand **Code intelligence**, click **Configuration**, click **Create new policy**, and fill in:
  - Name: `LSIF`
  - Click **add a repository pattern**
  - Repository pattern #1: set this to an existing repository on your Sourcegraph instance (e.g. `github.com/gorilla/mux`)
  - Type: `HEAD`
  - Auto-indexing: Enabled
10. Go to that repository's page, click **Code Intelligence**, click **Auto-indexing**, and check to see if an indexing job has appeared. If nothing is there:
  - Try clicking **Enqueue**
  - Try setting a higher update frequency: `PRECISE_CODE_INTEL_AUTO_INDEXING_TASK_INTERVAL=10s`
  - Try setting a lower delay: `PRECISE_CODE_INTEL_AUTO_INDEXING_REPOSITORY_PROCESS_DELAY=10s`
11. Once you have a completed indexing job, click **Uploads** and check to see that the LSIF dump has been uploaded.
12. Once the dump has been uploaded, you should see the **`PRECISE`** badge in the hover popover! 🎉
13. Optionally, add `.terraform`, `terraform.tfstate`, and `terraform.tfstate.backup` to your `.gitignore`.

OLD CONTENT:

## Deploy executors

First, [deploy the executor service](../../../../admin/deploy_executors.md) targeting your Sourcegraph instance. This will provide the necessary compute resources that clone the target Git repository, securely analyze the code to produce a precise code intelligence index, then upload that index to your Sourcegraph instance for processing.

## Enable index job scheduling

Next, enable the precise code intelligence auto-indexing feature by enabling the following feature flag in your Sourcegraph instance's site configuration.

```yaml
{
  "codeIntelAutoIndexing.enabled": true
}
```

This step will control the scheduling of indexing jobs which are made available to the executors deployed in the previous step.

## Configure auto-indexing policies

Once auto-indexing has been enabled, [create auto-indexing policies](configure_auto_indexing.md) to control the set of repositories and commits that are eligible for indexing.

## Tune the index scheduler

The frequency of index job scheduling can be tuned via the following environment variables read by `worker` service containers running the [`codeintel-auto-indexing`](../../../admin/workers.md#codeintel-auto-indexing) task.

**`PRECISE_CODE_INTEL_AUTO_INDEXING_TASK_INTERVAL`**: The frequency with which to run periodic codeintel auto-indexing tasks. Default is every 10 minutes.

**`PRECISE_CODE_INTEL_AUTO_INDEXING_REPOSITORY_PROCESS_DELAY`**: The minimum frequency that the same repository can be considered for auto-index scheduling. Default is every 24 hours.

**`PRECISE_CODE_INTEL_AUTO_INDEXING_REPOSITORY_BATCH_SIZE`**: The number of repositories to consider for auto-indexing scheduling at a time. Default is 100.

**`PRECISE_CODE_INTEL_AUTO_INDEX_MAXIMUM_REPOSITORIES_INSPECTED_PER_SECOND`**: The maximum number of repositories inspected for auto-indexing per second. Set to zero to disable limit. Default is 0.
