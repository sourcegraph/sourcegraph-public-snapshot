import assert from 'assert'
import * as path from 'path'

import { subDays } from 'date-fns'
import type * as sourcegraph from 'sourcegraph'

import { encodeURIPathComponent } from '@sourcegraph/common'
import { ExtensionManifest } from '@sourcegraph/shared/src/extensions/extensionManifest'
import { SharedGraphQlOperations } from '@sourcegraph/shared/src/graphql-operations'
import { ExternalServiceKind } from '@sourcegraph/shared/src/schema'
import { Settings } from '@sourcegraph/shared/src/settings/settings'
import { accessibilityAudit } from '@sourcegraph/shared/src/testing/accessibility'
import { createDriverForTest, Driver } from '@sourcegraph/shared/src/testing/driver'
import { afterEachSaveScreenshotIfFailed } from '@sourcegraph/shared/src/testing/screenshotReporter'

import { DiffHunkLineType, RepositoryContributorsResult, WebGraphQlOperations } from '../graphql-operations'

import { createWebIntegrationTestContext, WebIntegrationTestContext } from './context'
import {
    createResolveRepoRevisionResult,
    createFileExternalLinksResult,
    createTreeEntriesResult,
    createBlobContentResult,
    createRepoChangesetsStatsResult,
    createFileNamesResult,
} from './graphQlResponseHelpers'
import { commonWebGraphQlResults } from './graphQlResults'
import { createEditorAPI, percySnapshotWithVariants } from './utils'

export const getCommonRepositoryGraphQlResults = (
    repositoryName: string,
    repositoryUrl: string,
    fileEntries: string[] = []
): Partial<WebGraphQlOperations & SharedGraphQlOperations> => ({
    ...commonWebGraphQlResults,
    RepoChangesetsStats: () => createRepoChangesetsStatsResult(),
    ResolveRepoRev: () => createResolveRepoRevisionResult(repositoryName),
    FileNames: () => createFileNamesResult(),
    FileExternalLinks: ({ filePath }) => createFileExternalLinksResult(filePath),
    TreeEntries: () => createTreeEntriesResult(repositoryUrl, fileEntries),
    TreeCommits: () => ({
        node: {
            __typename: 'Repository',
            commit: { ancestors: { nodes: [], pageInfo: { hasNextPage: false } } },
        },
    }),
    Blob: ({ filePath }) => createBlobContentResult(`content for: ${filePath}\nsecond line\nthird line`),
})

