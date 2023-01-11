# Enterprise Getting Started Guide

If you're deploying a new Enterprise instance, this page covers our most frequently referenced pieces of documentation. Admins will be interested in the documentation for their specific deployment method; users will want to check out info on our search syntax, operators, batch changes, and browser extension.

## Admin articles

### General
- [Deployment overview](../admin/deploy/index.md)
- [Resource estimator](../admin/deploy/resource_estimator.md)
- [SAML config](../admin/auth/saml/index.md)
- [Configuring authorization and authentication](../admin/config/authorization_and_authentication.md) - We recommend starting here for access and permissions configuration before moving on to the more specific pages on these topics.
- [Built-in password authentication](../admin/auth/index.md#builtin-password-authentication)
- [GitHub authentication](../admin/auth/index.md#github)
- [GitLab authentication](../admin/auth/index.md#gitlab)
- [OpenID connect](../admin/auth/index.md#openid-connect)
- [HTTP authentication proxy](../admin/auth/index.md#http-authentication-proxies)
- [GitLab integration](../integration/gitlab.md)
- [GitHub integration](../integration/github.md)
- [All code host integrations (not GitLab or GitHub)](../integration/index.md#integrations)
- [Full guide to site config options](../admin/config/site_config.md#auth-sessionExpiry)
- [Changelog](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/CHANGELOG.md) to track releases and updates

### Docker-compose
- [Basic installation guide](../admin/deploy/docker-compose/index.md)
- [AWS installation](../admin/deploy/docker-compose/aws.md)
- [Digital Ocean installation](../admin/deploy/docker-compose/digitalocean.md)
- [Google Cloud installlation](../admin/deploy/docker-compose/google_cloud.md)

### Kubernetes admin
- [All Kubernetes with Helm guidance](../admin/deploy/kubernetes/helm.md)
- [Amazon EKS](../admin/deploy/kubernetes/helm.md#configure-sourcegraph-on-elastic-kubernetes-service-eks)
- [Google GKE](../admin/deploy/kubernetes/helm.md#configure-sourcegraph-on-google-kubernetes-engine-gke)
- [Azure](../admin/deploy/kubernetes/helm.md#configure-sourcegraph-on-azure-managed-kubernetes-service-aks)
- [Configure Sourcegraph on other Cloud providers or on-prem](../admin/deploy/kubernetes/helm.md#configure-sourcegraph-on-other-cloud-providers-or-on-prem)

## User articles
- [Search syntax](../code_search/reference/queries.md)
- [Search operators](../code_search/reference/queries.md#keywords-all-searches)
- [Example batch changes](../batch_changes/tutorials/index.md)
- [Sourcegraph browser extension](../integration/browser_extension.md)
