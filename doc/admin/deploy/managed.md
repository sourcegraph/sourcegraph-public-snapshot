# Managed Sourcegraph instances

Managed instances are provisioned and managed by the Sourcegraph team, making them a simple choice for customers to deploy Sourcegraph without having to worry about managing it.

Sourcegraph provisions your instance in a completely isolated and secure cloud infrastructure. It will be restricted to only your organization through your enterprise VPN and/or SSO provider of choice.

## Costs

Please [contact us](https://about.sourcegraph.com/contact/sales) to discuss [pricing](https://about.sourcegraph.com/pricing) and to start a trial.

## Scaling

Managed instances can be used by most enterprises of any size. Please [contact us](https://about.sourcegraph.com/contact/sales) to discuss options.

# Service

As part of this service you will receive a number of benefits from our team, including:

## Initial setup, configuration, and cost estimation

- Advising if managed instances are right for your organization.
- Initial resource estimations based on your organization & code size.
- Putting forward a transparent deployment & cost estimate plan.
- Your own `example.sourcegraph.com` domain with fully managed [DNS & HTTPS](../http_https_configuration.md).
- Hardware provisioning, software installation, and kernel configuration done for you.
- Direct assistance in:
  - [Adding repositories from all of your code hosts to Sourcegraph](../external_service/index.md)
  - [Integrating your single sign-on provider with Sourcegraph](../auth/index.md)
  - [Configuring Sourcegraph](../config/index.md)

## Access restrictions

- Granting your team application-level admin access to the instance.
- Configuring any IP-restrictions (e.g. VPN) and/or SSO restrictions to the instance.

## Monthly upgrades and maintenance

- Automatic monthly [upgrades](../updates/index.md) and maintenance.
- Regular reassessment of resource utilization based on your organization's unique usage to determine if costs can be reduced without impact to service. Additionally, you will automatically benefit from any committed use cloud provider discounts we receive.

## Health monitoring, support, and SLAs

- Instance performance and health [monitored](../observability/index.md) by our team's on-call engineers.
- [Responding to support requests and maintaining SLAs](https://handbook.sourcegraph.com/support#for-customers-with-managed-instances)

## Backup and Restore

- Snapshots are taken every week and kept for 90 days.

## Training, feedback, and engagement

As with any Sourcegraph enterprise customer, you will also receive support from us with:

- [Installing code host and code review integrations](../../integration/index.md)
- Monitoring and aggregating user feedback
- Understanding usage statistics of your deployment
- Internal rollout programs including:
  - Holding company-wide or team-by-team training sessions ([contact us](https://about.sourcegraph.com/contact/sales) for details)
  - Helping the maintainers of your internal engineer onboarding add a session on Sourcegraph
  - Holding ongoing brown bag lunches to introduce new feature releases
  - Advice and templates on how to introduce Sourcegraph to your engineering organization

# Requirements

## Business

- A dedicated project manager point of contact for the rollout process
- A mutual non-disclosure agreement, and any additional approvals or special status required to allow Sourcegraph to manage infrastructure access tokens (listed below)
- Acceptance of our [Terms of Service for private instances](https://about.sourcegraph.com/terms-private) or an enterprise contract

## Technical

- A dedicated technical point of contact for the installation process.
- [Tokens with read access to your code hosts](../external_service/index.md) (we will direct you on how to enter them)
- [Keys, access tokens, or any other setup required to integrate your SSO (single sign-on) provider with Sourcegraph](../auth/index.md), as well as support from a member of your team with administrator access to your SSO provider to help set up and test the integration.
- If you desire VPN/IP-restricted access, we will need to know the IP/CIDR source ranges of your enterprise VPN to allow access to the instance.

# Security

Your managed instance will be accessible over HTTPS/TLS, provide storage volumes that are encrypted at rest, and would have access restricted to only your team through your enterprise VPN and/or internal [SSO (single sign-on provider)](../auth/index.md) of choice.

If you decide for your manged instance to be public, we will provide security capabilities from Cloudflare such as WAF and rate-limiting. We will also provide a firewall to protect your instance from malicious traffic.

It will be hosted in completely isolated Google Cloud infrastructure, with minimal access even within the Sourcegraph team, both for security and billing purposes. See our [employee handbook](https://handbook.sourcegraph.com/engineering/enablement/delivery/managed#technical-details) to learn more about the cloud architecture we use. Both your team and limited Sourcegraph personnel will have application-level administrator access to the instance.

Only essential Sourcegraph personnel will have access to the instance, server, code, and any other sensitive materials, such as tokens or keys. The employees or contractors with access would be bound by the same terms as Sourcegraph itself. Learn more in our [network security policies for Sourcegraph Cloud](https://about.sourcegraph.com/security) or [contact us](https://about.sourcegraph.com/contact/sales) with any questions/concerns.

# Accommodating special requirements

We may be able to support special requests (network access policies, infrastructure requirements, custom version control systems, etc.) with additional time, support, and fees. [Contact us](https://about.sourcegraph.com/contact/sales) to discuss any special requirements you may have.
