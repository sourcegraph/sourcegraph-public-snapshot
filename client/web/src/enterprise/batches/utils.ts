import { findIndex } from 'lodash'

import { AuthenticatedUser } from '../../auth'
import {
    ChangesetCheckState,
    ChangesetReviewState,
    ChangesetSpecOperation,
    ChangesetState,
} from '../../graphql-operations'

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
    !!user && findIndex(user.permissions.nodes, permission => permission.displayName === 'BATCH_CHANGES#WRITE') !== -1
