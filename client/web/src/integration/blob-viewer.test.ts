import assert from 'assert'

import { SharedGraphQlOperations } from '@sourcegraph/shared/src/graphql-operations'
import { Driver, createDriverForTest } from '@sourcegraph/shared/src/testing/driver'
import { afterEachSaveScreenshotIfFailed } from '@sourcegraph/shared/src/testing/screenshotReporter'

import { WebGraphQlOperations } from '../graphql-operations'

import { WebIntegrationTestContext, createWebIntegrationTestContext } from './context'
import {
    createResolveRepoRevisionResult,
    createFileExternalLinksResult,
    createTreeEntriesResult,
    createBlobContentResult,
    createFileTreeEntriesResult,
} from './graphQlResponseHelpers'
import { commonWebGraphQlResults, createViewerSettingsGraphQLOverride } from './graphQlResults'

describe('Blob viewer', () => {
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

    const repositoryName = 'github.com/sourcegraph/jsonrpc2'
    const repositorySourcegraphUrl = `/${repositoryName}`
    const fileName = 'test.ts'
    const files = ['README.md', fileName]

    const commonBlobGraphQlResults: Partial<WebGraphQlOperations & SharedGraphQlOperations> = {
        ...commonWebGraphQlResults,
        ...createViewerSettingsGraphQLOverride({
            user: {
                experimentalFeatures: {
                    enableCodeMirrorFileView: false,
                },
            },
        }),
        ResolveRepoRev: () => createResolveRepoRevisionResult(repositorySourcegraphUrl),
        FileExternalLinks: ({ filePath }) =>
            createFileExternalLinksResult(`https://${repositoryName}/blob/master/${filePath}`),
        TreeEntries: () => createTreeEntriesResult(repositorySourcegraphUrl, ['README.md', fileName]),
        FileTreeEntries: () => createFileTreeEntriesResult(repositorySourcegraphUrl, ['README.md', fileName]),
        Blob: ({ filePath }) => createBlobContentResult(`content for: ${filePath}\nsecond line\nthird line`),
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
    }

    beforeEach(() => {
        testContext.overrideGraphQL(commonBlobGraphQlResults)
    })

    describe('general layout for viewing a file', () => {
        it('populates editor content and FILES tab', async () => {
            await driver.page.goto(`${driver.sourcegraphBaseUrl}/${repositoryName}/-/blob/${fileName}`)
            await driver.page.waitForSelector('[data-testid="repo-blob"]')
            const blobContent = await driver.page.evaluate(
                () => document.querySelector<HTMLElement>('[data-testid="repo-blob"]')?.textContent
            )

            // editor shows the return string content from Blob request
            assert.strictEqual(blobContent, `content for: ${fileName}\nsecond line\nthird line`)

            // collect all files/links visible the the "Files" tab
            const allFilesInTheTree = await driver.page.evaluate(() => {
                const allFiles = document.querySelectorAll<HTMLAnchorElement>('.test-tree-file-link')

                return [...allFiles].map(fileAnchor => ({
                    content: fileAnchor.textContent,
                    href: fileAnchor.href,
                }))
            })

            // files from TreeEntries request
            assert.deepStrictEqual(
                allFilesInTheTree,
                files.map(name => ({
                    content: name,
                    href: `${driver.sourcegraphBaseUrl}/${repositoryName}/-/blob/${name}`,
                }))
            )
        })
    })

    describe('line number redirects', () => {
        beforeEach(() => {
            testContext.overrideGraphQL({
                ...commonBlobGraphQlResults,
                Blob: () => ({
                    repository: {
                        commit: {
                            blob: null,
                            file: {
                                __typename: 'VirtualFile',
                                content: '// Log to console\nconsole.log("Hello world")\n// Third line',
                                totalLines: 3,
                                richHTML: '',
                                highlight: {
                                    aborted: false,
                                    html:
                                        // Note: whitespace in this string is significant.
                                        '<table class="test-log-token"><tbody><tr><td class="line" data-line="1"/>' +
                                        '<td class="code"><span class="hl-source hl-js hl-react"><span class="hl-comment hl-line hl-double-slash hl-js">' +
                                        '<span class="hl-punctuation hl-definition hl-comment hl-js">//</span> ' +
                                        'Log to console\n</span></span></td></tr>' +
                                        '<tr><td class="line" data-line="2"/><td class="code"><span class="hl-source hl-js hl-react">' +
                                        '<span class="hl-meta hl-function-call hl-method hl-js">' +
                                        '<span class="hl-support hl-type hl-object hl-console hl-js">console</span>' +
                                        '<span class="hl-punctuation hl-accessor hl-js">.</span>' +
                                        '<span class="hl-support hl-function hl-console hl-js test-log-token">log</span>' +
                                        '<span class="hl-meta hl-group hl-js"><span class="hl-punctuation hl-section hl-group hl-begin hl-js">(</span>' +
                                        '<span class="hl-meta hl-string hl-js"><span class="hl-string hl-quoted hl-double hl-js">' +
                                        '<span class="hl-punctuation hl-definition hl-string hl-begin hl-js">&quot;</span>Hello world' +
                                        '<span class="hl-punctuation hl-definition hl-string hl-end hl-js">&quot;</span></span></span>' +
                                        '<span class="hl-punctuation hl-section hl-group hl-end hl-js">)</span></span>\n</span></span></td></tr>' +
                                        '<tr><td class="line" data-line="3"/><td class="code"><span class="hl-source hl-js hl-react">' +
                                        '<span class="hl-meta hl-function-call hl-method hl-js"></span>' +
                                        '<span class="hl-comment hl-line hl-double-slash hl-js"><span class="hl-punctuation hl-definition hl-comment hl-js">//</span> ' +
                                        'Third line\n</span></span></td></tr></tbody></table>',
                                    lsif: '',
                                },
                            },
                        },
                    },
                }),
            })
        })

        it('should redirect from line number hash to query parameter', async () => {
            await driver.page.goto(`${driver.sourcegraphBaseUrl}/${repositoryName}/-/blob/${fileName}#2`)
            await driver.page.waitForSelector('[data-testid="repo-blob"]')
            await driver.assertWindowLocation(`/${repositoryName}/-/blob/${fileName}?L2`)
        })

        it('should redirect from line range hash to query parameter', async () => {
            await driver.page.goto(`${driver.sourcegraphBaseUrl}/${repositoryName}/-/blob/${fileName}#1-3`)
            await driver.page.waitForSelector('[data-testid="repo-blob"]')
            await driver.assertWindowLocation(`/${repositoryName}/-/blob/${fileName}?L1-3`)
        })
    })
})
