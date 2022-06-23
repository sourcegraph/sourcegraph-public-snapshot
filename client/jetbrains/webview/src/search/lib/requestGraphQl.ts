import { asError } from '@sourcegraph/common'
import { GraphQLResult, GRAPHQL_URI } from '@sourcegraph/http-client'

import { getAccessToken, getInstanceURL } from '..'

export const requestGraphQL = async <R, V = object>(request: string, variables: V): Promise<GraphQLResult<R>> => {
    const instanceURL = getInstanceURL()
    const accessToken = getAccessToken()

    const nameMatch = request.match(/^\s*(?:query|mutation)\s+(\w+)/)
    const apiURL = `${GRAPHQL_URI}${nameMatch ? '?' + nameMatch[1] : ''}`

    const headers = new Headers()
    headers.set('Content-Type', 'application/json')
    headers.set('X-Sourcegraph-Should-Trace', new URLSearchParams(window.location.search).get('trace') || 'false')
    if (accessToken) {
        headers.set('Authorization', `token ${accessToken}`)
    }

    try {
        const url = new URL(apiURL, instanceURL).href
        const response = await fetch(url, {
            body: JSON.stringify({
                query: request,
                variables,
            }),
            method: 'POST',
            headers,
        })
        // eslint-disable-next-line @typescript-eslint/return-await
        return response.json() as Promise<GraphQLResult<any>>
    } catch (error) {
        throw asError(error)
    }
}
