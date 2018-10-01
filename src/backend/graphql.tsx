import { Observable } from 'rxjs'
import { ajax, AjaxResponse } from 'rxjs/ajax'
import { catchError, map } from 'rxjs/operators'
import { createAggregateError, normalizeAjaxError } from '../util/errors'
import * as GQL from './graphqlschema'

export const graphQLContent = Symbol('graphQLContent')
export interface GraphQLDocument {
    [graphQLContent]: string
}

/**
 * Guarantees that the GraphQL query resulted in an error.
 */
export function isGraphQLError<T extends GQL.IQuery | GQL.IMutation>(
    result: GraphQLResult<T>
): result is ErrorGraphQLResult {
    return !!(result as ErrorGraphQLResult).errors && (result as ErrorGraphQLResult).errors.length > 0
}

export function dataOrThrowErrors<T extends GQL.IQuery | GQL.IMutation>(result: GraphQLResult<T>): T {
    if (isGraphQLError(result)) {
        throw createAggregateError(result.errors)
    }
    return result.data
}

export interface GraphQLError extends Error {
    queryName: string
}
export const createInvalidGraphQLQueryResponseError = (queryName: string): GraphQLError =>
    Object.assign(new Error(`Invalid GraphQL response: query ${queryName}`), {
        queryName,
    })
export const createInvalidGraphQLMutationResponseError = (queryName: string): GraphQLError =>
    Object.assign(new Error(`Invalid GraphQL response: mutation ${queryName}`), {
        queryName,
    })

/**
 * Use this template string tag for all GraphQL queries
 */
export const gql = (template: TemplateStringsArray, ...substitutions: any[]): GraphQLDocument => ({
    [graphQLContent]: String.raw(template, ...substitutions.map(s => s[graphQLContent] || s)),
})

export interface SuccessGraphQLResult<T extends GQL.IQuery | GQL.IMutation> {
    data: T
    errors: undefined
}
export interface ErrorGraphQLResult {
    data: undefined
    errors: GQL.IGraphQLResponseError[]
}

export type GraphQLResult<T extends GQL.IQuery | GQL.IMutation> = SuccessGraphQLResult<T> | ErrorGraphQLResult

/**
 * Does a GraphQL request to the Sourcegraph GraphQL API running under `/.api/graphql`
 *
 * @param request The GraphQL request (query or mutation)
 * @param variables A key/value object with variable values
 * @return Observable That emits the result or errors if the HTTP request failed
 */
function requestGraphQL(request: GraphQLDocument, variables: any = {}): Observable<GQL.IGraphQLResponseRoot> {
    const nameMatch = request[graphQLContent].match(/^\s*(?:query|mutation)\s+(\w+)/)
    return ajax({
        method: 'POST',
        url: '/.api/graphql' + (nameMatch ? '?' + nameMatch[1] : ''),
        headers: {
            ...window.context.xhrHeaders,
            Accept: 'application/json',
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({ query: request[graphQLContent], variables }),
    }).pipe(
        catchError<AjaxResponse, never>(err => {
            normalizeAjaxError(err)
            throw err
        }),
        map(({ response }) => response)
    )
}

/**
 * Does a GraphQL query to the Sourcegraph GraphQL API running under `/.api/graphql`
 *
 * @param query The GraphQL query
 * @param variables A key/value object with variable values
 * @return Observable That emits the result or errors if the HTTP request failed
 */
export function queryGraphQL(query: GraphQLDocument, variables: any = {}): Observable<GraphQLResult<GQL.IQuery>> {
    return requestGraphQL(query, variables) as Observable<GraphQLResult<GQL.IQuery>>
}

/**
 * Does a GraphQL mutation to the Sourcegraph GraphQL API running under `/.api/graphql`
 *
 * @param mutation The GraphQL mutation
 * @param variables A key/value object with variable values
 * @return Observable That emits the result or errors if the HTTP request failed
 */
export function mutateGraphQL(
    mutation: GraphQLDocument,
    variables: any = {}
): Observable<GraphQLResult<GQL.IMutation>> {
    return requestGraphQL(mutation, variables) as Observable<GraphQLResult<GQL.IMutation>>
}
