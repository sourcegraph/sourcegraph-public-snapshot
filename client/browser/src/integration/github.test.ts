import { createDriverForTest, Driver } from '@sourcegraph/shared/src/testing/driver'
import { BrowserIntegrationTestContext, createBrowserIntegrationTestContext } from './context'
import { closeInstallPageTab } from './shared'
import { afterEachSaveScreenshotIfFailed } from '@sourcegraph/shared/src/testing/screenshotReporter'
import assert from 'assert'
import delay from 'delay'

describe('GitHub', () => {
    let driver: Driver
    before(async () => {
        driver = await createDriverForTest({ loadExtension: true })
        // TODO(tj): Add retries in case this delay isn't large enough in CI
        await delay(1000)
        await closeInstallPageTab(driver.browser)
    })
    after(() => driver?.close())

    let testContext: BrowserIntegrationTestContext
    beforeEach(async function () {
        testContext = await createBrowserIntegrationTestContext({
            driver,
            currentTest: this.currentTest!,
            directory: __dirname,
        })
    })
    afterEachSaveScreenshotIfFailed(() => driver.page)
    afterEach(() => testContext?.dispose())

    it('works', async () => {
        await driver.page.goto('https://github.com')
        assert.strictEqual(true, true)
    })
})
