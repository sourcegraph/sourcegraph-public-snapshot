import { asError } from '@sourcegraph/common'
import { checkOk, GRAPHQL_URI, type GraphQLResult, isHTTPAuthError } from '@sourcegraph/http-client'

import { handleAccessTokenError, getAccessToken } from '../settings/accessTokenSetting'
import { endpointRequestHeadersSetting, endpointSetting } from '../settings/endpointSetting'

import { fetch, getProxyAgent, Headers, type HeadersInit } from './fetch'

export const requestGraphQLFromVSCode = async <R, V = object>(
    request: string,
    variables: V,
    overrideAccessToken?: string,
    overrideSourcegraphURL?: string
): Promise<GraphQLResult<R>> => {
    const sourcegraphURL = overrideSourcegraphURL || endpointSetting()
    const accessToken = overrideAccessToken || (await getAccessToken())
    const nameMatch = request.match(/^\s*(?:query|mutation)\s+(\w+)/)
    const apiURL = `${GRAPHQL_URI}${nameMatch ? '?' + nameMatch[1] : ''}`
    const customHeaders = endpointRequestHeadersSetting()
    // create a headers container based on the custom headers configuration if there are any
    // then add Access Token to request header, contributed by @ptxmac!
    const headers = new Headers(customHeaders as HeadersInit)
    headers.set('Content-Type', 'application/json')
    if (accessToken) {
        headers.set('Authorization', `token ${accessToken}`)
    }
    try {
        const url = new URL(apiURL, sourcegraphURL).href
        // Debt: intercepted requests in integration tests
        // have 0 status codes, so don't check in test environment.
        const checkFunction = process.env.IS_TEST ? <T>(value: T): T => value : checkOk
        const options: any = {
            agent: getProxyAgent(),
            body: JSON.stringify({
                query: request,
                variables,
            }),
            method: 'POST',
            headers,
        }

        const response = checkFunction(await fetch(url, options))
        // TODO request cancellation w/ VS Code cancellation tokens.
        return (await response.json()) as GraphQLResult<R>
    } catch (error) {
        // If `overrideAccessToken` is set, we're validating the token
        // and errors will be displayed in the UI.
        if (isHTTPAuthError(error) && !overrideAccessToken) {
            handleAccessTokenError(accessToken || '', sourcegraphURL).then(
                () => {},
                () => {}
            )
        }
        throw asError(error)
    }
}
