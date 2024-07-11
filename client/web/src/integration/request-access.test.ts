import { afterEach, beforeEach, describe, it } from 'mocha'

import { accessibilityAudit } from '@sourcegraph/shared/src/testing/accessibility'
import { createDriverForTest, type Driver } from '@sourcegraph/shared/src/testing/driver'
import { afterEachSaveScreenshotIfFailed } from '@sourcegraph/shared/src/testing/screenshotReporter'

import { createWebIntegrationTestContext, type WebIntegrationTestContext } from './context'
import { commonWebGraphQlResults } from './graphQlResults'

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
        testContext.overrideJsContext({
            isAuthenticatedUser: undefined,
            currentUser: null,
        })
        await driver.page.goto(driver.sourcegraphBaseUrl + '/sign-in')
        const selector = await driver.page.waitForSelector('a[href="/request-access"]')
        await selector?.click()

        await driver.page.waitForSelector('#name')
        await driver.page.waitForSelector('#email')
        await driver.page.waitForSelector('#additionalInfo')
        await accessibilityAudit(driver.page)
    })

    it('post-submit step is styled correctly', async () => {
        testContext.overrideGraphQL({
            ...commonWebGraphQlResults,
            CurrentAuthState: () => ({
                currentUser: null,
            }),
        })
        testContext.overrideJsContext({
            isAuthenticatedUser: undefined,
            currentUser: null,
        })
        await driver.page.goto(driver.sourcegraphBaseUrl + '/sign-in')
        // TODO: replace with driver.page.goto(driver.sourcegraphBaseUrl + '/request-access/done') once PR is merged and dogfood is updated
        // This is a workaround for the fact that we can't navigate to a page because integration tests uses dogfood backend instead of local backend
        await driver.page.evaluate(() => {
            window.history.replaceState({}, '', '/request-access/done')
        })
        await driver.page.waitForSelector('[data-testid="request-access-post-submit"]')
        await accessibilityAudit(driver.page)
    })
})
