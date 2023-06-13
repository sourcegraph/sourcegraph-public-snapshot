# Permissions

> NOTE: This page refers to repository access permissions synced between Sourcegraph and your code host, which has also historically been referred to as "authorization", "repository permissions", and "code host permissions". This is *not* the same as [in-product permissions](../access_control/index.md), which determine who can, for example, create a batch change or who is a site admin.

Sourcegraph can be configured to enforce the same access to repositories and underlying 
source files as your code host. If configured, Sourcegraph will allow the user to only 
see the entities that they can see on the code host. These permissions are enforced 
across the product for all the use cases that need to read data from a repository, 
including the existence of such repository on the code host.

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

## Permissions mechanisms in parallel

<span class="badge badge-experimental">Experimental</span>
<span class="badge badge-note">Sourcegraph 5.0+</span>

Up to version 5.0 it was not possible to use explicit permissions API alongside permission syncing. 
Meaning, if explicit permissions API was turned ON, synced permissions were turned OFF. Which also meant it was 
impossible to use explicit permissions for one code host and synced permissions for another one on the same Sourcegraph instance. 

**Example**:

User `alice` has existing synced permissions to repositories `horsegraph/global` and `horsegraph/hay-v1`. 
Alice also has explicit API permissions to repository `horsegraph/hay-dev`. So the overall repository permissions 
of `alice` are the following union set: [`horsegraph/global`, `horsegraph/hay-v1`, `horsegraph/hay-dev`]

### Configuration

**Prerequisites:** 
1. Sourcegraph version 5.0+
1. Go to **Site Admin > Migrations** page. There is a migration called `Migrate data from user_permissions table to unified user_repo_permissions.`. 
Make sure that it finished migrating all the data (it reports as 100%). Contact support if the migration does not seem to complete for a long time (multiple days). 

1. Enable the experimental feature in the [site configuration](../config/site_config.md):
```json
{
  "experimentalFeatures": {
    "unifiedPermissions": "enabled"
  }
  // ...
}
```
1. Continue [configuring the explicit permissions API](api.md#configuration) as you would before. 
Both mechanisms work at the same time, thanks to a new behind the scenes data model that allows 
what was previously impossible. 

### Permission updates

Each permission mechanism is going to update only its own data. This means, that permission syncing is not 
going to touch permissions created by explicit permissions API and vice versa. We consider webhooks permissions 
as part of the permission syncing mechanism as well, since it is using the same underlying database operations. 

What the above paragraph means is, that when an updated set of accessible repositories for a user is given via 
permission sync, it will replace the existing set of synced permissions for that user, but not the explicit permissions.

**Example**:

Let's follow the example from above, `alice` has existing synced permissions to repositories `horsegraph/global` and `horsegraph/hay-v1` 
and explicit permissions to `horsegraph/hay-dev`, meaning a unioned set of effective permissions of [`horsegraph/global`, `horsegraph-hay-v1`, `horsegraph/hay-dev`]. 

An update comes in from permission sync, now returning `alice` permissions as [`horsegraph/global`, `horsegraph/hay-v2`]. Notice 
the removal of `horsegraph-v1` from the set.

After the update, the synced permissions of `alice` will be [`horsegraph/global`, `horsegraph/hay-v2`], but explicit permissions 
were not touched, leading to effective permissions of [`horsegraph/global`, `horsegraph-hay-v2`, `horsegraph/hay-dev`]
