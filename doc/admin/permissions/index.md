# Permissions

Sourcegraph can be configured to enforce the same access to repositories and underlying 
source files as your code  host. If configured, Sourcegraph will allow the user to only 
see the entities that they can see on the code host. These permissions are enforced 
accross the product for all the use cases that need to read data from a repository, 
including the existence of such repository on the code host.

> NOTE: Historically, we have referred to permissions-related features under different 
names: "authorization", "repository permissions", "code host permissions". All of these 
terms were used interchangeably, but in general we do call it permissions or repository 
permissions. That's not to be confused with in-product permissions, which determine who 
can, for example, create a batch change or who is a site admin.

## Example

Imagine a scenario, with 2 users, `alice` and `bob`: 

- `alice` can access repositories `horsegraph/global` and `horsegraph/alice` on the code host
- `bob` has access to repositories `horsegraph/global` and `horsegraph/docs` on the code host
- there is also a public repository `horsegraph/public`
- all of the mentioned repositories are synced to Sourcegraph

If `alice` tries to run a search on Sourcegraph, the results will only contain data from the public `horsegraph/global` repository 
and the ones she has access to (`horsegraph/global` and `horsegraph/alice`). `alice` will not be able to see results 
from the `horsegraph/docs`. If `alice` creates a code insight, she will only see results from the repositories she has access to.

Same for `bob`, the search results or any other feature will not show him the existence of `horsegraph/alice` repository on 
Sourcegraph, since `bob` does not have access to it on the code host.

## Supported methods to sync permissions

Today, we support 3 different methods to get the permission data from code host to Sourcegraph:

1. [Permission syncing from the code host](syncing.md)
1. [Webhooks for getting permission events from code host](webhooks.md)
1. [Explicit permissions API](api.md)

To know more about each method that we support, please follow the link above.

## Supported code hosts

Support for repository permissions accross different code hosts is different. The following table captures current state of support (ordered alphabetically):
| Code host | [Permission Syncing](syncing.md) | [Webhooks for Permissions](webhooks.md) | [Explicit API](api.md) | [Scale supported](#supported-scale) |
| -------- | -------- | -------- | -------- | -------- |
| Bitbucket Cloud <span class="badge badge-beta">Beta</span> | ✓ | ✗ | ✓ | 10k users, 100k repositories |
| Bitbucket Server | ✓ | ✗ | ✓ | 10k users, 100k repositories |
| Gerrit <span class="badge badge-beta">Beta</span> | ✓ | ✗ | ✓ | 10k users, 100k repositories |
| GitHub   | ✓ | ✓ | ✓ | 40k users, 200k repositories |
| GitHub Enterprise | ✓ | ✓ | ✓ | 40k users, 200k repositories |
| GitLab | ✓ | ✗ | ✓ | 40k users, 200k repositories |
| GitLab Self-Managed | ✓ | ✗ | ✓ | 40k users, 200k repositories |
| Perforce <span class="badge badge-experimental">Experimental</span> | Yes <span class="badge">(with file-level permissions)</span> | ✓ | ✓ | 10k users, 250k repositories |

All the other code hosts only support [Explicit permissions API](./api.md). 

<span class="virtual-br"></span>

> NOTE: If your desired code host is not yet supported, please [open a feature request](https://github.com/sourcegraph/sourcegraph/issues/new?template=feature_request.md).

### Supported scale

If not otherwise stated in the table above, all code hosts should support up to 10k users and 100k repositories for permission syncing. 

These numbers come from testing the supported scale in a testing environment or running on a customer instance.

> NOTE: Sourcegraph might be able to support higher scale than specified, but was not rigorously tested to do so. 

Please contact support if you want to discuss bigger scale than specified.

## SLAs

Each method of getting permissions to Sourcegraph has different SLA on how long it takes for permissions to appear 
in Sourcegraph.

- [Permission syncing SLA](./syncing.md#sla)
- [Webhooks SLA](./webhooks.md#sla)
- [Explicit Permissions API SLA](./api.md#sla)
## License requirements

To have permission syncing available, the Sourcegraph instance needs to be configured with 
a license that has `acls` feature enabled. If it is not present, Sourcegraph will not enforce 
repository permissions and each repository will be treated as public - any user that has access 
to Sourcegraph will be able to access it.

## Site administrators

By default, site-admins bypass all of the permissions checks. Which means, all site admins are able to 
view all the repositories by default. The default can be changed by setting the site config 
option `authz.enforceForSiteAdmins` to `true`.

> NOTE: However, we recommend to be cautious with this option, as it might make some operations for site admins more complicated or impossible.

E.g. trying to figure out if a specific repository is syncing source code to Sourcegraph correctly 
might become an impossible task if the site admin cannot access that repository.
