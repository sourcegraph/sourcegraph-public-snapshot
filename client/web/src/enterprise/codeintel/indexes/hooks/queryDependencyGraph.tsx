import type { ApolloClient } from '@apollo/client'
import { from, type Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { getDocumentNode, gql } from '@sourcegraph/http-client'

import type {
    PreciseIndexConnectionFields,
    PreciseIndexDependencyGraphResult,
    PreciseIndexDependencyGraphVariables,
} from '../../../../graphql-operations'

import { preciseIndexFieldsFragment } from './types'

const PRECISE_INDEX_DEPENDENCY_GRAPH = gql`
    query PreciseIndexDependencyGraph(
        $dependencyOf: ID
        $dependentOf: ID
        $query: String
        $first: Int
        $after: String
    ) {
        preciseIndexes(
            dependencyOf: $dependencyOf
            dependentOf: $dependentOf
            query: $query
            first: $first
            after: $after
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

export const queryDependencyGraph = (
    { dependencyOf, dependentOf, query, first, after }: Partial<PreciseIndexDependencyGraphVariables>,
    client: ApolloClient<object>
): Observable<PreciseIndexConnectionFields> => {
    const variables: PreciseIndexDependencyGraphVariables = {
        dependencyOf: dependencyOf ?? null,
        dependentOf: dependentOf ?? null,
        query: query ?? null,
        first: first ?? null,
        after: after ?? null,
    }

    return from(
        client.query<PreciseIndexDependencyGraphResult, PreciseIndexDependencyGraphVariables>({
            query: getDocumentNode(PRECISE_INDEX_DEPENDENCY_GRAPH),
            variables: { ...variables },
            fetchPolicy: 'cache-first',
        })
    ).pipe(
        map(({ data }) => data),
        map(({ preciseIndexes }) => preciseIndexes)
    )
}
