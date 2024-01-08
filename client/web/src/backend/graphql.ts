import { memoize } from 'lodash'
import type { Observable } from 'rxjs'

import { getGraphQLClient, type GraphQLResult, requestGraphQLCommon } from '@sourcegraph/http-client'

import { getFeatureFlagOverrides } from '../featureFlags/lib/feature-flag-local-overrides'
import type { WebGraphQlOperations } from '../graphql-operations'
import { useDeveloperSettings } from '../stores'

import { getPersistentCache } from './getPersistentCache'

const getHeaders = (): { [header: string]: string } => {
    const headers: { [header: string]: string } = {
        ...window?.context?.xhrHeaders,
        Accept: 'application/json',
        'Content-Type': 'application/json',
    }
    const parameters = new URLSearchParams(window.location.search)
    const trace = parameters.get('trace')
    if (trace) {
        headers['X-Sourcegraph-Should-Trace'] = trace
    }

    // Get values from URL and local overrides
    let feat = parameters.getAll('feat')
    if (process.env.NODE_ENV === 'development' || useDeveloperSettings.getState().enabled) {
        // Use Set to dedupe values from the URL and local. At some point during page
        // rendering the overrides from the URL will be synced to local.
        // It's not necessary to dedupe them (duplicate values are not a problem),
        // but it's less surprising when inspecting requests.
        feat = Array.from(
            new Set(feat.concat(Array.from(getFeatureFlagOverrides(), ([flag, value]) => (value ? '' : '-') + flag)))
        )
    }

    if (feat.length) {
        headers['X-Sourcegraph-Override-Feature'] = feat.join(',')
    }
    return headers
}

/**
 * Does a GraphQL request to the Sourcegraph GraphQL API running under `/.api/graphql`
 * @param request The GraphQL request (query or mutation)
 * @param variables A key/value object with variable values
 * @returns Observable That emits the result or errors if the HTTP request failed
 * @template TResult The type of the query result (import from our auto-generated types).
 * @template TVariables The type of the query input variables (import from our auto-generated types).
 * @deprecated Prefer using Apollo-Client instead if possible. The migration is in progress.
 */
export const requestGraphQL = <TResult, TVariables = object>(
    request: string,
    variables?: TVariables
): Observable<GraphQLResult<TResult>> =>
    requestGraphQLCommon({
        request,
        variables,
        headers: getHeaders(),
    })

type WebGraphQlOperationResults = ReturnType<WebGraphQlOperations[keyof WebGraphQlOperations]>

/**
 * Does a GraphQL query to the Sourcegraph GraphQL API running under `/.api/graphql`
 * @param request The GraphQL query
 * @param variables A key/value object with variable values
 * @returns Observable That emits the result or errors if the HTTP request failed
 * @deprecated Prefer using Apollo-Client instead if possible. The migration is in progress.
 */
export const queryGraphQL = <TResult extends WebGraphQlOperationResults>(
    request: string,
    variables?: {}
): Observable<GraphQLResult<TResult>> =>
    requestGraphQLCommon<TResult>({
        request,
        variables,
        headers: getHeaders(),
    })

/**
 * Does a GraphQL mutation to the Sourcegraph GraphQL API running under `/.api/graphql`
 * @param request The GraphQL mutation
 * @param variables A key/value object with variable values
 * @returns Observable That emits the result or errors if the HTTP request failed
 * @deprecated Prefer using Apollo-Client instead if possible. The migration is in progress.
 */
export const mutateGraphQL = <TResult extends WebGraphQlOperationResults>(
    request: string,
    variables?: {}
): Observable<GraphQLResult<TResult>> =>
    requestGraphQLCommon<TResult>({
        request,
        variables,
        headers: getHeaders(),
    })

/**
 * Memoized Apollo Client getter. It should be executed once to restore the cache from the local storage.
 * After that, the same instance should be used by all consumers.
 */
export const getWebGraphQLClient = memoize(async () => {
    const persistentCache = await getPersistentCache({
        isAuthenticatedUser: window.context.isAuthenticatedUser,
        preloadedQueries: {
            temporarySettings: window.context.temporarySettings,
        },
    })

    const client = await getGraphQLClient({
        cache: persistentCache,
        getHeaders,
    })

    return client
})
