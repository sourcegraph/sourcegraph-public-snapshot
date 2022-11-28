import { memoize } from 'lodash'
import { Observable } from 'rxjs'

import {
    buildGraphQLUrl,
    checkOk,
    getGraphQLClient,
    GraphQLResult,
    requestGraphQLCommon,
} from '@sourcegraph/http-client'
import * as GQL from '@sourcegraph/shared/src/schema'

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
    const feat = parameters.getAll('feat')
    if (feat.length) {
        headers['X-Sourcegraph-Override-Feature'] = feat.join(',')
    }
    return headers
}

/**
 * Does a GraphQL request to the Sourcegraph GraphQL API running under `/.api/graphql`
 *
 * @param request The GraphQL request (query or mutation)
 * @param variables A key/value object with variable values
 * @returns Observable That emits the result or errors if the HTTP request failed
 * @template TResult The type of the query result (import from our auto-generated types).
 * @template TVariables The type of the query input variables (import from our auto-generated types).
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

export const requestGraphQLWithProgress = <TResult, TVariables = object>(
    request: string,
    variables?: TVariables,
    onProgress?: (loaded: number, total: number) => void
): Observable<GraphQLResult<TResult>> =>
    new Observable<GraphQLResult<TResult>>(async observer => {
        console.log('A')
        const response = checkOk(
            await fetch(buildGraphQLUrl({ request }), {
                method: 'POST',
                headers: getHeaders(),
                body: JSON.stringify({ query: request, variables }),
            })
        )

        console.log('B')

        if (!response.body) {
            throw new Error('ReadableStream not yet supported in this browser.')
        }
        const body = response.body

        // to access headers, server must send CORS header "Access-Control-Expose-Headers: content-encoding, content-length x-file-size"
        // server must send custom x-file-size header if gzip or other content-encoding is used
        const contentEncoding = response.headers.get('content-encoding')

        console.log({ contentEncoding })
        const contentLength = response.headers.get(contentEncoding ? 'x-file-size' : 'content-length')
        if (contentLength === null) {
            throw new Error('Response size header unavailable')
        }

        const total = parseInt(contentLength, 10)
        let loaded = 0
        const streamingResponse = new Response(
            new ReadableStream({
                start(controller) {
                    const reader = body.getReader()
                    read()
                    function read() {
                        console.log('C')
                        reader
                            .read()
                            .then(({ done, value }) => {
                                if (done) {
                                    controller.close()
                                    return
                                }
                                loaded += value.byteLength
                                onProgress?.(loaded, total)
                                controller.enqueue(value)
                                read()
                            })
                            .catch(error => {
                                console.error(error)
                                controller.error(error)
                            })
                        return undefined
                    }
                },
            })
        )
        const data: GraphQLResult<TResult> = await streamingResponse.json()
        console.log({ data })
        observer.next(data)
        //     method: 'POST',
        //     body: JSON.stringify({ query: request, variables }),
        //     selector: response => checkOk(response).json(),
        // })
    })

/**
 * Does a GraphQL query to the Sourcegraph GraphQL API running under `/.api/graphql`
 *
 * @param request The GraphQL query
 * @param variables A key/value object with variable values
 * @returns Observable That emits the result or errors if the HTTP request failed
 *
 * @deprecated Prefer using `requestGraphQL()` and passing auto-generated query types as type parameters.
 */
export const queryGraphQL = (request: string, variables?: {}): Observable<GraphQLResult<GQL.IQuery>> =>
    requestGraphQLCommon<GQL.IQuery>({
        request,
        variables,
        headers: getHeaders(),
    })

/**
 * Does a GraphQL mutation to the Sourcegraph GraphQL API running under `/.api/graphql`
 *
 * @param request The GraphQL mutation
 * @param variables A key/value object with variable values
 * @returns Observable That emits the result or errors if the HTTP request failed
 *
 * @deprecated Prefer using `requestGraphQL()` and passing auto-generated query types as type parameters.
 */
export const mutateGraphQL = (request: string, variables?: {}): Observable<GraphQLResult<GQL.IMutation>> =>
    requestGraphQLCommon<GQL.IMutation>({
        request,
        variables,
        headers: getHeaders(),
    })

/**
 * Memoized Apollo Client getter. It should be executed once to restore the cache from the local storage.
 * After that, the same instance should be used by all consumers.
 */
export const getWebGraphQLClient = memoize(() =>
    getGraphQLClient({
        isAuthenticated: window.context.isAuthenticatedUser,
        headers: getHeaders(),
    })
)
