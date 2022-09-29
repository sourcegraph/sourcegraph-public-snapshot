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
} from './graphQlResponseHelpers'
import { commonWebGraphQlResults, createViewerSettingsGraphQLOverride } from './graphQlResults'

describe('CodeMirror blob view', () => {
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

    const { graphqlResults, filePaths } = createBlobPageData({
        repoName: 'github.com/sourcegraph/jsonrpc2',
        blobInfo: {
            'test.ts': {
                content: 'line1\nLine2\nline3',
            },
        },
    })

    const commonBlobGraphQlResults: Partial<WebGraphQlOperations & SharedGraphQlOperations> = {
        ...commonWebGraphQlResults,
        ...createViewerSettingsGraphQLOverride({
            user: {
                experimentalFeatures: {
                    enableCodeMirrorFileView: true,
                },
            },
        }),
        ...graphqlResults,
    }

    beforeEach(() => {
        testContext.overrideGraphQL(commonBlobGraphQlResults)
    })

    const blobSelector = '[data-testid="repo-blob"] .cm-editor'

    describe('in-document search', () => {
        function getMatchCount(): Promise<number> {
            return driver.page.evaluate<() => number>(() => document.querySelectorAll('.cm-searchMatch').length)
        }

        async function pressCtrlF(): Promise<void> {
            await driver.page.keyboard.down('Control')
            await driver.page.keyboard.press('f')
            await driver.page.keyboard.up('Control')
        }

        function getSelectedMatch(): Promise<string | null | undefined> {
            return driver.page.evaluate<() => string | null | undefined>(
                () => document.querySelector('.cm-searchMatch-selected')?.textContent
            )
        }

        it('renders a working in-document search', async () => {
            await driver.page.goto(`${driver.sourcegraphBaseUrl}${filePaths['test.ts']}`)
            await driver.page.waitForSelector(blobSelector)
            // Wait for page to "settle" so that focus management works better
            await driver.page.waitForTimeout(1000)

            // Focus file view and trigger in-document search
            await driver.page.click(blobSelector)
            await pressCtrlF()
            await driver.page.waitForSelector('.search-container')

            // Start searching (which implies that the search input has focus)
            await driver.page.keyboard.type('line')
            // All three lines should have matches
            assert.strictEqual(await getMatchCount(), 3, 'finds three matches')

            // Enable case sensitive search. This should update the matches
            // immediately.
            await driver.page.click('[data-testid="blob-view-search-case-sensitive"]')
            assert.strictEqual(await getMatchCount(), 2, 'finds two matches')

            // Pressing CTRL+f again focuses the search input again and selects
            // the value so that it can be easily replaced.
            await pressCtrlF()
            await driver.page.keyboard.type('line\\d')
            assert.strictEqual(
                await driver.page.evaluate<() => string | null | undefined>(
                    () => document.querySelector<HTMLInputElement>('.search-container [name="search"]')?.value
                ),
                'line\\d'
            )

            // Enabling regexp search.
            await driver.page.click('[data-testid="blob-view-search-regexp"]')
            assert.strictEqual(await getMatchCount(), 2, 'finds two matches')

            // Pressing previous / next buttons focuses next/previous match
            await driver.page.click('[data-testid="blob-view-search-next"]')
            const selectedMatch = await getSelectedMatch()
            assert.strictEqual(!!selectedMatch, true, 'match is selected')

            await driver.page.click('[data-testid="blob-view-search-previous"]')
            assert.notStrictEqual(selectedMatch, await getSelectedMatch())

            // Pressing Esc closes the search form
            await driver.page.keyboard.press('Escape')
            assert.strictEqual(
                await driver.page.evaluate(() => document.querySelector('.search-container')),
                null,
                'search form is not presetn'
            )
        })

        it('opens in-document when pressing ctrl-f anywhere on the page', async () => {
            await driver.page.goto(`${driver.sourcegraphBaseUrl}${filePaths['test.ts']}`)
            await driver.page.waitForSelector(blobSelector)
            // Wait for page to "settle" so that focus management works better
            await driver.page.waitForTimeout(1000)

            // Focus file view and trigger in-document search
            await driver.page.click('body')
            await pressCtrlF()
            await driver.page.waitForSelector('.search-container')
        })
    })
})

interface BlobInfo {
    [fileName: string]: {
        content: string
        html?: string
    }
}

function createBlobPageData<T extends BlobInfo>({
    repoName,
    blobInfo,
}: {
    repoName: string
    blobInfo: T
}): {
    graphqlResults: Pick<WebGraphQlOperations, 'ResolveRepoRev' | 'FileExternalLinks' | 'Blob' | 'FileNames'> &
        Pick<SharedGraphQlOperations, 'TreeEntries'>
    filePaths: { [k in keyof T]: string }
} {
    const repositorySourcegraphUrl = `/${repoName}`
    const fileNames = Object.keys(blobInfo)

    return {
        filePaths: fileNames.reduce((paths, fileName) => {
            paths[fileName as keyof T] = `/${repoName}/-/blob/${fileName}`
            return paths
        }, {} as { [k in keyof T]: string }),
        graphqlResults: {
            ResolveRepoRev: () => createResolveRepoRevisionResult(repositorySourcegraphUrl),
            FileExternalLinks: ({ filePath }) =>
                createFileExternalLinksResult(`https://${repoName}/blob/master/${filePath}`),
            TreeEntries: () => createTreeEntriesResult(repositorySourcegraphUrl, fileNames),
            Blob: ({ filePath }) => createBlobContentResult(blobInfo[filePath].content, blobInfo[filePath].html),
            FileNames: () => ({
                repository: {
                    id: 'repo-123',
                    __typename: 'Repository',
                    commit: {
                        id: 'c0ff33',
                        __typename: 'GitCommit',
                        fileNames,
                    },
                },
            }),
        },
    }
}
