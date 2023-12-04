import { URL } from 'url'

import { isDefined } from '@sourcegraph/common'
import type { SharedGraphQlOperations } from '@sourcegraph/shared/src/graphql-operations'
import {
    createSharedIntegrationTestContext,
    type IntegrationTestContext,
    type IntegrationTestOptions,
} from '@sourcegraph/shared/src/testing/integration/context'

import type { BrowserGraphQlOperations } from '../graphql-operations'

import { commonBrowserGraphQlResults } from './graphql'

export interface BrowserIntegrationTestContext
    extends IntegrationTestContext<
        BrowserGraphQlOperations & SharedGraphQlOperations,
        string & keyof (BrowserGraphQlOperations & SharedGraphQlOperations)
    > {}

export const createBrowserIntegrationTestContext = async ({
    driver,
    currentTest,
    directory,
}: IntegrationTestOptions): Promise<BrowserIntegrationTestContext> => {
    const sharedTestContext = await createSharedIntegrationTestContext<
        BrowserGraphQlOperations & SharedGraphQlOperations,
        string & keyof (BrowserGraphQlOperations & SharedGraphQlOperations)
    >({ driver, currentTest, directory })
    sharedTestContext.overrideGraphQL(commonBrowserGraphQlResults)

    // The Chrome extension id is unstable in CI, so find it at runtime.
    const targets = driver.browser.targets()
    const host = targets
        .map(target => {
            try {
                const { protocol, host } = new URL(target.url())
                if (protocol === 'chrome-extension:') {
                    return host
                }
                return null
            } catch {
                return null
            }
        })
        .find(isDefined)

    sharedTestContext.server.any(`chrome-extension://${host ?? 'bmfbcejdknlknpncfpeloejonjoledha'}/*rest`).passthrough()

    sharedTestContext.server.any('http://localhost:8890/*').intercept((request, response) => {
        response.sendStatus(200)
    })

    return {
        ...sharedTestContext,
    }
}
