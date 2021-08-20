# Sourcegraph cloud (Public Beta)

Sourcegraph cloud is a newly released cloud-based version of Sourcegraph. Both Sourcegraph cloud and self-hosted Sourcegraph instances power the same search experience relied on by developers around the world. The Sourcegraph team is working on bringing Sourcegraph cloud to feature parity with our self-hosted Sourcegraph solution.

## Differences between Sourcegraph cloud and self-hosted Sourcegraph instances

Self-hosted Sourcegraph ensures no code leaves the Sourcegraph environment. It also provides user management, sharing of private code, and enterprise only features such as batch changes. This is the primary solution our customers use today. 

| Feature                               | Cloud                  | Self-hosted      |
| ------------------------------------- | ---------------------- | ---------------- |
| Search your public and private code   | Yes                    | Yes              |
| Code Intel                            | Yes                    | Yes              |
| Code Monitoring                       | Yes                    | Yes              |
| Code Insights [Prototype]             | Yes *                  | Yes              |
| Extensions                            | Yes                    | Yes              |
| Custom Search Contexts                | Yes                    | Yes              |
| SSO                                   | Via Github/Gitlab **   | Enterprise       |
| Easily share private code             | No                     | Yes              |
| Code never leaves your environment    | No                     | Yes              |
| User Management                       | No                     | Yes              |
| Batch Changes                         | No                     | Enterprise       |

<sub>\* Code insights on Sourcegraph cloud is performant across at most 50-75 repositories today. We're working to improve performance across more than 75 repositories.</sub>  
<sub>\** Sourcegraph cloud provides SSO through GitHub.com / GitLab.com as a proxy. For example, a company using Okta SSO can require team members to sign in to GitHub with Okta SSO, and then use GitHub to sign in to Sourcegraph cloud. [Repository permissions on code hosts are always respected](../code_search/explanations/code_visibility_on_sourcegraph_cloud.md).</sub>

Installation instructions for self-hosted Sourcegraph [can be found here](https://docs.sourcegraph.com/admin/install). It is free to install and comes with a 30 day enterprise trial.  [Speak to a product](https://about.sourcegraph.com/contact/sales) specialist to learn more. 
