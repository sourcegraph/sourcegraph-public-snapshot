# Private artifact registry on AWS without public access

<p>Please contact Sourcegraph directly via <a href="https://about.sourcegraph.com/contact">prefered contact method</a> for more informations</p>
</aside>

As part of the [Enterprise tier](https://about.sourcegraph.com/pricing), Sourcegraph Cloud supports connecting customer artifact registries on AWS using [AWS Private Link] and managed [site-to-site VPN] solution between GCP and AWS, so that access to a private artifact registry is secure and without the need to expose artifact registries to the public internet.

## How it works

Sourcegraph Cloud is a managed service hosted on GCP. Sourcegraph creates a secure connection between customer [AWS Virtual Private Cloud] (AWS VPC) and a Sourcegraph-managed AWS account using [AWS Private Link]. Then, Sourcegraph maintains a secure connection between the Sourcegraph-managed AWS VPC and GCP Project via a managed highly available [site-to-site VPN] solution.
> Note: if customer is already using AWS Private Link for code host, they will reuse the existing site to site VPN connection with Sourcergaph Cloud VPC.

[link](https://link.excalidraw.com/readonly/WvySMq7S2O4mVgIEvdFu)

<iframe src="https://link.excalidraw.com/readonly/WvySMq7S2O4mVgIEvdFu" width="100%" height="100%" style="border: none;"></iframe>

## Steps

### Initiate the process

Customer should reach out to their account manager to initiate the process. The account manager will work with the customer to collect the required information and initiate the process, including but not limited to:

- The DNS name of the private artifact registry, e.g., `artifactory.internal.company.net`.
- The region of the private artifact registry on AWS, e.g., `us-east-1`.
- The type of the TLS certificate used by the private artifact registry, one of self-signed by internal private CA, or issued by a public CA.

### Create the VPC Endpoint Service

When a customer has private artifact registry inside the AWS VPC and needs to expose it for Sourcegraph managed AWS VPC, customers can follow [AWS Documentation](https://docs.aws.amazon.com/vpc/latest/privatelink/create-endpoint-service.html). An example can be found from our [handbook](https://handbook.sourcegraph.com/departments/cloud/technical-docs/private-code-hosts/#aws-private-link-playbook-for-customer).

Sourcegraph will provide the Sourcegraph-managed AWS account ARN that needs to be allowlist in your VPC endpoint service, e.g., `arn:aws:iam::$accountId:root`.

The customer needs to share the following details with Sourcegraph:

- VPC endpoint serivce name in the format of `com.amazonaws.vpce.<REGION>.<VPC_ENDPOINT_SERVICE_ID>`.

Upon receiving the detail, Sourcegraph will create a connection to the customer artifact registry, and Sourcegraph will follow up with the customer to confirm the connection is established. Customer might then verify that auto-indexing is working with private registry dependecies.

## FAQ

### Why AWS Private Link?

Advantages of AWS Private Link include:

- connectivity to customer VPC is only available inside AWS network
- ability to select AWS Principal (AWS Account or more granular) that can connect to customer artifact registry
- allows customer to control incoming connections

### Why site-to-site VPN connection between GCP to AWS?

Advantages of the site-to-site GCP to AWS VPN include:

- encrypted connection between Sourcegraph Cloud and customer artifact registry
- multiple tunnels to provide high availability between Cloud instance and customer artifact registry

###  How can I restrict access to my private artifact registry?

The customer has full control over the exposed service and they can may terminate the connection at any point.

[AWS Virtual Private Cloud]: https://docs.aws.amazon.com/vpc/latest/userguide/what-is-amazon-vpc.html
[AWS Private Link]: https://docs.aws.amazon.com/vpc/latest/privatelink/what-is-privatelink.html
[site-to-site VPN]: https://cloud.google.com/network-connectivity/docs/vpn/tutorials/create-ha-vpn-connections-google-cloud-aws

### Can I use my internal dns name for artifact registry?

Yes, customer can expose their private registry with internal DNS name via AWS Private Link. Sourcegraph will provision dns-proxy, which translates customer private domain to AWS Private Link domain.
No changes in customer configuration are required.

### What are the next steps when connectivity is working?

Only if private artifact registry is protected by authentication, the customer will need to:
- create executor secrets containing credentials for Sourcegraph to access the private artifact registry - [how to configure executor secrets](../admin/executors/executor_secrets.md#executor-secrets)
- update auto-indexing inference configuration to create additional files from executor secrets for given programing language - [how to configure auto-indexing](../code_navigation/references/inference_configuration.md)
