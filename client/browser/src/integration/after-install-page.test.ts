import percySnapshot from '@percy/puppeteer'

import { createDriverForTest, Driver } from '@sourcegraph/shared/src/testing/driver'

import { BrowserIntegrationTestContext, createBrowserIntegrationTestContext } from './context'
import { closeInstallPageTab } from './shared'

describe('After install page', () => {
    let driver: Driver
    before(async () => {
        driver = await createDriverForTest({ loadExtension: true })
        await closeInstallPageTab(driver.browser)
        if (driver.sourcegraphBaseUrl !== 'https://sourcegraph.com') {
            await driver.setExtensionSourcegraphUrl()
        }
    })
    after(() => driver?.close())

    let testContext: BrowserIntegrationTestContext
    beforeEach(async function () {
        testContext = await createBrowserIntegrationTestContext({
            driver,
            currentTest: this.currentTest!,
            directory: __dirname,
        })

        // Requests to other origins that we need to ignore to prevent breaking tests.
        testContext.server
            .get('https://storage.googleapis.com/sourcegraph-assets/code-host-integration/*path')
            .intercept((request, response) => {
                response.sendStatus(200)
            })

        // Ensure that the same assets are requested in all environments.
        await driver.page.emulateMediaFeatures([{ name: 'prefers-color-scheme', value: 'light' }])
    })

    afterEach(() => testContext?.dispose())

    it('renders after install page content', async function () {
        await driver.openBrowserExtensionPage('after_install')
        await driver.page.waitForSelector("[data-testid='after-install-page-content']")
        // eslint-disable-next-line @typescript-eslint/no-non-null-asserted-optional-chain, @typescript-eslint/no-non-null-assertion
        await percySnapshot(driver.page, this.currentTest?.fullTitle()!)
    })
})
