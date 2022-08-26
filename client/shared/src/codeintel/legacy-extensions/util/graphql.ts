import { fromBase64 } from 'js-base64'
import * as sourcegraph from '@sourcegraph/extension-api-types'

type GraphQLResponse<T> = GraphQLResponseSuccess<T> | GraphQLResponseError

interface GraphQLResponseSuccess<T> {
    data: T
    errors: undefined
}

interface GraphQLResponseError {
    data: undefined
    errors: Error[]
}

/** The generic type of the queryGraphQL function. */
export type QueryGraphQLFn<T> = (query: string, vars?: { [name: string]: unknown }) => Promise<T>

/**
 * Perform a GraphQL query via the extension host.
 *
 * @param query The GraphQL query string.
 * @param vars The query variables.
 */
export async function queryGraphQL<T>(query: string, vars: { [name: string]: unknown } = {}): Promise<T> {
    const response = await sourcegraph.commands.executeCommand<GraphQLResponse<T>>('queryGraphQL', query, vars)

    if (response.errors !== undefined) {
        throw response.errors.length === 1 ? response.errors[0] : aggregateErrors(response.errors)
    }

    return response.data
}

function aggregateErrors(errors: Error[]): Error {
    return Object.assign(new Error(errors.map(error => error.message).join('\n')), {
        name: 'AggregateError',
        errors,
    })
}

export function graphqlIdToRepoId(id: string): number {
    const decodedId = fromBase64(id)
    return parseInt(decodedId.split(':')[1], 10)
}
