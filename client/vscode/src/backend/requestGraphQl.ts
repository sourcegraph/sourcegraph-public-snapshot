import { asError } from '@sourcegraph/common'
import { checkOk, GraphQLResult, GRAPHQL_URI, isHTTPAuthError } from '@sourcegraph/http-client'

import { accessTokenSetting, handleAccessTokenError } from '../settings/accessTokenSetting'
import { endpointSetting } from '../settings/endpointSetting'

let invalidated = false

/**
 * To be called when Sourcegraph URL changes.
 */
export function invalidateClient(): void {
    invalidated = true
}

export const requestGraphQLFromVSCode = async <R, V = object>(
    request: string,
    variables: V
): Promise<GraphQLResult<R>> => {
    if (invalidated) {
        throw new Error(
            'Sourcegraph GraphQL Client has been invalidated due to instance URL change. Restart VS Code to fix.'
        )
    }

    const nameMatch = request.match(/^\s*(?:query|mutation)\s+(\w+)/)
    const apiURL = `${GRAPHQL_URI}${nameMatch ? '?' + nameMatch[1] : ''}`

    const headers: HeadersInit = []
    const sourcegraphURL = endpointSetting()
    const accessToken = accessTokenSetting()

    // Add Access Token to request header
    if (accessToken) {
        headers.push(['Authorization', `token ${accessToken}`])
    } else {
        headers.push(['Content-Type', 'application/json'])
    }

    try {
        const url = new URL(apiURL, sourcegraphURL).href
        const response = checkOk(
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
        if (isHTTPAuthError(error)) {
            await handleAccessTokenError(accessToken ?? '')
        }
        throw asError(error)
    }
}
