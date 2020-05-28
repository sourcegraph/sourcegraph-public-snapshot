# Set up a managed instance

> NOTE: Using Sourcegraph as a managed service is a [paid feature](https://about.sourcegraph.com/pricing). [Contact us](https://about.sourcegraph.com/contact/sales) to discuss requirements and start a trial.

Sourcegraph can be installed and managed by the Sourcegraph team. 

## Hosting

Managed instances are hosted on Sourcegraph's cloud infrastructure, and are accessible at a https://example.sourcegraph.com subdomain. 

Need to use your own infrastructure? Please [contact us](https://about.sourcegraph.com/contact/sales) to discuss options and pricing.

## Services

With a managed instance, Sourcegraph's Distribution Engineering team manages the full lifecycle of your Sourcegraph deployment. This includes:

- Creating and providing a transparent deployment plan with cost estimates
- [Selecting the appropriate deployment model (Docker, Docker Compose, Kubernetes), depending on code and user scale](../install/index.md)
- Installing and configuring Sourcegraph
  - [Adding repositories from all of your code hosts to Sourcegraph](../external_service/index.md)
  - [Integrating your single sign-on provider with Sourcegraph](../auth/index.md)
  - [Providing a valid certificate and setting up HTTPS/TLS access](../http_https_configuration.md)
  - [Configuring Sourcegraph](../config/index.md)
- [Supporting your team in installing code host and code review integrations](../../integration/index.md)
- [Performing monthly upgrades and maintenance](../updates.md)
  - Running smoke tests to validate functionality after upgrades
- [Monitoring performance](../observability/index.md)
- Monitoring and aggregating user feedback and usage statistics
- [Responding to support requests and maintaining SLAs](https://about.sourcegraph.com/handbook/support#for-customers-with-managed-instances)

Additionally, Sourcegraph would support your internal rollout programs, including:

- Providing templates for introduction emails to the engineering organization
- Holding company-wide or team-by-team training sessions ([contact us](https://about.sourcegraph.com/contact/sales) for details)
- Helping the maintainers of your internal engineer onboarding add a session on Sourcegraph
- Holding ongoing brown bag lunches to introduce new feature releases

## Requirements

**Business**

- A dedicated project manager point of contact for the rollout process
- A mutual non-disclosure agreement, and any additional approvals or special status required to allow Sourcegraph to manage infrastructure access tokens (listed below)
- Acceptance of our [Terms of Service for private instances](https://about.sourcegraph.com/terms-private) or an enterprise contract

**Technical**

- A dedicated technical point of contact for the installation process
- [Tokens with read access to your code hosts](../external_service/index.md)
- [Keys, access tokens, or any other setup required to integrate your single sign-on provider with Sourcegraph](../auth/index.md), as well as support from a member of your team with administrator access to your single sign-on provider to help set up and test the integration

## Security

Your managed instance would be accessible over HTTPS/TLS, provide storage volumes that are encrypted at rest, and would have access restricted using your internal [single sign-on provider](../auth/index.md) of choice.

Only essential Sourcegraph teammates would have access to your instance frontend, server, code, and any other sensitive materials, such as tokens or keys. The employees or contractors with access would be bound by the same terms as Sourcegraph itself.

Learn more in our [network security policies for Sourcegraph Cloud](https://about.sourcegraph.com/security).

## Other requests

Special requests, including non-standard network access policies or integration with your VPN, hosting on custom infrastructure or with non-standard software, custom version control systems, and more, may require additional time, support, and fees. [Contact us to discuss your needs](https://about.sourcegraph.com/contact/sales).
