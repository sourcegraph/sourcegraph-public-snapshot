import { SharedGraphQlOperations } from '@sourcegraph/shared/src/graphql-operations'
import { SearchEvent } from '@sourcegraph/shared/src/search/stream'
import {
    createSharedIntegrationTestContext,
    IntegrationTestContext,
    IntegrationTestOptions,
} from '@sourcegraph/shared/src/testing/integration/context'

import { WebGraphQlOperations } from '../graphql-operations'

import { commonWebGraphQlResults } from './graphQlResults'

export interface WebIntegrationTestContext
    extends IntegrationTestContext<
        WebGraphQlOperations & SharedGraphQlOperations,
        string & keyof (WebGraphQlOperations & SharedGraphQlOperations)
    > {
    /**
     * Configures fake responses for streaming search
     *
     * @param overrides The array of events to return.
     */
    overrideSearchStreamEvents: (overrides: SearchEvent[]) => void
}

/**
 * Creates the integration test context for integration tests testing the web app.
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
        overrideSearchStreamEvents: overrides => {
            searchStreamEventOverrides = overrides
        },
    }
}
