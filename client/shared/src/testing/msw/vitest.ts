import { readFileSync } from 'node:fs'
import path from 'node:path'

import glob from 'glob'
import type { GraphQLSchema } from 'graphql'
import type { RequestHandler } from 'msw'
import { setupServer } from 'msw/node'
import { afterAll, beforeAll, afterEach } from 'vitest'

import { setupHandlers } from './handler'

interface MockRequestHandlerOptions {
    schema: GraphQLSchema
    registerMocks: (mocks: Record<string, () => unknown>) => () => void
}

export type MockRequestHandler = (options: MockRequestHandlerOptions) => RequestHandler

const SCHEMA_DIR = path.resolve(path.join(__dirname, '../../../../../cmd/frontend/graphqlbackend'))

interface MockServerOptions {
    typeDefs?: string
    inspect?: boolean
}

export function installMockServer(options: MockServerOptions, ...mocks: MockRequestHandler[]) {
    let typeDefs = options.typeDefs
    if (!typeDefs) {
        typeDefs = glob
            .sync('**/*.graphql', { cwd: SCHEMA_DIR })
            .map(file => readFileSync(path.join(SCHEMA_DIR, file), 'utf8'))
            .join('\n')
    }
    const graphqlHandlers = setupHandlers({ typeDefs })
    const server = setupServer(...graphqlHandlers.use(...mocks), graphqlHandlers.mockGraphql({}))

    beforeAll(() => server.listen())
    afterEach(() => server.resetHandlers())
    afterAll(() => server.close())

    if (options.inspect) {
        server.events.on('request:match', ({ request }) => {
            console.log(`[MSW] Intercepting ${request.method} ${request.url}`)
        })
    }

    server.events.on('request:unhandled', ({ request }) => {
        console.log(`[MSW] Unhandled!!!!! ${request.method} ${request.url}`)
    })

    return {
        server,
        use(...handlers: RequestHandler[]) {
            server.use(...handlers)
        },
        useGraphqlMock(...handlers: MockRequestHandler[]) {
            server.use(...graphqlHandlers.use(...handlers))
        },
        store: graphqlHandlers.store,
        mockGraphql: graphqlHandlers.mockGraphql,
    }
}
