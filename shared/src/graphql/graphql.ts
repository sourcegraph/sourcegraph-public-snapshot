import { Observable } from 'rxjs'
import { ajax, AjaxRequest, AjaxResponse } from 'rxjs/ajax'
import { catchError, map } from 'rxjs/operators'
import { Omit } from 'utility-types'
import { createAggregateError, normalizeAjaxError } from '../util/errors'
import * as GQL from './schema'

/**
 * Use this template string tag for all GraphQL queries.
 */
export const gql = (template: TemplateStringsArray, ...substitutions: any[]): string =>
    String.raw(template, ...substitutions)

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

export interface GraphQLRequestOptions {
    headers: AjaxRequest['headers']
    requestOptions?: Partial<Omit<AjaxRequest, 'url' | 'method' | 'headers' | 'body'>>
    baseUrl?: string
}

export function requestGraphQL<T extends GQL.IQuery | GQL.IMutation>({
    request,
    variables = {},
    headers,
    requestOptions = {},
    baseUrl = '',
}: GraphQLRequestOptions & {
    request: string
    variables?: {}
}): Observable<GraphQLResult<T>> {
    const nameMatch = request.match(/^\s*(?:query|mutation)\s+(\w+)/)
    return ajax({
        method: 'POST',
        url: `${baseUrl}/.api/graphql${nameMatch ? '?' + nameMatch[1] : ''}`,
        headers,
        body: JSON.stringify({ query: request, variables }),
        ...requestOptions,
    }).pipe(
        catchError<AjaxResponse, never>(err => {
            normalizeAjaxError(err)
            throw err
        }),
        map(({ response }) => response)
    )
}

/**
 * It is cumbersome to query a GraphQL union whose types share common fields, because you must
 * repeat the fields for each branch, as in `... on UnionType1 { a b }`, `... on UnionType2 { a b}`,
 * etc.
 *
 * This helper produces the query and fragment definitions to make this easier.
 *
 * @param types GraphQL type names of the union's types.
 * @param fields The set of GraphQL fields that are common to all of the types.
 * @param nestedFields Other fields to query that are not immediate child fields (and thus can't be
 * typechecked).
 */
export function queryAndFragmentForUnion<T extends string, F extends string>(
    typeNames: T[],
    fields: F[],
    nestedFields?: string[],
    extraFragments?: string[]
): { query: string; fragment: string } {
    const allFields = nestedFields ? [...fields, ...nestedFields] : fields
    const fragment = [
        ...typeNames.map(typeName => `fragment ${typeName}Fields on ${typeName} { ${allFields.join('\n')} }`),
        ...(extraFragments || []),
    ].join('\n')
    const query = gql`
__typename
${typeNames.map(typeName => `... on ${typeName} { ...${typeName}Fields }`).join('\n')}
`
    return { fragment, query }
}
