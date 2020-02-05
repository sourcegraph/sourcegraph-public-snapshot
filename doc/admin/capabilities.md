# Site administrator capabilities

Site administrators have full administrative access to the Sourcegraph instance. In many cases, they also control the deployment environment. Therefore, it is reasonable to have extra capabilities for site administrators.

## Access to all repositories

Site administrators are able to access all repositories added to the Sourcegraph instance, and manage settings of individual repository, including repository indexing status, statistics, mirroring information and code intelligence assets.

## Access to all GraphQL APIs

Site administrators are able to access all GraphQL APIs of the Sourcegraph instance. There are many GraphQL APIs require the client to have the token from a site administrator.

## Impersonate regular users

Site administrators are able to impersonate regular users in two ways:
1. View and update settings of individual user.
2. Initiate GraphQL API calls on behave of a regular user.

## Receive site alerts

Site administrators are able to see site alerts in the form of top banner about configuration problems of the Sourcegraph instance.
