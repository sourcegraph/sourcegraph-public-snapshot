import { findIndex } from 'lodash'

import type { AuthenticatedUser } from '../auth'
import type { RbacPermission } from '../rbac/constants'

export const doesUserHavePermission = (
    user: Pick<AuthenticatedUser, 'permissions'> | null,
    permissionToCheckFor: RbacPermission
): boolean => {
    if (user === null) {
        return false
    }

    return findIndex(user.permissions.nodes, permission => permission.displayName === permissionToCheckFor) !== -1
}
