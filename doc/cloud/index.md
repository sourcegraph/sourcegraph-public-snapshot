# Sourcegraph Cloud

Sourcegraph Cloud is a single-tenant cloud solution. Cloud instances are private, dedicated instances provisioned and managed by Sourcegraph. Sourcegraph Cloud was formerly known as managed instances.

Sourcegraph provisions each instance in an isolated and secure cloud environment. Access is restricted to only your organization through your SSO provider of choice. Enterprise VPN is available upon request.

## Start a Sourcegraph Cloud trial

<div class="grid">
  <!-- Sourcegraph Cloud -->
  <a class="btn-app btn" href="http://signup.sourcegraph.com">
			<img alt="sourcegraph-logo" src="https://handbook.sourcegraph.com/departments/engineering/design/brand_guidelines/logo/versions/Sourcegraph_Logomark_Color.svg"/>
			<h3>Sourcegraph Cloud</h3>
		  <p>Sign up for a trial</p>
  </a>
</div>

Use the button above to sign up for a free 30-day trial of Sourcegraph Cloud. Please [contact us](https://about.sourcegraph.com/contact/sales) if you have specific VPN requirements or you require a large deployment with >500 users, >1,000 repos, or monorepos >5 GB.

## Trial limitations

We currently have a limited capacity of single-tenant cloud instances and are prioritizing organizations with more than 100 developers. When you request a trial, you will receive an email indicating the status of your request.

If your organization has fewer than 100 developers, we recommend trying [Sourcegraph self-hosted](https://docs.sourcegraph.com/#self-hosted).

If you're eligible for a cloud instance, you will receive a link to the instance URL once it's provisioned. This normally takes less than one hour during business hours. From there, we recommend using the [onboarding checklist](../getting-started/cloud-instance.md) to set up your instance quickly.

Trials last 30 days. When the end of the trial approaches, Sourcegraph's Customer Support team will check in with you to either help you set up a Cloud subscription or end your trial.

# Cloud subscription

As part of this service you will receive a number of benefits from our team, including:

## Initial setup, configuration, and cost estimation

- Advising if managed instances are right for your organization.
- Initial resource estimations based on your organization & code size.
- Putting forward a transparent deployment & cost estimate plan.
- Your own `example.sourcegraph.com` domain with fully managed [DNS & HTTPS](../admin/http_https_configuration.md).
- Hardware provisioning, software installation, and kernel configuration done for you.
- Direct assistance in:
  - [Adding repositories from all of your code hosts to Sourcegraph](../admin/external_service/index.md)
  - [Integrating your single sign-on provider with Sourcegraph](../admin/auth/index.md)
  - [Configuring Sourcegraph](../admin/config/index.md)

## Access restrictions

- Granting your team application-level admin access to the instance.
- Configuring any IP-restrictions (e.g. VPN) and/or SSO restrictions to the instance.

## Monthly upgrades and maintenance

- Automatic monthly [upgrades](../admin/updates/index.md) and maintenance.
- Regular reassessment of resource utilization based on your organization's unique usage to determine if costs can be reduced without impact to service. Additionally, you will automatically benefit from any committed use cloud provider discounts we receive.

## Health monitoring, support, and SLAs

- Instance performance and health [monitored](../admin/observability/index.md) by our team's on-call engineers.
- [Support and SLAs](https://handbook.sourcegraph.com/support#for-customers-with-managed-instances).

## Backup and restore

<span class="badge badge-note">SOC2/CI-79</span>

Backup and restore capability is provided via automated snapshots.

- Frequency: Snapshots are produced daily.
- Retention period: Snapshots are kept for 90 days.

## Training, feedback, and engagement

As with any Sourcegraph enterprise customer, you will also receive support from us with:

- [Installing code host and code review integrations](../integration/index.md)
- Monitoring and aggregating user feedback
- Understanding usage statistics of your deployment
- Internal rollout programs including:
  - Holding company-wide or team-by-team training sessions ([contact us](https://about.sourcegraph.com/contact/sales) for details)
  - Helping the maintainers of your internal engineer onboarding add a session on Sourcegraph
  - Holding ongoing brown bag lunches to introduce new feature releases
  - Advice and templates on how to introduce Sourcegraph to your engineering organization

## Requirements

### Business

- A dedicated project manager who serves as the point of contact for the rollout process.
- A mutual non-disclosure agreement and any additional approvals or special status required to allow Sourcegraph to manage infrastructure access tokens (listed below).
- Acceptance of our [Terms of Service for private instances](https://about.sourcegraph.com/terms-private) or an enterprise contract.

### Technical

- A dedicated technical point of contact for the installation process.
- [Tokens with read access to your code hosts](../admin/external_service/index.md) (we will direct you on how to enter them).
- [Keys, access tokens, or any other setup required to integrate your SSO (single sign-on) provider with Sourcegraph](../admin/auth/index.md), as well as support from a member of your team with administrator access to your SSO provider to help set up and test the integration.
- If you desire VPN/IP-restricted access, we will need to know the IP/CIDR source ranges of your enterprise VPN to allow access to the instance.

## Security

Your managed instance will be accessible over HTTPS/TLS, provide storage volumes that are encrypted at rest, and have access restricted to only your team through your enterprise VPN and/or internal [SSO (single sign-on provider)](../admin/auth/index.md) of choice.

If you would like your managed instance to be public, we will provide security capabilities from Cloudflare such as WAF and rate-limiting. We will also provide a firewall to protect your instance from malicious traffic.

Your instance will be hosted in isolated Google Cloud infrastructure. See our [employee handbook](https://handbook.sourcegraph.com/departments/cloud/technical-docs/) to learn more about the cloud architecture we use. Both your team and limited Sourcegraph personnel will have application-level administrator access to the instance.

Only essential Sourcegraph personnel will have access to the instance, server, code, and any other sensitive materials, such as tokens or keys. The employees or contractors with access are bound by the same terms as Sourcegraph itself. Learn more in our [network security policies for Sourcegraph Cloud](https://about.sourcegraph.com/security) or [contact us](https://about.sourcegraph.com/contact/sales) with any questions or concerns.

## Accommodating special requirements

We may be able to support special requests (network access policies, infrastructure requirements, custom version control systems, etc.) with additional time, support, and fees. [Contact us](https://about.sourcegraph.com/contact/sales) to discuss any special requirements you may have.
