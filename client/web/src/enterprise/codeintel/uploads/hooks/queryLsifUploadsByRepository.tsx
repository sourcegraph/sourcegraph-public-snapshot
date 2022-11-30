import { ApolloClient } from '@apollo/client'
import { from, Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { gql, getDocumentNode } from '@sourcegraph/http-client'

import {
    LSIFUploadState,
    LsifUploadsForRepoResult,
    LsifUploadsForRepoVariables,
    LsifUploadConnectionFields,
    LsifUploadsVariables,
} from '../../../../graphql-operations'

import { lsifUploadConnectionFieldsFragment } from './types'

const LSIF_UPLOAD_LIST_BY_REPO_ID = gql`
    query LsifUploadsForRepo(
        $repository: ID!
        $state: LSIFUploadState
        $isLatestForRepo: Boolean
        $dependencyOf: ID
        $dependentOf: ID
        $first: Int
        $after: String
        $query: String
    ) {
        node(id: $repository) {
            __typename
            ... on Repository {
                lsifUploads(
                    query: $query
                    state: $state
                    isLatestForRepo: $isLatestForRepo
                    dependencyOf: $dependencyOf
                    dependentOf: $dependentOf
                    first: $first
                    after: $after
                ) {
                    ...LsifUploadConnectionFields
                }
            }
        }
    }

    ${lsifUploadConnectionFieldsFragment}
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

export const queryLsifUploadsByRepository = (
    {
        query,
        state,
        isLatestForRepo,
        dependencyOf,
        dependentOf,
        first,
        after,
        includeDeleted,
    }: Partial<LsifUploadsVariables>,
    repository: string,
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
        client.query<LsifUploadsForRepoResult, LsifUploadsForRepoVariables>({
            query: getDocumentNode(LSIF_UPLOAD_LIST_BY_REPO_ID),
            variables: { ...variables, repository },
        })
    ).pipe(
        map(({ data }) => data),
        map(({ node }) => {
            if (!node) {
                throw new Error('Invalid repository')
            }
            if (node.__typename !== 'Repository') {
                throw new Error(`The given ID is ${node.__typename}, not Repository`)
            }

            return node.lsifUploads
        })
    )
}
