import { useConnection, UseConnectionResult } from 'src/components/FilteredConnection/hooks/useConnection'

import { dataOrThrowErrors } from '@sourcegraph/http-client'

import { LockfileIndexesResult, LockfileIndexesVariables, LockfileIndexFields } from '../../../../graphql-operations'

import { LOCKFILE_INDEXES_LIST } from './queries'

export const LOCKFILES_PER_PAGE_COUNT = 50

export const useLockfileIndexes = (first?: number, after?: string): UseConnectionResult<LockfileIndexFields> =>
    useConnection<LockfileIndexesResult, LockfileIndexesVariables, LockfileIndexFields>({
        query: LOCKFILE_INDEXES_LIST,
        variables: {
            first: first ?? null,
            after: after ?? null,
        },
        options: {
            useURL: false,
            fetchPolicy: 'cache-and-network',
        },
        getConnection: result => {
            const data = dataOrThrowErrors(result)

            return data.lockfileIndexes
        },
    })
