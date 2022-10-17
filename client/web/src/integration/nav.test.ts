import { subDays } from 'date-fns'
import expect from 'expect'
import { test } from 'mocha'

import { encodeURIPathComponent } from '@sourcegraph/common'
import { mixedSearchStreamEvents, SearchGraphQlOperations } from '@sourcegraph/search'
import { SharedGraphQlOperations } from '@sourcegraph/shared/src/graphql-operations'
import { Driver, createDriverForTest } from '@sourcegraph/shared/src/testing/driver'
import { afterEachSaveScreenshotIfFailed } from '@sourcegraph/shared/src/testing/screenshotReporter'

import { NotebookFields, WebGraphQlOperations } from '../graphql-operations'

import { WebIntegrationTestContext, createWebIntegrationTestContext } from './context'
import {
    createResolveRepoRevisionResult,
    createFileExternalLinksResult,
    createBlobContentResult,
    createTreeEntriesResult,
    createFileNamesResult,
} from './graphQlResponseHelpers'
import { commonWebGraphQlResults } from './graphQlResults'
import { createEditorAPI } from './utils'

const commonSearchGraphQLResults: Partial<WebGraphQlOperations & SharedGraphQlOperations & SearchGraphQlOperations> = {
    ...commonWebGraphQlResults,
    FileNames: () => createFileNamesResult(),
    ResolveRepoRev: () => createResolveRepoRevisionResult('/github.com/sourcegraph/sourcegraph'),
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
    OverwriteSettings: () => ({
        settingsMutation: {
            overwriteSettings: {
                empty: {
                    alwaysNil: null,
                },
            },
        },
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
        testContext.overrideSearchStreamEvents(mixedSearchStreamEvents)
    })

    afterEachSaveScreenshotIfFailed(() => driver.page)
    afterEach(() => testContext?.dispose())

    describe('repo file page', () => {
        // This test covers the case that the query state shouldn't be updated
        // from the URL if it doesn't contain a query (it should not override
        // the repo and file information in the query input).
        // The initial load will work but updates to the URL that do not change
        // the repo or file name should also preserve the input value.
        test('query input contains repo and file name', async () => {
            await driver.page.goto(driver.sourcegraphBaseUrl + '/github.com/sourcegraph/sourcegraph/-/blob/README.md')
            await (await driver.page.waitForSelector('.test-breadcrumb-part-last'))?.click()

            const input = await createEditorAPI(driver, '.test-query-input')
            expect(await input.getValue()).toEqual('repo:^github\\.com/sourcegraph/sourcegraph$ file:^README\\.md')
        })
    })

    describe('Code Search Dropdown', () => {
        test('is highlighted on search page', async () => {
            await driver.page.goto(driver.sourcegraphBaseUrl + '/search?q=test&patternType=regexp')
            await driver.page.waitForSelector('[data-test-id="/search"]')
            await driver.page.waitForSelector('[data-test-active="true"]')

            const active = await driver.page.evaluate(() =>
                document.querySelector('[data-test-id="/search"]')?.getAttribute('data-test-active')
            )

            expect(active).toEqual('true')
        })

        test('is highlighted on repo page', async () => {
            await driver.page.goto(driver.sourcegraphBaseUrl + '/github.com/sourcegraph/sourcegraph')
            await driver.page.waitForSelector('[data-test-id="/search"]')
            await driver.page.waitForSelector('[data-test-active="true"]')

            const active = await driver.page.evaluate(() =>
                document.querySelector('[data-test-id="/search"]')?.getAttribute('data-test-active')
            )

            expect(active).toEqual('true')
        })

        test('is highlighted on repo file page', async () => {
            await driver.page.goto(driver.sourcegraphBaseUrl + '/github.com/sourcegraph/sourcegraph/-/blob/README.md')
            await driver.page.waitForSelector('[data-test-id="/search"]')
            await driver.page.waitForSelector('[data-test-active="true"]')

            const active = await driver.page.evaluate(() =>
                document.querySelector('[data-test-id="/search"]')?.getAttribute('data-test-active')
            )

            expect(active).toEqual('true')
        })

        test('is not highlighted on notebook page', async () => {
            await driver.page.goto(driver.sourcegraphBaseUrl + '/notebooks/id')
            await driver.page.waitForSelector('[data-test-id="/search"]')
            await driver.page.waitForSelector('[data-test-active="false"]')

            const active = await driver.page.evaluate(() =>
                document.querySelector('[data-test-id="/search"]')?.getAttribute('data-test-active')
            )

            expect(active).toEqual('false')
        })
    })
})
