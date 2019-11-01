/**
 * @jest-environment node
 */

import { TestResourceManager } from './util/TestResourceManager'
import { GraphQLClient } from './util/GraphQLClient'
import { Driver } from '../../../shared/src/e2e/driver'
import { getConfig } from '../../../shared/src/e2e/config'
import { getTestTools } from './util/init'
import { ensureLoggedInOrCreateTestUser, createAuthProviderGUI } from './util/helpers'
import { setUserSiteAdmin, getUser, getManagementConsoleState } from './util/api'
import {
    GitHubAuthProvider,
    GitLabAuthProvider,
    SAMLAuthProvider,
    OpenIDConnectAuthProvider,
} from '../schema/critical.schema'

async function testLogin(
    driver: Driver,
    resourceManager: TestResourceManager,
    {
        sourcegraphBaseUrl,
        managementConsoleUrl,
        managementConsolePassword,
        authProvider,
        loginToAuthProvider,
    }: {
        sourcegraphBaseUrl: string
        managementConsoleUrl: string
        managementConsolePassword: string
        authProvider: (GitHubAuthProvider | GitLabAuthProvider | SAMLAuthProvider | OpenIDConnectAuthProvider) & {
            displayName: string
        }
        loginToAuthProvider: () => Promise<void>
    }
): Promise<void> {
    resourceManager.add(
        'Authentication provider',
        authProvider.displayName,
        await createAuthProviderGUI(driver, managementConsoleUrl, managementConsolePassword, authProvider)
    )
    await driver.page.goto(sourcegraphBaseUrl + '/-/sign-out')
    await driver.page.goto(sourcegraphBaseUrl)
    await driver.page.reload()
    await driver.page.waitForNavigation()
    await (await driver.findElementWithText('Sign in with ' + authProvider.displayName, {
        tagName: 'a',
        wait: true,
    })).click()
    await loginToAuthProvider()
    try {
        await driver.page.waitForFunction(
            url => document.location.href === url,
            { timeout: 2000 },
            sourcegraphBaseUrl + '/search'
        )
    } catch (err) {
        throw new Error('unsuccessful login')
    }
}

describe('Auth regression test suite', () => {
    const testUsername = 'test-auth'
    const config = getConfig(
        'sudoToken',
        'sudoUsername',
        'headless',
        'slowMo',
        'keepBrowser',
        'gitHubToken',
        'sourcegraphBaseUrl',
        'noCleanup',
        'testUserPassword',
        'logBrowserConsole',
        'managementConsoleUrl',
        'gitHubClientID',
        'gitHubClientSecret',
        'gitHubUserAmyPassword',
        'gitLabClientID',
        'gitLabClientSecret',
        'gitLabUserAmyPassword'
    )

    let driver: Driver
    let gqlClient: GraphQLClient
    let resourceManager: TestResourceManager
    let managementConsolePassword: string
    beforeAll(async () => {
        ;({ driver, gqlClient, resourceManager } = await getTestTools(config))
        resourceManager.add(
            'User',
            testUsername,
            await ensureLoggedInOrCreateTestUser(driver, gqlClient, {
                username: testUsername,
                deleteIfExists: true,
                ...config,
            })
        )
        const user = await getUser(gqlClient, testUsername)
        if (!user) {
            throw new Error(`test user ${testUsername} does not exist`)
        }
        await setUserSiteAdmin(gqlClient, user.id, true)

        const { plaintextPassword } = await getManagementConsoleState(gqlClient)
        if (!plaintextPassword) {
            throw new Error('empty management console password')
        }
        managementConsolePassword = plaintextPassword
    })

    afterAll(async () => {
        if (!config.noCleanup) {
            await resourceManager.destroyAll()
        }
        if (driver) {
            await driver.close()
        }
    }, 10 * 1000)

    test('Sign in via GitHub', async () => {
        await testLogin(driver, resourceManager, {
            ...config,
            managementConsolePassword,
            authProvider: {
                type: 'github',
                displayName: '[TEST] GitHub.com',
                clientID: config.gitHubClientID,
                clientSecret: config.gitHubClientSecret,
                allowSignup: true,
            },
            loginToAuthProvider: async () => {
                await driver.page.waitForSelector('#login_field')
                await driver.replaceText({
                    selector: '#login_field',
                    newText: 'sg-e2e-regression-test-amy',
                    selectMethod: 'keyboard',
                    enterTextMethod: 'paste',
                })
                await driver.replaceText({
                    selector: '#password',
                    newText: config.gitHubUserAmyPassword,
                    selectMethod: 'keyboard',
                    enterTextMethod: 'paste',
                })
                await driver.page.keyboard.press('Enter')
            },
        })
    })

    test('Sign in with GitLab', async () => {
        await testLogin(driver, resourceManager, {
            ...config,
            managementConsolePassword,
            authProvider: {
                type: 'gitlab',
                displayName: '[TEST] GitLab.com',
                clientID: config.gitLabClientID,
                clientSecret: config.gitLabClientSecret,
                allowSignup: true,
            },
            loginToAuthProvider: async () => {
                await driver.page.waitForSelector('input[name="user[login]"]', { timeout: 2000 })
                await driver.replaceText({
                    selector: '#user_login',
                    newText: 'sg-e2e-regression-test-amy',
                })
                await driver.replaceText({
                    selector: '#user_password',
                    newText: config.gitLabUserAmyPassword,
                })
                await (await driver.page.waitForSelector('input[data-qa-selector="sign_in_button"]')).click()
            },
        })
    })
})
