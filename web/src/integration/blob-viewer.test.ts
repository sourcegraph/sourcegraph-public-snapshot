import assert from 'assert'
import { commonWebGraphQlResults } from './graphQlResults'
import { Driver, createDriverForTest } from '../../../shared/src/testing/driver'
import { WebIntegrationTestContext, createWebIntegrationTestContext } from './context'
import { test } from 'mocha'
import {
    makeRepositoryRedirectResult,
    makeResolveRevisionResult,
    makeFileExternalLinksResult,
    makeTreeEntriesResult,
    makeBlobContentResult,
} from './graphQlResponseHelpers'

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
    afterEach(() => testContext?.dispose())

    describe('tests a general layout for viewing a file', () => {
        test.only('it populates editor content and FILES tab', async () => {
            const repositoryName = 'github.com/sourcegraph/jsonrpc2'
            const repositorySourcegraphUrl = `/${repositoryName}`
            const fileName = 'async.go'

            testContext.overrideGraphQL({
                ...commonWebGraphQlResults,
                RepositoryRedirect: ({ repoName }) => makeRepositoryRedirectResult(repoName),

                ResolveRev: () => makeResolveRevisionResult(repositorySourcegraphUrl),

                FileExternalLinks: ({ filePath }) =>
                    makeFileExternalLinksResult(`https://${repositoryName}/blob/master/${filePath}`),

                TreeEntries: () => makeTreeEntriesResult(repositorySourcegraphUrl, ['README.md', fileName]),

                Blob: ({ filePath }) => makeBlobContentResult(`content for: ${filePath}`),
            })

            await driver.page.goto(`${driver.sourcegraphBaseUrl}/${repositoryName}/-/blob/${fileName}`)
            await driver.page.waitForSelector('.e2e-repo-blob')
            const blobContent = await driver.page.evaluate(() => {
                const editorArea = document.querySelector<HTMLDivElement>('.e2e-repo-blob')
                document.querySelector('a')
                return editorArea ? editorArea.textContent : null
            })

            // editor shows the return string content from Blob request
            assert.strictEqual(blobContent, `content for: ${fileName}`)

            // collect all files/links visible the the FILES tab
            const allFilesInTheTree = await driver.page.evaluate(() => {
                // TODO is there a better way to get all of them?
                const allFiles = document.querySelectorAll<HTMLAnchorElement>('.tree__row-contents')

                return [...allFiles].map(fileAnchor => ({
                    content: fileAnchor.textContent,
                    href: fileAnchor.href,
                }))
            })

            // files from TreeEntries request
            assert.deepStrictEqual(
                allFilesInTheTree,
                ['README.md', fileName].map(name => ({
                    content: name,
                    href: `${driver.sourcegraphBaseUrl}/${repositoryName}/-/blob/${name}`,
                }))
            )
        })
    })
})
