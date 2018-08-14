# Scope design

This lays out a proposed generalization from the existing set of scopes. The scopes currently defined are:

```
ScopeUserAll       = "user:all"        // Full control of all resources accessible to the user account.
ScopeSiteAdminSudo = "site-admin:sudo" // Ability to perform any action as any other user.
```

We have been asked to implement a scope which allows updating repos, but nothing else. I'd like to think ahead a bit and define a logical structure for scoping which is reasonably simple to implement now but allows future design flexibility.

## Basic Schema

Each scope, represented textually, consists of two or more colon-separated segments. Each segment consists of one or more comma-separated terms. A term may have leading punctuation (other than colons or commas), must start with a lowercase ASCII letter, and consists of lowercase ASCII letters, hyphens, and numbers.

A thing may have multiple scopes. Scopes are processed in order from most-specific (a scope associated with a particular token) to most-general (a default scope defined site-wide, for instance). The first scope to give a yes or no answer on a given capability determines the answer.

The first segment is called a domain, and denotes the category of functionality restricted. For instance, `user` for functionality associated with the token's user, or `site-admin` for functionality associated with a site admin. (Possibly confusingly, `user` does not denote control over user objects, rather, it refers to the token's user.)

The second segment denotes capabilities, and consists of zero or more comma-separated items.

The third segment, if present, provides parameters to a capability. The arguments are applied to all capabilities in the scope. It is an error to provide parameters to a capability which does not accept them. Parameters are comma-separated strings which can't include colons or spaces.

### Domains

Domains denote a category of functionality. Domains defined in this proposal are:

- `user`: The permissions associated with the given user.
- `site-admin`: Administrative functionality, such as the ability to run as another user.
- `repo`: Control over or interactions with repositories.

### Capabilities

Each domain defines its own capabilities. A scope may explicitly include a capability (by listing it), or explicitly deny it (by prefixing it with a hyphen). For example, the scope `site-admin:-sudo` revokes `sudo` permissions, but does not address any other permissions questions. If the next scope processed said `site-admin:all`, the net result would be permission to perform all site-admin functions _except_ sudo.

#### Capabilities for `user`

The only currently defined capability for `user` is `all`, which gives permission to do things that user could normally do through the web interface.

We would probably like to have `user:readonly`, which gives permission to do things if and only if those things do not modify data.

#### Capabilities for `site-admin`

The only currently defined capability for `site-admin` is `sudo`, which gives permission to use the privileges of a specific user.

#### Capabilities for `repo`

None of these exist yet, this is just a proposal:

- `list`: Obtain information about the existence and metadata of repos.
- `read`: See the contents of repos. The `read` capability may take as parameters one or more repos, using the repo name, with `*` as a wildcard.
- `update`: Request fetches/updates to repos. The `update` capability may take as parameters one or more repos, using the repo name.

## Examples

- A scope which permits requesting updates to repositories, but nothing else: `repo:update`
- A scope which permits reading the contents of all repos on github associated with Sourcegraph: `repo:read:github.com/sourcegraph/*`

## Implementation

The proposed implementation is to do nothing but add `repo:update` as a third special case in the auth code, and allow specification of a scope at token creation. The purpose of writing down a design here is to try to have guidance for future design, so that if and when we're ready to add new capabilities or scopes, we can do so consistently and not create new parsing hassles.

As a longer-term thing, I'd like to have structured data in the database making it easy to perform queries to determine the actual scope created by a `user:all` token. Remember that `user:all` implies permission to do whatever the user could; in future, we could allow admins to create scope rules which restrict what a given user can do.

It would probably make sense to create some kind of "group" feature, where a user can be a member of one or more groups, and also allow some structured way of grouping repos. (Possibly by identifying code host accounts/tokens.)
