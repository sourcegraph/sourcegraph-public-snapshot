import * as GQL from '../../../shared/src/graphql/schema'
import { Driver } from '../../../shared/src/e2e/driver'
import { GraphQLClient } from './util/GraphQLClient'
import { setTestDefaults, createAndInitializeDriver } from './util/init'
import { Config, getConfig } from '../../../shared/src/e2e/config'
import { ensureLoggedInOrCreateUser } from './util/helpers'
import { ensureExternalService, waitForRepos } from './util/api'
import { BoundingBox } from 'puppeteer'
import { Key } from 'ts-key-enum'

const testRepoSlugs = ['auth0/go-jwt-middleware', 'kyoshidajp/ghkw', 'PalmStoneGames/kube-cert-manager']

interface ExpectedScreenshot {
    screenshotFile: string
    description: string
}

/**
 * Utility class to verify screenshots match a particular description
 */
class ScreenshotVerifier {
    public screenshots: ExpectedScreenshot[]
    constructor(public driver: Driver) {
        this.screenshots = []
    }

    public async verifyScreenshot({
        filename,
        description,
        clip,
    }: {
        /**
         * The filename to which to save the screenshot. It should be a decsriptive name that captures both
         * what should be happening in the screenshot and the context around what's happening. E.g.,
         * "progress-bar-after-initial-search-is-half-green.png"
         */
        filename: string

        /**
         * A short description of what should happen in the screenshot.
         */
        description: string
        clip?: BoundingBox
    }) {
        await this.driver.page.screenshot({
            path: filename,
            clip,
        })
        this.screenshots.push({ screenshotFile: filename, description })
    }

    public async verifySelector(
        filename: string,
        description: string,
        selector: string,
        waitForSelectorToBeVisibleTimeout: number = 0
    ) {
        if (waitForSelectorToBeVisibleTimeout > 0) {
            await this.driver.page.waitForFunction(
                selector => {
                    const element = document.querySelector(selector)
                    if (!element) {
                        return false
                    }
                    const { width, height } = element.getBoundingClientRect()
                    return width > 0 && height > 0
                },
                { timeout: waitForSelectorToBeVisibleTimeout },
                selector
            )
        }

        const clip: BoundingBox | undefined = await this.driver.page.evaluate(selector => {
            const element = document.querySelector(selector)
            if (!element) {
                throw new Error(`element with selector ${JSON.stringify(selector)} not found`)
            }
            const { left, top, width, height } = element.getBoundingClientRect()
            return { x: left, y: top, width, height }
        }, selector)
        await this.verifyScreenshot({
            filename,
            description,
            clip,
        })
    }

    /**
     * Returns instructions to manually verify each screenshot stored in the to-verify list.
     */
    public verificationInstructions(): string {
        return `
        @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@
        @@@@ Manual verification steps required!!! @@@@
        @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@

        Please verify the following screenshots match the corresponding descriptions:

        ${this.screenshots.map(s => `${s.screenshotFile}:\t${JSON.stringify(s.description)}`).join('\n        ')}

        `
        // TODO
    }
}

describe('Onboarding', () => {
    let config: Pick<Config, 'sudoToken' | 'sudoUsername' | 'gitHubToken' | 'sourcegraphBaseUrl'>
    let driver: Driver
    let gqlClient: GraphQLClient
    let screenshots: ScreenshotVerifier

    beforeAll(
        async () => {
            config = getConfig(['sudoToken', 'sudoUsername', 'gitHubToken', 'sourcegraphBaseUrl'])
            driver = await createAndInitializeDriver(config.sourcegraphBaseUrl)
            gqlClient = GraphQLClient.newForPuppeteerTest({
                baseURL: config.sourcegraphBaseUrl,
                sudoToken: config.sudoToken,
                username: config.sudoUsername,
            })
            screenshots = new ScreenshotVerifier(driver)
            setTestDefaults(driver)
            await ensureLoggedInOrCreateUser({
                driver,
                gqlClient,
                username: 'test-onboarding-regression-test-user',
                password: 'test',
                deleteIfExists: true,
            })
            await ensureExternalService(gqlClient, {
                kind: GQL.ExternalServiceKind.GITHUB,
                uniqueDisplayName: 'GitHub (search-regression-test)',
                config: {
                    url: 'https://github.com',
                    token: config.gitHubToken,
                    repos: testRepoSlugs,
                    repositoryQuery: ['none'],
                },
            })
            await waitForRepos(gqlClient, ['github.com/' + testRepoSlugs[testRepoSlugs.length - 1]])
        },
        20 * 1000 // wait 20s for cloning
    )

    afterAll(async () => {
        if (driver) {
            await driver.close()
        }
        console.log(screenshots.verificationInstructions())
    })

    test(
        'Non-admin user onboarding',
        async () => {
            const statusBarSelector = '.activation-dropdown-button__progress-bar-container'

            // Initial status indicator
            await driver.page.goto(config.sourcegraphBaseUrl + '/search')
            await screenshots.verifySelector(
                'initial-progress-bar-is-gray-circle.png',
                'gray circle',
                statusBarSelector,
                2000
            )

            // Do a search
            await driver.page.type('.e2e-query-input', 'asdf')
            await driver.page.keyboard.press(Key.Enter)
            await new Promise(resolve => setTimeout(resolve, 500)) // allow some time for confetti to play
            await screenshots.verifyScreenshot({
                filename: 'confetti-appears-after-first-search.png',
                description: 'confetti',
            })
            await screenshots.verifySelector(
                'progress-bar-after-initial-search-is-half-green.png',
                '50% green circle',
                statusBarSelector
            )

            // Do a find references
            await driver.page.goto(
                config.sourcegraphBaseUrl +
                    '/ghe.sgdev.org/sourcegraph/gorillalabs-sparkling/-/blob/src/java/sparkling/function/FlatMapFunction.java#L10:12&tab=references'
            )
            const defTokenXPath = '//*[contains(@class, "blob-page__blob")]//span[./text()="FlatMapFunction"]'
            await driver.page.waitForXPath(defTokenXPath)
            const elems = await driver.page.$x(defTokenXPath)
            await elems[0].click()
            await Promise.all(elems.map(elem => elem.dispose()))
            const findRefsSelector = '.e2e-tooltip-find-references'
            await driver.page.waitForSelector(findRefsSelector)
            await driver.page.click(findRefsSelector)
            await driver.page.waitForSelector('.e2e-search-result')
            await new Promise(resolve => setTimeout(resolve, 500)) // allow some time for confetti to play

            await screenshots.verifyScreenshot({
                filename: 'confetti-appears-after-find-refs.png',
                description: 'confetti',
            })
            await screenshots.verifySelector(
                'progress-bar-after-search-and-fnd-refs-is-full-green.png',
                '100% green circle',
                statusBarSelector
            )

            await driver.page.reload()

            // Wait for status bar to appear but it should be invisible
            await driver.page.waitForFunction(
                statusBarSelector => {
                    const element = document.querySelector(statusBarSelector)
                    if (!element) {
                        return false
                    }
                    const { width, height } = element.getBoundingClientRect()
                    return width === 0 && height === 0
                },
                { timeout: 100000 },
                statusBarSelector
            )
        },
        30 * 1000
    )
})
