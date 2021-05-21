# Search contexts

Search contexts help you search the code you care about on Sourcegraph. A search context represents a set of repositories at specific revisions on a Sourcegraph instance that will be targeted by search queries by default.

Every search on Sourcegraph uses a search context. Search contexts can be defined with the contexts selector shown in the search input, or entered directly in a search query.

## Using search contexts

The search contexts selector is shown in the search input. All search queries will target the currently selected search context. 

To change the current search context, press the contexts selector. All of your search contexts will be shown in the search contexts dropdown. Select or use the filter to narrow down to a specific search context. Selecting a different context will immediately re-run your current search query using the currently selected search context.

Search contexts can also be used in the search query itself. Type `context:` to begin defining the context as part of the search query. When a context is defined in the search query itself, it overrides the context shown in the context selector.

You can also search across multiple contexts at once using the `OR` [boolean operator](/code_search/reference/queries#boolean-operators). For example:

`(context:release1 OR context:release2 OR context:release3) someTerribleBug` 

## Creating search contexts

**Note**: Creating search contexts is only supported on private Sourcegraph instances. Sourcegraph Cloud does not yet support custom search contexts. Want early access to custom contexts on Sourcegraph Cloud? [Let us know](mailto:feedback@sourcegraph.com).

When search contexts are [enabled on your private Sourcegraph instance](/code_search/explanations/features#search-contexts-beta), you can create your own search contexts.

A search context consists of a name, description, and a set of repositories at one or many revisions.

Contexts can be owned by a user, and can be private to the user or public to all users on the Sourcegraph instance.

Contexts can also be at the global instance level, and can be private to site admins or public to all users on the Sourcegraph instance.\

To create a search context:

- Go to **User menu > Search contexts** in the top navigation bar.
- Press the **+ Create search context** button.
- In the **Owner** field, choose whether you will own the context or if it will be global to the Sourcegraph instance. **Note**: At present, the owner of a search context cannot be changed after being created.
- In the **Context name** field, type in a short, semantic name for the context. The name can be 32 characters max, and contain alphanumeric and `.` `_` `/` `-` characters.
- Optionally, enter a **Description** for the context. Markdown is supported.
- Choose the **Visibility** of this context.
  - Public contexts are available to everyone on the Sourcegraph instance. Note that private repositories will only be visible to users that have permission to view the repository via the code host.
  - Private contexts can only be viewed by their owner, or in the case being globally owned, by site admins.
- In the **Repositories and revisions** configuration, define which repositories and revisions should be included in the search context. Press **Add repository** to quickly add a template to the configuration.
  - Define repositories with valid URLs.
  - Define revisions as objects in an array. To specify the default branch, you can set `"rev"` to `"HEAD"` or `""`.

For example:
  
```json
    [
      {
        "repository": "github.com/sourcegraph/sourcegraph",
        "revisions": [
          "3.15"
        ]
      }, {
        "repository": "github.com/sourcegraph/src-cli",
        "revisions": [
          "3.11.2"
        ]
      }, 
    ]
```

- Press **Test configuration** to validate the repositories and revisions.
- Press **Create search context** to finish creating your search context.

You will be returned to the list of search contexts. Your new search context will appear in the search contexts selector in the search input, and can be [used immediately](#using-search-contexts).

## Search contexts on Sourcegraph Cloud

Please see [searching across repositories you've added to Sourcegraph Cloud with search contexts](searching_with_search_contexts.md).
