import { subDays } from 'date-fns'
import expect from 'expect'
import { test } from 'mocha'

import { encodeURIPathComponent } from '@sourcegraph/common'
import { SearchGraphQlOperations } from '@sourcegraph/search'
import { SharedGraphQlOperations } from '@sourcegraph/shared/src/graphql-operations'
import { Driver, createDriverForTest } from '@sourcegraph/shared/src/testing/driver'
import { afterEachSaveScreenshotIfFailed } from '@sourcegraph/shared/src/testing/screenshotReporter'

import { NotebookFields, WebGraphQlOperations } from '../graphql-operations'

import { WebIntegrationTestContext, createWebIntegrationTestContext } from './context'
import {
    createRepositoryRedirectResult,
    createResolveRevisionResult,
    createFileExternalLinksResult,
    createBlobContentResult,
    createTreeEntriesResult,
    createFileNamesResult,
} from './graphQlResponseHelpers'
import { commonWebGraphQlResults } from './graphQlResults'

const commonSearchGraphQLResults: Partial<WebGraphQlOperations & SharedGraphQlOperations & SearchGraphQlOperations> = {
    ...commonWebGraphQlResults,
    FileNames: () => createFileNamesResult(),
    RepositoryRedirect: ({ repoName }) => createRepositoryRedirectResult(repoName),
    ResolveRev: () => createResolveRevisionResult('/github.com/sourcegraph/sourcegraph'),
    FileExternalLinks: ({ filePath, repoName, revision }) =>
        createFileExternalLinksResult(
            `https://${encodeURIPathComponent(repoName)}/blob/${encodeURIPathComponent(
                revision
            )}/${encodeURIPathComponent(filePath)}`
        ),
    TreeEntries: () => createTreeEntriesResult('/github.com/sourcegraph/sourcegraph', ['README.md']),
    Blob: ({ filePath }) => createBlobContentResult(`content for: ${filePath}\nsecond line\nthird line`),
    FetchNotebook: ({ id }) => ({
        node: notebookFixture(id, 'Notebook Title', [
            { __typename: 'MarkdownBlock', id: '1', markdownInput: '# Title' },
            { __typename: 'QueryBlock', id: '2', queryInput: 'query' },
        ]),
    }),
}

const now = new Date()

const notebookFixture = (id: string, title: string, blocks: NotebookFields['blocks']): NotebookFields => ({
    __typename: 'Notebook',
    id,
    title,
    createdAt: subDays(now, 5).toISOString(),
    updatedAt: subDays(now, 5).toISOString(),
    public: true,
    viewerCanManage: true,
    viewerHasStarred: true,
    namespace: { __typename: 'User', id: '1', namespaceName: 'user1' },
    stars: { totalCount: 123 },
    creator: { __typename: 'User', username: 'user1' },
    updater: { __typename: 'User', username: 'user1' },
    blocks,
})

describe('GlobalNavbar', () => {
    describe('Code Search Dropdown', () => {
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

            testContext.overrideGraphQL(commonSearchGraphQLResults)
        })

        afterEachSaveScreenshotIfFailed(() => driver.page)
        afterEach(() => testContext?.dispose())

        test('is highlighted on search page', async () => {
            await driver.page.goto(driver.sourcegraphBaseUrl + '/search?q=test&patternType=regexp')

            const active = await driver.page.evaluate(() =>
                document.querySelector('[data-test-id="/search"]')?.getAttribute('data-test-active')
            )

            expect(active).toEqual('true')
        })

        test('is highlighted on repo page', async () => {
            await driver.page.goto(driver.sourcegraphBaseUrl + '/github.com/sourcegraph/sourcegraph')

            const active = await driver.page.evaluate(() =>
                document.querySelector('[data-test-id="/search"]')?.getAttribute('data-test-active')
            )

            expect(active).toEqual('true')
        })

        test('is highlighted on repo file page', async () => {
            await driver.page.goto(driver.sourcegraphBaseUrl + '/github.com/sourcegraph/sourcegraph/-/blob/README.md')

            const active = await driver.page.evaluate(() =>
                document.querySelector('[data-test-id="/search"]')?.getAttribute('data-test-active')
            )

            expect(active).toEqual('true')
        })

        test('is not highlighted on notebook page', async () => {
            await driver.page.goto(driver.sourcegraphBaseUrl + '/notebooks/id')

            const active = await driver.page.evaluate(() =>
                document.querySelector('[data-test-id="/search"]')?.getAttribute('data-test-active')
            )

            expect(active).toEqual('false')
        })
    })
})
