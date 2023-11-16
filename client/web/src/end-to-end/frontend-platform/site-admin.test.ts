import { describe, test, before, after } from 'mocha'

import { getConfig } from '@sourcegraph/shared/src/testing/config'
import type { Driver } from '@sourcegraph/shared/src/testing/driver'
import { afterEachSaveScreenshotIfFailed } from '@sourcegraph/shared/src/testing/screenshotReporter'

import { cloneRepos } from '../utils/cloneRepos'
import { initEndToEndTest } from '../utils/initEndToEndTest'

const { sourcegraphBaseUrl } = getConfig('gitHubDotComToken', 'sourcegraphBaseUrl')

describe('Site Admin', () => {
    let driver: Driver



    // Flaky https://github.com/sourcegraph/sourcegraph/issues/45531
    // test('Overview', async () => {
    //     await driver.page.goto(sourcegraphBaseUrl + '/site-admin')
    //     await driver.page.waitForSelector('[data-testid="product-certificate"', { visible: true })
    // })
    //
    test('Repositories list', async () => {
        driver = await initEndToEndTest()

        await cloneRepos({
            driver,
            mochaContext: this,
            repoSlugs: ['gorilla/mux'],
        })

        //afterEachSaveScreenshotIfFailed(() => driver.page)

        await driver.page.goto(sourcegraphBaseUrl + '/site-admin/repositories?query=gorilla%2Fmux')
        await driver.page.waitForSelector('a[href="/github.com/gorilla/mux"]', { visible: true })

        driver?.close())
    })
})
