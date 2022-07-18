import { ApolloClient } from '@apollo/client'
import { from, Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { gql, getDocumentNode } from '@sourcegraph/http-client'
import * as GQL from '@sourcegraph/shared/src/schema'

import {
    LSIFUploadState,
    LsifUploadsVariables,
    LsifUploadsResult,
    LsifUploadConnectionFields,
} from '../../../../graphql-operations'

import { lsifUploadFieldsFragment } from './types'

const LSIF_UPLOAD_LIST = gql`
    query LsifUploads(
        $state: LSIFUploadState
        $isLatestForRepo: Boolean
        $dependencyOf: ID
        $dependentOf: ID
        $first: Int
        $after: String
        $query: String
        $includeDeleted: Boolean
    ) {
        lsifUploads(
            query: $query
            state: $state
            isLatestForRepo: $isLatestForRepo
            dependencyOf: $dependencyOf
            dependentOf: $dependentOf
            first: $first
            after: $after
            includeDeleted: $includeDeleted
        ) {
            nodes {
                ...LsifUploadFields
            }
            totalCount
            pageInfo {
                endCursor
                hasNextPage
            }
        }
    }

    ${lsifUploadFieldsFragment}
`

export interface UploadListVariables {
    state?: LSIFUploadState
    isLatestForRepo?: boolean
    dependencyOf?: string | null
    dependentOf?: string | null
    first?: number | null
    after?: string | null
    query?: string | null
    includeDeleted?: boolean | null
}

export const queryLsifUploadsList = (
    {
        query,
        state,
        isLatestForRepo,
        dependencyOf,
        dependentOf,
        first,
        after,
        includeDeleted,
    }: GQL.ILsifUploadsOnQueryArguments,
    client: ApolloClient<object>
): Observable<LsifUploadConnectionFields> => {
    const variables: LsifUploadsVariables = {
        query: query ?? null,
        state: state ?? null,
        isLatestForRepo: isLatestForRepo ?? null,
        dependencyOf: dependencyOf ?? null,
        dependentOf: dependentOf ?? null,
        first: first ?? null,
        after: after ?? null,
        includeDeleted: includeDeleted ?? null,
    }

    return from(
        client.query<LsifUploadsResult, LsifUploadsVariables>({
            query: getDocumentNode(LSIF_UPLOAD_LIST),
            variables: { ...variables },
        })
    ).pipe(
        map(({ data }) => data),
        map(({ lsifUploads }) => lsifUploads)
    )
}
