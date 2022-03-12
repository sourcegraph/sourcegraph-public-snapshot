import { createDriverForTest, Driver } from '@sourcegraph/shared/src/testing/driver'

import { BrowserIntegrationTestContext, createBrowserIntegrationTestContext } from './context'
import { closeInstallPageTab } from './shared'

describe('After install page', () => {
    let driver: Driver
    beforeAll(async () => {
        driver = await createDriverForTest({ loadExtension: true })
        await closeInstallPageTab(driver.browser)
        if (driver.sourcegraphBaseUrl !== 'https://sourcegraph.com') {
            await driver.setExtensionSourcegraphUrl()
        }
    })
    afterAll(() => driver?.close())

    let testContext: BrowserIntegrationTestContext
    beforeEach(async () => {
        testContext = await createBrowserIntegrationTestContext({
            driver,
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

    it('renders after install page content', async () => {
        await driver.openBrowserExtensionPage('after_install')
        await driver.page.$("[data-testid='after-install-page-content']")
    })
})
