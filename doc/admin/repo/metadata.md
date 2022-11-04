# Repository metadata

<aside class="experimental">
<span class="badge badge-experimental">Experimental</span> Tagging repositories with key/value pairs is an experimental feature in Sourcegraph 4.0. It's a <b>preview</b> of functionality we're currently exploring to make searching large numbers of repositories easier. If you have any feedback, please let us know!
</aside>

Repositories tracked by Sourcegraph can be associated with user-provided key-value pairs. Once this metadata is added, it can be used to filter searches to the subset of matching repositories.

Metadata can be added either as key-value pairs or as tags. Key-value pairs can be searched with the filter `repo:has(mykey:myvalue)`. Tags are just key-value pairs with a `null` value and can be searched with the filter `repo:has.tag(mytag)`.

## Examples
### Repository owners

One way this feature might be used is to add the owning team of each repository as a key-value pair. For example, the repository `github.com/sourcegraph/security-onboarding` repository is owned by the security team, so we could add `owning-team:security` as a key-value pair on that repository. 

Once those key-value pairs are added, they can be used to filter searches to only the code that is owned by a specific team with a search like `repo:has(owning-team:security) account creation`.

### GitHub topics

Another way this could be used is to ingest GitHub topics as tags so repositories can be searched by GitHub topic. Once ingested, if you wanted to search for repositories with the github topic `machine-learning`, you could run the search `repo:has.tag(machine-learning)`.

## Adding metadata

Currently, the only way to add metadata to a repo is through Sourcegraph's GraphQL API. Metadata can be added with the `addRepoKeyValuePair` mutation, updated with the `updateRepoKeyValuePair` mutation, and deleted with the `deleteRepoKeyValuePair` mutation. You will need the GraphQL ID for the repository being targeted.

```graphql
mutation AddSecurityOwner($repoID: ID!) {
  addRepoKeyValuePair(repo: $repoID, key: "owning-team", value: "security") {
    alwaysNil
  }
}
```
