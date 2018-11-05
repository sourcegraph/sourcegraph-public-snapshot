# URI schemes

Sourcegraph uses URIs to identify files when communicating with Sourcegraph extensions. (Other usages of URIs in Sourcegraph are [deprecated](#deprecated-usages-of-uris).)

For example, to fetch the contents of a hover, Sourcegraph sends a hover request to Sourcegraph extensions with:

- the URI of the currently viewed file
- the position (line and character) in the file

> NOTE: This document and the URI schemes described here are unrelated to repository names (i.e., the name given to a repository, such as `github.com/facebook/react` or `facebook/react`) and the URL shown in the user's browser when browsing Sourcegraph. Sourcegraph only uses URIs for internal communication with Sourcegraph extensions; the URIs discussed here are not shown to the user.

## Design requirements

When designing how Sourcegraph uses URIs, we had a few requirements:

- The identifier should work well with existing tools' assumptions about URIs. This makes it easier to use existing tools (such as language servers and compilers) with Sourcegraph.
  - Common assumption 1: Each file has a URI that identifies it. So, however we choose to identify files, it needs to have a canonical URI representation.
  - Common assumption 2: Two URIs refer to the same file if the [URI scheme, authority, and path components](https://tools.ietf.org/html/rfc3986#section-1.1.1) match. The URI query and fragment are ignored. So, we can't encode the file path in the URI query or fragment.
  - Common assumption 3: All URI schemes support [RFC 3986's relative URI reference resolution](https://tools.ietf.org/html/rfc3986#section-5). If a tool (such as a compiler) is reading a file and encounters a directive for it to read a second file relative to the first file (e.g., a `require` or `#include`), it can construct the second file's URI by itself (without knowing details of the URI scheme's composition). So, we need URIs that support the standard URI reference resolution scheme.
- The URI should not (and need not) imply that it encodes the necessary information to fetch the resource (e.g., to clone its repository).
  - The only way for the URI to *actually* guarantee its holder could fetch the content (without needing to know any other information) would be to use a content-addressable URI scheme, which is impossible given the above requirement.
  - A URI that imperfectly/sometimes allows its holder to fetch the content (without needing to know any other information) is bad because there are so many common cases where it wouldn't work or would become unwieldy:
    - Repository authorization (in Sourcegraph and in code hosts) is per-user, so URIs would need to encode ephemeral authentication information (thereby requiring the holder to occasionally refresh the URI).
    - Cloning repositories requires different authentication depending on the code host and site configuration. In some cases, the URI would need to include an entire SSH key, which would make it very long (and insecure). To avoid this, you could require tools to always clone via Sourcegraph, but then there is no benefit derived from encoding the (partial) cloning information in the URI because the client is treating it opaquely.
    - The clone URL of a repository is not (in general) derivable from the repository's name on Sourcegraph. So, both would need to be encoded.

## When to use (and not use) URIs

Only use URIs when absolutely necessary.

Use URIs when all or most of the following are true:

- The client application that receives/holds the URI needs to support dealing with files from many underlying data sources (i.e., needs to use a virtual file system-like abstraction).
- The client application can treat the URIs opaquely and does not need to know metadata, such as the repository name and revision of the URI's resource.
- The client application is written by people who are not familiar with Sourcegraph's concepts of repository name, input revision, SHA, etc.

Do not use URIs when all or most of the following are true:

- The client application only ever needs to fetch data from one source.
- The client application always needs to know metadata about the URI, such as the repository name and revision it refers to. (For example, `searcher` needs to know the repository name and revision to search inside of. These value should be passed as separate fields, not in a URI.)
- The client application needs to use nuanced Sourcegraph-specific concepts such as distinguishing between the user's unresolved input revision (e.g., `master~2`) and the resolved SHA.

## URI schemes

### repo: URI scheme

URIs of the form `repo://REPO/PATH` refer to a file at `PATH` in a repository named `REPO`.

These URIs are intentionally ambiguous and can't be parsed into their components (`REPO` and `PATH`) by lexicographical manipulation alone. The single source of truth for resolving URIs lives in Sourcegraph itself and is exposed to extensions and other clients in the GraphQL API. This constraint discourages clients from making poor design decisions (as discussed in the [design requirements](#design-requirements) and in the [deprecated git: URI scheme](#git-uri-scheme)).

Examples (which are described/justified in the following sections):

- `repo://github.com/facebook/react/a40da257309d280a266e48f831f7b573fdd51c3c/mydir/myfile.txt`
- `repo://123/a40da257309d280a266e48f831f7b573fdd51c3c/mydir/myfile.txt`
- `repo://cAIAT3mIf9KWQVqoVh8abjiX1Ok/mydir/myfile.txt`

#### Parsing repo: URIs

To parse the components of the `repo:` URI given a URI of the form `repo://COMPONENTS`:

- The set of repository names must be known.
- No repository name may be a prefix of another repository's name. (For example, the repositories `a/b` and `a/b/c` can't coexist.)

The parser scans all repository names to find one that is a prefix of `COMPONENTS`. That repository name becomes `REPO`. The rest of the `COMPONENTS` are the `PATH`.

For example, suppose we have a URI `repo://github.com/alice/myrepo/mydir/myfile.txt` and the known set of repository names `{github.com/alice/myrepo, github.com/bob/myrepo}`. The URI is parsed into `REPO=github.com/alice/myrepo` and `PATH=mydir/myfile.txt` because no other repository name matches.

#### Enforcing immutability and opaque URIs

> NOTE: This section describes an implementation detail that fits within (but is not required by) the spec.

It is useful to think of some URIs as mutable and some as immutable. We will define "immutable" loosely, as "either fails to resolve, or resolves to the same bytes each time, assuming no Git SHA-1 collisions and that you don't tinker with repository IDs in Sourcegraph's PostgreSQL database". For example:

- Example mutable URIs: `file:///` and `https://` URIs (you can rename files at any time, so these URIs might point to different bytes depending on a lot of external state)
- Example immutable URI: a URI that represents file path mydir/myfile.txt at revision `a40da257309d280a266e48f831f7b573fdd51c3c` in repository ID 123

Whether to use a mutable or immutable URI depends on the context:

1. If the URI is actually resolved each time it's fetched, without caching, then use a mutable URI. For example, your editor remembers the file path of each file, not its inode, because you want it to write to a file path, not to disk blocks.
1. If the URI resolution is cached or pinned, then use an immutable URI. For example, if you view a branch on Sourcegraph and navigate to many files in the branch, you will view the files at the SHA that the branch originally resolved to. This is to provide consistency. Likewise, if a push occurs while you're viewing a file, Sourcegraph won't automatically update your view. (It's a fair point to suggest implementing these, but we haven't yet.)

In case (2), if Sourcegraph extensions used mutable URIs, then the application would behave incorrectly when the repository name or Git revspec is changed between when the user opened the repository on Sourcegraph and when the extension fetches the file's content. This would be bad.

So, the web app and browser extension should use immutable URIs. This means URIs like `repo://123/a40da257309d280a266e48f831f7b573fdd51c3c/mydir/myfile.txt`. To discourage clients from constructing these URIs on their own (if that is desirable), Sourcegraph could encode the repository ID and SHA into a single value and use a URI like `repo://cAIAT3mIf9KWQVqoVh8abjiX1Ok/mydir/myfile.txt`.

#### Incorporating the revision in the URI

> NOTE: This section describes an implementation detail that fits within (but is not required by) the spec.

The `REPO` component in a `repo:` URI can include a Git revspec. This is useful because a revision is an important part of the logical identifier for a file. The holder of the URI treats the `REPO` path components as an opaque value, so this is allowed by the spec.

To do so, Sourcegraph can append the Git revspec path components to `REPO`. For example, a branch named `alice/mywip` would yield a URI like `repo://example.com/myrepo/alice/mywip/mydir/myfile.txt`.

At first, this seems ambiguous. By adding a few constraints, we remove the ambiguity:

- The set of Git ref names for the repository must be known. (This is a trivial additional requirement because the set of all repository names is already necessary for fetching resource contents given a URI, as noted above.)
- The following Git revision specifiers (see `man gitrevisions`) are not supported: `:/<text>`, `<rev>:<path>`, `:<n>:<path>`.

Also note that Git already enforces the constraint that no ref name may be a prefix of another ref name. (This may or may not hold for other VCS systems, but currently only Git is supported and in scope.)

To illustrate why this eliminates ambiguity, consider the following well-known example: a GitHub URL like https://github.com/facebook/nuclide/blob/master/scripts/create-package.py. How does GitHub know that the branch name is `master`, not `master/scripts`? It can eliminate the latter possibility by knowing that `master` is a branch and therefore `master/scripts` can't be a branch. This also makes sense given the structure of `.git/refs`: you couldn't simultaneously have `.git/refs/heads/master` and `.git/refs/heads/master/scripts` *both* be files.

### Deprecated URI schemes

#### git: URI scheme

> NOTE: This URI scheme is deprecated (see below).

URIs of the form `git://REPO?REV#PATH` refer to a file or directory (or other Git object) at `PATH` in a Git repository named `REPO` at revision `REV`. For example:

- `git://github.com/gorilla/mux?master#route.go` refers to https://github.com/gorilla/mux/blob/master/route.go
- `git://github.com/gorilla/mux?3d80bc801bb034e17cae38591335b3b1110f1c47#route.go` refers to https://github.com/gorilla/mux/blob/3d80bc801bb034e17cae38591335b3b1110f1c47/route.go

This URI scheme is used by `lsp-proxy`, search- and file location-related GraphQL APIs, and Sourcegraph extensions.

The `git:` URI scheme is **deprecated**.

When communicating with external tools that need a single URI to represent files/directories, use the opaque `repo:` and `repo+rev:` URI schemes defined above because:

- Many tools (such as language servers and editors) do not support it because they ignore the URI query and fragment when asking whether two URIs refer to the same file. Therefore they behave as though all `git:` URIs for the same repository are actually for the same file. This leads to very confusing behavior. It is possible to work around this problem by adding a translation layer, but that adds a lot of complexity.
- Being able to parse the repository name, revision, and file path from the URI (and construct the URIs manually) leads to clients making poor design decisions that frequently lead to bugs.

  Example: Sourcegraph compares repository names case insensitively. If a client constructs a URI manually from an uppercase repository name (or otherwise obtains such a URL) and compares it against an equivalent lowercase URI, the client must know that they can be compared case-insensitively, or else it will incorrectly treat them as distinct. The best solution is for Sourcegraph to canonicalize all URIs it generates, but this breaks if many clients are manually constructing URLs (and not canonicalizing them).

When possible (such as when communicating only among internal Sourcegraph services), pass along each individual field (`REPO`, `REV`, and `PATH`) separately in a Go struct or JSON object to avoid needing to depend on parsing and serializing those values. This also gives you more control over the behavior (such as wanting to preserve the user's input revision in the URL or UI but resolve it to a full SHA for the underlying operations).

##### Deprecated usage of URIs in search

Some parts of search (`searcher`, `indexed-search`, and the GraphQL API for search) use URIs to refer to files. The downsides of using URIs in these systems is:

- It introduces needless complexity because the clients of the response always need to immediately parse the URIs in the results to obtain the same values: repository name, revision, and file path.
- It results in duplicate data being sent. For `searcher`, all results in a response will be for the same repository and revision, so only the file paths need to be sent.
- It often requires clients to discard parts of the URI (and add custom logic to do so). For example, if the client resolves the user's input Git revspec to a SHA before calling `searcher`, the URIs in the response will all have the SHA as the revision. The client needs to remove the SHA and replace it with the input revision before returning the data to the user.

The consistency benefits of using URIs do not outweigh these downsides.

The recommended alternative is for these systems to just return the information in separate fields (some subset of file path, revision, and repository name).
