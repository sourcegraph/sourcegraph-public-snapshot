import fs from 'fs'
import path from 'path'

import { SharedGraphQlOperations } from '@sourcegraph/shared/src/graphql-operations'
import { SearchEvent } from '@sourcegraph/shared/src/search/stream'
import { TemporarySettings } from '@sourcegraph/shared/src/settings/temporary/TemporarySettings'
import { getConfig } from '@sourcegraph/shared/src/testing/config'
import {
    createSharedIntegrationTestContext,
    IntegrationTestContext,
    IntegrationTestOptions,
} from '@sourcegraph/shared/src/testing/integration/context'

import { WebpackManifest, getHTMLPage } from '../../dev/webpack/get-html-webpack-plugins'
import { WebGraphQlOperations } from '../graphql-operations'
import { SourcegraphContext } from '../jscontext'

import { isHotReloadEnabled } from './environment'
import { commonWebGraphQlResults } from './graphQlResults'
import { createJsContext } from './jscontext'
import { TemporarySettingsContext } from './temporarySettingsContext'

export interface WebIntegrationTestContext
    extends IntegrationTestContext<
        WebGraphQlOperations & SharedGraphQlOperations,
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

    /**
     * Configures initial values for temporary settings.
     */
    overrideInitialTemporarySettings: (overrides: TemporarySettings) => void
}

const rootDirectory = path.resolve(__dirname, '..', '..', '..', '..')
const manifestFile = path.resolve(rootDirectory, 'ui/assets/webpack.manifest.json')

const getManifestBundles = (): Partial<WebpackManifest> =>
    // eslint-disable-next-line no-sync
    JSON.parse(fs.readFileSync(manifestFile, 'utf-8')) as Partial<WebpackManifest>

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
    const { environment, ...bundles } = getManifestBundles()

    const sharedTestContext = await createSharedIntegrationTestContext<
        WebGraphQlOperations & SharedGraphQlOperations,
        string & keyof (WebGraphQlOperations & SharedGraphQlOperations)
    >({ driver, currentTest, directory })

    sharedTestContext.overrideGraphQL(commonWebGraphQlResults)
    let jsContext = createJsContext({ sourcegraphBaseUrl: driver.sourcegraphBaseUrl })

    const tempSettings = new TemporarySettingsContext()
    sharedTestContext.overrideGraphQL(tempSettings.getGraphQLOverrides())

    const prodChunks = {
        'app.js': bundles['app.js'] || '',
        'app.css': bundles['app.css'],
        'react.js': bundles['react.js'],
        'opentelemetry.js': bundles['opentelemetry.js'],
    }

    const devChunks = {
        'app.js': bundles['app.js'] || '',
        'runtime.js': isHotReloadEnabled ? bundles['runtime.js'] : undefined,
    }

    const appChunks = environment === 'production' ? prodChunks : devChunks
    if (!config.disableAppAssetsMocking) {
        // Serve all requests for index.html (everything that does not match the handlers above) the same index.html
        sharedTestContext.server
            .get(new URL('/*path', driver.sourcegraphBaseUrl).href)
            .filter(request => !request.pathname.startsWith('/-/'))
            .intercept((request, response) => {
                response.type('text/html').send(getHTMLPage(appChunks, { ...jsContext, ...customContext }))
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
        overrideInitialTemporarySettings: overrides => {
            tempSettings.overrideInitialTemporarySettings(overrides)
        },
    }
}
