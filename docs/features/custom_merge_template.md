+++
title = "Custom Merge Template"
+++

When a Changeset is merged via the web UI or `src changeset merge` CLI a merge
commit will be added to the history based on either the default template or one
you specify.

# Default Template

By default, the template used is:

```
{{.Title}}

Merge changeset #{{.ID}}

{{.Description}}
```

# Custom Template

Sourcegraph looks for a `.sourcegraph-merge-template` file at the base branch of
your changeset (e.g. `master`) and will use that template when available.

The syntax is [Go text/template syntax](golang.org/pkg/text/template) and thus
you can use any of the following in your template file:

| Template Syntax | Description          | Example                                       |
|-----------------|----------------------|-----------------------------------------------|
| `{{.ID}}`       | CS ID number         | 4                                             |
| `{{.Title}}`    | CS Title string      | `httpapi: fix a bug with the query method`    |
| `{{.URL}}`      | URL to view CS       | `https://src.myteam.com/myrepo/.changesets/4` |
| `{{.Author}}`   | CS Author's username | `bill33`                                      |
