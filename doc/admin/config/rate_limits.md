# Configuring rate limits on Sourcegraph

Sourcegraph has various mechanisms in place to ensure that communication with external services happen in a controlled and reliable manner.

## External rate limits

Many code hosts have their own rate limits configured, and will provide feedback on requests when these rate limits are triggered. When Sourcegraph encounters a rate limit from the code host, the appropriate amount of time, as specified by the code host, will be waited before a request is retried.

Sourcegraph monitors external rate limits for the following code hosts:

1. [GitHub](../external_service/github.md#github-com-rate-limits)
2. [GitLab](../external_service/gitlab.md#gitlab-com-rate-limits)
3. [Azure DevOps](../external_service/azuredevops.md#azure-devops-rate-limits)
4. [Bitbucket Cloud](../external_service/bitbucket_cloud.md#bitbucket-cloud-rate-limits)

## Internal rate limits

Sourcegraph can also be configured with internal rate limits. These rate limits control how often Sourcegraph interacts with external services. Internal rate limits differ from external rate limits in that they are self-imposed, and are not enforced by the code host, but instead by Sourcegraph itself.

Internal rate limits are useful in scenarios where the load on a code host needs to be controlled, for example to lighten the load on a self-hosted code host instance.

Internal rate limits can be configured for each code host connection:

1. [GitHub](../external_service/github.md#rateLimit)
2. [GitLab](../external_service/gitlab.md#rateLimit)
3. [Bitbucket Cloud](../external_service/bitbucket_cloud#rateLimit)
4. [Bitbucket Server](../external_service/bitbucket_server#rateLimit)

These internal rate limits will also be applied when syncing user permissions from the code host.

Note that, if there are multiple code host connections to the same code host, then each connection will have separate internal rate limits.
