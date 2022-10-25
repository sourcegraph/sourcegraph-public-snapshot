# Deploying Sourcegraph executors

<aside class="beta">
<p>
<span class="badge badge-beta">Beta</span> This feature is in beta and might change in the future.
</p>

<p><b>We're very much looking for input and feedback on this feature.</b> You can either <a href="https://about.sourcegraph.com/contact">contact us directly</a>, <a href="https://github.com/sourcegraph/sourcegraph">file an issue</a>, or <a href="https://twitter.com/sourcegraph">tweet at us</a>.</p>
</aside>

[Executors](executors.md) provide a sandbox that can run resource-intensive or untrusted tasks on behalf of the Sourcegraph instance, such as:

- [Automatically indexing a repository for precise code navigation](../code_navigation/explanations/auto_indexing.md)
- [Running batch changes](../batch_changes/explanations/server_side.md)

> NOTE: Executors are available with no additional setup required on Sourcegraph Cloud.

## Requirements

Executors by default use KVM-based micro VMs powered by [Firecracker](https://github.com/firecracker-microvm/firecracker) in accordance with our [sandboxing model](executors.md#how-it-works) to isolate jobs from each other and the host.
This requires executors to be run on machines capable of running Linux KVM extensions. On the most popular cloud providers, this either means running executors on bare-metal machines (AWS) or machines capable of nested virtualization (GCP).

Optionally, executors can be run without using KVM-based isolation, which is less secure but might be easier to run on common machines.

## Configure Sourcegraph

Executors must be run separately from your Sourcegraph instance.

Since they must still be able to reach the Sourcegraph instance in order to dequeue and perform work, requests between the Sourcegraph instance and the executors are authenticated via a shared secret.

Before starting any executors, generate an arbitrary secret string (with at least 20 characters) and [set it as the `executors.accessToken` key in your Sourcegraph instance's site-config](config/site_config.md#view-and-edit-site-configuration).

## Executor installation

Once the shared secret is set in Sourcegraph, you can start setting up executors that can use that access token to talk to the Sourcegraph instance.

### Supported installation types

<div class="grid">
  <a class="btn-app btn" href="/admin/deploy_executors_terraform">
    <h3>Terraform</h3>
    <p>Simply launch executors on AWS or GCP using Sourcegraph-maintained modules and machine images.</p>
    <p>Supports auto scaling.</p>
  </a>
  <a class="btn-app btn" href="/admin/deploy_executors_binary">
    <h3>Install executor on your machine</h3>
    <p>Run executors on any linux amd64 machine.</p>
  </a>
</div>

## Confirm executors are working

If executor instances boot correctly and can authenticate with the Sourcegraph frontend, they will show up in the _Executors_ page under _Site Admin_ > _Maintenance_.

![Executor list in UI](https://storage.googleapis.com/sourcegraph-assets/docs/images/code-intelligence/sg-3.34/executor-ui-test.png)
