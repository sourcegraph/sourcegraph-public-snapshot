import { addMocksToSchema, createMockStore } from '@graphql-tools/mock'
// @graphql-tools seems to import the CommonJS version of graphql. We need to import the same version
// otherwise we get errors like "Cannot use GraphQLSchema "[object Object]" from another module or realm."
// eslint-disable-next-line import/extensions
import { buildSchema, isObjectType, type GraphQLSchema } from 'graphql'
import { merge } from 'lodash'
import type { RequestHandler } from 'msw'

import { defaultMocks } from '../graphql/defaultMocks'
import { getDefaultResolvers } from '../graphql/resolvers'

import { type MockGraphqlOptions, mockGraphql } from './graphql'

interface MockRequestHandlerOptions {
    schema: GraphQLSchema
    registerMocks: (mocks: Record<string, () => any>) => () => void
}

export type MockRequestHandler = (options: MockRequestHandlerOptions) => RequestHandler

export interface HandlerSetupOptions {
    typeDefs: string
}

interface MockHandler {
    store: ReturnType<typeof createMockStore>
    mockGraphql(options: MockGraphqlOptions): RequestHandler
    use(...handlers: MockRequestHandler[]): RequestHandler[]
}

export function setupHandlers(options: HandlerSetupOptions): MockHandler {
    const schema = buildSchema(options.typeDefs)
    // Extract GraphQL type names from the schema.
    const typeNames = Object.values(schema.getTypeMap())
        .filter(type => !type.name.startsWith('__') && isObjectType(type))
        .map(type => type.name)
    let requestMocks: Record<string, () => unknown> = {}
    const mocks: Record<string, () => unknown> = { ...defaultMocks }
    for (const typeName of typeNames) {
        mocks[typeName] = () => {
            if (requestMocks[typeName]) {
                return requestMocks[typeName]()
            }
            if (defaultMocks[typeName]) {
                return defaultMocks[typeName]()
            }
            return {}
        }
    }
    const store = createMockStore({
        schema,
        mocks,
    })

    const mockedSchema = addMocksToSchema({
        schema,
        store,
        resolvers: store => merge(getDefaultResolvers(store)),
    })

    function registerMocks(mocks: Record<string, () => unknown>): () => void {
        requestMocks = mocks
        return () => {
            requestMocks = {}
        }
    }

    return {
        store,
        mockGraphql(options: MockGraphqlOptions): RequestHandler {
            return mockGraphql(options)({ schema: mockedSchema, registerMocks })
        },
        use(...handlers: MockRequestHandler[]) {
            return handlers.map(handler => handler({ schema: mockedSchema, registerMocks }))
        },
    }
}
