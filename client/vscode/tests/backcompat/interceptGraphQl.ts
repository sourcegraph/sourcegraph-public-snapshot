import { addMocksToSchema } from '@graphql-tools/mock'
import { makeExecutableSchema } from '@graphql-tools/schema'
import { graphql, GraphQLSchema } from 'graphql'

import { IntegrationTestOptions } from '@sourcegraph/shared/src/testing/integration/context'

import { decompressSchema } from './decompressOldSchema'

export function createInterceptGraphQLForOldSchema({
    sourcegraphBaseUrl,
    schemaWithMocks,
}: {
    sourcegraphBaseUrl: string
    schemaWithMocks: GraphQLSchema
}): NonNullable<IntegrationTestOptions['interceptGraphQL']> {
    return function (server, onGraphQLRequest) {
        // Resolver-based GQL mocks for backcompat testing against
        // the oldest supported Sourcegraph GQL schema.
        server.post(new URL('/.api/graphql', sourcegraphBaseUrl).href).intercept(async (request, response) => {
            response.setHeader('Access-Control-Allow-Origin', '*')

            const operationName = new URL(request.absoluteUrl).search.slice(1)

            const { variables, query } = request.jsonBody() as {
                query: string
                variables: Record<string, any>
            }
            onGraphQLRequest({ operationName, variables })

            const { data, errors } = await graphql({
                schema: schemaWithMocks,
                source: query,
                variableValues: variables,
            })
            // To be implemented: missing mock errors.
            if (errors) {
                response.json({ data: undefined, errors })
            } else {
                response.json({ data, errors: undefined })
            }
        })
    }
}

/**
 * Incomplete set of mocks to test compatibility with older versions
 * of the Sourcegraph GraphQL API. To be completed when we address
 * bugs revealed by this testing technique.
 */
export async function createSchemaWithMocks(): Promise<GraphQLSchema> {
    const oldSchemaString = await decompressSchema()
    const schema = makeExecutableSchema({
        typeDefs: oldSchemaString,
    })
    const schemaWithMocks = addMocksToSchema({
        schema,

        mocks: {
            Site: {
                productVersion: () => '3320',
            },
            ViewerSettings: {
                __typename: () => 'SettingsCascade',
                final: () => {},
                subjects: () => [{}],
            },
            SettingsCascade: {
                __typename: () => 'SettingsCascade',
                final: () => {},
                subjects: () => [
                    {
                        __typename: 'User',
                        displayName: 'Test User',
                        latestSettings: {
                            id: 123,
                            contents: '{}',
                        },
                    },
                ],
            },
            SettingsSubject: {
                __typename: () => 'User',
            },
            // Custom scalar types.
            DateTime: () => '2021-01-05T17:08:49.000-0430',
        },
    })
    return schemaWithMocks
}