const now = new Date()
describe('Repository', () => {
    let driver: Driver
    before(async () => {
        driver = await createDriverForTest()
    })
    after(() => driver?.close())
    let testContext: WebIntegrationTestContext
    beforeEach(async function () {
        testContext = await createWebIntegrationTestContext({
            driver,
            currentTest: this.currentTest!,
            directory: __dirname,
        })
    })
    afterEachSaveScreenshotIfFailed(() => driver.page)
    afterEach(() => testContext?.dispose())

    async function assertSelectorHasText(selector: string, text: string) {
        assert.strictEqual(
            await driver.page.evaluate(
                selector => document.querySelector<HTMLButtonElement>(selector)?.textContent,
                selector
            ),
            text
        )
    }

    describe('index page', () => {
        it('loads when accessed with a repo url', async () => {
            const shortRepositoryName = 'sourcegraph/jsonrpc2'
            const repositoryName = `github.com/${shortRepositoryName}`
            const repositorySourcegraphUrl = `/${repositoryName}`
            const commitUrl = `${repositorySourcegraphUrl}/-/commit/15c2290dcb37731cc4ee5a2a1c1e5a25b4c28f81?visible=1`
            const clickedFileName = 'async.go'
            const clickedCommit = ''
            const fileEntries = ['jsonrpc2.go', clickedFileName]

            testContext.overrideGraphQL({
                ...commonWebGraphQlResults,
                ...getCommonRepositoryGraphQlResults(repositoryName, repositorySourcegraphUrl, fileEntries),
                TreeCommits: () => ({
                    node: {
                        __typename: 'Repository',
                        commit: {
                            ancestors: {
                                nodes: [
                                    {
                                        id: 'CommitID1',
                                        oid: '15c2290dcb37731cc4ee5a2a1c1e5a25b4c28f81',
                                        abbreviatedOID: '15c2290',
                                        message: 'update LSIF indexing CI workflow\n',
                                        subject: 'update LSIF indexing CI workflow',
                                        body: null,
                                        author: {
                                            person: {
                                                avatarURL: '',
                                                name: 'garo (they/them)',
                                                email: 'gbrik@users.noreply.github.com',
                                                displayName: 'garo (they/them)',
                                                user: null,
                                            },
                                            date: '2020-04-29T18:40:54Z',
                                        },
                                        committer: {
                                            person: {
                                                avatarURL: '',
                                                name: 'GitHub',
                                                email: 'noreply@github.com',
                                                displayName: 'GitHub',
                                                user: null,
                                            },
                                            date: '2020-04-29T18:40:54Z',
                                        },
                                        parents: [
                                            {
                                                oid: '96c4efab7ee28f3d1cf1d248a0139cea37368b18',
                                                abbreviatedOID: '96c4efa',
                                                url:
                                                    '/github.com/sourcegraph/jsonrpc2/-/commit/96c4efab7ee28f3d1cf1d248a0139cea37368b18',
                                            },
                                            {
                                                oid: '9e615b1c32cc519130575e8d10d0d0fee8a5eb6c',
                                                abbreviatedOID: '9e615b1',
                                                url:
                                                    '/github.com/sourcegraph/jsonrpc2/-/commit/9e615b1c32cc519130575e8d10d0d0fee8a5eb6c',
                                            },
                                        ],
                                        url: commitUrl,
                                        canonicalURL: commitUrl,
                                        externalURLs: [
                                            {
                                                url:
                                                    'https://github.com/sourcegraph/jsonrpc2/commit/15c2290dcb37731cc4ee5a2a1c1e5a25b4c28f81',
                                                serviceKind: ExternalServiceKind.GITHUB,
                                            },
                                        ],
                                        tree: {
                                            canonicalURL:
                                                '/github.com/sourcegraph/jsonrpc2@15c2290dcb37731cc4ee5a2a1c1e5a25b4c28f81',
                                        },
                                    },
                                    {
                                        id: 'CommitID2',
                                        oid: '9e615b1c32cc519130575e8d10d0d0fee8a5eb6c',
                                        abbreviatedOID: '9e615b1',
                                        message: 'LSIF Indexing Campaign',
                                        subject: 'LSIF Indexing Campaign',
                                        body: null,
                                        author: {
                                            person: {
                                                avatarURL: '',
                                                name: 'Sourcegraph Bot',
                                                email: 'batch-changes@sourcegraph.com',
                                                displayName: 'Sourcegraph Bot',
                                                user: null,
                                            },
                                            date: '2020-04-29T16:57:20Z',
                                        },
                                        committer: {
                                            person: {
                                                avatarURL: '',
                                                name: 'Sourcegraph Bot',
                                                email: 'batch-changes@sourcegraph.com',
                                                displayName: 'Sourcegraph Bot',
                                                user: null,
                                            },
                                            date: '2020-04-29T16:57:20Z',
                                        },
                                        parents: [
                                            {
                                                oid: '96c4efab7ee28f3d1cf1d248a0139cea37368b18',
                                                abbreviatedOID: '96c4efa',
                                                url:
                                                    '/github.com/sourcegraph/jsonrpc2/-/commit/96c4efab7ee28f3d1cf1d248a0139cea37368b18',
                                            },
                                        ],
                                        url:
                                            '/github.com/sourcegraph/jsonrpc2/-/commit/9e615b1c32cc519130575e8d10d0d0fee8a5eb6c',
                                        canonicalURL:
                                            '/github.com/sourcegraph/jsonrpc2/-/commit/9e615b1c32cc519130575e8d10d0d0fee8a5eb6c',
                                        externalURLs: [
                                            {
                                                url:
                                                    'https://github.com/sourcegraph/jsonrpc2/commit/9e615b1c32cc519130575e8d10d0d0fee8a5eb6c',
                                                serviceKind: ExternalServiceKind.GITHUB,
                                            },
                                        ],
                                        tree: {
                                            canonicalURL:
                                                '/github.com/sourcegraph/jsonrpc2@9e615b1c32cc519130575e8d10d0d0fee8a5eb6c',
                                        },
                                    },
                                    {
                                        id: 'CommitID3',
                                        oid: '96c4efab7ee28f3d1cf1d248a0139cea37368b18',
                                        abbreviatedOID: '96c4efa',
                                        message:
                                            'Produce LSIF data for each commit for fast/precise code nav (#35)\n\n* Produce LSIF data for each commit for fast/precise code nav\r\n\r\n* Update lsif.yml\r',
                                        subject: 'Produce LSIF data for each commit for fast/precise code nav (#35)',
                                        body:
                                            '* Produce LSIF data for each commit for fast/precise code nav\r\n\r\n* Update lsif.yml',
                                        author: {
                                            person: {
                                                avatarURL: '',
                                                name: 'Quinn Slack',
                                                email: 'qslack@qslack.com',
                                                displayName: 'Quinn Slack',
                                                user: {
                                                    id: 'VXNlcjo2',
                                                    username: 'sqs',
                                                    url: '/users/sqs',
                                                    displayName: 'sqs',
                                                },
                                            },
                                            date: '2019-12-22T04:34:38Z',
                                        },
                                        committer: {
                                            person: {
                                                avatarURL: '',
                                                name: 'GitHub',
                                                email: 'noreply@github.com',
                                                displayName: 'GitHub',
                                                user: null,
                                            },
                                            date: '2019-12-22T04:34:38Z',
                                        },
                                        parents: [
                                            {
                                                oid: 'cee7209801bf50cee868f8e0696ba0b76ae21792',
                                                abbreviatedOID: 'cee7209',
                                                url:
                                                    '/github.com/sourcegraph/jsonrpc2/-/commit/cee7209801bf50cee868f8e0696ba0b76ae21792',
                                            },
                                        ],
                                        url:
                                            '/github.com/sourcegraph/jsonrpc2/-/commit/96c4efab7ee28f3d1cf1d248a0139cea37368b18',
                                        canonicalURL:
                                            '/github.com/sourcegraph/jsonrpc2/-/commit/96c4efab7ee28f3d1cf1d248a0139cea37368b18',
                                        externalURLs: [
                                            {
                                                url:
                                                    'https://github.com/sourcegraph/jsonrpc2/commit/96c4efab7ee28f3d1cf1d248a0139cea37368b18',
                                                serviceKind: ExternalServiceKind.GITHUB,
                                            },
                                        ],
                                        tree: {
                                            canonicalURL:
                                                '/github.com/sourcegraph/jsonrpc2@96c4efab7ee28f3d1cf1d248a0139cea37368b18',
                                        },
                                    },
                                ],
                                pageInfo: { hasNextPage: false },
                            },
                        },
                    },
                }),
                RepositoryCommit: () => ({
                    node: {
                        __typename: 'Repository',
                        commit: {
                            __typename: 'GitCommit',
                            id: 'CommitID1',
                            oid: '15c2290dcb37731cc4ee5a2a1c1e5a25b4c28f81',
                            abbreviatedOID: '15c2290',
                            message: 'update LSIF indexing CI workflow\n',
                            subject: 'update LSIF indexing CI workflow',
                            body: null,
                            author: {
                                person: {
                                    avatarURL: '',
                                    name: 'garo (they/them)',
                                    email: 'gbrik@users.noreply.github.com',
                                    displayName: 'garo (they/them)',
                                    user: null,
                                },
                                date: '2020-04-29T18:40:54Z',
                            },
                            committer: {
                                person: {
                                    avatarURL: '',
                                    name: 'GitHub',
                                    email: 'noreply@github.com',
                                    displayName: 'GitHub',
                                    user: null,
                                },
                                date: '2020-04-29T18:40:54Z',
                            },
                            parents: [
                                {
                                    oid: '96c4efab7ee28f3d1cf1d248a0139cea37368b18',
                                    abbreviatedOID: '96c4efa',
                                    url:
                                        '/github.com/sourcegraph/jsonrpc2/-/commit/96c4efab7ee28f3d1cf1d248a0139cea37368b18',
                                },
                                {
                                    oid: '9e615b1c32cc519130575e8d10d0d0fee8a5eb6c',
                                    abbreviatedOID: '9e615b1',
                                    url:
                                        '/github.com/sourcegraph/jsonrpc2/-/commit/9e615b1c32cc519130575e8d10d0d0fee8a5eb6c',
                                },
                            ],
                            url: commitUrl,
                            canonicalURL: commitUrl,
                            externalURLs: [
                                {
                                    url:
                                        'https://github.com/sourcegraph/jsonrpc2/commit/15c2290dcb37731cc4ee5a2a1c1e5a25b4c28f81',
                                    serviceKind: ExternalServiceKind.GITHUB,
                                },
                            ],
                            tree: {
                                canonicalURL:
                                    '/github.com/sourcegraph/jsonrpc2@15c2290dcb37731cc4ee5a2a1c1e5a25b4c28f81',
                            },
                        },
                    },
                }),
                RepositoryComparisonDiff: () => ({
                    node: {
                        comparison: {
                            fileDiffs: {
                                nodes: [
                                    {
                                        __typename: 'FileDiff',
                                        oldPath: '.github/workflows/lsif.yml',
                                        oldFile: { __typename: 'GitBlob', binary: false, byteSize: 381 },
                                        newFile: { __typename: 'GitBlob', binary: false, byteSize: 304 },
                                        newPath: '.github/workflows/lsif.yml',
                                        mostRelevantFile: {
                                            __typename: 'GitBlob',
                                            url:
                                                '/github.com/sourcegraph/jsonrpc2@15c2290dcb37731cc4ee5a2a1c1e5a25b4c28f81/-/blob/.github/workflows/lsif.yml',
                                        },
                                        hunks: [
                                            {
                                                oldRange: { startLine: 2, lines: 15 },
                                                oldNoNewlineAt: false,
                                                newRange: { startLine: 2, lines: 12 },
                                                section: 'name: LSIF',
                                                highlight: {
                                                    aborted: false,
                                                    lines: [
                                                        {
                                                            kind: DiffHunkLineType.DELETED,
                                                            html:
                                                                '<div><span style="color:#657b83;">  </span><span style="color:#268bd2;">build</span><span style="color:#657b83;">:\n</span></div>',
                                                        },
                                                        {
                                                            kind: DiffHunkLineType.ADDED,
                                                            html:
                                                                '<div><span style="color:#657b83;">  </span><span style="color:#268bd2;">lsif-go</span><span style="color:#657b83;">:\n</span></div>',
                                                        },
                                                    ],
                                                },
                                            },
                                        ],
                                        stat: { added: 1, changed: 3, deleted: 4 },
                                        internalID: '084bcb27838a8adbbbe10f664420f2d2',
                                    },
                                ],
                                totalCount: 1,
                                pageInfo: { endCursor: null, hasNextPage: false },
                                diffStat: { added: 1, changed: 3, deleted: 4, __typename: 'DiffStat' },
                            },
                        },
                    },
                }),
            })

            // Mock `Date.now` to stabilize timestamps
            await driver.page.evaluateOnNewDocument(() => {
                // Number of ms between Unix epoch and July 1, 2020 (arbitrary)
                const mockMs = new Date('July 1, 2020 00:00:00 UTC').getTime()
                Date.now = () => mockMs
            })

            await driver.page.goto(driver.sourcegraphBaseUrl + '/' + repositoryName)

            await driver.page.waitForSelector('div.test-tree-page-title')

            // Assert that the directory listing displays properly
            await driver.page.waitForSelector('.test-tree-entries')

            // Wait for extensions bar to be loaded before screenshotting
            await driver.page.waitForSelector('[data-testid="action-items-toggle-open"]')

            await percySnapshotWithVariants(driver.page, 'Repository index page')
            await accessibilityAudit(driver.page)

            const numberOfFileEntries = await driver.page.evaluate(
                () => document.querySelectorAll<HTMLButtonElement>('.test-tree-entry-file')?.length
            )

            assert.strictEqual(numberOfFileEntries, fileEntries.length, 'Number of files in directory listing')

            await testContext.waitForGraphQLRequest(async () => {
                await driver.findElementWithText(clickedFileName, {
                    selector: '.test-tree-entry-file',
                    action: 'click',
                })
            }, 'Blob')

            await driver.page.waitForSelector('[data-testid="repo-blob"]')
            await driver.assertWindowLocation(`/${repositoryName}/-/blob/${clickedFileName}`)

            // Assert breadcrumb order
            await driver.page.waitForSelector('.test-breadcrumb')
            const breadcrumbTexts = await driver.page.evaluate(() =>
                [...document.querySelectorAll('.test-breadcrumb')].map(breadcrumb => breadcrumb.textContent?.trim())
            )
            assert.deepStrictEqual(breadcrumbTexts, [shortRepositoryName, '@master', `/${clickedFileName}`])

            // Return to repo page
            await driver.page.waitForSelector('.test-repo-header-repo-link')
            await driver.page.click('.test-repo-header-repo-link')

            await driver.page.waitForSelector('div.test-tree-page-title')
            await assertSelectorHasText('div.test-tree-page-title', shortRepositoryName)
            await driver.assertWindowLocation(repositorySourcegraphUrl)

            await driver.findElementWithText(clickedCommit, {
                selector: '[data-testid="git-commit-node-oid"]',
                action: 'click',
            })
            await driver.page.waitForSelector('[data-testid="repository-commit-page"]')
            await driver.page.waitForSelector('[data-testid="git-commit-node-message-subject"]')
            await driver.assertWindowLocation(commitUrl)

            await assertSelectorHasText(
                '[data-testid="git-commit-node-message-subject"]',
                'update LSIF indexing CI workflow'
            )
        })

        it('works with files with spaces in the name', async () => {
            const shortRepositoryName = 'ggilmore/q-test'
            const repositoryName = `github.com/${shortRepositoryName}`
            const repositorySourcegraphUrl = `/${repositoryName}`
            const fileName = '% token.4288249258.sql'
            const directoryName = "Geoffrey's random queries.32r242442bf"
            const filePath = path.posix.join(directoryName, fileName)

            testContext.overrideGraphQL({
                ...commonWebGraphQlResults,
                ...getCommonRepositoryGraphQlResults(repositoryName, repositorySourcegraphUrl),
                FileExternalLinks: ({ filePath, repoName, revision }) =>
                    createFileExternalLinksResult(
                        `https://${encodeURIPathComponent(repoName)}/blob/${encodeURIPathComponent(
                            revision
                        )}/${encodeURIPathComponent(filePath)}`
                    ),
                TreeEntries: () => ({
                    repository: {
                        commit: {
                            tree: {
                                isRoot: false,
                                url: '/github.com/ggilmore/q-test/-/tree/Geoffrey%27s%20random%20queries.32r242442bf',
                                entries: [
                                    {
                                        name: fileName,
                                        path: filePath,
                                        isDirectory: false,
                                        url:
                                            '/github.com/ggilmore/q-test/-/blob/Geoffrey%27s%20random%20queries.32r242442bf/%25%20token.4288249258.sql',
                                        submodule: null,
                                        isSingleChild: false,
                                    },
                                ],
                            },
                        },
                    },
                }),
            })

            await driver.page.goto(
                `${driver.sourcegraphBaseUrl}/github.com/ggilmore/q-test/-/tree/Geoffrey's%20random%20queries.32r242442bf`
            )
            await driver.page.waitForSelector('.test-tree-file-link')
            assert.strictEqual(
                await driver.page.evaluate(
                    () =>
                        document.querySelector(
                            '.test-tree-file-link[href="/github.com/ggilmore/q-test/-/blob/Geoffrey%27s%20random%20queries.32r242442bf/%25%20token.4288249258.sql"]'
                        )?.textContent
                ),
                fileName
            )

            // page.click() fails for some reason with Error: Node is either not visible or not an HTMLElement
            await driver.page.$eval(
                '.test-tree-file-link[href="/github.com/ggilmore/q-test/-/blob/Geoffrey%27s%20random%20queries.32r242442bf/%25%20token.4288249258.sql"]',
                linkElement => (linkElement as HTMLElement).click()
            )
            await driver.page.waitForSelector('[data-testid="repo-blob"]')

            await driver.page.waitForSelector('.test-breadcrumb')
            const breadcrumbTexts = await driver.page.evaluate(() =>
                [...document.querySelectorAll('.test-breadcrumb')].map(breadcrumb => breadcrumb.textContent?.trim())
            )
            assert.deepStrictEqual(breadcrumbTexts, [
                shortRepositoryName,
                '@master',
                "/Geoffrey's random queries.32r242442bf /% token.4288249258.sql",
            ])

            {
                const queryInput = await createEditorAPI(driver, '[data-testid="searchbox"] .test-query-input')
                assert.strictEqual(
                    await queryInput.getValue(),
                    "repo:^github\\.com/ggilmore/q-test$ file:^Geoffrey's\\ random\\ queries\\.32r242442bf/%\\ token\\.4288249258\\.sql"
                )
            }

            await driver.page.waitForSelector('.test-go-to-code-host')
            assert.strictEqual(
                await driver.page.evaluate(
                    () => document.querySelector<HTMLAnchorElement>('.test-go-to-code-host')?.href
                ),
                "https://github.com/ggilmore/q-test/blob/master/Geoffrey's%20random%20queries.32r242442bf/%25%20token.4288249258.sql"
            )

            const blobContent = await driver.page.evaluate(
                () => document.querySelector('[data-testid="repo-blob"]')?.textContent
            )
            assert.strictEqual(blobContent, `content for: ${filePath}\nsecond line\nthird line`)
        })

        it('works with a plus sign in the repository name', async () => {
            const shortRepositoryName = 'ubuntu/+source/quemu'
            const repositoryName = `github.com/${shortRepositoryName}`
            const repositorySourcegraphUrl = `/${shortRepositoryName}`

            testContext.overrideGraphQL({
                ...commonWebGraphQlResults,
                ...getCommonRepositoryGraphQlResults(repositoryName, repositorySourcegraphUrl, ['readme.md']),
            })

            await driver.page.goto(driver.sourcegraphBaseUrl + repositorySourcegraphUrl)

            await driver.page.waitForSelector('div.test-tree-page-title')
            await assertSelectorHasText('div.test-tree-page-title', shortRepositoryName)
            await assertSelectorHasText('.test-tree-entry-file', 'readme.md')

            {
                const queryInput = await createEditorAPI(driver, '[data-testid="searchbox"] .test-query-input')
                assert.strictEqual(await queryInput.getValue(), 'repo:^ubuntu/\\+source/quemu$ ') // + should be escaped in regular expression
            }

            // page.click() fails for some reason with Error: Node is either not visible or not an HTMLElement
            await driver.page.$eval('.test-tree-file-link', linkElement => (linkElement as HTMLElement).click())
            await driver.page.waitForSelector('[data-testid="repo-blob"]')

            await driver.page.waitForSelector('.test-breadcrumb')
            const breadcrumbTexts = await driver.page.evaluate(() =>
                [...document.querySelectorAll('.test-breadcrumb')].map(breadcrumb => breadcrumb.textContent?.trim())
            )
            assert.deepStrictEqual(breadcrumbTexts, [shortRepositoryName, '@master', '/readme.md'])
        })

        it('works with spaces in the repository name', async () => {
            const shortRepositoryName = 'my org/repo with spaces'
            const repositoryName = `github.com/${shortRepositoryName}`
            const repositorySourcegraphUrl = '/github.com/my%20org/repo%20with%20spaces'

            testContext.overrideGraphQL({
                ...commonWebGraphQlResults,
                ...getCommonRepositoryGraphQlResults(repositoryName, repositorySourcegraphUrl, ['readme.md']),
            })

            await driver.page.goto(driver.sourcegraphBaseUrl + repositorySourcegraphUrl)

            await driver.page.waitForSelector('div.test-tree-page-title')
            await assertSelectorHasText('div.test-tree-page-title', 'my org/repo with spaces')
            await assertSelectorHasText('.test-tree-entry-file', 'readme.md')

            // page.click() fails for some reason with Error: Node is either not visible or not an HTMLElement
            await driver.page.$eval('.test-tree-file-link', linkElement => (linkElement as HTMLElement).click())
            await driver.page.waitForSelector('[data-testid="repo-blob"]')

            await driver.page.waitForSelector('.test-breadcrumb')
            const breadcrumbTexts = await driver.page.evaluate(() =>
                [...document.querySelectorAll('.test-breadcrumb')].map(breadcrumb => breadcrumb.textContent?.trim())
            )
            assert.deepStrictEqual(breadcrumbTexts, [shortRepositoryName, '@master', '/readme.md'])
        })
    })

    describe('commits page', () => {
        it('loads commits of a repository', async () => {
            const shortRepositoryName = 'sourcegraph/sourcegraph'
            const repositoryName = `github.com/${shortRepositoryName}`
            testContext.overrideGraphQL({
                ...commonWebGraphQlResults,
                ResolveRepoRev: () => createResolveRepoRevisionResult(repositoryName),
                RepositoryGitCommits: () => ({
                    __typename: 'Query',
                    node: {
                        __typename: 'GitCommit',
                        commit: {
                            __typename: 'GitCommit',
                            ancestors: {
                                __typename: 'GitCommitConnection',
                                nodes: [
                                    {
                                        id:
                                            'R2l0Q29tbWl0OnsiciI6IlVtVndiM05wZEc5eWVUb3hORGs9IiwiYyI6IjI4NGFiYTAyNGIxYjU1ODU5MGU4ZTJmOTdkYmMzNTUzYTVlMGM3NmIifQ==',
                                        oid: '284aba024b1b558590e8e2f97dbc3553a5e0c76b',
                                        abbreviatedOID: '284aba0',
                                        message: 'sg: create a test command to run e2e tests locally (#34627)\n',
                                        subject: 'sg: create a test command to run e2e tests locally (#34627)',
                                        body: null,
                                        author: {
                                            person: {
                                                avatarURL: null,
                                                name: 'Jean-Hadrien Chabran',
                                                email: 'jr9@gmail.com',
                                                displayName: 'Jean-Hadrien Chabran',
                                                user: null,
                                            },
                                            date: subDays(now, 5).toISOString(),
                                        },
                                        committer: {
                                            person: {
                                                avatarURL: null,
                                                name: 'GitHub',
                                                email: 'noreply@yahoo.com',
                                                displayName: 'GitHub',
                                                user: null,
                                            },
                                            date: subDays(now, 5).toISOString(),
                                        },
                                        parents: [
                                            {
                                                oid: 'a2d1fd474d79dc29af6c7b4c33f02fe22287bd11',
                                                abbreviatedOID: 'a2d1fd4',
                                                url:
                                                    '/github.com/sourcegraph/sourcegraph/-/commit/a2d1fd474d79dc29af6c7b4c33f02fe22287bd11',
                                            },
                                        ],
                                        url:
                                            '/github.com/sourcegraph/sourcegraph/-/commit/284aba024b1b558590e8e2f97dbc3553a5e0c76b',
                                        canonicalURL:
                                            '/github.com/sourcegraph/sourcegraph/-/commit/284aba024b1b558590e8e2f97dbc3553a5e0c76b',
                                        externalURLs: [
                                            {
                                                url:
                                                    'https://github.com/sourcegraph/sourcegraph/commit/284aba024b1b558590e8e2f97dbc3553a5e0c76b',
                                                serviceKind: 'GITHUB',
                                            },
                                        ],
                                        tree: {
                                            canonicalURL:
                                                '/github.com/sourcegraph/sourcegraph@284aba024b1b558590e8e2f97dbc3553a5e0c76b',
                                        },
                                    },
                                    {
                                        id:
                                            'R2l0Q29tbWl0OnsiciI6IlVtVndiM05wZEc5eWVUb3hORGs9IiwiYyI6ImEyZDFmZDQ3NGQ3OWRjMjlhZjZjN2I0YzMzZjAyZmUyMjI4N2JkMTEifQ==',
                                        oid: 'a2d1fd474d79dc29af6c7b4c33f02fe22287bd11',
                                        abbreviatedOID: 'a2d1fd4',
                                        message:
                                            'Wildcard V2: <Checkbox /> migration (#34324)\n\nCo-authored-by: gitstart-sourcegraph <gitstart@users.noreply.github.com>',
                                        subject: 'Wildcard V2: <Checkbox /> migration (#34324)',
                                        body:
                                            'Co-authored-by: gitstart-sourcegraph <gitstart@users.noreply.github.com>',
                                        author: {
                                            person: {
                                                avatarURL: null,
                                                name: 'GitStart-SourceGraph',
                                                email: '89894075h@facebook.net',
                                                displayName: 'GitStart-SourceGraph',
                                                user: null,
                                            },
                                            date: subDays(now, 5).toISOString(),
                                        },
                                        committer: {
                                            person: {
                                                avatarURL: null,
                                                name: 'GitHub',
                                                email: 'google@yahoo.com',
                                                displayName: 'GitHub',
                                                user: null,
                                            },
                                            date: subDays(now, 5).toISOString(),
                                        },
                                        parents: [
                                            {
                                                oid: '3a163b92b5c45921fbc730ff2047ff7e60d8689b',
                                                abbreviatedOID: '3a163b9',
                                                url:
                                                    '/github.com/sourcegraph/sourcegraph/-/commit/3a163b92b5c45921fbc730ff2047ff7e60d8689b',
                                            },
                                        ],
                                        url:
                                            '/github.com/sourcegraph/sourcegraph/-/commit/a2d1fd474d79dc29af6c7b4c33f02fe22287bd11',
                                        canonicalURL:
                                            '/github.com/sourcegraph/sourcegraph/-/commit/a2d1fd474d79dc29af6c7b4c33f02fe22287bd11',
                                        externalURLs: [
                                            {
                                                url:
                                                    'https://github.com/sourcegraph/sourcegraph/commit/a2d1fd474d79dc29af6c7b4c33f02fe22287bd11',
                                                serviceKind: 'GITHUB',
                                            },
                                        ],
                                        tree: {
                                            canonicalURL:
                                                '/github.com/sourcegraph/sourcegraph@a2d1fd474d79dc29af6c7b4c33f02fe22287bd11',
                                        },
                                    },
                                    {
                                        id:
                                            'R2l0Q29tbWl0OnsiciI6IlVtVndiM05wZEc5eWVUb3hORGs9IiwiYyI6IjNhMTYzYjkyYjVjNDU5MjFmYmM3MzBmZjIwNDdmZjdlNjBkODY4OWIifQ==',
                                        oid: '3a163b92b5c45921fbc730ff2047ff7e60d8689b',
                                        abbreviatedOID: '3a163b9',
                                        message: 'web: ban `reactstrap` imports (#34881)\n',
                                        subject: 'web: ban `reactstrap` imports (#34881)',
                                        body: null,
                                        author: {
                                            person: {
                                                avatarURL: null,
                                                name: 'Valery Bugakov',
                                                email: 'user23@gmail.com',
                                                displayName: 'Valery Bugakov',
                                                user: null,
                                            },
                                            date: subDays(now, 5).toISOString(),
                                        },
                                        committer: {
                                            person: {
                                                avatarURL: null,
                                                name: 'GitHub',
                                                email: 'user43@gmail.com',
                                                displayName: 'GitHub',
                                                user: null,
                                            },
                                            date: subDays(now, 5).toISOString(),
                                        },
                                        parents: [
                                            {
                                                oid: 'e2e91f0dcdc90811c4f2f4df638bc459b2358e7d',
                                                abbreviatedOID: 'e2e91f0',
                                                url:
                                                    '/github.com/sourcegraph/sourcegraph/-/commit/e2e91f0dcdc90811c4f2f4df638bc459b2358e7d',
                                            },
                                        ],
                                        url:
                                            '/github.com/sourcegraph/sourcegraph/-/commit/3a163b92b5c45921fbc730ff2047ff7e60d8689b',
                                        canonicalURL:
                                            '/github.com/sourcegraph/sourcegraph/-/commit/3a163b92b5c45921fbc730ff2047ff7e60d8689b',
                                        externalURLs: [
                                            {
                                                url:
                                                    'https://github.com/sourcegraph/sourcegraph/commit/3a163b92b5c45921fbc730ff2047ff7e60d8689b',
                                                serviceKind: 'GITHUB',
                                            },
                                        ],
                                        tree: {
                                            canonicalURL:
                                                '/github.com/sourcegraph/sourcegraph@3a163b92b5c45921fbc730ff2047ff7e60d8689b',
                                        },
                                    },
                                ],
                                pageInfo: {
                                    __typename: 'PageInfo',
                                    hasNextPage: true,
                                },
                            },
                        },
                    },
                }),
                FileNames: () => ({
                    repository: {
                        id: 'repo-123',
                        __typename: 'Repository',
                        commit: {
                            id: 'c0ff33',
                            __typename: 'GitCommit',
                            fileNames: ['README.md'],
                        },
                    },
                }),
            })
            await driver.page.goto(driver.sourcegraphBaseUrl + '/github.com/sourcegraph/sourcegraph/-/commits')
            await driver.page.waitForSelector('[data-testid="commits-page"]', { visible: true })
            await percySnapshotWithVariants(driver.page, 'Repository commits page')
            await accessibilityAudit(driver.page)
        })
    })

    // Describes the ways the directory viewer and tree sidebar can be extended through Sourcegraph extensions.
    describe('extensibility', () => {
        const shortRepoName = 'sourcegraph/file-decs'
        const repoName = `github.com/${shortRepoName}`
        const repositorySourcegraphUrl = `/${shortRepoName}`

        beforeEach(() => {
            const userSettings: Settings = {
                extensions: {
                    'test/test': true,
                },
            }
            const extensionManifest: ExtensionManifest = {
                url: new URL('/-/static/extension/0001-test-test.js?hash--test-test', driver.sourcegraphBaseUrl).href,
                activationEvents: ['*'],
            }

            testContext.overrideGraphQL({
                ...commonWebGraphQlResults,
                ...getCommonRepositoryGraphQlResults(repoName, repositorySourcegraphUrl),
                FileExternalLinks: ({ filePath, repoName, revision }) =>
                    createFileExternalLinksResult(
                        `https://${encodeURIPathComponent(repoName)}/blob/${encodeURIPathComponent(
                            revision
                        )}/${encodeURIPathComponent(filePath)}`
                    ),
                ViewerSettings: () => ({
                    viewerSettings: {
                        __typename: 'SettingsCascade',
                        final: JSON.stringify(userSettings),
                        subjects: [
                            {
                                __typename: 'User',
                                displayName: 'Test User',
                                id: 'TestUserSettingsID',
                                latestSettings: {
                                    id: 123,
                                    contents: JSON.stringify(userSettings),
                                },
                                username: 'test',
                                viewerCanAdminister: true,
                                settingsURL: '/users/test/settings',
                            },
                        ],
                    },
                }),
                TreeEntries: ({ filePath, repoName }) => {
                    if (filePath === '') {
                        return {
                            repository: {
                                commit: {
                                    tree: {
                                        isRoot: true,
                                        url: `/${repoName}`,
                                        entries: [
                                            {
                                                isDirectory: true,
                                                isSingleChild: true,
                                                name: 'nested',
                                                path: 'nested',
                                                url: `/${repoName}/-/tree/nested`,
                                                submodule: null,
                                            },
                                            // recursiveSingleChild is always true in the web app
                                            {
                                                name: 'test.ts',
                                                path: 'nested/test.ts',
                                                isDirectory: false,
                                                url: `/${repoName}/-/blob/nested/test.ts`,
                                                submodule: null,
                                                isSingleChild: false,
                                            },
                                            {
                                                name: 'ReactComponent.tsx',
                                                path: 'nested/ReactComponent.tsx',
                                                isDirectory: false,
                                                url: `/${repoName}/-/blob/nested/ReactComponent.tsx`,
                                                submodule: null,
                                                isSingleChild: false,
                                            },
                                            {
                                                name: 'doubly-nested',
                                                path: 'nested/doubly-nested',
                                                isDirectory: true,
                                                url: `/${repoName}/-/tree/nested/doubly-nested`,
                                                submodule: null,
                                                isSingleChild: false,
                                            },
                                        ],
                                    },
                                },
                            },
                        }
                    }

                    if (filePath === 'nested') {
                        return {
                            repository: {
                                commit: {
                                    tree: {
                                        isRoot: false,
                                        url: `/${repoName}/-/tree/nested`,
                                        entries: [
                                            {
                                                name: 'test.ts',
                                                path: 'nested/test.ts',
                                                isDirectory: false,
                                                url: `/${repoName}/-/blob/nested/test.ts`,
                                                submodule: null,
                                                isSingleChild: false,
                                            },
                                            {
                                                name: 'ReactComponent.tsx',
                                                path: 'nested/ReactComponent.tsx',
                                                isDirectory: false,
                                                url: `/${repoName}/-/blob/nested/ReactComponent.tsx`,
                                                submodule: null,
                                                isSingleChild: false,
                                            },
                                            {
                                                name: 'doubly-nested',
                                                path: 'nested/doubly-nested',
                                                isDirectory: true,
                                                url: `/${repoName}/-/tree/nested/doubly-nested`,
                                                submodule: null,
                                                isSingleChild: false,
                                            },
                                        ],
                                    },
                                },
                            },
                        }
                    }

                    if (filePath === 'nested/doubly-nested') {
                        return {
                            repository: {
                                commit: {
                                    tree: {
                                        isRoot: false,
                                        url: `/${repoName}/-/tree/nested/doubly-nested`,
                                        entries: [
                                            {
                                                name: 'triply-nested.ts',
                                                path: 'nested/doubly-nested/triply-nested.ts',
                                                isDirectory: false,
                                                url: `/${repoName}/-/blob/nested/doubly-nested/triply-nested.ts`,
                                                submodule: null,
                                                isSingleChild: false,
                                            },
                                        ],
                                    },
                                },
                            },
                        }
                    }

                    // unknown
                    return {
                        repository: {
                            commit: {
                                tree: {
                                    isRoot: false,
                                    url: `/${repoName}/${filePath}`,
                                    entries: [],
                                },
                            },
                        },
                    }
                },
                Extensions: () => ({
                    extensionRegistry: {
                        __typename: 'ExtensionRegistry',
                        extensions: {
                            nodes: [
                                {
                                    id: 'test',
                                    extensionID: 'test/test',
                                    manifest: {
                                        jsonFields: extensionManifest,
                                    },
                                },
                            ],
                        },
                    },
                }),
            })

            // Serve a mock extension bundle with a simple file decoration provider
            testContext.server
                .get(new URL(extensionManifest.url, driver.sourcegraphBaseUrl).href)
                .intercept((request, response) => {
                    function extensionBundle(): void {
                        // eslint-disable-next-line @typescript-eslint/no-require-imports, @typescript-eslint/no-var-requires
                        const sourcegraph = require('sourcegraph') as typeof import('sourcegraph')

                        const vowels = 'aeiouAEIOU'

                        function activate(context: sourcegraph.ExtensionContext): void {
                            context.subscriptions.add(
                                sourcegraph.app.registerFileDecorationProvider({
                                    provideFileDecorations: ({ files }) =>
                                        files.map(file => {
                                            const fragments = file.path.split('/')
                                            const name = fragments[fragments.length - 1]
                                            return {
                                                uri: file.uri,
                                                after: {
                                                    contentText: `${
                                                        name.split('').filter(char => vowels.includes(char)).length
                                                    } vowels`,
                                                    color: file.isDirectory ? 'red' : 'blue',
                                                },
                                                meter: {
                                                    value: file.isDirectory ? 50 : 100,
                                                },
                                            }
                                        }),
                                })
                            )
                        }

                        exports.activate = activate
                    }
                    // Create an immediately-invoked function expression for the extensionBundle function
                    const extensionBundleString = `(${extensionBundle.toString()})()`
                    response.type('application/javascript; charset=utf-8').send(extensionBundleString)
                })
        })
        async function getDecorationsByFilename(
            pageOrSidebar: 'page' | 'sidebar',
            filename: string
        ): Promise<{ textContent?: string | null; percentage?: string | null } | null> {
            return driver.page.evaluate(
                ({ pageOrSidebar, filename }) => {
                    const decorable = [
                        ...document.querySelectorAll('.test-' + String(pageOrSidebar) + '-file-decorable'),
                    ].find(decorable =>
                        decorable?.querySelector('.test-file-decorable-name')?.textContent?.includes(filename)
                    )

                    if (!decorable) {
                        return null
                    }

                    return {
                        textContent: decorable.querySelector('.test-file-decoration-text')?.textContent,
                        percentage: decorable.querySelector('.test-file-decoration-meter')?.getAttribute('value'),
                    }
                },
                { pageOrSidebar, filename }
            )
        }

        it('file decorations work on tree page and sidebar', async () => {
            testContext.overrideJsContext({ enableLegacyExtensions: true })
            await driver.page.goto(`${driver.sourcegraphBaseUrl}/${repoName}`)

            try {
                await driver.page.waitForSelector('.test-file-decoration-container', { timeout: 10000 })
            } catch {
                throw new Error('Expected to see file decorations')
            }

            // TREE SIDEBAR ASSERTIONS

            const nestedDecorations = await getDecorationsByFilename('sidebar', 'nested')

            assert.deepStrictEqual(
                nestedDecorations,
                {
                    textContent: '2 vowels',
                    percentage: '50', // dirs are 50% in mock extension
                },
                'Incorrect decorations for nested on tree sidebar'
            )

            // Since nested is a single child, its children should be visible and decorated as well

            const testDecorations = await getDecorationsByFilename('sidebar', 'test.ts')

            assert.deepStrictEqual(
                testDecorations,
                {
                    textContent: '1 vowels',
                    percentage: '100', // files are 100% in mock extension
                },
                'Incorrect decorations for test.ts on tree sidebar'
            )

            const doublyNestedDecorations = await getDecorationsByFilename('sidebar', 'doubly-nested')

            assert.deepStrictEqual(
                doublyNestedDecorations,
                {
                    textContent: '4 vowels',
                    percentage: '50',
                },
                'Incorrect decorations for doubly-nested on tree sidebar'
            )

            // Expand directory. we want to trigger "noopRowClick" handler in order to not navigate to new tree page
            await driver.page.evaluate(() =>
                ([...document.querySelectorAll('.test-sidebar-file-decorable')]
                    .find(directory => directory.textContent?.includes('doubly-nested'))
                    ?.querySelector('.test-tree-noop-link') as HTMLAnchorElement | undefined)?.click()
            )

            // Wait for file decorations to be sent from extension host
            try {
                await driver.page.waitForFunction(
                    () =>
                        !![...document.querySelectorAll('.test-sidebar-file-decorable')]
                            .find(file =>
                                file.querySelector('.test-file-decorable-name')?.textContent?.includes('triply-nested')
                            )
                            ?.querySelector('.test-file-decoration-container'),
                    { timeout: 5000 }
                )
            } catch {
                throw new Error('Timed out waiting for "triply-nested" decorations in tree sidebar')
            }
            const triplyNestedDecorations = await getDecorationsByFilename('sidebar', 'triply-nested')

            assert.deepStrictEqual(
                triplyNestedDecorations,
                {
                    textContent: '3 vowels',
                    percentage: '100',
                },
                'Incorrect decorations for triply-nested.ts on tree sidebar'
            )

            // TREE PAGE ASSERTIONS

            try {
                await driver.findElementWithText('nested', {
                    selector: '.test-page-file-decorable .test-file-decorable-name',
                    fuzziness: 'contains',
                    wait: {
                        timeout: 3000,
                    },
                })
            } catch {
                throw new Error('timed out waiting for "nested" in tree page')
            }

            // Wait for decorations
            try {
                await driver.page.waitForSelector('.test-page-file-decorable .test-file-decoration-container')
            } catch {
                throw new Error('Timed out waiting for "nested" decorations in tree page')
            }

            await driver.page.evaluate(() =>
                ([...document.querySelectorAll('.test-page-file-decorable .test-file-decorable-name')].find(name =>
                    name?.textContent?.includes('nested')
                ) as HTMLAnchorElement | undefined)?.click()
            )

            // Wait for decorations
            try {
                await driver.page.waitForSelector('.test-page-file-decorable .test-file-decoration-container')
            } catch {
                throw new Error('Timed out waiting for "ReactComponent.tsx" decorations in tree page')
            }

            const reactDecorations = await getDecorationsByFilename('page', 'ReactComponent.tsx')
            assert.deepStrictEqual(
                reactDecorations,
                {
                    textContent: '5 vowels',
                    percentage: '100',
                },
                'Incorrect decorations for ReactComponent.tsx on tree page'
            )

            const doublyNestedPageDecorations = await getDecorationsByFilename('page', 'doubly-nested')
            // This should be equal to its sidebar decorations
            assert.deepStrictEqual(
                doublyNestedPageDecorations,
                {
                    textContent: '4 vowels',
                    percentage: '50',
                },
                'Incorrect decorations for doubly-nested on tree page'
            )

            await driver.page.evaluate(() =>
                ([...document.querySelectorAll('.test-page-file-decorable .test-file-decorable-name')].find(name =>
                    name?.textContent?.includes('doubly-nested')
                ) as HTMLAnchorElement | undefined)?.click()
            )

            // Wait for new tree page
            await driver.findElementWithText('triply-nested', {
                selector: '.test-page-file-decorable .test-file-decorable-name',
                fuzziness: 'contains',
                wait: {
                    timeout: 3000,
                },
            })

            // Wait for decorations
            try {
                await driver.page.waitForSelector('.test-page-file-decorable .test-file-decoration-container')
            } catch {
                throw new Error('Timed out waiting for "triply-nested" decorations in tree page')
            }

            const triplyNestedPageDecorations = await getDecorationsByFilename('page', 'triply-nested.ts')
            // This should be equal to its sidebar decorations
            assert.deepStrictEqual(
                triplyNestedPageDecorations,
                {
                    textContent: '3 vowels',
                    percentage: '100',
                },
                'Incorrect decorations for triply-nested.ts on tree page'
            )
        })
    })

    describe('Accessibility', () => {
        const shortRepositoryName = 'sourcegraph/sourcegraph'
        const repositoryName = `github.com/${shortRepositoryName}`

        describe('Contributors page', () => {
            const repositorySourcegraphUrl = `/${repositoryName}/-/stats/contributors`

            it('Should render correctly all contributors', async () => {
                testContext.overrideGraphQL({
                    ...commonWebGraphQlResults,
                    ...getCommonRepositoryGraphQlResults(repositoryName, repositorySourcegraphUrl, []),
                    RepositoryContributors: (): RepositoryContributorsResult => ({
                        node: {
                            contributors: {
                                nodes: [
                                    {
                                        person: {
                                            name: 'alice',
                                            displayName: 'alice',
                                            email: 'alice@sourcegraph.test',
                                            avatarURL: null,
                                            user: null,
                                            __typename: 'Person',
                                        },
                                        count: 1,
                                        commits: {
                                            nodes: [
                                                {
                                                    oid: '1'.repeat(40),
                                                    abbreviatedOID: '1'.repeat(7),
                                                    url: `/${repositoryName}/-/commit/${'1'.repeat(40)}`,
                                                    subject: 'Commit message 1',
                                                    author: { date: subDays(new Date(), 1).toISOString() },
                                                    __typename: 'GitCommit',
                                                },
                                            ],
                                            __typename: 'GitCommitConnection',
                                        },
                                        __typename: 'RepositoryContributor',
                                    },
                                    {
                                        person: {
                                            name: 'jack',
                                            displayName: 'jack',
                                            email: 'jack@sourcegraph.test',
                                            avatarURL: null,
                                            user: null,
                                            __typename: 'Person',
                                        },
                                        count: 1,
                                        commits: {
                                            nodes: [
                                                {
                                                    oid: '2'.repeat(40),
                                                    abbreviatedOID: '2'.repeat(7),
                                                    url: `/${repositoryName}/-/commit/${'2'.repeat(40)}`,
                                                    subject: 'Commit message 2',
                                                    author: { date: subDays(new Date(), 2).toISOString() },
                                                    __typename: 'GitCommit',
                                                },
                                            ],
                                            __typename: 'GitCommitConnection',
                                        },
                                        __typename: 'RepositoryContributor',
                                    },
                                    {
                                        person: {
                                            name: 'jill',
                                            displayName: 'jill',
                                            email: 'jill@sourcegraph.test',
                                            avatarURL: null,
                                            user: null,
                                            __typename: 'Person',
                                        },
                                        count: 1,
                                        commits: {
                                            nodes: [
                                                {
                                                    oid: '3'.repeat(40),
                                                    abbreviatedOID: '3'.repeat(7),
                                                    url: `/${repositoryName}/-/commit/${'3'.repeat(40)}`,
                                                    subject: 'Commit message 3',
                                                    author: { date: subDays(new Date(), 3).toISOString() },
                                                    __typename: 'GitCommit',
                                                },
                                            ],
                                            __typename: 'GitCommitConnection',
                                        },
                                        __typename: 'RepositoryContributor',
                                    },
                                ],
                                totalCount: 3,
                                pageInfo: { hasNextPage: false },
                                __typename: 'RepositoryContributorConnection',
                            },
                            __typename: 'Repository',
                        },
                    }),
                })
                await driver.page.goto(driver.sourcegraphBaseUrl + repositorySourcegraphUrl)
                await driver.page.waitForSelector('.test-filtered-contributors-connection')
                await percySnapshotWithVariants(driver.page, 'Contributor list')
                await accessibilityAudit(driver.page)
            })
        })

        describe('Branches page', () => {
            it('should render correctly branches', async () => {
                const repositorySourcegraphUrl = `/${repositoryName}/-/branches`
                testContext.overrideGraphQL({
                    ...commonWebGraphQlResults,
                    ...getCommonRepositoryGraphQlResults(repositoryName, repositorySourcegraphUrl, []),
                    RepositoryGitBranchesOverview: () => ({
                        node: {
                            defaultBranch: {
                                id: 'QmV3b2Q=',
                                displayName: 'main',
                                name: 'refs/heads/main',
                                abbrevName: 'main',
                                url: `/${repositoryName}/-/branches/${'1'.repeat(40)}`,
                                target: {
                                    commit: {
                                        author: {
                                            __typename: 'Signature',
                                            person: {
                                                displayName: 'John Doe',
                                                user: {
                                                    username: 'johndoe',
                                                },
                                            },
                                            date: subDays(new Date(), 3).toISOString(),
                                        },
                                        committer: {
                                            __typename: 'Signature',
                                            person: {
                                                displayName: 'John Doe',
                                                user: null,
                                            },
                                            date: subDays(new Date(), 1).toISOString(),
                                        },
                                        behindAhead: {
                                            behind: 0,
                                            ahead: 0,
                                        },
                                    },
                                },
                            },
                            gitRefs: {
                                pageInfo: { hasNextPage: false },
                                nodes: [
                                    {
                                        id: 'BranchId1',
                                        displayName: 'integration-tests-trigramming',
                                        name: 'refs/heads/integration-tests-trigramming',
                                        abbrevName: 'integration-tests-trigramming',
                                        url: `/${repositoryName}/-/branches/${'1'.repeat(40)}`,
                                        target: {
                                            commit: {
                                                author: {
                                                    __typename: 'Signature',
                                                    person: {
                                                        displayName: 'John Doe',
                                                        user: {
                                                            username: 'johndoe',
                                                        },
                                                    },
                                                    date: subDays(new Date(), 1).toISOString(),
                                                },
                                                committer: {
                                                    __typename: 'Signature',
                                                    person: {
                                                        displayName: 'John Doe',
                                                        user: {
                                                            username: 'johndoe',
                                                        },
                                                    },
                                                    date: subDays(new Date(), 1).toISOString(),
                                                },
                                                behindAhead: {
                                                    behind: 12633,
                                                    ahead: 1,
                                                },
                                            },
                                        },
                                    },
                                    {
                                        id: 'BranchId2',
                                        displayName: 'integration-tests-quadgramming',
                                        name: 'refs/heads/integration-tests-quadgramming',
                                        abbrevName: 'integration-tests-quadgramming',
                                        url: `/${repositoryName}/-/branches/${'1'.repeat(40)}`,
                                        target: {
                                            commit: {
                                                author: {
                                                    __typename: 'Signature',
                                                    person: {
                                                        displayName: 'Alice',
                                                        user: {
                                                            username: 'alice',
                                                        },
                                                    },
                                                    date: subDays(new Date(), 1).toISOString(),
                                                },
                                                committer: {
                                                    __typename: 'Signature',
                                                    person: {
                                                        displayName: 'Alice',
                                                        user: {
                                                            username: 'alice',
                                                        },
                                                    },
                                                    date: subDays(new Date(), 1).toISOString(),
                                                },
                                                behindAhead: {
                                                    behind: 12633,
                                                    ahead: 1,
                                                },
                                            },
                                        },
                                    },
                                ],
                            },
                        },
                    }),
                })

                await driver.page.goto(driver.sourcegraphBaseUrl + repositorySourcegraphUrl)
                await driver.page.waitForSelector('[data-testid="active-branches-list"]')
                await percySnapshotWithVariants(driver.page, 'Repository branches page')
                await accessibilityAudit(driver.page)
            })
        })

        describe('Tags page', () => {
            const repositorySourcegraphUrl = `/${repositoryName}/-/tags`
            it('should render correctly tags list page', async () => {
                testContext.overrideGraphQL({
                    ...commonWebGraphQlResults,
                    ...getCommonRepositoryGraphQlResults(repositoryName, repositorySourcegraphUrl, []),
                    RepositoryGitRefs: () => ({
                        node: {
                            __typename: 'Repository',
                            gitRefs: {
                                __typename: 'GitRefConnection',
                                nodes: [
                                    {
                                        __typename: 'GitRef',
                                        name: 'refs/heads/main',
                                        abbrevName: 'v3.39.1',
                                        displayName: 'v3.39.1',
                                        id: 'GitRef:refs/heads/main',
                                        url: '/github.com/sourcegraph/sourcegraph@2'.repeat(70),
                                        target: {
                                            commit: {
                                                author: {
                                                    __typename: 'Signature',
                                                    person: {
                                                        displayName: 'John Doe',
                                                        user: {
                                                            username: 'johndoe',
                                                        },
                                                    },
                                                    date: subDays(new Date(), 1).toISOString(),
                                                },
                                                committer: {
                                                    __typename: 'Signature',
                                                    person: {
                                                        displayName: 'John Doe',
                                                        user: {
                                                            username: 'johndoe',
                                                        },
                                                    },
                                                    date: subDays(new Date(), 1).toISOString(),
                                                },
                                                behindAhead: {
                                                    __typename: 'BehindAheadCounts',
                                                    ahead: 0,
                                                    behind: 0,
                                                },
                                            },
                                        },
                                    },
                                    {
                                        __typename: 'GitRef',
                                        name: 'refs/heads/mai2n',
                                        abbrevName: 'v3.39.2',
                                        displayName: 'v3.39.2',
                                        id: 'GitRef:refs/heads/main2',
                                        url: '/github.com/sourcegraph/sourcegraph@2'.repeat(70),
                                        target: {
                                            commit: {
                                                author: {
                                                    __typename: 'Signature',
                                                    person: {
                                                        displayName: 'Alice',
                                                        user: {
                                                            username: 'alice',
                                                        },
                                                    },
                                                    date: subDays(new Date(), 1).toISOString(),
                                                },
                                                committer: {
                                                    __typename: 'Signature',
                                                    person: {
                                                        displayName: 'Alice',
                                                        user: {
                                                            username: 'alice',
                                                        },
                                                    },
                                                    date: subDays(new Date(), 1).toISOString(),
                                                },
                                                behindAhead: {
                                                    __typename: 'BehindAheadCounts',
                                                    ahead: 0,
                                                    behind: 0,
                                                },
                                            },
                                        },
                                    },
                                ],
                                totalCount: 2,
                                pageInfo: {
                                    hasNextPage: false,
                                },
                            },
                        },
                    }),
                })
                await driver.page.goto(driver.sourcegraphBaseUrl + repositorySourcegraphUrl)
                await driver.page.waitForSelector('.test-filtered-tags-connection')
                await driver.page.click('input[name="query"]')
                await driver.page.waitForSelector('input[name="query"].focus-visible')
                await percySnapshotWithVariants(driver.page, 'Repository tags page')
                await accessibilityAudit(driver.page)
            })
        })

        describe('Compare page', () => {
            const repositorySourcegraphUrl = `/${repositoryName}/-/compare/main...bl/readme?visible=1`
            it('should render correctly compare page, including diff view', async () => {
                testContext.overrideGraphQL({
                    ...commonWebGraphQlResults,
                    ...getCommonRepositoryGraphQlResults(repositoryName, repositorySourcegraphUrl, []),
                    RepositoryComparison: () => ({
                        node: {
                            comparison: {
                                range: {
                                    expr: 'main...bl/readme',
                                    baseRevSpec: { object: { oid: '1'.repeat(40) } },
                                    headRevSpec: { object: { oid: '2'.repeat(40) } },
                                },
                            },
                        },
                    }),
                    RepositoryComparisonCommits: () => ({
                        node: {
                            comparison: {
                                commits: {
                                    nodes: [
                                        {
                                            id: '1'.repeat(70),
                                            oid: '1'.repeat(40),
                                            abbreviatedOID: '1'.repeat(7),
                                            message: 'update README',
                                            subject: 'update README',
                                            body: null,
                                            author: {
                                                person: {
                                                    avatarURL: null,
                                                    name: 'alice',
                                                    email: 'alice@sourcegraph.com',
                                                    displayName: 'alice',
                                                    user: {
                                                        id: '1'.repeat(70),
                                                        username: 'alice',
                                                        url: '/users/alice',
                                                        displayName: 'alice',
                                                    },
                                                },
                                                date: subDays(new Date(), 1).toISOString(),
                                            },
                                            committer: {
                                                person: {
                                                    avatarURL: null,
                                                    name: 'alice',
                                                    email: 'alice@sourcegraph.com',
                                                    displayName: 'alice',
                                                    user: {
                                                        id: '1'.repeat(70),
                                                        username: 'alice',
                                                        url: '/users/alice',
                                                        displayName: 'alice',
                                                    },
                                                },
                                                date: subDays(new Date(), 1).toISOString(),
                                            },
                                            parents: [
                                                {
                                                    oid: '2'.repeat(40),
                                                    abbreviatedOID: '2'.repeat(7),
                                                    url: '/github.com/sourcegraph/sourcegraph@2'.repeat(70),
                                                },
                                            ],
                                            url: '/github.com/sourcegraph/sourcegraph@2'.repeat(70),
                                            canonicalURL: '/github.com/sourcegraph/sourcegraph@2'.repeat(70),
                                            externalURLs: [
                                                {
                                                    url: '/github.com/sourcegraph/sourcegraph@2'.repeat(70),
                                                    serviceKind: ExternalServiceKind.GITHUB,
                                                },
                                            ],
                                            tree: {
                                                canonicalURL: '/github.com/sourcegraph/sourcegraph@2'.repeat(70),
                                            },
                                        },
                                    ],
                                    pageInfo: { hasNextPage: false },
                                },
                            },
                        },
                    }),
                    RepositoryComparisonDiff: () => ({
                        node: {
                            comparison: {
                                fileDiffs: {
                                    nodes: [
                                        {
                                            oldPath: 'README.md',
                                            oldFile: { __typename: 'GitBlob', binary: false, byteSize: 2262 },
                                            newFile: { __typename: 'GitBlob', binary: false, byteSize: 3083 },
                                            newPath: 'README.md',
                                            mostRelevantFile: {
                                                __typename: 'GitBlob',
                                                url: `/${repositoryName}/-/commit/${'1'.repeat(40)}`,
                                            },
                                            hunks: [
                                                {
                                                    oldRange: { startLine: 3, lines: 47 },
                                                    oldNoNewlineAt: false,
                                                    newRange: { startLine: 3, lines: 64 },
                                                    section: null,
                                                    highlight: {
                                                        aborted: false,
                                                        lines: [
                                                            {
                                                                kind: DiffHunkLineType.UNCHANGED,
                                                                html:
                                                                    '\u003Cdiv\u003E\u003Cspan class="hl-text hl-html hl-markdown"\u003E[![build](https://badge.buildkite.com/00bbe6fa9986c78b8e8591cffeb0b0f2e8c4bb610d7e339ff6.svg?branch=master)](https://buildkite.com/sourcegraph/sourcegraph)\n\u003C/span\u003E\u003C/div\u003E',
                                                            },
                                                            {
                                                                kind: DiffHunkLineType.DELETED,
                                                                html:
                                                                    '\u003Cdiv\u003E\u003Cspan class="hl-text hl-html hl-markdown"\u003E[![build](https://badge.buildkite.com/00bbe6fa9986c78b8e8591cffeb0b0f2e8c4bb610d7e339ff6.svg?branch=master)](https://buildkite.com/sourcegraph/sourcegraph)\n\u003C/span\u003E\u003C/div\u003E',
                                                            },
                                                            {
                                                                kind: DiffHunkLineType.ADDED,
                                                                html:
                                                                    '\u003Cdiv\u003E\u003Cspan class="hl-text hl-html hl-markdown"\u003E[![build](https://badge.buildkite.com/00bbe6fa9986c78b8e8591cffeb0b0f2e8c4bb610d7e339ff6.svg?branch=master)](https://buildkite.com/sourcegraph/sourcegraph)\n\u003C/span\u003E\u003C/div\u003E',
                                                            },
                                                        ],
                                                    },
                                                },
                                            ],
                                            stat: { added: 31, changed: 13, deleted: 14 },
                                            internalID: '1'.repeat(70),
                                        },
                                    ],
                                    totalCount: 1,
                                    pageInfo: { endCursor: null, hasNextPage: false },
                                    diffStat: { __typename: 'DiffStat', added: 1, changed: 1, deleted: 1 },
                                },
                            },
                        },
                    }),
                })
                await driver.page.goto(driver.sourcegraphBaseUrl + repositorySourcegraphUrl)
                await driver.page.waitForSelector('.test-file-diff-connection')
                await percySnapshotWithVariants(driver.page, 'Repository compare page')
                await accessibilityAudit(driver.page)
            })
        })
    })
})
