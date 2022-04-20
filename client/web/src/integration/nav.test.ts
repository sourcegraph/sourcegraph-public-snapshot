import expect from 'expect'
import { test } from 'mocha'

import { encodeURIPathComponent } from '@sourcegraph/common'
import { SearchGraphQlOperations } from '@sourcegraph/search'
import { SharedGraphQlOperations } from '@sourcegraph/shared/src/graphql-operations'
import { Driver, createDriverForTest } from '@sourcegraph/shared/src/testing/driver'
import { afterEachSaveScreenshotIfFailed } from '@sourcegraph/shared/src/testing/screenshotReporter'

import { WebGraphQlOperations } from '../graphql-operations'

import { WebIntegrationTestContext, createWebIntegrationTestContext } from './context'
import {
    createRepositoryRedirectResult,
    createResolveRevisionResult,
    createFileExternalLinksResult,
    createBlobContentResult,
    createTreeEntriesResult,
} from './graphQlResponseHelpers'
import { commonWebGraphQlResults } from './graphQlResults'

const commonSearchGraphQLResults: Partial<WebGraphQlOperations & SharedGraphQlOperations & SearchGraphQlOperations> = {
    ...commonWebGraphQlResults,
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
}

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

        test('is not highlighted on batch changes page', async () => {
            await driver.page.goto(driver.sourcegraphBaseUrl + '/batch-changes')

            const active = await driver.page.evaluate(() =>
                document.querySelector('[data-test-id="/search"]')?.getAttribute('data-test-active')
            )

            expect(active).toEqual('false')
        })
    })
})
