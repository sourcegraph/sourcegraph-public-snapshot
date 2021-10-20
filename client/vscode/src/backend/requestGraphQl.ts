import { checkOk } from '@sourcegraph/shared/src/backend/fetch'
import { GRAPHQL_URI } from '@sourcegraph/shared/src/graphql/constants'
import { GraphQLResult } from '@sourcegraph/shared/src/graphql/graphql'

export const requestGraphQLFromVSCode = async (
    request: string,
    variables: any,
    sourcegraphURL: string
    // TODO access token
): Promise<GraphQLResult<any>> => {
    const nameMatch = request.match(/^\s*(?:query|mutation)\s+(\w+)/)
    const apiURL = `${GRAPHQL_URI}${nameMatch ? '?' + nameMatch[1] : ''}`

    const response = checkOk(
        await fetch(new URL(apiURL, sourcegraphURL).href, {
            body: JSON.stringify({
                query: request,
                variables,
            }),
            method: 'POST',
        })
    )

    return response.json() as Promise<GraphQLResult<any>>
}
