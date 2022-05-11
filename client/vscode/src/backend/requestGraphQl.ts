import { asError } from '@sourcegraph/common'
import { checkOk, GraphQLResult, GRAPHQL_URI, isHTTPAuthError } from '@sourcegraph/http-client'

import { accessTokenSetting, handleAccessTokenError } from '../settings/accessTokenSetting'
import { endpointSetting, endpointRequestHeadersSetting } from '../settings/endpointSetting'

let invalidated = false

/**
 * To be called when Sourcegraph URL changes.
 */
export function invalidateClient(): void {
    invalidated = true
}

export const requestGraphQLFromVSCode = async <R, V = object>(
    request: string,
    variables: V,
    overrideAccessToken?: string,
    overrideSourcegraphURL?: string
): Promise<GraphQLResult<R>> => {
    if (invalidated) {
        throw new Error(
            'Sourcegraph GraphQL Client has been invalidated due to instance URL change. Restart VS Code to fix.'
        )
    }

    const nameMatch = request.match(/^\s*(?:query|mutation)\s+(\w+)/)
    const apiURL = `${GRAPHQL_URI}${nameMatch ? '?' + nameMatch[1] : ''}`
    // load custom headers from user setting if any
    const customHeaders = endpointRequestHeadersSetting()
    // return empty array if no custom header is provided in configuration
    const headers: HeadersInit = Object.entries(customHeaders)
    const sourcegraphURL = overrideSourcegraphURL || endpointSetting()
    const accessToken = accessTokenSetting()

    // Add Access Token to request header
    // Used to validate access token.
    if (overrideAccessToken) {
        headers.push(['Authorization', `token ${overrideAccessToken}`])
    } else if (accessToken) {
        headers.push(['Authorization', `token ${accessToken}`])
    } else {
        headers.push(['Content-Type', 'application/json'])
    }

    try {
        const url = new URL(apiURL, sourcegraphURL).href

        // Debt: intercepted requests in integration tests
        // have 0 status codes, so don't check in test environment.
        const checkFunction = process.env.IS_TEST ? <T>(value: T): T => value : checkOk

        const response = checkFunction(
            await fetch(url, {
                body: JSON.stringify({
                    query: request,
                    variables,
                }),
                method: 'POST',
                headers,
            })
        )
        // TODO request cancellation w/ VS Code cancellation tokens.

        // eslint-disable-next-line @typescript-eslint/return-await
        return response.json() as Promise<GraphQLResult<any>>
    } catch (error) {
        // If `overrideAccessToken` is set, we're validating the token
        // and errors will be displayed in the UI.
        if (isHTTPAuthError(error) && !overrideAccessToken) {
            handleAccessTokenError(accessToken).then(
                () => {},
                () => {}
            )
        }
        throw asError(error)
    }
}
