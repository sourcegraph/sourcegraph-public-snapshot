# Rate limiting

When interfacing with code hosts, Sourcegraph continuously makes requests to the code hosts in order to keep the system as up to date as possible.
This includes things like checking if there are any new repositories that need to be cloned, checking for updates to existing repositories,
confirming user access to repositories, etc.

Ideally these tasks would happen as frequently as possible. However, resources are finite, and as a result we have to decide on reasonable
compromises. We do this by imposing rate limits: a pre-defined maximum number of requests that will be allowed to happen in a given time-frame.
Many code hosts have their own rate limits, and Sourcegraph respects these rate limits.

## Configuring rate limits

Note: Regardless of the rate limits configured, Sourcegraph will always respect a code host's rate limits if the code host provides rate limit headers in their responses.

Rate limits can be configured by visiting the Site Admin page, and selecting "Rate limits" under the "Configuration" header in the side bar.
Here, rate limits can be configured on various levels of interaction, namely

1. Host URL
1. Code host connection
1. Authentication provider
1. User

All rate limited requests are made to a specific host URL. If a rate limit is set for a host URL, then the total number of all requests made to this URL cannot exceed the rate limit.

The systems in Sourcegraph that make requests to this URL are the code host connection, used to discover and fetch repositories, as well as determining user access in some cases.
And user authentication, which users use to authenticate themselves via a code host, and to determine which repositories each user can access.

## Host URL

If a host URL rate limit is set, the total number of requests to this URL will not be able to exceed the rate limit.
For example, given the following Sourcegraph setup:
1 GitLab.com code host connection
2 Users: Alice and Sarah

If a host URL rate limit of 50 requests per hour is configured, and it takes 10 requests to check for updates to all configured repositories on the GitLab code host connection, and 20 requests each to see what repositories Alice and Sarah have access to,
then we will be able to check for repository updates and sync user permissions exactly once per hour. If more repositories get added, or if another user gets added, we will no longer be able to do a complete sync
in an hour, even if the sync could theoretically complete in 30 seconds.

If you determine it is necessary to rate limit requests to a specific URL, but are unsure of what the limit needs to be while still ensuring a functional system, you can confirm the current request throughput to a specific URL in the Site Admin area.

## Code host connection

Rate limits can also be applied to specific code host connections. The rate limit for each code host connection can be configured in the code host connection settings.

If there are multiple code host connections that point to the same host URL, the combined request throughput of all code host connections will be subject to the host URL rate limit.

## Authentication provider

If users need to connect external accounts in order to verify their access to private repositories, an authentication provider will need to be configured.
Rate limits can be set for all accounts that are connected through a specific authentication provider. This means that the total request throughput of all accounts connected through this specific authentication provider cannot exceed the specified limit.

## User accounts

Rate limits can also be set on the user account level. This is also configured in the authentication provider settings, but these are the limits for each individual user token. For example, user tokens can have a rate limit of 5000 requests per hour, and an authentication provider can have a rate limit of 50 000 requests per hour.
Certain code hosts also enforce rate limits at this level. For example, GitHub enforces a rate limit of 5 000 requests per hour per user. Specifying a higher user limit cannot overcome the code host's limit, and Sourcegraph will adhere to the code host's limit where possible.
