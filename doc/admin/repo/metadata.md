# Custom repository metadata

<aside class="experimental">
<span class="badge badge-experimental">Experimental</span> Tagging repositories with key/value pairs is an experimental feature in Sourcegraph 4.0. It's a <b>preview</b> of functionality we're currently exploring to make searching large numbers of repositories easier. To enable this feature, enable the `repository-metadata` feature flag for your org. If you have any feedback, please let us know!
</aside>

Repositories tracked by Sourcegraph can be associated with user-provided key-value pairs. Once this metadata is added, it can be used to filter searches to the subset of matching repositories.

Metadata can be added either as key-value pairs or as tags. Key-value pairs can be searched with the filter `repo:has(mykey:myvalue)`. `repo:has.key(mykey)` can be used to search over repositories with a given key irrespective of its value. Tags are just key-value pairs with a `null` value and can be searched with the filter `repo:has.tag(mytag)`.

## Examples
### Repository owners

One way this feature might be used is to add the owning team of each repository as a key-value pair. For example, the repository `github.com/sourcegraph/security-onboarding` repository is owned by the security team, so we could add `owning-team:security` as a key-value pair on that repository. 

Once those key-value pairs are added, they can be used to filter searches to only the code that is owned by a specific team with a search like `repo:has(owning-team:security) account creation`.

### Maintenance status

Another way this could be used is to associate repos with a maintenance status. Do you have a library that is commonly used but is unmaintained, deprecated, or replaced by a better solution? You can associate these custom statuses with repository metadata. After adding this info to your repositories, you can do things like `-repo:has(status:deprecated)` to exclude all results from deprecated repos.

## Adding metadata

Currently, there are two ways to add metadata to a repository: Sourcegraph's GraphQL API, and the [`src-cli` command line tool](https://github.com/sourcegraph/src-cli). 

### Limitations

- There are no scale limits in terms of number of pairs per repo, or number of pairs globally.
- The size of a field is unbounded, but practically it's better to keep it small for performance reasons.
- There are no limits on special characters in the key-value pairs, but in practice we recommend not using special characters because the search query language doesn’t have full support for escaping arbitrary sequences, in particular `:`, `(` and`)`.

### GraphQL

Metadata can be added with the `addRepoKeyValuePair` mutation, updated with the `updateRepoKeyValuePair` mutation, and deleted with the `deleteRepoKeyValuePair` mutation. You will need the GraphQL ID for the repository being targeted.

```graphql
mutation AddSecurityOwner($repoID: ID!) {
  addRepoKeyValuePair(repo: $repoID, key: "owning-team", value: "security") {
    alwaysNil
  }
}

mutation UpdateSecurityOwner($repoID: ID!) {
  updateRepoKeyValuePair(repo: $repoID, key: "owning-team", value: "security++") {
    alwaysNil
  }
}

mutation DeleteSecurityOwner($repoID: ID!) {
  deleteRepoKeyValuePair(repo: $repoID, key: "owning-team") {
    alwaysNil
  }
}
```

### src-cli

Metadata can be added using `src repos add-metadata`, updated using `src repos update-metadata`, and deleted using `src repos delete-metadata`. You will need the GraphQL ID for the repository being targeted.

```text
$ src repos add-metadata -repo=repoID -key=owning-team -value=security
Key-value pair 'owning-team:security' created.

$ src repos update-metadata -repo=repoID -key=owning-team -value=security++
Value of key 'owning-team' updated to 'security++'

$ src repos delete-metadata -repo=repoID -key=owning-team
Key-value pair with key 'owning-team' deleted.
```
