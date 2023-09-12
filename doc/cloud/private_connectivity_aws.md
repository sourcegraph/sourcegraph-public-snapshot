# Code hosts on AWS without public access

<aside class="experimental">
<p>
<span class="badge badge-experimental">Experimental</span> This feature is experimental
</p>

<p>Please contact Sourcegraph directly via <a href="https://about.sourcegraph.com/contact">prefered contact method</a> for more informations</p>
</aside>

As part of the [Enterprise tier](https://about.sourcegraph.com/pricing), Sourcegraph Cloud offers customers that have code hosts without public access deployed on AWS a [highly available site-to-site VPN solution](https://cloud.google.com/network-connectivity/docs/vpn/tutorials/create-ha-vpn-connections-google-cloud-aws) with [AWS Private Link](https://docs.aws.amazon.com/vpc/latest/privatelink/what-is-privatelink.html) inside AWS's network, so that access to a private code host never occurs over the public internet.

Solution architecture:
<img src="https://sourcegraphstatic.com/private-code-host-solution-vpn-aws-private-link.png" class="screenshot">

Advantages of the site-to-site GCP to AWS VPN include:
- encrypted connection between Sourcegraph Cloud and customer code host
- multiple tunnels to provide high availability between Cloud
instance and customer code host

Advantages of AWS Private Link include:
- connectivity to customer VPC is only available inside AWS network
-  ability to select AWS Principal (AWS Account or more granular) that can connect to customer code host
- allows customer to control incoming connections
- supports private DNS

When a customer has private code hosts inside the AWS VPC and needs to expose it for Sourcegraph managed AWS VPC, customers can follow [AWS Documentation](https://docs.aws.amazon.com/vpc/latest/privatelink/create-endpoint-service.html)
