import { useLoaderData } from 'react-router-dom'
import * as uuid from 'uuid'

import { getDocumentNode, gql, useSuspenseQuery } from '@sourcegraph/http-client'
import { getWebGraphQLClient } from '../../../backend/graphql'

import { SearchPageQueryResult, SearchPageQueryVariables } from '../../../graphql-operations'

/**
 * TODO: Create this GraphQL query combining fragments specified in underliying components
 * instead of hardcoding it here.
 *
 * TODO: We need the `evaluateFeatureFlags` query that can evaluate multiple feature flags.
 */
const SEARCH_PAGE_QUERY = gql`
    query SearchPageQuery($flagName: String!) {
        externalServices {
            totalCount
        }
        evaluateFeatureFlag(flagName: $flagName) {
            id
            name
            value
        }
    }
`

export const { pageLoader, usePageLoaderData } = createPreloadedQuery<SearchPageQueryResult, SearchPageQueryVariables>(
    SEARCH_PAGE_QUERY
)

export async function loader() {
    console.log('SearchPage loader is called')

    return pageLoader({ flagName: 'plg-enable-add-codehost-widget' })
}

/**
 * TODO: Add explainers on how these things are connected and type properly.
 */
function createReference(variables: object) {
    return {
        variables,
    }
}

export async function loadQuery(query, variables) {
    // https://www.apollographql.com/docs/react/performance/performance/
    //
    // We actually don't need to await here because the query will still
    // be started eagerly and we can instead wait for the suspense in the
    // component. This would be great because we can already start some
    // rendering work in the component.
    //
    // However it seems like suspending is not properly handled in that
    // case as the supsense boundaries are triggered. Perhaps an issue
    // in the useSuspenseQuery_experimental implementation? It may not
    // use startTransition in this case.
    await getWebGraphQLClient().then(client => client.query({ query: getDocumentNode(query), variables }))

    return createReference(variables)
}

export function createPreloadedQuery<D extends object, V extends object>(query) {
    const key = uuid.v4()

    async function pageLoader(variables: V) {
        const reference = await loadQuery(query, variables)
        return { [key]: reference }
    }

    function usePageLoaderData() {
        const loaderData = useLoaderData()

        const reference = loaderData[key]
        if (!reference) {
            throw new Error('loadQuery() was not properly called in the loader function')
        }

        return useSuspenseQuery<D>(query, reference)
    }

    return { pageLoader, usePageLoaderData }
}
