import type { SharedGraphQlOperations } from '@sourcegraph/shared/src/graphql-operations'
import type { SearchEvent } from '@sourcegraph/shared/src/search/stream'
import type { TemporarySettings } from '@sourcegraph/shared/src/settings/temporary/TemporarySettings'
import { getConfig } from '@sourcegraph/shared/src/testing/config'
import {
    createSharedIntegrationTestContext,
    type IntegrationTestContext,
    type IntegrationTestOptions,
} from '@sourcegraph/shared/src/testing/integration/context'

import { getWebBuildManifest, getIndexHTML } from '../../dev/utils/get-index-html'
import type { WebGraphQlOperations } from '../graphql-operations'
import type { SourcegraphContext } from '../jscontext'

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

    const tempSettings = new TemporarySettingsContext()
    sharedTestContext.overrideGraphQL(tempSettings.getGraphQLOverrides())

    if (!config.disableAppAssetsMocking) {
        // Serve all requests for index.html (everything that does not match the handlers above) the same index.html
        sharedTestContext.server
            .get(new URL('/*path', driver.sourcegraphBaseUrl).href)
            .filter(request => !request.pathname.startsWith('/-/'))
            .intercept((request, response) => {
                response.type('text/html').send(
                    getIndexHTML({
                        manifest: getWebBuildManifest(),
                        jsContext: { ...jsContext, ...customContext },
                    })
                )
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

    // Let browser handle data: URIs
    sharedTestContext.server.get('data:*rest').passthrough()

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
