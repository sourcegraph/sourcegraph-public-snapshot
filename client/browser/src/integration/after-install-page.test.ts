/** @jest-environment setup-polly-jest/jest-environment-node */

import { createDriverForTest, Driver } from '@sourcegraph/shared/src/testing/driver'
import { setupPollyServer } from '@sourcegraph/shared/src/testing/integration/context'

import { BrowserIntegrationTestContext, createBrowserIntegrationTestContext } from './context'
import { closeInstallPageTab } from './shared'

describe('After install page', () => {
    let driver: Driver
    const pollyServer = setupPollyServer()

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
            pollyServer: pollyServer.polly,
        })

        // Requests to other origins that we need to ignore to prevent breaking tests.
        pollyServer.polly.server
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
