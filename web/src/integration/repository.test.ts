import assert from 'assert'
import { createDriverForTest, Driver } from '../../../shared/src/testing/driver'
import { commonWebGraphQlResults } from './graphQlResults'
import { createWebIntegrationTestContext, WebIntegrationTestContext } from './context'
import {
    makeRepositoryRedirectResult,
    makeResolveRevisionResult,
    makeFileExternalLinksResult,
    makeTreeEntriesResult,
    makeBlobContentResult,
} from './graphQlResponseHelpers'

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
            const clickedFileName = 'async.go'
            const clickedCommit = ''

            testContext.overrideGraphQL({
                ...commonWebGraphQlResults,
                RepositoryRedirect: ({ repoName }) => makeRepositoryRedirectResult(repoName),
                ResolveRev: () => makeResolveRevisionResult(repositorySourcegraphUrl),
                FileExternalLinks: ({ filePath }) => makeFileExternalLinksResult(filePath),
                TreeEntries: () => makeTreeEntriesResult(repositorySourcegraphUrl, ['jsonrpc2.go', clickedFileName]),
                Blob: () => makeBlobContentResult('mock file blob'),
                TreeCommits: () => ({
                    node: {
                        __typename: 'Repository',
                        commit: {
                            ancestors: {
                                nodes: [
                                    {
                                        id:
                                            'R2l0Q29tbWl0OnsiciI6IlVtVndiM05wZEc5eWVUbzBNRGsxTXpnPSIsImMiOiIxNWMyMjkwZGNiMzc3MzFjYzRlZTVhMmExYzFlNWEyNWI0YzI4ZjgxIn0=',
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
                                        url:
                                            '/github.com/sourcegraph/jsonrpc2/-/commit/15c2290dcb37731cc4ee5a2a1c1e5a25b4c28f81',
                                        canonicalURL:
                                            '/github.com/sourcegraph/jsonrpc2/-/commit/15c2290dcb37731cc4ee5a2a1c1e5a25b4c28f81',
                                        externalURLs: [
                                            {
                                                url:
                                                    'https://github.com/sourcegraph/jsonrpc2/commit/15c2290dcb37731cc4ee5a2a1c1e5a25b4c28f81',
                                                serviceType: 'github',
                                            },
                                        ],
                                        tree: {
                                            canonicalURL:
                                                '/github.com/sourcegraph/jsonrpc2@15c2290dcb37731cc4ee5a2a1c1e5a25b4c28f81',
                                        },
                                    },
                                    {
                                        id:
                                            'R2l0Q29tbWl0OnsiciI6IlVtVndiM05wZEc5eWVUbzBNRGsxTXpnPSIsImMiOiI5ZTYxNWIxYzMyY2M1MTkxMzA1NzVlOGQxMGQwZDBmZWU4YTVlYjZjIn0=',
                                        oid: '9e615b1c32cc519130575e8d10d0d0fee8a5eb6c',
                                        abbreviatedOID: '9e615b1',
                                        message: 'LSIF Indexing Campaign',
                                        subject: 'LSIF Indexing Campaign',
                                        body: null,
                                        author: {
                                            person: {
                                                avatarURL: '',
                                                name: 'Sourcegraph Bot',
                                                email: 'campaigns@sourcegraph.com',
                                                displayName: 'Sourcegraph Bot',
                                                user: null,
                                            },
                                            date: '2020-04-29T16:57:20Z',
                                        },
                                        committer: {
                                            person: {
                                                avatarURL: '',
                                                name: 'Sourcegraph Bot',
                                                email: 'campaigns@sourcegraph.com',
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
                                                serviceType: 'github',
                                            },
                                        ],
                                        tree: {
                                            canonicalURL:
                                                '/github.com/sourcegraph/jsonrpc2@9e615b1c32cc519130575e8d10d0d0fee8a5eb6c',
                                        },
                                    },
                                    {
                                        id:
                                            'R2l0Q29tbWl0OnsiciI6IlVtVndiM05wZEc5eWVUbzBNRGsxTXpnPSIsImMiOiI5NmM0ZWZhYjdlZTI4ZjNkMWNmMWQyNDhhMDEzOWNlYTM3MzY4YjE4In0=',
                                        oid: '96c4efab7ee28f3d1cf1d248a0139cea37368b18',
                                        abbreviatedOID: '96c4efa',
                                        message:
                                            'Produce LSIF data for each commit for fast/precise code nav (#35)\n\n* Produce LSIF data for each commit for fast/precise code nav\r\n\r\n* Update lsif.yml\r',
                                        subject: 'Produce LSIF data for each commit for fast/precise code nav (#35)',
                                        body:
                                            '* Produce LSIF data for each commit for fast/precise code nav\r\n\r\n* Update lsif.yml',
                                        author: {
                                            person: {
                                                avatarURL: 'https://avatars0.githubusercontent.com/u/1976?v=4',
                                                name: 'Quinn Slack',
                                                email: 'qslack@qslack.com',
                                                displayName: 'Quinn Slack',
                                                user: { id: 'VXNlcjo2', username: 'sqs', url: '/users/sqs' },
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
                                                serviceType: 'github',
                                            },
                                        ],
                                        tree: {
                                            canonicalURL:
                                                '/github.com/sourcegraph/jsonrpc2@96c4efab7ee28f3d1cf1d248a0139cea37368b18',
                                        },
                                    },
                                ],
                                pageInfo: { hasNextPage: true },
                            },
                        },
                    },
                }),
                RepositoryCommit: () => ({
                    node: {
                        commit: {
                            __typename: 'GitCommit',
                            id:
                                'R2l0Q29tbWl0OnsiciI6IlVtVndiM05wZEc5eWVUbzBNRGsxTXpnPSIsImMiOiIxNWMyMjkwZGNiMzc3MzFjYzRlZTVhMmExYzFlNWEyNWI0YzI4ZjgxIn0=',
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
                            url: '/github.com/sourcegraph/jsonrpc2/-/commit/15c2290dcb37731cc4ee5a2a1c1e5a25b4c28f81',
                            canonicalURL:
                                '/github.com/sourcegraph/jsonrpc2/-/commit/15c2290dcb37731cc4ee5a2a1c1e5a25b4c28f81',
                            externalURLs: [
                                {
                                    url:
                                        'https://github.com/sourcegraph/jsonrpc2/commit/15c2290dcb37731cc4ee5a2a1c1e5a25b4c28f81',
                                    serviceType: 'github',
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
                                                            kind: 'DELETED',
                                                            html:
                                                                '<div><span style="color:#657b83;">  </span><span style="color:#268bd2;">build</span><span style="color:#657b83;">:\n</span></div>',
                                                        },
                                                        {
                                                            kind: 'ADDED',
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
                                diffStat: { added: 1, changed: 3, deleted: 4 },
                            },
                        },
                    },
                }),
            })

            await driver.page.goto(driver.sourcegraphBaseUrl + '/github.com/sourcegraph/jsonrpc2')

            await driver.page.waitForSelector('h2.tree-page__title')

            // Assert that the directory listing displays properly
            await driver.page.waitForSelector('.tree-page__entries--columns')

            const numberOfFileEntries = await driver.page.evaluate(
                () => document.querySelector<HTMLButtonElement>('.tree-page__entries--columns')?.children.length
            )

            assert.strictEqual(numberOfFileEntries, 2, 'Number of files in directory listing')

            await testContext.waitForGraphQLRequest(async () => {
                await driver.findElementWithText(clickedFileName, { selector: '.e2e-tree-entry-file', action: 'click' })
            }, 'Blob')

            await driver.page.waitForSelector('.e2e-repo-blob')
            await driver.assertWindowLocation(`/github.com/sourcegraph/jsonrpc2/-/blob/${clickedFileName}`)

            // Assert that the file is loaded
            await assertSelectorHasText('.breadcrumb .part-last', clickedFileName)

            // Return to repo page
            await driver.page.click('a.repo-header__repo')
            await driver.page.waitForSelector('h2.tree-page__title')
            await assertSelectorHasText('h2.tree-page__title', ' sourcegraph/jsonrpc2')
            await driver.assertWindowLocation('/github.com/sourcegraph/jsonrpc2')

            await driver.findElementWithText(clickedCommit, { selector: '.git-commit-node__oid', action: 'click' })
            await driver.page.waitForSelector('.git-commit-node__message-subject')
            await assertSelectorHasText('.git-commit-node__message-subject', 'update LSIF indexing CI workflow')
        })
    })
})
