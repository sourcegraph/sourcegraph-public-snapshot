import { checkOk, isHTTPAuthError } from '@sourcegraph/shared/src/backend/fetch'
import { GRAPHQL_URI } from '@sourcegraph/shared/src/graphql/constants'
import { GraphQLResult } from '@sourcegraph/shared/src/graphql/graphql'
import { asError } from '@sourcegraph/shared/src/util/errors'

import { accessTokenSetting, handleAccessTokenError } from '../settings/accessTokenSetting'
import { endpointSetting } from '../settings/endpointSetting'

export const requestGraphQLFromVSCode = async (request: string, variables: any): Promise<GraphQLResult<any>> => {
    const nameMatch = request.match(/^\s*(?:query|mutation)\s+(\w+)/)
    const apiURL = `${GRAPHQL_URI}${nameMatch ? '?' + nameMatch[1] : ''}`

    const headers: HeadersInit = []

    const sourcegraphURL = endpointSetting()
    const accessToken = accessTokenSetting()
    if (accessToken) {
        headers.push(['Authorization', `token ${accessToken}`])
    }

    try {
        const response = checkOk(
            await fetch(new URL(apiURL, sourcegraphURL).href, {
                body: JSON.stringify({
                    query: request,
                    variables,
                }),
                method: 'POST',
                headers,
            })
        )

        // eslint-disable-next-line @typescript-eslint/return-await
        return response.json() as Promise<GraphQLResult<any>>
    } catch (error) {
        if (isHTTPAuthError(error)) {
            await handleAccessTokenError(accessToken ?? '')
        }
        throw asError(error)
    }
}
