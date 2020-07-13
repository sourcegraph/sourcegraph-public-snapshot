import assert from 'assert'
import { createDriverForTest, Driver } from '../../../shared/src/testing/driver'
import { commonWebGraphQlResults } from './graphQlResults'
import { createWebIntegrationTestContext, WebIntegrationTestContext } from './context'

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

    describe('index page', () => {
        it('loads when accessed with a repo url', async () => {
            testContext.overrideGraphQL({
                ...commonWebGraphQlResults,
                RepositoryRedirect: () => ({
                    repositoryRedirect: {
                        __typename: 'Repository',
                        id: 'UmVwb3NpdG9yeTo0MDk1Mzg=',
                        name: 'github.com/sourcegraph/jsonrpc2',
                        url: '/github.com/sourcegraph/jsonrpc2',
                        externalURLs: [{ url: 'https://github.com/sourcegraph/jsonrpc2', serviceType: 'github' }],
                        description:
                            'Package jsonrpc2 provides a client and server implementation of JSON-RPC 2.0 (http://www.jsonrpc.org/specification)',
                        viewerCanAdminister: true,
                        defaultBranch: { displayName: 'master' },
                    },
                }),
                ResolveRev: () => ({
                    repositoryRedirect: {
                        __typename: 'Repository',
                        mirrorInfo: { cloneInProgress: false, cloneProgress: '', cloned: true },
                        commit: {
                            oid: '15c2290dcb37731cc4ee5a2a1c1e5a25b4c28f81',
                            tree: { url: '/github.com/sourcegraph/jsonrpc2' },
                        },
                        defaultBranch: { abbrevName: 'master' },
                    },
                }),
                TreeEntries: () => ({
                    repository: {
                        commit: {
                            tree: {
                                isRoot: true,
                                url: '/github.com/sourcegraph/jsonrpc2',
                                entries: [
                                    {
                                        name: '.github',
                                        path: '.github',
                                        isDirectory: true,
                                        url: '/github.com/sourcegraph/jsonrpc2/-/tree/.github',
                                        submodule: null,
                                        isSingleChild: false,
                                    },
                                    {
                                        name: 'websocket',
                                        path: 'websocket',
                                        isDirectory: true,
                                        url: '/github.com/sourcegraph/jsonrpc2/-/tree/websocket',
                                        submodule: null,
                                        isSingleChild: false,
                                    },
                                    {
                                        name: '.travis.yml',
                                        path: '.travis.yml',
                                        isDirectory: false,
                                        url: '/github.com/sourcegraph/jsonrpc2/-/blob/.travis.yml',
                                        submodule: null,
                                        isSingleChild: false,
                                    },
                                    {
                                        name: 'LICENSE',
                                        path: 'LICENSE',
                                        isDirectory: false,
                                        url: '/github.com/sourcegraph/jsonrpc2/-/blob/LICENSE',
                                        submodule: null,
                                        isSingleChild: false,
                                    },
                                    {
                                        name: 'README.md',
                                        path: 'README.md',
                                        isDirectory: false,
                                        url: '/github.com/sourcegraph/jsonrpc2/-/blob/README.md',
                                        submodule: null,
                                        isSingleChild: false,
                                    },
                                    {
                                        name: 'async.go',
                                        path: 'async.go',
                                        isDirectory: false,
                                        url: '/github.com/sourcegraph/jsonrpc2/-/blob/async.go',
                                        submodule: null,
                                        isSingleChild: false,
                                    },
                                    {
                                        name: 'call_opt.go',
                                        path: 'call_opt.go',
                                        isDirectory: false,
                                        url: '/github.com/sourcegraph/jsonrpc2/-/blob/call_opt.go',
                                        submodule: null,
                                        isSingleChild: false,
                                    },
                                    {
                                        name: 'codec_test.go',
                                        path: 'codec_test.go',
                                        isDirectory: false,
                                        url: '/github.com/sourcegraph/jsonrpc2/-/blob/codec_test.go',
                                        submodule: null,
                                        isSingleChild: false,
                                    },
                                    {
                                        name: 'conn_opt.go',
                                        path: 'conn_opt.go',
                                        isDirectory: false,
                                        url: '/github.com/sourcegraph/jsonrpc2/-/blob/conn_opt.go',
                                        submodule: null,
                                        isSingleChild: false,
                                    },
                                    {
                                        name: 'go.mod',
                                        path: 'go.mod',
                                        isDirectory: false,
                                        url: '/github.com/sourcegraph/jsonrpc2/-/blob/go.mod',
                                        submodule: null,
                                        isSingleChild: false,
                                    },
                                    {
                                        name: 'go.sum',
                                        path: 'go.sum',
                                        isDirectory: false,
                                        url: '/github.com/sourcegraph/jsonrpc2/-/blob/go.sum',
                                        submodule: null,
                                        isSingleChild: false,
                                    },
                                    {
                                        name: 'handler_with_error.go',
                                        path: 'handler_with_error.go',
                                        isDirectory: false,
                                        url: '/github.com/sourcegraph/jsonrpc2/-/blob/handler_with_error.go',
                                        submodule: null,
                                        isSingleChild: false,
                                    },
                                    {
                                        name: 'jsonrpc2.go',
                                        path: 'jsonrpc2.go',
                                        isDirectory: false,
                                        url: '/github.com/sourcegraph/jsonrpc2/-/blob/jsonrpc2.go',
                                        submodule: null,
                                        isSingleChild: false,
                                    },
                                    {
                                        name: 'jsonrpc2_test.go',
                                        path: 'jsonrpc2_test.go',
                                        isDirectory: false,
                                        url: '/github.com/sourcegraph/jsonrpc2/-/blob/jsonrpc2_test.go',
                                        submodule: null,
                                        isSingleChild: false,
                                    },
                                    {
                                        name: 'object_test.go',
                                        path: 'object_test.go',
                                        isDirectory: false,
                                        url: '/github.com/sourcegraph/jsonrpc2/-/blob/object_test.go',
                                        submodule: null,
                                        isSingleChild: false,
                                    },
                                    {
                                        name: 'stream.go',
                                        path: 'stream.go',
                                        isDirectory: false,
                                        url: '/github.com/sourcegraph/jsonrpc2/-/blob/stream.go',
                                        submodule: null,
                                        isSingleChild: false,
                                    },
                                ],
                            },
                        },
                    },
                }),
                Blob: () => ({
                    repository: {
                        commit: {
                            file: {
                                content: '',
                                richHTML: '',
                                highlight: {
                                    aborted: false,
                                    html: '',
                                },
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

            console.log('len', numberOfFileEntries)
            assert.strictEqual(numberOfFileEntries, 16, 'Number of files in directory listing')

            await testContext.waitForGraphQLRequest(async () => {
                await driver.findElementWithText('async.go', { selector: '.e2e-tree-entry-file', action: 'click' })
            }, 'Blob')

            await driver.page.waitForSelector('.e2e-repo-blob')
            await driver.assertWindowLocation('/github.com/sourcegraph/jsonrpc2/-/blob/async.go')

            // Assert that the file is loaded
            assert.strictEqual(
                await driver.page.evaluate(
                    () => document.querySelector<HTMLButtonElement>('.breadcrumb .part-last')?.textContent
                ),
                'async.go'
            )
        })
    })
})
