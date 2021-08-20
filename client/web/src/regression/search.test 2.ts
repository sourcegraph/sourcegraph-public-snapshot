import { describe, test } from 'mocha'

import * as GQL from '@sourcegraph/shared/src/graphql/schema'
import { getConfig } from '@sourcegraph/shared/src/testing/config'
import { Driver } from '@sourcegraph/shared/src/testing/driver'
import { afterEachSaveScreenshotIfFailed } from '@sourcegraph/shared/src/testing/screenshotReporter'

import { ensureTestExternalService } from './util/api'
import { GraphQLClient } from './util/GraphQlClient'
import { ensureLoggedInOrCreateTestUser } from './util/helpers'
import { getTestTools } from './util/init'
import { TestResourceManager } from './util/TestResourceManager'

describe('Search regression test suite', () => {
    /**
     * Test data
     */
    const testUsername = 'test-search'
    const testExternalServiceInfo = {
        kind: GQL.ExternalServiceKind.GITHUB,
        uniqueDisplayName: '[TEST] GitHub (search.test.ts)',
    }
    const testRepoSlugs = ['sgtest/jsonrpc2', 'sgtest/go-diff']
    const config = getConfig(
        'sudoToken',
        'sudoUsername',
        'gitHubToken',
        'sourcegraphBaseUrl',
        'noCleanup',
        'testUserPassword',
        'logStatusMessages',
        'logBrowserConsole',
        'slowMo',
        'headless',
        'keepBrowser'
    )

    describe('Search over a dozen repositories', () => {
        let driver: Driver
        let gqlClient: GraphQLClient
        let resourceManager: TestResourceManager
        before(async function () {
            this.timeout(10 * 60 * 1000 + 30 * 1000)
            ;({ driver, gqlClient, resourceManager } = await getTestTools(config))
            resourceManager.add(
                'User',
                testUsername,
                await ensureLoggedInOrCreateTestUser(driver, gqlClient, { username: testUsername, ...config })
            )
            resourceManager.add(
                'External service',
                testExternalServiceInfo.uniqueDisplayName,
                await ensureTestExternalService(
                    gqlClient,
                    {
                        ...testExternalServiceInfo,
                        config: {
                            url: 'https://github.com',
                            token: config.gitHubToken,
                            repos: testRepoSlugs,
                            repositoryQuery: ['none'],
                        },
                        waitForRepos: testRepoSlugs.map(slug => 'github.com/' + slug),
                    },
                    { ...config, timeout: 6 * 60 * 1000, indexed: true }
                )
            )
        })

        afterEachSaveScreenshotIfFailed(() => driver.page)

        after(async () => {
            if (!config.noCleanup) {
                await resourceManager.destroyAll()
            }
            if (driver) {
                await driver.close()
            }
        })

        test('Performs a search and displays results', async () => {
            await driver.page.goto(config.sourcegraphBaseUrl + '/search?q=fmt.Sprintf')
            await driver.page.waitForFunction(() => document.querySelectorAll('.test-search-result').length > 0)
        })
    })
})
