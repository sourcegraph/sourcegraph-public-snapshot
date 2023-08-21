import { GraphQLError } from 'graphql'
import { trimEnd } from 'lodash'
import type { Observable } from 'rxjs'
import { fromFetch } from 'rxjs/fetch'
import type { Omit } from 'utility-types'

import { createAggregateError } from '@sourcegraph/common'

import { checkOk } from '../http-status-error'

import { GRAPHQL_URI } from './constants'

/**
 * Use this template string tag for all GraphQL queries.
 */
export const gql = (template: TemplateStringsArray, ...substitutions: any[]): string =>
    String.raw(template, ...substitutions)

export interface SuccessGraphQLResult<T> {
    data: T
    errors: undefined
}
export interface ErrorGraphQLResult<T = unknown> {
    // It might be possible that even with errored response we have
    // a partially resolved data
    // See https://github.com/sourcegraph/sourcegraph/pull/36033
    data: T | null
    errors: readonly GraphQLError[]
}

export type GraphQLResult<T> = SuccessGraphQLResult<T> | ErrorGraphQLResult<T>

/**
 * Guarantees that the GraphQL query resulted in an error.
 */
export function isErrorGraphQLResult<T>(result: GraphQLResult<T>): result is ErrorGraphQLResult<T> {
    return !!(result as ErrorGraphQLResult<T>).errors && (result as ErrorGraphQLResult<T>).errors.length > 0
}

export function dataOrThrowErrors<T>(result: GraphQLResult<T>): T {
    if (isErrorGraphQLResult(result)) {
        throw createAggregateError(result.errors)
    }
    return result.data
}

export const createInvalidGraphQLQueryResponseError = (queryName: string): GraphQLError =>
    new GraphQLError(`Invalid GraphQL response: query ${queryName}`)

export const createInvalidGraphQLMutationResponseError = (queryName: string): GraphQLError =>
    new GraphQLError(`Invalid GraphQL response: mutation ${queryName}`)

export interface GraphQLRequestOptions extends Omit<RequestInit, 'method' | 'body'> {
    baseUrl?: string
}

interface BuildGraphQLUrlOptions {
    request?: string
    baseUrl?: string
}
/**
 * Constructs GraphQL Request URL
 */
export const buildGraphQLUrl = ({ request, baseUrl }: BuildGraphQLUrlOptions): string => {
    const nameMatch = request ? request.match(/^\s*(?:query|mutation)\s+(\w+)/) : ''
    const apiURL = `${GRAPHQL_URI}${nameMatch ? '?' + nameMatch[1] : ''}`
    return baseUrl ? new URL(trimEnd(baseUrl, '/') + apiURL).href : apiURL
}

/**
 * This function should not be called directly as it does not
 * add the necessary headers to authorize the GraphQL API call.
 * Use `requestGraphQL()` in `client/web/src/backend/graphql.ts` instead.
 *
 * @deprecated Prefer using Apollo-Client instead if possible. The migration is in progress.
 */
export function requestGraphQLCommon<T, V = object>({
    request,
    baseUrl,
    variables,
    ...options
}: GraphQLRequestOptions & {
    request: string
    variables?: V
}): Observable<GraphQLResult<T>> {
    return fromFetch<GraphQLResult<T>>(buildGraphQLUrl({ request, baseUrl }), {
        ...options,
        method: 'POST',
        body: JSON.stringify({ query: request, variables }),
        selector: response => checkOk(response).json(),
    })
}

export * from './apollo'
