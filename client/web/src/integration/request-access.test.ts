import { accessibilityAudit } from '@sourcegraph/shared/src/testing/accessibility'
import { createDriverForTest, Driver } from '@sourcegraph/shared/src/testing/driver'
import { afterEachSaveScreenshotIfFailed } from '@sourcegraph/shared/src/testing/screenshotReporter'

import { createWebIntegrationTestContext, WebIntegrationTestContext } from './context'
import { commonWebGraphQlResults } from './graphQlResults'
import { percySnapshotWithVariants } from './utils'

describe('RequestAccess', () => {
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

    it('form step is styled correctly', async () => {
        testContext.overrideGraphQL({
            ...commonWebGraphQlResults,
            CurrentAuthState: () => ({
                currentUser: null,
            }),
        })
        await driver.page.goto(driver.sourcegraphBaseUrl + '/request-access')
        await driver.page.waitForSelector('#name')
        await driver.page.waitForSelector('#email')
        await driver.page.waitForSelector('#additionalInfo')

        await percySnapshotWithVariants(driver.page, 'Request access page')
        await accessibilityAudit(driver.page)
    })

    it('post-submit step is styled correctly', async () => {
        testContext.overrideGraphQL({
            ...commonWebGraphQlResults,
            CurrentAuthState: () => ({
                currentUser: null,
            }),
        })
        await driver.page.goto(driver.sourcegraphBaseUrl + '/request-access/done')
        await driver.page.waitForSelector('[data-testid="request-access-post-submit"]')

        await percySnapshotWithVariants(driver.page, 'Request access page post-submit')
        await accessibilityAudit(driver.page)
    })
})
