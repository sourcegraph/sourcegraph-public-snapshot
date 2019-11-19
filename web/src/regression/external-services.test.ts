/**
 * @jest-environment node
 */

import * as jsoncEdit from '@sqs/jsonc-parser/lib/edit'
import { TestResourceManager } from './util/TestResourceManager'
import { GraphQLClient, createGraphQLClient } from './util/GraphQLClient'
import { Driver } from '../../../shared/src/e2e/driver'
import { getConfig } from '../../../shared/src/e2e/config'
import { getTestTools } from './util/init'
import { ensureLoggedInOrCreateTestUser, editCriticalSiteConfig, login, loginToGitHub } from './util/helpers'
import {
    setUserSiteAdmin,
    getUser,
    ensureNoTestExternalServices,
    ensureTestExternalService,
    getManagementConsoleState,
    waitForRepos,
    getExternalServices,
    updateExternalService,
} from './util/api'
import * as GQL from '../../../shared/src/graphql/schema'

describe('External services GUI', () => {
    const testUsername = 'test-extsvc'
    const config = getConfig(
        'sudoToken',
        'sudoUsername',
        'gitHubToken',
        'sourcegraphBaseUrl',
        'noCleanup',
        'testUserPassword',
        'logBrowserConsole',
        'slowMo',
        'headless',
        'keepBrowser',
        'logStatusMessages'
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

    test('External services: GitHub.com GUI and repositoryPathPattern', async () => {
        const externalServiceName = '[TEST] Regression test: GitHub.com'
        await ensureNoTestExternalServices(gqlClient, {
            kind: GQL.ExternalServiceKind.GITHUB,
            uniqueDisplayName: externalServiceName,
            deleteIfExist: true,
        })

        resourceManager.add(
            'External service',
            externalServiceName,
            await (async () => {
                await driver.page.goto(config.sourcegraphBaseUrl + '/site-admin/external-services')
                await (await driver.findElementWithText('Add external service', { wait: { timeout: 500 } })).click()
                await (
                    await driver.findElementWithText('Add GitHub.com repositories.', {
                        wait: { timeout: 500 },
                    })
                ).click()
                const repoSlugs = ['gorilla/mux']
                const githubConfig = `{
                    "url": "https://github.com",
                    "token": ${JSON.stringify(config.gitHubToken)},
                    "repos": ${JSON.stringify(repoSlugs)},
                    "repositoryQuery": ["none"],
                    "repos": ["gorilla/mux"],
                    "repositoryPathPattern": "github-prefix/{nameWithOwner}"
                }`
                await driver.replaceText({
                    selector: '#e2e-external-service-form-display-name',
                    newText: externalServiceName,
                    selectMethod: 'selectall',
                    enterTextMethod: 'paste',
                })
                await driver.replaceText({
                    selector: '.monaco-editor',
                    newText: githubConfig,
                    selectMethod: 'keyboard',
                    enterTextMethod: 'paste',
                })
                await (
                    await driver.findElementWithText('Add external service', {
                        selector: 'button',
                        wait: { timeout: 500 },
                    })
                ).click()
                return () =>
                    ensureNoTestExternalServices(gqlClient, {
                        kind: GQL.ExternalServiceKind.GITHUB,
                        uniqueDisplayName: externalServiceName,
                        deleteIfExist: true,
                    })
            })()
        )

        await waitForRepos(gqlClient, ['github-prefix/gorilla/mux'], config)
        const response = await driver.page.goto(config.sourcegraphBaseUrl + '/github-prefix/gorilla/mux')
        if (!response) {
            throw new Error('no response')
        }
        expect(response.status()).toBe(200)

        // Redirect
        await driver.page.goto(config.sourcegraphBaseUrl + '/github.com/gorilla/mux')
        await driver.waitUntilURL(config.sourcegraphBaseUrl + '/github-prefix/gorilla/mux')
    })
})

describe('External services API', () => {
    const config = getConfig(
        'sudoToken',
        'sudoUsername',
        'gitHubToken',
        'gitLabToken',
        'sourcegraphBaseUrl',
        'noCleanup',
        'testUserPassword',
        'logBrowserConsole',
        'slowMo',
        'headless',
        'keepBrowser',
        'logStatusMessages',
        'bitbucketCloudUserBobAppPassword'
    )

    const gqlClient = createGraphQLClient({
        baseUrl: config.sourcegraphBaseUrl,
        token: config.sudoToken,
        sudoUsername: config.sudoUsername,
    })
    const resourceManager = new TestResourceManager()
    afterAll(async () => {
        if (!config.noCleanup) {
            await resourceManager.destroyAll()
        }
    })

    test(
        'External services: GitLab',
        async () => {
            const externalService = {
                kind: GQL.ExternalServiceKind.GITLAB,
                uniqueDisplayName: '[TEST] Regression test: GitLab.com',
                config: {
                    url: 'https://gitlab.com',
                    token: config.gitLabToken,
                    projectQuery: ['none'],
                    projects: [
                        {
                            name: 'ase/ase',
                        },
                    ],
                },
            }
            const repos = ['gitlab.com/ase/ase']
            await ensureNoTestExternalServices(gqlClient, { ...externalService, deleteIfExist: true })
            await waitForRepos(gqlClient, repos, config, true)
            resourceManager.add(
                'External service',
                externalService.uniqueDisplayName,
                await ensureTestExternalService(gqlClient, { ...externalService, waitForRepos: repos }, config)
            )
        },
        5 * 1000
    )

    test(
        'External services: Bitbucket Cloud',
        async () => {
            const uniqueDisplayName = '[TEST] Regression test: Bitbucket Cloud (bitbucket.org)'
            const externalServiceInput = {
                kind: GQL.ExternalServiceKind.BITBUCKETCLOUD,
                uniqueDisplayName,
                config: {
                    url: 'https://bitbucket.org',
                    username: 'sg-e2e-regression-test-bob',
                    appPassword: config.bitbucketCloudUserBobAppPassword,
                    repositoryPathPattern: '{nameWithOwner}',
                },
            }
            const repos = [
                'sg-e2e-regression-test-bob/jsonrpc2',
                'sg-e2e-regression-test-bob/codeintellify',
                'sg-e2e-regression-test-bob/mux',
            ]
            await ensureNoTestExternalServices(gqlClient, { ...externalServiceInput, deleteIfExist: true })
            await waitForRepos(gqlClient, repos, config, true)
            resourceManager.add(
                'External service',
                uniqueDisplayName,
                await ensureTestExternalService(gqlClient, { ...externalServiceInput, waitForRepos: repos }, config)
            )
            // Update eternal service with an "exclude" property
            const { id } = (await getExternalServices(gqlClient, { uniqueDisplayName }))[0]
            await updateExternalService(gqlClient, {
                id,
                config: JSON.stringify({
                    ...externalServiceInput.config,
                    exclude: [{ name: 'sg-e2e-regression-test-bob/jsonrpc2' }],
                }),
            })
            // Check that the excluded repository is no longer synced
            await waitForRepos(gqlClient, ['sg-e2e-regression-test-bob/jsonrpc2'], config, true)
        },
        30 * 1000
    )
})

describe('External services permissions', () => {
    const formattingOptions = { eol: '\n', insertSpaces: true, tabSize: 2 }
    const config = getConfig(
        'sudoToken',
        'sudoUsername',
        'sourcegraphBaseUrl',
        'noCleanup',
        'logBrowserConsole',
        'slowMo',
        'headless',
        'keepBrowser',
        'logStatusMessages',
        'gitHubUserAmyPassword',
        'gitHubUserBobToken',
        'gitHubClientID',
        'gitHubClientSecret',
        'managementConsoleUrl'
    )
    let driver: Driver
    let gqlClient: GraphQLClient
    let resourceManager: TestResourceManager
    let managementConsolePassword: string
    beforeAll(async () => {
        ;({ driver, gqlClient, resourceManager } = await getTestTools(config))
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
        'External services permissions: GitHub',
        async () => {
            const externalService = {
                kind: GQL.ExternalServiceKind.GITHUB,
                uniqueDisplayName: '[TEST] Regression test: GitHub.com permissions',
                config: {
                    url: 'https://github.com',
                    token: config.gitHubUserBobToken,
                    repositoryQuery: ['affiliated'],
                    authorization: {},
                },
            }
            const repos = [
                'github.com/sg-e2e-regression-test-bob/about',
                'github.com/sg-e2e-regression-test-bob/shared-with-amy',
            ]
            await ensureNoTestExternalServices(gqlClient, { ...externalService, deleteIfExist: true })
            resourceManager.add(
                'External service',
                externalService.uniqueDisplayName,
                await ensureTestExternalService(gqlClient, { ...externalService, waitForRepos: repos }, config)
            )

            const authProvider = {
                type: 'github',
                allowSignup: true,
                clientID: config.gitHubClientID,
                clientSecret: config.gitHubClientSecret,
                displayName: '[TEST] GitHub.com permissions',
                url: 'https://github.com/',
            }
            resourceManager.add(
                'Authentication provider',
                authProvider.displayName,
                await editCriticalSiteConfig(config.managementConsoleUrl, managementConsolePassword, contents =>
                    jsoncEdit.setProperty(contents, ['auth.providers', -1], authProvider, formattingOptions)
                )
            )

            // Ensure user sg-e2e-regression-test-amy exists
            await login(driver, { ...config, authProviderDisplayName: authProvider.displayName }, () =>
                loginToGitHub(driver, 'sg-e2e-regression-test-amy', config.gitHubUserAmyPassword)
            )

            {
                const response = await driver.page.goto(
                    config.sourcegraphBaseUrl + '/github.com/sg-e2e-regression-test-bob/shared-with-amy'
                )
                if (!response) {
                    throw new Error('no response')
                }
                expect(response.status()).toBe(200)
            }
            await driver.findElementWithText('sg-e2e-regression-test-bob/shared-with-amy', {
                wait: { timeout: 2 * 1000 },
            })

            {
                const response = await driver.page.goto(
                    config.sourcegraphBaseUrl + '/github.com/sg-e2e-regression-test-bob/about'
                )
                if (!response) {
                    throw new Error('no response')
                }
                expect(response.status()).toBe(404)
            }
            await driver.findElementWithText('Repository not found', {
                wait: { timeout: 2 * 1000 },
            })
        },
        30 * 1000
    )
})
