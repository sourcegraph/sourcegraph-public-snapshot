/**
 * @jest-environment node
 */

import { TestResourceManager } from './util/TestResourceManager'
import { GraphQLClient, createGraphQLClient } from './util/GraphQLClient'
import { Driver } from '../../../shared/src/e2e/driver'
import { getConfig } from '../../../shared/src/e2e/config'
import { getTestTools } from './util/init'
import { ensureLoggedInOrCreateTestUser } from './util/helpers'
import {
    setUserSiteAdmin,
    getUser,
    ensureNoTestExternalServices,
    ensureTestExternalService,
    waitForRepos,
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
        'keepBrowser'
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

    test('External services: GitHub.com GUI', async () => {
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
                await (await driver.findElementWithText('Add GitHub.com repositories.', {
                    wait: { timeout: 500 },
                })).click()
                const repoSlugs = ['gorilla/mux']
                const githubConfig = `{
                    "url": "https://github.com",
                    "token": ${JSON.stringify(config.gitHubToken)},
                    "repos": ${JSON.stringify(repoSlugs)},
                    "repositoryQuery": ["none"],
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
                await (await driver.findElementWithText('Add external service', {
                    selector: 'button',
                    wait: { timeout: 500 },
                })).click()
                return () =>
                    ensureNoTestExternalServices(gqlClient, {
                        kind: GQL.ExternalServiceKind.GITHUB,
                        uniqueDisplayName: externalServiceName,
                        deleteIfExist: true,
                    })
            })()
        )
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
        'logStatusMessages'
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
})
