# Access control

<aside class="beta">
<p>
<span class="badge badge-beta">Beta</span> <strong>This feature is currently in beta.</strong>
</p>
</aside>

> NOTE: This page refers to in-product permissions, which determine who can, for example, create a batch change, or who is a site admin. This is *not* the same as [repository permissions](../permissions/index.md), which enforces the same repository access on Sourcegraph as your code host.

> NOTE: This functionality was added in Sourcegraph 5.0.

Sourcegraph uses [Role-Based Access Control (RBAC)](https://en.wikipedia.org/wiki/Role-based_access_control) to enable fine-grained control over different features and abilities of Sourcegraph, without having to modify permissions for each user individually. Currently, the scope of permissions control is limited to [Batch Changes](batch_changes.md) functionality, but it will be expanded to other areas in the future.

## Managing roles and permissions

<!-- TODO: Screenshots -->

Site admins can control which features each type of user has access to by creating custom roles and assigning permissions to them. You can see all available roles and create new ones under _Site admin_ > _Users & auth_ > _Roles_.

### System roles

Every Sourcegraph instance ships with two system roles:

- **Site Administrator**: This role is granted to any user who is promoted to site admin. It always has all features and permissions of Sourcegraph granted to it.
- **User:** This role is granted to every user of the Sourcegraph instance and cannot be unassigned. By default, it has all features and permissions of Sourcegraph granted to it, but this can be modified.

### Creating a new role and assigning it permissions

To create a new role, click the **+ Create role** button. Give the role a unique, descriptive name, then select which permissions to associate with it using the checkboxes. Then click **Create**.

### Editing permissions for an existing role

With the exception of the **Site Administrator** role, all role permissions can be modified. To edit the permissions granted to a role, click the role to expand it, then select the new set of permissions you want to grant to it. Then click **Update** to save your changes.

You can read about the specific permission types available for each RBAC-enabled product area below:

- [Batch Changes](./batch_changes.md)

> NOTE: While Batch Changes is the only RBAC-enabled product area today, we will be working on migrating other product areas in future releases of Sourcegraph. Please reach out to our [support team](mailto:support@sourcegraph.com) if you have further questions. 

### Deleting a role

With the exception of system roles, all roles can be deleted. To delete a role, click the **Delete** button on it. You will be prompted to confirm your choice. Once deleted, any user previously assigned that role will lose any permissions associated with it that are not also granted by their other roles.

## Managing user roles

Site admins can manage which roles are assigned to which users from _Site admin_ > _Users & auth_ > _Users_. To view or edit a user's roles, click the triple dots to open the context menu for that user, then click **Manage roles**. This will open a modal where you can see the user's current roles, assign new ones, or unassign current ones. You can type in the input field to search roles by name. Click **Update** to save any changes, or **Cancel** to discard. Note that system roles cannot be revoked or assigned via this modal.

To assign the **Site Administrator** system role to a user, open the same context menu from the triple dots, then click **Promote to site admin**. To unassign the **Site Administrator** role, open the same context menu from the triple dots, then click **Revoke site admin**.
