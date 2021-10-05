import { ApolloError, MutationFunctionOptions, FetchResult, ApolloClient, useMutation } from '@apollo/client'
import { from, Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { gql, getDocumentNode } from '@sourcegraph/shared/src/graphql/graphql'
import * as GQL from '@sourcegraph/shared/src/graphql/schema'

import {
    LsifIndexesForRepoResult,
    LsifIndexesForRepoVariables,
    LsifIndexesResult,
    LsifIndexesVariables,
    LsifIndexFields,
    CodeIntelligenceCommitGraphMetadataResult,
    CodeIntelligenceCommitGraphMetadataVariables,
    QueueAutoIndexJobsForRepoResult,
    QueueAutoIndexJobsForRepoVariables,
} from '../../../graphql-operations'
import { lsifIndexFieldsFragment } from '../shared/backend'

interface IndexConnection {
    nodes: LsifIndexFields[]
    totalCount: number | null
    pageInfo: { endCursor: string | null; hasNextPage: boolean }
}

const LSIF_INDEX_FOR_REPOSITORY = gql`
    query LsifIndexesForRepo($repository: ID!, $state: LSIFIndexState, $first: Int, $after: String, $query: String) {
        node(id: $repository) {
            __typename
            ... on Repository {
                lsifIndexes(query: $query, state: $state, first: $first, after: $after) {
                    nodes {
                        ...LsifIndexFields
                    }
                    totalCount
                    pageInfo {
                        endCursor
                        hasNextPage
                    }
                }
            }
        }
    }

    ${lsifIndexFieldsFragment}
`

export const queryLsifIndexListByRepository = (
    { query, state, first, after }: GQL.ILsifIndexesOnRepositoryArguments,
    repository: string,
    client: ApolloClient<object>
): Observable<IndexConnection> => {
    const vars = {
        repository,
        query: query ?? null,
        state: state ?? null,
        first: first ?? null,
        after: after ?? null,
    }

    return from(
        client.query<LsifIndexesForRepoResult, LsifIndexesForRepoVariables>({
            query: getDocumentNode(LSIF_INDEX_FOR_REPOSITORY),
            variables: { ...vars },
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

            return node.lsifIndexes
        })
    )
}

const LSIF_INDEXES = gql`
    query LsifIndexes($state: LSIFIndexState, $first: Int, $after: String, $query: String) {
        lsifIndexes(query: $query, state: $state, first: $first, after: $after) {
            nodes {
                ...LsifIndexFields
            }
            totalCount
            pageInfo {
                endCursor
                hasNextPage
            }
        }
    }

    ${lsifIndexFieldsFragment}
`

export const queryLsifIndexList = (
    { query, state, first, after }: GQL.ILsifIndexesOnRepositoryArguments,
    client: ApolloClient<object>
): Observable<IndexConnection> => {
    const vars = {
        query: query ?? null,
        state: state ?? null,
        first: first ?? null,
        after: after ?? null,
    }

    return from(
        client.query<LsifIndexesResult, LsifIndexesVariables>({
            query: getDocumentNode(LSIF_INDEXES),
            variables: { ...vars },
        })
    ).pipe(
        map(({ data }) => data),
        map(({ lsifIndexes }) => lsifIndexes)
    )
}
