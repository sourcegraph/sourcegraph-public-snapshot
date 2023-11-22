import expect from 'expect'
import { describe, test, before, after } from 'mocha'

import { getConfig } from '@sourcegraph/shared/src/testing/config'
import type { Driver } from '@sourcegraph/shared/src/testing/driver'
import { afterEachSaveScreenshotIfFailed } from '@sourcegraph/shared/src/testing/screenshotReporter'

import { initEndToEndTest } from '../utils/initEndToEndTest'

const { sourcegraphBaseUrl } = getConfig('gitHubDotComToken', 'sourcegraphBaseUrl')

// Since the test inside the describe is skipped the after does not execute. This is a known bug and since
describe.skip('Theme switcher', () => {
    let driver: Driver

    before(async () => {
        driver = await initEndToEndTest()
    })

    after('Close browser', () => driver?.close())

    afterEachSaveScreenshotIfFailed(() => driver.page)

    const getActiveThemeClasses = (): Promise<string[]> =>
        driver.page.evaluate(() => {
            const themeNode = document.querySelector('.theme')

            if (themeNode) {
                return [...themeNode.classList].filter(className => className.startsWith('theme-'))
            }

            return []
        })

    // Disabled as flaky
    test.skip('changes the theme', async () => {
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
