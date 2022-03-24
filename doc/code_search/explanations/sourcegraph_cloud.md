# Sourcegraph Cloud

[Sourcegraph Cloud](https://sourcegraph.com/search) lets you search across your code from GitHub.com or GitLab.com, and across any open-source project on GitHub.com or Gitlab.com. Sourcegraph Cloud is in Public Beta, allowing any individual to sign-up, connect personal repositories, and search across personal code. 

Note that you can search across a maximum of 2,000 repositories at once using Sourcegraph Cloud. To search across more than 2,000 repositories at once or to search code hosted in an on-prem environment, [run your own Sourcegraph instance](../../../admin/install/index.md).

## Explanations and how-tos

- [Adding repositories to Sourcegraph Cloud](../how-to/adding_repositories_to_cloud.md)
- [Searching across repositories you’ve added to Sourcegraph Cloud with search contexts](../how-to/searching_with_search_contexts.md)
- [Who can see your code on Sourcegraph Cloud](./code_visibility_on_sourcegraph_cloud.md)

## FAQ

### What is Sourcegraph Cloud?

Sourcegraph Cloud is a Software-as-a-Service version of Sourcegraph. This means that we handle hosting and updating Sourcegraph so you can focus on what matters: searching your code. Sourcegraph Cloud is available in Public Beta for any individual user to [sign up for free](https://sourcegraph.com/sign-up).

### Limitations

- **Adding repositories**: You can add a maximum of 2,000 repositories hosted on Github.com or Gitlab.com to Sourcegraph Cloud. To add more than 2,000 repositories or to search code hosted in environments other than GitHub.com or GitLab.com, [run your own Sourcegraph instance](../../../admin/install/index.md).
- **Searching code**: You can search across a maximum of 50 repositories at once with a `type:diff` or `type:commit` search using Sourcegraph Cloud. To search across more than 50 repositories at once, [run your own Sourcegraph instance](../../../admin/install/index.md).
- **Organizations and collaboration**: [Sourcegraph Cloud for Teams](../../cloud/index.md#sourcegraph-cloud-for-teams) is currently in Private Beta.

### Who is Sourcegraph Cloud for / why should I use this over Sourcegraph self-hosted?

Sourcegraph Cloud is designed for individual developers to connect and search personal code stored on Github.com or Gitlab.com. While our self-hosted product provides an incredible experience for enterprises, we've heard feedback that individual developers want a simple way to search personal code. 

[A local Sourcegraph instance](../../../admin/install/index.md) is a better fit for you if:

- You have source code stored on-premises
- You would like to create an organization for teammates who need to share code
- You are interested in enterprise solutions such as [Batch Changes](https://about.sourcegraph.com/batch-changes/) to make large-scale code 
- You require more robust admin and user management tooling

Learn more about [how to run your own Sourcegraph instance](../../../admin/install/index.md).

### What are the differences between Sourcegraph Cloud and self-hosted Sourcegraph instances?

Both Sourcegraph Cloud and self-hosted Sourcegraph instances power the same search experience relied on by developers around the world. The Sourcegraph team is working on bringing Sourcegraph Cloud to feature parity with our self-hosted Sourcegraph solution. See a [full breakdown between Sourcegraph Cloud, self-hosted, and enterprise](../../cloud/cloud_ent_on-prem_comparison.md).

### How secure is Sourcegraph Cloud? Can Sourcegraph see my code?

Even though Sourcegraph Cloud is in private beta, it has been designed with security and privacy at the core. No Sourcegraph user, admin, or Sourcegraph employee has access to your private code. This functionality has been rigorously tested during a 2 month private beta with hundreds of users who connected more than 15,000 repositories. In addition, prior to Public Beta Sourcegraph conducted a robust 3rd party penetration test and regularly conducts internal security audits. 

See also:

- [Who can see your code on Sourcegraph Cloud](./code_visibility_on_sourcegraph_cloud.md)
- [Our security infrastructure](https://handbook.sourcegraph.com/engineering/cloud/security/infrastructure)
- [Our Terms of Service](https://about.sourcegraph.com/terms-dotcom) and [Privacy Policy](https://about.sourcegraph.com/privacy/)

If you have further questions, reach out to our [security team](mailto:security@sourcegraph.com).

### How can I share this with my teammates?

It's easy to share Sourcegraph with your team. Each team member must [sign up for Sourcegraph](https://sourcegraph.com/sign-up). From there, anytime you want to share a search, simply search for what you're looking for in Sourcegraph, copy the URL, and share with your teammate. As long as they have permissions to see the code you're trying to share, they will see the search.

### How do I use Sourcegraph Cloud for my organization?

Sourcegraph Cloud only supports individual use today. This means that anyone can sign up for Sourcegraph.com, connect public or private repositories hosted on Github.com or Gitlab.com, and leverage the powerful code search of Sourcegraph. To create and manage an organization with Sourcegraph with team-oriented functionality, get started with the [self-hosted deployment](../../../admin/install/index.md).

### What if my code is not hosted on Github.com or Gitlab.com?

Today, only Github.com or Gitlab.com are supported on Sourcegraph Cloud. To search your code hosted on other code hosts, get started with the [self-hosted version of Sourcegraph](../../../admin/install/index.md).
