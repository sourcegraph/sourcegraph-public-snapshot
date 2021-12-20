import vscode from 'vscode'

import { checkOk, isHTTPAuthError } from '@sourcegraph/shared/src/backend/fetch'
import { GRAPHQL_URI } from '@sourcegraph/shared/src/graphql/constants'
import { GraphQLResult } from '@sourcegraph/shared/src/graphql/graphql'
import { asError } from '@sourcegraph/shared/src/util/errors'

import { accessTokenSetting, handleAccessTokenError } from '../settings/accessTokenSetting'
import { endpointSetting, endpointCorsSetting } from '../settings/endpointSetting'

let invalidated = false

/**
 * To be called when Sourcegraph URL changes.
 */
export function invalidateClient(): void {
    invalidated = true
}

// Check what platform is the user on
// return 'desktop', 'github.dev', 'codespaces', or 'web'
export const currentPlatform = vscode.env.appHost

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
    const corsUrl = endpointCorsSetting()
    // Add Access Token to request header
    if (accessToken) {
        headers.push(['Authorization', `token ${accessToken}`])
    }
    if(currentPlatform!=='desktop'&&!accessToken && !corsUrl){
        throw asError('You must have accessToken and corsUrl configured for Sourcegraph Search to work on VS Code Web')
    }
    try {
        // Add CORS if not on desktop
        const searchUrl = corsUrl
            ? `${new URL('/', corsUrl).href}${new URL(apiURL, sourcegraphURL).href}`
            : new URL(apiURL, sourcegraphURL).href
        const response = checkOk(
            await fetch(searchUrl, {
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
        if (!corsUrl) {
            await vscode.window.showErrorMessage(
                `Failed to connect to endpoint ${sourcegraphURL}. Please make sure you have CORS configured in your setting if you are on VS Code Web.`
            )
        } else {
            await vscode.window.showErrorMessage(
                `Failed to connect using ${corsUrl}. Try removing or using a different corsUrl in your setting.`
            )
        }
        throw asError(error)
    }
}
