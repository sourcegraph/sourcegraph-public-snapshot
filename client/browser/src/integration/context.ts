import { SharedGraphQlOperations } from '@sourcegraph/shared/src/graphql-operations'
import {
    createSharedIntegrationTestContext,
    IntegrationTestContext,
    IntegrationTestOptions,
} from '@sourcegraph/shared/src/testing/integration/context'
import { BrowserGraphQlOperations } from '../graphql-operations'
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

    sharedTestContext.server.any('chrome-extension://bmfbcejdknlknpncfpeloejonjoledha/*rest').passthrough()

    return {
        ...sharedTestContext,
    }
}
