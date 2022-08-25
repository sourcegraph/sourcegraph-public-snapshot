import fs from 'fs'
import path from 'path'

import html from 'tagged-template-noop'

import { SearchGraphQlOperations } from '@sourcegraph/search'
import { SearchUIGraphQlOperations } from '@sourcegraph/search-ui'
import { SharedGraphQlOperations } from '@sourcegraph/shared/src/graphql-operations'
import { SearchEvent } from '@sourcegraph/shared/src/search/stream'
import { getConfig } from '@sourcegraph/shared/src/testing/config'
import {
    createSharedIntegrationTestContext,
    IntegrationTestContext,
    IntegrationTestOptions,
} from '@sourcegraph/shared/src/testing/integration/context'

import { WebGraphQlOperations } from '../graphql-operations'
import { SourcegraphContext } from '../jscontext'

import { isHotReloadEnabled } from './environment'
import { commonWebGraphQlResults } from './graphQlResults'
import { createJsContext } from './jscontext'

export interface WebIntegrationTestContext
    extends IntegrationTestContext<
        WebGraphQlOperations & SharedGraphQlOperations & SearchGraphQlOperations & SearchUIGraphQlOperations,
        string & keyof (WebGraphQlOperations & SharedGraphQlOperations)
    > {
    /**
     * Overrides `window.context` from the default created by `createJsContext()`.
     */
    overrideJsContext: (jsContext: Partial<SourcegraphContext>) => void

    /**
     * Configures fake responses for streaming search
     *
     * @param overrides The array of events to return.
     */
    overrideSearchStreamEvents: (overrides: SearchEvent[]) => void
}

const rootDirectory = path.resolve(__dirname, '..', '..', '..', '..')
const manifestFile = path.resolve(rootDirectory, 'ui/assets/webpack.manifest.json')

const getAppBundle = (): string => {
    // eslint-disable-next-line no-sync
    const manifest = JSON.parse(fs.readFileSync(manifestFile, 'utf-8')) as Record<string, string>
    return manifest['app.js']
}

const getRuntimeAppBundle = (): string => {
    // eslint-disable-next-line no-sync
    const manifest = JSON.parse(fs.readFileSync(manifestFile, 'utf-8')) as Record<string, string>
    return manifest['runtime.js']
}

/**
 * Creates the integration test context for integration tests testing the web app.
 * This should be called in a `beforeEach()` hook and assigned to a variable `testContext` in the test scope.
 */
export const createWebIntegrationTestContext = async ({
    driver,
    currentTest,
    directory,
    customContext = {},
}: IntegrationTestOptions): Promise<WebIntegrationTestContext> => {
    const config = getConfig('disableAppAssetsMocking')

    const sharedTestContext = await createSharedIntegrationTestContext<
        WebGraphQlOperations & SharedGraphQlOperations,
        string & keyof (WebGraphQlOperations & SharedGraphQlOperations)
    >({ driver, currentTest, directory })

    sharedTestContext.overrideGraphQL(commonWebGraphQlResults)
    let jsContext = createJsContext({ sourcegraphBaseUrl: driver.sourcegraphBaseUrl })

    if (!config.disableAppAssetsMocking) {
        // On CI, we don't use `react-fast-refresh`, so we don't need the runtime bundle.
        // This branching will be redundant after switching to production bundles for integration tests:
        // https://github.com/sourcegraph/sourcegraph/issues/22831
        const runtimeChunkScriptTag = isHotReloadEnabled ? `<script src=${getRuntimeAppBundle()}></script>` : ''

        // Serve all requests for index.html (everything that does not match the handlers above) the same index.html
        sharedTestContext.server
            .get(new URL('/*path', driver.sourcegraphBaseUrl).href)
            .filter(request => !request.pathname.startsWith('/-/'))
            .intercept((request, response) => {
                response.type('text/html').send(html`
                    <html lang="en">
                        <head>
                            <title>Sourcegraph Test</title>
                        </head>
                        <body>
                            <div id="root"></div>
                            <script>
                                window.context = ${JSON.stringify({ ...jsContext, ...customContext })}
                            </script>
                            ${runtimeChunkScriptTag}
                            <script src=${getAppBundle()}></script>
                        </body>
                    </html>
                `)
            })
    }

    let searchStreamEventOverrides: SearchEvent[] = []
    sharedTestContext.server
        .get(new URL('/search/stream?*params', driver.sourcegraphBaseUrl).href)
        .intercept((request, response) => {
            if (!searchStreamEventOverrides || searchStreamEventOverrides.length === 0) {
                throw new Error(
                    'Search stream event overrides missing. Call overrideSearchStreamEvents() to set the events.'
                )
            }

            const responseContent = searchStreamEventOverrides
                .map(event => `event: ${event.type}\ndata: ${JSON.stringify(event.data)}\n\n`)
                .join('')
            response.status(200).type('text/event-stream').send(responseContent)
        })

    return {
        ...sharedTestContext,
        overrideJsContext: overrides => {
            jsContext = { ...jsContext, ...overrides }
        },
        overrideSearchStreamEvents: overrides => {
            searchStreamEventOverrides = overrides
        },
    }
}
