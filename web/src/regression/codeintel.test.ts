/**
 * @jest-environment node
 */

import * as GQL from '../../../shared/src/graphql/schema'
import { Driver } from '../../../shared/src/e2e/driver'
import { enableLSIF, uploadAndEnsure, disableLSIF, clearUploads } from './util/codeintel'
import { ensureLoggedInOrCreateTestUser } from './util/helpers'
import { ensureTestExternalService, getUser, setUserSiteAdmin } from './util/api'
import { getConfig } from '../../../shared/src/e2e/config'
import { getTestTools } from './util/init'
import { GraphQLClient } from './util/GraphQLClient'
import { testCodeNavigation } from './util/codenav'
import { TestResourceManager } from './util/TestResourceManager'

describe('Code intelligence regression test suite', () => {
    const prometheusCommonHeadCommit = 'b5fe7d854c42dc7842e48d1ca58f60feae09d77b' // HEAD
    const prometheusCommonLSIFCommit = '287d3e634a1e550c9e463dd7e5a75a422c614505' // 2 behind HEAD
    const prometheusCommonFallbackCommit = 'e8215224146358493faab0295ce364cd386223b9' // 2 behind LSIF
    const prometheusClientHeadCommit = '333f01cef0d61f9ef05ada3d94e00e69c8d5cdda'
    const prometheusRedefinitionsHeadCommit = 'c68f0e063cf8a98e7ce3428cfd50588746010f1f'

    const testRepoSlugs = [
        'sourcegraph/sourcegraph',
        'sourcegraph-testing/prometheus-common',
        'sourcegraph-testing/prometheus-client-golang',
        'sourcegraph-testing/prometheus-redefinitions',
    ]

    const testUsername = 'test-sg-codeintel'
    const config = getConfig(
        'gitHubToken',
        'headless',
        'keepBrowser',
        'logBrowserConsole',
        'logStatusMessages',
        'noCleanup',
        'slowMo',
        'sourcegraphBaseUrl',
        'sudoToken',
        'sudoUsername',
        'testUserPassword'
    )
    const testExternalServiceInfo = {
        kind: GQL.ExternalServiceKind.GITHUB,
        uniqueDisplayName: '[TEST] GitHub (codeintel.test.ts)',
    }

    let driver: Driver
    let gqlClient: GraphQLClient
    let outerResourceManager: TestResourceManager
    beforeAll(async () => {
        ;({ driver, gqlClient, resourceManager: outerResourceManager } = await getTestTools(config))
        outerResourceManager.add(
            'User',
            testUsername,
            await ensureLoggedInOrCreateTestUser(driver, gqlClient, {
                username: testUsername,
                deleteIfExists: true,
                ...config,
            })
        )
        outerResourceManager.add(
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
                    waitForRepos: testRepoSlugs.map(r => `github.com/${r}`),
                },
                config
            )
        )

        const user = await getUser(gqlClient, testUsername)
        if (!user) {
            throw new Error(`test user ${testUsername} does not exist`)
        }
        await setUserSiteAdmin(gqlClient, user.id, true)
    }, 30 * 1000)

    afterAll(async () => {
        if (!config.noCleanup) {
            await outerResourceManager.destroyAll()
        }
        if (driver) {
            await driver.close()
        }
    })

    describe('Basic code intelligence regression test suite', () => {
        const innerResourceManager = new TestResourceManager()
        beforeAll(async () => {
            innerResourceManager.add('Global setting', 'codeIntel.lsif', await disableLSIF(gqlClient))
        })
        afterAll(async () => {
            if (!config.noCleanup) {
                await innerResourceManager.destroyAll()
            }
        })

        test('Definitions, references, and hovers', async () => {
            const makeTestCase = (
                repository: string,
                sourceCommit: string,
                /** The commit to use for the definition (changes with nearest commit). */
                commonCommit: string,
                path: string,
                line: number
            ) => ({
                repoRev: `github.com/sourcegraph-testing/${repository}@${sourceCommit}`,
                files: [
                    {
                        path,
                        locations: [
                            {
                                line,
                                token: 'SamplePair',
                                precise: false,
                                expectedHoverContains: 'SamplePair pairs a SampleValue with a Timestamp.',
                                expectedDefinition: [
                                    {
                                        url: `/github.com/sourcegraph-testing/prometheus-common@${commonCommit}/-/blob/model/value.go#L78:1`,
                                        precise: false,
                                    },
                                    {
                                        url: `/github.com/sourcegraph-testing/prometheus-redefinitions@${prometheusRedefinitionsHeadCommit}/-/blob/sample.go#L7:1`,
                                        precise: false,
                                    },
                                ],
                                expectedReferences: [],
                            },
                        ],
                    },
                ],
            })

            await testCodeNavigation(driver, config, [
                makeTestCase(
                    'prometheus-client-golang',
                    prometheusClientHeadCommit,
                    prometheusCommonHeadCommit,
                    '/api/prometheus/v1/api.go',
                    41
                ),
            ])
        })

        test(
            'File sidebar, multiple levels of directories',
            async () => {
                await driver.page.goto(
                    config.sourcegraphBaseUrl +
                        '/github.com/sourcegraph/sourcegraph@c543dfd3936019befe94b881ade89e637d1a3dc3'
                )
                for (const file of ['cmd', 'frontend', 'auth', 'providers', 'providers.go']) {
                    await driver.findElementWithText(file, {
                        action: 'click',
                        selector: '.e2e-repo-rev-sidebar a',
                        wait: { timeout: 2 * 1000 },
                    })
                }
                await driver.waitUntilURL(
                    `${config.sourcegraphBaseUrl}/github.com/sourcegraph/sourcegraph@c543dfd3936019befe94b881ade89e637d1a3dc3/-/blob/cmd/frontend/auth/providers/providers.go`,
                    { timeout: 2 * 1000 }
                )
            },
            20 * 1000
        )

        test('Symbols sidebar', async () => {
            await driver.page.goto(
                config.sourcegraphBaseUrl +
                    '/github.com/sourcegraph/sourcegraph@c543dfd3936019befe94b881ade89e637d1a3dc3'
            )
            await driver.findElementWithText('SYMBOLS', {
                action: 'click',
                selector: '.e2e-repo-rev-sidebar button',
                wait: { timeout: 10 * 1000 },
            })
            await driver.findElementWithText('backgroundEntry', {
                action: 'click',
                selector: '.e2e-repo-rev-sidebar a span',
                wait: { timeout: 2 * 1000 },
            })
            await driver.replaceText({
                selector: 'input[placeholder="Search symbols..."]',
                newText: 'buildentry',
            })
            await driver.page.waitForFunction(
                () => {
                    const sidebar = document.querySelector<HTMLElement>('.e2e-repo-rev-sidebar')
                    return sidebar && !sidebar.innerText.includes('backgroundEntry')
                },
                {
                    timeout: 2 * 1000,
                }
            )
            await driver.findElementWithText('buildEntry', {
                action: 'click',
                selector: '.e2e-repo-rev-sidebar a span',
                wait: { timeout: 2 * 1000 },
            })
            await driver.waitUntilURL(
                `${config.sourcegraphBaseUrl}/github.com/sourcegraph/sourcegraph@c543dfd3936019befe94b881ade89e637d1a3dc3/-/blob/browser/config/webpack/base.config.ts#L6:7-6:17`,
                { timeout: 2 * 1000 }
            )
        })
    })

    describe('Precise code intelligence regression test suite', () => {
        const innerResourceManager = new TestResourceManager()
        beforeAll(async () => {
            for (const { repo, commit } of [
                { repo: 'prometheus-common', commit: prometheusCommonLSIFCommit },
                { repo: 'prometheus-client-golang', commit: prometheusClientHeadCommit },
            ]) {
                innerResourceManager.add(
                    'LSIF upload',
                    `${repo} upload`,
                    await uploadAndEnsure(
                        driver,
                        config,
                        gqlClient,
                        `github.com/sourcegraph-testing/${repo}`,
                        commit,
                        '/'
                    )
                )
            }

            await clearUploads(gqlClient, 'github.com/sourcegraph-testing/prometheus-redefinitions')
            innerResourceManager.add('Global setting', 'codeIntel.lsif', await enableLSIF(gqlClient))
        }, 30 * 1000)
        afterAll(async () => {
            if (!config.noCleanup) {
                await innerResourceManager.destroyAll()
            }
        })

        test(
            'Cross-repository definitions, references, and hovers',
            async () => {
                interface Location {
                    path: string
                    position: { line: number; character: number }
                }

                const commonLocations = [
                    { path: 'model/value.go', position: { line: 31, character: 19 } },
                    { path: 'model/value.go', position: { line: 78, character: 6 } },
                    { path: 'model/value.go', position: { line: 84, character: 9 } },
                    { path: 'model/value.go', position: { line: 97, character: 10 } },
                    { path: 'model/value.go', position: { line: 104, character: 10 } },
                    { path: 'model/value.go', position: { line: 108, character: 9 } },
                    { path: 'model/value.go', position: { line: 137, character: 43 } },
                    { path: 'model/value.go', position: { line: 147, character: 10 } },
                    { path: 'model/value.go', position: { line: 163, character: 10 } },
                    { path: 'model/value.go', position: { line: 225, character: 11 } },
                    { path: 'model/value_test.go', position: { line: 133, character: 9 } },
                    { path: 'model/value_test.go', position: { line: 137, character: 11 } },
                    { path: 'model/value_test.go', position: { line: 156, character: 10 } },

                    // These (should) also come back from the backend, but the UI merges
                    // these into one of the results above.
                    // { path: 'model/value.go', position: { line: 104, character: 31 } },
                    // { path: 'model/value.go', position: { line: 150, character: 10 } },
                    // { path: 'model/value.go', position: { line: 166, character: 10 } },
                ]

                const clientLocations = [
                    { path: 'api/prometheus/v1/api.go', position: { line: 41, character: 15 } },
                    { path: 'api/prometheus/v1/api.go', position: { line: 70, character: 17 } },
                    { path: 'api/prometheus/v1/api_test.go', position: { line: 1119, character: 18 } },
                    { path: 'api/prometheus/v1/api_test.go', position: { line: 1123, character: 20 } },
                    { path: 'api/prometheus/v1/api_test.go', position: { line: 1127, character: 20 } },
                    { path: 'api/prometheus/v1/api_test.go', position: { line: 1131, character: 20 } },
                    { path: 'api/prometheus/v1/api_test.go', position: { line: 1135, character: 20 } },
                    { path: 'api/prometheus/v1/api_test.go', position: { line: 1139, character: 20 } },
                    { path: 'api/prometheus/v1/api_test.go', position: { line: 1143, character: 20 } },
                    { path: 'api/prometheus/v1/api_test.go', position: { line: 1147, character: 20 } },
                    { path: 'api/prometheus/v1/api_test.go', position: { line: 1151, character: 20 } },
                    { path: 'api/prometheus/v1/api_test.go', position: { line: 1155, character: 20 } },
                    { path: 'api/prometheus/v1/api_test.go', position: { line: 1159, character: 20 } },
                    { path: 'api/prometheus/v1/api_test.go', position: { line: 1163, character: 20 } },
                    { path: 'api/prometheus/v1/api_test.go', position: { line: 1167, character: 20 } },
                    { path: 'api/prometheus/v1/api_test.go', position: { line: 1171, character: 20 } },
                    { path: 'api/prometheus/v1/api_test.go', position: { line: 1175, character: 20 } },
                    { path: 'api/prometheus/v1/api_test.go', position: { line: 1179, character: 20 } },
                    { path: 'api/prometheus/v1/api_test.go', position: { line: 1197, character: 17 } },
                    { path: 'api/prometheus/v1/api_bench_test.go', position: { line: 34, character: 26 } },

                    // These (should) also come back from the backend, but the UI merges
                    // these into one of the results above.
                    // { path: 'api/prometheus/v1/api_bench_test.go', position: { line: 37, character: 24 } },
                ]

                const redefinitionLocations = [
                    { path: 'sample.go', position: { line: 7, character: 6 } },
                    { path: 'sample.go', position: { line: 12, character: 10 } },
                    { path: 'sample.go', position: { line: 16, character: 10 } },
                ]

                const makePath = (
                    repository: string,
                    commit: string,
                    { path, position: { line, character } }: Location
                ) => `/github.com/sourcegraph-testing/${repository}@${commit}/-/blob/${path}#L${line}:${character}`

                const makeTestCase = (
                    repository: string,
                    sourceCommit: string,
                    /** The commit to use for the definition (changes with nearest commit). */
                    commonCommit: string,
                    path: string,
                    line: number
                ) => ({
                    repoRev: `github.com/sourcegraph-testing/${repository}@${sourceCommit}`,
                    files: [
                        {
                            path,
                            locations: [
                                {
                                    line,
                                    token: 'SamplePair',
                                    precise: true,
                                    expectedHoverContains: 'SamplePair pairs a SampleValue with a Timestamp.',
                                    expectedDefinition: {
                                        url: `/github.com/sourcegraph-testing/prometheus-common@${commonCommit}/-/blob/model/value.go#L78:6`,
                                        precise: true,
                                    },
                                    expectedReferences: commonLocations
                                        .map(r => ({
                                            url: makePath('prometheus-common', commonCommit, r),
                                            precise: true,
                                        }))
                                        .concat(
                                            clientLocations.map(r => ({
                                                url: makePath(
                                                    'prometheus-client-golang',
                                                    prometheusClientHeadCommit,
                                                    r
                                                ),
                                                precise: true,
                                            }))
                                        )
                                        .concat(
                                            redefinitionLocations.map(r => ({
                                                url: makePath(
                                                    'prometheus-redefinitions',
                                                    prometheusRedefinitionsHeadCommit,
                                                    r
                                                ),
                                                precise: false,
                                            }))
                                        ),
                                },
                            ],
                        },
                    ],
                })

                await testCodeNavigation(driver, config, [
                    makeTestCase(
                        'prometheus-common',
                        prometheusCommonLSIFCommit,
                        prometheusCommonLSIFCommit,
                        '/model/value.go',
                        31
                    ),
                    makeTestCase(
                        'prometheus-common',
                        prometheusCommonLSIFCommit,
                        prometheusCommonLSIFCommit,
                        '/model/value_test.go',
                        133
                    ),
                    makeTestCase(
                        'prometheus-client-golang',
                        prometheusClientHeadCommit,
                        prometheusCommonLSIFCommit,
                        '/api/prometheus/v1/api.go',
                        41
                    ),
                    // makeTestCase('prometheus-client-golang', clientCommit, commonCommit, '/api/prometheus/v1/api_test.go', 1119),
                    // makeTestCase('prometheus-client-golang', clientCommit, commonCommit, '/api/prometheus/v1/api_bench_test.go', 34),
                    makeTestCase(
                        'prometheus-common',
                        prometheusCommonFallbackCommit,
                        prometheusCommonLSIFCommit,
                        '/model/value.go',
                        31
                    ),
                    // makeTestCase('prometheus-common', olderCommonCommit, olderCommonCommit, '/model/value.go', 31),
                ])
            },
            60 * 1000
        )
    })
})
