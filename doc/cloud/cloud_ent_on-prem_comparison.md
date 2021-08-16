# Sourcegraph Cloud 
Sourcegraph cloud is a newly released Software-as-a-Service version of Sourcegraph. Fundamentally, both versions utilize the same search developers around the world rely on. Our team is actively working on building cloud to feature parity with our on-prem solution. 

## Differences between cloud and on-prem 
Please see the below breakdown of feature support for Sourcegraph cloud vs Sourcegraph on-prem. 

| Feature                             | Cloud                | On-Prem          |
| ----------------------------------- | -------------------- | ---------------- |
| Search your public and private code | Yes                  | Yes              |
| Code Intel                          | Yes                  | Yes              |
| Code Insights                       | Yes *                | Yes              |
| Extensions                          | Yes                  | Yes              |
| Custom Search Contexts              | Yes                  | Yes              |
| Batch Changes                       | No                   | Yes (Enterprise) |
| SSO                                 | Via Github/Gitlab ** | Enterprise Only  |
| User Management                     | Via Github/Gitlab ** | Yes              |

\* Code insights on cloud works with up to 50 repositories today. We're actively working to improve this limitation. 
\** Sourcegraph cloud today does not support explicit SSO or User Management, though Sourcegraph.com uses GitHub / GitLab SSO as a proxy to respecting SSO and User Management. For example, imagine a customer uses Okta SSO. This company would require their developers to use Okta SSO to sign in to GitHub, and then use GitHub to sign in to Sourcegraph. If the user leaves the company and SSO perms are removed, that user would no longer see private repositories on GitHub or on Sourcegraph. 
