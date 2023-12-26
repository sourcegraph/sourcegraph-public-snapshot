import { findIndex } from 'lodash'

import type { AuthenticatedUser } from '../auth'

export const doesUserHavePermission = (user: AuthenticatedUser | null, permissionToCheckFor: string): boolean => {
    if (user === null) {
        return false
    }

    return findIndex(user.permissions.nodes, permission => permission.displayName === permissionToCheckFor) !== -1
}
