import { ApolloClient } from '@apollo/client'
import { from, Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { gql, getDocumentNode } from '@sourcegraph/http-client'

import { LsifIndexesForRepoResult, LsifIndexesForRepoVariables, LsifIndexFields } from '../../../../graphql-operations'

import { lsifIndexFieldsFragment } from './types'

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
    { query, state, first, after }: Partial<LsifIndexesForRepoVariables>,
    repository: string,
    client: ApolloClient<object>
): Observable<IndexConnection> => {
    const variables: LsifIndexesForRepoVariables = {
        repository,
        query: query ?? null,
        state: state ?? null,
        first: first ?? null,
        after: after ?? null,
    }

    return from(
        client.query<LsifIndexesForRepoResult, LsifIndexesForRepoVariables>({
            query: getDocumentNode(LSIF_INDEX_FOR_REPOSITORY),
            variables: { ...variables },
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
