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

// In integration test environment, don't make requests until CDP session
// has been established, which seems to take longer for VS Code.
// In addition, the Sourcegraph VS Code extension is activated
// on startup, which may be before we set up request interception.
const TEST_INIT_DELAY = new Promise(resolve => setTimeout(resolve, 2500))

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
    const sourcegraphURL = overrideSourcegraphURL || endpointSetting()
    const accessToken = overrideAccessToken || accessTokenSetting()
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

        let checkFunction: (response: Response) => Response = checkOk
        if (process.env.IS_TEST) {
            // Debt: intercepted requests in integration tests
            // have 0 status codes, so don't check in test environment.
            checkFunction = <T>(value: T): T => value
            await TEST_INIT_DELAY
        }

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
            handleAccessTokenError(accessToken, sourcegraphURL).then(
                () => {},
                () => {}
            )
        }
        throw asError(error)
    }
}
