/**
 * @jest-environment node
 */

import * as GQL from '../../../shared/src/graphql/schema'
import { Driver } from '../../../shared/src/e2e/driver'
import { GraphQLClient } from './util/GraphQLClient'
import { getTestTools } from './util/init'
import { getConfig } from '../../../shared/src/e2e/config'
import { ensureLoggedInOrCreateTestUser } from './util/helpers'
import {
    ensureTestExternalService,
    waitForRepos,
    setUserSiteAdmin,
    getUser,
    ensureNoTestExternalServices,
    getExternalServices,
} from './util/api'
import { Key } from 'ts-key-enum'
import { retry } from '../../../shared/src/e2e/e2e-test-utils'
import { ScreenshotVerifier } from './util/ScreenshotVerifier'
import { TestResourceManager } from './util/TestResourceManager'
import delay from 'delay'

const activationNavBarSelector = '.e2e-activation-nav-item-toggle'

/**
 * Gets the activation status for the current user from the GUI. There's no easy way to fetch this
 * from the API, so we use a coarse method for extracting it from the GUI.
 */
async function getActivationStatus(driver: Driver): Promise<{ complete: number; total: number }> {
    await driver.page.goto(driver.sourcegraphBaseUrl + '/search')
    await driver.page.waitForSelector(activationNavBarSelector)
    await driver.page.click(activationNavBarSelector)
    await delay(2000) // TODO: replace/delete
    return driver.page.evaluate(() => {
        const dropdownMenu = document.querySelector('.activation-dropdown')
        if (!dropdownMenu) {
            throw new Error('No activation status dropdown menu')
        }
        const lineItems = Array.from(dropdownMenu.querySelectorAll('.activation-dropdown-item'))
        const complete = lineItems.flatMap(el => Array.from(el.querySelectorAll('.mdi-icon.text-success'))).length
        const incomplete = lineItems.flatMap(el => Array.from(el.querySelectorAll('.mdi-icon.text-muted'))).length
        return {
            complete,
            total: complete + incomplete,
        }
    })
}

