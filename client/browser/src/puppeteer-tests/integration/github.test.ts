import { Driver, createDriverForTest } from '../../../../shared/src/testing/driver'
import { BrowserIntegrationTestContext, createBrowserIntegrationTestContext } from './context'
import { BrowserGraphQlOperations } from '../../graphql-operations'
import { SharedGraphQlOperations } from '../../../../shared/src/graphql-operations'
import { afterEachSaveScreenshotIfFailed } from '../../../../shared/src/testing/screenshotReporter'
import { closeInstallPageTab, testSingleFilePage } from '../shared'
import delay from 'delay'

describe('GitHub', () => {
    let driver: Driver
    before(async () => {
        driver = await createDriverForTest({ loadExtension: true })
    })
    after(() => driver?.close())
    let testContext: BrowserIntegrationTestContext
    beforeEach(async function () {
        testContext = await createBrowserIntegrationTestContext({
            driver,
            currentTest: this.currentTest!,
            directory: __dirname,
        })
        testContext.server.get('https://collector.githubapp.com/*').intercept((request, response) => {
            response.sendStatus(200)
        })
        testContext.server.get('https://github.githubassets.com/favicons/*').intercept((request, response) => {
            response.sendStatus(200)
        })
        await delay(3000)
        await closeInstallPageTab(driver.browser)
        await driver.setExtensionSourcegraphUrl()
    })
    afterEachSaveScreenshotIfFailed(() => driver.page)
    afterEach(() => testContext?.dispose())

    // const repositoryName = 'github.com/sourcegraph/jsonrpc2'
    // const fileName = 'test.ts'

    const commonBlobGraphQlResults: Partial<BrowserGraphQlOperations & SharedGraphQlOperations> = {
        // ...commonBrowserGraphQlResults,
        // RepositoryRedirect: ({ repoName }) => createRepositoryRedirectResult(repoName),
        // ResolveRev: () => createResolveRevisionResult(repositorySourcegraphUrl),
        // FileExternalLinks: ({ filePath }) =>
        //     createFileExternalLinksResult(`https://${repositoryName}/blob/master/${filePath}`),
        // TreeEntries: () => createTreeEntriesResult(repositorySourcegraphUrl, ['README.md', fileName]),
        // Blob: ({ filePath }) => createBlobContentResult(`content for: ${filePath}\nsecond line\nthird line`),
    }
    beforeEach(() => {
        testContext.overrideGraphQL(commonBlobGraphQlResults)
    })

    testSingleFilePage({
        getDriver: () => driver,
        url: 'https://github.com/sourcegraph/jsonrpc2/blob/4fb7cd90793ee6ab445f466b900e6bffb9b63d78/call_opt.go',
        repoName: 'github.com/sourcegraph/jsonrpc2',
        // Not using '.js-file-line' because it breaks the reliance on :nth-child() in testSingleFilePage()
        lineSelector: '.js-file-line-container tr',
        goToDefinitionURL:
            'https://github.com/sourcegraph/jsonrpc2/blob/4fb7cd90793ee6ab445f466b900e6bffb9b63d78/call_opt.go#L5:6',
    })
})
