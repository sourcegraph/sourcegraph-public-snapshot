import type { ApolloClient } from '@apollo/client'
import { from, type Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { getDocumentNode, gql } from '@sourcegraph/http-client'

import type {
    PreciseIndexConnectionFields,
    PreciseIndexesResult,
    PreciseIndexesVariables,
    PreciseIndexState,
} from '../../../../graphql-operations'

import { preciseIndexFieldsFragment } from './types'

const PRECISE_INDEX_LIST = gql`
    query PreciseIndexes(
        $repo: ID
        $states: [PreciseIndexState!]
        $indexerKey: String
        $query: String
        $first: Int
        $after: String
    ) {
        preciseIndexes(
            repo: $repo
            states: $states
            query: $query
            first: $first
            after: $after
            indexerKey: $indexerKey
        ) {
            nodes {
                ...PreciseIndexFields
            }
            totalCount
            pageInfo {
                endCursor
                hasNextPage
            }
        }
    }

    ${preciseIndexFieldsFragment}
`

export const queryPreciseIndexes = (
    {
        repo,
        states,
        query,
        indexerKey,
        first,
        after,
    }: Partial<Omit<PreciseIndexesVariables, 'states'>> & { states?: string },
    client: ApolloClient<object>
): Observable<PreciseIndexConnectionFields> => {
    const typedStates = statesFromString(states)
    const variables: PreciseIndexesVariables = {
        repo: repo ?? null,
        states: typedStates.length > 0 ? typedStates : null,
        indexerKey: indexerKey ?? null,
        query: query ?? null,
        first: first ?? null,
        after: after ?? null,
    }

    return from(
        client.query<PreciseIndexesResult, PreciseIndexesVariables>({
            query: getDocumentNode(PRECISE_INDEX_LIST),
            variables: { ...variables },
        })
    ).pipe(
        map(({ data }) => data),
        map(({ preciseIndexes }) => preciseIndexes)
    )
}

export const statesFromString = (states?: string): PreciseIndexState[] =>
    (states || '')
        .split(',')
        .filter(state => state !== '')
        .map(state => state.toUpperCase() as PreciseIndexState)
