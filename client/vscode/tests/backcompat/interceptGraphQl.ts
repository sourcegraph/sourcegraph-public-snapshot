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
        server.post(new URL('/.api/graphql', sourcegraphBaseUrl).href).intercept((request, response) => {
            response.setHeader('Access-Control-Allow-Origin', '*')

            const operationName = new URL(request.absoluteUrl).search.slice(1)
            const { variables, query } = request.jsonBody() as {
                query: string
                variables: Record<string, any>
            }
            onGraphQLRequest({ operationName, variables })

            console.log({ operationName, variables, query })

            // TODO: this is working for most requests. work on refactoring, then
            // fixing individual backcompat cases
            graphql({
                schema: schemaWithMocks,
                source: query,
                variableValues: variables,
            })
                .then(gqlTest => {
                    console.dir(gqlTest, { depth: 4 })
                })
                .catch(() => {})

            response.send(400)
        })
    }
}

// TODO accept overrides
export async function createSchemaWithMocks(): Promise<GraphQLSchema> {
    const oldSchemaString = await decompressSchema()
    const schema = makeExecutableSchema({
        typeDefs: oldSchemaString,
    })
    const schemaWithMocks = addMocksToSchema({
        schema,
        mocks: {},
    })
    return schemaWithMocks
}
