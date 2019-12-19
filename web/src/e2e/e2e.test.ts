/**
 * @jest-environment node
 */

import * as path from 'path'
import { saveScreenshotsUponFailuresAndClosePage } from '../../../shared/src/e2e/screenshotReporter'
import { createDriverForTest, Driver } from '../../../shared/src/e2e/driver'
import { ExternalServiceKind } from '../../../shared/src/graphql/schema'
import { getConfig } from '../../../shared/src/e2e/config'

const { gitHubToken, sourcegraphBaseUrl } = getConfig('gitHubToken', 'sourcegraphBaseUrl')

// 1 minute test timeout. This must be greater than the default Puppeteer
// command timeout of 30s in order to get the stack trace to point to the
// Puppeteer command that failed instead of a cryptic Jest test timeout
// location.
jest.setTimeout(1 * 60 * 1000)

process.on('unhandledRejection', error => {
    console.error('Caught unhandledRejection:', error)
})

process.on('rejectionHandled', error => {
    console.error('Caught rejectionHandled:', error)
})

describe('e2e test suite', () => {
    let driver: Driver

    async function init(): Promise<void> {
        await driver.ensureLoggedIn({ username: 'test', password: 'test', email: 'test@test.com' })
        await driver.resetUserSettings()
        await driver.ensureHasExternalService({
            kind: ExternalServiceKind.GITHUB,
            displayName: 'e2e-test-github',
            config: JSON.stringify({
                url: 'https://github.com',
                token: gitHubToken,
                repos: ['gorilla/mux'],
            }),
            ensureRepos: ['github.com/gorilla/mux'],
        })
    }

    beforeAll(
        async () => {
            // Start browser.
            driver = await createDriverForTest({ sourcegraphBaseUrl, logBrowserConsole: true })
            await init()
        },
        // Cloning the repositories takes ~1 minute, so give initialization 2
        // minutes instead of 1 (which would be inherited from
        // `jest.setTimeout(1 * 60 * 1000)` above).
        2 * 60 * 1000
    )

    // Close browser.
    afterAll(async () => {
        if (driver) {
            await driver.close()
        }
    })

    // Take a screenshot when a test fails.
    saveScreenshotsUponFailuresAndClosePage(
        path.resolve(__dirname, '..', '..', '..'),
        path.resolve(__dirname, '..', '..', '..', 'puppeteer'),
        () => driver.page
    )

    test('search', async () => {
        // Visit search page
        await driver.page.goto(
            new URL(
                '/search?q=r:%5Egithub%5C.com/gorilla/mux%24+case:yes+mux.NewRouter%28%29&patternType=literal',
                sourcegraphBaseUrl
            ).href
        )
        // Expect results to appear
        await driver.page.waitForSelector('.e2e-file-match-children-item')
        expect(
            await driver.page.evaluate(() => document.querySelectorAll('.e2e-file-match-children-item').length)
        ).toBe(34)
    })

    test('code intelligence', async () => {
        await driver.page.goto(
            new URL(
                '/github.com/gorilla/mux@49c01487a141b49f8ffe06277f3dca3ee80a55fa/-/blob/mux.go#L47:6&tab=references',
                sourcegraphBaseUrl
            ).href
        )
        // Go-to-def button in the hover
        await driver.page.waitForSelector('.e2e-tooltip-go-to-definition')
        // References
        await driver.page.waitForSelector('.e2e-file-match-children-item')
    })
})
