# Sourcegraph Cloud

[Sourcegraph Cloud](https://sourcegraph.com/search) lets you search your personal (public or private) code on GitHub.com or GitLab.com, and more than a million of open-source projects.

Note that you can search across a maximum of 2,000 repositories at once using Sourcegraph Cloud. To search across more than 2,000 repositories at once or to search code hosted in an on-prem environment, [run your own Sourcegraph instance](../../../admin/install/index.md).

Note that **Organizations** feature is disabled in Sourcegraph Cloud. If you want to use it, you also need to [run your own Sourcegraph instance](../../../admin/install/index.md). 


## Explanations and how-tos

- [Adding repositories to Sourcegraph cloud](../how-to/adding_repositories_to_cloud.md)
- [Searching across repositories you’ve added to Sourcegraph Cloud with search contexts](../how-to/searching_with_search_contexts.md)
- [Who can see your code on Sourcegraph cloud](./code_visibility_on_sourcegraph_cloud.md)

## FAQ
### What is Sourcegraph Cloud?
Sourcegraph cloud is a Software-as-a-Service version of Sourcegraph. This means that we handle hosting and updating Sourcegraph so you can focus on what matters, searching your code. Sourcegraph cloud is available in public beta today for any individual user to [sign-up](https://sourcegraph.com/sign-up) for free. 

### What are the differences between Sourcegraph cloud and Sourcegraph on-prem/Enterprise?

Sourcegraph cloud is fundamentally similar to Sourcegraph on-prem. Both have the same search capabilities that help developers understand big code. Sourcegraph cloud is still under active development and missing several key organizational features. See a [full breakdown between cloud, on-prem, and enterprise](../../cloud/cloud_ent_on-prem_comparison.md).

### What if I want to use Sourcegraph for my organization?

Sourcegraph cloud only supports individual users today. This means that any user can sign up for Sourcegraph.com, connect public or private repositories hosted on Github.com or Gitlab.com, and leverage the powerful code search of Sourcegraph. Organizations are supported in Sourcegraph on-prem. Learn how to [run your own Sourcegraph instance](../../../admin/install/index.md).

### What if my code is not stored on Github.com or Gitlab.com?
Today, only Github.com or Gitlab.com are supported on Sourcegraph cloud, though many other code-hosts are supported in our self-hosted version of Sourcegraph. Learn how to [run your own Sourcegraph instance](../../../admin/install/index.md).

### Who is Sourcegraph cloud for / why should I use this over local install?
Sourcegraph cloud is designed for individual developers to connect and search personal code stored on Github.com or Gitlab.com. While our on-prem product provides an incredible experience for enterprises, we've heard feedbackthat individual developers want a simple way to search personal code. 

If you meet one of the below descriptions, [a local Sourcegraph instance](../../../admin/install/index.md) is a better fit for you: 
- You have source code stored on-prem 
- You are interested in using [Batch Changes](https://about.sourcegraph.com/batch-changes/) to make large-scale code 
- You require more robust admin and user management tooling

Learn more about how to [run your own Sourcegraph instance](../../../admin/install/index.md).

### How secure is Sourcegraph cloud? Can Sourcegraph see my code?
Sourcegraph cloud has been designed with security and privacy at the core. No Sourcegraph user, admin, or Sourcegraph employee has access to your private code. You can read more detail about [who can see your code on Sourcegraph cloud](./code_visibility_on_sourcegraph_cloud.md) Prior to public beta, Sourcegraph conducted a robust 3rd party penetration test and regularly conducts internal security audits. 

You can read more about our [security infrastructure](https://about.sourcegraph.com/handbook/engineering/security/infrastructure). Further, our [Terms of Service](https://about.sourcegraph.com/terms-dotcom) and [Privacy Policy](https://about.sourcegraph.com/privacy/) outline the ways we respect your privacy. If you have further questions, reach out to our [security team](mailto:security@sourcegraph.com).


### How can I share this with my teammates?
It is easy to share Sourcegraph with your team. Each team member must [sign-up](https://sourcegraph.com/sign-up) for Sourcegraph. From there, anytime you want to share a search, simply search for what you're looking for in Sourcegraph, copy the URL, and share with your teammate. As long as they have permissions to see the code you're trying to share, they will see the search. 


## Search contexts

>NOTE: This feature is still in active development.

Search contexts help you search the code you care about on Sourcegraph cloud. A search context represents a set of repositories on Sourcegraph cloud that will be targeted by search queries by default.

Sourcegraph Cloud supports two search contexts:

- Your personal context, `context:@username`, which automatically includes [all repositories you add to Sourcegraph](../how-to/adding_repositories_to_cloud.md).
- The global context, `context:global`, which includes all repositories on Sourcegraph cloud.

Keep an eye on our Sourcegraph Twitter for more information about upcoming releases.
