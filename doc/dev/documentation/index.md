# Sourcegraph documentation guidelines

> NOTE: Adapted from [GitLab documentation guidelines](https://gitlab.com/gitlab-org/gitlab-ee/raw/master/doc/development/documentation/index.md).

- **General documentation**: written by the [developers responsible by creating features](#contributing-to-docs). Should be submitted in the same pull request containing code. Feature proposals (by Sourcegraph contributors) should also be accompanied by its respective documentation. They can be improved later.
- **[Technical articles](#technical-articles)**: written by any Sourcegraph staff member, Sourcegraph contributors, or any member of the community.
- **Indexes per topic**: kept up-to-date by developers and PMs in the same pull request containing code. They gather all resources for that topic in a single page (user and admin documentation, articles, and third-party docs).

## Contributing to docs

Whenever a feature is changed, updated, introduced, or deprecated, the merge
request introducing these changes must be accompanied by the documentation
(either updating existing ones or creating new ones). This is also valid when
changes are introduced to the UI.

The one responsible for writing the first piece of documentation is the developer who
wrote the code. It's the job of the Product Manager to ensure all features are
shipped with its docs, whether is a small or big change. At the pace Sourcegraph evolves,
this is the only way to keep the docs up-to-date. If you have any questions about it,
ask a Technical Writer. Otherwise, when your content is ready, assign one of
them to review it for you.

We use the [monthly release blog post](https://about.sourcegraph.com/blog) as a changelog checklist to ensure everything
is documented.

## Documentation structure

Follow through the [documentation structure guide](structure.md) for learning how to structure
Sourcegraph docs.

## Documentation directory structure

The documentation is organized into the following top-level directories:

- [`user`](https://github.com/sourcegraph/sourcegraph/tree/master/doc/user) for users
- [`admin`](https://github.com/sourcegraph/sourcegraph/tree/master/doc/admin) for site admins
- [`extensions`](https://github.com/sourcegraph/sourcegraph/tree/master/doc/extensions) for Sourcegraph extensions
- [`integration`](https://github.com/sourcegraph/sourcegraph/tree/master/doc/integration) for integrations with other products
- [`api`](https://github.com/sourcegraph/sourcegraph/tree/master/doc/dev) for the Sourcegraph GraphQL API
- [`dev`](https://github.com/sourcegraph/sourcegraph/tree/master/doc/dev) for contributors

There is no global index or nav, so all docs should be linked from their parent index page. Every new document should be cross-linked to its related documentation, and linked from its topic-related index, when it exists.

### Documentation files

- When you create a new directory, always start with an `index.md` file. Do not use another file name and **do not** create `README.md` files
- **Do not** use special chars and spaces, or capital letters in file names, directory names, branch names, and anything that generates a path.
- Max screenshot size: 100KB
- We do not support videos (yet)

#### General rules & best practices

- When creating a new document and it has more than one word in its name,
  use underscores instead of spaces or dashes (`-`). For example,
  a proper naming would be `import_projects_from_github.md`. The same rule
  applies to images.
- Always cross-link to `.md` files, including the file extension, so that the docs are browseable as-is (e.g., in GitHub's file browser).
- Start a new directory with an `index.md` file.

If you are unsure where a document should live, you can mention `@ryan-blunden` or `@sqs` in your
pull request.

### Changing document location

Changing a document's location is not to be taken lightly. Remember that the
documentation is available to all installations under `/help` and not only to
Sourcegraph.com or http://docs.sourcegraph.com. Make sure this is discussed with the
Documentation team beforehand.

If you indeed need to change a document's location, do NOT remove the old
document, but rather replace all of its contents with a new line:

```
This document was moved to [another location](path/to/new_doc.md).
```

where `path/to/new_doc.md` is the relative path to the root directory `doc/`.

---

For example, if you were to move `doc/user/search/queries.md` to
`doc/user/search/query_syntax.md`, then the steps would be:

1.  Copy `doc/user/search/queries.md` to `doc/user/search/query_syntax.md`
1.  Replace the contents of `doc/user/search/queries.md` with:

    ```
    This document was moved to [another location](query_syntax.md).
    ```

1.  Find and replace any occurrences of the old location with the new one.

## Linting

Currently there is no automatic linting of documentation files. In the future we may add proselint and markdownlint.

## Testing

We treat documentation as code, so we've implemented some testing:

- `docsite check`: Check that all internal (relative) links work correctly.

## Previewing the changes live

When running Sourcegraph in development mode, you can access a live preview of the docs at http://localhost:3080/help. The Markdown rendering and structure is the same as on docs.sourcegraph.com.

You can also follow the [sourcegraph/docs.sourcegraph.com README](https://github.com/sourcegraph/docs.sourcegraph.com) to run the docs.sourcegraph.com site locally.

## Sourcegraph /help

Every Sourcegraph instance includes the documentation at the URL path `/help`
(`https://sourcegraph.example.com/help`), e.g., <https://sourcegraph.com/help>.

The documentation available on docs.sourcegraph.com is continuously <!-- TODO(sqs): set up continuous deploy of docs -->
deployed every hour from the `master` branch of `sourcegraph/sourcegraph`. Once a pull request gets merged, it will be available on docs.sourcegraph.com soon, and its doc changes will ship in the next release.

### Linking to /help

When you're building a new feature, you may need to link to the documentation from the product. In-product documentation links should refer to the documentation hosted on the product itself (at `/help`) so that the product and documentation versions are consistent. (The docs on docs.sourcegraph.com may be for a newer version than what the user is running.)

## Prefer primary documents

We want our written documents to be as up-to-date, accurate, and useful as possible. For this reason, we always prefer creating and using **primary documents** over secondary documents, such as in the case of documentation, plans, or design docs.

A **primary document** is a document that is actually used to specify, communicate, or document something to the entire intended audience and is discoverable by that audience. A secondary document is a document that (effectively) only its creator uses or knows about.

Examples:

- A project's [tracking issues](../product/index.md#planning) are primary documents for "what will ship" because they specify precisely that and everyone would be able to discover them.
  - Secondary documents for a project would be: a Google Doc with a list of TODOs, a checked-in Markdown project plan at `cmd/foo/PLAN.md`.
- A project's [long-term plan](../product/index.md#planning) is a primary document that's either a published blog post (for particularly user-facing projects) or a set of tracking issues on future milestones.
  - We don't have the product management capabilities right now for all projects to have a separate long-term plan separate from these existing documents that we already create.

Add links to your primary document liberally, from anywhere your audience might be looking!

To choose the right place for your primary document to live, ask yourself:

- Where would people expect to find this information?
- How can I maximize the probability that a developer who updates the underlying behavior described by the doc would realize they also need to update the doc? (To avoid the doc diverging from the actual behavior.)

It's OK to create secondary documents for your own temporary use. If anybody else would need to use or discover them, transfer the information to the appropriate primary document (and delete the secondary document or clearly mark it as moved).

> This is similar to saying "prefer a single source of truth". The problem with that, however, is that the designated source of truth might be a secondary document such as a technical spec that the intended audience is unaware of. Our use of the term "primary document" is intended to avoid that problem.

## Product documentation vs Technical articles

### Product documentation

Product documentation describes how to use a specific part of the product (e.g., a feature) for users, admins, extension users/authors, API consumers, and integrators. A good product documentation page includes mention of:

- What problems the feature solves
- Who should use the feature
- Common example use cases for the feature (in what cases would someone use this feature?)
- How to use the feature
- Site configuration and settings that affect the feature
- Security concerns related to the feature
- Performance characteristics of the feature
- Known issues related to using the feature
- Links to other related features
- Future product roadmap plans related to the feature

### Technical articles

Technical articles replace technical content that once lived in the [Sourcegraph Blog](https://about.sourcegraph.com/blog), where they got out-of-date and weren't easily found.

They are topic-related documentation, written with an user-friendly approach and language, aiming to provide the community with guidance on specific processes to achieve certain objectives.

A technical article guides users and/or admins to achieve certain objectives (within guides and tutorials), or provide an overview of that particular topic or feature (within technical overviews). It can also describe the use, implementation, or integration of third-party tools with Sourcegraph.

They should be placed in a new directory named `/article-title/index.md` under a topic-related folder, and their images should be placed in `/article-title/img/`. For example, a new article on Sourcegraph organizations should be placed in `doc/user/organizations/article-title/` and a new article on Sourcegraph extensions should be placed in `doc/extensions/article-title/`.

#### Types of technical articles

- **User guides**: technical content to guide regular users from point A to point B
- **Admin guides**: technical content to guide Sourcegraph site admins from point A to point B
- **Technical overviews**: technical content describing features, solutions, and third-party integrations
- **Tutorials**: technical content provided step-by-step on how to do things, or how to reach very specific objectives

#### Understanding guides, tutorials, and technical overviews

Suppose there's a process to go from point A to point B in 5 steps: `(A) 1 > 2 > 3 > 4 > 5 (B)`.

A **guide** can be understood as a description of certain processes to achieve a particular objective. A guide brings you from A to B describing the characteristics of that process, but not necessarily going over each step. It can mention, for example, steps 2 and 3, but does not necessarily explain how to accomplish them.

- Live example (from GitLab): "[Static sites and GitLab Pages domains (Part 1)](https://docs.gitlab.com/ee/user/project/pages/getting_started_part_one.html) to [Creating and Tweaking GitLab CI/CD for GitLab Pages (Part 4)](https://docs.gitlab.com/ee/user/project/pages/getting_started_part_four.html)"

A **tutorial** requires a clear **step-by-step** guidance to achieve a singular objective. It brings you from A to B, describing precisely all the necessary steps involved in that process, showing each of the 5 steps to go from A to B.
It does not only describes steps 2 and 3, but also shows you how to accomplish them.

- Live example (from GitLab): [Hosting on GitLab.com with GitLab Pages](https://about.gitlab.com/2016/04/07/gitlab-pages-setup/)

A **technical overview** is a description of what a certain feature is, and what it does, but does not walk
through the process of how to use it systematically.

- Live example (from GitLab): [GitLab Workflow, an overview](https://about.gitlab.com/2016/10/25/gitlab-workflow-an-overview/)
