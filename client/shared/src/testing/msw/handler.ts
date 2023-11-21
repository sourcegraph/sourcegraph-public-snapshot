import { addMocksToSchema, createMockStore } from '@graphql-tools/mock'
import { buildSchema, isObjectType, type GraphQLSchema } from 'graphql'
import { merge } from 'lodash'
import type { RequestHandler } from 'msw'

import type { TypeMocks } from '../../graphql-types'
import { getDefaultResolvers } from '../graphql/resolvers'

import { type MockGraphqlOptions, mockGraphql } from './graphql'

export type Mocks = Record<string, () => unknown>

interface MockRequestHandlerOptions<T extends Mocks> {
    schema: GraphQLSchema
    registerMocks: (mocks: T) => () => void
}

export type MockRequestHandler<T extends Mocks> = (options: MockRequestHandlerOptions<T>) => RequestHandler

export interface HandlerSetupOptions<T extends Mocks> {
    typeDefs: string
    defaultMocks?: T
}

export interface MockHandler<T extends Mocks> {
    store: ReturnType<typeof createMockStore>
    mockGraphql(options: MockGraphqlOptions<T>): RequestHandler
    use(...handlers: MockRequestHandler<T>[]): RequestHandler[]
}

export function setupHandlers<T extends Mocks>(options: HandlerSetupOptions<T>): MockHandler<T> {
    const schema = buildSchema(options.typeDefs)
    // Extract GraphQL type names from the schema.
    const typeNames = Object.values(schema.getTypeMap())
        .filter(type => !type.name.startsWith('__') && isObjectType(type))
        .map(type => type.name)
    const defaultMocks = options.defaultMocks
    let requestMocks: T | undefined
    const mocks: TypeMocks = { ...defaultMocks }
    for (const typeName of typeNames) {
        mocks[typeName] = () => {
            if (requestMocks?.[typeName]) {
                return requestMocks[typeName]()
            }
            if (defaultMocks?.[typeName]) {
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

    function registerMocks(mocks: T): () => void {
        requestMocks = mocks
        return () => {
            requestMocks = undefined
        }
    }

    return {
        store,
        mockGraphql(options: MockGraphqlOptions<T>): RequestHandler {
            return mockGraphql<T>(options)({ schema: mockedSchema, registerMocks })
        },
        use(...handlers: MockRequestHandler<T>[]) {
            return handlers.map(handler => handler({ schema: mockedSchema, registerMocks }))
        },
    }
}
