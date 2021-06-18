# Search snippets

Every project and team has a different set of repositories they commonly work with and search over. Custom search snippets enable users and organizations to quickly filter their searches with any search fragment.

A search snippet is any valid query. For example, a search snippet that defines all repositories in the "example" organization would be `repo:^github\.com/example/`. In the UI, search snippet appear as suggested filters in the search sidebar (as of v3.29).DF

NOTE: Search snippets are temporarily named search.scopes in site configuration files. 

---

## Creating custom search snippets

Custom search snippets can be specified at 3 different levels:

- By site admins for all users: in the **Global settings** in the site admin area.
- By organization admins for all organization members: in the organization profile **Settings** section
- By users for themselves only: in the user profile **Settings** section

You can configure search snippets by setting the `search.scopes` to a JSON array of `{name, value}` objects.

The `value` of a search snippet can be any valid query and can include any [search token](../reference/queries.md) (such as `repo:`, `file:`, etc.).

For example, this JSON will create two search snippets:

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

After editing and saving the configuration settings JSON in the profile page, your search snippets will be shown as suggested filters on search results pages.

