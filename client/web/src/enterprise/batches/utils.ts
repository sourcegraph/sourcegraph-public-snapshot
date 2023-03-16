import { findIndex } from 'lodash'

import { AuthenticatedUser } from '../../auth'
import {
    ChangesetCheckState,
    ChangesetReviewState,
    ChangesetSpecOperation,
    ChangesetState,
} from '../../graphql-operations'
import { BatchChangesWritePermission } from '../../rbac/constants'

export function isValidChangesetReviewState(input: string): input is ChangesetReviewState {
    return Object.values<string>(ChangesetReviewState).includes(input)
}

export function isValidChangesetCheckState(input: string): input is ChangesetCheckState {
    return Object.values<string>(ChangesetCheckState).includes(input)
}

export function isValidChangesetSpecOperation(input: string): input is ChangesetSpecOperation {
    return Object.values<string>(ChangesetSpecOperation).includes(input)
}

export function isValidChangesetState(input: string): input is ChangesetState {
    return Object.values<string>(ChangesetState).includes(input)
}

export const canWriteBatchChanges = (user: Pick<AuthenticatedUser, 'permissions'> | null): boolean =>
    !!user && findIndex(user.permissions.nodes, permission => permission.displayName === BatchChangesWritePermission) !== -1

export const NO_ACCESS_SOURCEGRAPH_COM = 'Batch changes are not available on Sourcegraph.com.'
export const NO_ACCESS_BATCH_CHANGES_WRITE =
    'Your user does not have sufficient permissions to create batch changes. Contact your site admin to request access.'
export const NO_ACCESS_NAMESPACE = 'Your user is not able to create batch changes in this namespace.'
