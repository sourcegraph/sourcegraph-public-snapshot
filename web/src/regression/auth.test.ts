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

const oktaUserAmy = 'beyang+sg-e2e-regression-test-amy@sourcegraph.com'

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
    await driver.newPage()
    await driver.page.goto(sourcegraphBaseUrl)
    await driver.page.reload()
    await (await driver.findElementWithText('Sign in with ' + authProvider.displayName, {
        tagName: 'a',
        wait: { timeout: 5000 },
    })).click()
    await driver.page.waitForNavigation()
    if (driver.page.url() !== sourcegraphBaseUrl + '/search') {
        await loginToAuthProvider()
        try {
            await driver.page.waitForFunction(
                url => document.location.href === url,
                { timeout: 5 * 1000 },
                sourcegraphBaseUrl + '/search'
            )
        } catch (err) {
            throw new Error('unsuccessful login')
        }
    }

    await (await driver.page.waitForSelector('.e2e-user-nav-item-toggle')).click()
    await (await driver.findElementWithText('Sign out', { wait: { timeout: 2000 } })).click()
    await driver.findElementWithText('Signed out of Sourcegraph', { wait: { timeout: 2000 } })
    await driver.page.goto(sourcegraphBaseUrl)
    // >>>>>> TODO: this assumes > 1 auth provider
    await driver.findElementWithText('Sign in', { wait: { timeout: 5000 } })
    await driver.findElementWithText('Forgot password?', { wait: { timeout: 5000 } })
}

async function loginToOkta(driver: Driver, username: string, password: string): Promise<void> {
    await driver.page.waitForSelector('#okta-signin-username')
    await driver.replaceText({
        selector: '#okta-signin-username',
        newText: username,
    })
    await driver.replaceText({
        selector: '#okta-signin-password',
        newText: password,
    })
    await (await driver.page.waitForSelector('#okta-signin-submit')).click()
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
        'gitLabUserAmyPassword',
        'oktaUserAmyPassword',
        'oktaMetadataUrl'
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

    test(
        'Sign in via GitHub',
        async () => {
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
        },
        20 * 1000
    )

    test(
        'Sign in with GitLab',
        async () => {
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
        },
        20 * 1000
    )

    test(
        'Sign in with Okta SAML',
        async () => {
            await testLogin(driver, resourceManager, {
                ...config,
                managementConsolePassword,
                authProvider: {
                    type: 'saml',
                    displayName: '[TEST] Okta SAML',
                    identityProviderMetadataURL: config.oktaMetadataUrl,
                },
                loginToAuthProvider: () => loginToOkta(driver, oktaUserAmy, config.oktaUserAmyPassword),
            })
        },
        20 * 1000
    )

    test(
        'Sign in with Okta OpenID Connect',
        async () => {
            await testLogin(driver, resourceManager, {
                ...config,
                managementConsolePassword,
                authProvider: {
                    type: 'openidconnect',
                    displayName: '[TEST] Okta OpenID Connect',
                    issuer: 'https://dev-433675.oktapreview.com',
                    clientID: '0oao8w32qpPNB8tnU0h7',
                    clientSecret: 'pHCg8h8Dr0yaBzBEqBGM4NWjXSAzLqp8OtcYGUqA',
                    requireEmailDomain: 'sourcegraph.com',
                },
                loginToAuthProvider: () => loginToOkta(driver, oktaUserAmy, config.oktaUserAmyPassword),
            })
        },
        20 * 1000
    )
})
