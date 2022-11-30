import assert from 'assert'

import { ElementHandle, MouseButton } from 'puppeteer'
import type * as sourcegraph from 'sourcegraph'

import { JsonDocument, SyntaxKind } from '@sourcegraph/shared/src/codeintel/scip'
import { SharedGraphQlOperations } from '@sourcegraph/shared/src/graphql-operations'
import { ExtensionManifest } from '@sourcegraph/shared/src/schema/extensionSchema'
import { Settings } from '@sourcegraph/shared/src/schema/settings.schema'
import { Driver, createDriverForTest, percySnapshot } from '@sourcegraph/shared/src/testing/driver'
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
import { createEditorAPI, EditorAPI } from './utils'

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
        ...createViewerSettingsGraphQLOverride({
            user: {
                experimentalFeatures: {
                    enableCodeMirrorFileView: true,
                },
            },
        }),
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
                Object.entries(filePaths).map(([name, path]) => ({
                    content: name,
                    href: `${driver.sourcegraphBaseUrl}${path}`,
                }))
            )
        })

        it('truncates long file paths properly', async () => {
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

        it('selects a line when clicking the line', async () => {
            await driver.page.goto(`${driver.sourcegraphBaseUrl}${filePaths['test.ts']}`)
            await waitForView()
            await driver.page.click(lineAt(1))

            // Line is selected
            await driver.page.waitForSelector(lineAt(1) + '.selected-line')

            // URL is updated
            await driver.assertWindowLocation(`${filePaths['test.ts']}?L1`)
        })

        // This should also test the "back' button, but that test passed with
        // puppeteer regardless of the implementation.
        for (const button of ['forward', 'middle', 'right'] as MouseButton[]) {
            it(`does not select a line on ${button} button click`, async () => {
                await driver.page.goto(`${driver.sourcegraphBaseUrl}${filePaths['test.ts']}`)
                await waitForView()

                await driver.page.click(lineAt(1), { button })
                await driver.page.waitForSelector(lineAt(1) + '.selected-line', { hidden: true, timeout: 5000 })
            })
        }

        it('does not select a line when clicking on content in the line', async () => {
            await driver.page.goto(`${driver.sourcegraphBaseUrl}${filePaths['test.ts']}`)
            await waitForView()
            await driver.page.click(wordSelector)

            // Line is not selected
            await driver.page.waitForSelector(lineAt(1) + '.selected-line', { hidden: true })

            // URL is not updated
            await driver.assertWindowLocation(`${filePaths['test.ts']}`)
        })

        it('selects a line when clicking the line number', async () => {
            await driver.page.goto(`${driver.sourcegraphBaseUrl}${filePaths['test.ts']}`)
            await waitForView()
            await (await getLineNumberElement(5)).click()

            // Line is selected
            await driver.page.waitForSelector(lineAt(5) + '.selected-line')

            // URL is updated
            await driver.assertWindowLocation(`${filePaths['test.ts']}?L5`)
        })

        describe('line range selection', () => {
            it('selects a line range when shift-clicking lines', async () => {
                await driver.page.goto(`${driver.sourcegraphBaseUrl}${filePaths['test.ts']}`)
                await waitForView()

                await driver.page.click(lineAt(1))
                await driver.page.keyboard.down('Shift')
                await driver.page.click(lineAt(3))
                await driver.page.keyboard.up('Shift')

                // Lines is selected
                await Promise.all(
                    [1, 2, 3].map(lineNumber => driver.page.waitForSelector(lineAt(lineNumber) + '.selected-line'))
                )

                // URL is updated
                await driver.assertWindowLocation(`${filePaths['test.ts']}?L1-3`)
            })

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
                        driver.page.waitForSelector(lineAt(lineNumber) + '.selected-line')
                    )
                )

                // URL is updated
                await driver.assertWindowLocation(`${filePaths['test.ts']}?L1-5`)
            })

            it('selects a line range when dragging over line numbers', async () => {
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
                        driver.page.waitForSelector(lineAt(lineNumber) + '.selected-line')
                    )
                )

                // URL is updated
                await driver.assertWindowLocation(`${filePaths['test.ts']}?L1-5`)
            })
        })
    })

    // Describes the ways the blob viewer can be extended through Sourcegraph extensions.
    describe('extensibility', () => {
        beforeEach(() => {
            testContext.overrideJsContext({ enableLegacyExtensions: true })
        })

        function lineDecorationAt(line: number): string {
            return `${lineAt(line)} [data-testid=line-decoration]`
        }

        describe('hovercards', () => {
            beforeEach(() => {
                const { graphqlResults: extensionGraphQlResult, intercept, userSettings } = createExtensionData([
                    {
                        id: 'test',
                        extensionID: 'test/test',
                        extensionManifest: {
                            url: new URL(
                                '/-/static/extension/0001-test-test.js?hash--test-test',
                                driver.sourcegraphBaseUrl
                            ).href,
                            activationEvents: ['*'],
                        },
                        bundle: function extensionBundle(): void {
                            // eslint-disable-next-line @typescript-eslint/no-require-imports, @typescript-eslint/no-var-requires
                            const sourcegraph = require('sourcegraph') as typeof import('sourcegraph')

                            function activate(context: sourcegraph.ExtensionContext): void {
                                context.subscriptions.add(
                                    sourcegraph.languages.registerHoverProvider([{ language: 'typescript' }], {
                                        provideHover: () => ({
                                            contents: {
                                                kind: sourcegraph.MarkupKind.Markdown,
                                                value: 'Test hover content',
                                            },
                                        }),
                                    })
                                )
                            }

                            exports.activate = activate
                        },
                    },
                ])
                testContext.overrideGraphQL({
                    ...commonBlobGraphQlResults,
                    ...createViewerSettingsGraphQLOverride({
                        user: {
                            ...userSettings,
                            experimentalFeatures: {
                                enableCodeMirrorFileView: true,
                            },
                        },
                    }),
                    ...extensionGraphQlResult,
                })

                // Serve a mock extension bundle with a simple hover provider
                intercept(testContext, driver)
            })

            it('shows a hover overlay from a hover provider when a token is hovered', async () => {
                await driver.page.goto(`${driver.sourcegraphBaseUrl}${filePaths['test.ts']}`)
                await waitForView()
                await driver.page.hover(wordSelector)
                await driver.page.waitForSelector('.cm-code-intel-hovercard')
                assert.strictEqual(
                    await driver.page.evaluate(
                        (): string =>
                            document.querySelector('[data-testid="hover-overlay-contents"]')?.textContent?.trim() ?? ''
                    ),
                    'Test hover content',
                    'hovercard is visible with correct content'
                )

                await driver.page.hover(lineAt(5))
                try {
                    await driver.page.waitForSelector('.cm-code-intel-hovercard', { hidden: true })
                } catch {
                    throw new Error('Timeout waiting for hovercard to disappear')
                }
            })

            it('pins a hovercard and unpins hovercards', async () => {
                await driver.page.goto(`${driver.sourcegraphBaseUrl}${filePaths['test.ts']}`)
                await waitForView()
                await driver.page.hover(wordSelector)
                await driver.page.waitForSelector('.cm-code-intel-hovercard [data-testid="hover-copy-link"]')

                await driver.page.click('.cm-code-intel-hovercard [data-testid="hover-copy-link"]')

                // URL gets updated
                await driver.assertWindowLocation(`${filePaths['test.ts']}?L1:1&popover=pinned`)

                // Close button is visible
                await driver.page.waitForSelector('.cm-code-intel-hovercard [aria-label="Close"]')

                // Hovercard stay open when moving the mouse away
                await driver.page.hover(lineAt(5))
                await driver.page.waitForSelector('.cm-code-intel-hovercard')

                // Closes hovercard when clicking on another line
                await driver.page.click(lineAt(5))
                try {
                    await driver.page.waitForSelector('.cm-code-intel-hovercard', { hidden: true })
                } catch {
                    throw new Error('Timeout waiting for hovercard to close after selecting another line')
                }

                // Opens pinned hovecard when navigating back
                await driver.page.goBack()
                await driver.page.waitForSelector('.cm-code-intel-hovercard')

                // Closes hover card when clicking the close button
                await driver.page.click('.cm-code-intel-hovercard [aria-label="Close"]')
                try {
                    await driver.page.waitForSelector('.cm-code-intel-hovercard', { hidden: true })
                } catch {
                    throw new Error('Timeout waiting for hovercard to close after clicking close button')
                }
            })

            it('opens a pinned hovercard on page load', async () => {
                await driver.page.goto(`${driver.sourcegraphBaseUrl}${filePaths['test.ts']}?L1:1&popover=pinned`)
                await waitForView()
                await driver.page.waitForSelector('.cm-code-intel-hovercard')
            })
        })

        /**
         * This test is meant to prevent regression: https://github.com/sourcegraph/sourcegraph/pull/15099
         */
        it('adds and clears line decoration attachments properly', async () => {
            const { graphqlResults: extensionsGraphqlResult, intercept, userSettings } = createExtensionData([
                {
                    id: 'test',
                    extensionID: 'test/fixed-line',
                    extensionManifest: {
                        url: new URL(
                            '/-/static/extension/0001-test-fixed-line.js?hash--test-fixed-line',
                            driver.sourcegraphBaseUrl
                        ).href,
                        activationEvents: ['*'],
                    },
                    bundle: function extensionBundle(): void {
                        // eslint-disable-next-line @typescript-eslint/no-require-imports, @typescript-eslint/no-var-requires
                        const sourcegraph = require('sourcegraph') as typeof import('sourcegraph')

                        const fixedLineDecorationType = sourcegraph.app.createDecorationType()

                        function decorateLineTwo(editor: sourcegraph.CodeEditor) {
                            // Always decorate line 2
                            editor.setDecorations(fixedLineDecorationType, [
                                {
                                    range: new sourcegraph.Range(
                                        new sourcegraph.Position(1, 0),
                                        new sourcegraph.Position(1, 0)
                                    ),
                                    after: {
                                        contentText: 'fixed line content',
                                        backgroundColor: 'red',
                                    },
                                },
                            ])
                        }

                        function activate(context: sourcegraph.ExtensionContext): void {
                            // check initial viewer in case it was already emitted before extension was activated
                            const initialViewer = sourcegraph.app.activeWindow?.activeViewComponent
                            if (initialViewer?.type === 'CodeEditor') {
                                decorateLineTwo(initialViewer)
                            }

                            // subscribe to viewer changes
                            let previousViewerSubscription: sourcegraph.Unsubscribable | undefined
                            context.subscriptions.add(
                                sourcegraph.app.activeWindowChanges.subscribe(activeWindow => {
                                    const viewerSubscription = activeWindow?.activeViewComponentChanges.subscribe(
                                        viewer => {
                                            if (viewer?.type === 'CodeEditor') {
                                                // Always decorate line 2
                                                decorateLineTwo(viewer)
                                            }
                                        }
                                    )
                                    if (previousViewerSubscription) {
                                        previousViewerSubscription.unsubscribe()
                                    }
                                    previousViewerSubscription = viewerSubscription
                                })
                            )
                        }

                        exports.activate = activate
                    },
                },
                {
                    id: 'selected-line',
                    extensionID: 'test/selected-line',
                    extensionManifest: {
                        url: new URL(
                            '/-/static/extension/0001-test-selected-line.js?hash--test-selected-line',
                            driver.sourcegraphBaseUrl
                        ).href,
                        activationEvents: ['*'],
                    },
                    bundle: function extensionBundle(): void {
                        // eslint-disable-next-line @typescript-eslint/no-require-imports, @typescript-eslint/no-var-requires
                        const sourcegraph = require('sourcegraph') as typeof import('sourcegraph')

                        const selectedLineDecorationType = sourcegraph.app.createDecorationType()

                        function activate(context: sourcegraph.ExtensionContext): void {
                            let previousViewerSubscription: sourcegraph.Unsubscribable | undefined
                            let previousSelectionsSubscription: sourcegraph.Unsubscribable | undefined

                            context.subscriptions.add(
                                sourcegraph.app.activeWindowChanges.subscribe(activeWindow => {
                                    const viewerSubscription = activeWindow?.activeViewComponentChanges.subscribe(
                                        viewer => {
                                            let selectionsSubscription: sourcegraph.Unsubscribable | undefined
                                            if (viewer?.type === 'CodeEditor') {
                                                selectionsSubscription = viewer.selectionsChanges.subscribe(
                                                    selections => {
                                                        viewer.setDecorations(
                                                            selectedLineDecorationType,
                                                            selections.map(selection => ({
                                                                range: new sourcegraph.Range(
                                                                    selection.start,
                                                                    selection.end
                                                                ),
                                                                after: {
                                                                    contentText: `selected line content for line ${selection.start.line}`,
                                                                    backgroundColor: 'green',
                                                                },
                                                            }))
                                                        )
                                                    }
                                                )
                                            }

                                            if (previousSelectionsSubscription) {
                                                previousSelectionsSubscription.unsubscribe()
                                            }
                                            previousSelectionsSubscription = selectionsSubscription
                                        }
                                    )

                                    if (previousViewerSubscription) {
                                        previousViewerSubscription.unsubscribe()
                                    }
                                    previousViewerSubscription = viewerSubscription
                                })
                            )
                        }

                        exports.activate = activate
                    },
                },
            ])

            testContext.overrideGraphQL({
                ...commonBlobGraphQlResults,
                ...createViewerSettingsGraphQLOverride({
                    user: {
                        experimentalFeatures: {
                            enableCodeMirrorFileView: true,
                        },
                        ...userSettings,
                    },
                }),
                ...extensionsGraphqlResult,
            })

            // Serve mock extension bundles
            intercept(testContext, driver)

            const timeout = 10000
            await driver.page.goto(`${driver.sourcegraphBaseUrl}${filePaths['test.ts']}`)

            await waitForView()

            // Select line 1. Line 1
            await driver.page.click(lineAt(1))
            await driver.page.waitForSelector(lineDecorationAt(1), { timeout })
            assert(await driver.page.$(lineDecorationAt(1)), 'Expected line 1 to have a decoration attachment portal')
            assert(await driver.page.$(lineDecorationAt(2)), 'Expected line 2 to have a decoration attachment portal')
            assert.strictEqual(
                await driver.page.evaluate(
                    (selector: string) => document.querySelector(selector)?.childElementCount,
                    lineDecorationAt(2)
                ),
                1,
                'Expected line 2 to have 1 decoration'
            )

            // Select line 2. Assert that everything is normal
            await driver.page.click(lineAt(2))
            await driver.page.waitForSelector(lineDecorationAt(1), { timeout, hidden: true })
            assert(
                !(await driver.page.$(lineDecorationAt(1))),
                'Expected line 1 to not have a decoration attachment portal after selecting line 2'
            )
            assert(await driver.page.$(lineDecorationAt(2)), 'Expected line 2 to have a decoration attachment portal')

            // Count child nodes of existing portals
            assert.strictEqual(
                await driver.page.evaluate(
                    (selector: string) => document.querySelector(selector)?.childElementCount,
                    lineDecorationAt(2)
                ),
                2,
                'Expected line 2 to have 2 decorations'
            )

            // Select line 1 again. before fix, line 2 will still have 2 decorations
            await driver.page.click(lineAt(1))
            await driver.page.waitForSelector(lineDecorationAt(1), { timeout })
            assert(
                await driver.page.$(lineDecorationAt(1)),
                'Expected line 1 to have a decoration attachment portal after it is reselected'
            )
            assert(
                await driver.page.$(lineDecorationAt(2)),
                'Expected line 2 to have a decoration attachment portal after selecting line 1 again'
            )

            // Count child nodes of existing portals
            assert.strictEqual(
                await driver.page.evaluate(
                    (selector: string) => document.querySelector(selector)?.childElementCount,
                    lineDecorationAt(2)
                ),
                1,
                'Expected line 2 to have 1 decoration'
            )
        })

        it('sends the latest document to extensions', async () => {
            // This test is meant to prevent regression of
            // "extensions receive wrong text documents": https://github.com/sourcegraph/sourcegraph/issues/14965

            /**
             * How can we verify that extensions receive the latest document?
             * The test extension has to cause some change detectable to the web application, and
             * this change must be dependent on the text document. This extension should be simple to
             * avoid bugs in the extension itself.
             *
             * Simplest possible extension that satisfies these requirements:
             * add attachment to lines that contain a certain word.
             */

            const { graphqlResults: extensionsGraphqlResult, intercept, userSettings } = createExtensionData([
                {
                    id: 'word-finder',
                    extensionID: 'test/word-finder',
                    extensionManifest: {
                        url: new URL(
                            '/-/static/extension/0001-test-word-finder.js?hash--test-word-finder',
                            driver.sourcegraphBaseUrl
                        ).href,
                        activationEvents: ['*'],
                    },
                    bundle: function extensionBundle(): void {
                        // eslint-disable-next-line @typescript-eslint/no-require-imports, @typescript-eslint/no-var-requires
                        const sourcegraph = require('sourcegraph') as typeof import('sourcegraph')

                        const fixedLineDecorationType = sourcegraph.app.createDecorationType()

                        // Match all occurrences of 'word', decorate lines
                        function decorateWordLines(editor: sourcegraph.CodeEditor) {
                            const text = editor.document.text ?? ''

                            const wordPattern = /word/g
                            const matchIndices: number[] = []

                            for (const match of text.matchAll(wordPattern)) {
                                if (match.index) {
                                    matchIndices.push(match.index)
                                }
                            }

                            editor.setDecorations(
                                fixedLineDecorationType,
                                matchIndices.map(index => {
                                    const line = editor.document.positionAt(index).line

                                    return {
                                        range: new sourcegraph.Range(
                                            new sourcegraph.Position(line, 0),
                                            new sourcegraph.Position(line, 0)
                                        ),
                                        after: {
                                            contentText: 'found word',
                                            backgroundColor: 'red',
                                        },
                                    }
                                })
                            )
                        }

                        function activate(context: sourcegraph.ExtensionContext): void {
                            // check initial viewer in case it was already emitted before extension was activated
                            const initialViewer = sourcegraph.app.activeWindow?.activeViewComponent
                            if (initialViewer?.type === 'CodeEditor') {
                                decorateWordLines(initialViewer)
                            }

                            // subscribe to viewer changes
                            let previousViewerSubscription: sourcegraph.Unsubscribable | undefined
                            context.subscriptions.add(
                                sourcegraph.app.activeWindowChanges.subscribe(activeWindow => {
                                    const viewerSubscription = activeWindow?.activeViewComponentChanges.subscribe(
                                        viewer => {
                                            if (viewer?.type === 'CodeEditor') {
                                                decorateWordLines(viewer)
                                            }
                                        }
                                    )
                                    if (previousViewerSubscription) {
                                        previousViewerSubscription.unsubscribe()
                                    }
                                    previousViewerSubscription = viewerSubscription
                                })
                            )
                        }

                        exports.activate = activate
                    },
                },
            ])

            const { graphqlResults: blobGraphqlResults, filePaths } = createBlobPageData({
                repoName,
                blobInfo: {
                    'test.ts': {
                        content: '// First line\n// Second word line\n// Third line',
                    },
                    'fake.ts': {
                        content: '// First word line\n// Second line\n// Third word line',
                    },
                },
            })

            testContext.overrideGraphQL({
                ...commonBlobGraphQlResults,
                ...blobGraphqlResults,
                ...extensionsGraphqlResult,
                ...createViewerSettingsGraphQLOverride({
                    user: {
                        experimentalFeatures: {
                            enableCodeMirrorFileView: true,
                        },
                        ...userSettings,
                    },
                }),
            })

            // Serve the word-finder extension bundle
            intercept(testContext, driver)

            const timeout = 5000
            await driver.page.goto(`${driver.sourcegraphBaseUrl}${filePaths['test.ts']}`)
            await waitForView()

            // File 1 (test.ts). Only line two contains 'word'
            try {
                await driver.page.waitForSelector(lineDecorationAt(2), { timeout })
            } catch {
                // Rethrow with contextual error message
                throw new Error(`Timeout waiting for ${lineDecorationAt(2)} (test.ts, first time)`)
            }
            // await driver.page.waitFor(1000)
            assert(
                !(await driver.page.$(lineDecorationAt(1))),
                'Expected line 1 to not have a decoration attachment on test.ts'
            )
            assert(
                await driver.page.$(lineDecorationAt(2)),
                'Expected line 2 to have a decoration attachment on test.ts'
            )
            assert(
                !(await driver.page.$(lineDecorationAt(3))),
                'Expected line 3 to not have a decoration attachment on test.ts'
            )

            // Go to File 2 (fake.ts). Lines one and three contain 'word'
            await driver.findElementWithText('fake.ts', {
                selector: '.test-tree-file-link',
                action: 'click',
            })
            try {
                await driver.page.waitForSelector(lineDecorationAt(1), { timeout })
            } catch {
                throw new Error(`Timeout waiting for ${lineDecorationAt(1)} (fake.ts)`)
            }
            assert(
                await driver.page.$(lineDecorationAt(1)),
                'Expected line 1 to have a decoration attachment on fake.ts'
            )
            assert(
                !(await driver.page.$(lineDecorationAt(2))),
                'Expected line 2 to not have a decoration attachment on fake.ts'
            )
            assert(
                await driver.page.$(lineDecorationAt(3)),
                'Expected line 3 to have a decoration attachment on fake.ts'
            )

            // Come back to File 1 (test.ts)
            await driver.findElementWithText('test.ts', {
                selector: '.test-tree-file-link',
                action: 'click',
            })
            try {
                await driver.page.waitForSelector(lineDecorationAt(2), { timeout })
            } catch {
                throw new Error(`Timeout waiting for ${lineDecorationAt(2)} (test.ts, second time)`)
            }
            assert(
                !(await driver.page.$(lineDecorationAt(1))),
                'Expected line 1 to not have a decoration attachment on test.ts (second visit)'
            )
            assert(
                await driver.page.$(lineDecorationAt(2)),
                'Expected line 2 to have a decoration attachment on test.ts (second visit)'
            )
            assert(
                !(await driver.page.$(lineDecorationAt(3))),
                'Expected line 3 to not have a decoration attachment on test.ts (second visit)'
            )
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

        it('renders a working in-document search', async () => {
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
    graphqlResults: Pick<WebGraphQlOperations, 'ResolveRepoRev' | 'FileExternalLinks' | 'Blob' | 'FileNames'> &
        Pick<SharedGraphQlOperations, 'TreeEntries' | 'LegacyRepositoryIntrospection' | 'LegacyResolveRepo2'>
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
            Blob: ({ filePath }) =>
                createBlobContentResult(blobInfo[filePath].content, blobInfo[filePath].html, blobInfo[filePath].lsif),
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
            LegacyRepositoryIntrospection: () => ({
                __type: {
                    fields: [
                        {
                            name: 'noFork',
                        },
                    ],
                },
            }),
            LegacyResolveRepo2: () => ({
                repository: {
                    id: repoName,
                    name: repoName,
                },
            }),
        },
    }
}

interface MockExtension {
    id: string
    extensionID: string
    extensionManifest: ExtensionManifest
    /**
     * A function whose body is a Sourcegraph extension.
     *
     * Bundle must import 'sourcegraph' (e.g. `const sourcegraph = require('sourcegraph')`)
     * */
    bundle: () => void
}

function createExtensionData(
    extensions: MockExtension[]
): {
    intercept: (testContext: WebIntegrationTestContext, driver: Driver) => void
    graphqlResults: Pick<SharedGraphQlOperations, 'Extensions'>
    userSettings: Required<Pick<Settings, 'extensions'>>
} {
    return {
        intercept(testContext: WebIntegrationTestContext, driver: Driver) {
            for (const extension of extensions) {
                testContext.server
                    .get(new URL(extension.extensionManifest.url, driver.sourcegraphBaseUrl).href)
                    .intercept((_request, response) => {
                        // Create an immediately-invoked function expression for the extensionBundle function
                        const extensionBundleString = `(${extension.bundle.toString()})()`
                        response.type('application/javascript; charset=utf-8').send(extensionBundleString)
                    })
            }
        },
        graphqlResults: {
            Extensions: () => ({
                extensionRegistry: {
                    __typename: 'ExtensionRegistry',
                    extensions: {
                        nodes: extensions.map(extension => ({
                            ...extension,
                            manifest: { jsonFields: extension.extensionManifest },
                        })),
                    },
                },
            }),
        },
        userSettings: {
            extensions: extensions.reduce((extensionsSettings: Record<string, boolean>, mockExtension) => {
                extensionsSettings[mockExtension.extensionID] = true
                return extensionsSettings
            }, {}),
        },
    }
}
