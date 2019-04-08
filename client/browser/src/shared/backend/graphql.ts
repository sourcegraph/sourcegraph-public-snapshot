import { from, Observable, throwError } from 'rxjs'
import { ajax } from 'rxjs/ajax'
import { catchError, map } from 'rxjs/operators'
import { dataOrThrowErrors, GraphQLResult } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { background } from '../../browser/runtime'
import { isBackground, isInPage } from '../../context'
import { DEFAULT_SOURCEGRAPH_URL, sourcegraphUrl } from '../util/context'
import { RequestContext } from './context'
import { createAuthRequiredError, PrivateRepoPublicSourcegraphComError } from './errors'
import { getHeaders } from './headers'

export interface GraphQLRequestArgs {
    ctx: RequestContext

    /** The GraphQL request (query or mutation) */
    request: string

    /** A key/value object with variable values */
    variables?: any

    /** the url the request is going to */
    url?: string
    requestMightContainPrivateInfo?: boolean
}

function privateRepoPublicSourcegraph({
    url = sourcegraphUrl,
    requestMightContainPrivateInfo,
    ctx,
}: Pick<GraphQLRequestArgs, 'url' | 'requestMightContainPrivateInfo' | 'ctx'>): boolean {
    return !!(ctx.privateRepository && requestMightContainPrivateInfo && url === DEFAULT_SOURCEGRAPH_URL)
}

/**
 * Does a GraphQL request to the Sourcegraph GraphQL API running under `/.api/graphql`
 *
 * @return Observable That emits the result or errors if the HTTP request failed
 */
export const requestGraphQL = <T extends GQL.IGraphQLResponseRoot>(args: GraphQLRequestArgs): Observable<T> => {
    // Make sure all GraphQL API requests are sent from the background page, so as to bypass CORS
    // restrictions when running on private code hosts with the public Sourcegraph instance.  This
    // allows us to run extensions on private code hosts without needing a private Sourcegraph
    // instance. See https://github.com/sourcegraph/sourcegraph/issues/1945.
    if (isBackground || isInPage) {
        return performRequest(args)
    }

    return from(background.requestGraphQL<T>(args))
}

function performRequest<T extends GQL.IQuery | GQL.IMutation>({
    ctx,
    request,
    url = sourcegraphUrl,
    variables = {},
    requestMightContainPrivateInfo = true,
}: GraphQLRequestArgs & { ajaxRequest?: typeof ajax }): Observable<GraphQLResult<T>> {
    const nameMatch = request.match(/^\s*(?:query|mutation)\s+(\w+)/)
    const queryName = nameMatch ? '?' + nameMatch[1] : ''

    // Check if it's a private repo - if so don't make a request to Sourcegraph.com.
    if (privateRepoPublicSourcegraph({ url, ctx, requestMightContainPrivateInfo })) {
        return throwError(new PrivateRepoPublicSourcegraphComError(nameMatch ? nameMatch[1] : '<unnamed>'))
    }

    return ajax({
        method: 'POST',
        url: `${url}/.api/graphql` + queryName,
        headers: getHeaders(),
        crossDomain: true,
        withCredentials: true,
        body: JSON.stringify({ query: request, variables }),
        async: true,
    }).pipe(
        map(({ response }) => {
            if (!response) {
                throw new Error('invalid response received from graphql endpoint')
            }
            return response
        }),
        catchError(err => {
            if (err.status === 401) {
                throw createAuthRequiredError(url)
            }
            throw err
        })
    )
}

/**
 * Does a GraphQL query to the Sourcegraph GraphQL API running under `/.api/graphql`
 */
export const queryGraphQL = (args: GraphQLRequestArgs) =>
    requestGraphQL<GQL.IQuery>(args).pipe(map(result => dataOrThrowErrors(result)))

/**
 * Does a GraphQL mutation to the Sourcegraph GraphQL API running under `/.api/graphql`
 */
export const mutateGraphQL = (args: GraphQLRequestArgs) =>
    requestGraphQL<GQL.IMutation>(args).pipe(map(result => dataOrThrowErrors(result)))
