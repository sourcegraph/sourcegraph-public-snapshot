# Private artifact registry in customer data center

<p>Please contact Sourcegraph directly via <a href="https://about.sourcegraph.com/contact">prefered contact method</a> for more informations</p>
</aside>

As part of the [Enterprise tier](https://about.sourcegraph.com/pricing), Sourcegraph Cloud supports connecting customer private artifact registries from customer data center via public load balancer on customer side.
> For artifact private registries in AWS or GCP, please refer to [other deployment methods](./index.md#private-artifact-registries-support)

## How it works

Sourcegraph Cloud is a managed service hosted on GCP. Customer will expose private artifact registry via load balancer with IP allowlist for 2 static IPs provided by Sourcegraph. Sourcegraph will then be able to access the private registry over HTTPS through the load balancer from the GCP project hosting Sourcegraph Cloud. Sourcegraph recommends to setup pass-through TCP load balancer to avoid adding private registry domain certificate to exposed load balancer.

[link](https://link.excalidraw.com/readonly/gc6P8SDOEMCcrIl9cl64)

<iframe src="https://link.excalidraw.com/readonly/gc6P8SDOEMCcrIl9cl64" width="100%" height="100%" style="border: none;"></iframe>

## Steps

### Initiate the process

Customer should reach out to their account manager to initiate the process. The account manager will work with the customer to collect the required information and initiate the process, including but not limited to:

- The DNS name of the private artifact registry, e.g., `artifactory.internal.company.net`.
- The public DNS name of the load balancer exposing private artifact registry, e.g., `artifactory.public.company.net`.

Sourcegraph will provide 2 static IPs for customer to allowlist ingress traffic for load balancer.

## FAQ

### Why pass-through TCP load balancer?

With pass-through TCP load balancer, the load balancer acts as a simple TCP proxy to forward traffic to the backend private registry without terminating TLS. This avoids the need to install private registry server certificate on the load balancer, reducing certificate management overhead.

### Can I use my internal dns name for artifact registry?

Yes, customer can expose their private registry with internal DNS name. Sourcegraph will provision dns-proxy, which translates customer private domain to public customer load balancer domain.
No changes in customer configuration are required.

### What are the next steps when connectivity is working?

Only if private artifact registry is protected by authentication, the customer will need to:
- create executor secrets containing credentials for Sourcegraph to access the private artifact registry - [how to configure executor secrets](../admin/executors/executor_secrets.md#executor-secrets)
- update auto-indexing inference configuration to create additional files from executor secrets for given programing language - [how to configure auto-indexing](../code_navigation/references/inference_configuration.md)
