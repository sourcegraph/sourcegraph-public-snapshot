import { describe, test, after, Context } from 'mocha'

import { getConfig } from '@sourcegraph/shared/src/testing/config'
import { afterEachRecordCoverage } from '@sourcegraph/shared/src/testing/coverage'
import type { Driver } from '@sourcegraph/shared/src/testing/driver'
import { afterEachSaveScreenshotIfFailed } from '@sourcegraph/shared/src/testing/screenshotReporter'

import { cloneRepos } from '../utils/cloneRepos'
import { initEndToEndTest } from '../utils/initEndToEndTest'

const { sourcegraphBaseUrl } = getConfig('gitHubDotComToken', 'sourcegraphBaseUrl')

describe('Site Admin', () => {
    let driver: Driver

    after('Close browser', () => driver?.close())

    afterEachSaveScreenshotIfFailed(() => driver.page)
    afterEachRecordCoverage(() => driver)

    // Flaky https://github.com/sourcegraph/sourcegraph/issues/45531
    test.skip('Overview', async () => {
        driver = await initEndToEndTest()
        const ctx = new Context()

        await cloneRepos({
            driver,
            mochaContext: ctx,
            repoSlugs: ['gorilla/mux'],
        })
        if (driver) {
            await driver.page.goto(sourcegraphBaseUrl + '/site-admin')
            await driver.page.waitForSelector('[data-testid="product-certificate"', { visible: true })
        }
    })

    test.skip('Repositories list', async () => {
        driver = await initEndToEndTest()
        if (driver) {
            await driver.page.goto(sourcegraphBaseUrl + '/site-admin/repositories?query=gorilla%2Fmux')
            await driver.page.waitForSelector('a[href="/github.com/gorilla/mux"]', { visible: true })
        }
    })
})
