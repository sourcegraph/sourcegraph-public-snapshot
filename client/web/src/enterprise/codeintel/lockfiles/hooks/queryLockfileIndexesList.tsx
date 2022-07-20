import { ApolloClient } from '@apollo/client'
import { from, Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { getDocumentNode } from '@sourcegraph/http-client'
import * as GQL from '@sourcegraph/shared/src/schema'

import {
    LockfileIndexConnectionFields,
    LockfileIndexesResult,
    LockfileIndexesVariables,
} from '../../../../graphql-operations'

import { LOCKFILE_INDEXES_LIST } from './queries'

export const queryLockfileIndexesList = (
    { first, after }: GQL.ILsifUploadsOnQueryArguments,
    client: ApolloClient<object>
): Observable<LockfileIndexConnectionFields> => {
    const variables: LockfileIndexesVariables = {
        first: first ?? null,
        after: after ?? null,
    }

    return from(
        client.query<LockfileIndexesResult, LockfileIndexesVariables>({
            query: getDocumentNode(LOCKFILE_INDEXES_LIST),
            variables: { ...variables },
        })
    ).pipe(map(({ data }) => data.lockfileIndexes))
}
