// Package authzchecked wraps stores with authorization checks.
//
// As a defensive programming practice, this package's store wrappers
// use a field for the underlying un-authz-checked store (type A
// struct{b B}) instead of embedding (type A struct{B}). This means
// that if another method is later added to type B, the authzchecked
// implementation will not silently delegate calls to the
// non-authz-checked method (as it would if struct embedding were
// used).
//
//
// DEVELOPMENT NOTES (@sqs)
//
// What happens when you list builds for a federated repo example.com/repo?
//
// 1. Top-level service notices that example.com != local; fans out query to (1) example.com and (2) local builds [because you’ve configured local builds of federated repos]
// 2. (1) is easy, all authz is handled at federation service level
// 3. (2) local builds service needs to be configured to use federated repo lookup (since example.com/repo doesn’t exist locally), so federated service invokes Builds.List with the local Builds (and all other services) but the federated Repos service
// 4. Local Builds.List checks authz for example.com/repo by calling the context-injected Repos service (which correctly returns YES for the authz to example.com/repo) OR
// 5. Local Builds.List proceeds with listing builds for the federated repo
//
// Upshot:
//
// * authz checking needs to consult service-level, not just store-level
//
// What happens when you list builds for a GitHub repo github.com/user/repo?
//
// 1. Top-level service notices that there's no federation origin at github.com, so it calls the local service
// 2. Local service notices that "github.com/..." is mapped to use local DB Builds store but GitHub-backed Repos store, so it invokes Builds.List on a local service and injects that configuration of stores
// 3. That local service calls Builds.List on its Builds store -- which is (authzchecked.Builds).List
// 4. (authzchecked.Builds).List notices the param opt.Repo=="github.com/user/repo", so it consults EITHER(which is best?) the context's repoPermChecker.
//
// Misc. principles:
//
// * The List methods of the lowest-level services (fs.Builds, db.Builds, etc.) should not have to make arbitrary numbers of queries to find out what perms they have. Except for site admins, these queries should be required to be scoped to a specific repo.
//
package authzchecked
