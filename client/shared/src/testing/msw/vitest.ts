import { readFileSync } from 'node:fs'
import path from 'node:path'

import type { IMockStore } from '@graphql-tools/mock'
import glob from 'glob'
import type { RequestHandler } from 'msw'
import { setupServer } from 'msw/node'
import { afterAll, beforeAll, afterEach } from 'vitest'

import type { TypeMocks } from '../../graphql-types'
import { defaultMocks } from '../graphql/defaultMocks'

import type { MockGraphqlOptions } from './graphql'
import { type MockRequestHandler, setupHandlers } from './handler'

const SCHEMA_DIR = path.resolve(path.join(__dirname, '../../../../../cmd/frontend/graphqlbackend'))

interface MockServerOptions {
    typeDefs?: string
    inspect?: boolean
}

interface MockServer {
    /**
     * The msw server instance.
     */
    server: ReturnType<typeof setupServer>

    /**
     * Convenience method for adding request handlers to the server.
     */
    use(...handlers: RequestHandler[]): void

    /**
     * Convenience method for adding graphql request handlers to the server.
     * @internal
     */
    useGraphqlMock(...handlers: MockRequestHandler<TypeMocks>[]): void

    /**
     * The mock store used by the graphql handlers to generate mock data.
     * You can use this to add additional mocks to the store.
     */
    store: IMockStore

    /**
     * Helper function for creating a graphql handler that mocks a specific operation/query.
     */
    mockGraphql(options: MockGraphqlOptions<TypeMocks>): RequestHandler
}

export function installMockServer(options?: MockServerOptions, ...mocks: MockRequestHandler<TypeMocks>[]): MockServer {
    let typeDefs = options?.typeDefs
    if (!typeDefs) {
        typeDefs = glob
            .sync('**/*.graphql', { cwd: SCHEMA_DIR })
            .map(file => readFileSync(path.join(SCHEMA_DIR, file), 'utf8'))
            .join('\n')
    }
    const graphqlHandlers = setupHandlers({ typeDefs, defaultMocks })
    const server = setupServer(...graphqlHandlers.use(...mocks), graphqlHandlers.mockGraphql({}))

    beforeAll(() => server.listen())
    afterEach(() => server.resetHandlers())
    afterAll(() => server.close())

    if (options?.inspect) {
        server.events.on('request:match', ({ request }) => {
            // eslint-disable-next-line no-console
            console.info(`[MSW] Intercepting ${request.method} ${request.url}`)
        })
    }

    server.events.on('request:unhandled', ({ request }) => {
        // eslint-disable-next-line no-console
        console.warn(`[MSW] Unhandled!!!!! ${request.method} ${request.url}`)
    })

    return {
        server,
        use(...handlers: RequestHandler[]) {
            server.use(...handlers)
        },
        useGraphqlMock(...handlers: MockRequestHandler<TypeMocks>[]) {
            server.use(...graphqlHandlers.use(...handlers))
        },
        store: graphqlHandlers.store,
        mockGraphql: graphqlHandlers.mockGraphql,
    }
}
