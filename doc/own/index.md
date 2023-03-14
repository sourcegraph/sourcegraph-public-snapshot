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

Owners can be defined by a username/team name or an email address (for person owners). 

Using email addresses is generally recommended, as email addresses are most likely the same across different platforms, and are independent of a user having registered yet. 
In Sourcegraph, a user can add multiple email addresses to their profile. All of those would match to the same user.

For committed CODEOWNERS files, the usernames are usually the username **on the code host**, so they don't necessarily match with the Sourcegraph username. 
This is a known limitation, and in the future we will provide ways to map external code host names to Sourcegraph users. 
For now, you can search for a user by their code host username, or switch to using emails in the codeowners files, which will work across both Sourcegraph and the code host.

```
TODO: DESCRIBE CODEOWNERS FORMAT HERE.
```

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

### Limitations

- Sourcegraph Own is being released as an MVP for 5.0. In the future of the product we intend to infer ownership beyond CODEOWNERS data.
- The feature has not been fully validated to work well on large repositories or large CODEOWNERS rulesets. This is a future area of improvement, but please contact us if you run into issues.