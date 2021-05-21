# Searching across repositories you've added to Sourcegraph Cloud with search contexts

Once you've [added repositories to Sourcegraph](adding_repositories_to_cloud.md), you can search across those repositories by default using search contexts.

Sourcegraph Cloud currently supports two search contexts: 

- Your personal context, `context:@username`, which automatically includes all repositories you add to Sourcegraph Cloud.
- The global context, `context:global`, which includes all repositories on Sourcegraph Cloud.

**Coming soon:** create your own search contexts that include the repositories you choose. Want early access? [Let us know](mailto:feedback@sourcegraph.com).


## Using search contexts

The search contexts selector is shown in the search input. All search queries will target the currently selected search context. 

To change the current search context, press the contexts selector. All of your search contexts will be shown in the search contexts dropdown. Select or use the filter to narrow down to a specific search context. Selecting a different context will immediately re-run your current search query using the currently selected search context.

Search contexts can also be used in the search query itself. Type `context:` to begin defining the context as part of the search query. When a context is defined in the search query itself, it overrides the context shown in the context selector.
