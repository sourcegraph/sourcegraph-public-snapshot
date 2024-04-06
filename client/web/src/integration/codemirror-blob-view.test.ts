import assert from 'assert'

import { afterEach, beforeEach, describe, it } from 'mocha'
import type { ElementHandle, MouseButton } from 'puppeteer'

import { type JsonDocument, SyntaxKind } from '@sourcegraph/shared/src/codeintel/scip'
import type { SharedGraphQlOperations } from '@sourcegraph/shared/src/graphql-operations'
import { type Driver, createDriverForTest, percySnapshot } from '@sourcegraph/shared/src/testing/driver'
import { afterEachSaveScreenshotIfFailed } from '@sourcegraph/shared/src/testing/screenshotReporter'

import type { WebGraphQlOperations } from '../graphql-operations'

import { type WebIntegrationTestContext, createWebIntegrationTestContext } from './context'
import {
    createResolveRepoRevisionResult,
    createFileExternalLinksResult,
    createTreeEntriesResult,
    createBlobContentResult,
    createFileTreeEntriesResult,
} from './graphQlResponseHelpers'
import { commonWebGraphQlResults } from './graphQlResults'
import { createEditorAPI, type EditorAPI } from './utils'

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

    const repoName = 'github.com/sourcegraph/jsonrpc2'
    const { graphqlResults: blobGraphqlResults, filePaths } = createBlobPageData({
        repoName,
        blobInfo: {
            'test.ts': {
                content: ['line1', 'line2', 'line3', 'line4', 'line5'].join('\n'),
                // This is used to create a span element around the text `line1`
                // which can later be target by tests (e.g. for hover). We
                // cannot specify a custom CSS class to add. Using
                // `SyntaxKind.Tag` will add the class `hl-typed-Tag`. This will
                // break when we decide to change the class name format.
                lsif: {
                    occurrences: [{ range: [0, 0, 5], syntaxKind: SyntaxKind.Tag }],
                },
            },
            'README.md': {
                content: 'README.md',
            },
            'this_is_a_long_file_path/apps/rest-showcase/src/main/java/org/demo/rest/example/OrdersController.java': {
                content: 'line1\nline2\nline3',
            },
        },
    })

    const commonBlobGraphQlResults: Partial<WebGraphQlOperations & SharedGraphQlOperations> = {
        ...commonWebGraphQlResults,
        ...blobGraphqlResults,
    }

    beforeEach(() => {
        testContext.overrideGraphQL(commonBlobGraphQlResults)
    })

    const blobSelector = '[data-testid="repo-blob"] .cm-editor'
    const wordSelector = blobSelector + ' .hl-typed-Tag'

    function waitForView(): Promise<EditorAPI> {
        return createEditorAPI(driver, '[data-testid="repo-blob"]')
    }

    function lineAt(line: number): string {
        return `${blobSelector} .cm-line:nth-child(${line})`
    }

    describe('general layout for viewing a file', () => {
        it('populates editor content and FILES tab', async () => {
            await driver.page.goto(`${driver.sourcegraphBaseUrl}${filePaths['test.ts']}`)
            const view = await waitForView()
            const blobContent = await view.getValue()

            // editor shows the return string content from Blob request
            assert.strictEqual(blobContent, 'line1\nline2\nline3\nline4\nline5')

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
                Object.entries(filePaths)
                    .filter(
                        ([name]) =>
                            name !==
                            // This file is not part of the same directory so it won't be shown in this test case
                            'this_is_a_long_file_path/apps/rest-showcase/src/main/java/org/demo/rest/example/OrdersController.java'
                    )
                    .map(([name, path]) => ({
                        content: name,
                        href: `${driver.sourcegraphBaseUrl}${path}`,
                    }))
            )
        })

        // TODO 53389: This test is disabled because it is flaky.
        it.skip('truncates long file paths properly', async () => {
            await driver.page.goto(
                `${driver.sourcegraphBaseUrl}${filePaths['this_is_a_long_file_path/apps/rest-showcase/src/main/java/org/demo/rest/example/OrdersController.java']}`
            )
            await waitForView()
            await driver.page.waitForSelector('.test-breadcrumb')
            await percySnapshot(driver.page, 'truncates long file paths properly')
        })
    })

    describe('line number redirects', () => {
        it('should redirect from line number hash to query parameter', async () => {
            await driver.page.goto(`${driver.sourcegraphBaseUrl}${filePaths['test.ts']}#2`)
            await waitForView()
            await driver.assertWindowLocation(`${filePaths['test.ts']}?L2`)
        })

        it('should redirect from line range hash to query parameter', async () => {
            await driver.page.goto(`${driver.sourcegraphBaseUrl}${filePaths['test.ts']}#1-3`)
            await waitForView()
            await driver.assertWindowLocation(`${filePaths['test.ts']}?L1-3`)
        })
    })

    describe('line selection', () => {
        async function getLineNumberElement(lineNumber: number): Promise<ElementHandle> {
            const lineNumberElement = (
                await driver.page.evaluateHandle(
                    (blobSelector: string, lineNumber: number): HTMLElement | null => {
                        const lineNumberElements = document.querySelectorAll<HTMLDivElement>(
                            `${blobSelector} .cm-lineNumbers .cm-gutterElement`
                        )
                        for (const element of lineNumberElements) {
                            if (Number(element.textContent) === lineNumber) {
                                return element
                            }
                        }
                        return null
                    },
                    blobSelector,
                    lineNumber
                )
            ).asElement()
            assert(lineNumberElement, `found line number element ${lineNumber}`)
            return lineNumberElement
        }

        // This should also test the "back' button, but that test passed with
        // puppeteer regardless of the implementation.
        for (const button of ['forward', 'middle', 'right'] as MouseButton[]) {
            it(`does not select a line on ${button} button click`, async () => {
                await driver.page.goto(`${driver.sourcegraphBaseUrl}${filePaths['test.ts']}`)
                await waitForView()

                await driver.page.click(lineAt(1), { button })
                await driver.page.waitForSelector(lineAt(1) + "[data-testid='selected-line']", {
                    hidden: true,
                    timeout: 5000,
                })
            })
        }

        it('does not select a line when clicking on content in the line', async () => {
            await driver.page.goto(`${driver.sourcegraphBaseUrl}${filePaths['test.ts']}`)
            await waitForView()
            await driver.page.click(wordSelector)

            // Line is not selected
            await driver.page.waitForSelector(lineAt(1) + "[data-testid='selected-line']", { hidden: true })

            // URL is not updated
            await driver.assertWindowLocation(`${filePaths['test.ts']}`)
        })

        it('selects a line when clicking the line number', async () => {
            await driver.page.goto(`${driver.sourcegraphBaseUrl}${filePaths['test.ts']}`)
            await waitForView()
            await (await getLineNumberElement(5)).click()

            // Line is selected
            await driver.page.waitForSelector(lineAt(5) + "[data-testid='selected-line']")

            // URL is updated
            await driver.assertWindowLocation(`${filePaths['test.ts']}?L5`)
        })

        describe('line range selection', () => {
            it('selects a line range when shift-clicking line numbers', async () => {
                await driver.page.goto(`${driver.sourcegraphBaseUrl}${filePaths['test.ts']}`)
                await waitForView()

                await (await getLineNumberElement(1)).click()
                await driver.page.keyboard.down('Shift')
                await (await getLineNumberElement(5)).click()
                await driver.page.keyboard.up('Shift')

                // Line is selected
                await Promise.all(
                    [1, 2, 3, 4, 5].map(lineNumber =>
                        driver.page.waitForSelector(lineAt(lineNumber) + "[data-testid='selected-line']")
                    )
                )

                // URL is updated
                await driver.assertWindowLocation(`${filePaths['test.ts']}?L1-5`)
            })

            it.skip('selects a line range when dragging over line numbers', async () => {
                await driver.page.goto(`${driver.sourcegraphBaseUrl}${filePaths['test.ts']}`)
                await waitForView()

                {
                    const startLineNumberPoint = await (await getLineNumberElement(1)).clickablePoint()
                    const endLineNumberPoint = await (await getLineNumberElement(5)).clickablePoint()
                    await driver.page.mouse.move(startLineNumberPoint.x, startLineNumberPoint.y)
                    await driver.page.mouse.down()
                    await driver.page.mouse.move(endLineNumberPoint.x, endLineNumberPoint.y)
                    await driver.page.mouse.up()
                }

                // Line is selected
                await Promise.all(
                    [1, 2, 3, 4, 5].map(lineNumber =>
                        driver.page.waitForSelector(lineAt(lineNumber) + "[data-testid='selected-line']")
                    )
                )

                // URL is updated
                await driver.assertWindowLocation(`${filePaths['test.ts']}?L1-5`)
            })
        })
    })

    describe('in-document search', () => {
        const { graphqlResults: blobGraphqlResults, filePaths } = createBlobPageData({
            repoName,
            blobInfo: {
                'test.ts': {
                    content: 'line1\nLine2\nline3',
                },
            },
        })
        beforeEach(() => {
            testContext.overrideGraphQL({
                ...commonBlobGraphQlResults,
                ...blobGraphqlResults,
            })
        })

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

        it.skip('renders a working in-document search', async () => {
            await driver.page.goto(`${driver.sourcegraphBaseUrl}${filePaths['test.ts']}`)
            await driver.page.waitForSelector(blobSelector)
            // Wait for page to "settle" so that focus management works better
            await driver.page.waitForTimeout(1000)

            // Focus file view and trigger in-document search
            await driver.page.click(blobSelector)
            await pressCtrlF()
            await driver.page.waitForSelector('.cm-sg-search-container')

            // Start searching (which implies that the search input has focus)
            await driver.page.keyboard.type('line')
            // Wait for search input debounce timeout (100ms)
            await driver.page.waitForTimeout(150)
            // All three lines should have matches
            assert.strictEqual(await getMatchCount(), 3, 'finds three matches')

            // Enable case sensitive search. This should update the matches
            // immediately.
            await driver.page.click('.test-blob-view-search-case-sensitive')
            assert.strictEqual(await getMatchCount(), 2, 'finds two matches')

            // Pressing CTRL+f again focuses the search input again and selects
            // the value so that it can be easily replaced.
            await pressCtrlF()
            await driver.page.keyboard.type('line\\d')
            assert.strictEqual(
                await driver.page.evaluate<() => string | null | undefined>(
                    () => document.querySelector<HTMLInputElement>('.cm-sg-search-container [name="search"]')?.value
                ),
                'line\\d'
            )

            // Enabling regexp search.
            await driver.page.click('.test-blob-view-search-regexp')
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
                await driver.page.evaluate(() => document.querySelector('.cm-sg-search-container')),
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
            await driver.page.waitForSelector('.cm-sg-search-container')
        })
    })
})

interface BlobInfo {
    [fileName: string]: {
        content: string
        html?: string
        lsif?: JsonDocument
    }
}

function createBlobPageData<T extends BlobInfo>({
    repoName,
    blobInfo,
}: {
    repoName: string
    blobInfo: T
}): {
    graphqlResults: Pick<
        WebGraphQlOperations,
        'ResolveRepoRev' | 'FileTreeEntries' | 'FileExternalLinks' | 'Blob' | 'FileNames'
    > &
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
            FileTreeEntries: () => createFileTreeEntriesResult(repositorySourcegraphUrl, fileNames),
            Blob: ({ filePath }) => createBlobContentResult(blobInfo[filePath].content, blobInfo[filePath].lsif),
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
