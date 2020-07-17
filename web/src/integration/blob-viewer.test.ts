import assert from 'assert'
import { commonWebGraphQlResults } from './graphQlResults'
import { Driver, createDriverForTest } from '../../../shared/src/testing/driver'
import { WebIntegrationTestContext, createWebIntegrationTestContext } from './context'
import {
    createRepositoryRedirectResult,
    createResolveRevisionResult,
    createFileExternalLinksResult,
    createTreeEntriesResult,
    createBlobContentResult,
} from './graphQlResponseHelpers'
import { saveScreenshotsUponFailures } from '../../../shared/src/testing/screenshotReporter'

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
    saveScreenshotsUponFailures(() => driver.page)
    afterEach(() => testContext?.dispose())

    describe('general layout for viewing a file', () => {
        const repositoryName = 'github.com/sourcegraph/jsonrpc2'
        const repositorySourcegraphUrl = `/${repositoryName}`
        const fileName = 'async.go'
        const files = ['README.md', fileName]

        const prepareTwoFilesStubs = () =>
            testContext.overrideGraphQL({
                ...commonWebGraphQlResults,
                RepositoryRedirect: ({ repoName }) => createRepositoryRedirectResult(repoName),

                ResolveRev: () => createResolveRevisionResult(repositorySourcegraphUrl),

                FileExternalLinks: ({ filePath }) =>
                    createFileExternalLinksResult(`https://${repositoryName}/blob/master/${filePath}`),

                TreeEntries: () => createTreeEntriesResult(repositorySourcegraphUrl, files),

                Blob: ({ filePath }) => createBlobContentResult(`content for: ${filePath}`),
            })

        it('populates blob viewer with file content', async () => {
            prepareTwoFilesStubs()

            await driver.page.goto(`${driver.sourcegraphBaseUrl}/${repositoryName}/-/blob/${fileName}`)
            await driver.page.waitForSelector('.test-repo-blob')
            const blobContent = await driver.page.evaluate(() => {
                const editorArea = document.querySelector<HTMLDivElement>('.test-repo-blob')
                return editorArea ? editorArea.textContent : null
            })

            // editor shows the return string content from Blob request
            assert.strictEqual(blobContent, `content for: ${fileName}`)
        })

        it('populates files tree view', async () => {
            prepareTwoFilesStubs()

            await driver.page.goto(`${driver.sourcegraphBaseUrl}/${repositoryName}/-/blob/${fileName}`)
            await driver.page.waitForSelector('.test-tree-file-link')

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
})
