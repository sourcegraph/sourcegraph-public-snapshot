import {
    createSharedIntegrationTestContext,
    IntegrationTestContext,
    IntegrationTestOptions,
} from '../../../shared/src/testing/integration/context'
import { createJsContext } from './jscontext'
import { SourcegraphContext } from '../jscontext'
import { WebGraphQlOperations } from '../graphql-operations'
import { SharedGraphQlOperations } from '../../../shared/src/graphql-operations'
import html from 'tagged-template-noop'
import { commonWebGraphQlResults } from './graphQlResults'
import { SearchEvent } from '../search/stream'

export interface WebIntegrationTestContext
    extends IntegrationTestContext<
        WebGraphQlOperations & SharedGraphQlOperations,
        string & keyof (WebGraphQlOperations & SharedGraphQlOperations)
    > {
    /**
     * Overrides `window.context` from the default created by `createJsContext()`.
     */
    overrideJsContext: (jsContext: SourcegraphContext) => void

    /**
     * Configures fake responses for streaming search
     *
     * @param overrides The array of events to return.
     */
    overrideSearchStreamEvents: (overrides: SearchEvent[]) => void
}

/**
 * Creates the intergation test context for integration tests testing the web app.
 * This should be called in a `beforeEach()` hook and assigned to a variable `testContext` in the test scope.
 */
export const createWebIntegrationTestContext = async ({
    driver,
    currentTest,
    directory,
}: IntegrationTestOptions): Promise<WebIntegrationTestContext> => {
    const sharedTestContext = await createSharedIntegrationTestContext<
        WebGraphQlOperations & SharedGraphQlOperations,
        string & keyof (WebGraphQlOperations & SharedGraphQlOperations)
    >({ driver, currentTest, directory })
    sharedTestContext.overrideGraphQL(commonWebGraphQlResults)

    // Serve all requests for index.html (everything that does not match the handlers above) the same index.html
    let jsContext = createJsContext({ sourcegraphBaseUrl: sharedTestContext.driver.sourcegraphBaseUrl })
    sharedTestContext.server
        .get(new URL('/*path', driver.sourcegraphBaseUrl).href)
        .filter(request => !request.pathname.startsWith('/-/'))
        .intercept((request, response) => {
            response.type('text/html').send(html`
                <html>
                    <head>
                        <title>Sourcegraph Test</title>
                    </head>
                    <body>
                        <div id="root"></div>
                        <script>
                            window.context = ${JSON.stringify(jsContext)}
                        </script>
                        <script src="/.assets/scripts/app.bundle.js"></script>
                    </body>
                </html>
            `)
        })

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
            jsContext = overrides
        },
        overrideSearchStreamEvents: overrides => {
            searchStreamEventOverrides = overrides
        },
    }
}
