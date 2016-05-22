/*

Package server contains subpackages which implement the high-level
logic of the Sourcegraph API. The main implementation is in
"server/local", which relies on "stores" (which may be backed
by PostgreSQL, the filesystem, GitHub, Papertrail, etc.)  for
persistence. The local server implementations are wrapped with
middleware and context initialization functions in this package's
Config func.


SERVERS VS. STORES

The public API service interfaces are defined in ./api/sourcegraph's
package sourcegraph (in the file sourcegraph.proto). The server
package contains multiple implementations of these services
interfaces. The primary implementations are in server/local;
there is also middleware in svc/middleware and
server/internal/middleware that wraps them.

The stores (defined in the store package and implemented in
services/ext/... and server/internal/localstore/...) implement the minimal set of
operations needed by the higher-level servers in server.

This separation has benefits and drawbacks:

1. It enables usage and swappability of multiple backends (PostgreSQL,
   local filesystem, GitHub, etc.), but it increases the number of
   possible configurations we need to (or might be expected to)
   support.

2. It makes each individual method implementation simpler and shorter
   by separating logic and persistence, but it adds complexity to the
   overall call flow.

3. It makes it easier to unit-test server methods, but it makes it
   harder to integration-test them because they can be used with
   (potentially) multiple stores and need additional code to
   instantiate each combination properly in tests.

How do you know whether to implement new functionality in a server
method vs. a store method? If it's specific to the way the data is
persisted, you must put it in the store. If it's independent of how
the data is persisted ("business logic"), use a server method.


AUTHORIZATION AND PERMISSIONS (WORK-IN-PROGRESS)

**Note: This section's suggestions are not fully implemented and are
still open for discussion.**

Stores MUST ensure the actor is authorized before performing any
action or returning any data. The servers in package server MUST NOT
contain any authorization logic; they MUST assume that any methods
they invoke on stores will be auth-checked and any data returned by
store methods has been auth-checked. The goal of this is to reduce the
likelihood of security bugs by making auth-checking code simpler and
independent of implementation code.

There are 2 ways that stores perform authorization:

1. Stores that persist no local state and merely wrap external APIs
   (such as GitHub, Papertrail, etc.) SHOULD rely on the external
   server for authorization checking (e.g., by assuming that GitHub's
   "get repo" API only returns private repos to authorized users). To
   do so, they MUST pass the actor's credentials (if any) with each
   external API call and MUST NOT persist any state.

2. Stores that persist local state/data MUST perform their own
   authorization checking and MUST NOT return data (directly or via
   side channel, to the extent possible) that the actor is not
   authorized to see. Stores SHOULD check authorization when their
	 methods are invoked. When the behavior of a method implicitly
	 depends on the actor's authorization (e.g., Repos.List returns
	 private repos for the actor only), they MUST implement their own
	 authorization logic.

Here are 4 examples to illustrate various cases:

1. The GitHub-backed ext/github.ReposServer doesn't need to perform
   any authorization checks because it merely wraps GitHub's API and
   passes the actor's credentials. It assumes GitHub's API properly
   checks authorization.

2. The DB-backed server/internal/store/pgsql.ReposServer List method
   must ensure that the repositories list it returns are all visible
   to the current actor (i.e., it implicitly depends on the actor's
   authorization). Therefore it must contain custom auth logic.

3. Builds server methods (in all implementations) work with data that
   is related to repositories. A user shouldn't be able to view a
   build for a repository they don't have access to. This means that
   each method should ensure a user can read or write to the
   associated repository before fetching or modifying builds for that
   repository. This is a hybrid case, so it's worth mentioning.

(Historical note: Previously, permission checks were necessarily
intertwined with persistence and logic code. The move to using base
servers means that it is easier to create authz-checking server
wrappers, since the simpler store methods can determine what
permissions to check without needing to duplicate logic. Also, no
longer persisting GitHub data means we can eliminate the vast majority
of permission checks. The necessary optimizations for listing
repositories and auth-checking multiple repositories, in particular,
required us to mix logic and auth checking to achieve decent
performance. Those optimizations are no longer needed.)

*/
package server
