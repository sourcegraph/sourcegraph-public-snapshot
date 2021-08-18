# Sourcegraph cloud (Public Beta)

Sourcegraph cloud is a newly released cloud-based version of Sourcegraph. Both Sourcegraph cloud and self-hosted Sourcegraph instances power the same search experience relied on by developers around the world. The Sourcegraph team is working on bringing Sourcegraph cloud to feature parity with our self-hosted Sourcegraph solution.

## Differences between Sourcegraph cloud and self-hosted Sourcegraph instances

| Feature                             | Cloud                | Self-hosted      |
| ----------------------------------- | -------------------- | ---------------- |
| Search your public and private code | Yes                  | Yes              |
| Code Intel                          | Yes                  | Yes              |
| Code Monitoring                     | Yes                  | Yes              |
| Code Insights                       | Yes *                | Yes              |
| Extensions                          | Yes                  | Yes              |
| Custom Search Contexts              | Yes                  | Yes              |
| Batch Changes                       | No                   | Enterprise       |
| SSO                                 | Via Github/Gitlab ** | Enterprise       |
| User Management                     | Via Github/Gitlab ** | Yes              |

\* Code insights on Sourcegraph cloud is performant across at most 50-75 repositories today. We're working to improve performance across more than 75 repositories.  
\** Sourcegraph cloud provides SSO through GitHub.com / GitLab.com as a proxy. For example, a company using Okta SSO can require team members to sign in to GitHub with Okta SSO, and then use GitHub to sign in to Sourcegraph cloud. [Repository permissions on code hosts are always respected](../code_search/explanations/code_visibility_on_sourcegraph_cloud.md).
