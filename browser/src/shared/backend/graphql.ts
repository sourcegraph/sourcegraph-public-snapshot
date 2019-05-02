const options: GraphQLRequestOptions = {
    headers: getHeaders(),
    requestOptions: {
        crossDomain: true,
        withCredentials: true,
        async: true,
    },
}

/**
 * Does a GraphQL request to the Sourcegraph GraphQL API running under `/.api/graphql`
 *
 * @param request The GraphQL request (query or mutation)
 * @param variables A key/value object with variable values
 * @return Observable That emits the result or errors if the HTTP request failed
 */
export const requestGraphQLFromBackground: PlatformContext['requestGraphQL'] = (
    request,
    variables,
    mightContainPrivateInfo
) => {
    if (isBackground) {
        throw new Error('Should not be called from the background page')
    }
    return from(
        browser.runtime
            .sendMessage({
                type: 'requestGraphQL',
                payload: {
                    request: request[graphQLContent],
                    variables: variables || {},
                    mightContainPrivateInfo,
                },
            })
            .then(response => {
                if (!response || (!response.result && !response.err)) {
                    throw new Error('Invalid requestGraphQL response received from background page')
                }
                const { result, err } = response
                if (err) {
                    throw err
                }
                return result
            })
    )
}
