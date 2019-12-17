/**
 * @jest-environment node
 */

import { TestResourceManager } from './util/TestResourceManager'
import { GraphQLClient } from './util/GraphQLClient'
import { Driver } from '../../../shared/src/e2e/driver'
import { getConfig } from '../../../shared/src/e2e/config'
import { getTestTools } from './util/init'
import {
    ensureLoggedInOrCreateTestUser,
    login,
    loginToOkta,
    loginToGitHub,
    loginToGitLab,
    createAuthProvider,
} from './util/helpers'
import { setUserSiteAdmin, getUser } from './util/api'
import {
    GitHubAuthProvider,
    GitLabAuthProvider,
    SAMLAuthProvider,
    OpenIDConnectAuthProvider,
} from '../schema/site.schema'

const oktaUserAmy = 'beyang+sg-e2e-regression-test-amy@sourcegraph.com'

async function testLogin(
    driver: Driver,
    gqlClient: GraphQLClient,
    resourceManager: TestResourceManager,
    {
        sourcegraphBaseUrl,
        authProvider,
        loginToAuthProvider,
    }: {
        sourcegraphBaseUrl: string

        authProvider: (GitHubAuthProvider | GitLabAuthProvider | SAMLAuthProvider | OpenIDConnectAuthProvider) & {
            displayName: string
        }
        loginToAuthProvider: () => Promise<void>
    }
): Promise<void> {
    resourceManager.add(
        'Authentication provider',
        authProvider.displayName,
        await createAuthProvider(gqlClient, authProvider)
    )
    await login(driver, { sourcegraphBaseUrl, authProviderDisplayName: authProvider.displayName }, loginToAuthProvider)

    await (await driver.page.waitForSelector('.e2e-user-nav-item-toggle')).click()
    await (await driver.findElementWithText('Sign out', { wait: { timeout: 2000 } })).click()
    await driver.findElementWithText('Signed out of Sourcegraph', { wait: { timeout: 2000 } })
    await driver.page.goto(sourcegraphBaseUrl)
    await driver.findElementWithText('Sign in', { wait: { timeout: 5000 } })
    await driver.findElementWithText('Forgot password?', { wait: { timeout: 5000 } })
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
            await testLogin(driver, gqlClient, resourceManager, {
                ...config,
                authProvider: {
                    type: 'github',
                    displayName: '[TEST] GitHub.com',
                    clientID: config.gitHubClientID,
                    clientSecret: config.gitHubClientSecret,
                    allowSignup: true,
                },
                loginToAuthProvider: () =>
                    loginToGitHub(driver, 'sg-e2e-regression-test-amy', config.gitHubUserAmyPassword),
            })
        },
        20 * 1000
    )

    test(
        'Sign in with GitLab',
        async () => {
            await testLogin(driver, gqlClient, resourceManager, {
                ...config,
                authProvider: {
                    type: 'gitlab',
                    displayName: '[TEST] GitLab.com',
                    clientID: config.gitLabClientID,
                    clientSecret: config.gitLabClientSecret,
                },
                loginToAuthProvider: () =>
                    loginToGitLab(driver, 'sg-e2e-regression-test-amy', config.gitLabUserAmyPassword),
            })
        },
        20 * 1000
    )

    test(
        'Sign in with Okta SAML',
        async () => {
            await testLogin(driver, gqlClient, resourceManager, {
                ...config,
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
            await testLogin(driver, gqlClient, resourceManager, {
                ...config,
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
