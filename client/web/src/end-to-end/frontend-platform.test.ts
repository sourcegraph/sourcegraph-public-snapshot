import expect from 'expect'
import { describe, test, before, after } from 'mocha'

import { getConfig } from '@sourcegraph/shared/src/testing/config'
import { afterEachRecordCoverage } from '@sourcegraph/shared/src/testing/coverage'
import { Driver } from '@sourcegraph/shared/src/testing/driver'
import { afterEachSaveScreenshotIfFailed } from '@sourcegraph/shared/src/testing/screenshotReporter'

import { cloneRepos } from './utils/cloneRepos'
import { initEndToEndTest } from './utils/initEndToEndTest'

const { sourcegraphBaseUrl } = getConfig('gitHubDotComToken', 'sourcegraphBaseUrl')

describe('e2e test suite', () => {
    let driver: Driver

    before(async function () {
        driver = await initEndToEndTest()

        await cloneRepos({
            driver,
            mochaContext: this,
            repoSlugs: ['gorilla/mux'],
        })
    })

    after('Close browser', () => driver?.close())

    afterEachSaveScreenshotIfFailed(() => driver.page)
    afterEachRecordCoverage(() => driver)

    describe('Visual tests', () => {
        test('Repositories list', async () => {
            await driver.page.goto(sourcegraphBaseUrl + '/site-admin/repositories?query=gorilla%2Fmux')
            await driver.page.waitForSelector('a[href="/github.com/gorilla/mux"]', { visible: true })
        })

        test('Site admin overview', async () => {
            await driver.page.goto(sourcegraphBaseUrl + '/site-admin')
            await driver.page.waitForSelector('.test-site-admin-overview-menu', { visible: true })
            await driver.page.waitForSelector('.test-product-certificate', { visible: true })
        })
    })

    describe('Theme switcher', () => {
        const getActiveThemeClasses = (): Promise<string[]> =>
            driver.page.evaluate(() => {
                const themeNode = document.querySelector('.theme')

                if (themeNode) {
                    return [...themeNode.classList].filter(className => className.startsWith('theme-'))
                }

                return []
            })

        test('changes the theme', async () => {
            await driver.page.goto(sourcegraphBaseUrl + '/github.com/gorilla/mux/-/blob/mux.go')
            await driver.page.waitForSelector('.theme.theme-dark, .theme.theme-light', { visible: true })

            expect(await getActiveThemeClasses()).toHaveLength(1)
            await driver.page.waitForSelector('[data-testid="user-nav-item-toggle"')
            await driver.page.click('[data-testid="user-nav-item-toggle"')

            // Switch to dark
            await driver.page.select('[data-testid="theme-toggle', 'dark')
            await driver.page.waitForSelector('.theme.theme-dark', { visible: true })
            expect(await getActiveThemeClasses()).toEqual(expect.arrayContaining(['theme-dark']))

            // Switch to light
            await driver.page.select('[data-testid="theme-toggle', 'light')
            await driver.page.waitForSelector('.theme.theme-light', { visible: true })
            expect(await getActiveThemeClasses()).toEqual(expect.arrayContaining(['theme-light']))
        })
    })
})
