import type { AuthenticatedUser } from '../auth'
import { doesUserHavePermission } from '../rbac/checker'
import { BatchChangesWritePermission } from '../rbac/constants'

export const canWriteBatchChanges = (user: AuthenticatedUser | null): boolean =>
    doesUserHavePermission(user, BatchChangesWritePermission)

export const NO_ACCESS_SOURCEGRAPH_COM = 'Batch changes are not available on Sourcegraph.com.'
export const NO_ACCESS_BATCH_CHANGES_WRITE =
    'Your user does not have sufficient permissions to create batch changes. Contact your site admin to request access.'
export const NO_ACCESS_NAMESPACE = 'Your user is not able to create batch changes in this namespace.'
