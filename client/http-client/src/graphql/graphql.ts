import { gql as apolloGql } from '@apollo/client'
import { GraphQLError, TypedQueryDocumentNode } from 'graphql'
import { trimEnd } from 'lodash'
import { Observable } from 'rxjs'
import { fromFetch } from 'rxjs/fetch'
import { Omit } from 'utility-types'

import { createAggregateError } from '@sourcegraph/common'

import { checkOk } from '../http-status-error'

import { GRAPHQL_URI } from './constants'

/**
 * Use this template string tag for all GraphQL queries.
 */
export const gql = apolloGql

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
export type GraphQLRequest<T, V> = string | TypedQueryDocumentNode<T, V>

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

interface BuildGraphQLUrlOptions<T, V> {
    request?: GraphQLRequest<T, V>
    baseUrl?: string
}
/**
 * Constructs GraphQL Request URL
 */
export function buildGraphQLUrl<T, V = object>({ request, baseUrl }: BuildGraphQLUrlOptions<T, V>): string {
    let nameMatch: string | undefined

    if (typeof request === 'string') {
        nameMatch = request.match(/^\s*(?:query|mutation)\s+(\w+)/)?.[1]
    } else {
        // TODO: Update
        nameMatch = undefined
    }

    const apiURL = `${GRAPHQL_URI}${nameMatch ? '?' + nameMatch : ''}`
    return baseUrl ? new URL(trimEnd(baseUrl, '/') + apiURL).href : apiURL
}

/**
 * This function should not be called directly as it does not
 * add the necessary headers to authorize the GraphQL API call.
 * Use `requestGraphQL()` in `client/web/src/backend/graphql.ts` instead.
 */
export function requestGraphQLCommon<T, V = object>({
    request,
    baseUrl,
    variables,
    ...options
}: GraphQLRequestOptions & {
    request: GraphQLRequest<T, V>
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
