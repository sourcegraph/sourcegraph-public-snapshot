# Search scopes

Every project and team has a different set of repositories they commonly work with and search over. Custom search scopes enable users and organizations to quickly filter their searches to predefined subsets of files and repositories.

A search scope is any valid query. For example, a search scope that defines all repositories in the "example" organization would be `repo:^github\.com/example/`.In the UI, search scopes appear as buttons on the homepage and below the search bar when you execute a search.

---

## Creating custom search scopes

Custom search scopes can be specified at 3 different levels:

- By site admins for all users: in the **Global settings** in the site admin area.
- By org admins for all org members: in the org profile **Configuration** section
- By users for themselves only: in the user profile **Configuration** section

You can configure search scopes by setting the `search.scopes` to a JSON array of `{name, value}` objects.

The `value` of a search scope can be any valid query and can include any [search token](queries.md) (such as `repo:`, `file:`, etc.).

For example, this JSON will create two search scopes:

```json
{
  // ...
  "search.scopes": [
    {
      "name": "Test code",
      "value": "file:(test|spec)"
    },
    {
      "name": "Non-vendor code",
      "value": "-file:vendor/ -file:node_modules/"
    }
  ]
  // ...
}
```

After editing and saving the configuration settings JSON in the profile page, the new search scopes take effect immediately. Instead of typing the set of files and repos you want to search over, you can now select your search scope from the buttons on the search homepage whenever you need.

---

## Creating search scope pages

You can also create search scope pages. These pages will automatically set the search scope and list all repositories included in the scope. This is useful for users who will always search within a defined scope. To create a search page for this scope, include the optional `id` and `description` fields. The scope page URL will be `$SOURCEGRAPH_URL/search/scope/{id}`.

For example:

```json
{
  // ...
  "search.scopes": [
    {
      "name": "Test code",
      "value": "file:(test|spec)",
      "id": "test-code",
      "description": "Search over test files"
    }
  ]
  // ...
}
```

The search scope page is only accessible to users that can use this search scope. (Only the single user can access the search scope page for a search scope defined in user settings. All organization members can access an organization search scope's page. All users can access a global search scope's page.)

Example scope pages for open source repositories:

- [Top 1000 Rust crates](https://sourcegraph.com/search/scope/crates)
- [Popular npm packages](https://sourcegraph.com/search/scope/npm)
