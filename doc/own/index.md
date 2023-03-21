# Sourcegraph Own

<aside class="experimental">
<p>
<span class="badge badge-experimental">Experimental</span> This feature is experimental and might change in the future.
</p>

<p><b>We're very much looking for input and feedback on this feature.</b> You can either <a href="https://about.sourcegraph.com/contact">contact us directly</a>, <a href="https://github.com/sourcegraph/sourcegraph">file an issue</a>, or <a href="https://twitter.com/sourcegraph">tweet at us</a>.</p>
</aside>

Sourcegraph Own is a new, experimental product aimed at helping find the right person and team to contact, for any question, at any time. We are starting out with code ownership, and are exploring ways to help you find someone to answer _every_ question.

## Enabling Sourcegraph Own

As an experimental feature, Sourcegraph Own is disabled by default. If you like to try it, a site-admin can enable the feature flag `search-ownership`:

- Go to **Site-admin > Feature flags**
- If the feature flag `search-ownership` doesn't yet exist, click **Create feature flag**
- Under **Name**, put `search-ownership`
- For **Type** select **Boolean**
- And set **Value** to **True**
- Click **Create flag**

## Concepts

**Owner**: An owner is defined as a person or a team in Sourcegraph.

A _person_ can be:
- a Sourcegraph user which we were able to resolve from the `CODEOWNERS` handle or email, in which case we link to their profile.
- an unknown user for which we were unable to resolve a profile, in which case we will return the `CODEOWNERS` data we have.

A _team_ is a group of Sourcegraph users represented by a common handle, which is a new feature that we added. 
[Read more about how to manage teams in Sourcegraph](../admin/teams).

## Code ownership

Code ownership is defined as a strict ruleset. Files can be assigned owners. 
To define rulesets for codeownership, we make use of the CODEOWNERS format.

### The CODEOWNERS format

`CODEOWNERS` files contain a sequence of matching rules - a glob pattern and zero or more owners. 
A repository has at most one CODEOWNERS file. 

#### Specifying Owner information 

Owners can be defined by a username/team name or an email address. 

Using email addresses is generally recommended, as email addresses are most likely the same across different platforms, and are independent of a user having registered yet. 
In Sourcegraph, a user can add multiple email addresses to their profile. All of those would match to the same user.

For committed CODEOWNERS files, the usernames are usually the username **on the code host**, so they don't necessarily match with the Sourcegraph username. 
This is a known limitation, and in the future we will provide ways to map external code host names to Sourcegraph users. 
For now, you can search for a user by their code host username, or switch to using emails in the codeowners files, which will work across both Sourcegraph and the code host.

#### File format

The following snippet shows an example of a valid CODEOWNERS file. 

```
*.txt @text-team
/build/logs/ alice@sourcegraph.com 
/cmd/**/test @qa-team @user
```

- Asterisk * is a wildcard that matches N tokens in a path segment. Example: `doc/*/own` will match `doc/ref/own` and `doc/tutorial/own`, but not `doc/a/b/own`
- Double ** asterisk matches any sub-path. Example: `doc/**/own` will match `doc/ref/own` and `doc/a/b/own`
- Starting a pattern with / anchors matches at the repository root. Example: `/docs/*` matches `/docs/a.md` and `/docs/b.md` but not `/src/docs/a.md`.
- Trailing slash / matches any file within the directory tree (so it is equivalent to trailing /**). `Example: `docs/` matches `/testing/docs/foo` and `/docs/foo/bar`, but does not match `/docs` or `/testing/docs`.


The rules are considered independently and in order. Rules farther down the file take precedence. Only **one** rule matches.

#### Limitations

- GitLab allows sections in CODEOWNERS files, these are not yet supported and section markers are ignored
- [Code Owners for Bitbucket](https://marketplace.atlassian.com/apps/1218598/code-owners-for-bitbucket?tab=overview&hosting=cloud) inline defined groups are not yet supported

To configure ownership in Sourcegraph, you have two options:

### Committing a CODEOWNERS file to your repositories

> Use this approach if you prefer versioned ownership data.

You can simply commit a CODEOWNERS file at any of the following locations for it to be picked up automatically by Own:

```
CODEOWNERS
.github/CODEOWNERS
.gitlab/CODEOWNERS
docs/CODEOWNERS
```

Searches at specific commits will return any CODEOWNERS data that exists at that specific commit.

### Uploading a CODEOWNERS file to Sourcegraph

> Use this approach if you don't want to commit CODEOWNERS files to your repos, or if you have an existing system that tracks ownership data and want to sync that data with Sourcegraph.

Read more on how to [manually ingest CODEOWNERS data](codeowners_ingestion.md) into your Sourcegraph instance.

The docs detail how to use the UI or `src-cli` to upload CODEOWNERS files to Sourcegraph.

## Limitations

- Sourcegraph Own is being released as an MVP for 5.0. In the future of the product we intend to infer ownership beyond CODEOWNERS data.
- The feature has not been fully validated to work well on large repositories or large CODEOWNERS rulesets. This is a future area of improvement, but please contact us if you run into issues.

## Browsing ownership

The ownership information is available for browsing once ownership data is available through [a CODEOWNERS file](#code-ownership).

When displaying a source file, there is a bar above the file contents.

*   On the left hand side, it displays the most recent change to the file.
*   On the right hand side it displays the Own bar with at most 2 file owners. Any additional number of owners is also displayed.

![File view showing own bar on the right hand side above the file contents](https://storage.googleapis.com/sourcegraph-assets/docs/own/blob-view.png)

After clicking on the Own bar, a bottom panel appears listing all the owners.

![File view with the ownership tab selected in the bottom panel](https://storage.googleapis.com/sourcegraph-assets/docs/own/blob-view-panel.png)

There is always a single rule in a CODEOWNERS file that determines ownership (if any). Each owner listed in the bottom panel has a description found by clicking the collapsible arrow: _Owner is associated with a rule in a CODEOWNERS file_. Clicking this description links to the line containing the responsible rule in the CODEOWNERS file.

If any email information has been found for the owner, clicking the mail icon will  start an email to them. 

## Ownership search

Code ownership is a first-class citizen in search. Ownership can be either a query input or a search result:

*   `file:has.owner(user@example.com)` keeps only the search results associated with given user (here referred to by e-mail).
*   `-file:has.owner(@username)` removes all results owned by specific user (here referred to by name).

Ownership predicate can also be used without parameters:

*   `file:has.owner()` will only include files with an owner assigned to them.
*   `-file:has.owner()` will only include files without an owner.

When performing a search the `select:file.owners` predicate will return the owners for the result of that search.

For instance one can find all the owners of TypeScript files in a given repository by using `repo:^github\.com/sourcegraph/sourcegraph$ lang:TypeScript select:file.owners`.
