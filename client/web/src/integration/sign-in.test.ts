/** @jest-environment setup-polly-jest/jest-environment-node */

import { createDriverForTest, Driver } from '@sourcegraph/shared/src/testing/driver'
import { setupPollyServer } from '@sourcegraph/shared/src/testing/integration/context'
import { afterEachSaveScreenshotIfFailedWithJest } from '@sourcegraph/shared/src/testing/screenshotReporter'

import { createWebIntegrationTestContext, WebIntegrationTestContext } from './context'
import { commonWebGraphQlResults } from './graphQlResults'
import { percySnapshotWithVariants } from './utils'

describe('SignIn', () => {
    let driver: Driver
    const pollyServer = setupPollyServer(__dirname)

    beforeAll(async () => {
        driver = await createDriverForTest()
    })
    afterAll(() => driver?.close())
    let testContext: WebIntegrationTestContext
    beforeEach(async () => {
        testContext = await createWebIntegrationTestContext({
            driver,
            directory: __dirname,
            pollyServer: pollyServer.polly,
        })
    })
    afterEachSaveScreenshotIfFailedWithJest(() => driver.page)
    afterEach(() => testContext?.dispose())

    it('is styled correctly', async () => {
        testContext.overrideGraphQL({
            ...commonWebGraphQlResults,
            CurrentAuthState: () => ({
                currentUser: null,
            }),
        })
        await driver.page.goto(driver.sourcegraphBaseUrl + '/sign-in')
        await driver.page.waitForSelector('#username-or-email')
        await driver.page.waitForSelector('input[name="password"]')

        await percySnapshotWithVariants(driver.page, 'Sign in page')
    })
})
