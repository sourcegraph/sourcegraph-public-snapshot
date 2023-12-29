import { findIndex } from 'lodash'

import type { AuthenticatedUser } from '@sourcegraph/shared/src/auth'

import type { RbacPermission } from './constants'

export const doesUserHavePermission = (
    user: AuthenticatedUser | null,
    permissionToCheckFor: RbacPermission
): boolean => {
    if (user === null) {
        return false
    }

    return findIndex(user.permissions.nodes, permission => permission.displayName === permissionToCheckFor) !== -1
}
