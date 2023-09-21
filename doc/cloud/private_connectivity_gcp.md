# Code hosts on GCP without public access

<p>Please contact Sourcegraph directly via <a href="https://about.sourcegraph.com/contact">prefered contact method</a> for more informations</p>
</aside>

As part of the [Enterprise tier](https://about.sourcegraph.com/pricing), Sourcegraph Cloud supports connecting to customer code hosts on GCP using [GCP Private Service Connect](https://cloud.google.com/vpc/docs/private-service-connect). It creates a secure connection between customer GCP project and Sourcegraph Cloud instance, so that access to a private code host never occurs over the public internet.

When a customer has private code hosts inside the GCP and needs to expose it for Sourcegraph Cloud instance, please reach out to your account manager to initiate the process.

## How it works

Sourcegraph supports connecting to private code hosts on GCP using GCP [Private Service Connect] (PSC). It is used to securely expose and connect services across the project boundary within GCP.

The customer is the Service Producer (the "producer"), and the Sourcegraph Cloud instance is the Service Consumer (the "consumer"). PSC can expose an internal regional load balancer for the private code host to the consumer. The consumer can then connect to the private code host over PSC transparently on their Sourcegraph Cloud instance.

[link](https://link.excalidraw.com/readonly/Xiz9LWNPCa3DERBJUiZI)

<iframe src="https://link.excalidraw.com/readonly/Xiz9LWNPCa3DERBJUiZI" width="100%" height="100%" style="border: none;"></iframe>

## Limitation

Cross-region connectivity is not supported by Google Cloud for [Private Service Connect]. The Sourcegraph Cloud instance and the customer code host must be in the same region, learn more from our [supported regions](../index.md#multiple-region-availability).

## Steps

### Initiate the process

Customer should reach out to their account manager to initiate the process. The account manager will work with the customer to collect the required information and initiate the process, including but not limited to:

- The DNS name of the private code host, e.g., `gitlab.internal.company.net`.
- The region of the private code host on GCP, e.g., `us-central1`.
- The type of the TLS certificate used by the private code host, one of self-signed by internal private CA, or issued by a public CA.
- The location of where the TLS connection is terminated, one of the load balancer, or the private code host node.

Finally, Sourcegraph will provide the following:

- A reference architecture in Terraform to demonstrate the setup on customer end.
- The GCP Project ID of the Sourcegraph Cloud instance.

### Create Private Serivce Connect connection

Customer should publish their services using PSC by follow [GCP documentation](https://cloud.google.com/vpc/docs/configure-private-service-connect-producer). The customer needs to [permit connection](https://cloud.google.com/vpc/docs/manage-private-service-connect-services#access) from the provided GCP Project ID earlier. The customer needs to provide the [Service Attachment] URI to Sourcegraph. The Service Attachment URI is in the format of `projects/:id/regions/:region/serviceAttachments/:name`. 

Upon receiving the Service Attachment URI, Sourcegraph will create a connection to the customer service using PSC, and Sourcegraph will follow up with the customer to confirm the connection is established.

### Create the code host connection

Once the connection is established, the customer can create the [code host connection](../../admin/external_service/index.md) on their Sourcegraph Cloud instance.

## FAQ

### How can I restrict access to my private code host connection?

The customer has full control over the access to the [Service Attachment] by configuring the [accept and reject lists](https://cloud.google.com/vpc/docs/private-service-connect-security#consumer-lists). Sourcegraph will provide the GCP Project ID to be added to the accept list. 

Additionally, you may terminate the connection at any point. You can do so by running:

```sh
gcloud compute service-attachments delete [SERVICE_ATTACHMENT_NAME] --region=[REGION] --project=[YOUR_PROJECT_ID]
```

Learn more from documentation of [gcloud compute service-attachments delete](https://cloud.google.com/sdk/gcloud/reference/compute/service-attachments/delete) and [management of published services](https://cloud.google.com/vpc/docs/manage-private-service-connect-services).

### How secure is the connection?

All traffic between the producer and consumer is encrypted in transit. You may learn more from Google's whitepaper about [encryption in transit] and [Sourcegraph's security practices].

[private service connect]: https://cloud.google.com/vpc/docs/private-service-connect
[service attachment]: https://cloud.google.com/vpc/docs/private-service-connect#service-attachments
[encryption in transit]: https://cloud.google.com/docs/security/encryption-in-transit
[Sourcegraph's security practices]: https://about.sourcegraph.com/security
