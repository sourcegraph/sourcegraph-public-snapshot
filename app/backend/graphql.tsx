import { without } from 'lodash'
import sortBy from 'lodash/sortBy'
import { Observable, throwError } from 'rxjs'
import { ajax } from 'rxjs/ajax'
import { catchError, map } from 'rxjs/operators'
import { isPrivateRepository, repoUrlCache, serverUrls, sourcegraphUrl } from '../util/context'
import { RequestContext } from './context'
import { AuthRequiredError, createAuthRequiredError, NoSourcegraphURLError } from './errors'
import { getHeaders } from './headers'

/**
 * Interface for the response result of a GraphQL query
 */
export interface QueryResult {
    data?: GQL.IQuery
    errors?: GQL.IGraphQLResponseError[]
}

/**
 * Interface for the response result of a GraphQL mutation
 */
export interface MutationResult {
    data?: GQL.IMutation
    errors?: GQL.IGraphQLResponseError[]
}

/**
 * Does a GraphQL request to the Sourcegraph GraphQL API running under `/.api/graphql`
 *
 * @param request The GraphQL request (query or mutation)
 * @param variables A key/value object with variable values
 * @return Observable That emits the result or errors if the HTTP request failed
 */
function requestGraphQL(
    ctx: RequestContext,
    request: string,
    variables: any = {},
    urlsToTry: string[],
    retry = true,
    authError?: AuthRequiredError
): Observable<GQL.IGraphQLResponseRoot> {
    let urls: string[] = ctx.blacklist ? without(urlsToTry, ...ctx.blacklist) : urlsToTry
    const defaultUrl = repoUrlCache[ctx.repoKey] || sourcegraphUrl
    urls = sortBy(urls, url => (url === defaultUrl ? 0 : 1))
    // Check if it's a private repo - if so don't make a request to Sourcegraph.com.
    if (isPrivateRepository()) {
        urls = without(urls, 'https://sourcegraph.com')
    }

    if (urls.length === 0) {
        return throwError(new NoSourcegraphURLError())
    }
    const url = urls[0]

    const nameMatch = request.match(/^\s*(?:query|mutation)\s+(\w+)/)
    const queryName = nameMatch ? '?' + nameMatch[1] : ''

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
            if (shouldResponseTriggerRetryOrError(response)) {
                delete repoUrlCache[ctx.repoKey]
                throw response
            }
            if (ctx.isRepoSpecific && response.data.repository) {
                repoUrlCache[ctx.repoKey] = url
            }
            return response
        }),
        catchError(err => {
            if (err.status === 401) {
                // Ensure all urls are tried and update authError to be the last seen 401.
                // This ensures that the correct URL is used for sign in and also that all possible
                // urls were checked.
                authError = createAuthRequiredError(url)
            }

            if (!retry || urls.length === 1) {
                // If there was an auth error and we tried all of the possible URLs throw the auth error.
                if (authError) {
                    throw authError
                }
                delete repoUrlCache[ctx.repoKey]
                // We just tried the last url
                throw err
            }
            return requestGraphQL(ctx, request, variables, urls.slice(1), retry, authError)
        })
    )
}

/**
 * Checks the GraphQL response to determine if the response should trigger a retry.
 * The browser extension can have multiple Sourcegraph Server URLs and it is not always known which URL will return
 * a repository or if any of the Server URLs have a repository. This means in some cases we need to check if we should trigger
 * the retry block by throwing an error.
 *
 * Conditions:
 * 1. There is no response data.
 * 2. Attempting to fetch a repository returned null. response.data.repository will be undefined if the GraphQL query did not request a repository.
 * 3. resolveRev return null for a commit and the repository was also not currently cloning.
 */
function shouldResponseTriggerRetryOrError(response: any): boolean {
    if (!response || !response.data) {
        return true
    }
    const { repository } = response.data
    if (repository === undefined) {
        return false
    }
    if (repository === null) {
        return true
    }
    if (
        repository.commit === null &&
        (!response.data.repository.mirrorInfo || !response.data.repository.mirrorInfo.cloneInProgress)
    ) {
        return true
    }
    return false
}

/**
 * Does a GraphQL query to the Sourcegraph GraphQL API running under `/.api/graphql`
 *
 * @param query The GraphQL query
 * @param variables A key/value object with variable values
 * @return Observable That emits the result or errors if the HTTP request failed
 */
export function queryGraphQL(ctx: RequestContext, query: string, variables: any = {}): Observable<QueryResult> {
    return requestGraphQL(ctx, query, variables, serverUrls) as Observable<QueryResult>
}

/**
 * Does a GraphQL mutation to the Sourcegraph GraphQL API running under `/.api/graphql`
 *
 * @param mutation The GraphQL mutation
 * @param variables A key/value object with variable values
 * @return Observable That emits the result or errors if the HTTP request failed
 */
export function mutateGraphQL(ctx: RequestContext, mutation: string, variables: any = {}): Observable<MutationResult> {
    return requestGraphQL(ctx, mutation, variables, serverUrls) as Observable<MutationResult>
}

/**
 * Does a GraphQL mutation to the Sourcegraph GraphQL API running under `/.api/graphql`.
 * Unlike mutateGraphQL, if the first request fails, this will not retry with the rest of the server URLs.
 *
 * @param mutation The GraphQL mutation
 * @param variables A key/value object with variable values
 * @return Observable That emits the result or errors if the HTTP request failed
 */
export function mutateGraphQLNoRetry(
    ctx: RequestContext,
    mutation: string,
    variables: any = {}
): Observable<MutationResult> {
    return requestGraphQL(ctx, mutation, variables, serverUrls, false) as Observable<MutationResult>
}
