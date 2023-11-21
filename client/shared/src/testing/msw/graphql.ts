import { fakerEN as faker } from '@faker-js/faker'
import { type DocumentNode, graphqlSync, type GraphQLError } from 'graphql'
import { graphql as mswgraphql, HttpResponse } from 'msw'

import { getDocumentNode } from '@sourcegraph/http-client'

import type { MockRequestHandler, Mocks } from './handler'

export interface MockGraphqlOptions<T extends Mocks> {
    /**
     * The graphql query to mock. If this is not specified, the name option must be specified.
     */
    query?: DocumentNode | string
    /**
     * The name of the graphql operation to mock. If this is not specified, the query option must be specified.
     */
    name?: string
    /**
     * The handler tries to determine the typename of the node to mock from the query.
     * This only works if the node query contains an inline fragment. If it doesn't you can
     * specify the typename here.
     */
    nodeTypename?: string
    /**
     * Additional mock generators to use.
     */
    mocks?: T

    /**
     * When set to true, the mock result will be logged to the console.
     */
    inspect?: boolean

    /**
     * Seed to use to initialze the random data generator. The default seed is
     * 1. If you want to generate different data for each test run, you can
     * set this to a random number or to undefined.
     */
    seed?: number
}

/**
 * Helper function for creating a graphql handler that mocks a specific operation/query.
 */
export function mockGraphql<T extends Mocks = Mocks>(options: MockGraphqlOptions<T>): MockRequestHandler<T> {
    return ({ schema, registerMocks }) => {
        let name: string | undefined = options.name
        if (options.query) {
            // Get operation name from document node in options.query
            const document = getDocumentNode(options.query)
            for (const definition of document.definitions) {
                if (definition.kind === 'OperationDefinition' && definition.operation === 'query') {
                    name = definition.name?.value
                    break
                }
            }
        }

        const context = {
            nodeTypename: options.nodeTypename,
            operationName: name,
        }

        return mswgraphql.operation(({ query, variables, operationName }) => {
            if (!name || operationName === name) {
                const unregister = options.mocks ? registerMocks(options.mocks) : null
                let data: unknown
                let errors: readonly GraphQLError[] | undefined

                try {
                    faker.seed(options.seed ?? 1)
                    ;({ data, errors } = graphqlSync(schema, query, undefined, context, variables))
                } catch (error) {
                    errors = [error]
                } finally {
                    unregister?.()
                }
                if (errors) {
                    // eslint-disable-next-line no-console
                    console.error(
                        `[MSW] Operation '${operationName}' with ${JSON.stringify(variables)} errord:\n${errors
                            .map(error => error.message)
                            .join('\n')}`
                    )
                }
                if (options.inspect) {
                    // eslint-disable-next-line no-console
                    console.log(
                        `[MSW] Mocked operation '${operationName}' with ${JSON.stringify(variables)}: ${JSON.stringify(
                            { data, errors },
                            null,
                            2
                        )}`
                    )
                }
                // eslint-disable-next-line @typescript-eslint/no-explicit-any
                return HttpResponse.json({ data: (data as any) ?? undefined, errors: (errors as any) ?? undefined })
            }
            return undefined
        })
    }
}
