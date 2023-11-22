import { afterEach, beforeEach, describe, it } from 'mocha'

import { accessibilityAudit } from '@sourcegraph/shared/src/testing/accessibility'
import { createDriverForTest, type Driver } from '@sourcegraph/shared/src/testing/driver'
import { afterEachSaveScreenshotIfFailed } from '@sourcegraph/shared/src/testing/screenshotReporter'

import { createWebIntegrationTestContext, type WebIntegrationTestContext } from './context'
import { commonWebGraphQlResults } from './graphQlResults'
import { percySnapshotWithVariants } from './utils'

describe('SignIn', () => {
    let driver: Driver
    before(async () => {
        driver = await createDriverForTest()
    })
    after(() => driver?.close())
    let testContext: WebIntegrationTestContext
    beforeEach(async function () {
        testContext = await createWebIntegrationTestContext({
            driver,
            currentTest: this.currentTest!,
            directory: __dirname,
        })
    })
    afterEachSaveScreenshotIfFailed(() => driver.page)
    afterEach(() => testContext?.dispose())

    it('is styled correctly', async () => {
        testContext.overrideGraphQL({
            ...commonWebGraphQlResults,
            CurrentAuthState: () => ({
                currentUser: null,
            }),
        })
        testContext.overrideJsContext({
            currentUser: undefined,
        })

        await driver.page.goto(driver.sourcegraphBaseUrl + '/sign-in')
        await driver.page.waitForSelector('#username-or-email')
        await driver.page.waitForSelector('input[name="password"]')

        await percySnapshotWithVariants(driver.page, 'Sign in page')
        await accessibilityAudit(driver.page)
    })
})