describe('Onboarding', () => {
    const config = getConfig(
        'sudoToken',
        'sudoUsername',
        'gitHubToken',
        'sourcegraphBaseUrl',
        'includeAdminOnboarding',
        'noCleanup',
        'testUserPassword',
        'headless',
        'slowMo',
        'logBrowserConsole',
        'logStatusMessages',
        'keepBrowser'
    )
    const testExternalServiceConfig = {
        kind: GQL.ExternalServiceKind.GITHUB,
        uniqueDisplayName: '[TEST] GitHub (onboarding.test.ts)',
        config: {
            url: 'https://github.com',
            token: config.gitHubToken,
            repos: ['auth0/go-jwt-middleware', 'kyoshidajp/ghkw', 'PalmStoneGames/kube-cert-manager'],
            repositoryQuery: ['none'],
        },
    }
    const testUsername = 'test-onboarding-regression-test-user'

    let driver: Driver
    let gqlClient: GraphQLClient
    let resourceManager: TestResourceManager
    let screenshots: ScreenshotVerifier
    beforeAll(
        async () => {
            ;({ driver, gqlClient, resourceManager } = await getTestTools(config))
            screenshots = new ScreenshotVerifier(driver)

            resourceManager.add(
                'User',
                testUsername,
                await ensureLoggedInOrCreateTestUser(driver, gqlClient, {
                    ...config,
                    username: testUsername,
                    deleteIfExists: true,
                })
            )
        },
        20 * 1000 // wait 20s for cloning
    )

    afterAll(async () => {
        if (!config.noCleanup) {
            await resourceManager.destroyAll()
        }
        if (driver) {
            await driver.close()
        }
        if (screenshots.screenshots.length > 0) {
            console.log(screenshots.verificationInstructions())
        }
    })

    /**
     * Only run site-admin onboarding test if the appropriate environment variable is set.
     * This test assumes an instance of Sourcegraph that does not yet have external services.
     */
    const testAdminOnboarding = config.includeAdminOnboarding ? test : test.skip
    testAdminOnboarding(
        'Site-admin onboarding',
        async () => {
            // TODO: need to destroy?
            await ensureNoTestExternalServices(gqlClient, {
                ...testExternalServiceConfig,
                deleteIfExist: true,
            })
            if ((await getExternalServices(gqlClient)).length > 0) {
                throw new Error(
                    'other external services exist and this test should be run on an instance with no user-created external services'
                )
            }

            const testUser = await getUser(gqlClient, 'test-onboarding-regression-test-user')
            if (!testUser) {
                throw new Error(`Could not obtain userID of user ${testUsername}`)
            }
            await setUserSiteAdmin(gqlClient, testUser.id, true)

            const activationStatus = await getActivationStatus(driver)
            if (activationStatus.total !== 4) {
                throw new Error(`Expected 4 onboarding steps for site admins, but only found ${activationStatus.total}`)
            }

            // Verify add-external-service onboarding step
            await driver.page.waitForSelector(activationNavBarSelector)
            // Check that onboarding status menu contains menu item and it goes where it should
            await retry(async () => {
                // for some reason, first click doesn't work here
                await driver.page.click(activationNavBarSelector)
                await driver.page.click(activationNavBarSelector)
                await (await driver.findElementWithText('Connect your code host')).click()
            })
            await driver.waitUntilURL(driver.sourcegraphBaseUrl + '/site-admin/external-services')
            // Verify confetti plays
            await driver.ensureHasExternalService({
                kind: testExternalServiceConfig.kind,
                displayName: testExternalServiceConfig.uniqueDisplayName,
                config: JSON.stringify(testExternalServiceConfig.config),
            })
            await delay(500) // wait for confetti to play a bit
            await screenshots.verifyScreenshot({
                filename: 'confetti-appears-after-adding-first-external-service.png',
                description: 'confetti coming out of "Setup" navbar item',
            })
        },
        30 * 1000
    )

    test(
        'Non-admin user onboarding',
        async () => {
            await ensureTestExternalService(gqlClient, testExternalServiceConfig, config)
            const repoSlugs = testExternalServiceConfig.config.repos
            await waitForRepos(gqlClient, ['github.com/' + repoSlugs[repoSlugs.length - 1]], config)

            const testUser = await getUser(gqlClient, testUsername)
            if (!testUser) {
                throw new Error(`Could not obtain userID of user ${testUsername}`)
            }
            await setUserSiteAdmin(gqlClient, testUser.id, false)

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
            await delay(500) // allow some time for confetti to play
            await screenshots.verifyScreenshot({
                filename: 'confetti-appears-after-first-search.png',
                description: 'confetti coming out of "Setup" navbar item',
            })
            await screenshots.verifySelector(
                'progress-bar-after-initial-search-is-half-green.png',
                '50% green circle',
                statusBarSelector
            )

            // Do a find references
            await driver.page.goto(
                config.sourcegraphBaseUrl + '/github.com/auth0/go-jwt-middleware/-/blob/jwtmiddleware.go'
            )
            // await driver.page.mouse.move(100, 100)
            const defTokenXPath =
                '//*[contains(@class, "blob-page__blob")]//span[starts-with(text(), "TokenExtractor")]'
            await driver.page.waitForXPath(defTokenXPath)
            const elems = await driver.page.$x(defTokenXPath)
            await Promise.all(elems.map(e => e.click()))
            await Promise.all(elems.map(elem => elem.dispose()))
            const findRefsSelector = '.e2e-tooltip-find-references'
            await driver.page.waitForSelector(findRefsSelector)
            await driver.page.click(findRefsSelector)
            await driver.page.waitForSelector('.e2e-search-result')

            await delay(500) // allow some time for confetti to play
            await screenshots.verifyScreenshot({
                filename: 'confetti-appears-after-find-refs.png',
                description: 'confetti coming out of "Setup" navbar item',
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
