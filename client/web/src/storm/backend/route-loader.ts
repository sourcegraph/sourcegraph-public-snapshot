import type { OperationVariables, SuspenseQueryHookOptions } from '@apollo/client'
import { useLoaderData } from 'react-router-dom'
import * as uuid from 'uuid'

import { getDocumentNode, useSuspenseQuery } from '@sourcegraph/http-client'

import { getWebGraphQLClient } from '../../backend/graphql'

export type LoaderQuery = Parameters<typeof useSuspenseQuery>[0]
export interface QueryReference extends SuspenseQueryHookOptions {}

/**
 * TODO: Add explainers on how these things are connected.
 */
function createReference(variables: OperationVariables): QueryReference {
    return {
        variables,
        fetchPolicy: 'cache-first',
    }
}

async function loadQuery(query: LoaderQuery, variables: OperationVariables): Promise<QueryReference> {
    // https://www.apollographql.com/docs/react/performance/performance/
    //
    // We actually don't need to await here because the query will still
    // be started eagerly and we can instead wait for the suspense in the
    // component. This would be great because we can already start some
    // rendering work in the component.
    //
    // However it seems like suspending is not properly handled in that
    // case as the suspense boundaries are triggered. Perhaps an issue
    // in the useSuspenseQuery_experimental implementation? It may not
    // use startTransition in this case.
    await getWebGraphQLClient().then(client => client.query({ query: getDocumentNode(query), variables }))

    return createReference(variables)
}

interface CreatePreloadedQueryResult<D extends object, V extends OperationVariables> {
    queryLoader: (variables: V) => Promise<Record<string, QueryReference | undefined>>
    usePreloadedQueryData: () => ReturnType<typeof useSuspenseQuery<D, V>>
}

export function createPreloadedQuery<D extends object, V extends OperationVariables>(
    query: LoaderQuery
): CreatePreloadedQueryResult<D, V> {
    const key = uuid.v4()

    return {
        queryLoader: async variables => {
            const reference = await loadQuery(query, variables)

            return { [key]: reference }
        },
        usePreloadedQueryData: () => {
            const loaderData = useLoaderData() as Record<string, V | undefined>

            const reference = loaderData[key]
            if (!reference) {
                throw new Error('loadQuery() was not properly called in the loader function')
            }

            return useSuspenseQuery<D, V>(query, reference)
        },
    }
}
